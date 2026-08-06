package httpapi

import (
	"context"
	"log"

	"compliance/api/internal/store"
)

type auditDataStore interface {
	WriteAudit(context.Context, store.AuditEvent) error
}

func (h *Handler) writeAudit(rUser store.User, ctx context.Context, event store.AuditEvent) {
	if h.Auth == nil {
		return
	}
	auditor, ok := h.Store.(auditDataStore)
	if !ok {
		return
	}
	event.ActorUserID = rUser.ID
	if err := auditor.WriteAudit(ctx, event); err != nil {
		log.Printf("audit write failed: %v", err)
	}
}
