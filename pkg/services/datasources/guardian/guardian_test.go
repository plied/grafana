package guardian

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/services/datasources"
	"github.com/grafana/grafana/pkg/services/org"
	"github.com/grafana/grafana/pkg/services/user"
)

func TestRoleBasedGuardian_FilterDatasourcesByRole(t *testing.T) {
	tests := []struct {
		name         string
		userRole     org.RoleType
		datasources  []*datasources.DataSource
		expectedLen  int
		expectedUIDs []string
	}{
		{
			name:     "admin user sees all datasources",
			userRole: org.RoleAdmin,
			datasources: []*datasources.DataSource{
				{UID: "ds1", AllowedRoles: ""},
				{UID: "ds2", AllowedRoles: "Admin"},
				{UID: "ds3", AllowedRoles: "Editor,Admin"},
			},
			expectedLen:  3,
			expectedUIDs: []string{"ds1", "ds2", "ds3"},
		},
		{
			name:     "editor user sees appropriate datasources",
			userRole: org.RoleEditor,
			datasources: []*datasources.DataSource{
				{UID: "ds1", AllowedRoles: ""},
				{UID: "ds2", AllowedRoles: "Admin"},
				{UID: "ds3", AllowedRoles: "Editor,Admin"},
				{UID: "ds4", AllowedRoles: "Editor"},
			},
			expectedLen:  3,
			expectedUIDs: []string{"ds1", "ds3", "ds4"},
		},
		{
			name:     "viewer user sees only unrestricted and viewer-allowed datasources",
			userRole: org.RoleViewer,
			datasources: []*datasources.DataSource{
				{UID: "ds1", AllowedRoles: ""},
				{UID: "ds2", AllowedRoles: "Admin"},
				{UID: "ds3", AllowedRoles: "Editor,Admin"},
				{UID: "ds4", AllowedRoles: "Viewer"},
				{UID: "ds5", AllowedRoles: "Viewer,Editor"},
			},
			expectedLen:  3,
			expectedUIDs: []string{"ds1", "ds4", "ds5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &user.SignedInUser{
				OrgRole: tt.userRole,
			}
			guardian := NewRoleBasedGuardian(user)

			filtered, err := guardian.FilterDatasourcesByReadPermissions(tt.datasources)
			require.NoError(t, err)
			
			assert.Len(t, filtered, tt.expectedLen)
			
			actualUIDs := make([]string, len(filtered))
			for i, ds := range filtered {
				actualUIDs[i] = ds.UID
			}
			assert.ElementsMatch(t, tt.expectedUIDs, actualUIDs)
		})
	}
}

func TestOSSProvider_New(t *testing.T) {
	provider := &OSSProvider{}
	user := &user.SignedInUser{OrgRole: org.RoleViewer}

	t.Run("always returns RoleBasedGuardian", func(t *testing.T) {
		datasources := []datasources.DataSource{
			{UID: "ds1", AllowedRoles: ""},
			{UID: "ds2", AllowedRoles: ""},
		}

		guardian := provider.New(1, user, datasources...)
		_, isRoleBasedGuardian := guardian.(*RoleBasedGuardian)
		assert.True(t, isRoleBasedGuardian)
	})

	t.Run("returns RoleBasedGuardian when role restrictions exist", func(t *testing.T) {
		datasources := []datasources.DataSource{
			{UID: "ds1", AllowedRoles: ""},
			{UID: "ds2", AllowedRoles: "Admin"},
		}

		guardian := provider.New(1, user, datasources...)
		_, isRoleBasedGuardian := guardian.(*RoleBasedGuardian)
		assert.True(t, isRoleBasedGuardian)
	})
}