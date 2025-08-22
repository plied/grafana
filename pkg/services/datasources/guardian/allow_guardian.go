package guardian

import (
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
	user identity.Requester
}

func NewRoleBasedGuardian(user identity.Requester) *RoleBasedGuardian {
	return &RoleBasedGuardian{user: user}
}

func (r *RoleBasedGuardian) CanQuery(datasourceID int64) (bool, error) {
	// For now, always allow query. This could be extended to check specific datasource permissions
	return true, nil
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
