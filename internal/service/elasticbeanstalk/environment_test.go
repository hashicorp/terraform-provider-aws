// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk_test

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk"
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

func TestAccElasticBeanstalkEnvironment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"
	beanstalkAsgNameRegexp := regexache.MustCompile("awseb.+?AutoScalingGroup[^,]+")
	beanstalkElbNameRegexp := regexache.MustCompile("awseb.+?EBLoa[^,]+")
	beanstalkInstancesNameRegexp := regexache.MustCompile("i-([0-9A-Fa-f]{8}|[0-9A-Fa-f]{17})")
	beanstalkLcNameRegexp := regexache.MustCompile("awseb.+?AutoScalingLaunch[^,]+")
	beanstalkEndpointURL := regexache.MustCompile("awseb.+?EBLoa[^,].+?elb.amazonaws.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticbeanstalk", fmt.Sprintf("environment/%s/%s", rName, rName)),
					resource.TestMatchResourceAttr(resourceName, "autoscaling_groups.0", beanstalkAsgNameRegexp),
					resource.TestMatchResourceAttr(resourceName, "endpoint_url", beanstalkEndpointURL),
					resource.TestMatchResourceAttr(resourceName, "instances.0", beanstalkInstancesNameRegexp),
					resource.TestMatchResourceAttr(resourceName, "launch_configurations.0", beanstalkLcNameRegexp),
					resource.TestMatchResourceAttr(resourceName, "load_balancers.0", beanstalkElbNameRegexp),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"setting",
					"wait_for_ready_timeout",
				},
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelasticbeanstalk.ResourceEnvironment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_tier(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"
	beanstalkQueuesNameRegexp := regexache.MustCompile("https://sqs.+?awseb[^,]+")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_worker(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
					resource.TestMatchResourceAttr(resourceName, "queues.0", beanstalkQueuesNameRegexp),
					resource.TestCheckResourceAttr(resourceName, "tier", "Worker"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"setting",
					"wait_for_ready_timeout",
				},
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_cnamePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"
	beanstalkCnameRegexp := regexache.MustCompile("^" + rName + ".+?elasticbeanstalk.com$")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_cnamePrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
					resource.TestMatchResourceAttr(resourceName, "cname", beanstalkCnameRegexp),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"setting",
					"wait_for_ready_timeout",
				},
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_beanstalkEnv(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_template(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
					testAccCheckEnvironmentConfigValue(ctx, resourceName, acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"setting",
					"template_name",
					"wait_for_ready_timeout",
				},
			},
			{
				Config: testAccEnvironmentConfig_template(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
					testAccCheckEnvironmentConfigValue(ctx, resourceName, acctest.Ct2),
				),
			},
			{
				Config: testAccEnvironmentConfig_template(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
					testAccCheckEnvironmentConfigValue(ctx, resourceName, acctest.Ct3),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_resource(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_resourceOptionSetting(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"setting",
					"wait_for_ready_timeout",
				},
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"setting",
					"wait_for_ready_timeout",
				},
			},
			{
				Config: testAccEnvironmentConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccEnvironmentConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_changeStack(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_templateChangeStack(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
				),
			},
			{
				Config: testAccEnvironmentConfig_templateChangeTemp(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
				),
			},
			{
				Config: testAccEnvironmentConfig_templateChangeStack(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_update(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
					testAccVerifyConfig(ctx, &app, []string{}),
				),
			},
			{
				Config: testAccEnvironmentConfig_settings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
					testAccVerifyConfig(ctx, &app, []string{"ENV_STATIC", "ENV_UPDATE"}),
				),
			},
			{
				Config: testAccEnvironmentConfig_settings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
					testAccVerifyConfig(ctx, &app, []string{"ENV_STATIC", "ENV_UPDATE"}),
				),
			},
			{
				Config: testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
					testAccVerifyConfig(ctx, &app, []string{}),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_label(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_applicationVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttrSet(resourceName, "version_label"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"setting",
					"wait_for_ready_timeout",
				},
			},
			{
				Config: testAccEnvironmentConfig_applicationVersionUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
					resource.TestCheckResourceAttrSet(resourceName, "version_label"),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_settingWithJSONValue(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"
	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)

	if err != nil {
		t.Fatalf("generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_settingJSONValue(rName, publicKey, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"setting",
					"wait_for_ready_timeout",
				},
			},
		},
	})
}

func TestAccElasticBeanstalkEnvironment_platformARN(t *testing.T) {
	ctx := acctest.Context(t)
	var app awstypes.EnvironmentDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elastic_beanstalk_environment.test"
	platformNameWithVersion1 := "Python 3.8 running on 64bit Amazon Linux 2/3.7.0"
	rValue1 := sdkacctest.RandIntRange(1000, 2000)
	rValue1Str := strconv.Itoa(rValue1)
	platformNameWithVersion2 := "Python 3.9 running on 64bit Amazon Linux 2023/4.1.1"
	rValue2 := sdkacctest.RandIntRange(3000, 4000)
	rValue2Str := strconv.Itoa(rValue2)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElasticBeanstalkServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_platformARN(rName, platformNameWithVersion1, rValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
					acctest.CheckResourceAttrRegionalARNNoAccount(resourceName, "platform_arn", "elasticbeanstalk", "platform/"+platformNameWithVersion1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "5"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", rValue1Str),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", rValue1Str),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", rValue1Str),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", rValue1Str),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", rValue1Str),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"setting",
					"wait_for_ready_timeout",
				},
			},
			{
				Config: testAccEnvironmentConfig_platformARN(rName, platformNameWithVersion2, rValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &app),
					acctest.CheckResourceAttrRegionalARNNoAccount(resourceName, "platform_arn", "elasticbeanstalk", "platform/"+platformNameWithVersion2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "5"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", rValue2Str),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", rValue2Str),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", rValue2Str),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", rValue2Str),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", rValue2Str),
				),
			},
		},
	})
}

func testAccCheckEnvironmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticBeanstalkClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elastic_beanstalk_environment" {
				continue
			}

			_, err := tfelasticbeanstalk.FindEnvironmentByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Elastic Beanstalk Environment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckEnvironmentExists(ctx context.Context, n string, v *awstypes.EnvironmentDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Elastic Beanstalk Environment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticBeanstalkClient(ctx)

		output, err := tfelasticbeanstalk.FindEnvironmentByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVerifyConfig(ctx context.Context, env *awstypes.EnvironmentDescription, expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if env == nil {
			return fmt.Errorf("Nil environment in testAccVerifyConfig")
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticBeanstalkClient(ctx)

		resp, err := conn.DescribeConfigurationSettings(ctx, &elasticbeanstalk.DescribeConfigurationSettingsInput{
			ApplicationName: env.ApplicationName,
			EnvironmentName: env.EnvironmentName,
		})

		if err != nil {
			return fmt.Errorf("Error describing config settings in testAccVerifyConfig: %s", err)
		}

		// should only be 1 environment
		if len(resp.ConfigurationSettings) != 1 {
			return fmt.Errorf("Expected only 1 set of Configuration Settings in testAccVerifyConfig, got (%d)", len(resp.ConfigurationSettings))
		}

		cs := resp.ConfigurationSettings[0]

		var foundEnvs []string
		testStrings := []string{"ENV_STATIC", "ENV_UPDATE"}
		for _, os := range cs.OptionSettings {
			for _, k := range testStrings {
				if *os.OptionName == k {
					foundEnvs = append(foundEnvs, k)
				}
			}
		}

		// if expected is zero, then we should not have found any of the predefined
		// env vars
		if len(expected) == 0 {
			if len(foundEnvs) > 0 {
				return fmt.Errorf("Found configs we should not have: %#v", foundEnvs)
			}
			return nil
		}

		sort.Strings(testStrings)
		sort.Strings(expected)
		if !reflect.DeepEqual(testStrings, expected) {
			return fmt.Errorf("error matching strings, expected:\n\n%#v\n\ngot:\n\n%#v", testStrings, foundEnvs)
		}

		return nil
	}
}

func testAccCheckEnvironmentConfigValue(ctx context.Context, n string, expectedValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticBeanstalkClient(ctx)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Elastic Beanstalk ENV is not set")
		}

		resp, err := conn.DescribeConfigurationOptions(ctx, &elasticbeanstalk.DescribeConfigurationOptionsInput{
			ApplicationName: aws.String(rs.Primary.Attributes["application"]),
			EnvironmentName: aws.String(rs.Primary.Attributes[names.AttrName]),
			Options: []awstypes.OptionSpecification{
				{
					Namespace:  aws.String("aws:elasticbeanstalk:application:environment"),
					OptionName: aws.String("TEMPLATE"),
				},
			},
		})
		if err != nil {
			return err
		}

		if len(resp.Options) != 1 {
			return fmt.Errorf("Found %d options, expected 1.", len(resp.Options))
		}

		log.Printf("[DEBUG] %d Elastic Beanstalk Option values returned.", len(resp.Options[0].ValueOptions))

		for _, value := range resp.Options[0].ValueOptions {
			if value != expectedValue {
				return fmt.Errorf("Option setting value: %s. Expected %s", value, expectedValue)
			}
		}

		return nil
	}
}

func testAccEnvironmentConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
data "aws_elastic_beanstalk_solution_stack" "test" {
  most_recent = true
  name_regex  = "64bit Amazon Linux .* running Python .*"
}

data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test.id
  route_table_id         = aws_vpc.test.main_route_table_id
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_elastic_beanstalk_application" "test" {
  description = "tf-test-desc"
  name        = %[1]q
}

# Create custom service role per test to remove dependency on
# Service-Linked Role existing.
resource "aws_iam_role" "service_role" {
  name = "%[1]s-service"
  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Condition = {
        StringEquals = {
          "sts:ExternalId" = "elasticbeanstalk"
        }
      }
      Effect = "Allow"
      Principal = {
        Service = "elasticbeanstalk.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "service_role-AWSElasticBeanstalkEnhancedHealth" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSElasticBeanstalkEnhancedHealth"
  role       = aws_iam_role.service_role.id
}

resource "aws_iam_role_policy_attachment" "service_role-AWSElasticBeanstalkService" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSElasticBeanstalkService"
  role       = aws_iam_role.service_role.id
}

# Amazon Linux 2 environments require IAM Instance Profile.
resource "aws_iam_role" "instance_profile" {
  name = "%[1]s-instance"
  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "instance_profile-AWSElasticBeanstalkWebTier" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AWSElasticBeanstalkWebTier"
  role       = aws_iam_role.instance_profile.id
}

resource "aws_iam_role_policy_attachment" "instance_profile-AWSElasticBeanstalkWorkerTier" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AWSElasticBeanstalkWorkerTier"
  role       = aws_iam_role.instance_profile.id
}

# Since the IAM Instance Profile is required anyways, also use it to
# ensure IAM Role permissions for both roles are attached.
resource "aws_iam_instance_profile" "test" {
  depends_on = [
    aws_iam_role_policy_attachment.instance_profile-AWSElasticBeanstalkWebTier,
    aws_iam_role_policy_attachment.instance_profile-AWSElasticBeanstalkWorkerTier,
    aws_iam_role_policy_attachment.service_role-AWSElasticBeanstalkEnhancedHealth,
    aws_iam_role_policy_attachment.service_role-AWSElasticBeanstalkService,
  ]

  name = %[1]q
  role = aws_iam_role.instance_profile.name
}
`, rName))
}

func testAccEnvironmentConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
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

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }
}
`, rName))
}

