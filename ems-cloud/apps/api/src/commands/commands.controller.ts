import { Body, Controller, ForbiddenException, Get, Param, Post, Req, UseGuards } from '@nestjs/common';
import { IsIn, IsObject, IsOptional } from 'class-validator';
import { Request } from 'express';
import { AuthGuard } from '../auth/auth.guard';
import { CommandsService } from './commands.service';

class CreateCommandDto {
  @IsIn(['identify', 'update_tariff', 'update_config', 'restart']) type!: string;
  @IsOptional() @IsObject() payload?: object;
}

@Controller()
@UseGuards(AuthGuard)
export class CommandsController {
  constructor(private readonly commands: CommandsService) {}
  @Post('sites/:siteId/commands') create(@Param('siteId') siteId: string, @Body() dto: CreateCommandDto, @Req() req: Request) {
    if (req.user!.role === 'VIEWER') throw new ForbiddenException('Viewer cannot issue commands');
    return this.commands.create(siteId, req.user!.tenantId, dto.type, dto.payload);
  }
  @Get('commands/:id') get(@Param('id') id: string, @Req() req: Request) { return this.commands.get(id, req.user!.tenantId); }
}
