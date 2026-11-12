// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsecr "github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/signer"
	signertypes "github.com/aws/aws-sdk-go-v2/service/signer/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfecr "github.com/hashicorp/terraform-provider-aws/internal/service/ecr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRSigningConfiguration_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccSigningConfiguration_basic,
		acctest.CtDisappears: testAccSigningConfiguration_disappears,
		"update":             testAccSigningConfiguration_update,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccSigningConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awsecr.GetSigningConfigurationOutput
	resourceName := "aws_ecr_signing_configuration.test"
	signingProfileName := "aws_signer_signing_profile.test"
	rName := fmt.Sprintf("tf-acc-test-%d", acctest.RandInt(t))

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSignerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningConfigurationExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.signing_profile_arn", signingProfileName, names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.repository_filter.*", map[string]string{
						names.AttrFilter: "team-a/*",
						"filter_type":    "WILDCARD_MATCH",
					}),
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

func testAccSigningConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awsecr.GetSigningConfigurationOutput
	resourceName := "aws_ecr_signing_configuration.test"
	rName := fmt.Sprintf("tf-acc-test-%d", acctest.RandInt(t))

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSignerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningConfigurationExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfecr.ResourceSigningConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSigningConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awsecr.GetSigningConfigurationOutput
	resourceName := "aws_ecr_signing_configuration.test"
	rName := fmt.Sprintf("tf-acc-test-%d", acctest.RandInt(t))

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSignerSigningProfile(ctx, t, "AWSLambda-SHA384-ECDSA")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.repository_filter.*", map[string]string{
						names.AttrFilter: "team-a/*",
						"filter_type":    "WILDCARD_MATCH",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSigningConfigurationConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSigningConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.signing_profile_arn", "aws_signer_signing_profile.test_a", names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.repository_filter.*", map[string]string{
						names.AttrFilter: "team-a/*",
						"filter_type":    "WILDCARD_MATCH",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.signing_profile_arn", "aws_signer_signing_profile.test_b", names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.repository_filter.*", map[string]string{
						names.AttrFilter: "team-b/prod",
						"filter_type":    "WILDCARD_MATCH",
					}),
				),
			},
		},
	})
}

func testAccPreCheckSignerSigningProfile(ctx context.Context, t *testing.T, platformID string) {
	conn := acctest.ProviderMeta(ctx, t).SignerClient(ctx)

	pages := signer.NewListSigningPlatformsPaginator(conn, &signer.ListSigningPlatformsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			t.Fatalf("unexpected PreCheck error: %s", err)
		}

		if page == nil {
			t.Skip("skipping acceptance testing: empty response")
		}

		for _, platform := range page.Platforms {
			if platform == (signertypes.SigningPlatform{}) {
				continue
			}

			if aws.ToString(platform.PlatformId) == platformID {
				return
			}
		}
	}

	t.Skipf("skipping acceptance testing: Signing Platform (%s) not found", platformID)
}

func testAccCheckSigningConfigurationExists(ctx context.Context, t *testing.T, n string, v *awsecr.GetSigningConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ECRClient(ctx)
		output, err := tfecr.FindSigningConfiguration(ctx, conn)
		if err != nil {
			return err
		}

		*v = *output
		return nil
	}
}

func testAccCheckSigningConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ECRClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecr_signing_configuration" {
				continue
			}

			_, err := tfecr.FindSigningConfiguration(ctx, conn)
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ECR Signing Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccSigningConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name        = %[1]q
}

resource "aws_ecr_signing_configuration" "test" {
  rule {
    signing_profile_arn = aws_signer_signing_profile.test.arn

    repository_filter {
      filter      = "team-a/*"
      filter_type = "WILDCARD_MATCH"
    }
  }
}
`, rName)
}

func testAccSigningConfigurationConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test_a" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name        = "%[1]s-a"
}

resource "aws_signer_signing_profile" "test_b" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name        = "%[1]s-b"
}

resource "aws_ecr_signing_configuration" "test" {
  rule {
    signing_profile_arn = aws_signer_signing_profile.test_a.arn

    repository_filter {
      filter      = "team-a/*"
      filter_type = "WILDCARD_MATCH"
    }
  }

  rule {
    signing_profile_arn = aws_signer_signing_profile.test_b.arn

    repository_filter {
      filter      = "team-b/prod"
      filter_type = "WILDCARD_MATCH"
    }
  }
}
`, rName)
}
