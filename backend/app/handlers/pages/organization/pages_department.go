package organization

import (
	core "ems/app/handlers/pages/core"
	orgmodel "ems/app/models/organization"
	orgsvc "ems/app/services/organization"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type DepartmentListItem struct {
	OrganizationName string
	Department       orgmodel.Department
}

type DepartmentsPageData struct {
	Departments   []DepartmentListItem
	Organizations []orgmodel.Organization
}

func (h *OrganizationHandler) Departments(c *fiber.Ctx) error {
	data := DepartmentsPageData{}

	orgs, _ := orgsvc.OrganizationServiceInst.FindAll()
	data.Organizations = orgs
	orgNames := map[int64]string{}
	for _, o := range orgs {
		orgNames[o.ID] = o.Name
	}

	departments, _ := orgsvc.DepartmentServiceInst.FindAll()
	for _, d := range departments {
		data.Departments = append(data.Departments, DepartmentListItem{Department: d, OrganizationName: orgNames[d.OrganizationID]})
	}
	return core.RenderPage(c, "page-departments", data)
}

func (h *OrganizationHandler) UpdateDepartment(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	name := strings.TrimSpace(c.FormValue("name"))
	if name == "" {
		return core.RenderError(c, fiber.StatusBadRequest, "Please provide a name.")
	}
	if _, err := orgsvc.DepartmentServiceInst.Update(id, name, c.FormValue("description"), c.FormValue("status")); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to update department.")
	}
	return h.Departments(c)
}

func (h *OrganizationHandler) DeleteDepartment(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	if err := orgsvc.DepartmentServiceInst.Delete(id); err != nil {
		return core.RenderError(c, fiber.StatusInternalServerError, "Failed to delete department.")
	}
	return h.Departments(c)
}
