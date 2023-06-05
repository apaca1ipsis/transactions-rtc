CREATE TABLE "transactions" (
	"id" INTEGER NOT NULL DEFAULT 'nextval(''transactions_id_seq''::regclass)',
	"amount" INTEGER NOT NULL,
	"account_from_id" INTEGER NOT NULL,
	"account_to_id" INTEGER NOT NULL,
	PRIMARY KEY ("id")
)
;
COMMENT ON COLUMN "transactions"."id" IS '';
COMMENT ON COLUMN "transactions"."amount" IS '';
COMMENT ON COLUMN "transactions"."account_from_id" IS '';
COMMENT ON COLUMN "transactions"."account_to_id" IS '';


INSERT INTO "transactions" ("id", "amount", "account_from_id", "account_to_id") VALUES
	(1, -100, 1, 2),
