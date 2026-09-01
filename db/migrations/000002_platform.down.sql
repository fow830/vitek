DROP TABLE IF EXISTS avito_accounts;
DROP TABLE IF EXISTS user_service_entitlements;
DROP TABLE IF EXISTS product_services;
DROP TABLE IF EXISTS magic_link_challenges;
ALTER TABLE proxies DROP COLUMN IF EXISTS label;
ALTER TABLE users DROP COLUMN IF EXISTS role;
DROP TYPE IF EXISTS avito_account_status;
DROP TYPE IF EXISTS user_role;
