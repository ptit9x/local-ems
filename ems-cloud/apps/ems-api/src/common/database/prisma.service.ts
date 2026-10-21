import { Injectable, OnModuleInit, OnModuleDestroy } from '@nestjs/common';
import { PrismaClient } from '@prisma/client';

@Injectable()
export class PrismaService
  extends PrismaClient
  implements OnModuleInit, OnModuleDestroy
{
  async onModuleInit() {
    await this.$connect();
  }

  async onModuleDestroy() {
    await this.$disconnect();
  }

  /**
   * Create TimescaleDB hypertable for telemetry_records.
   * Run this once after the first migration.
   */
  async createHypertable(): Promise<void> {
    try {
      await this.$executeRawUnsafe(`
        SELECT create_hypertable('telemetry_records', 'timestamp',
          if_not_exists => TRUE,
          migrate_data => TRUE
        );
      `);
      console.log('TimescaleDB: hypertable created for telemetry_records');
    } catch (error) {
      // Ignore if TimescaleDB extension is not available (dev without Docker)
      console.warn('TimescaleDB: could not create hypertable:', error);
    }
  }
}
