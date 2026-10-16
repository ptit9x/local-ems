import { UserRole } from '@prisma/client';

export interface AuthUser {
  sub: string;
  tenantId: string;
  email: string;
  role: UserRole;
}

declare global {
  namespace Express {
    interface Request { user?: AuthUser }
  }
}
