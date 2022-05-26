package outposts_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccOutpostsOutpostInstanceTypeDataSource_instanceType(t *testing.T) {
	dataSourceName := "data.aws_outposts_outpost_instance_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, outposts.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostInstanceTypeDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "instance_type", regexp.MustCompile(`^.+$`)),
				),
			},
		},
	})
}

func TestAccOutpostsOutpostInstanceTypeDataSource_preferredInstanceTypes(t *testing.T) {
	dataSourceName := "data.aws_outposts_outpost_instance_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, outposts.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostInstanceTypeDataSourceConfig_preferreds(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "instance_type", regexp.MustCompile(`^.+$`)),
				),
			},
		},
	})
}

func testAccOutpostInstanceTypeDataSourceConfig_basic() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost_instance_types" "test" {
  arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}

data "aws_outposts_outpost_instance_type" "test" {
  arn           = tolist(data.aws_outposts_outposts.test.arns)[0]
  instance_type = tolist(data.aws_outposts_outpost_instance_types.test.instance_types)[0]
}
`
}

func testAccOutpostInstanceTypeDataSourceConfig_preferreds() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost_instance_types" "test" {
  arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}

data "aws_outposts_outpost_instance_type" "test" {
  arn                      = tolist(data.aws_outposts_outposts.test.arns)[0]
  preferred_instance_types = data.aws_outposts_outpost_instance_types.test.instance_types
}
`
}
