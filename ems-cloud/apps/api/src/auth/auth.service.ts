import { Injectable, UnauthorizedException } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { compare } from 'bcryptjs';
import { createHash, randomUUID } from 'crypto';
import { sign, verify } from 'jsonwebtoken';
import { PrismaService } from '../prisma/prisma.service';
import { AuthUser } from './auth.types';

@Injectable()
export class AuthService {
  constructor(private readonly prisma: PrismaService, private readonly config: ConfigService) {}

  private secret() { return this.config.get('JWT_SECRET', 'dev-only-change-me'); }
  private hash(value: string) { return createHash('sha256').update(value).digest('hex'); }

  async login(email: string, password: string) {
    const user = await this.prisma.user.findUnique({ where: { email } });
    if (!user || !(await compare(password, user.passwordHash))) throw new UnauthorizedException('Invalid credentials');
    return this.issue({ sub: user.id, tenantId: user.tenantId, email: user.email, role: user.role });
  }

  async refresh(refreshToken: string) {
    let payload: AuthUser;
    try { payload = verify(refreshToken, this.secret()) as AuthUser; } catch { throw new UnauthorizedException('Invalid refresh token'); }
    const stored = await this.prisma.refreshToken.findUnique({ where: { tokenHash: this.hash(refreshToken) } });
    if (!stored || stored.revokedAt || stored.expiresAt < new Date()) throw new UnauthorizedException('Refresh token expired');
    await this.prisma.refreshToken.update({ where: { id: stored.id }, data: { revokedAt: new Date() } });
    return this.issue(payload);
  }

  async logout(refreshToken: string) {
    await this.prisma.refreshToken.updateMany({ where: { tokenHash: this.hash(refreshToken), revokedAt: null }, data: { revokedAt: new Date() } });
    return { ok: true };
  }

  private async issue(user: AuthUser) {
    const accessToken = sign(user, this.secret(), { expiresIn: '15m' });
    const refreshToken = sign({ ...user, jti: randomUUID() }, this.secret(), { expiresIn: '7d' });
    await this.prisma.refreshToken.create({ data: { tokenHash: this.hash(refreshToken), userId: user.sub, expiresAt: new Date(Date.now() + 7 * 86400000) } });
    return { access_token: accessToken, refresh_token: refreshToken, user: { email: user.email, role: user.role } };
  }
}
