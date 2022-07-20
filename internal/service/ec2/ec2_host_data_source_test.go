package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2HostDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_ec2_host.test"
	resourceName := "aws_ec2_host.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHostDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "auto_placement", resourceName, "auto_placement"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone", resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(dataSourceName, "cores"),
					resource.TestCheckResourceAttrPair(dataSourceName, "host_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "host_recovery", resourceName, "host_recovery"),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_family", resourceName, "instance_family"),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_type", resourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "outpost_arn", resourceName, "outpost_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "sockets"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrSet(dataSourceName, "total_vcpus"),
				),
			},
		},
	})
}

func TestAccEC2HostDataSource_filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_host.test"
	resourceName := "aws_ec2_host.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHostDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "auto_placement", resourceName, "auto_placement"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone", resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(dataSourceName, "cores"),
					resource.TestCheckResourceAttrPair(dataSourceName, "host_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "host_recovery", resourceName, "host_recovery"),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_family", resourceName, "instance_family"),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_type", resourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "outpost_arn", resourceName, "outpost_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "sockets"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrSet(dataSourceName, "total_vcpus"),
				),
			},
		},
	})
}

func testAccHostDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_host" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = "a1.large"

  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_host" "test" {
  host_id = aws_ec2_host.test.id
}
`, rName))
}

func testAccHostDataSourceConfig_filter(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_host" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = "a1.large"

  tags = {
    %[1]q = "True"
  }
}

data "aws_ec2_host" "test" {
  filter {
    name   = "availability-zone"
    values = [aws_ec2_host.test.availability_zone]
  }

  filter {
    name   = "instance-type"
    values = [aws_ec2_host.test.instance_type]
  }

  filter {
    name   = "tag-key"
    values = [%[1]q]
  }
}
`, rName))
}
