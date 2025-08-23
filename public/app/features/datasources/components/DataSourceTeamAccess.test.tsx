import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { getBackendSrv } from '@grafana/runtime';

import { DataSourceTeamAccess, Props } from './DataSourceTeamAccess';

// Mock the backend service
jest.mock('@grafana/runtime', () => ({
  getBackendSrv: jest.fn(),
}));

const mockTeams = [
  { id: 1, name: 'Team Alpha' },
  { id: 2, name: 'Team Beta' },
  { id: 3, name: 'Team Gamma' },
];

const setup = (props?: Partial<Props>) => {
  const defaultProps: Props = {
    allowedTeams: '',
    onAllowedTeamsChange: jest.fn(),
    disabled: false,
  };

  // Mock the teams API call
  (getBackendSrv as jest.Mock).mockReturnValue({
    get: jest.fn().mockResolvedValue({ teams: mockTeams }),
  });

  const user = userEvent.setup();
  return {
    user,
    ...render(<DataSourceTeamAccess {...defaultProps} {...props} />),
    onAllowedTeamsChange: props?.onAllowedTeamsChange || defaultProps.onAllowedTeamsChange,
  };
};

describe('DataSourceTeamAccess', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should render the team access component', async () => {
    setup();

    expect(screen.getByText('Team Access')).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByText('All teams (no restrictions)')).toBeInTheDocument();
    });
  });

  it('should load teams on mount', async () => {
    setup();

    await waitFor(() => {
      expect(getBackendSrv().get).toHaveBeenCalledWith('/api/teams/search?perpage=100&page=1');
    });
  });

  it('should display selected teams when allowedTeams is provided', async () => {
    setup({ allowedTeams: '1,2' });

    await waitFor(() => {
      expect(screen.getByDisplayValue('Team Alpha')).toBeInTheDocument();
      expect(screen.getByDisplayValue('Team Beta')).toBeInTheDocument();
    });
  });

  it('should call onAllowedTeamsChange when teams are selected', async () => {
    const onAllowedTeamsChange = jest.fn();
    const { user } = setup({ onAllowedTeamsChange });

    await waitFor(() => {
      expect(screen.getByRole('combobox')).toBeInTheDocument();
    });

    const multiSelect = screen.getByRole('combobox');
    await user.click(multiSelect);

    await waitFor(() => {
      expect(screen.getByText('Team Alpha')).toBeInTheDocument();
    });

    const teamAlphaOption = screen.getByText('Team Alpha');
    await user.click(teamAlphaOption);

    expect(onAllowedTeamsChange).toHaveBeenCalledWith('1');
  });

  it('should handle multiple team selections', async () => {
    const onAllowedTeamsChange = jest.fn();
    const { user } = setup({ onAllowedTeamsChange });

    await waitFor(() => {
      expect(screen.getByRole('combobox')).toBeInTheDocument();
    });

    const multiSelect = screen.getByRole('combobox');
    
    // Select first team
    await user.click(multiSelect);
    await waitFor(() => {
      expect(screen.getByText('Team Alpha')).toBeInTheDocument();
    });
    await user.click(screen.getByText('Team Alpha'));

    // Select second team
    await user.click(multiSelect);
    await waitFor(() => {
      expect(screen.getByText('Team Beta')).toBeInTheDocument();
    });
    await user.click(screen.getByText('Team Beta'));

    // Should be called with comma-separated team IDs
    expect(onAllowedTeamsChange).toHaveBeenLastCalledWith('1,2');
  });

  it('should be disabled when disabled prop is true', async () => {
    setup({ disabled: true });

    await waitFor(() => {
      const multiSelect = screen.getByRole('combobox');
      expect(multiSelect).toBeDisabled();
    });
  });

  it('should parse existing allowedTeams correctly', async () => {
    setup({ allowedTeams: '1, 2 , 3' });

    await waitFor(() => {
      // Should handle whitespace correctly
      expect(screen.getByDisplayValue('Team Alpha')).toBeInTheDocument();
      expect(screen.getByDisplayValue('Team Beta')).toBeInTheDocument();
      expect(screen.getByDisplayValue('Team Gamma')).toBeInTheDocument();
    });
  });

  it('should handle empty allowedTeams', async () => {
    setup({ allowedTeams: '' });

    await waitFor(() => {
      // Should show placeholder when no teams are selected
      expect(screen.getByText('All teams (no restrictions)')).toBeInTheDocument();
    });
  });

  it('should handle teams API error gracefully', async () => {
    const onAllowedTeamsChange = jest.fn();
    
    // Mock API error
    (getBackendSrv as jest.Mock).mockReturnValue({
      get: jest.fn().mockRejectedValue(new Error('API Error')),
    });

    const { user } = setup({ onAllowedTeamsChange });

    await waitFor(() => {
      expect(screen.getByRole('combobox')).toBeInTheDocument();
    });

    // Should still be able to interact with the component even if teams failed to load
    const multiSelect = screen.getByRole('combobox');
    await user.click(multiSelect);

    // Should show placeholder
    expect(screen.getByText('All teams (no restrictions)')).toBeInTheDocument();
  });
});
