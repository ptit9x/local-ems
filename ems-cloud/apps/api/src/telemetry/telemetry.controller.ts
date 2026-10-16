import { Controller, Get, Param, Query, Req, UseGuards } from '@nestjs/common';
import { Request } from 'express';
import { AuthGuard } from '../auth/auth.guard';
import { TelemetryService } from './telemetry.service';
import { TelemetryQueryDto } from './dto/telemetry-query.dto';

@Controller('sites/:siteId/telemetry')
@UseGuards(AuthGuard)
export class TelemetryController {
  constructor(private readonly telemetryService: TelemetryService) {}

  /**
   * GET /api/telemetry/:siteId?from=2025-01-01&to=2025-01-02&limit=100
   */
  @Get()
  async query(
    @Param('siteId') siteId: string,
    @Query() dto: TelemetryQueryDto,
    @Req() req: Request,
  ) {
    return this.telemetryService.query(siteId, req.user!.tenantId, dto);
  }

  /**
   * GET /api/telemetry/:siteId/latest
   */
  @Get('latest')
  async latest(@Param('siteId') siteId: string, @Req() req: Request) {
    return this.telemetryService.latest(siteId, req.user!.tenantId);
  }

  /**
   * GET /api/telemetry/:siteId/daily/:date
   */
  @Get('daily/:date')
  async dailySummary(
    @Param('siteId') siteId: string,
    @Param('date') date: string,
    @Req() req: Request,
  ) {
    return this.telemetryService.dailySummary(siteId, req.user!.tenantId, date);
  }
}
