package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"ems/app/enums"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DevAdminEmail / DevAdminPassword are the credentials of the development
	// administrator seeded by SeedAdminUser. The same account is available as
	// portable SQL in seed.sql.
	DevAdminEmail    = "admin@example.com"
	DevAdminName     = "Administrator"
	DevAdminPassword = "admin123"
)

// SeedDefaultRBAC inserts the default permission catalog plus a default role
// set and their permissions. It is idempotent: existing rows are left intact
// and default roles are only created when no roles exist yet.
func SeedDefaultRBAC() {
	seedPermissions()

	if tableEmpty("roles") {
		seedDefaultRoles()
		return
	}
	grantAdminAll()
	seedMissingDefaultRoles()
}

func grantAdminAll() {
	rows, err := DB.Query("SELECT id FROM roles WHERE LOWER(name) = 'admin'")
	if err != nil {
		log.Printf("Failed to find admin role for seeding: %v", err)
		return
	}
	defer rows.Close()

	var adminIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err == nil {
			adminIDs = append(adminIDs, id)
		}
	}
	if len(adminIDs) == 0 {
		return
	}

	for _, permission := range enums.AllPermissionNames() {
		var permissionID int64
		if err := DB.QueryRow("SELECT id FROM permissions WHERE name = ?", permission).Scan(&permissionID); err != nil {
			continue
		}
		for _, adminID := range adminIDs {
			DB.Exec("INSERT OR IGNORE INTO role_permissions (role_id, permission_id) VALUES (?, ?)", adminID, permissionID)
		}
	}
	fmt.Println("Granted all permissions to the admin role")
}

func seedPermissions() {
	for _, p := range enums.PermissionCatalog {
		if _, err := DB.Exec(
			"INSERT OR IGNORE INTO permissions (name, description) VALUES (?, ?)",
			string(p.Name), p.Description,
		); err != nil {
			log.Printf("Failed to seed permission %s: %v", p.Name, err)
		}
	}
	fmt.Printf("Seeded %d permissions\n", len(enums.PermissionCatalog))
}

type defaultRole struct {
	role        string
	description string
	permissions []string
}

var defaultRoles = []defaultRole{
	{
		role: "admin", description: "Full system access",
		permissions: enums.AllPermissionNames(),
	},
	{
		role: "hrm", description: "HR management: employees, evaluations, reports",
		permissions: []string{
			string(enums.PermissionDashboardView),
			string(enums.PermissionReportView),
			string(enums.PermissionEmployeeManage),
			string(enums.PermissionOrganizationManage),
			string(enums.PermissionRecruitmentManage),
			string(enums.PermissionInterviewView),
			string(enums.PermissionInterviewAssign),
			string(enums.PermissionInterviewScore),
			string(enums.PermissionInterviewApprove),
			string(enums.PermissionEvaluationManage),
			string(enums.PermissionEvaluationAssign),
			string(enums.PermissionEvaluationEvaluate),
			string(enums.PermissionEvaluationApprove),
			string(enums.PermissionAuditView),
		},
	},
	{
		role: "manager", description: "Manages workers and evaluates reports to them",
		permissions: []string{
			string(enums.PermissionDashboardView),
			string(enums.PermissionReportView),
			string(enums.PermissionInterviewView),
			string(enums.PermissionInterviewAssign),
			string(enums.PermissionInterviewScore),
			string(enums.PermissionInterviewApprove),
			string(enums.PermissionEvaluationAssign),
			string(enums.PermissionEvaluationEvaluate),
			string(enums.PermissionEvaluationApprove),
		},
	},
	{
		role: "employee", description: "Fills in the evaluations assigned to them",
		permissions: []string{
			string(enums.PermissionDashboardView),
			string(enums.PermissionInterviewView),
			string(enums.PermissionInterviewScore),
			string(enums.PermissionEvaluationEvaluate),
		},
	},
	{
		role: "interviewer", description: "Scores the interviews they are assigned to",
		permissions: []string{
			string(enums.PermissionDashboardView),
			string(enums.PermissionInterviewView),
			string(enums.PermissionInterviewScore),
		},
	},
}

func seedDefaultRoles() {
	for _, d := range defaultRoles {
		createDefaultRole(d)
	}
	fmt.Println("Seeded default roles and role permissions")
}

// seedMissingDefaultRoles adds default roles introduced after an installation
// was first created. Existing roles are left untouched: an administrator may
// have deliberately adjusted their permissions.
func seedMissingDefaultRoles() {
	for _, d := range defaultRoles {
		var id int64
		if err := DB.QueryRow("SELECT id FROM roles WHERE LOWER(name) = LOWER(?)", d.role).Scan(&id); err == nil {
			continue
		}
		createDefaultRole(d)
	}
}

func createDefaultRole(d defaultRole) {
	result, err := DB.Exec(
		"INSERT OR IGNORE INTO roles (name, description) VALUES (?, ?)",
		d.role, d.description,
	)
	if err != nil {
		log.Printf("Failed to seed role %s: %v", d.role, err)
		return
	}
	id, _ := result.LastInsertId()
	if id == 0 {
		if err := DB.QueryRow("SELECT id FROM roles WHERE name = ?", d.role).Scan(&id); err != nil {
			log.Printf("Failed to find seeded role %s: %v", d.role, err)
			return
		}
	}
	for _, permission := range d.permissions {
		if _, err := DB.Exec(
			`INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
			 SELECT ?, id FROM permissions WHERE name = ?`,
			id, permission,
		); err != nil {
			log.Printf("Failed to grant %s to %s: %v", permission, d.role, err)
		}
	}
}

func tableEmpty(table string) bool {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
	return err != nil || count == 0
}

// SeedAdminUser creates the development administrator account (see
// DevAdminEmail/DevAdminPassword) and guarantees it holds the admin role.
// It is idempotent: an existing account is left untouched apart from the role
// assignment. Set APP_ENV=production to disable it.
//
// The portable equivalent lives in seed.sql:
//
//	sqlite3 ./storage/db < backend/database/seed.sql
func SeedAdminUser() {
	if DB == nil {
		return
	}
	if strings.EqualFold(os.Getenv("APP_ENV"), "production") {
		return
	}

	var userID int64
	err := DB.QueryRow("SELECT id FROM users WHERE email = ?", DevAdminEmail).Scan(&userID)
	switch {
	case err == nil:
		grantAdminRole(userID)
		return
	case err != sql.ErrNoRows:
		log.Printf("Failed to look up development admin: %v", err)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(DevAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash development admin password: %v", err)
		return
	}
	result, err := DB.Exec(
		"INSERT INTO users (name, email, password, status) VALUES (?, ?, ?, 'active')",
		DevAdminName, DevAdminEmail, string(hash),
	)
	if err != nil {
		log.Printf("Failed to seed development admin: %v", err)
		return
	}
	if id, err := result.LastInsertId(); err == nil {
		grantAdminRole(id)
	}
	log.Printf("Seeded development admin %s (password: %s)", DevAdminEmail, DevAdminPassword)
}

// grantAdminRole assigns the admin role to userID, creating the link when it
// is missing. It is a no-op when the admin role does not exist yet.
func grantAdminRole(userID int64) {
	var roleID int64
	if err := DB.QueryRow("SELECT id FROM roles WHERE LOWER(name) = 'admin'").Scan(&roleID); err != nil {
		return
	}
	DB.Exec("INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?)", userID, roleID)
}
