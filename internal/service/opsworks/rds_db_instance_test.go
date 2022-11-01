package opsworks_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/opsworks"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfopsworks "github.com/hashicorp/terraform-provider-aws/internal/service/opsworks"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccOpsWorksRDSDBInstance_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v opsworks.RdsDbInstance
	resourceName := "aws_opsworks_rds_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRDSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRDSDBInstanceConfig_basic(rName, "user1", "password1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRDSDBInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "db_password", "password1"),
					resource.TestCheckResourceAttr(resourceName, "db_user", "user1"),
				),
			},
			{
				Config: testAccRDSDBInstanceConfig_basic(rName, "user2", "password1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRDSDBInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "db_password", "password1"),
					resource.TestCheckResourceAttr(resourceName, "db_user", "user2"),
				),
			},
			{
				Config: testAccRDSDBInstanceConfig_basic(rName, "user2", "password2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRDSDBInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "db_password", "password2"),
					resource.TestCheckResourceAttr(resourceName, "db_user", "user2"),
				),
			},
		},
	})
}

func TestAccOpsWorksRDSDBInstance_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v opsworks.RdsDbInstance
	resourceName := "aws_opsworks_rds_db_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRDSDBInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRDSDBInstanceConfig_basic(rName, "user1", "password1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRDSDBInstanceExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfopsworks.ResourceRDSDBInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRDSDBInstanceExists(n string, v *opsworks.RdsDbInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No OpsWorks RDS DB Instance ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn

		output, err := tfopsworks.FindRDSDBInstanceByTwoPartKey(conn, rs.Primary.Attributes["rds_db_instance_arn"], rs.Primary.Attributes["stack_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckRDSDBInstanceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_opsworks_rds_db_instance" {
			continue
		}

		_, err := tfopsworks.FindRDSDBInstanceByTwoPartKey(conn, rs.Primary.Attributes["rds_db_instance_arn"], rs.Primary.Attributes["stack_id"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("OpsWorks RDS DB Instance %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccRDSDBInstanceConfig_basic(rName, userName, password string) string {
	return acctest.ConfigCompose(testAccStackConfig_basic(rName), fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine         = "mysql"
  engine_version = "8.0.25"
  license_model  = "general-public-license"
  storage_type   = "standard"

  preferred_instance_classes = ["db.t3.micro", "db.t2.micro", "db.t2.medium"]
}

resource "aws_db_instance" "test" {
  identifier              = %[1]q
  allocated_storage       = 10
  backup_retention_period = 0
  db_name                 = "test"
  engine                  = data.aws_rds_orderable_db_instance.test.engine
  engine_version          = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  maintenance_window      = "Fri:09:00-Fri:09:30"
  parameter_group_name    = "default.mysql8.0"
  skip_final_snapshot     = true
  password                = "avoid-plaintext-passwords"
  username                = "tfacctest"
}

resource "aws_opsworks_rds_db_instance" "test" {
  stack_id = aws_opsworks_stack.test.id

  rds_db_instance_arn = aws_db_instance.test.arn
  db_user             = %[1]q
  db_password         = %[2]q
}
`, userName, password))
}
