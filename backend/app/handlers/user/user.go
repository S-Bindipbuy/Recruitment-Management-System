package user

import (
	"ems/app/handlers/authz"
	usersvc "ems/app/services/user"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct{}

func (h *UserHandler) List(c *fiber.Ctx) error {
	if err := authz.RequireAdmin(c); err != nil {
		return err
	}

	users, err := usersvc.UserServiceInstance.FindAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}
	return c.JSON(users)
}

func (h *UserHandler) Get(c *fiber.Ctx) error {
	if err := authz.RequireAdmin(c); err != nil {
		return err
	}

	id, _ := c.ParamsInt("id")
	user, err := usersvc.UserServiceInstance.FindByID(int64(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}
	return c.JSON(user)
}

func (h *UserHandler) Update(c *fiber.Ctx) error {
	if err := authz.RequireAdmin(c); err != nil {
		return err
	}

	id, _ := c.ParamsInt("id")
	var req struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Status string `json:"status"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	user, err := usersvc.UserServiceInstance.Update(int64(id), req.Name, req.Email, req.Status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user",
		})
	}
	return c.JSON(user)
}

func (h *UserHandler) Delete(c *fiber.Ctx) error {
	if err := authz.RequireAdmin(c); err != nil {
		return err
	}

	id, _ := c.ParamsInt("id")
	if err := usersvc.UserServiceInstance.Delete(int64(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete user",
		})
	}
	return c.JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}
