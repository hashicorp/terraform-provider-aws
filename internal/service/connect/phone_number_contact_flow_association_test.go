// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPhoneNumberContactFlowAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.FlowAssociationSummary
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_connect_phone_number_contact_flow_association.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPhoneNumberContactFlowAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneNumberContactFlowAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPhoneNumberContactFlowAssociationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "phone_number_id",
				ImportStateIdFunc:                    testAccPhoneNumberContactFlowAssociationImportStateIDFunc(resourceName),
			},
		},
	})
}

func testAccPhoneNumberContactFlowAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.FlowAssociationSummary
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_connect_phone_number_contact_flow_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPhoneNumberContactFlowAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneNumberContactFlowAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPhoneNumberContactFlowAssociationExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfconnect.ResourcePhoneNumberContactFlowAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPhoneNumberContactFlowAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ConnectClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_connect_phone_number_contact_flow_association" {
				continue
			}

			_, err := tfconnect.FindPhoneNumberContactFlowAssociationByThreePartKey(ctx, conn, rs.Primary.Attributes["phone_number_id"], rs.Primary.Attributes[names.AttrInstanceID], rs.Primary.Attributes["contact_flow_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Connect Phone Number %s Contact Flow %s Association still exists", rs.Primary.Attributes["phone_number_id"], rs.Primary.Attributes["contact_flow_id"])
		}

		return nil
	}
}

func testAccCheckPhoneNumberContactFlowAssociationExists(ctx context.Context, t *testing.T, n string, v *awstypes.FlowAssociationSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ConnectClient(ctx)

		output, err := tfconnect.FindPhoneNumberContactFlowAssociationByThreePartKey(ctx, conn, rs.Primary.Attributes["phone_number_id"], rs.Primary.Attributes[names.AttrInstanceID], rs.Primary.Attributes["contact_flow_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPhoneNumberContactFlowAssociationImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", n)
		}

		return fmt.Sprintf("%s,%s,%s", rs.Primary.Attributes["phone_number_id"], rs.Primary.Attributes[names.AttrInstanceID], rs.Primary.Attributes["contact_flow_id"]), nil
	}
}

func testAccPhoneNumberContactFlowAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_connect_phone_number" "test" {
  target_arn   = aws_connect_instance.test.arn
  country_code = "US"
  type         = "DID"

  tags = {
    Name = %[1]q
  }
}

resource "aws_connect_contact_flow" "test" {
  instance_id  = aws_connect_instance.test.id
  name         = %[1]q
  filename     = "test-fixtures/connect_contact_flow.json"
  content_hash = filebase64sha256("test-fixtures/connect_contact_flow.json")

  tags = {
    Name = %[1]q
  }
}

resource "aws_connect_phone_number_contact_flow_association" "test" {
  phone_number_id = aws_connect_phone_number.test.id
  instance_id     = aws_connect_instance.test.id
  contact_flow_id = aws_connect_contact_flow.test.contact_flow_id
}
`, rName)
}
