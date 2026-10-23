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

func seedDefaultRoles() {
	permissionIDs := map[string]int64{}
	rows, err := DB.Query("SELECT id, name FROM permissions")
	if err != nil {
		log.Printf("Failed to load permissions for seeding: %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err == nil {
			permissionIDs[name] = id
		}
	}

	roleID := func(name, description string) int64 {
		result, err := DB.Exec(
			"INSERT OR IGNORE INTO roles (name, description) VALUES (?, ?)",
			name, description,
		)
		if err != nil {
			log.Printf("Failed to seed role %s: %v", name, err)
			return 0
		}
		if id, err := result.LastInsertId(); err == nil && id > 0 {
			return id
		}
		var existing int64
		if err := DB.QueryRow("SELECT id FROM roles WHERE name = ?", name).Scan(&existing); err == nil {
			return existing
		}
		return 0
	}
	grant := func(roleID int64, name string) {
		if roleID == 0 {
			return
		}
		if permsID, ok := permissionIDs[name]; ok {
			DB.Exec("INSERT OR IGNORE INTO role_permissions (role_id, permission_id) VALUES (?, ?)", roleID, permsID)
		}
	}

	defaults := []struct {
		role        string
		description string
		permissions []string
	}{
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
				string(enums.PermissionEvaluationAssign),
				string(enums.PermissionEvaluationEvaluate),
				string(enums.PermissionEvaluationApprove),
			},
		},
		{
			role: "employee", description: "Fills in the evaluations assigned to them",
			permissions: []string{
				string(enums.PermissionDashboardView),
				string(enums.PermissionEvaluationEvaluate),
			},
		},
	}

	for _, d := range defaults {
		id := roleID(d.role, d.description)
		for _, p := range d.permissions {
			grant(id, p)
		}
	}
	fmt.Println("Seeded default roles and role permissions")
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
