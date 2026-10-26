package user

import (
	usermodel "ems/app/models/user"
	"ems/app/repositories"
	"ems/database"
	"strings"
	"time"
)

type UserRepository struct{}

func (r *UserRepository) Create(user *usermodel.User) error {
	result, err := database.DB.Exec(
		"INSERT INTO users (name, email, password, status) VALUES (?, ?, ?, ?)",
		user.Name, user.Email, user.Password, user.Status,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	user.ID = id
	return nil
}

func (r *UserRepository) FindByEmail(email string) (*usermodel.User, error) {
	user := &usermodel.User{}
	err := database.DB.Get(user,
		"SELECT id, name, email, password, status, created_at, updated_at FROM users WHERE email = ?",
		email,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByID(id int64) (*usermodel.User, error) {
	user := &usermodel.User{}
	err := database.DB.Get(user,
		"SELECT id, name, email, password, status, created_at, updated_at FROM users WHERE id = ?",
		id,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindAll() ([]usermodel.User, error) {
	var users []usermodel.User
	err := database.DB.Select(&users, "SELECT id, name, email, status, created_at, updated_at FROM users")
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) Update(user *usermodel.User) error {
	_, err := database.DB.Exec(
		"UPDATE users SET name = ?, email = ?, status = ?, updated_at = ? WHERE id = ?",
		user.Name, user.Email, user.Status, time.Now(), user.ID,
	)
	return err
}

func (r *UserRepository) UpdatePassword(id int64, hashedPassword string) error {
	_, err := database.DB.Exec("UPDATE users SET password = ?, updated_at = ? WHERE id = ?", hashedPassword, time.Now(), id)
	return err
}

func (r *UserRepository) Delete(id int64) error {
	_, err := database.DB.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

func (r *UserRepository) GetUserRoleIDs(userID int64) ([]int64, error) {
	var ids []int64
	err := database.DB.Select(&ids,
		"SELECT role_id FROM user_roles WHERE user_id = ?", userID,
	)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *UserRepository) GetUserRoles(userID int64) ([]string, error) {
	var roles []string
	err := database.DB.Select(&roles,
		`SELECT r.name FROM roles r 
		 INNER JOIN user_roles ur ON r.id = ur.role_id 
		 WHERE ur.user_id = ?`, userID,
	)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *UserRepository) AssignRoles(userID int64, roleIDs []int64) error {
	tx, err := database.DB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM user_roles WHERE user_id = ?", userID)
	if err != nil {
		return err
	}

	for _, roleID := range roleIDs {
		_, err = tx.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", userID, roleID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

type RoleRepository struct{}

func (r *RoleRepository) Create(role *usermodel.Role) error {
	result, err := database.DB.Exec(
		"INSERT INTO roles (name, description) VALUES (?, ?)",
		role.Name, role.Description,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	role.ID = id
	return nil
}

func (r *RoleRepository) FindAll() ([]usermodel.Role, error) {
	var roles []usermodel.Role
	err := database.DB.Select(&roles,
		"SELECT id, name, COALESCE(description, '') AS description, created_at, updated_at FROM roles",
	)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *RoleRepository) FindByID(id int64) (*usermodel.Role, error) {
	role := &usermodel.Role{}
	err := database.DB.Get(role,
		"SELECT id, name, description, created_at, updated_at FROM roles WHERE id = ?", id,
	)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *RoleRepository) Update(role *usermodel.Role) error {
	_, err := database.DB.Exec(
		"UPDATE roles SET name = ?, description = ?, updated_at = ? WHERE id = ?",
		role.Name, role.Description, time.Now(), role.ID,
	)
	return err
}

func (r *RoleRepository) Delete(id int64) error {
	_, err := database.DB.Exec("DELETE FROM roles WHERE id = ?", id)
	return err
}

func (r *RoleRepository) IsAdminByUserID(userID int64) (bool, error) {
	var ok bool
	err := database.DB.Get(&ok, `
		SELECT EXISTS(
			SELECT 1 FROM user_roles ur
			INNER JOIN roles r ON r.id = ur.role_id
			WHERE ur.user_id = ? AND LOWER(r.name) = 'admin')`, userID)
	return ok, err
}

// FindActiveUserWithPermission returns the ID of an active user who holds the
// named permission (e.g. "evaluation.approve") and is not in excludeIDs.
// It is used to route an evaluation to a real approver instead of the person
// who just submitted it. The lowest matching ID wins so the choice is stable.
func (r *UserRepository) FindActiveUserWithPermission(excludeIDs []int64, permission string) (int64, error) {
	query := `
		SELECT u.id
		FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN role_permissions rp ON rp.role_id = ur.role_id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE u.status = 'active' AND p.name = ?`
	args := []any{permission}
	if len(excludeIDs) > 0 {
		placeholders := make([]string, len(excludeIDs))
		for i, id := range excludeIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		query += " AND u.id NOT IN (" + strings.Join(placeholders, ",") + ")"
	}
	query += " ORDER BY u.id LIMIT 1"

	var id int64
	err := database.DB.QueryRow(query, args...).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// FindActiveNonEmployees lists active accounts that hold no employee record:
// the people who can be chosen as a manager before they are hired as one.
func (r *UserRepository) FindActiveNonEmployees() ([]usermodel.UserOption, error) {
	var users []usermodel.UserOption
	err := database.DB.Select(&users, `
		SELECT id, name FROM users
		WHERE status = 'active'
		  AND NOT EXISTS (SELECT 1 FROM employees e WHERE e.user_id = users.id AND e.user_id IS NOT NULL)
		ORDER BY name`)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// FindNames returns the display names behind a set of user ids, so a page can
// label its rows in one query instead of one per row.
func (r *UserRepository) FindNames(ids []int64) (map[int64]string, error) {
	names := map[int64]string{}
	clause, args, ok := repositories.InClause(ids)
	if !ok {
		return names, nil
	}
	rows, err := database.DB.Query("SELECT id, name FROM users WHERE id IN ("+clause+")", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err == nil {
			names[id] = name
		}
	}
	return names, rows.Err()
}

// CountUsersWithRole reports how many accounts hold a role, so the roles page
// can refuse to drop one that is still in use.
func (r *RoleRepository) CountUsersWithRole(roleID int64) (int, error) {
	var count int
	err := database.DB.Get(&count, "SELECT COUNT(*) FROM user_roles WHERE role_id = ?", roleID)
	return count, err
}
