// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmacie2 "github.com/hashicorp/terraform-provider-aws/internal/service/macie2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetMacieSessionOutput
	resourceName := "aws_macie2_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", macie2.FindingPublishingFrequencyFifteenMinutes),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, macie2.MacieStatusEnabled),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrServiceRole, "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
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

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_finding(macie2.FindingPublishingFrequencyFifteenMinutes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", macie2.FindingPublishingFrequencyFifteenMinutes),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, macie2.MacieStatusEnabled),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrServiceRole, "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccAccountConfig_finding(macie2.FindingPublishingFrequencyOneHour),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", macie2.FindingPublishingFrequencyOneHour),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, macie2.MacieStatusEnabled),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrServiceRole, "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
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

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_status(macie2.MacieStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", macie2.FindingPublishingFrequencyFifteenMinutes),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, macie2.MacieStatusEnabled),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrServiceRole, "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccAccountConfig_status(macie2.MacieStatusPaused),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", macie2.FindingPublishingFrequencyFifteenMinutes),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, macie2.MacieStatusPaused),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrServiceRole, "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
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

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_findingAndStatus(macie2.FindingPublishingFrequencyFifteenMinutes, macie2.MacieStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", macie2.FindingPublishingFrequencyFifteenMinutes),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, macie2.MacieStatusEnabled),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrServiceRole, "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
				),
			},
			{
				Config: testAccAccountConfig_findingAndStatus(macie2.FindingPublishingFrequencyOneHour, macie2.MacieStatusPaused),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName, &macie2Output),
					resource.TestCheckResourceAttr(resourceName, "finding_publishing_frequency", macie2.FindingPublishingFrequencyOneHour),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, macie2.MacieStatusPaused),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrServiceRole, "iam", "role/aws-service-role/macie.amazonaws.com/AWSServiceRoleForAmazonMacie"),
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

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfmacie2.ResourceAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAccountDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Conn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_macie2_account" {
				continue
			}

			input := &macie2.GetMacieSessionInput{}
			resp, err := conn.GetMacieSessionWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, macie2.ErrCodeAccessDeniedException) {
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

func testAccCheckAccountExists(ctx context.Context, resourceName string, macie2Session *macie2.GetMacieSessionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Conn(ctx)
		input := &macie2.GetMacieSessionInput{}

		resp, err := conn.GetMacieSessionWithContext(ctx, input)

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
