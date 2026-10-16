import { NotFoundException } from '@nestjs/common';
import { CommandsService } from './commands.service';

describe('CommandsService', () => {
  const prisma = {
    site: { findFirst: jest.fn() },
    command: { create: jest.fn(), update: jest.fn(), findFirst: jest.fn() },
  };
  const mqtt = { publish: jest.fn() };
  const config = { get: jest.fn((_key: string, fallback: string) => fallback) };
  let service: CommandsService;

  beforeEach(() => {
    jest.clearAllMocks();
    service = new CommandsService(prisma as never, mqtt as never, config as never);
  });

  it('publishes an allowlisted command only for a tenant-owned site', async () => {
    prisma.site.findFirst.mockResolvedValue({ id: 'site-a', siteCode: 'sg-01' });
    prisma.command.create.mockResolvedValue({ id: 'command-a' });
    prisma.command.update.mockResolvedValue({ id: 'command-a', status: 'PUBLISHED' });
    mqtt.publish.mockResolvedValue(undefined);

    await service.create('site-a', 'tenant-a', 'identify');

    expect(prisma.site.findFirst).toHaveBeenCalledWith({ where: { id: 'site-a', tenantId: 'tenant-a' } });
    expect(mqtt.publish).toHaveBeenCalledWith('ems/sg-01/command', expect.objectContaining({ command: 'identify', site_id: 'sg-01' }));
  });

  it('rejects commands for sites outside the tenant', async () => {
    prisma.site.findFirst.mockResolvedValue(null);
    await expect(service.create('site-b', 'tenant-a', 'restart')).rejects.toBeInstanceOf(NotFoundException);
    expect(mqtt.publish).not.toHaveBeenCalled();
  });

  it('rejects command types outside the allowlist', async () => {
    prisma.site.findFirst.mockResolvedValue({ id: 'site-a', siteCode: 'sg-01' });
    await expect(service.create('site-a', 'tenant-a', 'shell')).rejects.toBeInstanceOf(NotFoundException);
    expect(mqtt.publish).not.toHaveBeenCalled();
  });
});
