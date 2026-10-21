import { Module } from '@nestjs/common';
import { FailoverController } from './failover.controller';
import { FailoverService } from './failover.service';
import { CloudOcppService } from './cloud-ocpp.service';

@Module({ providers: [FailoverService, CloudOcppService], controllers: [FailoverController], exports: [FailoverService, CloudOcppService] })
export class FailoverModule {}

