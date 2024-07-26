// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcegroups_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/resourcegroups/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfresourcegroups "github.com/hashicorp/terraform-provider-aws/internal/service/resourcegroups"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccResourceGroupsResource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var r types.ListGroupResourcesItem
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_resourcegroups_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ResourceGroupsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, names.AttrResourceType, "AWS::EC2::Host"),
					resource.TestCheckResourceAttrSet(resourceName, "group_arn"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceARN),
				),
			},
		},
	})
}

func testAccCheckResourceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ResourceGroupsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_resourcegroups_resource" {
				continue
			}

			_, err := tfresourcegroups.FindResourceByTwoPartKey(ctx, conn, rs.Primary.Attributes["group_arn"], rs.Primary.Attributes[names.AttrResourceARN])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Resource Groups Resource %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckResourceExists(ctx context.Context, n string, v *types.ListGroupResourcesItem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ResourceGroupsClient(ctx)

		output, err := tfresourcegroups.FindResourceByTwoPartKey(ctx, conn, rs.Primary.Attributes["group_arn"], rs.Primary.Attributes[names.AttrResourceARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccResourceConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ec2_host" "test" {
  auto_placement    = "on"
  availability_zone = data.aws_availability_zones.available.names[0]
  host_recovery     = "off"
  instance_family   = "c5"

  tags = {
    Name = %[1]q
  }
}

resource "aws_resourcegroups_group" "test" {
  name = %[1]q

  configuration {
    type = "AWS::EC2::HostManagement"
    parameters {
      name = "any-host-based-license-configuration"
      values = [
        "true"
      ]
    }

    parameters {
      name = "auto-allocate-host"
      values = [
        "false"
      ]
    }

    parameters {
      name = "auto-host-recovery"
      values = [
        "false"
      ]
    }

    parameters {
      name = "auto-release-host"
      values = [
        "false"
      ]
    }
  }

  configuration {
    type = "AWS::ResourceGroups::Generic"

    parameters {
      name = "allowed-resource-types"
      values = [
        "AWS::EC2::Host"
      ]
    }

    parameters {
      name = "deletion-protection"
      values = [
        "UNLESS_EMPTY"
      ]
    }
  }

  depends_on = [aws_ec2_host.test]
}
`, rName))
}

func testAccResourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccResourceConfig_base(rName), `
resource "aws_resourcegroups_resource" "test" {
  group_arn    = aws_resourcegroups_group.test.arn
  resource_arn = aws_ec2_host.test.arn
}
	`)
}
