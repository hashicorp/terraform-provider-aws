package elasticbeanstalk_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelasticbeanstalk "github.com/hashicorp/terraform-provider-aws/internal/service/elasticbeanstalk"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccElasticBeanstalkConfigurationTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var config elasticbeanstalk.ConfigurationSettingsDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_configuration_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticbeanstalk.EndpointsID),
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
	var config elasticbeanstalk.ConfigurationSettingsDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_configuration_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticbeanstalk.EndpointsID),
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

func TestAccElasticBeanstalkConfigurationTemplate_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	var config elasticbeanstalk.ConfigurationSettingsDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_configuration_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticbeanstalk.EndpointsID),
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
	var config elasticbeanstalk.ConfigurationSettingsDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_configuration_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticbeanstalk.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationTemplateConfig_setting(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationTemplateExists(ctx, resourceName, &config),
					resource.TestCheckResourceAttr(resourceName, "setting.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "setting.*", map[string]string{
						"value": "m1.small",
					}),
				),
			},
		},
	})
}

func testAccCheckConfigurationTemplateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticBeanstalkConn()

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

func testAccCheckConfigurationTemplateExists(ctx context.Context, n string, v *elasticbeanstalk.ConfigurationSettingsDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Elastic Beanstalk Configuration Template ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticBeanstalkConn()

		output, err := tfelasticbeanstalk.FindConfigurationSettingsByTwoPartKey(ctx, conn, rs.Primary.Attributes["application"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccConfigurationTemplateConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "test" {
  name        = %[1]q
  description = "testing"
}

resource "aws_elastic_beanstalk_configuration_template" "test" {
  name                = %[1]q
  application         = aws_elastic_beanstalk_application.test.name
  solution_stack_name = "64bit Amazon Linux running Python"
}
`, rName)
}

func testAccConfigurationTemplateConfig_vpc(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "test" {
  name        = %[1]q
  description = "testing"
}

resource "aws_elastic_beanstalk_configuration_template" "test" {
  name        = %[1]q
  application = aws_elastic_beanstalk_application.test.name

  solution_stack_name = "64bit Amazon Linux running Python"

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
	return fmt.Sprintf(`
resource "aws_elastic_beanstalk_application" "test" {
  name        = %[1]q
  description = "testing"
}

resource "aws_elastic_beanstalk_configuration_template" "test" {
  name        = %[1]q
  application = aws_elastic_beanstalk_application.test.name

  solution_stack_name = "64bit Amazon Linux running Python"

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "InstanceType"
    value     = "m1.small"
  }
}
`, rName)
}
