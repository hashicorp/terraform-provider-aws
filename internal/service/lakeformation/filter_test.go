// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestFilterTablePermissions(t *testing.T) {
	t.Parallel()

	// primitives to make test cases easier
	accountID := acctest.Ct12Digit
	dbName := "Hiliji"
	altDBName := "Hiuhbum"
	tableName := "Ladocmoc"
	altTableName := "AltTable"

	//lintignore:AWSAT005
	principalIdentifier := fmt.Sprintf("arn:aws-us-gov:iam::%s:role/Zepotiz-Bulgaria", accountID)

	principal := &awstypes.DataLakePrincipal{
		DataLakePrincipalIdentifier: aws.String(principalIdentifier),
	}

	testCases := []struct {
		Name          string
		Input         *lakeformation.ListPermissionsInput
		All           []awstypes.PrincipalResourcePermissions
		ExpectedClean []awstypes.PrincipalResourcePermissions
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
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionSelect},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						Table: &awstypes.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(altDBName),
							Name:         aws.String(altTableName),
						},
					},
				},
			},
			ExpectedClean: []awstypes.PrincipalResourcePermissions{},
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
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			filter := tflakeformation.FilterTablePermissions(principalIdentifier, testCase.Input.Resource.Table)

			got := tfslices.Filter(testCase.All, filter)

			if diff := cmp.Diff(got, testCase.ExpectedClean,
				cmpopts.IgnoreUnexported(awstypes.PrincipalResourcePermissions{}),
				cmpopts.IgnoreUnexported(awstypes.DataLakePrincipal{}),
				cmpopts.IgnoreUnexported(awstypes.Resource{}),
				cmpopts.IgnoreUnexported(awstypes.TableResource{}),
				cmpopts.IgnoreUnexported(awstypes.TableWithColumnsResource{}),
				cmpopts.IgnoreUnexported(awstypes.ColumnWildcard{}),
			); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFilterTableWithColumnsPermissions(t *testing.T) {
	t.Parallel()

	// primitives to make test cases easier
	accountID := acctest.Ct12Digit
	dbName := "Hiliji"
	altDBName := "Hiuhbum"
	tableName := "Ladocmoc"
	altTableName := "AltTable"

	//lintignore:AWSAT005
	principalIdentifier := fmt.Sprintf("arn:aws-us-gov:iam::%s:role/Zepotiz-Bulgaria", accountID)

	principal := &awstypes.DataLakePrincipal{
		DataLakePrincipalIdentifier: aws.String(principalIdentifier),
	}

	testCases := []struct {
		Name                string
		Input               *lakeformation.ListPermissionsInput
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
			ColumnNames:   []string{names.AttrValue},
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
			ColumnNames: []string{names.AttrValue},
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
				{
					Permissions:                []awstypes.Permission{awstypes.PermissionSelect},
					PermissionsWithGrantOption: []awstypes.Permission{},
					Principal:                  principal,
					Resource: &awstypes.Resource{
						Table: &awstypes.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(altDBName),
							Name:         aws.String(altTableName),
						},
					},
				},
			},
			ExpectedClean: []awstypes.PrincipalResourcePermissions{},
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
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			filter := tflakeformation.FilterTableWithColumnsPermissions(principalIdentifier, testCase.Input.Resource.Table, testCase.ColumnNames, testCase.ExcludedColumnNames, testCase.ColumnWildcard)

			got := tfslices.Filter(testCase.All, filter)

			if diff := cmp.Diff(got, testCase.ExpectedClean,
				cmpopts.IgnoreUnexported(awstypes.PrincipalResourcePermissions{}),
				cmpopts.IgnoreUnexported(awstypes.DataLakePrincipal{}),
				cmpopts.IgnoreUnexported(awstypes.Resource{}),
				cmpopts.IgnoreUnexported(awstypes.TableResource{}),
				cmpopts.IgnoreUnexported(awstypes.TableWithColumnsResource{}),
				cmpopts.IgnoreUnexported(awstypes.ColumnWildcard{}),
			); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
