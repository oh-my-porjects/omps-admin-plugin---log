package main

// audit_handlers.go — log 公共模块的审计日志接口（task/inner_plugin.md §4.4）
//
// 提供：
//   POST /api/log/audit  写入一条审计日志
//   GET  /api/log/list   列查（按 actor / action / since / until 过滤）

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type auditWriteReq struct {
	Actor   string `json:"actor"`
	Action  string `json:"action"`
	Before  any    `json:"before,omitempty"`
	After   any    `json:"after,omitempty"`
	Extra   any    `json:"extra,omitempty"`
	Module  string `json:"module,omitempty"`
}

// handleAuditWrite POST /api/log/audit
func (p *LogPlugin) handleAuditWrite(w http.ResponseWriter, r *http.Request) {
	var req auditWriteReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, -1, nil, "解析失败: "+err.Error())
		return
	}
	if strings.TrimSpace(req.Actor) == "" || strings.TrimSpace(req.Action) == "" {
		writeJSON(w, -1, nil, "actor / action 必填")
		return
	}
	beforeJSON, _ := json.Marshal(req.Before)
	afterJSON, _ := json.Marshal(req.After)
	extraJSON, _ := json.Marshal(req.Extra)

	_, err := p.db.ExecContext(r.Context(), `
		INSERT INTO log_audit (actor, action, module, before_state, after_state, extra)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, req.Actor, req.Action, req.Module, string(beforeJSON), string(afterJSON), string(extraJSON))
	if err != nil {
		writeJSON(w, -1, nil, "写入失败: "+err.Error())
		return
	}
	writeJSON(w, 0, map[string]any{"ok": true}, "")
}

// handleAuditList GET /api/log/list?actor=&action=&since=&until=&page=&page_size=
func (p *LogPlugin) handleAuditList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	actor := q.Get("actor")
	action := q.Get("action")
	module := q.Get("module")
	since := q.Get("since")
	until := q.Get("until")
	page, _ := strconv.Atoi(q.Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 50
	}

	whereParts := []string{"1=1"}
	args := []any{}
	idx := 1
	if actor != "" {
		whereParts = append(whereParts, fmt.Sprintf("actor = $%d", idx))
		args = append(args, actor)
		idx++
	}
	if action != "" {
		whereParts = append(whereParts, fmt.Sprintf("action = $%d", idx))
		args = append(args, action)
		idx++
	}
	if module != "" {
		whereParts = append(whereParts, fmt.Sprintf("module = $%d", idx))
		args = append(args, module)
		idx++
	}
	if since != "" {
		whereParts = append(whereParts, fmt.Sprintf("created_at >= $%d", idx))
		args = append(args, since)
		idx++
	}
	if until != "" {
		whereParts = append(whereParts, fmt.Sprintf("created_at <= $%d", idx))
		args = append(args, until)
		idx++
	}
	whereSQL := strings.Join(whereParts, " AND ")

	var total int
	if err := p.db.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM log_audit WHERE "+whereSQL, args...).Scan(&total); err != nil {
		writeJSON(w, -1, nil, "统计失败: "+err.Error())
		return
	}

	offset := (page - 1) * pageSize
	rows, err := p.db.QueryContext(r.Context(),
		"SELECT id, actor, action, module, before_state, after_state, extra, created_at "+
			"FROM log_audit WHERE "+whereSQL+
			fmt.Sprintf(" ORDER BY id DESC LIMIT $%d OFFSET $%d", idx, idx+1),
		append(args, pageSize, offset)...)
	if err != nil {
		writeJSON(w, -1, nil, "查询失败: "+err.Error())
		return
	}
	defer rows.Close()

	items := []map[string]any{}
	for rows.Next() {
		var id int64
		var a, act, mod, before, after, extra, createdAt string
		if err := rows.Scan(&id, &a, &act, &mod, &before, &after, &extra, &createdAt); err == nil {
			items = append(items, map[string]any{
				"id":         id,
				"actor":      a,
				"action":     act,
				"module":     mod,
				"before":     json.RawMessage(before),
				"after":      json.RawMessage(after),
				"extra":      json.RawMessage(extra),
				"created_at": createdAt,
			})
		}
	}
	writeJSON(w, 0, map[string]any{"total": total, "items": items}, "")
}
