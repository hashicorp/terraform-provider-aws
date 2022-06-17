package rds_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRDSInstanceDataSource_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "address"),
					resource.TestCheckResourceAttrSet(dataSourceName, "allocated_storage"),
					resource.TestCheckResourceAttrSet(dataSourceName, "auto_minor_version_upgrade"),
					resource.TestCheckResourceAttrSet(dataSourceName, "db_instance_class"),
					resource.TestCheckResourceAttrSet(dataSourceName, "db_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "db_subnet_group"),
					resource.TestCheckResourceAttrSet(dataSourceName, "endpoint"),
					resource.TestCheckResourceAttrSet(dataSourceName, "engine"),
					resource.TestCheckResourceAttrSet(dataSourceName, "hosted_zone_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "master_username"),
					resource.TestCheckResourceAttrSet(dataSourceName, "port"),
					resource.TestCheckResourceAttrSet(dataSourceName, "multi_az"),
					resource.TestCheckResourceAttrSet(dataSourceName, "enabled_cloudwatch_logs_exports.0"),
					resource.TestCheckResourceAttrSet(dataSourceName, "enabled_cloudwatch_logs_exports.1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "resource_id", "aws_db_instance.test", "resource_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", "aws_db_instance.test", "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Environment", "aws_db_instance.test", "tags.Environment"),
				),
			},
		},
	})
}

func TestAccRDSInstanceDataSource_ec2Classic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_db_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckEC2Classic(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceDataSourceConfig_ec2Classic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "db_subnet_group", ""),
				),
			},
		},
	})
}

func testAccInstanceDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_orderableClassMariadb(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_db_instance" "test" {
  allocated_storage       = 10
  backup_retention_period = 0
  db_subnet_group_name    = aws_db_subnet_group.test.name
  engine                  = data.aws_rds_engine_version.default.engine
  engine_version          = data.aws_rds_engine_version.default.version
  identifier              = %[1]q
  instance_class          = data.aws_rds_orderable_db_instance.test.instance_class
  name                    = "baz"
  password                = "barbarbarbar"
  skip_final_snapshot     = true
  username                = "foo"

  enabled_cloudwatch_logs_exports = [
    "audit",
    "error",
  ]

  tags = {
    Environment = "test"
  }
}

data "aws_db_instance" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
}
`, rName))
}

func testAccInstanceDataSourceConfig_ec2Classic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigEC2ClassicRegionProvider(),
		fmt.Sprintf(`
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

# EC2-Classic specific
data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  preferred_instance_classes = ["db.m3.medium", "db.m3.large", "db.r3.large"]
}

resource "aws_db_instance" "test" {
  identifier           = %[1]q
  allocated_storage    = 10
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  storage_type         = data.aws_rds_orderable_db_instance.test.storage_type
  db_name              = "baz"
  password             = "barbarbarbar"
  username             = "foo"
  publicly_accessible  = true
  security_group_names = ["default"]
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
  skip_final_snapshot  = true
}

data "aws_db_instance" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
}
`, rName))
}
