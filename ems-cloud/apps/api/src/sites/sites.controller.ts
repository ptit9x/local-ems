import { Controller, Get, NotFoundException, Param, Post, Req, UseGuards } from '@nestjs/common';
import { Request } from 'express';
import { AuthGuard } from '../auth/auth.guard';
import { SitesService } from './sites.service';

@Controller('sites')
@UseGuards(AuthGuard)
export class SitesController {
  constructor(private readonly sitesService: SitesService) {}

  @Get()
  async findAll(@Req() req: Request) {
    return this.sitesService.findAll(req.user!.tenantId);
  }

  @Get(':id')
  async findOne(@Param('id') id: string, @Req() req: Request) {
    const site = await this.sitesService.findById(id, req.user!.tenantId);
    if (!site) throw new NotFoundException('Site not found');
    return site;
  }

  @Get(':id/devices')
  async devices(@Param('id') id: string, @Req() req: Request) {
    const devices = await this.sitesService.devices(id, req.user!.tenantId);
    if (!devices) throw new NotFoundException('Site not found');
    return devices;
  }

  @Get(':id/alerts')
  async alerts(@Param('id') id: string, @Req() req: Request) {
    const alerts = await this.sitesService.alerts(id, req.user!.tenantId);
    if (!alerts) throw new NotFoundException('Site not found');
    return alerts;
  }

  @Post(':id/alerts/:alertId/acknowledge')
  acknowledge(@Param('id') id: string, @Param('alertId') alertId: string, @Req() req: Request) { return this.sitesService.acknowledgeAlert(id, alertId, req.user!.tenantId); }
}
