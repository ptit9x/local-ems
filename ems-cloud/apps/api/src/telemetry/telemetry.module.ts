import { Module } from '@nestjs/common';
import { SitesModule } from '../sites/sites.module';
import { TelemetryService } from './telemetry.service';
import { TelemetryController } from './telemetry.controller';
import { MqttIngestionService } from './mqtt-ingestion.service';
import { FailoverModule } from '../failover/failover.module';

@Module({
  imports: [SitesModule, FailoverModule],
  providers: [TelemetryService, MqttIngestionService],
  controllers: [TelemetryController],
  exports: [TelemetryService, MqttIngestionService],
})
export class TelemetryModule {}
