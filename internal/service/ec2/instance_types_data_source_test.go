package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccEC2InstanceTypesDataSource_filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_instance_types.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckInstanceTypes(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypesFilterDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2InstanceTypes(dataSourceName),
				),
			},
		},
	})
}

func testAccCheckEc2InstanceTypes(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", dataSourceName)
		}

		if v := rs.Primary.Attributes["instance_types.#"]; v == "0" {
			return fmt.Errorf("expected at least one instance_types result, got none")
		}

		return nil
	}
}

func testAccPreCheckInstanceTypes(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeInstanceTypesInput{
		MaxResults: aws.Int64(5),
	}

	_, err := conn.DescribeInstanceTypes(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccInstanceTypesFilterDataSourceConfig() string {
	return `
data "aws_ec2_instance_types" "test" {
  filter {
    name   = "auto-recovery-supported"
    values = ["true"]
  }

  filter {
    name   = "network-info.encryption-in-transit-supported"
    values = ["true"]
  }

  filter {
    name   = "instance-storage-supported"
    values = ["true"]
  }
}
`
}
