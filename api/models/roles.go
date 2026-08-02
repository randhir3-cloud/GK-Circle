package models

import (
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
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

var SystemRoles = map[string]struct{}{
	SystemRoleUser:       {},
	SystemRoleModerator:  {},
	SystemRoleAdmin:      {},
	SystemRoleSuperAdmin: {},
}

// ParseRoles parses, normalizes, and returns unique valid roles from a comma-separated string
func ParseRoles(value string) []string {
	seen := make(map[string]struct{})
	roles := make([]string, 0)

	for _, rawRole := range strings.Split(value, ",") {
		role := strings.ToLower(strings.TrimSpace(rawRole))
		if role == "" {
			continue
		}

		if _, valid := SystemRoles[role]; !valid {
			continue
		}

		if _, exists := seen[role]; exists {
			continue
		}

		seen[role] = struct{}{}
		roles = append(roles, role)
	}

	return roles
}

// HasRole checks if the userRoles string contains any of the target roles
func HasRole(userRoles string, targets ...string) bool {
	roles := ParseRoles(userRoles)

	for _, target := range targets {
		normalizedTarget := strings.ToLower(strings.TrimSpace(target))

		for _, role := range roles {
			if role == normalizedTarget {
				return true
			}
		}
	}

	return false
}

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

// CanManageCourses checks if a user has the super_admin or admin roles required to manage courses.
func CanManageCourses(user *User) bool {
	if user == nil {
		return false
	}
	return HasRole(user.Roles, SystemRoleSuperAdmin, SystemRoleAdmin)
}

// CanManageQuizzes checks if a user is authorized to manage quizzes (super_admin, admin, or is public quiz admin).
func CanManageQuizzes(user *User, appConfig *config.AppConfig) bool {
	if user == nil {
		return false
	}
	return HasRole(user.Roles, SystemRoleSuperAdmin, SystemRoleAdmin) || (appConfig != nil && appConfig.Quiz.IsPublicQuizAdmin(user.Email))
}
