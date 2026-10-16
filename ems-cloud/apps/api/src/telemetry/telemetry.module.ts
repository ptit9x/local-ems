import { Module } from '@nestjs/common';
import { SitesModule } from '../sites/sites.module';
import { TelemetryService } from './telemetry.service';
import { TelemetryController } from './telemetry.controller';
import { MqttIngestionService } from './mqtt-ingestion.service';

@Module({
  imports: [SitesModule],
  providers: [TelemetryService, MqttIngestionService],
  controllers: [TelemetryController],
  exports: [TelemetryService, MqttIngestionService],
})
export class TelemetryModule {}
