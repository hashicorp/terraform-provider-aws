package applicationinsights_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/applicationinsights"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapplicationinsights "github.com/hashicorp/terraform-provider-aws/internal/service/applicationinsights"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccApplicationInsightsApplication_basic(t *testing.T) {
	var app applicationinsights.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_applicationinsights_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, applicationinsights.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "resource_group_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "applicationinsights", fmt.Sprintf("application/resource-group/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "auto_config_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "cwe_monitor_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "ops_center_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "resource_group_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "applicationinsights", fmt.Sprintf("application/resource-group/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "auto_config_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "cwe_monitor_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "ops_center_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccApplicationInsightsApplication_autoConfig(t *testing.T) {
	var app applicationinsights.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_applicationinsights_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, applicationinsights.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "resource_group_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "applicationinsights", fmt.Sprintf("application/resource-group/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "auto_config_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "cwe_monitor_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "ops_center_enabled", "false"),
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

func TestAccApplicationInsightsApplication_tags(t *testing.T) {
	var app applicationinsights.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_applicationinsights_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, applicationinsights.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccApplicationConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &app),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccApplicationInsightsApplication_disappears(t *testing.T) {
	var app applicationinsights.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_applicationinsights_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, applicationinsights.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &app),
					acctest.CheckResourceDisappears(acctest.Provider, tfapplicationinsights.ResourceApplication(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfapplicationinsights.ResourceApplication(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckApplicationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ApplicationInsightsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_applicationinsights_application" {
			continue
		}

		app, err := tfapplicationinsights.FindApplicationByName(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		if aws.StringValue(app.ResourceGroupName) == rs.Primary.ID {
			return fmt.Errorf("applicationinsights Application %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckApplicationExists(n string, app *applicationinsights.ApplicationInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No applicationinsights Application ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ApplicationInsightsConn
		resp, err := tfapplicationinsights.FindApplicationByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*app = *resp

		return nil
	}
}

func testAccApplicationConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_resourcegroups_group" "test" {
  name = %[1]q

  resource_query {
    query = <<JSON
	{
		"ResourceTypeFilters": [
		  "AWS::EC2::Instance"
		],
		"TagFilters": [
		  {
			"Key": "Stage",
			"Values": [
			  "Test"
			]
		  }
		]
	  }
JSON
  }
}
`, rName)
}

func testAccApplicationConfig_basic(rName string) string {
	return testAccApplicationConfigBase(rName) + `
resource "aws_applicationinsights_application" "test" {
  resource_group_name = aws_resourcegroups_group.test.name
}
`
}

func testAccApplicationConfig_updated(rName string) string {
	return testAccApplicationConfigBase(rName) + `
resource "aws_applicationinsights_application" "test" {
  resource_group_name = aws_resourcegroups_group.test.name
  auto_config_enabled = true
}
`
}

func testAccApplicationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return testAccApplicationConfigBase(rName) + fmt.Sprintf(`
resource "aws_applicationinsights_application" "test" {
  resource_group_name = aws_resourcegroups_group.test.name

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccApplicationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccApplicationConfigBase(rName) + fmt.Sprintf(`
resource "aws_applicationinsights_application" "test" {
  resource_group_name = aws_resourcegroups_group.test.name

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
