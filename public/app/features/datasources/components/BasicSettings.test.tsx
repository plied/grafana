import { render, screen } from '@testing-library/react';

import { selectors } from '@grafana/e2e-selectors';

import { BasicSettings, Props } from './BasicSettings';

// Mock the DataSourceTeamAccess component since it has external dependencies
jest.mock('./DataSourceTeamAccess', () => ({
  DataSourceTeamAccess: () => <div data-testid="team-access-component">Team Access</div>,
}));

const setup = () => {
  const props: Props = {
    dataSourceName: 'Graphite',
    isDefault: false,
    allowedTeams: '',
    onDefaultChange: jest.fn(),
    onNameChange: jest.fn(),
    onAllowedTeamsChange: jest.fn(),
  };

  return render(<BasicSettings {...props} />);
};

describe('<BasicSettings>', () => {
  it('should render component', () => {
    setup();

    expect(screen.getByTestId(selectors.pages.DataSource.name)).toBeInTheDocument();
    expect(screen.getByLabelText(/Default/)).toBeInTheDocument();
    expect(screen.getByTestId('team-access-component')).toBeInTheDocument();
  });
});
