import { Body, Controller, ForbiddenException, Get, Param, Post, Req, UseGuards } from '@nestjs/common';
import { z } from 'zod';
import { Request } from 'express';
import { AuthGuard } from '../../common/guards/auth.guard';
import { CommandsService } from './commands.service';
import { ZodPipe } from '../../common/pipes/zod.pipe';

const CreateCommandSchema = z.object({
  type: z.enum(['identify', 'update_tariff', 'update_config', 'restart']),
  payload: z.object({}).passthrough().optional(),
});
type CreateCommandDto = z.infer<typeof CreateCommandSchema>;

@Controller()
@UseGuards(AuthGuard)
export class CommandsController {
  constructor(private readonly commands: CommandsService) {}
  @Post('sites/:siteId/commands') create(@Param('siteId') siteId: string, @Body(new ZodPipe(CreateCommandSchema)) dto: CreateCommandDto, @Req() req: Request) {
    if (req.user!.role === 'VIEWER') throw new ForbiddenException('Viewer cannot issue commands');
    return this.commands.create(siteId, req.user!.tenantId, dto.type, dto.payload);
  }
  @Get('commands/:id') get(@Param('id') id: string, @Req() req: Request) { return this.commands.get(id, req.user!.tenantId); }
}
