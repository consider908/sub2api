-- Add group-level Kiro sticky session binding TTL.
ALTER TABLE groups
  ADD COLUMN IF NOT EXISTS kiro_sticky_session_ttl_seconds INTEGER NOT NULL DEFAULT 3600;

