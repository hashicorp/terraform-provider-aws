// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestFilterPermissions(t *testing.T) {
	t.Parallel()

	// primitives to make test cases easier
	accountID := "481516234248"
	dbName := "Hiliji"
	altDBName := "Hiuhbum"
	tableName := "Ladocmoc"

	principal := &awstypes.DataLakePrincipal{
		//lintignore:AWSAT005
		DataLakePrincipalIdentifier: aws.String(fmt.Sprintf("arn:aws-us-gov:iam::%s:role/Zepotiz-Bulgaria", accountID)),
	}

	testCases := []struct {
		Name                string
		Input               *lakeformation.ListPermissionsInput
		TableType           string
		ColumnNames         []string
		ExcludedColumnNames []string
		ColumnWildcard      bool
		All                 []awstypes.PrincipalResourcePermissions
		ExpectedClean       []awstypes.PrincipalResourcePermissions
	}{
		{
			Name: "empty",
			Input: &lakeformation.ListPermissionsInput{
				Principal: principal,
				Resource:  &awstypes.Resource{},
			},
			All:           nil,
			ExpectedClean: nil,
		},
		{
			Name: "emptyWithInput",
			Input: &lakeformation.ListPermissionsInput{
				Principal: principal,
				Resource: &awstypes.Resource{
					Table: &awstypes.TableResource{
						CatalogId:    aws.String(accountID),
						DatabaseName: aws.String(dbName),
						Name:         aws.String(tableName),
					},
				},
			},
			All:           nil,
			ExpectedClean: nil,
		},
		{
			Name: "wrongTableResource", // this may not actually be possible but we account for it
			Input: &lakeformation.ListPermissionsInput{
				Principal: principal,
				Resource: &awstypes.Resource{
					Table: &awstypes.TableResource{
						CatalogId:    aws.String(accountID),
						DatabaseName: aws.String(dbName),
						Name:         aws.String(tableName),
					},
				},
			},
			All: []awstypes.PrincipalResourcePermissions{
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionSelect},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						Table: &awstypes.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(altDBName),
							Name:         aws.String(tableName),
						},
					},
				},
			},
			ExpectedClean: nil,
		},
		{
			Name: "tableResource",
			Input: &lakeformation.ListPermissionsInput{
				Principal: principal,
				Resource: &awstypes.Resource{
					Table: &awstypes.TableResource{
						CatalogId:    aws.String(accountID),
						DatabaseName: aws.String(dbName),
						Name:         aws.String(tableName),
					},
				},
			},
			All: []awstypes.PrincipalResourcePermissions{
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionSelect},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						Table: &awstypes.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
			},
			ExpectedClean: []awstypes.PrincipalResourcePermissions{
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionSelect},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						Table: &awstypes.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
			},
		},
		{
			Name: "tableResourceSelectPerm",
			Input: &lakeformation.ListPermissionsInput{
				Principal: principal,
				Resource: &awstypes.Resource{
					Table: &awstypes.TableResource{
						CatalogId:    aws.String(accountID),
						DatabaseName: aws.String(dbName),
						Name:         aws.String(tableName),
					},
				},
			},
			All: []awstypes.PrincipalResourcePermissions{
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionAlter, awstypes.PermissionDelete},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						Table: &awstypes.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionSelect},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						TableWithColumns: &awstypes.TableWithColumnsResource{
							CatalogId:      aws.String(accountID),
							DatabaseName:   aws.String(dbName),
							Name:           aws.String(tableName),
							ColumnWildcard: &awstypes.ColumnWildcard{},
						},
					},
				},
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionAlter},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						TableWithColumns: &awstypes.TableWithColumnsResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
			},
			ExpectedClean: []awstypes.PrincipalResourcePermissions{
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionAlter, awstypes.PermissionDelete},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						Table: &awstypes.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionSelect},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						TableWithColumns: &awstypes.TableWithColumnsResource{
							CatalogId:      aws.String(accountID),
							DatabaseName:   aws.String(dbName),
							Name:           aws.String(tableName),
							ColumnWildcard: &awstypes.ColumnWildcard{},
						},
					},
				},
			},
		},
		{
			Name: "tableResourceSelectPermGrant",
			Input: &lakeformation.ListPermissionsInput{
				Principal: principal,
				Resource: &awstypes.Resource{
					Table: &awstypes.TableResource{
						CatalogId:    aws.String(accountID),
						DatabaseName: aws.String(dbName),
						Name:         aws.String(tableName),
					},
				},
			},
			All: []awstypes.PrincipalResourcePermissions{
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionAlter, awstypes.PermissionDelete},
					PermissionsWithGrantOption: []awstypes.Permission{awstypes.PermissionAlter, awstypes.PermissionDelete},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						Table: &awstypes.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionSelect},
					PermissionsWithGrantOption: []awstypes.Permission{awstypes.PermissionSelect},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						TableWithColumns: &awstypes.TableWithColumnsResource{
							CatalogId:      aws.String(accountID),
							DatabaseName:   aws.String(dbName),
							Name:           aws.String(tableName),
							ColumnWildcard: &awstypes.ColumnWildcard{},
						},
					},
				},
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionAlter},
					PermissionsWithGrantOption: []awstypes.Permission{awstypes.PermissionAlter},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						TableWithColumns: &awstypes.TableWithColumnsResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
			},
			ExpectedClean: []awstypes.PrincipalResourcePermissions{
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionAlter, awstypes.PermissionDelete},
					PermissionsWithGrantOption: []awstypes.Permission{awstypes.PermissionAlter, awstypes.PermissionDelete},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						Table: &awstypes.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionSelect},
					PermissionsWithGrantOption: []awstypes.Permission{awstypes.PermissionSelect},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						TableWithColumns: &awstypes.TableWithColumnsResource{
							CatalogId:      aws.String(accountID),
							DatabaseName:   aws.String(dbName),
							Name:           aws.String(tableName),
							ColumnWildcard: &awstypes.ColumnWildcard{},
						},
					},
				},
			},
		},
		{
			Name: "twcBasic",
			Input: &lakeformation.ListPermissionsInput{
				Principal: principal,
				Resource: &awstypes.Resource{
					Table: &awstypes.TableResource{
						CatalogId:    aws.String(accountID),
						DatabaseName: aws.String(dbName),
						Name:         aws.String(tableName),
					},
				},
			},
			TableType:   tflakeformation.TableTypeTableWithColumns,
			ColumnNames: []string{names.AttrValue},
			All: []awstypes.PrincipalResourcePermissions{
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionAlter, awstypes.PermissionDelete},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						Table: &awstypes.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionSelect},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						TableWithColumns: &awstypes.TableWithColumnsResource{
							CatalogId:    aws.String(accountID),
							ColumnNames:  []string{names.AttrValue},
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionSelect},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						TableWithColumns: &awstypes.TableWithColumnsResource{
							CatalogId:    aws.String(accountID),
							ColumnNames:  []string{"fred"},
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
			},
			ExpectedClean: []awstypes.PrincipalResourcePermissions{
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionAlter, awstypes.PermissionDelete},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						Table: &awstypes.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionSelect},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						TableWithColumns: &awstypes.TableWithColumnsResource{
							CatalogId:    aws.String(accountID),
							ColumnNames:  []string{names.AttrValue},
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
			},
		},
		{
			Name: "twcWildcard",
			Input: &lakeformation.ListPermissionsInput{
				Principal: principal,
				Resource: &awstypes.Resource{
					Table: &awstypes.TableResource{
						CatalogId:    aws.String(accountID),
						DatabaseName: aws.String(dbName),
						Name:         aws.String(tableName),
					},
				},
			},
			TableType:      tflakeformation.TableTypeTableWithColumns,
			ColumnWildcard: true,
			All: []awstypes.PrincipalResourcePermissions{
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionAlter, awstypes.PermissionDelete},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						Table: &awstypes.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionSelect},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						TableWithColumns: &awstypes.TableWithColumnsResource{
							CatalogId:      aws.String(accountID),
							ColumnWildcard: &awstypes.ColumnWildcard{},
							DatabaseName:   aws.String(dbName),
							Name:           aws.String(tableName),
						},
					},
				},
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionSelect},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						TableWithColumns: &awstypes.TableWithColumnsResource{
							CatalogId:    aws.String(accountID),
							ColumnNames:  []string{"fred"},
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
			},
			ExpectedClean: []awstypes.PrincipalResourcePermissions{
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionAlter, awstypes.PermissionDelete},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						Table: &awstypes.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionSelect},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						TableWithColumns: &awstypes.TableWithColumnsResource{
							CatalogId:      aws.String(accountID),
							ColumnWildcard: &awstypes.ColumnWildcard{},
							DatabaseName:   aws.String(dbName),
							Name:           aws.String(tableName),
						},
					},
				},
			},
		},
		{
			Name: "twcWildcardExcluded",
			Input: &lakeformation.ListPermissionsInput{
				Principal: principal,
				Resource: &awstypes.Resource{
					Table: &awstypes.TableResource{
						CatalogId:    aws.String(accountID),
						DatabaseName: aws.String(dbName),
						Name:         aws.String(tableName),
					},
				},
			},
			TableType:           tflakeformation.TableTypeTableWithColumns,
			ColumnWildcard:      true,
			ExcludedColumnNames: []string{names.AttrValue},
			All: []awstypes.PrincipalResourcePermissions{
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionAlter, awstypes.PermissionDelete},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						Table: &awstypes.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionSelect},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						TableWithColumns: &awstypes.TableWithColumnsResource{
							CatalogId: aws.String(accountID),
							ColumnWildcard: &awstypes.ColumnWildcard{
								ExcludedColumnNames: []string{names.AttrValue},
							},
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionSelect},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						TableWithColumns: &awstypes.TableWithColumnsResource{
							CatalogId:    aws.String(accountID),
							ColumnNames:  []string{"fred"},
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
			},
			ExpectedClean: []awstypes.PrincipalResourcePermissions{
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionAlter, awstypes.PermissionDelete},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						Table: &awstypes.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionSelect},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						TableWithColumns: &awstypes.TableWithColumnsResource{
							CatalogId: aws.String(accountID),
							ColumnWildcard: &awstypes.ColumnWildcard{
								ExcludedColumnNames: []string{names.AttrValue},
							},
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			got := tflakeformation.FilterPermissions(testCase.Input, testCase.TableType, testCase.ColumnNames, testCase.ExcludedColumnNames, testCase.ColumnWildcard, testCase.All)

			if !reflect.DeepEqual(testCase.ExpectedClean, got) {
				t.Errorf("got %v, expected %v, input %v", got, testCase.ExpectedClean, testCase.Input)
			}
		})
	}
}
