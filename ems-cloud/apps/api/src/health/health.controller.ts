import { Controller, Get, ServiceUnavailableException } from '@nestjs/common';
import { PrismaService } from '../prisma/prisma.service';

@Controller('health')
export class HealthController {
  constructor(private readonly prisma: PrismaService) {}

  @Get('live')
  live() { return { status: 'ok', timestamp: new Date().toISOString() }; }

  @Get('ready')
  async ready() {
    try {
      await this.prisma.$queryRaw`SELECT 1`;
    } catch (error) {
      throw new ServiceUnavailableException({ status: 'error', services: { database: 'error' } });
    }
    return {
      status: 'ok',
      timestamp: new Date().toISOString(),
      services: { database: 'ok' },
    };
  }
}
