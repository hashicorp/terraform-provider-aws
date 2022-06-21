package cloudwatch_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatch "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatch"
)

func TestAccCloudWatchDashboard_basic(t *testing.T) {
	var dashboard cloudwatch.GetDashboardOutput
	resourceName := "aws_cloudwatch_dashboard.test"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDashboardExists(resourceName, &dashboard),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", testAccDashboardName(rInt)),
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

func TestAccCloudWatchDashboard_update(t *testing.T) {
	var dashboard cloudwatch.GetDashboardOutput
	resourceName := "aws_cloudwatch_dashboard.test"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDashboardExists(resourceName, &dashboard),
					testAccCheckDashboardBodyIsExpected(resourceName, basicWidget),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", testAccDashboardName(rInt)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDashboardConfig_updateBody(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDashboardExists(resourceName, &dashboard),
					testAccCheckDashboardBodyIsExpected(resourceName, updatedWidget),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", testAccDashboardName(rInt)),
				),
			},
		},
	})
}

func TestAccCloudWatchDashboard_updateName(t *testing.T) {
	var dashboard cloudwatch.GetDashboardOutput
	resourceName := "aws_cloudwatch_dashboard.test"
	rInt := sdkacctest.RandInt()
	rInt2 := sdkacctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDashboardExists(resourceName, &dashboard),
					testAccCheckDashboardBodyIsExpected(resourceName, basicWidget),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", testAccDashboardName(rInt)),
				),
			},
			{
				Config: testAccDashboardConfig_basic(rInt2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDashboardExists(resourceName, &dashboard),
					testAccCheckDashboardBodyIsExpected(resourceName, basicWidget),
					resource.TestCheckResourceAttr(resourceName, "dashboard_name", testAccDashboardName(rInt2)),
					testAccCheckDashboardDestroyPrevious(testAccDashboardName(rInt)),
				),
			},
		},
	})
}

func testAccCheckDashboardExists(n string, dashboard *cloudwatch.GetDashboardOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchConn
		params := cloudwatch.GetDashboardInput{
			DashboardName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetDashboard(&params)
		if err != nil {
			return err
		}

		*dashboard = *resp

		return nil
	}
}

func testAccCheckDashboardDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_dashboard" {
			continue
		}

		params := cloudwatch.GetDashboardInput{
			DashboardName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetDashboard(&params)
		if err == nil {
			return fmt.Errorf("Dashboard still exists: %s", rs.Primary.ID)
		}
		if !tfcloudwatch.IsDashboardNotFoundErr(err) {
			return err
		}
	}

	return nil
}

func testAccCheckDashboardDestroyPrevious(dashboardName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchConn

		params := cloudwatch.GetDashboardInput{
			DashboardName: aws.String(dashboardName),
		}

		_, err := conn.GetDashboard(&params)

		if err == nil {
			return fmt.Errorf("Dashboard still exists: %s", dashboardName)
		}

		if !tfcloudwatch.IsDashboardNotFoundErr(err) {
			return err
		}

		return nil
	}
}

const (
	basicWidget = `{
  "widgets": [
    {
      "type": "text",
      "x": 0,
      "y": 0,
      "width": 6,
      "height": 6,
      "properties": {
        "markdown": "Hi there from Terraform: CloudWatch"
      }
    }
  ]
}`

	updatedWidget = `{
  "widgets": [
    {
      "type": "text",
      "x": 0,
      "y": 0,
      "width": 6,
      "height": 6,
      "properties": {
        "markdown": "Hi there from Terraform: CloudWatch - updated"
      }
    }
  ]
}`
)

func testAccDashboardName(rInt int) string {
	return fmt.Sprintf("terraform-test-dashboard-%d", rInt)
}

func testAccDashboardConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_dashboard" "test" {
  dashboard_name = "terraform-test-dashboard-%d"

  dashboard_body = <<EOF
  %s
EOF
}
`, rInt, basicWidget)
}

func testAccDashboardConfig_updateBody(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_dashboard" "test" {
  dashboard_name = "terraform-test-dashboard-%d"

  dashboard_body = <<EOF
  %s
EOF
}
`, rInt, updatedWidget)
}

func testAccCheckDashboardBodyIsExpected(resourceName, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchConn
		params := cloudwatch.GetDashboardInput{
			DashboardName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetDashboard(&params)
		if err != nil {
			return err
		}

		var (
			bodyReader     = strings.NewReader(*resp.DashboardBody)
			expectedReader = strings.NewReader(expected)

			body         = make(map[string]interface{})
			expectedBody = make(map[string]interface{})
		)

		if err := json.NewDecoder(bodyReader).Decode(&body); err != nil {
			return fmt.Errorf("failed to parse received body: %s", err)
		} else if err := json.NewDecoder(expectedReader).Decode(&expectedBody); err != nil {
			return fmt.Errorf("failed to parse expected body: %s", err)
		}

		if !reflect.DeepEqual(body, expectedBody) {
			return fmt.Errorf("Expected %q dashboard body, got %q", expectedBody, body)
		}

		return nil
	}
}
