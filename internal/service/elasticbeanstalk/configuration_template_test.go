// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelasticbeanstalk "github.com/hashicorp/terraform-provider-aws/internal/service/elasticbeanstalk"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElasticBeanstalkConfigurationTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var config awstypes.ConfigurationSettingsDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_configuration_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationTemplateExists(ctx, resourceName, &config),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkConfigurationTemplate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var config awstypes.ConfigurationSettingsDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_configuration_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationTemplateExists(ctx, resourceName, &config),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelasticbeanstalk.ResourceConfigurationTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccElasticBeanstalkConfigurationTemplate_Disappears_application(t *testing.T) {
	ctx := acctest.Context(t)
	var config awstypes.ConfigurationSettingsDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_configuration_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationTemplateExists(ctx, resourceName, &config),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelasticbeanstalk.ResourceApplication(), "aws_elastic_beanstalk_application.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccElasticBeanstalkConfigurationTemplate_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	var config awstypes.ConfigurationSettingsDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_configuration_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationTemplateConfig_vpc(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationTemplateExists(ctx, resourceName, &config),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkConfigurationTemplate_settings(t *testing.T) {
	ctx := acctest.Context(t)
	var config awstypes.ConfigurationSettingsDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_configuration_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationTemplateConfig_setting(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationTemplateExists(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "setting.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "setting.*", map[string]string{
						names.AttrValue: "m1.small",
					}),
				),
			},
		},
	})
}

func testAccCheckConfigurationTemplateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticBeanstalkClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elastic_beanstalk_configuration_template" {
				continue
			}

			_, err := tfelasticbeanstalk.FindConfigurationSettingsByTwoPartKey(ctx, conn, rs.Primary.Attributes["application"], rs.Primary.ID)

			if tfresource.NotFound(err) {
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

func testAccCheckConfigurationTemplateExists(ctx context.Context, n string, v *awstypes.ConfigurationSettingsDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Elastic Beanstalk Configuration Template ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticBeanstalkClient(ctx)

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
