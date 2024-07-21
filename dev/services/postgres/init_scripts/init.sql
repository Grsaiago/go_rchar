CREATE TABLE "Users" (
  "uuid" UUID UNIQUE PRIMARY KEY NOT NULL DEFAULT (gen_random_uuid()),
  "username" VARCHAR(70) NOT NULL,
  "password" VARCHAR NOT NULL,
  "created_at" DATE NOT NULL DEFAULT (NOW()),
  "updated_at" DATE NOT NULL DEFAULT (NOW())
);
