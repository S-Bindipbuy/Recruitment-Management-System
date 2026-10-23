package httperr

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

// ResponseError is an error whose HTTP response is fully controlled by the
// application. Handlers and middleware return it to abort a request with a
// specific status and message. Returning a plain nil from a gate or middleware
// would let the handler chain continue and overwrite the response body, so
// every policy/auth abort goes through this type.
type ResponseError struct {
	Code    int
	Message string
	Served  bool // response body already written; keep it as-is
}

func (e *ResponseError) Error() string { return e.Message }

// New builds a ResponseError carrying a status and a message.
func New(code int, message string) *ResponseError {
	return &ResponseError{Code: code, Message: message}
}

// Served builds a ResponseError for a request whose response body has already
// been written (e.g. an app-shell file) but whose handler chain must stop.
func Served(code int, message string) *ResponseError {
	return &ResponseError{Code: code, Message: message, Served: true}
}

// Handler is the app-wide Fiber error handler. Every error leaves the server as
// JSON; rendering it is the frontend's job. The one exception is a body that was
// already written (Served), which is kept untouched.
func Handler(c *fiber.Ctx, err error) error {
	c.Set("Cache-Control", "no-store")

	var re *ResponseError
	if errors.As(err, &re) {
		if re.Served {
			return nil
		}
		return c.Status(re.Code).JSON(fiber.Map{"error": re.Message})
	}

	fe := &fiber.Error{}
	if errors.As(err, &fe) {
		return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
}
