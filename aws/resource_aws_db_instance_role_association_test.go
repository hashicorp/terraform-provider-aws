package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAWSDbInstanceRoleAssociation_basic(t *testing.T) {
	var dbInstanceRole1 rds.DBInstanceRole
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	dbInstanceResourceName := "aws_db_instance.test"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_db_instance_role_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSDbInstanceRoleAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDbInstanceRoleAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDbInstanceRoleAssociationExists(resourceName, &dbInstanceRole1),
					resource.TestCheckResourceAttrPair(resourceName, "db_instance_identifier", dbInstanceResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "feature_name", "S3_INTEGRATION"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", iamRoleResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDbInstanceRoleAssociation_disappears(t *testing.T) {
	var dbInstance1 rds.DBInstance
	var dbInstanceRole1 rds.DBInstanceRole
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	dbInstanceResourceName := "aws_db_instance.test"
	resourceName := "aws_db_instance_role_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, rds.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSDbInstanceRoleAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDbInstanceRoleAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBInstanceExists(dbInstanceResourceName, &dbInstance1),
					testAccCheckAWSDbInstanceRoleAssociationExists(resourceName, &dbInstanceRole1),
					testAccCheckAWSDbInstanceRoleAssociationDisappears(&dbInstance1, &dbInstanceRole1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSDbInstanceRoleAssociationExists(resourceName string, dbInstanceRole *rds.DBInstanceRole) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		dbInstanceIdentifier, roleArn, err := resourceAwsDbInstanceRoleAssociationDecodeId(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error reading resource ID: %s", err)
		}

		conn := acctest.Provider.Meta().(*AWSClient).rdsconn

		role, err := rdsDescribeDbInstanceRole(conn, dbInstanceIdentifier, roleArn)

		if err != nil {
			return err
		}

		if role == nil {
			return fmt.Errorf("RDS DB Instance IAM Role Association not found")
		}

		if aws.StringValue(role.Status) != "ACTIVE" {
			return fmt.Errorf("RDS DB Instance (%s) IAM Role (%s) association exists in non-ACTIVE (%s) state", dbInstanceIdentifier, roleArn, aws.StringValue(role.Status))
		}

		*dbInstanceRole = *role

		return nil
	}
}

func testAccCheckAWSDbInstanceRoleAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_instance_role_association" {
			continue
		}

		dbInstanceIdentifier, roleArn, err := resourceAwsDbInstanceRoleAssociationDecodeId(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error reading resource ID: %s", err)
		}

		dbInstanceRole, err := rdsDescribeDbInstanceRole(conn, dbInstanceIdentifier, roleArn)

		if tfawserr.ErrMessageContains(err, rds.ErrCodeDBInstanceNotFoundFault, "") {
			continue
		}

		if err != nil {
			return err
		}

		if dbInstanceRole == nil {
			continue
		}

		return fmt.Errorf("RDS DB Instance (%s) IAM Role (%s) association still exists in non-deleted (%s) state", dbInstanceIdentifier, roleArn, aws.StringValue(dbInstanceRole.Status))
	}

	return nil
}

func testAccCheckAWSDbInstanceRoleAssociationDisappears(dbInstance *rds.DBInstance, dbInstanceRole *rds.DBInstanceRole) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*AWSClient).rdsconn

		input := &rds.RemoveRoleFromDBInstanceInput{
			DBInstanceIdentifier: dbInstance.DBInstanceIdentifier,
			FeatureName:          dbInstanceRole.FeatureName,
			RoleArn:              dbInstanceRole.RoleArn,
		}

		_, err := conn.RemoveRoleFromDBInstance(input)

		if err != nil {
			return err
		}

		return waitForRdsDbInstanceRoleDisassociation(conn, aws.StringValue(dbInstance.DBInstanceIdentifier), aws.StringValue(dbInstanceRole.RoleArn))
	}
}

func testAccAWSDbInstanceRoleAssociationConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_rds_orderable_db_instance" "test" {
  engine        = "oracle-se2"
  license_model = "bring-your-own-license"
  storage_type  = "standard"

  preferred_instance_classes = ["db.m5.large", "db.m4.large", "db.r4.large"]
}

data "aws_iam_policy_document" "rds_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      identifiers = ["rds.${data.aws_partition.current.dns_suffix}"]
      type        = "Service"
    }
  }
}

resource "aws_iam_role" "test" {
  assume_role_policy = data.aws_iam_policy_document.rds_assume_role_policy.json
  name               = %[1]q
}

resource "aws_db_instance" "test" {
  allocated_storage   = 10
  engine              = data.aws_rds_orderable_db_instance.test.engine
  identifier          = %[1]q
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  license_model       = data.aws_rds_orderable_db_instance.test.license_model
  password            = "avoid-plaintext-passwords"
  username            = "tfacctest"
  skip_final_snapshot = true
}

resource "aws_db_instance_role_association" "test" {
  db_instance_identifier = aws_db_instance.test.id
  feature_name           = "S3_INTEGRATION"
  role_arn               = aws_iam_role.test.arn
}
`, rName)
}
