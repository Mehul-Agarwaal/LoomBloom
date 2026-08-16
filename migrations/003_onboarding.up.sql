-- Alter organizations table to add GST, owner email, and OTP state fields
ALTER TABLE organizations 
ADD COLUMN IF NOT EXISTS gst_number TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS owner_email TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS phone_verified BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS otp_code TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS otp_expires_at TIMESTAMPTZ;

-- Create organization sessions table for authentication
CREATE TABLE IF NOT EXISTS organization_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create index for session token lookup
CREATE INDEX IF NOT EXISTS idx_sessions_token ON organization_sessions(token);
