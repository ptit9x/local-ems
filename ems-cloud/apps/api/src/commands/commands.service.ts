import { Injectable, NotFoundException } from '@nestjs/common';
import { randomUUID } from 'crypto';
import { PrismaService } from '../prisma/prisma.service';
import { MqttIngestionService } from '../telemetry/mqtt-ingestion.service';
import { ConfigService } from '@nestjs/config';

@Injectable()
export class CommandsService {
  constructor(private readonly prisma: PrismaService, private readonly mqtt: MqttIngestionService, private readonly config: ConfigService) {}

  async create(siteId: string, tenantId: string, type: string, payload?: object) {
    const site = await this.prisma.site.findFirst({ where: { id: siteId, tenantId } });
    if (!site) throw new NotFoundException('Site not found');
    if (!['identify', 'update_tariff', 'update_config', 'restart'].includes(type)) throw new NotFoundException('Unsupported command');
    const requestId = randomUUID();
    const command = await this.prisma.command.create({ data: { requestId, siteId, type, payload } });
    try {
      const prefix = this.config.get('MQTT_TOPIC_PREFIX', 'ems');
      await this.mqtt.publish(`${prefix}/${site.siteCode}/command`, { schema_version: 1, site_id: site.siteCode, command: type, request_id: requestId, timestamp: new Date().toISOString(), payload });
      return this.prisma.command.update({ where: { id: command.id }, data: { status: 'PUBLISHED', publishedAt: new Date() } });
    } catch (error) {
      await this.prisma.command.update({ where: { id: command.id }, data: { status: 'FAILED', result: { error: String(error) } } });
      throw error;
    }
  }

  async get(id: string, tenantId: string) {
    const command = await this.prisma.command.findFirst({ where: { id, site: { tenantId } } });
    if (!command) throw new NotFoundException('Command not found');
    return command;
  }
}
