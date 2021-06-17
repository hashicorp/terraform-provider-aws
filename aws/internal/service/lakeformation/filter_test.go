package lakeformation

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
)

func TestFilterPermissions(t *testing.T) {
	// primitives to make test cases easier
	accountID := "481516234248"
	dbName := "Hiliji"
	altDBName := "Hiuhbum"
	tableName := "Ladocmoc"

	principal := &lakeformation.DataLakePrincipal{
		DataLakePrincipalIdentifier: aws.String(fmt.Sprintf("arn:aws-us-gov:iam::%s:role/Zepotiz-Bulgaria", accountID)),
	}

	testCases := []struct {
		Name                string
		Input               *lakeformation.ListPermissionsInput
		TableType           string
		ColumnNames         []*string
		ExcludedColumnNames []*string
		ColumnWildcard      bool
		All                 []*lakeformation.PrincipalResourcePermissions
		ExpectedClean       []*lakeformation.PrincipalResourcePermissions
	}{
		{
			Name: "empty",
			Input: &lakeformation.ListPermissionsInput{
				Principal: principal,
				Resource:  &lakeformation.Resource{},
			},
			All:           nil,
			ExpectedClean: nil,
		},
		{
			Name: "emptyWithInput",
			Input: &lakeformation.ListPermissionsInput{
				Principal: principal,
				Resource: &lakeformation.Resource{
					Table: &lakeformation.TableResource{
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
			Name: "wrongTableResource",
			Input: &lakeformation.ListPermissionsInput{
				Principal: principal,
				Resource: &lakeformation.Resource{
					Table: &lakeformation.TableResource{
						CatalogId:    aws.String(accountID),
						DatabaseName: aws.String(dbName),
						Name:         aws.String(tableName),
					},
				},
			},
			All: []*lakeformation.PrincipalResourcePermissions{
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionSelect}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						Table: &lakeformation.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(altDBName),
							Name:         aws.String(tableName),
						},
					},
				},
			},
			ExpectedClean: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			got := FilterPermissions(testCase.Input, testCase.TableType, testCase.ColumnNames, testCase.ExcludedColumnNames, testCase.ColumnWildcard, testCase.All)

			if !reflect.DeepEqual(testCase.ExpectedClean, got) {
				t.Errorf("got %v, expected %v", got, testCase.ExpectedClean)
			}
		})
	}
}
