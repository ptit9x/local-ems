import { FailoverService } from './failover.service';

describe('FailoverService', () => {
  const prisma: any = {
    site: { findUnique: jest.fn(), findFirst: jest.fn() },
    failoverSession: { findFirst: jest.fn(), create: jest.fn(), update: jest.fn() },
    failoverEvent: { create: jest.fn(), findMany: jest.fn() },
  };
  let service: FailoverService;

  const ocpp: any = { applySafeMode: jest.fn().mockResolvedValue([]) };
  beforeEach(() => { jest.clearAllMocks(); ocpp.applySafeMode.mockResolvedValue([]); service = new FailoverService(prisma, ocpp); });

  it('increments the lease and enters cloud safe mode once', async () => {
    prisma.failoverSession.findFirst.mockResolvedValue({ leaseEpoch: 4n, state: 'LOCAL_ACTIVE' });
    prisma.failoverSession.create.mockResolvedValue({ id: 'session-5' });
    prisma.failoverSession.update.mockResolvedValue({});
    prisma.failoverEvent.create.mockResolvedValue({});
    const command = await service.beginTakeover('site-db-id', 'sg-01', 'heartbeat timeout');
    expect(command?.lease_epoch).toBe(5);
    expect(command?.safe_mode.allow_new_sessions).toBe(false);
    expect(ocpp.applySafeMode).toHaveBeenCalledWith(7000);
    expect(prisma.failoverSession.create).toHaveBeenCalledWith(expect.objectContaining({ data: expect.objectContaining({ leaseEpoch: 5n, state: 'CLOUD_TAKEOVER_PENDING' }) }));
  });

  it('does not create a second takeover while cloud owns control', async () => {
    prisma.failoverSession.findFirst.mockResolvedValue({ leaseEpoch: 5n, state: 'CLOUD_ACTIVE' });
    expect(await service.beginTakeover('site-db-id', 'sg-01', 'heartbeat timeout')).toBeNull();
    expect(prisma.failoverSession.create).not.toHaveBeenCalled();
  });

  it('reasserts the Cloud lease when recovery uses a stale epoch', async () => {
    prisma.site.findUnique.mockResolvedValue({ id: 'site-db-id' });
    prisma.failoverSession.findFirst.mockResolvedValue({ id: 'session-5', leaseEpoch: 5n, state: 'CLOUD_ACTIVE' });
    const response = await service.completeHandover('sg-01', 4n);
    expect(response?.action).toBe('takeover');
    expect(response?.lease_epoch).toBe(5);
    expect(prisma.failoverSession.update).not.toHaveBeenCalled();
  });
});
