package configservice_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconfig "github.com/hashicorp/terraform-provider-aws/internal/service/configservice"
)

func TestAccConfigServiceConfigurationAggregator_account(t *testing.T) {
	var ca configservice.ConfigurationAggregator
	//Name is upper case on purpose to test https://github.com/hashicorp/terraform-provider-aws/issues/8432
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_configuration_aggregator.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationAggregatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationAggregatorConfig_account(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationAggregatorExists(resourceName, &ca),
					testAccCheckConfigurationAggregatorName(resourceName, rName, &ca),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(`config-aggregator/config-aggregator-.+`)),
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.0.account_ids.#", "1"),
					acctest.CheckResourceAttrAccountID(resourceName, "account_aggregation_source.0.account_ids.0"),
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.0.regions.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "account_aggregation_source.0.regions.0", "data.aws_region.current", "name"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccConfigServiceConfigurationAggregator_organization(t *testing.T) {
	var ca configservice.ConfigurationAggregator
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_configuration_aggregator.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationAggregatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationAggregatorConfig_organization(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationAggregatorExists(resourceName, &ca),
					testAccCheckConfigurationAggregatorName(resourceName, rName, &ca),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "organization_aggregation_source.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_aggregation_source.0.role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "organization_aggregation_source.0.all_regions", "true"),
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

func TestAccConfigServiceConfigurationAggregator_switch(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_configuration_aggregator.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationAggregatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationAggregatorConfig_account(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "organization_aggregation_source.#", "0"),
				),
			},
			{
				Config: testAccConfigurationAggregatorConfig_organization(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "organization_aggregation_source.#", "1"),
				),
			},
		},
	})
}

func TestAccConfigServiceConfigurationAggregator_tags(t *testing.T) {
	var ca configservice.ConfigurationAggregator
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_configuration_aggregator.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationAggregatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationAggregatorConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationAggregatorExists(resourceName, &ca),
					testAccCheckConfigurationAggregatorName(resourceName, rName, &ca),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccConfigurationAggregatorConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationAggregatorExists(resourceName, &ca),
					testAccCheckConfigurationAggregatorName(resourceName, rName, &ca),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationAggregatorConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationAggregatorExists(resourceName, &ca),
					testAccCheckConfigurationAggregatorName(resourceName, rName, &ca),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccConfigServiceConfigurationAggregator_disappears(t *testing.T) {
	var ca configservice.ConfigurationAggregator
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_configuration_aggregator.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConfigurationAggregatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationAggregatorConfig_account(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationAggregatorExists(resourceName, &ca),
					acctest.CheckResourceDisappears(acctest.Provider, tfconfig.ResourceConfigurationAggregator(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckConfigurationAggregatorName(n, desired string, obj *configservice.ConfigurationAggregator) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.Attributes["name"] != aws.StringValue(obj.ConfigurationAggregatorName) {
			return fmt.Errorf("expected name: %q, given: %q", desired, aws.StringValue(obj.ConfigurationAggregatorName))
		}
		return nil
	}
}

func testAccCheckConfigurationAggregatorExists(n string, obj *configservice.ConfigurationAggregator) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No config configuration aggregator ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceConn
		out, err := conn.DescribeConfigurationAggregators(&configservice.DescribeConfigurationAggregatorsInput{
			ConfigurationAggregatorNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		})
		if err != nil {
			return fmt.Errorf("Failed to describe config configuration aggregator: %s", err)
		}
		if len(out.ConfigurationAggregators) < 1 {
			return fmt.Errorf("No config configuration aggregator found when describing %q", rs.Primary.Attributes["name"])
		}

		ca := out.ConfigurationAggregators[0]
		*obj = *ca

		return nil
	}
}

func testAccCheckConfigurationAggregatorDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_config_configuration_aggregator" {
			continue
		}

		resp, err := conn.DescribeConfigurationAggregators(&configservice.DescribeConfigurationAggregatorsInput{
			ConfigurationAggregatorNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		})

		if err == nil {
			if len(resp.ConfigurationAggregators) != 0 &&
				aws.StringValue(resp.ConfigurationAggregators[0].ConfigurationAggregatorName) == rs.Primary.Attributes["name"] {
				return fmt.Errorf("config configuration aggregator still exists: %s", rs.Primary.Attributes["name"])
			}
		}
	}

	return nil
}

func testAccConfigurationAggregatorConfig_account(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_config_configuration_aggregator" "test" {
  name = %[1]q

  account_aggregation_source {
    account_ids = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.name]
  }
}
`, rName)
}

func testAccConfigurationAggregatorConfig_organization(rName string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["config.${data.aws_partition.current.dns_suffix}"]
}

resource "aws_config_configuration_aggregator" "test" {
  depends_on = [aws_iam_role_policy_attachment.test, aws_organizations_organization.test]

  name = %[1]q

  organization_aggregation_source {
    all_regions = true
    role_arn    = aws_iam_role.test.arn
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "config.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSConfigRoleForOrganizations"
}
`, rName)
}

func testAccConfigurationAggregatorConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_config_configuration_aggregator" "test" {
  name = %[1]q

  account_aggregation_source {
    account_ids = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.name]
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccConfigurationAggregatorConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_config_configuration_aggregator" "test" {
  name = %[1]q

  account_aggregation_source {
    account_ids = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.name]
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
