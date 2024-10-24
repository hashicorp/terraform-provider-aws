// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/shield"
	"github.com/aws/aws-sdk-go-v2/service/shield/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfshield "github.com/hashicorp/terraform-provider-aws/internal/service/shield"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccProactiveEngagement_basic(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()
	address1 := acctest.RandomEmailAddress(domain)
	address2 := acctest.RandomEmailAddress(domain)
	var proactiveengagementassociation []types.EmergencyContact
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_shield_proactive_engagement.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckProactiveEngagement(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProactiveEngagementAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProactiveEngagementConfig_basic(rName, address1, address2, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProactiveEngagementAssociationExists(ctx, resourceName, &proactiveengagementassociation),
					resource.TestCheckResourceAttr(resourceName, "emergency_contact.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
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

func testAccProactiveEngagement_disabled(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()
	address1 := acctest.RandomEmailAddress(domain)
	address2 := acctest.RandomEmailAddress(domain)
	var proactiveengagementassociation []types.EmergencyContact
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_shield_proactive_engagement.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckProactiveEngagement(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProactiveEngagementAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProactiveEngagementConfig_basic(rName, address1, address2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProactiveEngagementAssociationExists(ctx, resourceName, &proactiveengagementassociation),
					resource.TestCheckResourceAttr(resourceName, "emergency_contact.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
		},
	})
}

func testAccProactiveEngagement_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()
	address1 := acctest.RandomEmailAddress(domain)
	address2 := acctest.RandomEmailAddress(domain)
	var proactiveengagementassociation []types.EmergencyContact
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_shield_proactive_engagement.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckProactiveEngagement(ctx, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProactiveEngagementAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProactiveEngagementConfig_basic(rName, address1, address2, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProactiveEngagementAssociationExists(ctx, resourceName, &proactiveengagementassociation),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfshield.ResourceProactiveEngagement, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckProactiveEngagementAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_shield_proactive_engagement" {
				continue
			}

			_, err := tfshield.FindEmergencyContactSettings(ctx, conn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Shield Proactive Engagement %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckProactiveEngagementAssociationExists(ctx context.Context, n string, v *[]types.EmergencyContact) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldClient(ctx)

		output, err := tfshield.FindEmergencyContactSettings(ctx, conn)

		if err != nil {
			return err
		}

		*v = output

		return nil
	}
}

func testAccPreCheckProactiveEngagement(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldClient(ctx)

	input := &shield.DescribeEmergencyContactSettingsInput{}
	_, err := conn.DescribeEmergencyContactSettings(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccProactiveEngagementConfig_basic(rName, email1, email2 string, enabled bool) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        "Sid" : "",
        "Effect" : "Allow",
        "Principal" : {
          "Service" : "drt.shield.amazonaws.com"
        },
        "Action" : "sts:AssumeRole"
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSShieldDRTAccessPolicy"
}

resource "aws_shield_drt_access_role_arn_association" "test" {
  role_arn = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy_attachment.test]
}

resource "aws_shield_proactive_engagement" "test" {
  enabled = %[4]t

  emergency_contact {
    contact_notes = "Notes"
    email_address = %[2]q
    phone_number  = "+12358132134"
  }
  emergency_contact {
    contact_notes = "Notes 2"
    email_address = %[3]q
    phone_number  = "+12358132134"
  }

  depends_on = [aws_shield_drt_access_role_arn_association.test]
}

`, rName, email1, email2, enabled)
}