func testAccEnvironmentConfig_platformARN(rName, platformNameWithVersion string, rValue int) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application  = aws_elastic_beanstalk_application.test.name
  name         = %[1]q
  platform_arn = "arn:${data.aws_partition.current.partition}:elasticbeanstalk:${data.aws_region.current.name}::platform/%[2]s"

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

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }

  tags = {
    Key1 = "%[3]d"
    Key2 = "%[3]d"
    Key3 = "%[3]d"
    Key4 = "%[3]d"
    Key5 = "%[3]d"
  }
}
`, rName, platformNameWithVersion, rValue))
}

func testAccEnvironmentConfig_settings(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
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

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "ENV_STATIC"
    value     = "true"
  }

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "ENV_UPDATE"
    value     = "true"
  }

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "ENV_REMOVE"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:scheduledaction"
    resource  = "ScheduledAction01"
    name      = "MinSize"
    value     = 2
  }

  setting {
    namespace = "aws:autoscaling:scheduledaction"
    resource  = "ScheduledAction01"
    name      = "MaxSize"
    value     = 3
  }

  setting {
    namespace = "aws:autoscaling:scheduledaction"
    resource  = "ScheduledAction01"
    name      = "StartTime"
    value     = "2016-07-28T04:07:02Z"
  }
}
`, rName))
}

func testAccEnvironmentConfig_worker(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name
  tier                = "Worker"

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

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }
}
`, rName))
}

func testAccEnvironmentConfig_cnamePrefix(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  cname_prefix        = %[1]q
  name                = %[1]q
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

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }
}
`, rName))
}

func testAccEnvironmentConfig_template(rName string, cfgTplValue int) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application   = aws_elastic_beanstalk_application.test.name
  name          = %[1]q
  template_name = aws_elastic_beanstalk_configuration_template.test.name
}

resource "aws_elastic_beanstalk_configuration_template" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
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

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:application:environment"
    name      = "TEMPLATE"
    value     = %[2]d
  }
}
`, rName, cfgTplValue))
}

func testAccEnvironmentConfig_resourceOptionSetting(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
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

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }

  setting {
    namespace = "aws:autoscaling:scheduledaction"
    resource  = "ScheduledAction01"
    name      = "MinSize"
    value     = "2"
  }

  setting {
    namespace = "aws:autoscaling:scheduledaction"
    resource  = "ScheduledAction01"
    name      = "MaxSize"
    value     = "6"
  }

  setting {
    namespace = "aws:autoscaling:scheduledaction"
    resource  = "ScheduledAction01"
    name      = "Recurrence"
    value     = "0 8 * * *"
  }
}
`, rName))
}

func testAccEnvironmentConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
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

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccEnvironmentConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
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

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccEnvironmentConfig_templateChangeStack(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
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

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }
}

resource "aws_elastic_beanstalk_configuration_template" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name
}
`, rName))
}

func testAccEnvironmentConfig_templateChangeTemp(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_elastic_beanstalk_environment" "test" {
  application   = aws_elastic_beanstalk_application.test.name
  name          = %[1]q
  template_name = aws_elastic_beanstalk_configuration_template.test.name

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

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }
}

resource "aws_elastic_beanstalk_configuration_template" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name
}
`, rName))
}

func testAccEnvironmentConfig_applicationVersion(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "python-v1.zip"
  source = "test-fixtures/python-v1.zip"
}

resource "aws_elastic_beanstalk_application_version" "test" {
  application = aws_elastic_beanstalk_application.test.name
  bucket      = aws_s3_bucket.test.id
  key         = aws_s3_object.test.id
  name        = "%[1]s-1"
}

resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name
  version_label       = aws_elastic_beanstalk_application_version.test.name

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

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }
}
`, rName))
}

func testAccEnvironmentConfig_applicationVersionUpdate(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "python-v1.zip"
  source = "test-fixtures/python-v1.zip"
}

resource "aws_elastic_beanstalk_application_version" "test" {
  application = aws_elastic_beanstalk_application.test.name
  bucket      = aws_s3_bucket.test.id
  key         = aws_s3_object.test.id
  name        = "%[1]s-2"
}

resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name
  version_label       = aws_elastic_beanstalk_application_version.test.name

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

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }
}
`, rName))
}

