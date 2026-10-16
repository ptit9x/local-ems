import { SitesService } from './sites.service';

describe('SitesService tenant isolation', () => {
  const prisma = {
    site: { findMany: jest.fn(), findFirst: jest.fn(), findUnique: jest.fn(), update: jest.fn() },
    device: { findMany: jest.fn() },
    alert: { findMany: jest.fn(), updateMany: jest.fn() },
  };
  const service = new SitesService(prisma as never);

  beforeEach(() => jest.clearAllMocks());

  it('always scopes site listing to the authenticated tenant', async () => {
    prisma.site.findMany.mockResolvedValue([]);
    await service.findAll('tenant-a');
    expect(prisma.site.findMany).toHaveBeenCalledWith(expect.objectContaining({ where: { tenantId: 'tenant-a' } }));
  });

  it('does not query devices when the site belongs to another tenant', async () => {
    prisma.site.findFirst.mockResolvedValue(null);
    await expect(service.devices('site-b', 'tenant-a')).resolves.toBeNull();
    expect(prisma.site.findFirst).toHaveBeenCalledWith({ where: { id: 'site-b', tenantId: 'tenant-a' } });
    expect(prisma.device.findMany).not.toHaveBeenCalled();
  });
});
