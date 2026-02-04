import { Module } from '@nestjs/common';
import { NestFactory } from '@nestjs/core';

// Temporary bootstrap module; expand with controllers/services as features land.
@Module({})
class AppModule {}

async function bootstrap() {
  const app = await NestFactory.create(AppModule, { bufferLogs: true });

  // Configure global API settings (CORS, pipes) once requirements are settled.
  await app.listen(process.env.PORT ? Number(process.env.PORT) : 3000);
}

void bootstrap();
