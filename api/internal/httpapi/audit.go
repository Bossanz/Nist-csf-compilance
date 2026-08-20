package httpapi

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"

	"compliance/api/internal/store"
	"github.com/google/uuid"
)

type auditDataStore interface {
	WriteAudit(context.Context, store.AuditEvent) error
}

type auditRequestContextKey struct{}

type auditRequestContext struct {
	RequestID string
	IPAddress string
	UserAgent string
}

func withAuditRequestContext(r *http.Request) *http.Request {
	requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if requestID == "" {
		requestID = uuid.NewString()
	}
	if len(requestID) > 128 {
		requestID = requestID[:128]
	}
	ipAddress := strings.TrimSpace(r.RemoteAddr)
	if host, _, err := net.SplitHostPort(ipAddress); err == nil {
		ipAddress = host
	}
	userAgent := r.UserAgent()
	if len(userAgent) > 512 {
		userAgent = userAgent[:512]
	}
	return r.WithContext(context.WithValue(r.Context(), auditRequestContextKey{}, auditRequestContext{RequestID: requestID, IPAddress: ipAddress, UserAgent: userAgent}))
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
	event.ActorRole = rUser.Role
	if event.Result == "" {
		event.Result = "success"
	}
	if request, ok := ctx.Value(auditRequestContextKey{}).(auditRequestContext); ok {
		if event.RequestID == "" {
			event.RequestID = request.RequestID
		}
		if event.IPAddress == "" {
			event.IPAddress = request.IPAddress
		}
		if event.UserAgent == "" {
			event.UserAgent = request.UserAgent
		}
	}
	if err := auditor.WriteAudit(ctx, event); err != nil {
		log.Printf("audit write failed: %v", err)
	}
}
