import { Module } from '@nestjs/common';
import { AuthController } from './auth.controller';
import { AuthGuard } from '../../common/guards/auth.guard';
import { AuthService } from './auth.service';

@Module({ controllers: [AuthController], providers: [AuthService, AuthGuard], exports: [AuthGuard] })
export class AuthModule {}
