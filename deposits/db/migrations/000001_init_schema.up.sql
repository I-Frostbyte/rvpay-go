CREATE TYPE "payer" AS (
    "type" VARCHAR(3) DEFAULT 'MMO'
    
)

CREATE TABLE "deposits" (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "amount" TEXT NOT NULL,
    "currency" VARCHAR(3) NOT NULL,
)