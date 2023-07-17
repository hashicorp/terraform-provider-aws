// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappconfig "github.com/hashicorp/terraform-provider-aws/internal/service/appconfig"
)

func TestAccAppConfigHostedConfigurationVersion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_hosted_configuration_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHostedConfigurationVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHostedConfigurationVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedConfigurationVersionExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appconfig", regexp.MustCompile(`application/[a-z0-9]{4,7}/configurationprofile/[a-z0-9]{4,7}/hostedconfigurationversion/[0-9]+`)),
					resource.TestCheckResourceAttrPair(resourceName, "application_id", "aws_appconfig_application.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_profile_id", "aws_appconfig_configuration_profile.test", "configuration_profile_id"),
					resource.TestCheckResourceAttr(resourceName, "content", "{\"foo\":\"bar\"}"),
					resource.TestCheckResourceAttr(resourceName, "content_type", "application/json"),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "version_number", "1"),
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

func TestAccAppConfigHostedConfigurationVersion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_hosted_configuration_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHostedConfigurationVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHostedConfigurationVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedConfigurationVersionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappconfig.ResourceHostedConfigurationVersion(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckHostedConfigurationVersionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppConfigConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appconfig_hosted_configuration_version" {
				continue
			}

			appID, confProfID, versionNumber, err := tfappconfig.HostedConfigurationVersionParseID(rs.Primary.ID)

			if err != nil {
				return err
			}

			input := &appconfig.GetHostedConfigurationVersionInput{
				ApplicationId:          aws.String(appID),
				ConfigurationProfileId: aws.String(confProfID),
				VersionNumber:          aws.Int64(int64(versionNumber)),
			}

			output, err := conn.GetHostedConfigurationVersionWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, appconfig.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error reading AppConfig Hosted Configuration Version (%s): %w", rs.Primary.ID, err)
			}

			if output != nil {
				return fmt.Errorf("AppConfig Hosted Configuration Version (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckHostedConfigurationVersionExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		appID, confProfID, versionNumber, err := tfappconfig.HostedConfigurationVersionParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppConfigConn(ctx)

		output, err := conn.GetHostedConfigurationVersionWithContext(ctx, &appconfig.GetHostedConfigurationVersionInput{
			ApplicationId:          aws.String(appID),
			ConfigurationProfileId: aws.String(confProfID),
			VersionNumber:          aws.Int64(int64(versionNumber)),
		})

		if err != nil {
			return fmt.Errorf("error reading AppConfig Hosted Configuration Version (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("AppConfig Hosted Configuration Version (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccHostedConfigurationVersionConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccConfigurationProfileConfig_name(rName),
		fmt.Sprintf(`
resource "aws_appconfig_hosted_configuration_version" "test" {
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.configuration_profile_id
  content_type             = "application/json"

  content = jsonencode({
    foo = "bar"
  })

  description = %q
}
`, rName))
}
