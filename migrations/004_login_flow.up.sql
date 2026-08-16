-- Add indexes on owner_email and phone for login lookups
CREATE INDEX IF NOT EXISTS idx_organizations_owner_email ON organizations(owner_email);
CREATE INDEX IF NOT EXISTS idx_organizations_phone ON organizations(phone);
