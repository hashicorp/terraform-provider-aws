// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package macie2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/macie2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/macie2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfmacie2 "github.com/hashicorp/terraform-provider-aws/internal/service/macie2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetMacieSessionOutput
	resourceName := "aws_macie2_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", string(awstypes.FindingPublishingFrequencyFifteenMinutes)),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.MacieStatusEnabled)),
					acctest.CheckResourceAttrGlobalARN(ctx, resourceName, names.AttrServiceRole, "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
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

func testAccAccount_FindingPublishingFrequency(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetMacieSessionOutput
	resourceName := "aws_macie2_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_finding(string(awstypes.FindingPublishingFrequencyFifteenMinutes)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", string(awstypes.FindingPublishingFrequencyFifteenMinutes)),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.MacieStatusEnabled)),
					acctest.CheckResourceAttrGlobalARN(ctx, resourceName, names.AttrServiceRole, "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccAccountConfig_finding(string(awstypes.FindingPublishingFrequencyOneHour)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", string(awstypes.FindingPublishingFrequencyOneHour)),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.MacieStatusEnabled)),
					acctest.CheckResourceAttrGlobalARN(ctx, resourceName, names.AttrServiceRole, "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
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

func testAccAccount_WithStatus(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetMacieSessionOutput
	resourceName := "aws_macie2_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_status(string(awstypes.MacieStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", string(awstypes.FindingPublishingFrequencyFifteenMinutes)),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.MacieStatusEnabled)),
					acctest.CheckResourceAttrGlobalARN(ctx, resourceName, names.AttrServiceRole, "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccAccountConfig_status(string(awstypes.MacieStatusPaused)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", string(awstypes.FindingPublishingFrequencyFifteenMinutes)),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.MacieStatusPaused)),
					acctest.CheckResourceAttrGlobalARN(ctx, resourceName, names.AttrServiceRole, "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
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

func testAccAccount_WithFindingAndStatus(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetMacieSessionOutput
	resourceName := "aws_macie2_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_findingAndStatus(string(awstypes.FindingPublishingFrequencyFifteenMinutes), string(awstypes.MacieStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", string(awstypes.FindingPublishingFrequencyFifteenMinutes)),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.MacieStatusEnabled)),
					acctest.CheckResourceAttrGlobalARN(ctx, resourceName, names.AttrServiceRole, "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccAccountConfig_findingAndStatus(string(awstypes.FindingPublishingFrequencyOneHour), string(awstypes.MacieStatusPaused)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", string(awstypes.FindingPublishingFrequencyOneHour)),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.MacieStatusPaused)),
					acctest.CheckResourceAttrGlobalARN(ctx, resourceName, names.AttrServiceRole, "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
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

func testAccAccount_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetMacieSessionOutput
	resourceName := "aws_macie2_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName, &macie2Output),
					acctest.CheckSDKResourceDisappears(ctx, t, tfmacie2.ResourceAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAccountDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).Macie2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_macie2_account" {
				continue
			}

			input := &macie2.GetMacieSessionInput{}
			resp, err := conn.GetMacieSession(ctx, input)

			if errs.IsA[*awstypes.AccessDeniedException](err) {
				continue
			}

			if err != nil {
				return err
			}

			if resp != nil {
				return fmt.Errorf("macie account %q still enabled", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckAccountExists(ctx context.Context, t *testing.T, resourceName string, macie2Session *macie2.GetMacieSessionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.ProviderMeta(ctx, t).Macie2Client(ctx)
		input := &macie2.GetMacieSessionInput{}

		resp, err := conn.GetMacieSession(ctx, input)

		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("macie account %q does not exist", rs.Primary.ID)
		}

		*macie2Session = *resp

		return nil
	}
}

func testAccAccountConfig_basic() string {
	return `
resource "aws_macie2_account" "test" {}
`
}

func testAccAccountConfig_finding(finding string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {
  finding_publishing_frequency = "%s"
}
`, finding)
}

func testAccAccountConfig_status(status string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {
  status = "%s"
}
`, status)
}

func testAccAccountConfig_findingAndStatus(finding, status string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {
  finding_publishing_frequency = "%s"
  status                       = "%s"
}
`, finding, status)
}
