// prisma/schema.prisma
generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

model User {
  id         String   @id @default(uuid())
  email      String   @unique
  appleSub   String?  @unique
  googleSub  String?  @unique
  createdAt  DateTime @default(now())

  projects   Project[] @relation("OwnerProjects")
  memberships Membership[]
  deviceTokens DeviceToken[]
}

model Project {
  id         String   @id @default(uuid())
  ownerId    String
  owner      User     @relation("OwnerProjects", fields: [ownerId], references: [id])
  name       String
  updatedAt  DateTime @updatedAt
  deletedAt  DateTime? // soft delete
  version    BigInt    // set by server on write

  tasks       Task[]
  memberships Membership[]
}

model Task {
  id           String   @id @default(uuid())
  projectId    String
  project      Project  @relation(fields: [projectId], references: [id])
  title        String
  notes        String?
  dueAt        DateTime?
  completedAt  DateTime?
  priority     Int       @default(0)
  order        Decimal   @default(0) // or BigInt if you prefer; decimal helps "between" inserts
  updatedAt    DateTime  @updatedAt
  deletedAt    DateTime?
  version      BigInt    // set by server on write
}

model Membership {
  // (userId, projectId) composite PK
  userId    String
  projectId String
  role      Role       @default(EDITOR)
  createdAt DateTime   @default(now())

  user      User       @relation(fields: [userId], references: [id])
  project   Project    @relation(fields: [projectId], references: [id])

  @@id([userId, projectId])
}

model DeviceToken {
  id        String   @id @default(uuid())
  userId    String
  user      User     @relation(fields: [userId], references: [id])
  deviceId  String
  platform  Platform
  token     String   @unique
  updatedAt DateTime @updatedAt

  @@unique([userId, deviceId, platform])
}

model SyncCursor {
  id         String   @id @default(uuid())
  userId     String
  deviceId   String
  lastVersion BigInt   @default(0)
  updatedAt  DateTime @updatedAt

  user      User      @relation(fields: [userId], references: [id])
}

enum Platform {
  IOS
  MACOS
  WINDOWS
}

enum Role {
  OWNER
  EDITOR
  READER
}

// Helpful indexes
// Run additional raw SQL migrations to add partial indexes on deletedAt IS NULL if needed.
