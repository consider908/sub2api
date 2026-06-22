ALTER TABLE usage_logs
ADD COLUMN IF NOT EXISTS kiro_credits numeric NULL;
