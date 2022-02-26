package opsworks_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccOpsWorksRDSDBInstance_basic(t *testing.T) {
	resourceName := "aws_opsworks_rds_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var opsdb opsworks.RdsDbInstance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck: acctest.ErrorCheck(t, opsworks.EndpointsID),
		Providers:  acctest.Providers,
		//CheckDestroy: testAccCheckRDSDBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRDSDBInstance(rName, "foo", "barbarbarbar"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRDSDBExists(resourceName, &opsdb),
					testAccCheckCreateRDSDBAttributes(&opsdb, "foo"),
					resource.TestCheckResourceAttr(resourceName, "db_user", "foo"),
				),
			},
			/*{
				Config: testAccRDSDBInstance(rName, "bar", "barbarbarbar"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRDSDBExists(resourceName, &opsdb),
					testAccCheckCreateRDSDBAttributes(&opsdb, "bar"),
					resource.TestCheckResourceAttr(resourceName, "db_user", "bar"),
				),
			},
			{
				Config: testAccRDSDBInstance(rName, "bar", "foofoofoofoofoo"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRDSDBExists(resourceName, &opsdb),
					testAccCheckCreateRDSDBAttributes(&opsdb, "bar"),
					resource.TestCheckResourceAttr(resourceName, "db_user", "bar"),
				),
			},
			{
				Config: testAccRDSDBInstanceForceNew(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRDSDBExists(resourceName, &opsdb),
					testAccCheckCreateRDSDBAttributes(&opsdb, "foo"),
					resource.TestCheckResourceAttr(resourceName, "db_user", "foo"),
				),
			},*/
		},
	})
}

func testAccCheckRDSDBExists(
	n string, opsdb *opsworks.RdsDbInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		if _, ok := rs.Primary.Attributes["stack_id"]; !ok {
			return fmt.Errorf("Rds Db stack id is missing, should be set.")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn

		params := &opsworks.DescribeRdsDbInstancesInput{
			StackId: aws.String(rs.Primary.Attributes["stack_id"]),
		}
		resp, err := conn.DescribeRdsDbInstances(params)

		if err != nil {
			return err
		}

		if v := len(resp.RdsDbInstances); v != 1 {
			return fmt.Errorf("Expected 1 response returned, got %d", v)
		}

		*opsdb = *resp.RdsDbInstances[0]

		return nil
	}
}

func testAccCheckCreateRDSDBAttributes(
	opsdb *opsworks.RdsDbInstance, user string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *opsdb.DbUser != user {
			return fmt.Errorf("Unnexpected user: %s", *opsdb.DbUser)
		}
		if *opsdb.Engine != "mysql" {
			return fmt.Errorf("Unnexpected engine: %s", *opsdb.Engine)
		}
		return nil
	}
}

func testAccCheckRDSDBDestroy(s *terraform.State) error {
	client := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_opsworks_rds_db_instance" {
			continue
		}

		req := &opsworks.DescribeRdsDbInstancesInput{
			StackId: aws.String(rs.Primary.Attributes["stack_id"]),
		}

		resp, err := client.DescribeRdsDbInstances(req)
		if err == nil {
			if len(resp.RdsDbInstances) > 0 {
				return fmt.Errorf("OpsWorks Rds db instances  still exist.")
			}
		}

		if !tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
			return err
		}
		/*
			if awserr, ok := err.(awserr.Error); ok {
				if awserr.Code() != "ResourceNotFoundException" {
					return fmt.Errorf("checking RDS DB destroy: %w", err)
				}
			}
		*/
	}
	return nil
}

func testAccRDSDBInstance(rName, userName, password string) string {
	return acctest.ConfigCompose(
		testAccStackVPCCreateConfig(rName),
		testAccDBInstanceBasicConfig(),
		fmt.Sprintf(`
resource "aws_opsworks_rds_db_instance" "test" {
  stack_id = aws_opsworks_stack.tf-acc.id

  rds_db_instance_arn = aws_db_instance.bar.arn
  db_user             = %[1]q
  db_password         = %[2]q
}
`, userName, password))
}

func testAccRDSDBInstanceForceNew(rName string) string {
	return acctest.ConfigCompose(
		testAccStackVPCCreateConfig(rName),
		`
resource "aws_opsworks_rds_db_instance" "test" {
  stack_id = aws_opsworks_stack.tf-acc.id

  rds_db_instance_arn = aws_db_instance.foo.arn
  db_user             = "foo"
  db_password         = "foofoofoofoo"
}

resource "aws_db_instance" "foo" {
  allocated_storage    = 10
  engine               = "mysql"
  engine_version       = "8.0.25"
  instance_class       = "db.t2.micro"
  name                 = "baz"
  password             = "foofoofoofoo"
  username             = "foo"
  parameter_group_name = "default.mysql8.0"

  skip_final_snapshot = true
}
`)
}

func testAccDBInstanceBasicConfig() string {
	return acctest.ConfigCompose(
		testAccDBInstanceConfig_orderableClassMySQL(),
		`
resource "aws_db_instance" "bar" {
  allocated_storage       = 10
  backup_retention_period = 0
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  name                    = "baz"
  parameter_group_name    = "default.mysql8.0"
  password                = "barbarbarbar"
  skip_final_snapshot     = true
  username                = "foo"

  # Maintenance Window is stored in lower case in the API, though not strictly
  # documented. Terraform will downcase this to match (as opposed to throw a
  # validation error).
  maintenance_window = "Fri:09:00-Fri:09:30"
}
`)
}

func testAccDBInstanceConfig_orderableClass(engine, version, license string) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine         = %[1]q
  engine_version = %[2]q
  license_model  = %[3]q
  storage_type   = "standard"

  preferred_instance_classes = ["db.t3.micro", "db.t2.micro", "db.t2.medium"]
}
`, engine, version, license)
}

func testAccDBInstanceConfig_orderableClassMySQL() string {
	return testAccDBInstanceConfig_orderableClass("mysql", "8.0.25", "general-public-license")
}
