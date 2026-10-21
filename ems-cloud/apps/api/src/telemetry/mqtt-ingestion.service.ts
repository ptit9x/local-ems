import { Injectable, Logger, OnModuleInit, OnModuleDestroy } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { connect, MqttClient } from 'mqtt';
import { TelemetryService } from './telemetry.service';
import { FailoverService } from '../failover/failover.service';

/**
 * MqttIngestionService subscribes to MQTT telemetry topics
 * and ingests incoming messages into the database.
 */
@Injectable()
export class MqttIngestionService implements OnModuleInit, OnModuleDestroy {
  private readonly logger = new Logger(MqttIngestionService.name);
  private client: MqttClient | null = null;
  private readonly topicPrefix: string;
  private readonly maxPayloadBytes = 256 * 1024;
  private offlineTimer?: NodeJS.Timeout;

  constructor(
    private readonly config: ConfigService,
    private readonly telemetryService: TelemetryService,
    private readonly failoverService: FailoverService,
  ) {
    this.topicPrefix = this.config.get<string>('MQTT_TOPIC_PREFIX', 'ems');
  }

  async onModuleInit() {
	this.offlineTimer = setInterval(() => this.markOfflineSites().catch(err => this.logger.error(`Offline monitor: ${err.message}`)), 30000);
    const brokerUrl = this.config.get<string>('MQTT_BROKER_URL');
    if (!brokerUrl) {
      this.logger.warn('MQTT_BROKER_URL not configured — MQTT ingestion disabled');
      return;
    }

    this.client = connect(brokerUrl, {
      clientId: this.config.get<string>('MQTT_CLIENT_ID', 'ems-cloud-api'),
      username: this.config.get<string>('MQTT_USERNAME') || undefined,
      password: this.config.get<string>('MQTT_PASSWORD') || undefined,
      clean: false, // durable session
      reconnectPeriod: 5000,
      connectTimeout: 15000,
    });

    this.client.on('connect', () => {
      this.logger.log(`Connected to MQTT broker: ${brokerUrl}`);
      this.subscribe();
    });

    this.client.on('reconnect', () => {
      this.logger.log('Reconnecting to MQTT broker...');
    });

    this.client.on('error', (err) => {
      this.logger.error(`MQTT error: ${err.message}`);
    });

    this.client.on('message', (topic, payload) => {
      this.handleMessage(topic, payload).catch((err) => {
        this.logger.error(`Failed to handle message on ${topic}: ${err.message}`);
      });
    });
  }

  async onModuleDestroy() {
    if (this.offlineTimer) clearInterval(this.offlineTimer);
    if (this.client) {
      this.client.end(true);
      this.logger.log('MQTT client disconnected');
    }
  }

  private subscribe() {
    if (!this.client) return;

    // Subscribe to all site telemetry: ems/+/telemetry
    const telemetryTopic = `${this.topicPrefix}/+/telemetry`;
    this.client.subscribe(telemetryTopic, { qos: 1 }, (err) => {
      if (err) {
        this.logger.error(`Failed to subscribe to ${telemetryTopic}: ${err.message}`);
      } else {
        this.logger.log(`Subscribed to ${telemetryTopic}`);
      }
    });

    const ackTopic = `${this.topicPrefix}/+/command/ack`;
    this.client.subscribe(ackTopic, { qos: 1 }, (err) => {
      if (err) {
        this.logger.error(`Failed to subscribe to ${ackTopic}: ${err.message}`);
      } else {
        this.logger.log(`Subscribed to ${ackTopic}`);
      }
    });

    this.client.subscribe(`${this.topicPrefix}/+/heartbeat`, { qos: 0 });
    this.client.subscribe(`${this.topicPrefix}/+/failover/recovery`, { qos: 1 });
  }

  private async handleMessage(topic: string, payload: Buffer) {
    if (payload.length > this.maxPayloadBytes) throw new Error('payload exceeds 256 KiB limit');
    const parts = topic.split('/');
    if (parts.length < 3) return;

    const isCommandAck = parts.slice(-2).join('/') === 'command/ack';
    const isFailoverRecovery = parts.slice(-2).join('/') === 'failover/recovery';
    const siteCode = parts[parts.length - (isCommandAck || isFailoverRecovery ? 3 : 2)];
    const messageType = parts[parts.length - 1];

    if (messageType === 'telemetry') {
      const data = JSON.parse(payload.toString());
      if (data.schema_version !== 1 || !data.message_id || data.site_id !== siteCode || !data.timestamp || !data.interval_sec || !data.cycle_count) {
        throw new Error('invalid telemetry contract');
      }
      await this.telemetryService.ingest(siteCode, data);
    } else if (messageType === 'heartbeat') {
      const data = JSON.parse(payload.toString());
      await this.telemetryService.heartbeat(siteCode, data);
      if (data.schema_version === 2 && data.control_owner === 'local') {
        const command = await this.failoverService.recoverOnHeartbeat(siteCode);
        if (command) await this.publish(`${this.topicPrefix}/${siteCode}/failover/control`, command);
      }
    } else if (isCommandAck) {
      const data = JSON.parse(payload.toString());
      await this.telemetryService.acknowledgeCommand(siteCode, data);
    } else if (isFailoverRecovery) {
      const data = JSON.parse(payload.toString());
      if (data.schema_version !== 2 || data.site_id !== siteCode || !Number.isSafeInteger(data.lease_epoch)) throw new Error('invalid recovery contract');
      const command = await this.failoverService.completeHandover(siteCode, BigInt(data.lease_epoch));
      if (command) await this.publish(`${this.topicPrefix}/${siteCode}/failover/control`, command);
    }
  }

  isConnected() { return this.client?.connected ?? false; }

  publish(topic: string, payload: object): Promise<void> {
    if (!this.client?.connected) return Promise.reject(new Error('MQTT unavailable'));
    return new Promise((resolve, reject) => this.client!.publish(topic, JSON.stringify(payload), { qos: 1 }, (err) => err ? reject(err) : resolve()));
  }

  private async markOfflineSites() {
    const cutoff = new Date(Date.now() - 90000);
    const sites = await this.telemetryService.findStaleSites(cutoff);
    for (const site of sites) {
      await this.telemetryService.markOffline(site.id);
      const command = await this.failoverService.beginTakeover(site.id, site.siteCode, 'No heartbeat received for 90 seconds');
      if (command) await this.publish(`${this.topicPrefix}/${site.siteCode}/failover/control`, command);
    }
  }
}
