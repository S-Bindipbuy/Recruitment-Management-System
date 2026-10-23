package report

import (
	"ems/app/enums"
	core "ems/app/handlers/pages/core"
	evalmodel "ems/app/models/evaluation"
	evalsvc "ems/app/services/evaluation"
	orgsvc "ems/app/services/organization"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ReportHandler struct{}

type ResultRow struct {
	AssignmentID int64
	Score        string
	Percentage   string
	Grade        string
	GradeClass   string
}

type DeptRow struct {
	DepartmentID int64
	AvgPct       float64
	WidthPct     float64
	Count        int
}

type ReportsPageData struct {
	Stats       core.DashboardStats
	Results     []ResultRow
	Departments []DeptRow
}

func gradeClass(grade string) string {
	switch strings.ToUpper(grade) {
	case "A":
		return "bg-green-100 text-green-800"
	case "B":
		return "bg-blue-100 text-blue-800"
	case "C":
		return "bg-yellow-100 text-yellow-800"
	case "D":
		return "bg-orange-100 text-orange-800"
	case "F":
		return "bg-red-100 text-red-800"
	default:
		return "bg-gray-100 text-gray-800"
	}
}

func (h *ReportHandler) Reports(c *fiber.Ctx) error {
	data := ReportsPageData{Stats: core.LoadDashboardStats(core.CurrentUserID(c))}

	assignments := h.reviewedAssignments()
	type deptTotal struct {
		sum   float64
		count int
	}
	totals := map[int64]*deptTotal{}
	var deptOrder []int64

	for _, a := range assignments {
		result, err := evalsvc.ResultServiceInstance.GetResult(a.ID)
		if err != nil || result == nil {
			continue
		}
		data.Results = append(data.Results, ResultRow{
			AssignmentID: a.ID,
			Score:        fmt.Sprintf("%.2f", result.TotalScore),
			Percentage:   fmt.Sprintf("%.1f%%", result.Percentage),
			Grade:        result.Grade,
			GradeClass:   gradeClass(result.Grade),
		})

		deptID := int64(0)
		if emp, err := orgsvc.EmployeeServiceInstance.FindByUserID(a.EvaluateeID); err == nil && emp != nil {
			deptID = emp.DepartmentID
		}
		total, ok := totals[deptID]
		if !ok {
			total = &deptTotal{}
			totals[deptID] = total
			deptOrder = append(deptOrder, deptID)
		}
		total.sum += result.Percentage
		total.count++
	}

	for _, deptID := range deptOrder {
		total := totals[deptID]
		avg := total.sum / float64(total.count)
		data.Departments = append(data.Departments, DeptRow{
			DepartmentID: deptID,
			AvgPct:       avg,
			WidthPct:     min(avg, 100),
			Count:        total.count,
		})
	}

	return core.RenderPage(c, "page-reports", data)
}

func (h *ReportHandler) reviewedAssignments() []evalmodel.EvaluationAssignment {
	var assignments []evalmodel.EvaluationAssignment
	for _, status := range []string{
		string(enums.AssignmentStatusSubmitted),
		string(enums.AssignmentStatusCompleted),
	} {
		found, err := evalsvc.AssignmentServiceInstance.FindByStatus(status)
		if err != nil {
			continue
		}
		assignments = append(assignments, found...)
	}
	return assignments
}
