import { Injectable, Logger } from '@nestjs/common';
import { Prisma } from '@prisma/client';
import { PrismaService } from '../prisma/prisma.service';
import { TelemetryQueryDto } from './dto/telemetry-query.dto';

interface TelemetryPayload {
  schema_version: 1;
  message_id: string;
  site_id: string;
  timestamp: string;
  interval_sec: number;
  avg_solar_w?: number;
  avg_load_w?: number;
  avg_grid_w?: number;
  avg_ess_w?: number;
  min_soc?: number;
  max_soc?: number;
  avg_soc?: number;
  max_temp_c?: number;
  energy_charge_wh?: number;
  energy_discharge_wh?: number;
  cycle_count: number;
  alerts?: string[];
}

@Injectable()
export class TelemetryService {
  private readonly logger = new Logger(TelemetryService.name);

  constructor(private readonly prisma: PrismaService) {}

  findStaleSites(cutoff: Date) {
    return this.prisma.site.findMany({ where: { status: 'ONLINE', OR: [{ lastSeenAt: null }, { lastSeenAt: { lt: cutoff } }] }, select: { id: true } });
  }

  async markOffline(siteId: string) {
    await this.prisma.$transaction(async tx => {
      await tx.site.update({ where: { id: siteId }, data: { status: 'OFFLINE' } });
      const existing = await tx.alert.findFirst({ where: { siteId, code: 'SITE_OFFLINE', active: true } });
      if (!existing) await tx.alert.create({ data: { siteId, code: 'SITE_OFFLINE', severity: 'CRITICAL', message: 'No heartbeat received for 90 seconds' } });
    });
  }

  /**
   * Ingest a telemetry message from MQTT into the database.
   */
  async ingest(siteCode: string, payload: TelemetryPayload): Promise<void> {
    // Find site by siteCode
    const site = await this.prisma.site.findUnique({
      where: { siteCode },
    });

    if (!site) {
      this.logger.warn(`Ignoring telemetry for unknown site: ${siteCode}`);
      return;
    }

    // Insert telemetry record
    try {
      await this.prisma.telemetryRecord.create({ data: {
        siteId: site.id,
        messageId: payload.message_id,
        schemaVersion: payload.schema_version,
        timestamp: new Date(payload.timestamp),
        intervalSec: payload.interval_sec,
        avgSolarW: payload.avg_solar_w ?? 0,
        avgLoadW: payload.avg_load_w ?? 0,
        avgGridW: payload.avg_grid_w ?? 0,
        avgEssW: payload.avg_ess_w ?? 0,
        minSoc: payload.min_soc ?? 0,
        maxSoc: payload.max_soc ?? 0,
        avgSoc: payload.avg_soc ?? 0,
        maxTempC: payload.max_temp_c ?? 0,
        energyChargeWh: payload.energy_charge_wh ?? 0,
        energyDischargeWh: payload.energy_discharge_wh ?? 0,
        cycleCount: payload.cycle_count,
        alerts: payload.alerts ? JSON.parse(JSON.stringify(payload.alerts)) : undefined,
      } });
    } catch (error) {
      // MQTT QoS 1 permits redelivery. The composite unique constraint makes the
      // database the final idempotency boundary, and duplicates are successful no-ops.
      if (error instanceof Prisma.PrismaClientKnownRequestError && error.code === 'P2002') {
        this.logger.debug(`Ignoring duplicate telemetry message: ${payload.message_id}`);
        return;
      }
      throw error;
    }

    // Update site last_seen_at
    await this.prisma.site.update({
      where: { siteCode },
      data: { lastSeenAt: new Date(), status: 'ONLINE' },
    });

    this.logger.log(`Ingested telemetry for ${siteCode}: ${payload.cycle_count} cycles`);
  }

  async heartbeat(siteCode: string, payload: { schema_version?: number; site_id?: string; timestamp?: string }) {
    if (payload.schema_version !== 1 || payload.site_id !== siteCode || !payload.timestamp) throw new Error('invalid heartbeat contract');
    const site = await this.prisma.site.findUnique({ where: { siteCode } });
    if (!site) return;
    const timestamp = new Date(payload.timestamp);
    await this.prisma.$transaction([
      this.prisma.heartbeat.create({ data: { siteId: site.id, timestamp, schemaVersion: 1 } }),
      this.prisma.site.update({ where: { id: site.id }, data: { lastSeenAt: timestamp, status: 'ONLINE' } }),
      this.prisma.alert.updateMany({ where: { siteId: site.id, code: 'SITE_OFFLINE', active: true }, data: { active: false, resolvedAt: new Date() } }),
    ]);
  }

  async acknowledgeCommand(siteCode: string, payload: { schema_version?: number; site_id?: string; request_id?: string; status?: string; result?: object }) {
    if (payload.schema_version !== 1 || payload.site_id !== siteCode || !payload.request_id) throw new Error('invalid command acknowledgment');
    await this.prisma.command.updateMany({
      where: { requestId: payload.request_id, site: { siteCode } },
      data: { status: payload.status === 'ok' || payload.status === 'accepted' ? 'ACKNOWLEDGED' : 'FAILED', acknowledgedAt: new Date(), result: payload.result },
    });
  }

  /**
   * Query telemetry records for a site within a time range.
   */
  async query(siteId: string, tenantId: string, dto: TelemetryQueryDto) {
    const site = await this.prisma.site.findFirst({ where: { id: siteId, tenantId } });
    if (!site) return [];
    const to = dto.to ? new Date(dto.to) : new Date();
    const from = dto.from ? new Date(dto.from) : new Date(to.getTime() - 24 * 3600000);
    return this.prisma.telemetryRecord.findMany({
      where: {
        siteId,
        timestamp: {
          gte: from,
          lte: to,
        },
      },
      orderBy: { timestamp: 'asc' },
      take: dto.limit || 1000,
    });
  }

  /**
   * Get the latest telemetry record for a site.
   */
  async latest(siteId: string, tenantId: string) {
    const site = await this.prisma.site.findFirst({ where: { id: siteId, tenantId } });
    if (!site) return null;
    return this.prisma.telemetryRecord.findFirst({
      where: { siteId },
      orderBy: { timestamp: 'desc' },
    });
  }

  /**
   * Get daily summary aggregation for a site.
   */
  async dailySummary(siteId: string, tenantId: string, date: string) {
    const site = await this.prisma.site.findFirst({ where: { id: siteId, tenantId } });
    if (!site) return { date, records: 0, summary: null };
    const startOfDay = new Date(date);
    const endOfDay = new Date(date);
    endOfDay.setDate(endOfDay.getDate() + 1);

    const records = await this.prisma.telemetryRecord.findMany({
      where: {
        siteId,
        timestamp: { gte: startOfDay, lt: endOfDay },
      },
    });

    if (records.length === 0) {
      return { date, records: 0, summary: null };
    }

    const totalCycles = records.reduce((sum, r) => sum + r.cycleCount, 0);
    const totalChargeWh = records.reduce((sum, r) => sum + r.energyChargeWh, 0);
    const totalDischargeWh = records.reduce((sum, r) => sum + r.energyDischargeWh, 0);
    const avgSoc = records.reduce((sum, r) => sum + r.avgSoc, 0) / records.length;
    const maxTemp = Math.max(...records.map((r) => r.maxTempC));

    return {
      date,
      records: records.length,
      summary: {
        total_cycles: totalCycles,
        energy_charge_kwh: totalChargeWh / 1000,
        energy_discharge_kwh: totalDischargeWh / 1000,
        avg_soc: Math.round(avgSoc * 10) / 10,
        max_temp_c: maxTemp,
      },
    };
  }
}
