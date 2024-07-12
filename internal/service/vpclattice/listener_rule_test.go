// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeListenerRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var listenerRule vpclattice.GetRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccChecklistenerRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, resourceName, &listenerRule),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "20"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile(`service/svc-.*/listener/listener-.*/rule/rule.+`)),
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

func TestAccVPCLatticeListenerRule_fixedResponse(t *testing.T) {
	ctx := acctest.Context(t)
	var listenerRule vpclattice.GetRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccChecklistenerRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleConfig_fixedResponse(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, resourceName, &listenerRule),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "action.0.fixed_response.0.status_code", "404"),
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

func TestAccVPCLatticeListenerRule_methodMatch(t *testing.T) {
	ctx := acctest.Context(t)
	var listenerRule vpclattice.GetRuleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccChecklistenerRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleConfig_methodMatch(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, resourceName, &listenerRule),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "40"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_listener_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccChecklistenerRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerRuleConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, resourceName, &listenerRule),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerRuleExists(ctx, resourceName, &listenerRule),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckListenerRuleExists(ctx context.Context, name string, rule *vpclattice.GetRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameListenerRule, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameListenerRule, name, errors.New("not set"))
		}

		serviceIdentifier := rs.Primary.Attributes["service_identifier"]
		listenerIdentifier := rs.Primary.Attributes["listener_identifier"]

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)
		resp, err := conn.GetRule(ctx, &vpclattice.GetRuleInput{
			RuleIdentifier:     aws.String(rs.Primary.Attributes[names.AttrARN]),
			ListenerIdentifier: aws.String(listenerIdentifier),
			ServiceIdentifier:  aws.String(serviceIdentifier),
		})

		if err != nil {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameListenerRule, rs.Primary.ID, err)
		}

		*rule = *resp

		return nil
	}
}

func testAccChecklistenerRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_listener_rule" {
				continue
			}

			listenerIdentifier := rs.Primary.Attributes["listener_identifier"]
			serviceIdentifier := rs.Primary.Attributes["service_identifier"]

			_, err := conn.GetRule(ctx, &vpclattice.GetRuleInput{
				RuleIdentifier:     aws.String(rs.Primary.Attributes[names.AttrARN]),
				ListenerIdentifier: aws.String(listenerIdentifier),
				ServiceIdentifier:  aws.String(serviceIdentifier),
			})
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.VPCLattice, create.ErrActionCheckingDestroyed, tfvpclattice.ResNameListenerRule, rs.Primary.ID, errors.New("not destroyed"))
		}

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
  name                = %[1]q
  listener_identifier = aws_vpclattice_listener.test.listener_id
  service_identifier  = aws_vpclattice_service.test.id
  priority            = 20
  match {
    http_match {

      header_matches {
        name           = "example-header"
        case_sensitive = false

        match {
          exact = "example-contains"
        }
      }

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
        weight                  = 1
      }
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test[1].id
        weight                  = 2
      }
    }
  }
}
`, rName))
}

func testAccListenerRuleConfig_fixedResponse(rName string) string {
	return acctest.ConfigCompose(testAccListenerRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener_rule" "test" {
  name                = %[1]q
  listener_identifier = aws_vpclattice_listener.test.listener_id
  service_identifier  = aws_vpclattice_service.test.id
  priority            = 10
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

func testAccListenerRuleConfig_methodMatch(rName string) string {
	return acctest.ConfigCompose(testAccListenerRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpclattice_listener_rule" "test" {
  name                = %[1]q
  listener_identifier = aws_vpclattice_listener.test.listener_id
  service_identifier  = aws_vpclattice_service.test.id
  priority            = 40
  match {
    http_match {

      method = "POST"

      header_matches {
        name           = "example-header"
        case_sensitive = false

        match {
          contains = "example-contains"
        }
      }

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
        weight                  = 1
      }
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.test[1].id
        weight                  = 2
      }
    }
  }
}
`, rName))
}
