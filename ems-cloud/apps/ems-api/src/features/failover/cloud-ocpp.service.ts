import { Injectable, Logger, OnModuleDestroy, OnModuleInit } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { randomUUID } from 'crypto';
import { IncomingMessage } from 'http';
import { WebSocket, WebSocketServer } from 'ws';

interface ChargerConnection { id: string; socket: WebSocket; status: string; lastSeenAt: Date; }

@Injectable()
export class CloudOcppService implements OnModuleInit, OnModuleDestroy {
  private readonly logger = new Logger(CloudOcppService.name);
  private readonly chargers = new Map<string, ChargerConnection>();
  private readonly pending = new Map<string, { resolve: (value: unknown) => void; reject: (error: Error) => void; timer: NodeJS.Timeout }>();
  private server?: WebSocketServer;

  constructor(private readonly config: ConfigService) {}

  onModuleInit() {
    if (!this.config.get<boolean>('OCPP_CLOUD_ENABLED', false)) return;
    const port = this.config.get<number>('OCPP_CLOUD_PORT', 9001);
    this.server = new WebSocketServer({ port, maxPayload: 1024 * 1024 });
    this.server.on('connection', (socket, request) => this.register(socket, request));
    this.logger.log(`Cloud OCPP backup endpoint listening on :${port}/ocpp/{chargerId}`);
  }

  onModuleDestroy() {
    for (const charger of this.chargers.values()) charger.socket.close(1001, 'server shutdown');
    this.server?.close();
    for (const request of this.pending.values()) { clearTimeout(request.timer); request.reject(new Error('OCPP service stopped')); }
    this.pending.clear();
  }

  list() { return [...this.chargers.values()].map(({ id, status, lastSeenAt, socket }) => ({ id, status, lastSeenAt, connected: socket.readyState === WebSocket.OPEN })); }

  async applySafeMode(maxPowerW = 7000) {
    const commands = [...this.chargers.keys()].map(id => this.call(id, 'SetChargingProfile', {
      connectorId: 0,
      csChargingProfiles: { chargingProfileId: 9001, stackLevel: 100, chargingProfilePurpose: 'ChargePointMaxProfile', chargingProfileKind: 'Absolute', chargingSchedule: { chargingRateUnit: 'W', chargingSchedulePeriod: [{ startPeriod: 0, limit: maxPowerW }] } },
    }));
    return Promise.allSettled(commands);
  }

  private register(socket: WebSocket, request: IncomingMessage) {
    const expectedAuth = this.config.get<string>('OCPP_CLOUD_BASIC_AUTH');
    if (expectedAuth && request.headers.authorization !== `Basic ${Buffer.from(expectedAuth).toString('base64')}`) {
      socket.close(1008, 'authentication required');
      return;
    }
    const match = request.url?.match(/^\/ocpp\/([^/?]+)$/);
    if (!match) { socket.close(1008, 'expected /ocpp/{chargerId}'); return; }
    const id = decodeURIComponent(match[1]);
    const existing = this.chargers.get(id);
    existing?.socket.close(1000, 'replaced by new connection');
    this.chargers.set(id, { id, socket, status: 'Available', lastSeenAt: new Date() });
    socket.on('message', data => this.onMessage(id, data.toString()));
    socket.on('close', () => { if (this.chargers.get(id)?.socket === socket) this.chargers.delete(id); });
    socket.on('error', error => this.logger.warn(`OCPP ${id}: ${error.message}`));
  }

  private onMessage(id: string, raw: string) {
    let message: any[];
    try { message = JSON.parse(raw); } catch { return; }
    const charger = this.chargers.get(id);
    if (charger) charger.lastSeenAt = new Date();
    if (message[0] === 3 || message[0] === 4) {
      const pending = this.pending.get(message[1]);
      if (!pending) return;
      clearTimeout(pending.timer); this.pending.delete(message[1]);
      message[0] === 3 ? pending.resolve(message[2]) : pending.reject(new Error(message[3] || message[2] || 'OCPP command failed'));
      return;
    }
    if (message[0] !== 2 || typeof message[1] !== 'string' || typeof message[2] !== 'string') return;
    const [, uid, action, payload] = message;
    if (charger && action === 'StatusNotification') charger.status = payload?.status ?? charger.status;
    const response = action === 'BootNotification'
      ? { status: 'Accepted', currentTime: new Date().toISOString(), interval: 30 }
      : action === 'Heartbeat' ? { currentTime: new Date().toISOString() }
      : action === 'StartTransaction' ? { transactionId: Date.now(), idTagInfo: { status: 'Accepted' } }
      : action === 'StopTransaction' ? { idTagInfo: { status: 'Accepted' } }
      : {};
    charger?.socket.send(JSON.stringify([3, uid, response]));
  }

  private call(chargerId: string, action: string, payload: object, maxAttempts = 3) {
    const charger = this.chargers.get(chargerId);
    if (!charger || charger.socket.readyState !== WebSocket.OPEN) return Promise.reject(new Error(`charger ${chargerId} is offline`));
    const uid = randomUUID();
    return new Promise((resolve, reject) => {
      const frame = JSON.stringify([2, uid, action, payload]);
      let attempt = 0;
      const send = () => {
        attempt += 1;
        const timer = setTimeout(() => {
          if (attempt < maxAttempts && charger.socket.readyState === WebSocket.OPEN) { send(); return; }
          this.pending.delete(uid); reject(new Error(`OCPP ${action} timeout after ${attempt} attempts`));
        }, 10_000);
        this.pending.set(uid, { resolve, reject, timer });
        charger.socket.send(frame, error => {
          if (!error) return;
          clearTimeout(timer);
          if (attempt < maxAttempts && charger.socket.readyState === WebSocket.OPEN) { send(); return; }
          this.pending.delete(uid); reject(error);
        });
      };
      send();
    });
  }
}

