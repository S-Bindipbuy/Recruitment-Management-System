package repositories

import (
	"ems/database"
	"strings"
)

// InClause turns ids into an "IN (?,?)" fragment and its arguments, dropping
// zeros and duplicates so callers can hand it ids straight from a form. The
// bool reports whether there is anything to match at all.
func InClause(ids []int64) (string, []any, bool) {
	seen := make(map[int64]bool, len(ids))
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		if id > 0 && !seen[id] {
			seen[id] = true
			args = append(args, id)
		}
	}
	if len(args) == 0 {
		return "", nil, false
	}
	return strings.TrimSuffix(strings.Repeat("?,", len(args)), ","), args, true
}

// ActiveRecordExists reports whether a row with that id still exists and is
// active. Callers pass a fixed table name of their own choosing - never
// anything that came from a request.
func ActiveRecordExists(table string, id int64) (bool, error) {
	var ok bool
	err := database.DB.Get(&ok,
		"SELECT EXISTS(SELECT 1 FROM "+table+" WHERE id = ? AND status = 'active')", id)
	return ok, err
}
