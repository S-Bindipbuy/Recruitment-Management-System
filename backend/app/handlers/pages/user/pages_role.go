package user

import (
	core "ems/app/handlers/pages/core"
	usermodel "ems/app/models/user"
	userrepo "ems/app/repositories/user"
	usersvc "ems/app/services/user"
	"github.com/gofiber/fiber/v2"
	"sort"
	"strconv"
	"strings"
)

// PermissionLevel is a preset of permissions inside a group, shown in the role
// picker as a single selectable "access level" (e.g. View only, Full control).
type PermissionLevel struct {
	Key             string
	Label           string
	PermissionNames []string
	PermissionIDs   []int64
}

// PermissionGroup buckets permissions into labelled feature areas. Each group
// carries a Description plus preset levels so a role can be granted access to
// a whole feature at once instead of ticking individual raw permissions.
type PermissionGroup struct {
	Key         string
	Label       string
	Description string
	Permissions []usermodel.Permission
	Levels      []PermissionLevel
}

// permissionGroupDefs declares each feature area and its available access
// levels. Level keys are stable identifiers used by the picker; "none" and
// "full" are added automatically.
var permissionGroupDefs = []struct {
	Key, Label, Description string
	Match                   func(string) bool
	Levels                  []PermissionLevel
}{
	{
		Key: "overview", Label: "General", Description: "Dashboard overview access.",
		Match: func(n string) bool { return strings.HasPrefix(n, "dashboard.") },
		Levels: []PermissionLevel{
			{Key: "view", Label: "View only", PermissionNames: []string{"dashboard.view"}},
		},
	},
	{
		Key: "organization", Label: "Organization", Description: "Employees, departments and positions.",
		Match: func(n string) bool { return strings.HasPrefix(n, "organization.") || strings.HasPrefix(n, "employee.") },
		Levels: []PermissionLevel{
			{Key: "manage", Label: "Manage", PermissionNames: []string{"employee.manage", "organization.manage"}},
		},
	},
	{
		Key: "recruitment", Label: "Recruitment", Description: "Job postings, interviews and candidates.",
		Match: func(n string) bool { return strings.HasPrefix(n, "recruitment.") },
		Levels: []PermissionLevel{
			{Key: "manage", Label: "Full control", PermissionNames: []string{"recruitment.manage"}},
		},
	},
	{
		Key: "evaluation", Label: "Evaluation", Description: "Run, assign and review performance evaluations.",
		Match: func(n string) bool { return strings.HasPrefix(n, "evaluation.") },
		Levels: []PermissionLevel{
			{Key: "worker", Label: "Evaluate (fill my assigned)", PermissionNames: []string{"evaluation.evaluate"}},
			{Key: "manager", Label: "Assign and manage", PermissionNames: []string{"evaluation.assign", "evaluation.approve", "evaluation.evaluate"}},
		},
	},
	{
		Key: "analytics", Label: "Analytics", Description: "Reports and audit logs.",
		Match: func(n string) bool { return strings.HasPrefix(n, "report.") || strings.HasPrefix(n, "audit.") },
		Levels: []PermissionLevel{
			{Key: "view", Label: "View only", PermissionNames: []string{"report.view", "audit.view"}},
		},
	},
	{
		Key: "administration", Label: "Administration", Description: "Users and role management.",
		Match: func(n string) bool { return strings.HasPrefix(n, "user.") || strings.HasPrefix(n, "role.") },
		Levels: []PermissionLevel{
			{Key: "manage", Label: "Full control", PermissionNames: []string{"user.manage", "role.manage"}},
		},
	},
}

func groupPermissions(permissions []usermodel.Permission) []PermissionGroup {
	byName := make(map[string]usermodel.Permission, len(permissions))
	for _, p := range permissions {
		byName[p.Name] = p
	}

	var groups []PermissionGroup
	for _, def := range permissionGroupDefs {
		g := PermissionGroup{Key: def.Key, Label: def.Label, Description: def.Description}
		var fullIDs []int64
		for _, p := range permissions {
			if def.Match(p.Name) {
				g.Permissions = append(g.Permissions, p)
				fullIDs = append(fullIDs, p.ID)
			}
		}
		if len(g.Permissions) == 0 {
			continue
		}

		g.Levels = append(g.Levels, PermissionLevel{Key: "none", Label: "No access"})
		for _, l := range def.Levels {
			resolved := PermissionLevel{Key: l.Key, Label: l.Label}
			for _, name := range l.PermissionNames {
				if p, ok := byName[name]; ok {
					resolved.PermissionIDs = append(resolved.PermissionIDs, p.ID)
				}
			}
			g.Levels = append(g.Levels, resolved)
		}
		if !hasIDSet(g.Levels, fullIDs) {
			g.Levels = append(g.Levels, PermissionLevel{Key: "full", Label: "Full control", PermissionIDs: fullIDs})
		}

		groups = append(groups, g)
	}
	return groups
}

