package store

import (
	"context"
	"encoding/json"
	"strings"
)

func (s *Store) WriteAudit(ctx context.Context, event AuditEvent) error {
	if event.Result == "" {
		event.Result = "success"
	}
	metadata, err := json.Marshal(event.Metadata)
	if err != nil {
		return err
	}
	if event.Metadata == nil {
		metadata = []byte(`{}`)
	}
	var requestID any
	if strings.TrimSpace(event.RequestID) != "" {
		requestID = strings.TrimSpace(event.RequestID)
	}
	var ipAddress any
	if strings.TrimSpace(event.IPAddress) != "" {
		ipAddress = strings.TrimSpace(event.IPAddress)
	}
	_, err = s.DB.Exec(ctx, `INSERT INTO audit_logs(actor_user_id,organization_id,project_id,actor_role,result,request_id,ip_address,user_agent,action,entity_type,entity_id,metadata)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, event.ActorUserID, event.OrganizationID, event.ProjectID, event.ActorRole, event.Result, requestID, ipAddress, event.UserAgent, event.Action, event.EntityType, event.EntityID, metadata)
	return err
}
