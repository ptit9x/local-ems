import { TelemetryService } from './telemetry.service';

describe('TelemetryService', () => {
  const prisma = {
    site: { findUnique: jest.fn(), findFirst: jest.fn(), update: jest.fn() },
    telemetryRecord: { create: jest.fn(), findMany: jest.fn(), findFirst: jest.fn() },
    heartbeat: { create: jest.fn() },
    alert: { updateMany: jest.fn() },
    command: { updateMany: jest.fn() },
    $transaction: jest.fn(),
  };
  let service: TelemetryService;

  beforeEach(() => {
    jest.clearAllMocks();
    service = new TelemetryService(prisma as never);
  });

  it('maps a valid telemetry contract and marks the site online', async () => {
    prisma.site.findUnique.mockResolvedValue({ id: 'site-a' });
    prisma.telemetryRecord.create.mockResolvedValue({});
    prisma.site.update.mockResolvedValue({});

    await service.ingest('sg-01', {
      schema_version: 1, message_id: 'message-a', site_id: 'sg-01',
      timestamp: '2026-10-16T00:00:00.000Z', interval_sec: 10, cycle_count: 10,
      avg_solar_w: 1200, avg_soc: 55.5,
    });

    expect(prisma.telemetryRecord.create).toHaveBeenCalledWith({ data: expect.objectContaining({
      siteId: 'site-a', messageId: 'message-a', avgSolarW: 1200, avgSoc: 55.5,
    }) });
    expect(prisma.site.update).toHaveBeenCalledWith({ where: { siteCode: 'sg-01' }, data: expect.objectContaining({ status: 'ONLINE' }) });
  });

  it('returns no data when a tenant queries a site it does not own', async () => {
    prisma.site.findFirst.mockResolvedValue(null);
    await expect(service.query('site-b', 'tenant-a', {})).resolves.toEqual([]);
    expect(prisma.telemetryRecord.findMany).not.toHaveBeenCalled();
  });

  it('rejects malformed heartbeat payloads before database writes', async () => {
    await expect(service.heartbeat('sg-01', { schema_version: 1, site_id: 'other' })).rejects.toThrow('invalid heartbeat contract');
    expect(prisma.site.findUnique).not.toHaveBeenCalled();
  });
});
