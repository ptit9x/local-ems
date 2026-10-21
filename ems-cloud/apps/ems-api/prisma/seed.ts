import { PrismaClient, UserRole, DeviceType } from '@prisma/client';
import { hash } from 'bcryptjs';

const prisma = new PrismaClient();
async function main() {
  const passwordHash = await hash('demo1234', 12);
  for (const tenantDef of [{ slug: 'demo-a', name: 'Demo Tenant A', siteCode: 'cdi-singapore' }, { slug: 'demo-b', name: 'Demo Tenant B', siteCode: 'demo-b-site' }]) {
    const tenant = await prisma.tenant.upsert({ where: { slug: tenantDef.slug }, update: {}, create: { slug: tenantDef.slug, name: tenantDef.name } });
    const site = await prisma.site.upsert({ where: { siteCode: tenantDef.siteCode }, update: {}, create: { siteCode: tenantDef.siteCode, name: `${tenantDef.name} Simulator`, location: 'Simulator', timezone: 'Asia/Singapore', tenantId: tenant.id } });
    for (const [prefix, count, type] of [['bms',4,DeviceType.BMS],['pcs',24,DeviceType.PCS],['ev',6,DeviceType.EV_CHARGER],['meter',3,DeviceType.METER]] as const) {
      for (let i=0;i<count;i++) await prisma.device.upsert({ where: { siteId_externalId: { siteId: site.id, externalId: `${prefix}-${i}` } }, update: {}, create: { siteId: site.id, externalId: `${prefix}-${i}`, type, metadata: { simulated: true } } });
    }
    for (const role of [UserRole.ADMIN, UserRole.OPERATOR, UserRole.VIEWER]) {
      const email = `${role.toLowerCase()}@${tenantDef.slug === 'demo-a' ? 'demo.local' : 'demo-b.local'}`;
      await prisma.user.upsert({ where: { email }, update: { passwordHash }, create: { email, name: `${tenantDef.name} ${role}`, role, tenantId: tenant.id, passwordHash } });
    }
  }
}
main().finally(() => prisma.$disconnect());
