package organization

import (
	core "ems/app/handlers/pages/core"
	orgmodel "ems/app/models/organization"
	orgsvc "ems/app/services/organization"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type PositionRow struct {
	Position       orgmodel.Position
	DepartmentName string
}

type PositionsPageData struct {
	Positions   []PositionRow
	Departments []orgmodel.Department
}

func (h *OrganizationHandler) Positions(c *fiber.Ctx) error {
	data := PositionsPageData{}

	departments, _ := orgsvc.DepartmentServiceInst.FindAll()
	data.Departments = departments
	deptNames := map[int64]string{}
	for _, d := range departments {
		deptNames[d.ID] = d.Name
	}

	positions, _ := orgsvc.PositionServiceInst.FindAll()
	for _, p := range positions {
		data.Positions = append(data.Positions, PositionRow{Position: p, DepartmentName: deptNames[p.DepartmentID]})
	}
	return core.RenderPage(c, "page-positions", data)
}

func (h *OrganizationHandler) UpdatePosition(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	name := strings.TrimSpace(c.FormValue("name"))
	if name == "" {
		return core.RenderError(c, fiber.StatusBadRequest, "Please provide a name.")
	}
	if _, err := orgsvc.PositionServiceInst.Update(id, name, c.FormValue("description"), c.FormValue("status")); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to update position.")
	}
	return h.Positions(c)
}

func (h *OrganizationHandler) DeletePosition(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if err := orgsvc.PositionServiceInst.Delete(id); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to delete position.")
	}
	return h.Positions(c)
}
