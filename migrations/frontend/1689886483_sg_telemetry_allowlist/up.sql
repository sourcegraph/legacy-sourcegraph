-- This migration was generated by the command `sg telemetry add`
INSERT INTO event_logs_export_allowlist (event_name) VALUES (UNNEST('{SIMPLE_SEARCH_SUBMIT_SEARCH,SIMPLE_SEARCH_SELECT_JOB,SIMPLE_SEARCH_BACK_BUTTON_CLICK,SIMPLE_SEARCH_SIMPLE_SEARCH_TOGGLE,VIEW_SIMPLE_SEARCH_HOME}'::TEXT[])) ON CONFLICT DO NOTHING;
