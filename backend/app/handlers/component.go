package handlers

import (
	"os"
	"path/filepath"
	"strings"

	"ems/config"

	"github.com/gofiber/fiber/v2"
)

// allowedComponents is the whole set of fragments the front end may ask for.
// The endpoint is public, so without this list it would hand out every .html
// file that happens to sit in the components folder.
var allowedComponents = map[string]bool{
	"auth":    true,
	"header":  true,
	"sidebar": true,
}

type ComponentHandler struct{}

func (h *ComponentHandler) GetComponent(c *fiber.Ctx) error {
	name := c.Params("name")
	name = filepath.Base(name)
	name = strings.TrimSuffix(name, ".html")

	if !allowedComponents[name] {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Component not found",
		})
	}

	filePath := filepath.Join(config.AppConfig.FrontendPath, "components", name+".html")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Component not found",
		})
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendFile(filePath)
}
