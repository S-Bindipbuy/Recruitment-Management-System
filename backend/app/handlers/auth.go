package handlers

import (
	"ems/app/enums"
	"ems/app/middleware"
	usersvc "ems/app/services/user"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct{}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	user, err := usersvc.UserServiceInstance.FindByEmail(req.Email)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	if user.Status != string(enums.UserStatusActive) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Account is inactive",
		})
	}

	roles, _ := usersvc.UserServiceInstance.GetUserRoles(user.ID)
	if roles == nil {
		roles = []string{}
	}

	token, err := middleware.GenerateToken(user.ID, user.Email, roles)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"roles": roles,
		},
	})
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}

	user, err := usersvc.UserServiceInstance.Register(req.Name, req.Email, string(hashedPassword))
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Email already exists",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"user": fiber.Map{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)
	user, err := usersvc.UserServiceInstance.FindByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	roles, _ := usersvc.UserServiceInstance.GetUserRoles(user.ID)
	permissions := usersvc.PolicyServiceInstance.PermissionsForUser(user.ID)

	return c.JSON(fiber.Map{
		"id":          user.ID,
		"name":        user.Name,
		"email":       user.Email,
		"roles":       roles,
		"permissions": permissions,
		"is_admin":    usersvc.PolicyServiceInstance.IsAdmin(user.ID),
	})
}
