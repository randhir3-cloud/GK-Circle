package models

import (
	"testing"
)

func TestAddRole(t *testing.T) {
	tests := []struct {
		existing string
		role     string
		expected string
	}{
		{"", "super_admin", "super_admin"},
		{"user", "super_admin", "user,super_admin"},
		{"user,admin", "super_admin", "user,admin,super_admin"},
		{"user, super_admin", "super_admin", "user,super_admin"},
		{"user,super_admin,super_admin", "super_admin", "user,super_admin"},
		{" USER , Moderator ", "super_admin", "user,moderator,super_admin"},
	}

	for _, tc := range tests {
		result := AddRole(tc.existing, tc.role)
		if result != tc.expected {
			t.Errorf("AddRole(%q, %q) = %q; expected %q", tc.existing, tc.role, result, tc.expected)
		}
	}
}
