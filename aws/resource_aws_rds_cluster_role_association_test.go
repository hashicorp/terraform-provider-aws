package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfrds "github.com/hashicorp/terraform-provider-aws/aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/rds/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

func TestAccAWSRDSClusterRoleAssociation_basic(t *testing.T) {
	var dbClusterRole rds.DBClusterRole
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dbClusterResourceName := "aws_rds_cluster.test"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_rds_cluster_role_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRDSClusterRoleAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterRoleAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRDSClusterRoleAssociationExists(resourceName, &dbClusterRole),
					resource.TestCheckResourceAttrPair(resourceName, "db_cluster_identifier", dbClusterResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "feature_name", "s3Import"),
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

func TestAccAWSRDSClusterRoleAssociation_disappears(t *testing.T) {
	var dbClusterRole rds.DBClusterRole
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_cluster_role_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRDSClusterRoleAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterRoleAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRDSClusterRoleAssociationExists(resourceName, &dbClusterRole),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRDSClusterRoleAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRDSClusterRoleAssociation_disappears_cluster(t *testing.T) {
	var dbClusterRole rds.DBClusterRole
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_cluster_role_association.test"
	clusterResourceName := "aws_rds_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRDSClusterRoleAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterRoleAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRDSClusterRoleAssociationExists(resourceName, &dbClusterRole),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRDSCluster(), clusterResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRDSClusterRoleAssociation_disappears_role(t *testing.T) {
	var dbClusterRole rds.DBClusterRole
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_rds_cluster_role_association.test"
	roleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRDSClusterRoleAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSClusterRoleAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRDSClusterRoleAssociationExists(resourceName, &dbClusterRole),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsIamRole(), roleResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSRDSClusterRoleAssociationExists(resourceName string, v *rds.DBClusterRole) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		dbClusterID, roleARN, err := tfrds.ClusterRoleAssociationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		role, err := finder.DBClusterRoleByDBClusterIDAndRoleARN(conn, dbClusterID, roleARN)

		if err != nil {
			return err
		}

		*v = *role

		return nil
	}
}

func testAccCheckAWSRDSClusterRoleAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_cluster_role_association" {
			continue
		}

		dbClusterID, roleARN, err := tfrds.ClusterRoleAssociationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = finder.DBClusterRoleByDBClusterIDAndRoleARN(conn, dbClusterID, roleARN)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("RDS DB Cluster IAM Role Association %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAWSRDSClusterRoleAssociationConfig(rName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
data "aws_iam_policy_document" "rds_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      identifiers = ["rds.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_iam_role" "test" {
  assume_role_policy = data.aws_iam_policy_document.rds_assume_role_policy.json
  name               = %[1]q
}

resource "aws_rds_cluster" "test" {
  cluster_identifier      = %[1]q
  engine                  = "aurora-postgresql"
  availability_zones      = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1], data.aws_availability_zones.available.names[2]]
  database_name           = "mydb"
  master_username         = "foo"
  master_password         = "foobarfoobarfoobar"
  backup_retention_period = 5
  preferred_backup_window = "07:00-09:00"
  skip_final_snapshot     = true
}

resource "aws_rds_cluster_role_association" "test" {
  db_cluster_identifier = aws_rds_cluster.test.id
  feature_name          = "s3Import"
  role_arn              = aws_iam_role.test.arn
}
`, rName))
}
