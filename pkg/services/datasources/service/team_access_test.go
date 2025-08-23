package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/services/datasources"
	fakeDS "github.com/grafana/grafana/pkg/services/datasources/fakes"
	"github.com/grafana/grafana/pkg/services/datasources/guardian"
	"github.com/grafana/grafana/pkg/services/org"
	"github.com/grafana/grafana/pkg/services/team"
	"github.com/grafana/grafana/pkg/services/user"
)

// Mock team service for testing
type mockTeamIDsByUserGetter struct {
	userTeams map[int64][]int64 // maps userID to team IDs
}

func (m *mockTeamIDsByUserGetter) GetTeamIDsByUser(ctx context.Context, query *team.GetTeamIDsByUserQuery) ([]int64, error) {
	if teams, ok := m.userTeams[query.UserID]; ok {
		return teams, nil
	}
	return []int64{}, nil
}

func TestTeamBasedAccess_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Simple unit test that demonstrates the team-based access control functionality
	// without requiring full database setup

	orgID := int64(1)

	// Create test datasources with different team restrictions
	datasources := []*datasources.DataSource{
		{
			ID:           1,
			Name:         "unrestricted-ds",
			AllowedTeams: "",
		},
		{
			ID:           2,
			Name:         "team1-only-ds",
			AllowedTeams: "1",
		},
		{
			ID:           3,
			Name:         "team1-team2-ds",
			AllowedTeams: "1,2",
		},
		{
			ID:           4,
			Name:         "team3-ds",
			AllowedTeams: "3",
		},
	}

	// Test different user roles and team memberships
	testCases := []struct {
		name           string
		userRole       org.RoleType
		userID         int64
		userTeams      []int64
		expectedCount  int
		expectedNames  []string
	}{
		{
			name:          "Admin sees all datasources regardless of team membership",
			userRole:      org.RoleAdmin,
			userID:        1,
			userTeams:     []int64{},
			expectedCount: 4,
			expectedNames: []string{"unrestricted-ds", "team1-only-ds", "team1-team2-ds", "team3-ds"},
		},
		{
			name:          "Editor in team 1 sees appropriate datasources",
			userRole:      org.RoleEditor,
			userID:        2,
			userTeams:     []int64{1},
			expectedCount: 3,
			expectedNames: []string{"unrestricted-ds", "team1-only-ds", "team1-team2-ds"},
		},
		{
			name:          "Viewer in team 2 sees appropriate datasources",
			userRole:      org.RoleViewer,
			userID:        3,
			userTeams:     []int64{2},
			expectedCount: 2,
			expectedNames: []string{"unrestricted-ds", "team1-team2-ds"},
		},
		{
			name:          "User with no teams sees only unrestricted datasources",
			userRole:      org.RoleViewer,
			userID:        4,
			userTeams:     []int64{},
			expectedCount: 1,
			expectedNames: []string{"unrestricted-ds"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user := &user.SignedInUser{
				UserID:  tc.userID,
				OrgID:   orgID,
				OrgRole: tc.userRole,
			}

			// Create guardian with mock team service
			fakeDSService := &fakeDS.FakeDataSourceService{
				DataSources: datasources,
			}
			mockTeamSvc := &mockTeamIDsByUserGetter{
				userTeams: map[int64][]int64{
					tc.userID: tc.userTeams,
				},
			}
			guardianProvider := &guardian.OSSProvider{}
			guard := guardian.NewTeamBasedGuardianWithGetter(user, orgID, fakeDSService, mockTeamSvc)

			// Filter datasources by team
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

	// Test team checking methods
	t.Run("IsTeamAllowed method works correctly", func(t *testing.T) {
		assert.True(t, datasources[0].IsTeamAllowed(1))
		assert.True(t, datasources[0].IsTeamAllowed(2))
		assert.True(t, datasources[0].IsTeamAllowed(3))

		assert.True(t, datasources[1].IsTeamAllowed(1))
		assert.False(t, datasources[1].IsTeamAllowed(2))
		assert.False(t, datasources[1].IsTeamAllowed(3))

		assert.True(t, datasources[2].IsTeamAllowed(1))
		assert.True(t, datasources[2].IsTeamAllowed(2))
		assert.False(t, datasources[2].IsTeamAllowed(3))

		assert.False(t, datasources[3].IsTeamAllowed(1))
		assert.False(t, datasources[3].IsTeamAllowed(2))
		assert.True(t, datasources[3].IsTeamAllowed(3))
	})
}