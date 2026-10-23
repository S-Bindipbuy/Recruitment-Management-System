package enums

type Permission string

const (
	PermissionAuditView          Permission = "audit.view"
	PermissionDashboardView      Permission = "dashboard.view"
	PermissionReportView         Permission = "report.view"
	PermissionUserManage         Permission = "user.manage"
	PermissionRoleManage         Permission = "role.manage"
	PermissionEmployeeManage     Permission = "employee.manage"
	PermissionOrganizationManage Permission = "organization.manage"
	PermissionRecruitmentManage  Permission = "recruitment.manage"
	PermissionEvaluationManage   Permission = "evaluation.manage"
	PermissionEvaluationAssign   Permission = "evaluation.assign"
	PermissionEvaluationEvaluate Permission = "evaluation.evaluate"
	PermissionEvaluationApprove  Permission = "evaluation.approve"
)

type PermissionInfo struct {
	Name        Permission
	Description string
}

// PermissionCatalog lists every permission the system knows about.
var PermissionCatalog = []PermissionInfo{
	{PermissionAuditView, "View audit logs"},
	{PermissionDashboardView, "View the dashboard"},
	{PermissionReportView, "View reports"},
	{PermissionUserManage, "Manage users and their roles"},
	{PermissionRoleManage, "Manage roles and role permissions"},
	{PermissionEmployeeManage, "Manage employees"},
	{PermissionOrganizationManage, "Manage organizations, departments and positions"},
	{PermissionRecruitmentManage, "Manage recruitment (job postings, interviews, candidates)"},
	{PermissionEvaluationManage, "Manage evaluation periods, templates, sections and questions"},
	{PermissionEvaluationAssign, "Assign workers to evaluations"},
	{PermissionEvaluationEvaluate, "Start, fill and submit assigned evaluations"},
	{PermissionEvaluationApprove, "Create and decide evaluation approval requests"},
}

func AllPermissionNames() []string {
	names := make([]string, 0, len(PermissionCatalog))
	for _, p := range PermissionCatalog {
		names = append(names, string(p.Name))
	}
	return names
}
