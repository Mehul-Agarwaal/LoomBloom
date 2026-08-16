DROP TABLE IF EXISTS organization_sessions;
ALTER TABLE organizations 
DROP COLUMN IF EXISTS gst_number,
DROP COLUMN IF EXISTS owner_email,
DROP COLUMN IF EXISTS phone_verified,
DROP COLUMN IF EXISTS otp_code,
DROP COLUMN IF EXISTS otp_expires_at;
