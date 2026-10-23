package user

import (
	core "ems/app/handlers/pages/core"
	usermodel "ems/app/models/user"
	usersvc "ems/app/services/user"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct{}

func (h *UserHandler) Permissions(c *fiber.Ctx) error {
	if err := core.RequireAdmin(c); err != nil {
		return err
	}

	permissions, _ := usersvc.PermissionServiceInstance.FindAll()

	type PermissionWithCount struct {
		usermodel.Permission
		RoleCount int
	}
	var perms []PermissionWithCount
	for _, p := range permissions {
		count, _ := usersvc.PermissionServiceInstance.CountRolesWithPermission(p.ID)
		perms = append(perms, PermissionWithCount{Permission: p, RoleCount: count})
	}

	data := struct {
		Permissions []PermissionWithCount
	}{Permissions: perms}
	return core.RenderPage(c, "page-permissions", data)
}

func (h *UserHandler) CreatePermission(c *fiber.Ctx) error {
	if err := core.RequireAdmin(c); err != nil {
		return err
	}

	name := strings.TrimSpace(c.FormValue("name"))
	description := strings.TrimSpace(c.FormValue("description"))

	if name == "" {
		return core.RenderError(c, fiber.StatusBadRequest, "Permission name is required.")
	}

	if _, err := usersvc.PermissionServiceInstance.Create(name, description); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to create permission. Name may already exist.")
	}

	return h.Permissions(c)
}

func (h *UserHandler) DeletePermission(c *fiber.Ctx) error {
	if err := core.RequireAdmin(c); err != nil {
		return err
	}

	id, _ := c.ParamsInt("id")
	if id <= 0 {
		return core.RenderError(c, fiber.StatusBadRequest, "Invalid permission ID.")
	}

	count, _ := usersvc.PermissionServiceInstance.CountRolesWithPermission(int64(id))
	if count > 0 {
		return core.RenderError(c, fiber.StatusBadRequest,
			"Cannot delete: assigned to "+strconv.Itoa(count)+" role(s). Remove it from roles first.")
	}

	if err := usersvc.PermissionServiceInstance.Delete(int64(id)); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to delete permission.")
	}

	return h.Permissions(c)
}
