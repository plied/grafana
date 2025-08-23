import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { DataSourceRoleAccess, Props } from './DataSourceRoleAccess';

const setup = (props?: Partial<Props>) => {
  const defaultProps: Props = {
    allowedRoles: '',
    onAllowedRolesChange: jest.fn(),
    disabled: false,
  };

  const user = userEvent.setup();
  return {
    user,
    ...render(<DataSourceRoleAccess {...defaultProps} {...props} />),
    onAllowedRolesChange: props?.onAllowedRolesChange || defaultProps.onAllowedRolesChange,
  };
};

describe('DataSourceRoleAccess', () => {
  it('should render the role access component', () => {
    setup();

    expect(screen.getByText('Role Access')).toBeInTheDocument();
    expect(screen.getByText('All roles (no restrictions)')).toBeInTheDocument();
  });

  it('should display selected roles when allowedRoles is provided', () => {
    setup({ allowedRoles: 'Admin,Editor' });

    expect(screen.getByDisplayValue('Admin')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Editor')).toBeInTheDocument();
  });

  it('should call onAllowedRolesChange when roles are selected', async () => {
    const onAllowedRolesChange = jest.fn();
    const { user } = setup({ onAllowedRolesChange });

    const multiSelect = screen.getByRole('combobox');
    await user.click(multiSelect);

    const adminOption = screen.getByText('Admin');
    await user.click(adminOption);

    expect(onAllowedRolesChange).toHaveBeenCalledWith('Admin');
  });

  it('should handle multiple role selections', async () => {
    const onAllowedRolesChange = jest.fn();
    const { user } = setup({ onAllowedRolesChange });

    const multiSelect = screen.getByRole('combobox');
    await user.click(multiSelect);

    const adminOption = screen.getByText('Admin');
    await user.click(adminOption);

    await user.click(multiSelect);
    const editorOption = screen.getByText('Editor');
    await user.click(editorOption);

    // Should be called with comma-separated roles
    expect(onAllowedRolesChange).toHaveBeenLastCalledWith('Admin,Editor');
  });

  it('should be disabled when disabled prop is true', () => {
    setup({ disabled: true });

    const multiSelect = screen.getByRole('combobox');
    expect(multiSelect).toBeDisabled();
  });

  it('should parse existing allowedRoles correctly', () => {
    setup({ allowedRoles: 'Viewer, Editor , Admin' });

    // Should handle whitespace correctly
    expect(screen.getByDisplayValue('Viewer')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Editor')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Admin')).toBeInTheDocument();
  });

  it('should handle empty allowedRoles', () => {
    setup({ allowedRoles: '' });

    // Should show placeholder when no roles are selected
    expect(screen.getByText('All roles (no restrictions)')).toBeInTheDocument();
  });
});
