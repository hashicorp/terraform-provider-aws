// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53HealthCheckSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.HealthCheckSource
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_health_check_source.test"
	vpcResourceName := "aws_vpc.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHealthCheckSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHealthCheckSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHealthCheckSourceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_configuration.*.subnet_id", "aws_subnet.test.0", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_configuration.*.availability_zone_id", "aws_subnet.test.0", "availability_zone_id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_configuration.*.security_group_ids.*", "aws_security_group.test", names.AttrID),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), knownvalue.StringRegexp(regexache.MustCompile(`^arn:aws[a-z-]*:route53:::healthchecksource/.+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.HealthCheckSourceStatusCreateCompleted)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("subnet_configuration"), knownvalue.SetSizeExact(3)),
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

func TestAccRoute53HealthCheckSource_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.HealthCheckSource
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_health_check_source.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHealthCheckSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHealthCheckSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHealthCheckSourceExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfroute53.ResourceHealthCheckSource, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

// Tags are the only thing that can change without replacing the source.
func TestAccRoute53HealthCheckSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.HealthCheckSource
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_health_check_source.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHealthCheckSourceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHealthCheckSourceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHealthCheckSourceExists(ctx, t, resourceName, &v),
				),
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
				Config: testAccHealthCheckSourceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHealthCheckSourceExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
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
				Config: testAccHealthCheckSourceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHealthCheckSourceExists(ctx, t, resourceName, &v),
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

// Plan-time validation only: this test creates nothing.
func TestAccRoute53HealthCheckSource_duplicateAvailabilityZoneID(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config:      testAccHealthCheckSourceConfig_duplicateAvailabilityZoneID(),
				ExpectError: regexache.MustCompile(`Duplicate Availability Zone`),
			},
		},
	})
}

func testAccCheckHealthCheckSourceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).Route53Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_health_check_source" {
				continue
			}

			_, err := tfroute53.FindHealthCheckSourceByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route 53 Health Check Source %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckHealthCheckSourceExists(ctx context.Context, t *testing.T, n string, v *awstypes.HealthCheckSource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).Route53Client(ctx)

		output, err := tfroute53.FindHealthCheckSourceByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

// testAccHealthCheckSourceConfig_base creates one subnet in each of three
// Availability Zones and a security group that allows egress inside the VPC.
func testAccHealthCheckSourceConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 3), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = [aws_vpc.test.cidr_block]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccHealthCheckSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccHealthCheckSourceConfig_base(rName), `
resource "aws_route53_health_check_source" "test" {
  vpc_id = aws_vpc.test.id

  dynamic "subnet_configuration" {
    for_each = aws_subnet.test

    content {
      availability_zone_id = subnet_configuration.value.availability_zone_id
      security_group_ids   = [aws_security_group.test.id]
      subnet_id            = subnet_configuration.value.id
    }
  }
}
`)
}

func testAccHealthCheckSourceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccHealthCheckSourceConfig_base(rName), fmt.Sprintf(`
resource "aws_route53_health_check_source" "test" {
  vpc_id = aws_vpc.test.id

  dynamic "subnet_configuration" {
    for_each = aws_subnet.test

    content {
      availability_zone_id = subnet_configuration.value.availability_zone_id
      security_group_ids   = [aws_security_group.test.id]
      subnet_id            = subnet_configuration.value.id
    }
  }

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccHealthCheckSourceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccHealthCheckSourceConfig_base(rName), fmt.Sprintf(`
resource "aws_route53_health_check_source" "test" {
  vpc_id = aws_vpc.test.id

  dynamic "subnet_configuration" {
    for_each = aws_subnet.test

    content {
      availability_zone_id = subnet_configuration.value.availability_zone_id
      security_group_ids   = [aws_security_group.test.id]
      subnet_id            = subnet_configuration.value.id
    }
  }

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccHealthCheckSourceConfig_duplicateAvailabilityZoneID() string {
	return `
resource "aws_route53_health_check_source" "test" {
  vpc_id = "vpc-12345678"

  subnet_configuration {
    availability_zone_id = "use1-az1"
    security_group_ids   = ["sg-12345678"]
    subnet_id            = "subnet-11111111"
  }

  subnet_configuration {
    availability_zone_id = "use1-az1"
    security_group_ids   = ["sg-12345678"]
    subnet_id            = "subnet-22222222"
  }
}
`
}
