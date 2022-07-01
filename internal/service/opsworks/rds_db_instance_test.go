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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_opsworks_rds_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var opsdb opsworks.RdsDbInstance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRDSDBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRDSDBInstanceConfig_basic(rName, "foo", "barbarbarbar"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRDSDBExists(resourceName, &opsdb),
					testAccCheckCreateRDSDBAttributes(&opsdb, "foo"),
					resource.TestCheckResourceAttr(resourceName, "db_user", "foo"),
				),
			},
			{
				Config: testAccRDSDBInstanceConfig_basic(rName, "bar", "barbarbarbar"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRDSDBExists(resourceName, &opsdb),
					testAccCheckCreateRDSDBAttributes(&opsdb, "bar"),
					resource.TestCheckResourceAttr(resourceName, "db_user", "bar"),
				),
			},
			{
				Config: testAccRDSDBInstanceConfig_basic(rName, "bar", "foofoofoofoofoo"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRDSDBExists(resourceName, &opsdb),
					testAccCheckCreateRDSDBAttributes(&opsdb, "bar"),
					resource.TestCheckResourceAttr(resourceName, "db_user", "bar"),
				),
			},
			{
				Config: testAccRDSDBInstanceConfig_forceNew(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRDSDBExists(resourceName, &opsdb),
					testAccCheckCreateRDSDBAttributes(&opsdb, "foo"),
					resource.TestCheckResourceAttr(resourceName, "db_user", "foo"),
				),
			},
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
		if aws.StringValue(opsdb.DbUser) != user {
			return fmt.Errorf("Unnexpected user: %s", *opsdb.DbUser)
		}
		if aws.StringValue(opsdb.Engine) != "mysql" {
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
	}
	return nil
}

func testAccRDSDBInstanceConfig_basic(rName, userName, password string) string {
	return acctest.ConfigCompose(
		testAccStackConfig_vpcCreate(rName),
		testAccDBInstanceBasicConfig(),
		fmt.Sprintf(`
resource "aws_opsworks_rds_db_instance" "test" {
  stack_id = aws_opsworks_stack.test.id

  rds_db_instance_arn = aws_db_instance.test.arn
  db_user             = %[1]q
  db_password         = %[2]q
}
`, userName, password))
}

func testAccRDSDBInstanceConfig_forceNew(rName string) string {
	return acctest.ConfigCompose(
		testAccStackConfig_vpcCreate(rName),
		testAccDBInstanceConfig_orderableClassMySQL(),
		`
resource "aws_opsworks_rds_db_instance" "test" {
  stack_id = aws_opsworks_stack.test.id

  rds_db_instance_arn = aws_db_instance.test2.arn
  db_user             = "foo"
  db_password         = "foofoofoofoo"
}

resource "aws_db_instance" "test2" {
  allocated_storage       = 10
  backup_retention_period = 0
  db_name                 = "baz"
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  maintenance_window      = "Fri:09:00-Fri:09:30"
  parameter_group_name    = "default.mysql8.0"
  password                = "foofoofoofoo"
  skip_final_snapshot     = true
  username                = "foo"
}
`)
}

func testAccDBInstanceBasicConfig() string {
	return acctest.ConfigCompose(
		testAccDBInstanceConfig_orderableClassMySQL(),
		`
resource "aws_db_instance" "test" {
  allocated_storage       = 10
  backup_retention_period = 0
  db_name                 = "baz"
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  maintenance_window      = "Fri:09:00-Fri:09:30"
  parameter_group_name    = "default.mysql8.0"
  password                = "barbarbarbar"
  skip_final_snapshot     = true
  username                = "foo"
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
