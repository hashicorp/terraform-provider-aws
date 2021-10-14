package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func TestAccAWSVpc_coreMismatchedDiffs(t *testing.T) {
	var vpc ec2.Vpc

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testMatchedDiffs,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists("aws_vpc.test", &vpc),
					testAccCheckVpcCidr(&vpc, "10.0.0.0/16"),
					resource.TestCheckResourceAttr(
						"aws_vpc.test", "cidr_block", "10.0.0.0/16"),
				),
			},
		},
	})
}

const testMatchedDiffs = `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-repro-GH-4965"
  }

  lifecycle {
    ignore_changes = ["tags"]
  }
}`
