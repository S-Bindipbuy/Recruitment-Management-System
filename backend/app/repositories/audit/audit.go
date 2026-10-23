package audit

import (
	"ems/app/models"
	"ems/database"
)

type AuditLogRepository struct{}

func (r *AuditLogRepository) FindAll(limit, offset int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := database.DB.Select(&logs,
		`SELECT id, user_id, action, entity_type, entity_id,
		 COALESCE(old_values, '') AS old_values, COALESCE(new_values, '') AS new_values,
		 COALESCE(ip_address, '') AS ip_address, COALESCE(user_agent, '') AS user_agent, created_at
		 FROM audit_logs ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	return logs, nil
}
