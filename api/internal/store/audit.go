package store

import (
	"context"
	"encoding/json"
)

func (s *Store) WriteAudit(ctx context.Context, event AuditEvent) error {
	metadata, err := json.Marshal(event.Metadata)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(ctx, `INSERT INTO audit_logs(actor_user_id,organization_id,project_id,action,entity_type,entity_id,metadata)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`, event.ActorUserID, event.OrganizationID, event.ProjectID, event.Action, event.EntityType, event.EntityID, metadata)
	return err
}
