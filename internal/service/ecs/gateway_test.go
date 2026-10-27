// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSGateway_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ecs", regexache.MustCompile(`gateway/`+rName+`/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "access_type", "PUBLIC"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.host_shared_alb.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.host_shared_alb.0.network_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.host_shared_alb.0.network_configuration.0.subnets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.host_shared_alb.0.scheme", "internet-facing"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "infrastructure_role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccECSGateway_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfecs.ResourceGateway, resourceName),
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

func TestAccECSGateway_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccGatewayConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccGatewayConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccECSGateway_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "configuration.0.host_shared_alb.0.network_configuration.0.inbound_security_group"),
				),
			},
			{
				Config: testAccGatewayConfig_updated(rName, "monitor", 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.host_shared_alb.0.http_desync", "monitor"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.host_shared_alb.0.max_alb_count", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.host_shared_alb.0.network_configuration.0.inbound_security_group", "aws_security_group.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.host_shared_alb.0.test_traffic_rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.host_shared_alb.0.test_traffic_rules.0.header.0.name", "x-test"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, "GATEWAY"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccGatewayConfig_updated(rName, "strictest", 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.host_shared_alb.0.http_desync", "strictest"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.host_shared_alb.0.max_alb_count", "3"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccECSGateway_internal(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_internal(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_type", "PRIVATE"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.host_shared_alb.0.scheme", "internal"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func testAccCheckGatewayDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecs_gateway" {
				continue
			}

			_, err := tfecs.FindGatewayByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ECS Gateway %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckGatewayExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

		_, err := tfecs.FindGatewayByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		return err
	}
}

func testAccGatewayConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ecs.${data.aws_partition.current.dns_suffix}" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonECSInfrastructureRoleForGatewayHostSharedALB"
}
`, rName))
}

func testAccGatewayConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_gateway" "test" {
  name                    = %[1]q
  infrastructure_role_arn = aws_iam_role.test.arn
  deletion_protection     = false

  configuration {
    host_shared_alb {
      network_configuration {
        subnets = aws_subnet.test[*].id
      }
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccGatewayConfig_internal(rName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_gateway" "test" {
  name                    = %[1]q
  infrastructure_role_arn = aws_iam_role.test.arn
  deletion_protection     = false

  configuration {
    host_shared_alb {
      scheme = "internal"

      network_configuration {
        subnets = aws_subnet.test[*].id
      }
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccGatewayConfig_updated(rName, httpDesync string, maxALBCount int) string {
	return acctest.ConfigCompose(testAccGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ecs_gateway" "test" {
  name                    = %[1]q
  infrastructure_role_arn = aws_iam_role.test.arn
  deletion_protection     = false
  propagate_tags          = "GATEWAY"

  configuration {
    host_shared_alb {
      http_desync   = %[2]q
      max_alb_count = %[3]d

      network_configuration {
        inbound_security_group = aws_security_group.test.id
        subnets                = aws_subnet.test[*].id
      }

      test_traffic_rules {
        header {
          name   = "x-test"
          values = ["blue"]
        }
      }
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, httpDesync, maxALBCount))
}

func testAccGatewayConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_gateway" "test" {
  name                    = %[1]q
  infrastructure_role_arn = aws_iam_role.test.arn
  deletion_protection     = false

  configuration {
    host_shared_alb {
      network_configuration {
        subnets = aws_subnet.test[*].id
      }
    }
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, tagKey1, tagValue1))
}

func testAccGatewayConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_gateway" "test" {
  name                    = %[1]q
  infrastructure_role_arn = aws_iam_role.test.arn
  deletion_protection     = false

  configuration {
    host_shared_alb {
      network_configuration {
        subnets = aws_subnet.test[*].id
      }
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
