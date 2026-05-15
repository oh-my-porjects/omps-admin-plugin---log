package main

// audit_storage.go — log 模块表初始化（task/inner_plugin.md §4.4）

import "context"

// ensureLogAuditTable 启动时建表（幂等）
func (p *LogPlugin) ensureLogAuditTable(ctx context.Context) error {
	_, err := p.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS log_audit (
			id            BIGSERIAL PRIMARY KEY,
			actor         TEXT NOT NULL,
			action        TEXT NOT NULL,
			module        TEXT NOT NULL DEFAULT '',
			before_state  JSONB NOT NULL DEFAULT '{}'::jsonb,
			after_state   JSONB NOT NULL DEFAULT '{}'::jsonb,
			extra         JSONB NOT NULL DEFAULT '{}'::jsonb,
			created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS log_audit_actor_idx ON log_audit (actor, created_at DESC);
		CREATE INDEX IF NOT EXISTS log_audit_action_idx ON log_audit (action, created_at DESC);
		CREATE INDEX IF NOT EXISTS log_audit_module_idx ON log_audit (module, created_at DESC);
	`)
	return err
}
