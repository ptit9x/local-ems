import { Controller, Get, Param, Req, UseGuards } from '@nestjs/common';
import { AuthGuard } from '../auth/auth.guard';
import { FailoverService } from './failover.service';
import { CloudOcppService } from './cloud-ocpp.service';

@Controller('sites/:siteId/failover')
@UseGuards(AuthGuard)
export class FailoverController {
  constructor(private readonly failover: FailoverService, private readonly ocpp: CloudOcppService) {}

  @Get('chargers')
  chargers() { return this.ocpp.list(); }

  @Get()
  async current(@Param('siteId') siteId: string, @Req() req: any) {
    const value = await this.failover.currentForTenant(siteId, req.user.tenantId);
    if (!value) return null;
    if (value.siteId !== siteId) return null;
    return { ...value, leaseEpoch: value.leaseEpoch.toString() };
  }

  @Get('timeline')
  async timeline(@Param('siteId') siteId: string, @Req() req: any) {
    const events = await this.failover.timeline(siteId, req.user.tenantId);
    return events.map(event => ({ ...event, leaseEpoch: event.leaseEpoch.toString() }));
  }
}

