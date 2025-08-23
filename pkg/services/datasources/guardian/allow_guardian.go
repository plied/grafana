package guardian

import (
	"context"

	"github.com/grafana/grafana/pkg/apimachinery/identity"
	"github.com/grafana/grafana/pkg/services/datasources"
	"github.com/grafana/grafana/pkg/services/org"
	"github.com/grafana/grafana/pkg/services/team"
)

var _ DatasourceGuardian = new(AllowGuardian)

// AllowGuardian is used whenever an enterprise build is running without a license.
// It allows every one to Query all data sources and will not filter out any of them
type AllowGuardian struct{}

func (n AllowGuardian) CanQuery(datasourceID int64) (bool, error) {
	return true, nil
}

func (n AllowGuardian) FilterDatasourcesByReadPermissions(ds []*datasources.DataSource) ([]*datasources.DataSource, error) {
	return ds, nil
}

func (n AllowGuardian) FilterDatasourcesByQueryPermissions(ds []*datasources.DataSource) ([]*datasources.DataSource, error) {
	return ds, nil
}

var _ DatasourceGuardian = new(TeamBasedGuardian)

// TeamBasedGuardian implements team-based access control for datasources
type TeamBasedGuardian struct {
	user                identity.Requester
	orgID               int64
	dsService           datasources.DataSourceService
	teamIDsByUserGetter TeamIDsByUserGetter
}

func NewTeamBasedGuardian(user identity.Requester, orgID int64, dsService datasources.DataSourceService, teamService team.Service) *TeamBasedGuardian {
	return &TeamBasedGuardian{
		user:                user,
		orgID:               orgID,
		dsService:           dsService,
		teamIDsByUserGetter: teamService,
	}
}

func NewTeamBasedGuardianWithGetter(user identity.Requester, orgID int64, dsService datasources.DataSourceService, teamGetter TeamIDsByUserGetter) *TeamBasedGuardian {
	return &TeamBasedGuardian{
		user:                user,
		orgID:               orgID,
		dsService:           dsService,
		teamIDsByUserGetter: teamGetter,
	}
}

func (t *TeamBasedGuardian) CanQuery(datasourceID int64) (bool, error) {
	// Admins always have access to all datasources
	if t.getUserRole() == "Admin" {
		return true, nil
	}

	// Get the datasource to check its allowed teams
	ds, err := t.dsService.GetDataSource(context.Background(), &datasources.GetDataSourceQuery{
		ID:    datasourceID,
		OrgID: t.orgID,
	})
	if err != nil {
		return false, err
	}

	// If no teams are specified, all users have access
	if ds.AllowedTeams == "" {
		return true, nil
	}

	// Get user's team memberships
	userID, err := t.user.GetInternalID()
	if err != nil {
		return false, err
	}
	userTeams, err := t.teamIDsByUserGetter.GetTeamIDsByUser(context.Background(), &team.GetTeamIDsByUserQuery{
		OrgID:  t.orgID,
		UserID: userID,
	})
	if err != nil {
		return false, err
	}

	// Check if user is member of any allowed team
	for _, userTeam := range userTeams {
		if ds.IsTeamAllowed(userTeam) {
			return true, nil
		}
	}

	return false, nil
}

func (t *TeamBasedGuardian) FilterDatasourcesByReadPermissions(ds []*datasources.DataSource) ([]*datasources.DataSource, error) {
	return t.filterDatasourcesByTeam(ds), nil
}

func (t *TeamBasedGuardian) FilterDatasourcesByQueryPermissions(ds []*datasources.DataSource) ([]*datasources.DataSource, error) {
	return t.filterDatasourcesByTeam(ds), nil
}

func (t *TeamBasedGuardian) filterDatasourcesByTeam(ds []*datasources.DataSource) []*datasources.DataSource {
	// Admins always have access to all datasources
	if t.getUserRole() == "Admin" {
		return ds
	}

	// Get user's team memberships
	userID, err := t.user.GetInternalID()
	if err != nil {
		// If we can't get user ID, return empty to be safe
		return []*datasources.DataSource{}
	}
	userTeams, err := t.teamIDsByUserGetter.GetTeamIDsByUser(context.Background(), &team.GetTeamIDsByUserQuery{
		OrgID:  t.orgID,
		UserID: userID,
	})
	if err != nil {
		// If we can't get team info, return empty to be safe
		return []*datasources.DataSource{}
	}

	var filtered []*datasources.DataSource
	for _, dataSource := range ds {
		// If no teams are specified, all users have access
		if dataSource.AllowedTeams == "" {
			filtered = append(filtered, dataSource)
			continue
		}

		// Check if user is member of any allowed team
		for _, userTeam := range userTeams {
			if dataSource.IsTeamAllowed(userTeam) {
				filtered = append(filtered, dataSource)
				break
			}
		}
	}

	return filtered
}

func (t *TeamBasedGuardian) getUserRole() string {
	// Extract role from user identity. Default to lowest permission role if uncertain.
	if t.user.GetOrgRole() == org.RoleAdmin {
		return "Admin"
	} else if t.user.GetOrgRole() == org.RoleEditor {
		return "Editor"
	}
	return "Viewer"
}
