package organization

import (
	core "ems/app/handlers/pages/core"
	orgmodel "ems/app/models/organization"
	orgsvc "ems/app/services/organization"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type OrganizationHandler struct{}

type OrganizationsPageData struct {
	Organizations []orgmodel.Organization
}

func (h *OrganizationHandler) Organizations(c *fiber.Ctx) error {
	data := OrganizationsPageData{}
	data.Organizations, _ = orgsvc.OrganizationServiceInst.FindAll()
	return core.RenderPage(c, "page-organizations", data)
}

func (h *OrganizationHandler) CreateOrganization(c *fiber.Ctx) error {
	name := strings.TrimSpace(c.FormValue("name"))
	if name == "" {
		return core.RenderError(c, fiber.StatusBadRequest, "Please provide a name.")
	}
	if _, err := orgsvc.OrganizationServiceInst.Create(name, c.FormValue("description")); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to create organization.")
	}
	return h.Organizations(c)
}

func (h *OrganizationHandler) UpdateOrganization(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	name := strings.TrimSpace(c.FormValue("name"))
	if name == "" {
		return core.RenderError(c, fiber.StatusBadRequest, "Please provide a name.")
	}
	if _, err := orgsvc.OrganizationServiceInst.Update(id, name, c.FormValue("description"), c.FormValue("status")); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to update organization.")
	}
	return h.Organizations(c)
}

func (h *OrganizationHandler) DeleteOrganization(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if err := orgsvc.OrganizationServiceInst.Delete(id); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to delete organization.")
	}
	return h.Organizations(c)
}

func (h *OrganizationHandler) CreateDepartment(c *fiber.Ctx) error {
	name := strings.TrimSpace(c.FormValue("name"))
	orgID := core.FormInt64(c, "organization_id")
	if name == "" || orgID <= 0 {
		return core.RenderError(c, fiber.StatusBadRequest, "Please provide a name and an organization.")
	}
	if !core.ActiveRecordExists("organizations", orgID) {
		return core.RenderError(c, fiber.StatusBadRequest, "The selected organization was not found.")
	}
	if _, err := orgsvc.DepartmentServiceInst.Create(orgID, name, c.FormValue("description")); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to create department.")
	}
	return h.Organizations(c)
}

func (h *OrganizationHandler) CreatePosition(c *fiber.Ctx) error {
	name := strings.TrimSpace(c.FormValue("name"))
	deptID := core.FormInt64(c, "department_id")
	if name == "" || deptID <= 0 {
		return core.RenderError(c, fiber.StatusBadRequest, "Please provide a name and a department.")
	}
	if !core.ActiveRecordExists("departments", deptID) {
		return core.RenderError(c, fiber.StatusBadRequest, "The selected department was not found.")
	}
	if _, err := orgsvc.PositionServiceInst.Create(deptID, name, c.FormValue("description")); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to create position.")
	}
	return h.Organizations(c)
}
