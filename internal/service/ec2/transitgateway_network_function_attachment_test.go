// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTransitGatewayNetworkFunctionAttachment_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayNetworkFunctionAttachment
	resourceName := "aws_ec2_transit_gateway_network_function_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayNetworkFunctionAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayNetworkFunctionAttachmentConfig_availabilityZoneIDs(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayNetworkFunctionAttachmentExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceName, "aws_vpc_endpoint_service.test", names.AttrServiceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayID, "aws_ec2_transit_gateway.test", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "availability_zone_ids.*", "aws_subnet.test.0", "availability_zone_id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "availability_zones.*", "aws_subnet.test.0", names.AttrAvailabilityZone),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("ec2", regexache.MustCompile(`transit-gateway-attachment/tgw-attach-.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("availability_zone_ids"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAvailabilityZones), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTransitGatewayNetworkFunctionAttachment_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayNetworkFunctionAttachment
	resourceName := "aws_ec2_transit_gateway_network_function_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayNetworkFunctionAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayNetworkFunctionAttachmentConfig_availabilityZoneIDs(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayNetworkFunctionAttachmentExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfec2.ResourceTransitGatewayNetworkFunctionAttachment, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccTransitGatewayNetworkFunctionAttachment_tags(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v awstypes.TransitGatewayNetworkFunctionAttachment
	resourceName := "aws_ec2_transit_gateway_network_function_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayNetworkFunctionAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayNetworkFunctionAttachmentConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayNetworkFunctionAttachmentExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayNetworkFunctionAttachmentConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayNetworkFunctionAttachmentExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						// A tags-only change must not touch the zones.
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("availability_zone_ids"), knownvalue.SetSizeExact(1)),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAvailabilityZones), knownvalue.SetSizeExact(1)),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccTransitGatewayNetworkFunctionAttachmentConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayNetworkFunctionAttachmentExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccTransitGatewayNetworkFunctionAttachment_availabilityZoneIDs(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v1, v2, v3 awstypes.TransitGatewayNetworkFunctionAttachment
	resourceName := "aws_ec2_transit_gateway_network_function_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayNetworkFunctionAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayNetworkFunctionAttachmentConfig_availabilityZoneIDs(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayNetworkFunctionAttachmentExists(ctx, t, resourceName, &v1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("availability_zone_ids"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAvailabilityZones), knownvalue.SetSizeExact(1)),
				},
			},
			{
				// Add a zone: updated in place.
				Config: testAccTransitGatewayNetworkFunctionAttachmentConfig_availabilityZoneIDs(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayNetworkFunctionAttachmentExists(ctx, t, resourceName, &v2),
					testAccCheckTransitGatewayNetworkFunctionAttachmentNotRecreated(&v1, &v2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "availability_zones.*", "aws_subnet.test.1", names.AttrAvailabilityZone),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("availability_zone_ids"), knownvalue.SetSizeExact(2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAvailabilityZones), knownvalue.SetSizeExact(2)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Remove a zone: updated in place.
				Config: testAccTransitGatewayNetworkFunctionAttachmentConfig_availabilityZoneIDs(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayNetworkFunctionAttachmentExists(ctx, t, resourceName, &v3),
					testAccCheckTransitGatewayNetworkFunctionAttachmentNotRecreated(&v2, &v3),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("availability_zone_ids"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAvailabilityZones), knownvalue.SetSizeExact(1)),
				},
			},
		},
	})
}

func testAccTransitGatewayNetworkFunctionAttachment_availabilityZones(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.TransitGatewayNetworkFunctionAttachment
	resourceName := "aws_ec2_transit_gateway_network_function_attachment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayNetworkFunctionAttachmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayNetworkFunctionAttachmentConfig_availabilityZones(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayNetworkFunctionAttachmentExists(ctx, t, resourceName, &v1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "availability_zones.*", "aws_subnet.test.0", names.AttrAvailabilityZone),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "availability_zone_ids.*", "aws_subnet.test.0", "availability_zone_id"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("availability_zone_ids"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAvailabilityZones), knownvalue.SetSizeExact(1)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Add a zone by name: updated in place, and the zone IDs follow.
				Config: testAccTransitGatewayNetworkFunctionAttachmentConfig_availabilityZones(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTransitGatewayNetworkFunctionAttachmentExists(ctx, t, resourceName, &v2),
					testAccCheckTransitGatewayNetworkFunctionAttachmentNotRecreated(&v1, &v2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "availability_zone_ids.*", "aws_subnet.test.1", "availability_zone_id"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("availability_zone_ids"), knownvalue.SetSizeExact(2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAvailabilityZones), knownvalue.SetSizeExact(2)),
				},
			},
		},
	})
}

func testAccTransitGatewayNetworkFunctionAttachment_identity(t *testing.T, _ tfsync.Semaphore) {
	testAccTransitGatewayTransitGatewayNetworkFunctionAttachment_identitySerial(t)
}

func testAccCheckTransitGatewayNetworkFunctionAttachmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_transit_gateway_network_function_attachment" {
				continue
			}

			_, err := tfec2.FindTransitGatewayNetworkFunctionAttachmentByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Transit Gateway Network Function Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTransitGatewayNetworkFunctionAttachmentExists(ctx context.Context, t *testing.T, n string, v *awstypes.TransitGatewayNetworkFunctionAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EC2Client(ctx)

		output, err := tfec2.FindTransitGatewayNetworkFunctionAttachmentByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTransitGatewayNetworkFunctionAttachmentNotRecreated(before, after *awstypes.TransitGatewayNetworkFunctionAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := *before.TransitGatewayAttachmentId, *after.TransitGatewayAttachmentId; before != after {
			return fmt.Errorf("EC2 Transit Gateway Network Function Attachment recreated: %s -> %s", before, after)
		}

		return nil
	}
}

// A Gateway Load Balancer in two Availability Zones, exposed as a VPC endpoint service
// that accepts connections automatically, and a Transit Gateway to attach it to.
func testAccTransitGatewayNetworkFunctionAttachmentConfig_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_lb" "test" {
  load_balancer_type = "gateway"
  name               = %[1]q
  subnets            = aws_subnet.test[*].id
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  gateway_load_balancer_arns = [aws_lb.test.arn]

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}
`, rName),
	)
}

