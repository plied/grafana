package guardian

import (
	"github.com/grafana/grafana/pkg/apimachinery/identity"
	"github.com/grafana/grafana/pkg/services/datasources"
)

type DatasourceGuardianProvider interface {
	New(orgID int64, user identity.Requester, dataSources ...datasources.DataSource) DatasourceGuardian
}

type DatasourceGuardian interface {
	CanQuery(datasourceID int64) (bool, error)
	FilterDatasourcesByReadPermissions([]*datasources.DataSource) ([]*datasources.DataSource, error)
	FilterDatasourcesByQueryPermissions([]*datasources.DataSource) ([]*datasources.DataSource, error)
}

func ProvideGuardian() *OSSProvider {
	return &OSSProvider{}
}

type OSSProvider struct{}

func (p *OSSProvider) New(orgID int64, user identity.Requester, dataSources ...datasources.DataSource) DatasourceGuardian {
	// Check if any datasource has role restrictions
	for _, ds := range dataSources {
		if ds.AllowedRoles != "" {
			return NewRoleBasedGuardian(user)
		}
	}
	// Default to allowing all access if no role restrictions are set
	return &AllowGuardian{}
}
