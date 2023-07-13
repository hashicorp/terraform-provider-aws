// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package synthetics_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/synthetics"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsynthetics "github.com/hashicorp/terraform-provider-aws/internal/service/synthetics"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccSyntheticsGroupAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_group_association.test"
	var groupSummary synthetics.GroupSummary

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, synthetics.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupAssociationExists(ctx, resourceName, &groupSummary),
					acctest.MatchResourceAttrRegionalARN(resourceName, "canary_arn", synthetics.ServiceName, regexp.MustCompile(`canary:.+`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "group_arn", synthetics.ServiceName, regexp.MustCompile(`group:.+`)),
					resource.TestCheckResourceAttr(resourceName, "group_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "group_id"),
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

func TestAccSyntheticsGroupAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_synthetics_group_association.test"
	var groupSummary synthetics.GroupSummary

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, synthetics.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGroupAssociationExists(ctx, resourceName, &groupSummary),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsynthetics.ResourceGroupAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckGroupAssociationExists(ctx context.Context, name string, v *synthetics.GroupSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Synthetics Group Association ID is set")
		}

		canaryArn, groupName, err := tfsynthetics.GroupAssociationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SyntheticsConn(ctx)
		output, err := tfsynthetics.FindAssociatedGroup(ctx, conn, canaryArn, groupName)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckGroupAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SyntheticsConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_synthetics_group_association" {
				continue
			}

			canaryArn, groupName, err := tfsynthetics.GroupAssociationParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfsynthetics.FindAssociatedGroup(ctx, conn, canaryArn, groupName)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("association to group (%s) for canary (%s) still exists", groupName, canaryArn)
		}

		return nil
	}
}

func testAccGroupAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccCanaryConfig_basic(rName), testAccGroupConfig_basic(rName), `
resource "aws_synthetics_group_association" "test" {
  group_name = aws_synthetics_group.test.name
  canary_arn = aws_synthetics_canary.test.arn
}
`)
}
