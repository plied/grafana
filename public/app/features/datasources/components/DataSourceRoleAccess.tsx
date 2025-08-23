import { OrgRole } from '@grafana/data';
import { selectors } from '@grafana/e2e-selectors';
import { t } from '@grafana/i18n';
import { InlineField, MultiSelect } from '@grafana/ui';

export interface Props {
  allowedRoles: string;
  onAllowedRolesChange: (allowedRoles: string) => void;
  disabled?: boolean;
}

const roleOptions = [
  { label: 'Viewer', value: OrgRole.Viewer },
  { label: 'Editor', value: OrgRole.Editor },
  { label: 'Admin', value: OrgRole.Admin },
];

export function DataSourceRoleAccess({ allowedRoles, onAllowedRolesChange, disabled }: Props) {
  const selectedRoles = allowedRoles
    ? allowedRoles.split(',').map((role) => role.trim()).filter(Boolean)
    : [];

  const selectedValues = selectedRoles.map((role) => ({ label: role, value: role }));

  const handleRoleChange = (values: Array<{ label: string; value: string }>) => {
    const roles = values.map((item) => item.value).join(',');
    onAllowedRolesChange(roles);
  };

  return (
    <div className="gf-form-group">
      <div className="gf-form-inline">
        <InlineField
          label={t('datasources.role-access.label', 'Access')}
          tooltip={t(
            'datasources.role-access.tooltip',
            'Restrict access to this datasource to specific organization roles. Leave empty to allow all roles to access this datasource.'
          )}
          grow
          disabled={disabled}
          labelWidth={14}
        >
          <MultiSelect
            options={roleOptions}
            value={selectedValues}
            onChange={handleRoleChange}
            placeholder={t('datasources.role-access.placeholder', 'All roles (no restrictions)')}
            disabled={disabled}
            width={45}
            data-testid={selectors.pages.DataSource.roleAccess}
          />
        </InlineField>
      </div>
    </div>
  );
}

