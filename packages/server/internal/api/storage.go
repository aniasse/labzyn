package api

import (
	"context"

	"github.com/aniasse/labzyn/server/internal/metadata"
	"github.com/aniasse/labzyn/server/internal/records"
)

// Record storage dispatch: route to the JSONB repo (storage="jsonb") or the
// real-table store (storage="table"), keeping handlers storage-agnostic.

func isTableBacked(obj metadata.Object) bool { return obj.Storage == "table" }

func fieldNameSet(fields []metadata.Field) map[string]struct{} {
	m := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		m[f.Name] = struct{}{}
	}
	return m
}

func (s *Server) recList(ctx context.Context, obj metadata.Object, fields []metadata.Field, wsID string, opt records.ListOptions) ([]records.Record, int, error) {
	if isTableBacked(obj) {
		return s.tables.List(ctx, obj, fieldNameSet(fields), wsID, opt)
	}
	return s.rec.List(ctx, obj.ID, wsID, opt)
}

func (s *Server) recCreate(ctx context.Context, obj metadata.Object, wsID string, data map[string]any) (records.Record, error) {
	if isTableBacked(obj) {
		return s.tables.Create(ctx, obj, wsID, data)
	}
	return s.rec.Create(ctx, obj.ID, wsID, data)
}

func (s *Server) recGet(ctx context.Context, obj metadata.Object, wsID, id string) (records.Record, error) {
	if isTableBacked(obj) {
		return s.tables.Get(ctx, obj, wsID, id)
	}
	return s.rec.Get(ctx, obj.ID, wsID, id)
}

func (s *Server) recUpdate(ctx context.Context, obj metadata.Object, wsID, id string, patch map[string]any) (records.Record, error) {
	if isTableBacked(obj) {
		return s.tables.Update(ctx, obj, wsID, id, patch)
	}
	return s.rec.Update(ctx, obj.ID, wsID, id, patch)
}

func (s *Server) recDelete(ctx context.Context, obj metadata.Object, wsID, id string) error {
	if isTableBacked(obj) {
		return s.tables.SoftDelete(ctx, obj, wsID, id)
	}
	return s.rec.SoftDelete(ctx, obj.ID, wsID, id)
}
