# Datasource Role-Based Access Control

This document describes the role-based access control feature for datasources in Grafana.

## Overview

The role-based access control feature allows administrators to restrict access to specific datasources based on user roles. This provides fine-grained control over which users can access which datasources within an organization.

## How It Works

### Roles
Grafana supports three organization roles:
- **Admin**: Full access to all datasources and administrative functions
- **Editor**: Can create and edit dashboards, limited datasource access based on restrictions
- **Viewer**: Read-only access to dashboards, limited datasource access based on restrictions

### Datasource Configuration
Each datasource can specify which roles are allowed to access it through the `allowedRoles` field:

- **Empty/null**: All roles can access the datasource (default behavior)
- **"Admin"**: Only administrators can access the datasource
- **"Editor,Admin"**: Editors and administrators can access the datasource
- **"Viewer,Editor,Admin"**: All roles can access the datasource (explicit)

## API Usage

### Creating a Datasource with Role Restrictions
```bash
POST /api/datasources
{
  "name": "Admin Only Database",
  "type": "postgres",
  "url": "localhost:5432",
  "database": "admin_db",
  "user": "admin",
  "allowedRoles": "Admin"
}
```

### Updating Datasource Role Restrictions
```bash
PUT /api/datasources/1
{
  "id": 1,
  "name": "Limited Access Database",
  "type": "postgres",
  "url": "localhost:5432", 
  "database": "shared_db",
  "allowedRoles": "Editor,Admin"
}
```

### Removing Role Restrictions
```bash
PUT /api/datasources/1
{
  "id": 1,
  "name": "Public Database",
  "type": "postgres",
  "url": "localhost:5432",
  "database": "public_db",
  "allowedRoles": ""
}
```

## Behavior

### Listing Datasources
When users list datasources (via `/api/datasources`), only datasources that their role is allowed to access will be returned. This filtering happens automatically based on the user's organization role.

### Individual Datasource Access
When accessing a specific datasource (by ID or UID), the system checks if the user's role is allowed. If not, access is denied.

### Dashboard Queries
Dashboard queries to restricted datasources will fail if the user's role is not allowed to access the datasource.

## Examples

### Admin-Only Datasource
```json
{
  "name": "Financial Database",
  "type": "mysql",
  "allowedRoles": "Admin"
}
```
Only organization administrators can see and query this datasource.

### Editor-Level Datasource
```json
{
  "name": "Analytics Database", 
  "type": "prometheus",
  "allowedRoles": "Editor,Admin"
}
```
Editors and administrators can access this datasource, but viewers cannot.

### Public Datasource
```json
{
  "name": "Public Metrics",
  "type": "influxdb", 
  "allowedRoles": ""
}
```
All users can access this datasource regardless of role.

## Migration

When upgrading to a version with this feature:
1. All existing datasources will have empty `allowedRoles` (no restrictions)
2. Access behavior remains the same until roles are explicitly configured
3. The feature is backward compatible

## Security Considerations

- Role checking is **case-sensitive**: "admin" ≠ "Admin"
- Empty `allowedRoles` grants access to all roles for backward compatibility
- Role restrictions are enforced at both list and individual access levels
- Roles are comma-separated with optional whitespace: "Editor, Admin" works
- Role restrictions complement but do not replace existing permission systems

## Implementation Details

The feature is implemented through:
1. Database schema addition of `allowed_roles` column
2. Enhanced datasource guardian system for role-based filtering  
3. Integration with existing access control mechanisms
4. Comprehensive test coverage for various scenarios

The implementation maintains backward compatibility and integrates seamlessly with Grafana's existing access control architecture.