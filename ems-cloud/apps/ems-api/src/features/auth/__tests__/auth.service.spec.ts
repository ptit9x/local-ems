import { UnauthorizedException } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { hash } from 'bcryptjs';
import { AuthService } from '../auth.service';

describe('AuthService', () => {
  const prisma = {
    user: { findUnique: jest.fn() },
    refreshToken: {
      create: jest.fn(),
      findUnique: jest.fn(),
      update: jest.fn(),
      updateMany: jest.fn(),
    },
  };
  const config = { get: jest.fn((_key: string, fallback: string) => fallback) };
  let service: AuthService;

  beforeEach(() => {
    jest.clearAllMocks();
    prisma.refreshToken.create.mockResolvedValue({});
    service = new AuthService(prisma as never, config as unknown as ConfigService);
  });

  it('issues an access token for valid credentials without exposing the password hash', async () => {
    prisma.user.findUnique.mockResolvedValue({
      id: 'user-a', tenantId: 'tenant-a', email: 'admin@example.com', role: 'ADMIN',
      passwordHash: await hash('correct-password', 4),
    });

    const result = await service.login('admin@example.com', 'correct-password');

    expect(result.access_token).toBeTruthy();
    expect(result.refresh_token).toBeTruthy();
    expect(result.user).toEqual({ email: 'admin@example.com', role: 'ADMIN' });
    expect(result.user).not.toHaveProperty('passwordHash');
    expect(prisma.refreshToken.create).toHaveBeenCalledTimes(1);
  });

  it('rejects invalid credentials', async () => {
    prisma.user.findUnique.mockResolvedValue(null);
    await expect(service.login('missing@example.com', 'wrong')).rejects.toBeInstanceOf(UnauthorizedException);
  });
});
