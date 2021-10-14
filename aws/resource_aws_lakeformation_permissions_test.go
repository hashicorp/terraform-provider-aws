package aws

import (
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	iamwaiter "github.com/hashicorp/terraform-provider-aws/aws/internal/service/iam/waiter"
	tflakeformation "github.com/hashicorp/terraform-provider-aws/aws/internal/service/lakeformation"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccAWSLakeFormationPermissions_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", lakeformation.PermissionCreateDatabase),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", "true"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_twcBasic(rName, "\"event\", \"timestamp\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					acctest.CheckResourceDisappears(testAccProvider, resourceAwsLakeFormationPermissions(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_database(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	dbName := "aws_glue_catalog_database.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_database(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "database.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "database.0.name", dbName, "name"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", lakeformation.PermissionAlter),
					resource.TestCheckResourceAttr(resourceName, "permissions.1", lakeformation.PermissionCreateTable),
					resource.TestCheckResourceAttr(resourceName, "permissions.2", lakeformation.PermissionDrop),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.0", lakeformation.PermissionCreateTable),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_databaseIAMAllowed(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	dbName := "aws_glue_catalog_database.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_databaseIAMAllowed(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "principal", tflakeformation.IAMAllowedPrincipals),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", "false"),
					resource.TestCheckResourceAttr(resourceName, "database.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "database.0.name", dbName, "name"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", lakeformation.PermissionAll),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "0"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_databaseMultiple(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	resourceName2 := "aws_lakeformation_permissions.test2"
	roleName := "aws_iam_role.test"
	roleName2 := "aws_iam_role.test2"
	dbName := "aws_glue_catalog_database.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_databaseMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "database.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "database.0.name", dbName, "name"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", lakeformation.PermissionAlter),
					resource.TestCheckResourceAttr(resourceName, "permissions.1", lakeformation.PermissionCreateTable),
					resource.TestCheckResourceAttr(resourceName, "permissions.2", lakeformation.PermissionDrop),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.0", lakeformation.PermissionCreateTable),
					testAccCheckAWSLakeFormationPermissionsExists(resourceName2),
					resource.TestCheckResourceAttrPair(resourceName2, "principal", roleName2, "arn"),
					resource.TestCheckResourceAttr(resourceName2, "catalog_resource", "false"),
					resource.TestCheckResourceAttrPair(resourceName2, "principal", roleName2, "arn"),
					resource.TestCheckResourceAttr(resourceName2, "database.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName2, "database.0.name", dbName, "name"),
					resource.TestCheckResourceAttr(resourceName2, "permissions.#", "2"),
					resource.TestCheckResourceAttr(resourceName2, "permissions.0", lakeformation.PermissionAlter),
					resource.TestCheckResourceAttr(resourceName2, "permissions.1", lakeformation.PermissionDrop),
					resource.TestCheckResourceAttr(resourceName2, "permissions_with_grant_option.#", "0"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_dataLocation(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	bucketName := "aws_s3_bucket.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_dataLocation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", lakeformation.PermissionDataLocationAccess),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", "false"),
					resource.TestCheckResourceAttr(resourceName, "data_location.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "data_location.0.arn", bucketName, "arn"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_tableBasic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tableName := "aws_glue_catalog_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_tableBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(roleName, "arn", resourceName, "principal"),
					resource.TestCheckResourceAttr(resourceName, "table.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", tableName, "database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.name", tableName, "name"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", lakeformation.PermissionAlter),
					resource.TestCheckResourceAttr(resourceName, "permissions.1", lakeformation.PermissionDelete),
					resource.TestCheckResourceAttr(resourceName, "permissions.2", lakeformation.PermissionDescribe),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_tableIAMAllowed(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	dbName := "aws_glue_catalog_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_tableIAMAllowed(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "principal", tflakeformation.IAMAllowedPrincipals),
					resource.TestCheckResourceAttr(resourceName, "catalog_resource", "false"),
					resource.TestCheckResourceAttr(resourceName, "table.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", dbName, "database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.name", dbName, "name"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", lakeformation.PermissionAll),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "0"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_tableImplicit(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tableName := "aws_glue_catalog_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_tableImplicit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "table.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", tableName, "database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.name", tableName, "name"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "7"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "7"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_tableMultipleRoles(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	resourceName2 := "aws_lakeformation_permissions.test2"
	roleName := "aws_iam_role.test"
	roleName2 := "aws_iam_role.test2"
	tableName := "aws_glue_catalog_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_tableMultipleRoles(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(roleName, "arn", resourceName, "principal"),
					resource.TestCheckResourceAttr(resourceName, "table.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", tableName, "database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.name", tableName, "name"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", lakeformation.PermissionAlter),
					resource.TestCheckResourceAttr(resourceName, "permissions.1", lakeformation.PermissionDelete),
					resource.TestCheckResourceAttr(resourceName, "permissions.2", lakeformation.PermissionDescribe),
					testAccCheckAWSLakeFormationPermissionsExists(resourceName2),
					resource.TestCheckResourceAttrPair(roleName2, "arn", resourceName2, "principal"),
					resource.TestCheckResourceAttr(resourceName2, "table.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName2, "table.0.database_name", tableName, "database_name"),
					resource.TestCheckResourceAttrPair(resourceName2, "table.0.name", tableName, "name"),
					resource.TestCheckResourceAttr(resourceName2, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName2, "permissions.0", lakeformation.PermissionSelect),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_tableSelectOnly(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tableName := "aws_glue_catalog_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_tableSelectOnly(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(roleName, "arn", resourceName, "principal"),
					resource.TestCheckResourceAttr(resourceName, "table.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", tableName, "database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.name", tableName, "name"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", lakeformation.PermissionSelect),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_tableSelectPlus(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_tableSelectPlus(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "7"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "7"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_tableWildcardNoSelect(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	databaseResourceName := "aws_glue_catalog_database.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_tableWildcardNoSelect(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "table.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table.0.database_name", databaseResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "table.0.wildcard", "true"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_tableWildcardSelectOnly(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_tableWildcardSelectOnly(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", lakeformation.PermissionSelect),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "0"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_tableWildcardSelectPlus(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_tableWildcardSelectPlus(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "7"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "7"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_twcBasic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tableName := "aws_glue_catalog_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_twcBasic(rName, "\"event\", \"timestamp\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.database_name", tableName, "database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.name", tableName, "name"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.0", "event"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.1", "timestamp"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", lakeformation.PermissionSelect),
				),
			},
			{
				Config: testAccAWSLakeFormationPermissionsConfig_twcBasic(rName, "\"timestamp\", \"event\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.database_name", tableName, "database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.name", tableName, "name"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.0", "event"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.1", "timestamp"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", lakeformation.PermissionSelect),
				),
			},
			{
				Config: testAccAWSLakeFormationPermissionsConfig_twcBasic(rName, "\"timestamp\", \"event\", \"transactionamount\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.database_name", tableName, "database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.name", tableName, "name"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.0", "event"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.1", "timestamp"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.2", "transactionamount"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", lakeformation.PermissionSelect),
				),
			},
			{
				Config: testAccAWSLakeFormationPermissionsConfig_twcBasic(rName, "\"event\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.database_name", tableName, "database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.name", tableName, "name"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.0", "event"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", lakeformation.PermissionSelect),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_twcImplicit(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tableName := "aws_glue_catalog_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_twcImplicit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.database_name", tableName, "database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.name", tableName, "name"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.wildcard", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "7"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "7"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_twcWildcardExcludedColumns(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_twcWildcardExcludedColumns(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "0"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_twcWildcardSelectOnly(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"
	tableName := "aws_glue_catalog_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_twcWildcardSelectOnly(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.database_name", tableName, "database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "table_with_columns.0.name", tableName, "name"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.column_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "table_with_columns.0.wildcard", "true"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", lakeformation.PermissionSelect),
				),
			},
		},
	})
}

func testAccAWSLakeFormationPermissions_twcWildcardSelectPlus(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_permissions.test"
	roleName := "aws_iam_role.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, lakeformation.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_twcWildcardSelectPlus(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "principal", roleName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "7"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSLakeFormationPermissionsDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lakeformationconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lakeformation_permissions" {
			continue
		}

		permCount, err := permissionCountForLakeFormationResource(conn, rs)

		if err != nil {
			return fmt.Errorf("acceptance test: error listing Lake Formation permissions (%s): %w", rs.Primary.ID, err)
		}

		if permCount != 0 {
			return fmt.Errorf("acceptance test: Lake Formation permissions (%s) still exist: %d", rs.Primary.ID, permCount)
		}

		return nil
	}

	return nil
}

func testAccCheckAWSLakeFormationPermissionsExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("acceptance test: resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).lakeformationconn

		permCount, err := permissionCountForLakeFormationResource(conn, rs)

		if err != nil {
			return fmt.Errorf("acceptance test: error listing Lake Formation permissions (%s): %w", rs.Primary.ID, err)
		}

		if permCount == 0 {
			return fmt.Errorf("acceptance test: Lake Formation permissions (%s) do not exist or could not be found", rs.Primary.ID)
		}

		return nil
	}
}

func permissionCountForLakeFormationResource(conn *lakeformation.LakeFormation, rs *terraform.ResourceState) (int, error) {
	input := &lakeformation.ListPermissionsInput{
		Principal: &lakeformation.DataLakePrincipal{
			DataLakePrincipalIdentifier: aws.String(rs.Primary.Attributes["principal"]),
		},
		Resource: &lakeformation.Resource{},
	}

	noResource := true

	if v, ok := rs.Primary.Attributes["catalog_id"]; ok && v != "" {
		input.CatalogId = aws.String(v)

		noResource = false
	}

	if v, ok := rs.Primary.Attributes["catalog_resource"]; ok && v != "" && v == "true" {
		input.Resource.Catalog = expandLakeFormationCatalogResource()

		noResource = false
	}

	if v, ok := rs.Primary.Attributes["data_location.#"]; ok && v != "" && v != "0" {
		tfMap := map[string]interface{}{}

		if v := rs.Primary.Attributes["data_location.0.catalog_id"]; v != "" {
			tfMap["catalog_id"] = v
		}

		if v := rs.Primary.Attributes["data_location.0.arn"]; v != "" {
			tfMap["arn"] = v
		}

		input.Resource.DataLocation = expandLakeFormationDataLocationResource(tfMap)

		noResource = false
	}

	if v, ok := rs.Primary.Attributes["database.#"]; ok && v != "" && v != "0" {
		tfMap := map[string]interface{}{}

		if v := rs.Primary.Attributes["database.0.catalog_id"]; v != "" {
			tfMap["catalog_id"] = v
		}

		if v := rs.Primary.Attributes["database.0.name"]; v != "" {
			tfMap["name"] = v
		}

		input.Resource.Database = expandLakeFormationDatabaseResource(tfMap)

		noResource = false
	}

	tableType := ""

	if v, ok := rs.Primary.Attributes["table.#"]; ok && v != "" && v != "0" {
		tableType = tflakeformation.TableTypeTable

		tfMap := map[string]interface{}{}

		if v := rs.Primary.Attributes["table.0.catalog_id"]; v != "" {
			tfMap["catalog_id"] = v
		}

		if v := rs.Primary.Attributes["table.0.database_name"]; v != "" {
			tfMap["database_name"] = v
		}

		if v := rs.Primary.Attributes["table.0.name"]; v != "" && v != tflakeformation.TableNameAllTables {
			tfMap["name"] = v
		}

		if v := rs.Primary.Attributes["table.0.wildcard"]; v != "" && v == "true" {
			tfMap["wildcard"] = true
		}

		input.Resource.Table = expandLakeFormationTableResource(tfMap)

		noResource = false
	}

	if v, ok := rs.Primary.Attributes["table_with_columns.#"]; ok && v != "" && v != "0" {
		tableType = tflakeformation.TableTypeTableWithColumns

		tfMap := map[string]interface{}{}

		if v := rs.Primary.Attributes["table_with_columns.0.catalog_id"]; v != "" {
			tfMap["catalog_id"] = v
		}

		if v := rs.Primary.Attributes["table_with_columns.0.database_name"]; v != "" {
			tfMap["database_name"] = v
		}

		if v := rs.Primary.Attributes["table_with_columns.0.name"]; v != "" {
			tfMap["name"] = v
		}

		input.Resource.Table = expandLakeFormationTableWithColumnsResourceAsTable(tfMap)

		noResource = false
	}

	if noResource {
		// if after read, there is no resource, it has been deleted
		return 0, nil
	}

	log.Printf("[DEBUG] Reading Lake Formation permissions: %v", input)
	var allPermissions []*lakeformation.PrincipalResourcePermissions

	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		err := conn.ListPermissionsPages(input, func(resp *lakeformation.ListPermissionsOutput, lastPage bool) bool {
			for _, permission := range resp.PrincipalResourcePermissions {
				if permission == nil {
					continue
				}

				allPermissions = append(allPermissions, permission)
			}
			return !lastPage
		})

		if err != nil {
			if tfawserr.ErrMessageContains(err, lakeformation.ErrCodeInvalidInputException, "Invalid principal") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(fmt.Errorf("acceptance test: error listing Lake Formation Permissions getting permission count: %w", err))
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		err = conn.ListPermissionsPages(input, func(resp *lakeformation.ListPermissionsOutput, lastPage bool) bool {
			for _, permission := range resp.PrincipalResourcePermissions {
				if permission == nil {
					continue
				}

				allPermissions = append(allPermissions, permission)
			}
			return !lastPage
		})
	}

	if tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeEntityNotFoundException) {
		return 0, nil
	}

	if tfawserr.ErrMessageContains(err, "AccessDeniedException", "Resource does not exist") {
		return 0, nil
	}

	if err != nil {
		return 0, fmt.Errorf("acceptance test: error listing Lake Formation permissions after retry %v: %w", input, err)
	}

	columnNames := make([]*string, 0)
	excludedColumnNames := make([]*string, 0)
	columnWildcard := false

	if tableType == tflakeformation.TableTypeTableWithColumns {
		if v := rs.Primary.Attributes["table_with_columns.0.wildcard"]; v != "" && v == "true" {
			columnWildcard = true
		}

		colCount := 0

		if v := rs.Primary.Attributes["table_with_columns.0.column_names.#"]; v != "" {
			colCount, err = strconv.Atoi(rs.Primary.Attributes["table_with_columns.0.column_names.#"])

			if err != nil {
				return 0, fmt.Errorf("acceptance test: could not convert string (%s) Atoi for column_names: %w", rs.Primary.Attributes["table_with_columns.0.column_names.#"], err)
			}
		}

		for i := 0; i < colCount; i++ {
			columnNames = append(columnNames, aws.String(rs.Primary.Attributes[fmt.Sprintf("table_with_columns.0.column_names.%d", i)]))
		}

		colCount = 0

		if v := rs.Primary.Attributes["table_with_columns.0.excluded_column_names.#"]; v != "" {
			colCount, err = strconv.Atoi(rs.Primary.Attributes["table_with_columns.0.excluded_column_names.#"])

			if err != nil {
				return 0, fmt.Errorf("acceptance test: could not convert string (%s) Atoi for excluded_column_names: %w", rs.Primary.Attributes["table_with_columns.0.excluded_column_names.#"], err)
			}
		}

		for i := 0; i < colCount; i++ {
			excludedColumnNames = append(excludedColumnNames, aws.String(rs.Primary.Attributes[fmt.Sprintf("table_with_columns.0.excluded_column_names.%d", i)]))
		}
	}

	// clean permissions = filter out permissions that do not pertain to this specific resource
	cleanPermissions := tflakeformation.FilterPermissions(input, tableType, columnNames, excludedColumnNames, columnWildcard, allPermissions)

	return len(cleanPermissions), nil
}

func testAccAWSLakeFormationPermissionsConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  principal        = aws_iam_role.test.arn
  permissions      = ["CREATE_DATABASE"]
  catalog_resource = true

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_database(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions                   = ["ALTER", "CREATE_TABLE", "DROP"]
  permissions_with_grant_option = ["CREATE_TABLE"]
  principal                     = aws_iam_role.test.arn

  database {
    name = aws_glue_catalog_database.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_databaseIAMAllowed(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "transactionamount"
      type = "double"
    }
  }
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALL"]
  principal   = "IAM_ALLOWED_PRINCIPALS"

  database {
    name = aws_glue_catalog_database.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_databaseMultiple(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role" "test2" {
  name = "%[1]s-2"
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions                   = ["ALTER", "CREATE_TABLE", "DROP"]
  permissions_with_grant_option = ["CREATE_TABLE"]
  principal                     = aws_iam_role.test.arn

  database {
    name = aws_glue_catalog_database.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_permissions" "test2" {
  permissions = ["ALTER", "DROP"]
  principal   = aws_iam_role.test2.arn

  database {
    name = aws_glue_catalog_database.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_dataLocation(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
      }, {
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "s3.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

resource "aws_lakeformation_resource" "test" {
  arn      = aws_s3_bucket.test.arn
  role_arn = aws_iam_role.test.arn
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  principal   = aws_iam_role.test.arn
  permissions = ["DATA_LOCATION_ACCESS"]

  data_location {
    arn = aws_s3_bucket.test.arn
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_tableBasic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALTER", "DELETE", "DESCRIBE"]
  principal   = aws_iam_role.test.arn

  table {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_tableIAMAllowed(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "transactionamount"
      type = "double"
    }
  }
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALL"]
  principal   = "IAM_ALLOWED_PRINCIPALS"

  table {
    database_name = aws_glue_catalog_database.test.name
    name          = aws_glue_catalog_table.test.name
  }
}
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_tableImplicit(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "value"
      type = "double"
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  # this will give the principal implicit permissions
  admins = [aws_iam_role.test.arn, data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  principal                     = aws_iam_role.test.arn
  permissions                   = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT", "SELECT"]
  permissions_with_grant_option = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT", "SELECT"]

  table {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_tableMultipleRoles(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role" "test2" {
  name = "%[1]s-2"
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALTER", "DELETE", "DESCRIBE"]
  principal   = aws_iam_role.test.arn

  table {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}

resource "aws_lakeformation_permissions" "test2" {
  permissions = ["SELECT"]
  principal   = aws_iam_role.test2.arn

  table {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_tableSelectOnly(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["SELECT"]
  principal   = aws_iam_role.test.arn

  table {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_tableSelectPlus(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "value"
      type = "double"
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  principal = aws_iam_role.test.arn

  permissions                   = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT", "SELECT"]
  permissions_with_grant_option = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT", "SELECT"]

  table {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_tableWildcardNoSelect(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions                   = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT"]
  permissions_with_grant_option = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT"]
  principal                     = aws_iam_role.test.arn

  table {
    database_name = aws_glue_catalog_database.test.name
    wildcard      = true
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_tableWildcardSelectOnly(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "value"
      type = "double"
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  principal = aws_iam_role.test.arn

  permissions = ["SELECT"]

  table {
    database_name = aws_glue_catalog_table.test.database_name
    wildcard      = true
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_tableWildcardSelectPlus(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "value"
      type = "double"
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  principal = aws_iam_role.test.arn

  permissions                   = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT", "SELECT"]
  permissions_with_grant_option = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT", "SELECT"]

  table {
    database_name = aws_glue_catalog_table.test.database_name
    wildcard      = true
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_twcBasic(rName string, columns string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "transactionamount"
      type = "double"
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["SELECT"]
  principal   = aws_iam_role.test.arn

  table_with_columns {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
    column_names  = [%[2]s]
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName, columns)
}

func testAccAWSLakeFormationPermissionsConfig_twcImplicit(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "value"
      type = "double"
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  # this will give the principal implicit permissions
  admins = [aws_iam_role.test.arn, data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  principal                     = aws_iam_role.test.arn
  permissions                   = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT", "SELECT"]
  permissions_with_grant_option = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT", "SELECT"]

  table_with_columns {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
    wildcard      = true
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_twcWildcardExcludedColumns(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "value"
      type = "double"
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["SELECT"]
  principal   = aws_iam_role.test.arn

  table_with_columns {
    database_name         = aws_glue_catalog_table.test.database_name
    name                  = aws_glue_catalog_table.test.name
    wildcard              = true
    excluded_column_names = ["value"]
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_twcWildcardSelectOnly(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["SELECT"]
  principal   = aws_iam_role.test.arn

  table_with_columns {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
    wildcard      = true
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_twcWildcardSelectPlus(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "glue.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name

  storage_descriptor {
    columns {
      name = "event"
      type = "string"
    }

    columns {
      name = "timestamp"
      type = "date"
    }

    columns {
      name = "value"
      type = "double"
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALL", "ALTER", "DELETE", "DESCRIBE", "DROP", "INSERT", "SELECT"]
  principal   = aws_iam_role.test.arn

  table_with_columns {
    database_name = aws_glue_catalog_table.test.database_name
    name          = aws_glue_catalog_table.test.name
    wildcard      = true
  }

  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}
