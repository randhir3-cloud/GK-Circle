package models

import (
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
)

type Role string

type AllowedRoles struct {
	roles map[Role]any
}

const (
	SystemRoleUser       = "user"
	SystemRoleModerator  = "moderator"
	SystemRoleAdmin      = "admin"
	SystemRoleSuperAdmin = "super_admin"
)

// QuizModel implements quiz related database operations
type RoleModel struct {
	db          *goqu.Database
	systemRoles []Role
}

func InitRoleModel(db *goqu.Database) *RoleModel {
	return &RoleModel{
		db:          db,
		systemRoles: []Role{SystemRoleSuperAdmin, SystemRoleAdmin, SystemRoleModerator, SystemRoleUser},
	}
}

// AddRole normailzes comma-separated roles and additively appends a new role.
func AddRole(existing string, role string) string {
	seen := make(map[string]struct{})
	roles := make([]string, 0)

	for _, value := range strings.Split(existing, ",") {
		normalized := strings.ToLower(strings.TrimSpace(value))
		if normalized == "" {
			continue
		}

		if _, exists := seen[normalized]; exists {
			continue
		}

		seen[normalized] = struct{}{}
		roles = append(roles, normalized)
	}

	normalizedRole := strings.ToLower(strings.TrimSpace(role))
	if _, exists := seen[normalizedRole]; !exists {
		roles = append(roles, normalizedRole)
	}

	return strings.Join(roles, ",")
}


func (rm *RoleModel) NewAllowedRoles(roles ...string) (AllowedRoles, error) {
	allowedRoles := AllowedRoles{roles: make(map[Role]any)}
	validRoles := []Role{}

	for _, r := range roles {
		role := Role(r)
		matched := false
		for _, ra := range rm.systemRoles {
			if ra == role {
				validRoles = append(validRoles, role)
				matched = true
			}
		}
		if !matched {
			return AllowedRoles{}, fmt.Errorf("Role not found: %s", r)
		}
	}

	for _, role := range validRoles {
		allowedRoles.roles[role] = any(nil)
	}

	return allowedRoles, nil
}

func (ar *AllowedRoles) IsAllowed(role Role) bool {
	_, found := ar.roles[role]
	return found
}
