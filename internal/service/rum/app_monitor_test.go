package rum_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchrum"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatchrum "github.com/hashicorp/terraform-provider-aws/internal/service/rum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRUMAppMonitor_basic(t *testing.T) {
	var appMon cloudwatchrum.AppMonitor
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rum_app_monitor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatchrum.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppMonitorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "rum", fmt.Sprintf("appmonitor/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "domain", "localhost"),
					resource.TestCheckResourceAttr(resourceName, "cw_log_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.0.session_sample_rate", "0.1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAppMonitorConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "rum", fmt.Sprintf("appmonitor/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "domain", "localhost"),
					resource.TestCheckResourceAttr(resourceName, "cw_log_enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "cw_log_group"),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "app_monitor_configuration.0.session_sample_rate", "0.1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccRUMAppMonitor_tags(t *testing.T) {
	var appMon cloudwatchrum.AppMonitor
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rum_app_monitor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatchrum.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppMonitorConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(resourceName, &appMon),
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
				Config: testAccAppMonitorConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAppMonitorConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(resourceName, &appMon),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccRUMAppMonitor_disappears(t *testing.T) {
	var appMon cloudwatchrum.AppMonitor
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rum_app_monitor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatchrum.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAppMonitorConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppMonitorExists(resourceName, &appMon),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudwatchrum.ResourceAppMonitor(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudwatchrum.ResourceAppMonitor(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAppMonitorDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RUMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_rum_app_monitor" {
			continue
		}

		appMon, err := tfcloudwatchrum.FindAppMonitorByName(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		if aws.StringValue(appMon.Name) == rs.Primary.ID {
			return fmt.Errorf("cloudwatchrum App Monitor %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAppMonitorExists(n string, appMon *cloudwatchrum.AppMonitor) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No cloudwatchrum App Monitor ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RUMConn
		resp, err := tfcloudwatchrum.FindAppMonitorByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*appMon = *resp

		return nil
	}
}

func testAccAppMonitorConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_rum_app_monitor" "test" {
  name   = %[1]q
  domain = "localhost"
}
`, rName)
}

func testAccAppMonitorConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_rum_app_monitor" "test" {
  name           = %[1]q
  domain         = "localhost"
  cw_log_enabled = true
}
`, rName)
}

func testAccAppMonitorConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_rum_app_monitor" "test" {
  name   = %[1]q
  domain = "localhost"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAppMonitorConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_rum_app_monitor" "test" {
  name   = %[1]q
  domain = "localhost"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
