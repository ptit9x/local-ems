import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { PrismaModule } from './common/database/prisma.module';
import { TelemetryModule } from './features/telemetry/telemetry.module';
import { SitesModule } from './features/sites/sites.module';
import { HealthModule } from './features/health/health.module';
import { AuthModule } from './features/auth/auth.module';
import { CommandsModule } from './features/commands/commands.module';
import { FailoverModule } from './features/failover/failover.module';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      envFilePath: ['.env', '../../.env'],
    }),
    PrismaModule,
    HealthModule,
    AuthModule,
    CommandsModule,
    SitesModule,
    TelemetryModule,
    FailoverModule,
  ],
})
export class AppModule {}
