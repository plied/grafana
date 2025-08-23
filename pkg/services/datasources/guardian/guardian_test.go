package guardian

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/services/datasources"
	fakeDS "github.com/grafana/grafana/pkg/services/datasources/fakes"
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

func TestTeamBasedGuardian_FilterDatasourcesByTeam(t *testing.T) {
	tests := []struct {
		name         string
		userRole     org.RoleType
		userID       int64
		userTeams    []int64
		datasources  []*datasources.DataSource
		expectedLen  int
		expectedUIDs []string
	}{
		{
			name:     "admin user sees all datasources",
			userRole: org.RoleAdmin,
			userID:   1,
			userTeams: []int64{1, 2},
			datasources: []*datasources.DataSource{
				{UID: "ds1", AllowedTeams: ""},
				{UID: "ds2", AllowedTeams: "1"},
				{UID: "ds3", AllowedTeams: "2,3"},
			},
			expectedLen:  3,
			expectedUIDs: []string{"ds1", "ds2", "ds3"},
		},
		{
			name:     "editor user sees team-accessible datasources",
			userRole: org.RoleEditor,
			userID:   2,
			userTeams: []int64{1, 3},
			datasources: []*datasources.DataSource{
				{UID: "ds1", AllowedTeams: ""},
				{UID: "ds2", AllowedTeams: "1"},
				{UID: "ds3", AllowedTeams: "2"},
				{UID: "ds4", AllowedTeams: "3"},
			},
			expectedLen:  3,
			expectedUIDs: []string{"ds1", "ds2", "ds4"},
		},
		{
			name:     "viewer user with no teams sees only unrestricted datasources",
			userRole: org.RoleViewer,
			userID:   3,
			userTeams: []int64{},
			datasources: []*datasources.DataSource{
				{UID: "ds1", AllowedTeams: ""},
				{UID: "ds2", AllowedTeams: "1"},
				{UID: "ds3", AllowedTeams: "2"},
			},
			expectedLen:  1,
			expectedUIDs: []string{"ds1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &user.SignedInUser{
				UserID:  tt.userID,
				OrgRole: tt.userRole,
			}
			fakeDSService := &fakeDS.FakeDataSourceService{}
			mockTeamSvc := &mockTeamIDsByUserGetter{
				userTeams: map[int64][]int64{
					tt.userID: tt.userTeams,
				},
			}
			guardian := NewTeamBasedGuardianWithGetter(user, 1, fakeDSService, mockTeamSvc)

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
	fakeDSService := &fakeDS.FakeDataSourceService{}
	mockTeamSvc := &mockTeamIDsByUserGetter{userTeams: map[int64][]int64{}}
	provider := &OSSProvider{dsService: fakeDSService, teamService: mockTeamSvc}
	user := &user.SignedInUser{OrgRole: org.RoleViewer}

	t.Run("always returns TeamBasedGuardian", func(t *testing.T) {
		datasources := []datasources.DataSource{
			{UID: "ds1", AllowedTeams: ""},
			{UID: "ds2", AllowedTeams: ""},
		}

		guardian := provider.New(1, user, datasources...)
		_, isTeamBasedGuardian := guardian.(*TeamBasedGuardian)
		assert.True(t, isTeamBasedGuardian)
	})

	t.Run("returns TeamBasedGuardian when team restrictions exist", func(t *testing.T) {
		datasources := []datasources.DataSource{
			{UID: "ds1", AllowedTeams: ""},
			{UID: "ds2", AllowedTeams: "1"},
		}

		guardian := provider.New(1, user, datasources...)
		_, isTeamBasedGuardian := guardian.(*TeamBasedGuardian)
		assert.True(t, isTeamBasedGuardian)
	})
}

func TestTeamBasedGuardian_CanQuery(t *testing.T) {
	tests := []struct {
		name       string
		userRole   org.RoleType
		userID     int64
		userTeams  []int64
		datasource *datasources.DataSource
		expected   bool
	}{
		{
			name:     "admin can query team-restricted datasource",
			userRole: org.RoleAdmin,
			userID:   1,
			userTeams: []int64{},
			datasource: &datasources.DataSource{
				ID:           1,
				OrgID:        1,
				AllowedTeams: "1",
			},
			expected: true,
		},
		{
			name:     "user with team access can query team-restricted datasource",
			userRole: org.RoleEditor,
			userID:   2,
			userTeams: []int64{1, 2},
			datasource: &datasources.DataSource{
				ID:           1,
				OrgID:        1,
				AllowedTeams: "1",
			},
			expected: true,
		},
		{
			name:     "user without team access cannot query team-restricted datasource",
			userRole: org.RoleEditor,
			userID:   3,
			userTeams: []int64{2, 3},
			datasource: &datasources.DataSource{
				ID:           1,
				OrgID:        1,
				AllowedTeams: "1",
			},
			expected: false,
		},
		{
			name:     "any user can query unrestricted datasource",
			userRole: org.RoleViewer,
			userID:   4,
			userTeams: []int64{},
			datasource: &datasources.DataSource{
				ID:           1,
				OrgID:        1,
				AllowedTeams: "",
			},
			expected: true,
		},
		{
			name:     "user can query datasource with multiple allowed teams",
			userRole: org.RoleEditor,
			userID:   5,
			userTeams: []int64{3},
			datasource: &datasources.DataSource{
				ID:           1,
				OrgID:        1,
				AllowedTeams: "1,2,3",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &user.SignedInUser{
				UserID:  tt.userID,
				OrgID:   1,
				OrgRole: tt.userRole,
			}
			
			fakeDSService := &fakeDS.FakeDataSourceService{
				DataSources: []*datasources.DataSource{tt.datasource},
			}
			
			mockTeamSvc := &mockTeamIDsByUserGetter{
				userTeams: map[int64][]int64{
					tt.userID: tt.userTeams,
				},
			}
			
			guardian := NewTeamBasedGuardianWithGetter(user, 1, fakeDSService, mockTeamSvc)

			canQuery, err := guardian.CanQuery(tt.datasource.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, canQuery)
		})
	}
}