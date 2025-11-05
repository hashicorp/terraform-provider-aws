// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkflowmonitor_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/networkflowmonitor"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkflowmonitor/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfnetworkflowmonitor "github.com/hashicorp/terraform-provider-aws/internal/service/networkflowmonitor"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkFlowMonitorScope_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Scope": {
			acctest.CtBasic:      testAccNetworkFlowMonitorScope_basic,
			acctest.CtDisappears: testAccNetworkFlowMonitorScope_disappears,
			"tags":               testAccNetworkFlowMonitorScope_tags,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccNetworkFlowMonitorScope_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var scope networkflowmonitor.GetScopeOutput
	resourceName := "aws_networkflowmonitor_scope.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFlowMonitorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScopeConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckScopeExists(ctx, resourceName, &scope),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "networkflowmonitor", "scope/*"),
					resource.TestCheckResourceAttrSet(resourceName, "scope_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
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

func testAccNetworkFlowMonitorScope_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var scope networkflowmonitor.GetScopeOutput
	resourceName := "aws_networkflowmonitor_scope.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFlowMonitorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScopeConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScopeExists(ctx, resourceName, &scope),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfnetworkflowmonitor.ResourceScope, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccNetworkFlowMonitorScope_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var scope networkflowmonitor.GetScopeOutput
	resourceName := "aws_networkflowmonitor_scope.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFlowMonitorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckScopeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccScopeConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScopeExists(ctx, resourceName, &scope),
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
				Config: testAccScopeConfig_tags2(acctest.CtKey1, "value1updated", acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScopeExists(ctx, resourceName, &scope),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, "value1updated"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccScopeConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScopeExists(ctx, resourceName, &scope),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckScopeExists(ctx context.Context, n string, v *networkflowmonitor.GetScopeOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFlowMonitorClient(ctx)

		output, err := tfnetworkflowmonitor.FindScopeByID(ctx, conn, rs.Primary.Attributes["scope_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckScopeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFlowMonitorClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkflowmonitor_scope" {
				continue
			}

			_, err := tfnetworkflowmonitor.FindScopeByID(ctx, conn, rs.Primary.Attributes["scope_id"])

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Flow Monitor Scope %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccScopeConfig_basic() string {
	return `
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_networkflowmonitor_scope" "test" {
  targets {
    region = data.aws_region.current.name
    target_identifier {
      target_id   = data.aws_caller_identity.current.account_id
      target_type = "ACCOUNT"
    }
  }
}
`
}

func testAccScopeConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_networkflowmonitor_scope" "test" {
  targets {
    region = data.aws_region.current.name
    target_identifier {
      target_id   = data.aws_caller_identity.current.account_id
      target_type = "ACCOUNT"
    }
  }

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccScopeConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

resource "aws_networkflowmonitor_scope" "test" {
  targets {
    region = data.aws_region.current.name
    target_identifier {
      target_id   = data.aws_caller_identity.current.account_id
      target_type = "ACCOUNT"
    }
  }

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
