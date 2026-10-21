import { Body, Controller, Post, Req, Res, UnauthorizedException } from '@nestjs/common';
import { z } from 'zod';
import { Request, Response } from 'express';
import { AuthService } from './auth.service';
import { ZodPipe } from '../../common/pipes/zod.pipe';

const LoginSchema = z.object({
  email: z.string().email(),
  password: z.string().min(1),
});
type LoginDto = z.infer<typeof LoginSchema>;

@Controller('auth')
export class AuthController {
  constructor(private readonly auth: AuthService) {}
  @Post('login') async login(@Body(new ZodPipe(LoginSchema)) dto: LoginDto, @Res({ passthrough: true }) res: Response) { return this.withCookie(await this.auth.login(dto.email, dto.password), res); }
  @Post('refresh') async refresh(@Req() req: Request, @Res({ passthrough: true }) res: Response) { return this.withCookie(await this.auth.refresh(this.cookie(req)), res); }
  @Post('logout') async logout(@Req() req: Request, @Res({ passthrough: true }) res: Response) { const result = await this.auth.logout(this.cookie(req)); res.clearCookie('ems_refresh'); return result; }

  private cookie(req: Request) { const value = req.cookies?.ems_refresh; if (!value) throw new UnauthorizedException('Refresh cookie required'); return value; }
  private withCookie(tokens: { access_token: string; refresh_token: string; user: object }, res: Response) {
    res.cookie('ems_refresh', tokens.refresh_token, { httpOnly: true, sameSite: 'lax', secure: process.env.NODE_ENV === 'production', maxAge: 7 * 86400000 });
    return { access_token: tokens.access_token, user: tokens.user };
  }
}
