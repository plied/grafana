import { SelectableValue } from '@grafana/data';
import { selectors } from '@grafana/e2e-selectors';
import { t } from '@grafana/i18n';
import { getBackendSrv } from '@grafana/runtime';
import { InlineField, MultiSelect } from '@grafana/ui';
import { Component } from 'react';
import { Team } from 'app/types/teams';

export interface Props {
  allowedTeams: string;
  onAllowedTeamsChange: (allowedTeams: string) => void;
  disabled?: boolean;
}

interface State {
  teamOptions: Array<SelectableValue<string>>;
  isLoading: boolean;
}

export class DataSourceTeamAccess extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      teamOptions: [],
      isLoading: false,
    };
  }

  componentDidMount() {
    this.loadTeams();
  }

  loadTeams = async () => {
    this.setState({ isLoading: true });
    try {
      const response = await getBackendSrv().get('/api/teams/search?perpage=100&page=1');
      const teamOptions = response.teams.map((team: Team) => ({
        label: team.name,
        value: team.id.toString(),
      }));
      this.setState({ teamOptions, isLoading: false });
    } catch (error) {
      console.error('Failed to load teams:', error);
      this.setState({ teamOptions: [], isLoading: false });
    }
  };

  render() {
    const { allowedTeams, onAllowedTeamsChange, disabled } = this.props;
    const { teamOptions, isLoading } = this.state;
    
    const selectedTeams = allowedTeams
      ? allowedTeams.split(',').map((team) => team.trim()).filter(Boolean)
      : [];

    const selectedValues = selectedTeams.map((teamId) => {
      const team = teamOptions.find((option) => option.value === teamId);
      return team || { label: teamId, value: teamId };
    });

    const handleTeamChange = (values: Array<SelectableValue<string>>) => {
      const teams = values.map((item) => item.value!).join(',');
      onAllowedTeamsChange(teams);
    };

    return (
      <div className="gf-form-group">
        <div className="gf-form-inline">
          <InlineField
            label={t('datasources.team-access.label', 'Team')}
            tooltip={t(
              'datasources.team-access.tooltip',
              'Restrict access to this datasource to specific teams. Leave empty to allow all users to access this datasource.'
            )}
            grow
            disabled={disabled}
            labelWidth={14}
          >
            <MultiSelect
              options={teamOptions}
              value={selectedValues}
              onChange={handleTeamChange}
              placeholder={t('datasources.team-access.placeholder', 'All teams (no restrictions)')}
              disabled={disabled || isLoading}
              width={45}
              data-testid={selectors.pages.DataSource.teamAccess}
            />
          </InlineField>
        </div>
      </div>
    );
  }
}

