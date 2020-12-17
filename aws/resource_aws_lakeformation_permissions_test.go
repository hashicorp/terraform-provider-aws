package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSLakeFormationPermissions_basic(t *testing.T) {
	//rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceAddr := "aws_lakeformation_resource.test"
	bucketAddr := "aws_s3_bucket.test"
	roleAddr := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_catalog(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceAddr),
					resource.TestCheckResourceAttrPair(resourceAddr, "role_arn", roleAddr, "arn"),
					resource.TestCheckResourceAttrPair(resourceAddr, "resource_arn", bucketAddr, "arn"),
				),
			},
		},
	})
}

func TestAccAWSLakeFormationPermissions_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationPermissionsExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsLakeFormationPermissions(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLakeFormationPermissions_full(t *testing.T) {
	rName := acctest.RandomWithPrefix("lakeformation-test-bucket")
	dName := acctest.RandomWithPrefix("lakeformation-test-db")
	tName := acctest.RandomWithPrefix("lakeformation-test-table")

	roleName := "data.aws_iam_role.test"
	resourceName := "aws_lakeformation_permissions.test"
	bucketName := "aws_s3_bucket.test"
	dbName := "aws_glue_catalog_database.test"
	tableName := "aws_glue_catalog_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationPermissionsConfig_catalog(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrAccountID(resourceName, "catalog_id"),
					resource.TestCheckResourceAttrPair(roleName, "arn", resourceName, "principal"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", "CREATE_DATABASE"),
				),
			},
			{
				Config: testAccAWSLakeFormationPermissionsConfig_location(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrAccountID(resourceName, "catalog_id"),
					resource.TestCheckResourceAttrPair(roleName, "arn", resourceName, "principal"),
					resource.TestCheckResourceAttrPair(bucketName, "arn", resourceName, "location"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", "DATA_LOCATION_ACCESS"),
				),
			},
			{
				Config: testAccAWSLakeFormationPermissionsConfig_database(rName, dName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrAccountID(resourceName, "catalog_id"),
					resource.TestCheckResourceAttrPair(roleName, "arn", resourceName, "principal"),
					resource.TestCheckResourceAttrPair(dbName, "name", resourceName, "database"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", "ALTER"),
					resource.TestCheckResourceAttr(resourceName, "permissions.1", "CREATE_TABLE"),
					resource.TestCheckResourceAttr(resourceName, "permissions.2", "DROP"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions_with_grant_option.0", "CREATE_TABLE"),
				),
			},
			{
				Config: testAccAWSLakeFormationPermissionsConfig_table(rName, dName, tName, "\"ALL\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrAccountID(resourceName, "catalog_id"),
					resource.TestCheckResourceAttrPair(roleName, "arn", resourceName, "principal"),
					resource.TestCheckResourceAttr(resourceName, "table.#", "1"),
					resource.TestCheckResourceAttrPair(tableName, "database_name", resourceName, "table.0.database"),
					resource.TestCheckResourceAttrPair(tableName, "name", resourceName, "table.0.name"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", "ALL"),
				),
			},
			{
				Config: testAccAWSLakeFormationPermissionsConfig_table(rName, dName, tName, "\"ALL\", \"SELECT\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrAccountID(resourceName, "catalog_id"),
					resource.TestCheckResourceAttrPair(roleName, "arn", resourceName, "principal"),
					resource.TestCheckResourceAttr(resourceName, "table.#", "1"),
					resource.TestCheckResourceAttrPair(tableName, "database_name", resourceName, "table.0.database"),
					resource.TestCheckResourceAttrPair(tableName, "name", resourceName, "table.0.name"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "permissions.1", "SELECT"),
				),
			},
			{
				Config: testAccAWSLakeFormationPermissionsConfig_tableWithColumns(rName, dName, tName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrAccountID(resourceName, "catalog_id"),
					resource.TestCheckResourceAttrPair(roleName, "arn", resourceName, "principal"),
					resource.TestCheckResourceAttr(resourceName, "table.#", "1"),
					resource.TestCheckResourceAttrPair(tableName, "database_name", resourceName, "table.0.database"),
					resource.TestCheckResourceAttrPair(tableName, "name", resourceName, "table.0.name"),
					resource.TestCheckResourceAttr(resourceName, "table.0.column_names.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "table.0.column_names.0", "event"),
					resource.TestCheckResourceAttr(resourceName, "table.0.column_names.1", "timestamp"),
					resource.TestCheckResourceAttr(resourceName, "permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions.0", "SELECT"),
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

		principal := rs.Primary.Attributes["principal"]
		catalogId := rs.Primary.Attributes["catalog_id"]

		input := &lakeformation.ListPermissionsInput{
			CatalogId: aws.String(catalogId),
			Principal: &lakeformation.DataLakePrincipal{
				DataLakePrincipalIdentifier: aws.String(principal),
			},
		}

		out, err := conn.ListPermissions(input)
		if err == nil {
			fmt.Print(out)
			return fmt.Errorf("Resource still registered: %s %s", catalogId, principal)
		}
	}

	return nil
}

func testAccCheckAWSLakeFormationPermissionsExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).lakeformationconn

		input := &lakeformation.ListPermissionsInput{
			MaxResults: aws.Int64(1),
			Principal: &lakeformation.DataLakePrincipal{
				DataLakePrincipalIdentifier: aws.String(rs.Primary.Attributes["principal"]),
			},
		}

		if rs.Primary.Attributes["catalog_resource"] == "true" {
			input.ResourceType = aws.String(lakeformation.DataLakeResourceTypeCatalog)
		}

		if rs.Primary.Attributes["data_location.#"] != "0" {
			input.ResourceType = aws.String(lakeformation.DataLakeResourceTypeDataLocation)
		}

		if rs.Primary.Attributes["database.#"] != "0" {
			input.ResourceType = aws.String(lakeformation.DataLakeResourceTypeDatabase)
		}

		if rs.Primary.Attributes["table.#"] != "0" {
			input.ResourceType = aws.String(lakeformation.DataLakeResourceTypeTable)
		}

		if rs.Primary.Attributes["table_with_columns.#"] != "0" {
			input.ResourceType = aws.String(lakeformation.DataLakeResourceTypeTable)
		}

		_, err := conn.ListPermissions(input)

		if err != nil {
			return fmt.Errorf("error getting Lake Formation resource (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccAWSLakeFormationPermissionsConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "workflow_role" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "managed" {
  role       = aws_iam_role.workflow_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSGlueServiceRole"
}

resource "aws_iam_role_policy" "inline" {
  name = %[1]q
  role = aws_iam_role.workflow_role.name

  policy = <<-EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "lakeformation:GetDataAccess",
        "lakeformation:GrantPermissions"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": ["iam:PassRole"],
      "Resource": [
        "${aws_iam_role.workflow_role.arn}"
      ]
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true

  tags = {
    Name = %[1]q
  }
}

// register the s3 bucket as the data lake
resource "aws_lakeformation_resource" "test" {
  resource_arn = aws_s3_bucket.test.arn
}

data "aws_caller_identity" "current" {}

resource "aws_lakeformation_data_lake_settings" "test" {
  data_lake_admins = [data.aws_caller_identity.current.arn]
}

// grants permissions to workflow role in catalog on bucket
resource "aws_lakeformation_permissions" "grants" {
  principal_arn = data.aws_caller_identity.current.arn
  permissions   = ["CREATE_DATABASE"]
  
  //data_location {
  //  resource_arn = aws_s3_bucket.test.arn
  //}
  catalog_resource = true

  depends_on = ["aws_lakeformation_resource.test", "aws_iam_role.workflow_role", "aws_iam_role_policy_attachment.managed"]
}
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_catalog() string {
	return `
data "aws_caller_identity" "current" {}

data "aws_iam_role" "test" {
  name = "AWSServiceRoleForLakeFormationDataAccess"
}

resource "aws_lakeformation_data_lake_settings" "test" {
  data_lake_admins = [
    data.aws_caller_identity.current.arn
  ]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["CREATE_DATABASE"]
  principal_arn   = data.aws_iam_role.test.arn
  catalog_resource = true

  depends_on = ["aws_lakeformation_data_lake_settings.test"]
}
`
}

func testAccAWSLakeFormationPermissionsConfig_location(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_role" "test" {
  name = "AWSServiceRoleForLakeFormationDataAccess"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [
    data.aws_caller_identity.current.arn
  ]
}

resource "aws_lakeformation_resource" "test" {
  resource_arn            = aws_s3_bucket.test.arn
  use_service_linked_role = true

  depends_on = ["aws_lakeformation_data_lake_settings.test"]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["DATA_LOCATION_ACCESS"]
  principal   = data.aws_iam_role.test.arn

  location = aws_lakeformation_resource.test.resource_arn

  depends_on = ["aws_lakeformation_data_lake_settings.test"]
}
`, rName)
}

func testAccAWSLakeFormationPermissionsConfig_database(rName, dName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_role" "test" {
  name = "AWSServiceRoleForLakeFormationDataAccess"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_glue_catalog_database" "test" {
  name = %[2]q
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [
    data.aws_caller_identity.current.arn
  ]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALTER", "CREATE_TABLE", "DROP"]
  permissions_with_grant_option = ["CREATE_TABLE"]
  principal   = data.aws_iam_role.test.arn

  database = aws_glue_catalog_database.test.name

  depends_on = ["aws_lakeformation_data_lake_settings.test"]
}
`, rName, dName)
}

func testAccAWSLakeFormationPermissionsConfig_table(rName, dName, tName, permissions string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_role" "test" {
  name = "AWSServiceRoleForLakeFormationDataAccess"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_glue_catalog_database" "test" {
  name = %[2]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[3]q
  database_name = aws_glue_catalog_database.test.name
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [
    data.aws_caller_identity.current.arn
  ]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = [%s]
  principal   = data.aws_iam_role.test.arn

  table {
  	database = aws_glue_catalog_table.test.database_name
  	name = aws_glue_catalog_table.test.name
  }

  depends_on = ["aws_lakeformation_data_lake_settings.test"]
}
`, rName, dName, tName, permissions)
}

func testAccAWSLakeFormationPermissionsConfig_tableWithColumns(rName, dName, tName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_role" "test" {
  name = "AWSServiceRoleForLakeFormationDataAccess"
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_glue_catalog_database" "test" {
  name = %[2]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[3]q
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

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [
    data.aws_caller_identity.current.arn
  ]
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["SELECT"]
  principal   = data.aws_iam_role.test.arn

  table {
  	database = aws_glue_catalog_table.test.database_name
  	name = aws_glue_catalog_table.test.name
  	column_names = ["event", "timestamp"]
  }

  depends_on = ["aws_lakeformation_data_lake_settings.test"]
}
`, rName, dName, tName)
}
