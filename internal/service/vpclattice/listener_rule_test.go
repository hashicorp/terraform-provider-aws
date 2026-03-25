// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeListenerRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var listenerRule vpclattice.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, resourceName, &listenerRule),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.fixed_response.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.forward.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.forward.0.target_groups.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "action.0.forward.0.target_groups.0.target_group_identifier", "aws_vpclattice_target_group.test.0", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "action.0.forward.0.target_groups.0.weight", "100"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "vpc-lattice", "service/{service_identifier}/listener/{listener_identifier}/rule/{rule_id}"),
					resource.TestCheckResourceAttrPair(resourceName, "listener_identifier", "aws_vpclattice_listener.test", "listener_id"),
					resource.TestCheckResourceAttr(resourceName, "match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.method", "GET"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.path_match.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "20"),
					resource.TestCheckResourceAttrPair(resourceName, "service_identifier", "aws_vpclattice_service.test", names.AttrID),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccListenerRuleConfig_ARNs(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccVPCLatticeListenerRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var listenerRule vpclattice.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, resourceName, &listenerRule),
					acctest.CheckSDKResourceDisappears(ctx, t, tfvpclattice.ResourceListenerRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCLatticeListenerRule_ARNs(t *testing.T) {
	ctx := acctest.Context(t)
	var listenerRule vpclattice.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleConfig_ARNs(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, resourceName, &listenerRule),
					resource.TestCheckResourceAttrPair(resourceName, "listener_identifier", "aws_vpclattice_listener.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "service_identifier", "aws_vpclattice_service.test", names.AttrARN),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateCheck: acctest.ComposeAggregateImportStateCheckFunc(
					acctest.ImportMatchResourceAttr("listener_identifier", regexache.MustCompile("^listener-[[:xdigit:]]+$")),
					acctest.ImportMatchResourceAttr("service_identifier", regexache.MustCompile("^svc-[[:xdigit:]]+$")),
				),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"listener_identifier",
					"service_identifier",
				},
			},
			{
				Config: testAccListenerRuleConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccVPCLatticeListenerRule_action_fixedResponse(t *testing.T) {
	ctx := acctest.Context(t)
	var listenerRule vpclattice.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleConfig_action_fixedResponse(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, resourceName, &listenerRule),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.fixed_response.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.fixed_response.0.status_code", "404"),
					resource.TestCheckResourceAttr(resourceName, "action.0.forward.#", "0"),
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

func TestAccVPCLatticeListenerRule_action_forward_Multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var listenerRule vpclattice.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleConfig_action_forward_Multiple(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, resourceName, &listenerRule),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.fixed_response.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "action.0.forward.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.forward.0.target_groups.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "action.0.forward.0.target_groups.0.target_group_identifier", "aws_vpclattice_target_group.test.0", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "action.0.forward.0.target_groups.0.weight", "25"),
					resource.TestCheckResourceAttrPair(resourceName, "action.0.forward.0.target_groups.1.target_group_identifier", "aws_vpclattice_target_group.test.1", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "action.0.forward.0.target_groups.1.weight", "75"),
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

func TestAccVPCLatticeListenerRule_match_HeaderMatches_Single(t *testing.T) {
	ctx := acctest.Context(t)
	var listenerRule vpclattice.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleConfig_match_HeaderMatches_Single(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, resourceName, &listenerRule),
					resource.TestCheckResourceAttr(resourceName, "match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.0.case_sensitive", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.0.name", "example-header"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.0.match.0.contains", "example-contains"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.0.match.0.exact", ""),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.0.match.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.path_match.#", "0"),
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

func TestAccVPCLatticeListenerRule_match_HeaderMatches_Multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var listenerRule vpclattice.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleConfig_match_HeaderMatches_Multiple(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, resourceName, &listenerRule),
					resource.TestCheckResourceAttr(resourceName, "match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.0.case_sensitive", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.0.name", "example-header"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.0.match.0.contains", "example-contains"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.0.match.0.exact", ""),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.0.match.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.1.case_sensitive", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.1.name", "other-header"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.1.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.1.match.0.contains", ""),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.1.match.0.exact", "example-exact"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.1.match.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.path_match.#", "0"),
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

func TestAccVPCLatticeListenerRule_match_PathMatch(t *testing.T) {
	ctx := acctest.Context(t)
	var listenerRule vpclattice.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleConfig_match_PathMatch(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, resourceName, &listenerRule),
					resource.TestCheckResourceAttr(resourceName, "match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.header_matches.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.method", ""),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.path_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.path_match.0.case_sensitive", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.path_match.0.match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.path_match.0.match.0.exact", ""),
					resource.TestCheckResourceAttr(resourceName, "match.0.http_match.0.path_match.0.match.0.prefix", "/example-path"),
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

func TestAccVPCLatticeListenerRule_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var listenerRule vpclattice.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, resourceName, &listenerRule),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccListenerRuleConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, t, resourceName, &listenerRule),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckListenerRuleDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).VPCLatticeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_listener_rule" {
				continue
			}

			_, err := tfvpclattice.FindListenerRuleByThreePartKey(ctx, conn, rs.Primary.Attributes["service_identifier"], rs.Primary.Attributes["listener_identifier"], rs.Primary.Attributes["rule_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPC Lattice Listener Rule %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckListenerRuleExists(ctx context.Context, t *testing.T, n string, v *vpclattice.GetRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).VPCLatticeClient(ctx)

		output, err := tfvpclattice.FindListenerRuleByThreePartKey(ctx, conn, rs.Primary.Attributes["service_identifier"], rs.Primary.Attributes["listener_identifier"], rs.Primary.Attributes["rule_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccListenerRuleConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 0), fmt.Sprintf(`
resource "aws_vpclattice_service" "test" {
  name = %[1]q
}

resource "aws_vpclattice_target_group" "test" {
  count = 2

  name = "%[1]s-${count.index}"
  type = "INSTANCE"

  config {
    port           = 80
    protocol       = "HTTP"
    vpc_identifier = aws_vpc.test.id
  }
}

resource "aws_vpclattice_listener" "test" {
  name               = %[1]q
  protocol           = "HTTP"
  service_identifier = aws_vpclattice_service.test.id
  default_action {
    fixed_response {
      status_code = 404
    }
  }
}
`, rName))
}

func testAccListenerRuleConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccListenerRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener_rule" "test" {
  name = %[1]q

  listener_identifier = aws_vpclattice_listener.test.listener_id
  service_identifier  = aws_vpclattice_service.test.id

  priority = 20

  match {
    http_match {
      method = "GET"
    }
  }

  action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test[0].id
      }
    }
  }
}
`, rName))
}

func testAccListenerRuleConfig_ARNs(rName string) string {
	return acctest.ConfigCompose(testAccListenerRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener_rule" "test" {
  name = %[1]q

  listener_identifier = aws_vpclattice_listener.test.arn
  service_identifier  = aws_vpclattice_service.test.arn

  priority = 20

  match {
    http_match {
      method = "GET"
    }
  }

  action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test[0].id
      }
    }
  }
}
`, rName))
}

func testAccListenerRuleConfig_action_fixedResponse(rName string) string {
	return acctest.ConfigCompose(testAccListenerRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener_rule" "test" {
  name = %[1]q

  listener_identifier = aws_vpclattice_listener.test.listener_id
  service_identifier  = aws_vpclattice_service.test.id

  priority = 10

  match {
    http_match {
      path_match {
        case_sensitive = false
        match {
          exact = "/example-path"
        }
      }
    }
  }

  action {
    fixed_response {
      status_code = 404
    }
  }
}
`, rName))
}

