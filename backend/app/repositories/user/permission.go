package user

import (
	usermodel "ems/app/models/user"
	"ems/database"
)

type PermissionRepository struct{}

func (r *PermissionRepository) Create(permission *usermodel.Permission) error {
	result, err := database.DB.Exec(
		"INSERT INTO permissions (name, description) VALUES (?, ?)",
		permission.Name, permission.Description,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	permission.ID = id
	return nil
}

func (r *PermissionRepository) FindAll() ([]usermodel.Permission, error) {
	var permissions []usermodel.Permission
	err := database.DB.Select(&permissions,
		"SELECT id, name, COALESCE(description, '') AS description, created_at, updated_at FROM permissions ORDER BY name",
	)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *PermissionRepository) Delete(id int64) error {
	_, err := database.DB.Exec("DELETE FROM permissions WHERE id = ?", id)
	return err
}

func (r *PermissionRepository) CountRolesWithPermission(permissionID int64) (int, error) {
	var count int
	err := database.DB.Get(&count,
		"SELECT COUNT(*) FROM role_permissions WHERE permission_id = ?", permissionID,
	)
	return count, err
}

func (r *PermissionRepository) SetRolePermissions(roleID int64, permissionIDs []int64) error {
	tx, err := database.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM role_permissions WHERE role_id = ?", roleID); err != nil {
		return err
	}

	for _, permissionID := range permissionIDs {
		if _, err := tx.Exec(
			"INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)",
			roleID, permissionID,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PermissionRepository) GetPermissionsByRole(roleID int64) ([]usermodel.Permission, error) {
	var permissions []usermodel.Permission
	err := database.DB.Select(&permissions,
		`SELECT p.id, p.name, COALESCE(p.description, '') AS description, p.created_at, p.updated_at
		 FROM permissions p
		 INNER JOIN role_permissions rp ON rp.permission_id = p.id
		 WHERE rp.role_id = ? ORDER BY p.name`, roleID,
	)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *PermissionRepository) GetPermissionsForUser(userID int64) ([]string, error) {
	var permissions []string
	err := database.DB.Select(&permissions,
		`SELECT DISTINCT p.name
		 FROM permissions p
		 INNER JOIN role_permissions rp ON rp.permission_id = p.id
		 INNER JOIN user_roles ur ON ur.role_id = rp.role_id
		 WHERE ur.user_id = ? ORDER BY p.name`, userID,
	)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}
