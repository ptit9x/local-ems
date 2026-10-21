import { Injectable } from '@nestjs/common';
import { PrismaService } from '../../common/database/prisma.service';

@Injectable()
export class SitesService {
  constructor(private readonly prisma: PrismaService) {}

  async findAll(tenantId?: string) {
    return this.prisma.site.findMany({
      where: tenantId ? { tenantId } : undefined,
      orderBy: { createdAt: 'desc' },
    });
  }

  async findBySiteCode(siteCode: string) {
    return this.prisma.site.findUnique({ where: { siteCode } });
  }

  async findById(id: string, tenantId?: string) {
    return this.prisma.site.findFirst({ where: { id, ...(tenantId ? { tenantId } : {}) } });
  }

  async devices(id: string, tenantId: string) {
    const site = await this.findById(id, tenantId);
    if (!site) return null;
    return this.prisma.device.findMany({ where: { siteId: id }, orderBy: { externalId: 'asc' } });
  }

  async alerts(id: string, tenantId: string) {
    const site = await this.findById(id, tenantId);
    if (!site) return null;
    return this.prisma.alert.findMany({ where: { siteId: id }, orderBy: { createdAt: 'desc' }, take: 200 });
  }

  async acknowledgeAlert(siteId: string, alertId: string, tenantId: string) {
    const site = await this.findById(siteId, tenantId);
    if (!site) return null;
    return this.prisma.alert.updateMany({ where: { id: alertId, siteId }, data: { acknowledgedAt: new Date() } });
  }

  async updateLastSeen(siteCode: string) {
    return this.prisma.site.update({
      where: { siteCode },
      data: { lastSeenAt: new Date(), status: 'ONLINE' },
    });
  }
}