func testAccListenerRuleConfig_action_forward_Multiple(rName string) string {
	return acctest.ConfigCompose(testAccListenerRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener_rule" "test" {
  name = %[1]q

  listener_identifier = aws_vpclattice_listener.test.listener_id
  service_identifier  = aws_vpclattice_service.test.id

  priority = 20

  match {
    http_match {
      method = "GET"
    }
  }

  action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test[0].id
        weight                  = 25
      }
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test[1].id
        weight                  = 75
      }
    }
  }
}
`, rName))
}

func testAccListenerRuleConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccListenerRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener_rule" "test" {
  name                = %[1]q
  listener_identifier = aws_vpclattice_listener.test.listener_id
  service_identifier  = aws_vpclattice_service.test.id
  priority            = 30
  match {
    http_match {
      path_match {
        case_sensitive = false
        match {
          prefix = "/example-path"
        }
      }
    }
  }
  action {
    fixed_response {
      status_code = 404
    }
  }
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccListenerRuleConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccListenerRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener_rule" "test" {
  name                = %[1]q
  listener_identifier = aws_vpclattice_listener.test.listener_id
  service_identifier  = aws_vpclattice_service.test.id
  priority            = 30
  match {
    http_match {
      path_match {
        case_sensitive = false
        match {
          prefix = "/example-path"
        }
      }
    }
  }
  action {
    fixed_response {
      status_code = 404
    }
  }
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccListenerRuleConfig_match_HeaderMatches_Single(rName string) string {
	return acctest.ConfigCompose(testAccListenerRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener_rule" "test" {
  name = %[1]q

  listener_identifier = aws_vpclattice_listener.test.listener_id
  service_identifier  = aws_vpclattice_service.test.id

  priority = 40

  match {
    http_match {
      header_matches {
        name = "example-header"

        match {
          contains = "example-contains"
        }
      }
    }
  }

  action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test[0].id
      }
    }
  }
}
`, rName))
}

func testAccListenerRuleConfig_match_HeaderMatches_Multiple(rName string) string {
	return acctest.ConfigCompose(testAccListenerRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener_rule" "test" {
  name = %[1]q

  listener_identifier = aws_vpclattice_listener.test.listener_id
  service_identifier  = aws_vpclattice_service.test.id

  priority = 40

  match {
    http_match {
      header_matches {
        name = "example-header"

        match {
          contains = "example-contains"
        }
      }
      header_matches {
        name           = "other-header"
        case_sensitive = true

        match {
          exact = "example-exact"
        }
      }
    }
  }

  action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test[0].id
      }
    }
  }
}
`, rName))
}

func testAccListenerRuleConfig_match_PathMatch(rName string) string {
	return acctest.ConfigCompose(testAccListenerRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener_rule" "test" {
  name = %[1]q

  listener_identifier = aws_vpclattice_listener.test.listener_id
  service_identifier  = aws_vpclattice_service.test.id

  priority = 40

  match {
    http_match {
      path_match {
        case_sensitive = true
        match {
          prefix = "/example-path"
        }
      }
    }
  }

  action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test[0].id
      }
    }
  }
}
`, rName))
}
