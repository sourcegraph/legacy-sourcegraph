-- This migration was generated by the command `sg telemetry remove`
INSERT INTO event_logs_export_allowlist (event_name) VALUES (UNNEST('{SearchSubmitted}'::TEXT[])) ON CONFLICT DO NOTHING;