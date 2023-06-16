package resourcegroups_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfresourcegroups "github.com/hashicorp/terraform-provider-aws/internal/service/resourcegroups"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccResourceGroupsResource_basic(t *testing.T) {
	ctx := context.Background()
	var r resourcegroups.ListGroupResourcesItem
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_resourcegroups_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, resourcegroups.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExists(ctx, resourceName, &r),
					resource.TestCheckResourceAttr(resourceName, "resource_type", "AWS::EC2::Host"),
					resource.TestCheckResourceAttrSet(resourceName, "group_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_arn"),
				),
			},
		},
	})
}

func testAccCheckResourceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ResourceGroupsConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_resourcegroups_resource" {
				continue
			}

			_, err := tfresourcegroups.FindResourceByARN(ctx, conn, rs.Primary.Attributes["group_arn"], rs.Primary.Attributes["resource_arn"])

			if err != nil {
				if tfawserr.ErrCodeEquals(err, resourcegroups.ErrCodeNotFoundException) {
					return nil
				}
				return err
			}

			return create.Error(names.ResourceGroups, create.ErrActionCheckingDestroyed, tfresourcegroups.ResNameResource, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckResourceExists(ctx context.Context, name string, resource *resourcegroups.ListGroupResourcesItem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ResourceGroups, create.ErrActionCheckingExistence, tfresourcegroups.ResNameResource, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ResourceGroups, create.ErrActionCheckingExistence, tfresourcegroups.ResNameResource, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ResourceGroupsConn(ctx)

		resp, err := tfresourcegroups.FindResourceByARN(ctx, conn, rs.Primary.Attributes["group_arn"], rs.Primary.Attributes["resource_arn"])

		if err != nil {
			return create.Error(names.ResourceGroups, create.ErrActionCheckingExistence, tfresourcegroups.ResNameResource, rs.Primary.ID, err)
		}

		*resource = *resp

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
