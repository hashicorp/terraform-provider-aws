package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSEc2HostDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_ec2_host.test"
	resourceName := "aws_ec2_host.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, ec2.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedHostDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_type", resourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone", resourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_family", resourceName, "instance_family"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cores", resourceName, "cores"),
					resource.TestCheckResourceAttrPair(dataSourceName, "total_vcpus", resourceName, "total_vcpus"),
					resource.TestCheckResourceAttrPair(dataSourceName, "sockets", resourceName, "sockets"),
					resource.TestCheckResourceAttrPair(dataSourceName, "host_recovery", resourceName, "host_recovery"),
					resource.TestCheckResourceAttrPair(dataSourceName, "auto_placement", resourceName, "auto_placement"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.tag1", "test-value1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.tag2", "test-value2"),
				),
			},
		},
	})
}

const testAccDedicatedHostDataSourceConfig = `
resource "aws_ec2_host" "test" {
   instance_type = "c5.xlarge"
   availability_zone = "${data.aws_availability_zones.available.names[0]}"
   host_recovery = "on"
   auto_placement = "on"
	tags = {
    tag1 = "test-value1"
    tag2 = "test-value2"
  }
}


data "aws_ec2_host" "test" {
  host_id = aws_ec2_host.test.id
}

data "aws_availability_zones" "available" {
	state = "available"
  
	filter {
	  name   = "opt-in-status"
	  values = ["opt-in-not-required"]
	}
  }
  
	

`
