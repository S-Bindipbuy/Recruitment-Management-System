package user

import (
	core "ems/app/handlers/pages/core"
	usermodel "ems/app/models/user"
	usersvc "ems/app/services/user"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func (h *UserHandler) Users(c *fiber.Ctx) error {
	if err := core.RequireAdmin(c); err != nil {
		return err
	}

	users, _ := usersvc.UserServiceInstance.FindAll()
	roles, _ := usersvc.RoleServiceInstance.FindAll()

	type UserWithRoles struct {
		usermodel.User
		Roles   []string
		RoleIDs []int64
	}
	var usersWithRoles []UserWithRoles
	for _, u := range users {
		userRoles, _ := usersvc.UserServiceInstance.GetUserRoles(u.ID)
		roleIDs, _ := usersvc.UserServiceInstance.GetUserRoleIDs(u.ID)
		usersWithRoles = append(usersWithRoles, UserWithRoles{User: u, Roles: userRoles, RoleIDs: roleIDs})
	}

	data := struct {
		Users   []UserWithRoles
		Roles   []usermodel.Role
		IsAdmin bool
	}{
		Users:   usersWithRoles,
		Roles:   roles,
		IsAdmin: true,
	}
	return core.RenderPage(c, "page-users", data)
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	if err := core.RequireAdmin(c); err != nil {
		return err
	}

	name := strings.TrimSpace(c.FormValue("name"))
	email := strings.TrimSpace(c.FormValue("email"))
	password := c.FormValue("password")

	if name == "" || email == "" || password == "" {
		return core.RenderError(c, fiber.StatusBadRequest, "Name, email, and password are required.")
	}

	_, err := usersvc.UserServiceInstance.Register(name, email, password)
	if err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to create user. Email may already be taken.")
	}

	// Assign role
	roleID := core.FormInt64(c, "role_id")
	user, _ := usersvc.UserServiceInstance.FindByEmail(email)
	if user != nil {
		if roleID > 0 {
			usersvc.UserServiceInstance.AssignRoles(user.ID, []int64{roleID})
		} else {
			usersvc.UserServiceInstance.AssignRoles(user.ID, nil)
		}
	}

	return h.Users(c)
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	if err := core.RequireAdmin(c); err != nil {
		return err
	}

	id, _ := c.ParamsInt("id")
	name := strings.TrimSpace(c.FormValue("name"))
	email := strings.TrimSpace(c.FormValue("email"))
	status := c.FormValue("status")

	if id <= 0 || name == "" || email == "" {
		return core.RenderError(c, fiber.StatusBadRequest, "Name and email are required.")
	}

	// Update role
	roleID := core.FormInt64(c, "role_id")

	// An administrator must not be able to strip the admin role from their own
	// account: nobody should be able to change their own role.
	if int64(id) == core.CurrentUserID(c) {
		keepsAdmin := false
		if roleID > 0 {
			if role, err := usersvc.RoleServiceInstance.FindByID(roleID); err == nil {
				keepsAdmin = strings.EqualFold(role.Name, "admin")
			}
		}
		hadAdmin := usersvc.PolicyServiceInstance.IsAdmin(int64(id))
		if hadAdmin && !keepsAdmin {
			return core.RenderError(c, fiber.StatusBadRequest, "You cannot change or remove your own admin role.")
		}
	}

	_, err := usersvc.UserServiceInstance.Update(int64(id), name, email, status)
	if err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to update user.")
	}

	if roleID > 0 {
		usersvc.UserServiceInstance.AssignRoles(int64(id), []int64{roleID})
	} else {
		usersvc.UserServiceInstance.AssignRoles(int64(id), nil)
	}

	return h.Users(c)
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	if err := core.RequireAdmin(c); err != nil {
		return err
	}

	id, _ := c.ParamsInt("id")
	if id <= 0 {
		return core.RenderError(c, fiber.StatusBadRequest, "Invalid user ID.")
	}

	if int64(id) == core.CurrentUserID(c) {
		return core.RenderError(c, fiber.StatusBadRequest, "You cannot delete your own account.")
	}

	if err := usersvc.UserServiceInstance.Delete(int64(id)); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to delete user.")
	}

	return h.Users(c)
}
