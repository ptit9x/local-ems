import { NestFactory } from '@nestjs/core';
import { ValidationPipe } from '@nestjs/common';
import { AppModule } from './app.module';
import cookieParser from 'cookie-parser';
import { enableBigIntJson } from './json';

async function bootstrap() {
  // Prisma uses bigint identifiers for TimescaleDB-compatible composite keys.
  // Express delegates to JSON.stringify, which otherwise throws on bigint.
  enableBigIntJson();

  const app = await NestFactory.create(AppModule);
  app.use(cookieParser());

  // Global prefix
  app.setGlobalPrefix('api');

  // Global validation pipe
  app.useGlobalPipes(
    new ValidationPipe({
      whitelist: true,
      forbidNonWhitelisted: true,
      transform: true,
      transformOptions: { enableImplicitConversion: true },
    }),
  );

  // CORS
  app.enableCors({
    origin: process.env.CORS_ORIGIN || 'http://localhost:5173',
    methods: 'GET,HEAD,PUT,PATCH,POST,DELETE',
    credentials: true,
  });

  const port = process.env.API_PORT || 3000;
  await app.listen(port);

  console.log(`\n╔════════════════════════════════════════════════╗`);
  console.log(`║   EMS Cloud API listening on port ${port}          ║`);
  console.log(`║   Docs: http://localhost:${port}/api              ║`);
  console.log(`╚════════════════════════════════════════════════╝\n`);
}
bootstrap();
