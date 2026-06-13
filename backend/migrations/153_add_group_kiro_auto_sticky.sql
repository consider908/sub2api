-- Add group-level Kiro automatic sticky routing switch.
ALTER TABLE groups
  ADD COLUMN IF NOT EXISTS kiro_auto_sticky_enabled BOOLEAN NOT NULL DEFAULT TRUE;

