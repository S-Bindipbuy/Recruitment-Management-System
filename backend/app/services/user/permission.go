package user

import (
	"ems/app/enums"
	usermodel "ems/app/models/user"
	userrepo "ems/app/repositories/user"
)

// PermissionService manages permissions and role→permission mapping.
type PermissionService struct{}

func (s *PermissionService) Create(name, description string) (*usermodel.Permission, error) {
	p := &usermodel.Permission{Name: name, Description: description}
	if err := userrepo.PermissionRepo.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PermissionService) FindAll() ([]usermodel.Permission, error) {
	return userrepo.PermissionRepo.FindAll()
}

func (s *PermissionService) GetPermissionsByRole(roleID int64) ([]usermodel.Permission, error) {
	return userrepo.PermissionRepo.GetPermissionsByRole(roleID)
}

func (s *PermissionService) SetRolePermissions(roleID int64, permissionIDs []int64) error {
	return userrepo.PermissionRepo.SetRolePermissions(roleID, permissionIDs)
}

func (s *PermissionService) Delete(id int64) error {
	return userrepo.PermissionRepo.Delete(id)
}

func (s *PermissionService) CountRolesWithPermission(permissionID int64) (int, error) {
	return userrepo.PermissionRepo.CountRolesWithPermission(permissionID)
}

// PolicyService answers access-control questions. Handlers enforce policy
// decisions through PolicyService rather than grouping routes in middleware.
type PolicyService struct{}

func (s *PolicyService) CanAny(userID int64, permissions ...enums.Permission) bool {
	if userID <= 0 || len(permissions) == 0 {
		return false
	}
	userPermissions, err := userrepo.PermissionRepo.GetPermissionsForUser(userID)
	if err != nil {
		return false
	}
	has := make(map[string]bool, len(userPermissions))
	for _, p := range userPermissions {
		has[p] = true
	}
	for _, p := range permissions {
		if has[string(p)] {
			return true
		}
	}
	return false
}

func (s *PolicyService) PermissionsForUser(userID int64) []string {
	permissions, err := userrepo.PermissionRepo.GetPermissionsForUser(userID)
	if err != nil {
		return []string{}
	}
	return permissions
}

// IsAdmin reports whether the user holds a role named "admin" (case-insensitive).
// Role and permission management, plus user role assignment, are only allowed
// for administrators.
func (s *PolicyService) IsAdmin(userID int64) bool {
	isAdmin, err := userrepo.RoleRepo.IsAdminByUserID(userID)
	return err == nil && isAdmin
}
