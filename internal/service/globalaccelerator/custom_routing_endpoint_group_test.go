// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/globalaccelerator/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfglobalaccelerator "github.com/hashicorp/terraform-provider-aws/internal/service/globalaccelerator"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlobalAcceleratorCustomRoutingEndpointGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.CustomRoutingEndpointGroup
	resourceName := "aws_globalaccelerator_custom_routing_endpoint_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomRoutingEndpointGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomRoutingEndpointGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomRoutingEndpointGroupExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "globalaccelerator", regexache.MustCompile(`accelerator/`+verify.UUIDRegexPattern+`/listener/[a-z0-9]{8}/endpoint-group/[a-z0-9]{12}$`)),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.from_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.protocols.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "destination_configuration.0.protocols.*", "TCP"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.to_port", "8443"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_group_region", acctest.Region()),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccGlobalAcceleratorCustomRoutingEndpointGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.CustomRoutingEndpointGroup
	resourceName := "aws_globalaccelerator_custom_routing_endpoint_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomRoutingEndpointGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomRoutingEndpointGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomRoutingEndpointGroupExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfglobalaccelerator.ResourceCustomRoutingEndpointGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGlobalAcceleratorCustomRoutingEndpointGroup_endpointConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.CustomRoutingEndpointGroup
	resourceName := "aws_globalaccelerator_custom_routing_endpoint_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomRoutingEndpointGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomRoutingEndpointGroupConfig_endpointConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomRoutingEndpointGroupExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "globalaccelerator", regexache.MustCompile(`accelerator/`+verify.UUIDRegexPattern+`/listener/[a-z0-9]{8}/endpoint-group/[a-z0-9]{12}$`)),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.from_port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.protocols.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "destination_configuration.0.protocols.*", "TCP"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.to_port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_configuration.0.endpoint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_group_region"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccGlobalAcceleratorCustomRoutingEndpointGroup_endpointGroupRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.CustomRoutingEndpointGroup
	resourceName := "aws_globalaccelerator_custom_routing_endpoint_group.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomRoutingEndpointGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomRoutingEndpointGroupConfig_endpointGroupRegion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomRoutingEndpointGroupExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "globalaccelerator", regexache.MustCompile(`accelerator/`+verify.UUIDRegexPattern+`/listener/[a-z0-9]{8}/endpoint-group/[a-z0-9]{12}$`)),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.from_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.protocols.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "destination_configuration.0.protocols.*", "TCP"),
					resource.TestCheckResourceAttr(resourceName, "destination_configuration.0.to_port", "8443"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_group_region", acctest.AlternateRegion()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckCustomRoutingEndpointGroupExists(ctx context.Context, t *testing.T, n string, v *awstypes.CustomRoutingEndpointGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).GlobalAcceleratorClient(ctx)

		output, err := tfglobalaccelerator.FindCustomRoutingEndpointGroupByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckCustomRoutingEndpointGroupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).GlobalAcceleratorClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_globalaccelerator_custom_routing_endpoint_group" {
				continue
			}

			_, err := tfglobalaccelerator.FindCustomRoutingEndpointGroupByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Global Accelerator Custom Routing Endpoint Group %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCustomRoutingEndpointGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_custom_routing_accelerator" "test" {
  name = %[1]q
}

resource "aws_globalaccelerator_custom_routing_listener" "test" {
  accelerator_arn = aws_globalaccelerator_custom_routing_accelerator.test.arn

  port_range {
    from_port = 443
    to_port   = 443
  }
}

resource "aws_globalaccelerator_custom_routing_endpoint_group" "test" {
  listener_arn = aws_globalaccelerator_custom_routing_listener.test.arn

  destination_configuration {
    from_port = 443
    to_port   = 8443
    protocols = ["TCP"]
  }
}
`, rName)
}

func testAccCustomRoutingEndpointGroupConfig_endpointConfiguration(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_globalaccelerator_custom_routing_accelerator" "test" {
  name = %[1]q
}

resource "aws_globalaccelerator_custom_routing_listener" "test" {
  accelerator_arn = aws_globalaccelerator_custom_routing_accelerator.test.arn

  port_range {
    from_port = 1
    to_port   = 65534
  }
}

resource "aws_globalaccelerator_custom_routing_endpoint_group" "test" {
  listener_arn = aws_globalaccelerator_custom_routing_listener.test.arn

  destination_configuration {
    from_port = 8080
    to_port   = 8081
    protocols = ["TCP"]
  }

  endpoint_configuration {
    endpoint_id = aws_subnet.test.id
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/28"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccCustomRoutingEndpointGroupConfig_endpointGroupRegion(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_custom_routing_accelerator" "test" {
  name = %[1]q
}

resource "aws_globalaccelerator_custom_routing_listener" "test" {
  accelerator_arn = aws_globalaccelerator_custom_routing_accelerator.test.arn

  port_range {
    from_port = 443
    to_port   = 443
  }
}

resource "aws_globalaccelerator_custom_routing_endpoint_group" "test" {
  listener_arn = aws_globalaccelerator_custom_routing_listener.test.arn

  destination_configuration {
    from_port = 443
    to_port   = 8443
    protocols = ["TCP"]
  }

  endpoint_group_region = %[2]q
}
`, rName, acctest.AlternateRegion())
}