// hasIDSet reports whether any level in the given list grants exactly the
// supplied permission set, so a redundant "Full control" preset is not added
// twice under a different label.
func hasIDSet(levels []PermissionLevel, ids []int64) bool {
	norm := func(v []int64) []int64 {
		s := append([]int64(nil), v...)
		sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
		return s
	}
	target := norm(ids)
	for _, l := range levels {
		got := norm(l.PermissionIDs)
		if len(got) != len(target) {
			continue
		}
		same := true
		for i := range got {
			if got[i] != target[i] {
				same = false
				break
			}
		}
		if same {
			return true
		}
	}
	return false
}

func (h *UserHandler) Roles(c *fiber.Ctx) error {
	if err := core.RequireAdmin(c); err != nil {
		return err
	}

	roles, _ := usersvc.RoleServiceInstance.FindAll()
	permissions, _ := usersvc.PermissionServiceInstance.FindAll()

	type RoleWithCount struct {
		usermodel.Role
		UserCount  int
		Permission []string
	}
	var rolesWithCount []RoleWithCount
	for _, r := range roles {
		count, _ := countUsersWithRole(r.ID)
		rolePermissions, _ := usersvc.PermissionServiceInstance.GetPermissionsByRole(r.ID)
		var perms []string
		for _, p := range rolePermissions {
			perms = append(perms, p.Name)
		}
		rolesWithCount = append(rolesWithCount, RoleWithCount{Role: r, UserCount: count, Permission: perms})
	}

	data := struct {
		Roles            []RoleWithCount
		Permissions      []usermodel.Permission
		PermissionGroups []PermissionGroup
	}{
		Roles:            rolesWithCount,
		Permissions:      permissions,
		PermissionGroups: groupPermissions(permissions),
	}
	return core.RenderPage(c, "page-roles", data)
}

func (h *UserHandler) CreateRole(c *fiber.Ctx) error {
	if err := core.RequireAdmin(c); err != nil {
		return err
	}

	name := strings.TrimSpace(c.FormValue("name"))
	description := strings.TrimSpace(c.FormValue("description"))

	if name == "" {
		return core.RenderError(c, fiber.StatusBadRequest, "Role name is required.")
	}

	_, err := usersvc.RoleServiceInstance.Create(name, description)
	if err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to create role. Name may already exist.")
	}

	return h.Roles(c)
}

func (h *UserHandler) UpdateRole(c *fiber.Ctx) error {
	if err := core.RequireAdmin(c); err != nil {
		return err
	}

	id, _ := c.ParamsInt("id")
	name := strings.TrimSpace(c.FormValue("name"))
	description := strings.TrimSpace(c.FormValue("description"))

	if id <= 0 || name == "" {
		return core.RenderError(c, fiber.StatusBadRequest, "Role name is required.")
	}

	_, err := usersvc.RoleServiceInstance.Update(int64(id), name, description)
	if err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to update role.")
	}

	return h.Roles(c)
}

func (h *UserHandler) DeleteRole(c *fiber.Ctx) error {
	if err := core.RequireAdmin(c); err != nil {
		return err
	}

	id, _ := c.ParamsInt("id")
	if id <= 0 {
		return core.RenderError(c, fiber.StatusBadRequest, "Invalid role ID.")
	}

	if err := usersvc.RoleServiceInstance.Delete(int64(id)); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to delete role.")
	}

	return h.Roles(c)
}

func (h *UserHandler) SetRolePermissions(c *fiber.Ctx) error {
	if err := core.RequireAdmin(c); err != nil {
		return err
	}

	id, _ := c.ParamsInt("id")
	if id <= 0 {
		return core.RenderError(c, fiber.StatusBadRequest, "Invalid role ID.")
	}

	permissionValues := c.Context().PostArgs().PeekMulti("permission_id")
	var ids []int64
	for _, v := range permissionValues {
		if pid, err := strconv.ParseInt(string(v), 10, 64); err == nil && pid > 0 {
			ids = append(ids, pid)
		}
	}

	if err := usersvc.PermissionServiceInstance.SetRolePermissions(int64(id), ids); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to update role permissions.")
	}

	return h.Roles(c)
}

func countUsersWithRole(roleID int64) (int, error) {
	return userrepo.RoleRepo.CountUsersWithRole(roleID)
}
