// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfelasticbeanstalk "github.com/hashicorp/terraform-provider-aws/internal/service/elasticbeanstalk"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElasticBeanstalkConfigurationTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var config awstypes.ConfigurationSettingsDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_configuration_template.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationTemplateExists(ctx, t, resourceName, &config),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkConfigurationTemplate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var config awstypes.ConfigurationSettingsDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_configuration_template.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationTemplateExists(ctx, t, resourceName, &config),
					acctest.CheckSDKResourceDisappears(ctx, t, tfelasticbeanstalk.ResourceConfigurationTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccElasticBeanstalkConfigurationTemplate_Disappears_application(t *testing.T) {
	ctx := acctest.Context(t)
	var config awstypes.ConfigurationSettingsDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_configuration_template.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationTemplateExists(ctx, t, resourceName, &config),
					acctest.CheckSDKResourceDisappears(ctx, t, tfelasticbeanstalk.ResourceApplication(), "aws_elastic_beanstalk_application.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccElasticBeanstalkConfigurationTemplate_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	var config awstypes.ConfigurationSettingsDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_configuration_template.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationTemplateConfig_vpc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationTemplateExists(ctx, t, resourceName, &config),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkConfigurationTemplate_settings(t *testing.T) {
	ctx := acctest.Context(t)
	var config awstypes.ConfigurationSettingsDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_configuration_template.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationTemplateConfig_setting(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationTemplateExists(ctx, t, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "setting.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "setting.*", map[string]string{
						names.AttrValue: "m1.small",
					}),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkConfigurationTemplate_migrate_settingsResourceDefault(t *testing.T) {
	ctx := acctest.Context(t)
	var config awstypes.ConfigurationSettingsDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_configuration_template.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		CheckDestroy: testAccCheckConfigurationTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.14.1",
					},
				},
				Config: testAccConfigurationTemplateConfig_setting(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationTemplateExists(ctx, t, resourceName, &config),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccConfigurationTemplateConfig_setting(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationTemplateExists(ctx, t, resourceName, &config),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func testAccCheckConfigurationTemplateDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ElasticBeanstalkClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elastic_beanstalk_configuration_template" {
				continue
			}

			_, err := tfelasticbeanstalk.FindConfigurationSettingsByTwoPartKey(ctx, conn, rs.Primary.Attributes["application"], rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Elastic Beanstalk Configuration Template %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckConfigurationTemplateExists(ctx context.Context, t *testing.T, n string, v *awstypes.ConfigurationSettingsDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ElasticBeanstalkClient(ctx)

		output, err := tfelasticbeanstalk.FindConfigurationSettingsByTwoPartKey(ctx, conn, rs.Primary.Attributes["application"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

const testAccConfigurationTemplateConfig_base = `
data "aws_elastic_beanstalk_solution_stack" "test" {
  most_recent = true
  name_regex  = "64bit Amazon Linux .* running Python .*"
}
`

func testAccConfigurationTemplateConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccConfigurationTemplateConfig_base,
		fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "test" {
  name        = %[1]q
  description = "testing"
}

resource "aws_elastic_beanstalk_configuration_template" "test" {
  name                = %[1]q
  application         = aws_elastic_beanstalk_application.test.name
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name
}
`, rName))
}

func testAccConfigurationTemplateConfig_vpc(rName string) string {
	return acctest.ConfigCompose(
		testAccConfigurationTemplateConfig_base,
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "test" {
  name        = %[1]q
  description = "testing"
}

resource "aws_elastic_beanstalk_configuration_template" "test" {
  name        = %[1]q
  application = aws_elastic_beanstalk_application.test.name

  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name

  setting {
    namespace = "aws:ec2:vpc"
    name      = "VPCId"
    value     = aws_vpc.test.id
  }

  setting {
    namespace = "aws:ec2:vpc"
    name      = "Subnets"
    value     = aws_subnet.test[0].id
  }
}
`, rName))
}

func testAccConfigurationTemplateConfig_setting(rName string) string {
	return acctest.ConfigCompose(
		testAccConfigurationTemplateConfig_base,
		fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "test" {
  name        = %[1]q
  description = "testing"
}

resource "aws_elastic_beanstalk_configuration_template" "test" {
  name        = %[1]q
  application = aws_elastic_beanstalk_application.test.name

  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "InstanceType"
    value     = "m1.small"
  }
}
`, rName))
}
