package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSIotEndpoint(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSIotEndpointConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aws_iot_endpoint.example", "id"),
					resource.TestCheckResourceAttrSet("aws_iot_endpoint.example", "endpoint_address"),
				),
			},
		},
	})
}

const testAccAWSIotEndpointConfig = `
data "aws_iot_endpoint" "example" {}
`
