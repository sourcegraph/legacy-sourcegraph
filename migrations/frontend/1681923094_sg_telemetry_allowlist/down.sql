-- This migration was generated by the command `sg telemetry add`
DELETE FROM event_logs_export_allowlist WHERE event_name IN (SELECT * FROM UNNEST('{CodyVSCodeExtension:recipe:rewrite-to-functional:executed,CodyVSCodeExtension:recipe:improve-variable-names:executed,CodyVSCodeExtension:recipe:replace:executed,CodyVSCodeExtension:recipe:generate-docstring:executed,CodyVSCodeExtension:recipe:generate-unit-test:executed,CodyVSCodeExtension:recipe:rewrite-functional:executed,CodyVSCodeExtension:recipe:code-refactor:executed,CodyVSCodeExtension:recipe:fixup:executed,CodyVSCodeExtension:recipe:explain-code-high-level:executed,CodyVSCodeExtension:recipe:explain-code-detailed:executed,CodyVSCodeExtension:recipe:find-code-smells:executed,CodyVSCodeExtension:recipe:git-history:executed,CodyVSCodeExtension:recipe:rate-code:executed,CodyVSCodeExtension:recipe:chat-question:executed,CodyVSCodeExtension:recipe:translate-to-language:executed}'::TEXT[]));
