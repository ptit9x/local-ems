import { Injectable } from '@nestjs/common';
import { randomUUID } from 'crypto';
import { PrismaService } from '../prisma/prisma.service';
import { CloudOcppService } from './cloud-ocpp.service';

export type FailoverAction = 'takeover_pending' | 'takeover' | 'request_recovery' | 'handover_complete' | 'fault';

@Injectable()
export class FailoverService {
  constructor(private readonly prisma: PrismaService, private readonly ocpp: CloudOcppService) {}

  async current(siteId: string) {
    return this.prisma.failoverSession.findFirst({ where: { siteId }, orderBy: { startedAt: 'desc' } });
  }

  async currentForTenant(siteId: string, tenantId: string) {
    const site = await this.prisma.site.findFirst({ where: { id: siteId, tenantId } });
    return site ? this.current(siteId) : null;
  }

  async beginTakeover(siteId: string, siteCode: string, reason: string) {
    const current = await this.current(siteId);
    if (current?.state === 'CLOUD_ACTIVE' || current?.state === 'CLOUD_TAKEOVER_PENDING') return null;
    const epoch = (current?.leaseEpoch ?? 0n) + 1n;
    const session = await this.prisma.failoverSession.create({ data: { siteId, state: 'CLOUD_TAKEOVER_PENDING', controlOwner: 'cloud', leaseEpoch: epoch, reason } });
    await this.event(siteId, session.id, 'CLOUD_TAKEOVER_PENDING', 'cloud', epoch, reason);
    await this.prisma.failoverSession.update({ where: { id: session.id }, data: { state: 'CLOUD_ACTIVE' } });
    await this.event(siteId, session.id, 'CLOUD_ACTIVE', 'cloud', epoch, 'Safe mode activated');
    await this.ocpp.applySafeMode(7000);
    return this.envelope(siteCode, 'takeover', epoch, reason);
  }

  async completeHandover(siteCode: string, reportedEpoch: bigint) {
    const site = await this.prisma.site.findUnique({ where: { siteCode } });
    if (!site) return null;
    const current = await this.current(site.id);
    if (!current) return this.envelope(siteCode, 'handover_complete', reportedEpoch, 'No Cloud lease; Local control confirmed');
    if (current.state === 'CLOUD_ACTIVE' && current.leaseEpoch !== reportedEpoch) {
      return this.envelope(siteCode, 'takeover', current.leaseEpoch, 'Cloud lease remains authoritative');
    }
    if (current.state === 'LOCAL_ACTIVE') return this.envelope(siteCode, 'handover_complete', current.leaseEpoch, 'Local control confirmed');
    if (current.leaseEpoch !== reportedEpoch || current.state !== 'CLOUD_ACTIVE') return null;
    await this.prisma.failoverSession.update({ where: { id: current.id }, data: { state: 'HANDOVER_PENDING' } });
    await this.event(site.id, current.id, 'HANDOVER_PENDING', 'cloud', current.leaseEpoch, 'Local recovery requested');
    await this.prisma.failoverSession.update({ where: { id: current.id }, data: { state: 'LOCAL_ACTIVE', controlOwner: 'local', endedAt: new Date() } });
    await this.event(site.id, current.id, 'LOCAL_ACTIVE', 'local', current.leaseEpoch, 'Handover complete');
    return this.envelope(siteCode, 'handover_complete', current.leaseEpoch, 'Local control restored');
  }

  async recoverOnHeartbeat(siteCode: string) {
    const site = await this.prisma.site.findUnique({ where: { siteCode } });
    if (!site) return null;
    const current = await this.current(site.id);
    if (!current || current.state !== 'CLOUD_ACTIVE') return null;
    return this.completeHandover(siteCode, current.leaseEpoch);
  }

  async timeline(siteId: string, tenantId: string) {
    const site = await this.prisma.site.findFirst({ where: { id: siteId, tenantId } });
    if (!site) return [];
    return this.prisma.failoverEvent.findMany({ where: { siteId }, orderBy: { createdAt: 'desc' }, take: 200 });
  }

  private event(siteId: string, sessionId: string, state: any, controlOwner: string, leaseEpoch: bigint, reason: string) {
    return this.prisma.failoverEvent.create({ data: { siteId, sessionId, state, controlOwner, leaseEpoch, reason } });
  }

  private envelope(siteId: string, action: FailoverAction, epoch: bigint, reason: string) {
    return { schema_version: 2, site_id: siteId, action, request_id: randomUUID(), timestamp: new Date().toISOString(), lease_epoch: Number(epoch), reason, safe_mode: { allow_new_sessions: false, max_power_per_charger_w: 7000 } };
  }
}

