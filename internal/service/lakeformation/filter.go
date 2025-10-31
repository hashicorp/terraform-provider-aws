// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
)

func filterCatalogPermissions(principalIdentifier string) PermissionsFilter {
	return func(permissions awstypes.PrincipalResourcePermissions) bool {
		return principalIdentifier == aws.ToString(permissions.Principal.DataLakePrincipalIdentifier) && permissions.Resource.Catalog != nil
	}
}

func filterDataCellsFilter(principalIdentifier string) PermissionsFilter {
	return func(permissions awstypes.PrincipalResourcePermissions) bool {
		return principalIdentifier == aws.ToString(permissions.Principal.DataLakePrincipalIdentifier) && permissions.Resource.DataCellsFilter != nil
	}
}

func filterDataLocationPermissions(principalIdentifier string) PermissionsFilter {
	return func(permissions awstypes.PrincipalResourcePermissions) bool {
		return principalIdentifier == aws.ToString(permissions.Principal.DataLakePrincipalIdentifier) && permissions.Resource.DataLocation != nil
	}
}

func filterDatabasePermissions(principalIdentifier string) PermissionsFilter {
	return func(permissions awstypes.PrincipalResourcePermissions) bool {
		return principalIdentifier == aws.ToString(permissions.Principal.DataLakePrincipalIdentifier) && permissions.Resource.Database != nil
	}
}

func filterLFTagPermissions(principalIdentifier string) PermissionsFilter {
	return func(permissions awstypes.PrincipalResourcePermissions) bool {
		return principalIdentifier == aws.ToString(permissions.Principal.DataLakePrincipalIdentifier) && permissions.Resource.LFTag != nil
	}
}

func filterLFTagPolicyPermissions(principalIdentifier string) PermissionsFilter {
	return func(permissions awstypes.PrincipalResourcePermissions) bool {
		return principalIdentifier == aws.ToString(permissions.Principal.DataLakePrincipalIdentifier) && permissions.Resource.LFTagPolicy != nil
	}
}

func filterTablePermissions(principalIdentifier string, table *awstypes.TableResource) PermissionsFilter {
	// CREATE PERMS (in)     = ALL, ALTER, DELETE, DESCRIBE, DROP, INSERT, SELECT on Table, Name = (Table Name)
	//      LIST PERMS (out) = ALL, ALTER, DELETE, DESCRIBE, DROP, INSERT         on Table, Name = (Table Name)
	//      LIST PERMS (out) = SELECT                                             on TableWithColumns, Name = (Table Name), ColumnWildcard

	// CREATE PERMS (in)       = ALL, ALTER, DELETE, DESCRIBE, DROP, INSERT, SELECT on Table, TableWildcard
	//        LIST PERMS (out) = ALL, ALTER, DELETE, DESCRIBE, DROP, INSERT         on Table, TableWildcard, Name = ALL_TABLES
	//        LIST PERMS (out) = SELECT                                             on TableWithColumns, Name = ALL_TABLES, ColumnWildcard

	return func(perm awstypes.PrincipalResourcePermissions) bool {
		if principalIdentifier != aws.ToString(perm.Principal.DataLakePrincipalIdentifier) {
			return false
		}

		if perm.Resource.TableWithColumns != nil && perm.Resource.TableWithColumns.ColumnWildcard != nil {
			if aws.ToString(perm.Resource.TableWithColumns.Name) == aws.ToString(table.Name) || (table.TableWildcard != nil && aws.ToString(perm.Resource.TableWithColumns.Name) == TableNameAllTables) {
				if len(perm.Permissions) > 0 && perm.Permissions[0] == awstypes.PermissionSelect {
					return true
				}

				if len(perm.PermissionsWithGrantOption) > 0 && perm.PermissionsWithGrantOption[0] == awstypes.PermissionSelect {
					return true
				}
			}
		}

		if perm.Resource.Table != nil && aws.ToString(perm.Resource.Table.DatabaseName) == aws.ToString(table.DatabaseName) {
			if aws.ToString(perm.Resource.Table.Name) == aws.ToString(table.Name) {
				return true
			}

			if perm.Resource.Table.TableWildcard != nil && table.TableWildcard != nil {
				return true
			}
		}

		return false
	}
}

func filterTableWithColumnsPermissions(principalIdentifier string, twc *awstypes.TableResource, columnNames []string, excludedColumnNames []string, columnWildcard bool) PermissionsFilter {
	// CREATE PERMS (in)       = ALL, ALTER, DELETE, DESCRIBE, DROP, INSERT, SELECT on TableWithColumns, Name = (Table Name), ColumnWildcard
	//        LIST PERMS (out) = ALL, ALTER, DELETE, DESCRIBE, DROP, INSERT         on Table, Name = (Table Name)
	//        LIST PERMS (out) = SELECT                                             on TableWithColumns, Name = (Table Name), ColumnWildcard

	return func(perm awstypes.PrincipalResourcePermissions) bool {
		if principalIdentifier != aws.ToString(perm.Principal.DataLakePrincipalIdentifier) {
			return false
		}

		if perm.Resource.TableWithColumns != nil && perm.Resource.TableWithColumns.ColumnNames != nil {
			if stringSlicesEqualIgnoreOrder(perm.Resource.TableWithColumns.ColumnNames, columnNames) {
				return true
			}
		}

		if perm.Resource.TableWithColumns != nil && perm.Resource.TableWithColumns.ColumnWildcard != nil && (columnWildcard || len(excludedColumnNames) > 0) {
			if perm.Resource.TableWithColumns.ColumnWildcard.ExcludedColumnNames == nil && len(excludedColumnNames) == 0 {
				return true
			}

			if len(excludedColumnNames) > 0 && stringSlicesEqualIgnoreOrder(perm.Resource.TableWithColumns.ColumnWildcard.ExcludedColumnNames, excludedColumnNames) {
				return true
			}
		}

		if perm.Resource.Table != nil && aws.ToString(perm.Resource.Table.DatabaseName) == aws.ToString(twc.DatabaseName) {
			if aws.ToString(perm.Resource.Table.Name) == aws.ToString(twc.Name) {
				return true
			}
		}

		return false
	}
}
