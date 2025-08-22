package datasources

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataSource_IsRoleAllowed(t *testing.T) {
	tests := []struct {
		name         string
		allowedRoles string
		testRole     string
		expected     bool
	}{
		{
			name:         "empty allowed roles allows all",
			allowedRoles: "",
			testRole:     "Admin",
			expected:     true,
		},
		{
			name:         "single role match",
			allowedRoles: "Admin",
			testRole:     "Admin",
			expected:     true,
		},
		{
			name:         "single role no match",
			allowedRoles: "Admin",
			testRole:     "Editor",
			expected:     false,
		},
		{
			name:         "multiple roles match first",
			allowedRoles: "Admin,Editor",
			testRole:     "Admin",
			expected:     true,
		},
		{
			name:         "multiple roles match second",
			allowedRoles: "Admin,Editor",
			testRole:     "Editor",
			expected:     true,
		},
		{
			name:         "multiple roles no match",
			allowedRoles: "Admin,Editor",
			testRole:     "Viewer",
			expected:     false,
		},
		{
			name:         "roles with spaces",
			allowedRoles: "Admin, Editor, Viewer",
			testRole:     "Editor",
			expected:     true,
		},
		{
			name:         "case sensitive",
			allowedRoles: "admin",
			testRole:     "Admin",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &DataSource{
				AllowedRoles: tt.allowedRoles,
			}
			
			result := ds.IsRoleAllowed(tt.testRole)
			assert.Equal(t, tt.expected, result)
		})
	}
}