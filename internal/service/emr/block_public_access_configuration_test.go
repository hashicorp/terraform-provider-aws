// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package emr_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/emr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfemr "github.com/hashicorp/terraform-provider-aws/internal/service/emr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEMRBlockPublicAccessConfiguration_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:      testAccBlockPublicAccessConfiguration_basic,
		acctest.CtDisappears: testAccBlockPublicAccessConfiguration_disappears,
		"default":            testAccBlockPublicAccessConfiguration_default,
		"enabledMultiRange":  testAccBlockPublicAccessConfiguration_enabledMultiRange,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccBlockPublicAccessConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_emr_block_public_access_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EMREndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBlockPublicAccessConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBlockPublicAccessConfigurationConfig_basic(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockPublicAccessConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "block_public_security_group_rules", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "permitted_public_security_group_rule_range.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBlockPublicAccessConfigurationConfig_basic(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockPublicAccessConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "block_public_security_group_rules", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "permitted_public_security_group_rule_range.#", "0"),
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

func testAccBlockPublicAccessConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_emr_block_public_access_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EMREndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBlockPublicAccessConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBlockPublicAccessConfigurationConfig_basic(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockPublicAccessConfigurationExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfemr.ResourceBlockPublicAccessConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccBlockPublicAccessConfiguration_default(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_emr_block_public_access_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EMREndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBlockPublicAccessConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: blockPublicAccessConfigurationConfig_defaultString,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockPublicAccessConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "block_public_security_group_rules", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "permitted_public_security_group_rule_range.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permitted_public_security_group_rule_range.0.min_range", "22"),
					resource.TestCheckResourceAttr(resourceName, "permitted_public_security_group_rule_range.0.max_range", "22"),
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

func testAccBlockPublicAccessConfiguration_enabledMultiRange(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_emr_block_public_access_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EMREndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBlockPublicAccessConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: blockPublicAccessConfigurationConfig_enabledMultiRangeString,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBlockPublicAccessConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "block_public_security_group_rules", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "permitted_public_security_group_rule_range.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "permitted_public_security_group_rule_range.0.min_range", "22"),
					resource.TestCheckResourceAttr(resourceName, "permitted_public_security_group_rule_range.0.max_range", "22"),
					resource.TestCheckResourceAttr(resourceName, "permitted_public_security_group_rule_range.1.min_range", "100"),
					resource.TestCheckResourceAttr(resourceName, "permitted_public_security_group_rule_range.1.max_range", "101"),
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

func testAccCheckBlockPublicAccessConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EMRClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_emr_block_public_access_configuration" {
				continue
			}

			output, err := tfemr.FindBlockPublicAccessConfiguration(ctx, conn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			// See defaultBlockPublicAccessConfiguration().
			if aws.ToBool(output.BlockPublicSecurityGroupRules) &&
				len(output.PermittedPublicSecurityGroupRuleRanges) == 1 &&
				aws.ToInt32(output.PermittedPublicSecurityGroupRuleRanges[0].MinRange) == 22 &&
				aws.ToInt32(output.PermittedPublicSecurityGroupRuleRanges[0].MaxRange) == 22 {
				return nil
			}

			return fmt.Errorf("EMR Block Public Access Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBlockPublicAccessConfigurationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EMRClient(ctx)

		_, err := tfemr.FindBlockPublicAccessConfiguration(ctx, conn)

		return err
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).EMRClient(ctx)

	input := &emr.GetBlockPublicAccessConfigurationInput{}
	_, err := conn.GetBlockPublicAccessConfiguration(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccBlockPublicAccessConfigurationConfig_basic(enabled bool) string {
	return fmt.Sprintf(`
resource "aws_emr_block_public_access_configuration" "test" {
  block_public_security_group_rules = %[1]t
}
`, enabled)
}

const blockPublicAccessConfigurationConfig_defaultString = `
resource "aws_emr_block_public_access_configuration" "test" {
  block_public_security_group_rules = true

  permitted_public_security_group_rule_range {
    min_range = 22
    max_range = 22
  }
}
`

const blockPublicAccessConfigurationConfig_enabledMultiRangeString = `
resource "aws_emr_block_public_access_configuration" "test" {
  block_public_security_group_rules = true

  permitted_public_security_group_rule_range {
    min_range = 22
    max_range = 22
  }

  permitted_public_security_group_rule_range {
    min_range = 100
    max_range = 101
  }
}
`
