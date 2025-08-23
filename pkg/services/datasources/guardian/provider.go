package guardian

import (
	"context"
	
	"github.com/grafana/grafana/pkg/apimachinery/identity"
	"github.com/grafana/grafana/pkg/services/datasources"
	"github.com/grafana/grafana/pkg/services/team"
)

type DatasourceGuardianProvider interface {
	New(orgID int64, user identity.Requester, dataSources ...datasources.DataSource) DatasourceGuardian
}

type DatasourceGuardian interface {
	CanQuery(datasourceID int64) (bool, error)
	FilterDatasourcesByReadPermissions([]*datasources.DataSource) ([]*datasources.DataSource, error)
	FilterDatasourcesByQueryPermissions([]*datasources.DataSource) ([]*datasources.DataSource, error)
}

// TeamIDsByUserGetter is a minimal interface for getting team IDs by user
type TeamIDsByUserGetter interface {
	GetTeamIDsByUser(ctx context.Context, query *team.GetTeamIDsByUserQuery) ([]int64, error)
}

func ProvideGuardian(dsService datasources.DataSourceService, teamService team.Service) *OSSProvider {
	return &OSSProvider{dsService: dsService, teamService: teamService}
}

type OSSProvider struct {
	dsService   datasources.DataSourceService
	teamService TeamIDsByUserGetter
}

func (p *OSSProvider) New(orgID int64, user identity.Requester, dataSources ...datasources.DataSource) DatasourceGuardian {
	// Always use team-based guardian as it can handle both restricted and unrestricted datasources
	return NewTeamBasedGuardianWithGetter(user, orgID, p.dsService, p.teamService)
}