func testAccTransitGatewayNetworkFunctionAttachmentConfig_availabilityZoneIDs(rName string, zoneCount int) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayNetworkFunctionAttachmentConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway_network_function_attachment" "test" {
  availability_zone_ids = slice(aws_subnet.test[*].availability_zone_id, 0, %[1]d)
  service_name          = aws_vpc_endpoint_service.test.service_name
  transit_gateway_id    = aws_ec2_transit_gateway.test.id
}
`, zoneCount),
	)
}

func testAccTransitGatewayNetworkFunctionAttachmentConfig_availabilityZones(rName string, zoneCount int) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayNetworkFunctionAttachmentConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway_network_function_attachment" "test" {
  availability_zones = slice(aws_subnet.test[*].availability_zone, 0, %[1]d)
  service_name       = aws_vpc_endpoint_service.test.service_name
  transit_gateway_id = aws_ec2_transit_gateway.test.id
}
`, zoneCount),
	)
}

func testAccTransitGatewayNetworkFunctionAttachmentConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayNetworkFunctionAttachmentConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway_network_function_attachment" "test" {
  availability_zone_ids = [aws_subnet.test[0].availability_zone_id]
  service_name          = aws_vpc_endpoint_service.test.service_name
  transit_gateway_id    = aws_ec2_transit_gateway.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tag1Key, tag1Value),
	)
}

func testAccTransitGatewayNetworkFunctionAttachmentConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(
		testAccTransitGatewayNetworkFunctionAttachmentConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ec2_transit_gateway_network_function_attachment" "test" {
  availability_zone_ids = [aws_subnet.test[0].availability_zone_id]
  service_name          = aws_vpc_endpoint_service.test.service_name
  transit_gateway_id    = aws_ec2_transit_gateway.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tag1Key, tag1Value, tag2Key, tag2Value),
	)
}