func testAccEnvironmentConfig_settingJSONValue(rName, publicKey, email string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q
}

resource "aws_elastic_beanstalk_environment" "test" {
  application         = aws_elastic_beanstalk_application.test.name
  name                = %[1]q
  solution_stack_name = data.aws_elastic_beanstalk_solution_stack.test.name
  tier                = "Worker"

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

  setting {
    namespace = "aws:ec2:vpc"
    name      = "AssociatePublicIpAddress"
    value     = "true"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "SecurityGroups"
    value     = aws_security_group.test.id
  }

  setting {
    namespace = "aws:elasticbeanstalk:command"
    name      = "BatchSize"
    value     = "30"
  }

  setting {
    namespace = "aws:elasticbeanstalk:command"
    name      = "BatchSizeType"
    value     = "Percentage"
  }

  setting {
    namespace = "aws:elasticbeanstalk:command"
    name      = "DeploymentPolicy"
    value     = "Rolling"
  }

  setting {
    namespace = "aws:elasticbeanstalk:sns:topics"
    name      = "Notification Endpoint"
    value     = %[3]q
  }

  setting {
    namespace = "aws:elasticbeanstalk:sqsd"
    name      = "ErrorVisibilityTimeout"
    value     = "2"
  }

  setting {
    namespace = "aws:elasticbeanstalk:sqsd"
    name      = "HttpPath"
    value     = "/event-message"
  }

  setting {
    namespace = "aws:elasticbeanstalk:sqsd"
    name      = "WorkerQueueURL"
    value     = aws_sqs_queue.test.id
  }

  setting {
    namespace = "aws:elasticbeanstalk:sqsd"
    name      = "VisibilityTimeout"
    value     = "300"
  }

  setting {
    namespace = "aws:elasticbeanstalk:sqsd"
    name      = "HttpConnections"
    value     = "10"
  }

  setting {
    namespace = "aws:elasticbeanstalk:sqsd"
    name      = "InactivityTimeout"
    value     = "299"
  }

  setting {
    namespace = "aws:elasticbeanstalk:sqsd"
    name      = "MimeType"
    value     = "application/json"
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "EnvironmentType"
    value     = "LoadBalanced"
  }

  setting {
    namespace = "aws:elasticbeanstalk:application"
    name      = "Application Healthcheck URL"
    value     = "/health"
  }

  setting {
    namespace = "aws:elasticbeanstalk:healthreporting:system"
    name      = "SystemType"
    value     = "enhanced"
  }

  setting {
    namespace = "aws:elasticbeanstalk:healthreporting:system"
    name      = "HealthCheckSuccessThreshold"
    value     = "Ok"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "IamInstanceProfile"
    value     = aws_iam_instance_profile.test.name
  }

  setting {
    namespace = "aws:elasticbeanstalk:environment"
    name      = "ServiceRole"
    value     = aws_iam_role.service_role.name
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "InstanceType"
    value     = "t2.micro"
  }

  setting {
    namespace = "aws:autoscaling:launchconfiguration"
    name      = "EC2KeyName"
    value     = aws_key_pair.test.key_name
  }

  setting {
    namespace = "aws:autoscaling:updatepolicy:rollingupdate"
    name      = "RollingUpdateEnabled"
    value     = "false"
  }

  setting {
    namespace = "aws:elasticbeanstalk:healthreporting:system"
    name      = "ConfigDocument"

    value = <<EOF
{
	"Version": 1,
	"CloudWatchMetrics": {
		"Instance": {
			"ApplicationRequestsTotal": 60
		},
		"Environment": {
			"ApplicationRequests5xx": 60,
			"ApplicationRequests4xx": 60,
			"ApplicationRequests2xx": 60
		}
	}
}
EOF
  }
}
`, rName, publicKey, email))
}
