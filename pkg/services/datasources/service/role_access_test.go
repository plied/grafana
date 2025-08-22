package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/services/datasources"
	"github.com/grafana/grafana/pkg/services/datasources/guardian"
	"github.com/grafana/grafana/pkg/services/org"
	"github.com/grafana/grafana/pkg/services/user"
)

func TestRoleBasedAccess_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Simple unit test that demonstrates the role-based access control functionality
	// without requiring full database setup

	orgID := int64(1)

	// Create test datasources with different role restrictions
	datasources := []*datasources.DataSource{
		{
			ID:           1,
			Name:         "unrestricted-ds",
			AllowedRoles: "",
		},
		{
			ID:           2,
			Name:         "admin-only-ds",
			AllowedRoles: "Admin",
		},
		{
			ID:           3,
			Name:         "editor-admin-ds",
			AllowedRoles: "Editor,Admin",
		},
		{
			ID:           4,
			Name:         "all-roles-ds",
			AllowedRoles: "Viewer,Editor,Admin",
		},
	}

	// Test different user roles
	testCases := []struct {
		name           string
		userRole       org.RoleType
		expectedCount  int
		expectedNames  []string
	}{
		{
			name:          "Admin sees all datasources",
			userRole:      org.RoleAdmin,
			expectedCount: 4,
			expectedNames: []string{"unrestricted-ds", "admin-only-ds", "editor-admin-ds", "all-roles-ds"},
		},
		{
			name:          "Editor sees appropriate datasources",
			userRole:      org.RoleEditor,
			expectedCount: 3,
			expectedNames: []string{"unrestricted-ds", "editor-admin-ds", "all-roles-ds"},
		},
		{
			name:          "Viewer sees limited datasources",
			userRole:      org.RoleViewer,
			expectedCount: 2,
			expectedNames: []string{"unrestricted-ds", "all-roles-ds"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user := &user.SignedInUser{
				OrgID:   orgID,
				OrgRole: tc.userRole,
			}

			// Create guardian
			guardianProvider := guardian.ProvideGuardian()
			guard := guardianProvider.New(orgID, user)

			// Filter datasources by role
			filtered, err := guard.FilterDatasourcesByReadPermissions(datasources)
			require.NoError(t, err)

			// Verify results
			assert.Len(t, filtered, tc.expectedCount)

			actualNames := make([]string, len(filtered))
			for i, ds := range filtered {
				actualNames[i] = ds.Name
			}
			assert.ElementsMatch(t, tc.expectedNames, actualNames)
		})
	}

	// Test role checking methods
	t.Run("IsRoleAllowed method works correctly", func(t *testing.T) {
		assert.True(t, datasources[0].IsRoleAllowed("Admin"))
		assert.True(t, datasources[0].IsRoleAllowed("Editor"))
		assert.True(t, datasources[0].IsRoleAllowed("Viewer"))

		assert.True(t, datasources[1].IsRoleAllowed("Admin"))
		assert.False(t, datasources[1].IsRoleAllowed("Editor"))
		assert.False(t, datasources[1].IsRoleAllowed("Viewer"))

		assert.True(t, datasources[2].IsRoleAllowed("Admin"))
		assert.True(t, datasources[2].IsRoleAllowed("Editor"))
		assert.False(t, datasources[2].IsRoleAllowed("Viewer"))

		assert.True(t, datasources[3].IsRoleAllowed("Admin"))
		assert.True(t, datasources[3].IsRoleAllowed("Editor"))
		assert.True(t, datasources[3].IsRoleAllowed("Viewer"))
	})
}