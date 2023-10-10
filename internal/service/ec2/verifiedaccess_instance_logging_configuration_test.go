// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVerifiedAccessInstanceLoggingConfiguration_accessLogsIncludeTrustContext(t *testing.T) {
	ctx := acctest.Context(t)

	var v types.VerifiedAccessInstanceLoggingConfiguration
	resourceName := "aws_verifiedaccess_instance_logging_configuration.test"
	instanceResourceName := "aws_verifiedaccess_instance.test"

	include_trust_context_original := true
	include_trust_context_updated := false

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccessInstanceLoggingConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessInstanceLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_basic_accessLogsIncludeTrustContext(include_trust_context_original),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.include_trust_context", strconv.FormatBool(include_trust_context_original)),
					resource.TestCheckResourceAttrPair(resourceName, "verifiedaccess_instance_id", instanceResourceName, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccLoggingConfigurationConfig_basic_accessLogsIncludeTrustContext(include_trust_context_updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.include_trust_context", strconv.FormatBool(include_trust_context_updated)),
					resource.TestCheckResourceAttrPair(resourceName, "verifiedaccess_instance_id", instanceResourceName, "id"),
				),
			},
		},
	})
}

func TestAccVerifiedAccessInstanceLoggingConfiguration_accessLogsLogVersion(t *testing.T) {
	ctx := acctest.Context(t)

	var v types.VerifiedAccessInstanceLoggingConfiguration
	resourceName := "aws_verifiedaccess_instance_logging_configuration.test"
	instanceResourceName := "aws_verifiedaccess_instance.test"

	log_version_original := "ocsf-0.1"
	log_version_updated := "ocsf-1.0.0-rc.2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckVerifiedAccessInstanceLoggingConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVerifiedAccessInstanceLoggingConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfigurationConfig_basic_accessLogsLogVersion(log_version_original),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.log_version", log_version_original),
					resource.TestCheckResourceAttrPair(resourceName, "verifiedaccess_instance_id", instanceResourceName, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
			{
				Config: testAccLoggingConfigurationConfig_basic_accessLogsLogVersion(log_version_updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVerifiedAccessInstanceLoggingConfigurationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "access_logs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "access_logs.0.log_version", log_version_updated),
					resource.TestCheckResourceAttrPair(resourceName, "verifiedaccess_instance_id", instanceResourceName, "id"),
				),
			},
		},
	})
}

func testAccCheckVerifiedAccessInstanceLoggingConfigurationExists(ctx context.Context, n string, v *types.VerifiedAccessInstanceLoggingConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindVerifiedAccessInstanceLoggingConfigurationByInstanceID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckVerifiedAccessInstanceLoggingConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_verifiedaccess_instance_logging_configuration" {
				continue
			}

			_, err := tfec2.FindVerifiedAccessInstanceLoggingConfigurationByInstanceID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Verified Access Instance Logging Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPreCheckVerifiedAccessInstanceLoggingConfiguration(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeVerifiedAccessInstanceLoggingConfigurationsInput{}
	_, err := conn.DescribeVerifiedAccessInstanceLoggingConfigurations(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccVerifiedAccessInstanceLoggingConfigurationConfig_instance() string {
	return `
resource "aws_verifiedaccess_instance" "test" {}
`
}

func testAccLoggingConfigurationConfig_basic_accessLogsIncludeTrustContext(includeTrustContext bool) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_instance(),
		fmt.Sprintf(`
resource "aws_verifiedaccess_instance_logging_configuration" "test" {
  access_logs {
    include_trust_context = %[1]t
  }

  verifiedaccess_instance_id = aws_verifiedaccess_instance.test.id
}
`, includeTrustContext))
}

func testAccLoggingConfigurationConfig_basic_accessLogsLogVersion(logVersion string) string {
	return acctest.ConfigCompose(
		testAccVerifiedAccessInstanceLoggingConfigurationConfig_instance(),
		fmt.Sprintf(`
resource "aws_verifiedaccess_instance_logging_configuration" "test" {
  access_logs {
    log_version = %[1]q
  }

  verifiedaccess_instance_id = aws_verifiedaccess_instance.test.id
}
`, logVersion))
}
