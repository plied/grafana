package guardian

import (
	"context"

	"github.com/grafana/grafana/pkg/apimachinery/identity"
	"github.com/grafana/grafana/pkg/services/datasources"
	"github.com/grafana/grafana/pkg/services/org"
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

var _ DatasourceGuardian = new(RoleBasedGuardian)

// RoleBasedGuardian implements role-based access control for datasources
type RoleBasedGuardian struct {
	user      identity.Requester
	orgID     int64
	dsService datasources.DataSourceService
}

func NewRoleBasedGuardian(user identity.Requester, orgID int64, dsService datasources.DataSourceService) *RoleBasedGuardian {
	return &RoleBasedGuardian{
		user:      user,
		orgID:     orgID,
		dsService: dsService,
	}
}

func (r *RoleBasedGuardian) CanQuery(datasourceID int64) (bool, error) {
	// Get the datasource to check its allowed roles
	ds, err := r.dsService.GetDataSource(context.Background(), &datasources.GetDataSourceQuery{
		ID:    datasourceID,
		OrgID: r.orgID,
	})
	if err != nil {
		return false, err
	}

	// Check if user's role is allowed to access this datasource
	userRole := r.getUserRole()
	return ds.IsRoleAllowed(userRole), nil
}

func (r *RoleBasedGuardian) FilterDatasourcesByReadPermissions(ds []*datasources.DataSource) ([]*datasources.DataSource, error) {
	return r.filterDatasourcesByRole(ds), nil
}

func (r *RoleBasedGuardian) FilterDatasourcesByQueryPermissions(ds []*datasources.DataSource) ([]*datasources.DataSource, error) {
	return r.filterDatasourcesByRole(ds), nil
}

func (r *RoleBasedGuardian) filterDatasourcesByRole(ds []*datasources.DataSource) []*datasources.DataSource {
	userRole := r.getUserRole()
	
	var filtered []*datasources.DataSource
	for _, dataSource := range ds {
		if dataSource.IsRoleAllowed(userRole) {
			filtered = append(filtered, dataSource)
		}
	}
	
	return filtered
}

func (r *RoleBasedGuardian) getUserRole() string {
	// Extract role from user identity. Default to lowest permission role if uncertain.
	if r.user.GetOrgRole() == org.RoleAdmin {
		return "Admin"
	} else if r.user.GetOrgRole() == org.RoleEditor {
		return "Editor"
	}
	return "Viewer"
}
