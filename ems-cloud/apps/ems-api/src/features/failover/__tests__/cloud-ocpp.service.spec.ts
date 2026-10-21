import { ConfigService } from '@nestjs/config';
import WebSocket from 'ws';
import { CloudOcppService } from '../cloud-ocpp.service';

describe('CloudOcppService integration', () => {
  let service: CloudOcppService;
  const clients: WebSocket[] = [];

  afterEach(() => {
    for (const client of clients) client.close();
    clients.length = 0;
    service?.onModuleDestroy();
  });

  it('registers six backup chargers and applies safe mode', async () => {
    service = new CloudOcppService(new ConfigService({ OCPP_CLOUD_ENABLED: true, OCPP_CLOUD_PORT: 0 }));
    service.onModuleInit();
    const server: any = (service as any).server;
    await new Promise<void>(resolve => server.once('listening', resolve));
    const port = server.address().port;

    for (let i = 1; i <= 6; i += 1) {
      const client = new WebSocket(`ws://127.0.0.1:${port}/ocpp/ev-${i}`);
      clients.push(client);
      await new Promise<void>((resolve, reject) => { client.once('open', resolve); client.once('error', reject); });
      client.send(JSON.stringify([2, `boot-${i}`, 'BootNotification', { chargePointVendor: 'Simulator', chargePointModel: 'E2E' }]));
      await new Promise<void>(resolve => client.once('message', () => resolve()));
    }

    const responders = clients.map(client => new Promise<void>((resolve, reject) => {
      client.once('message', raw => {
        const message = JSON.parse(raw.toString());
        try {
          expect(message[2]).toBe('SetChargingProfile');
          expect(message[3].csChargingProfiles.chargingSchedule.chargingSchedulePeriod[0].limit).toBe(7000);
          client.send(JSON.stringify([3, message[1], { status: 'Accepted' }]));
          resolve();
        } catch (error) { reject(error); }
      });
    }));
    const resultPromise = service.applySafeMode(7000);
    await Promise.all(responders);
    const results = await resultPromise;
    expect(results).toHaveLength(6);
    expect(results.every(result => result.status === 'fulfilled')).toBe(true);
    expect(service.list()).toHaveLength(6);
  });
});
