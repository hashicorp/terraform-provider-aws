package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/route53"
	r53rcc "github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/route53/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_route53_health_check", &resource.Sweeper{
		Name: "aws_route53_health_check",
		F:    testSweepRoute53Healthchecks,
	})
}

func testSweepRoute53Healthchecks(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).r53conn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	input := &route53.ListHealthChecksInput{}

	err = conn.ListHealthChecksPages(input, func(page *route53.ListHealthChecksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, detail := range page.HealthChecks {
			if detail == nil {
				continue
			}

			id := aws.StringValue(detail.Id)

			r := resourceAwsRoute53HealthCheck()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Route53 Health Checks for %s: %w", region, err))
	}

	if err = testSweepResourceOrchestratorContext(context.Background(), sweepResources, 0*time.Minute, 1*time.Minute, 10*time.Second, 18*time.Second, 10*time.Minute); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Route53 Health Checks for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Route53 Health Checks sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSRoute53HealthCheck_basic(t *testing.T) {
	var check route53.HealthCheck
	resourceName := "aws_route53_health_check.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, route53.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53HealthCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53HealthCheckConfigBasic("2", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName, &check),
					testAccMatchResourceAttrGlobalARNNoAccount(resourceName, "arn", "route53", regexp.MustCompile("healthcheck/.+")),
					resource.TestCheckResourceAttr(resourceName, "measure_latency", "true"),
					resource.TestCheckResourceAttr(resourceName, "port", "80"),
					resource.TestCheckResourceAttr(resourceName, "failure_threshold", "2"),
					resource.TestCheckResourceAttr(resourceName, "invert_healthcheck", "true"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoute53HealthCheckConfigBasic("5", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "failure_threshold", "5"),
					resource.TestCheckResourceAttr(resourceName, "invert_healthcheck", "false"),
				),
			},
		},
	})
}

func TestAccAWSRoute53HealthCheck_tags(t *testing.T) {
	var check route53.HealthCheck
	resourceName := "aws_route53_health_check.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, route53.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53HealthCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53HealthCheckConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName, &check),
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
				Config: testAccRoute53HealthCheckConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRoute53HealthCheckConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSRoute53HealthCheck_withSearchString(t *testing.T) {
	var check route53.HealthCheck
	resourceName := "aws_route53_health_check.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, route53.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53HealthCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53HealthCheckSearchStringConfig("OK", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "invert_healthcheck", "false"),
					resource.TestCheckResourceAttr(resourceName, "search_string", "OK"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoute53HealthCheckSearchStringConfig("FAILED", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "invert_healthcheck", "true"),
					resource.TestCheckResourceAttr(resourceName, "search_string", "FAILED"),
				),
			},
		},
	})
}

func TestAccAWSRoute53HealthCheck_withChildHealthChecks(t *testing.T) {
	var check route53.HealthCheck
	resourceName := "aws_route53_health_check.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, route53.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53HealthCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53HealthCheckConfig_withChildHealthChecks,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName, &check),
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

func TestAccAWSRoute53HealthCheck_withHealthCheckRegions(t *testing.T) {
	var check route53.HealthCheck
	resourceName := "aws_route53_health_check.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionPreCheck("aws", t) }, // GovCloud has 2 regions, test requires 3
		ErrorCheck:   testAccErrorCheck(t, route53.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53HealthCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53HealthCheckConfig_withHealthCheckRegions(endpoints.UsWest2RegionID, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "3"),
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

func TestAccAWSRoute53HealthCheck_IpConfig(t *testing.T) {
	var check route53.HealthCheck
	resourceName := "aws_route53_health_check.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, route53.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53HealthCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53HealthCheckIpConfig("1.2.3.4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "ip_address", "1.2.3.4"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoute53HealthCheckIpConfig("1.2.3.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "ip_address", "1.2.3.5"),
				),
			},
		},
	})
}

func TestAccAWSRoute53HealthCheck_Ipv6Config(t *testing.T) {
	var check route53.HealthCheck
	resourceName := "aws_route53_health_check.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, route53.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53HealthCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53HealthCheckIpConfig("1234:5678:9abc:6811:0:0:0:4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "ip_address", "1234:5678:9abc:6811:0:0:0:4"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:   testAccRoute53HealthCheckIpConfig("1234:5678:9abc:6811:0:0:0:4"),
				PlanOnly: true,
			},
		},
	})
}

func TestAccAWSRoute53HealthCheck_CloudWatchAlarmCheck(t *testing.T) {
	var check route53.HealthCheck
	resourceName := "aws_route53_health_check.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, route53.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53HealthCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53HealthCheckCloudWatchAlarm,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm_name", "cloudwatch-healthcheck-alarm"),
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

func TestAccAWSRoute53HealthCheck_withSNI(t *testing.T) {
	var check route53.HealthCheck
	resourceName := "aws_route53_health_check.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, route53.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53HealthCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53HealthCheckConfigWithoutSNI,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "enable_sni", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoute53HealthCheckSNIConfig(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "enable_sni", "false"),
				),
			},
			{
				Config: testAccRoute53HealthCheckSNIConfig(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "enable_sni", "true"),
				),
			},
		},
	})
}

func TestAccAWSRoute53HealthCheck_Disabled(t *testing.T) {
	var check route53.HealthCheck
	resourceName := "aws_route53_health_check.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, route53.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53HealthCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53HealthCheckConfigDisabled(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "disabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoute53HealthCheckConfigDisabled(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "disabled", "false"),
				),
			},
			{
				Config: testAccRoute53HealthCheckConfigDisabled(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "disabled", "true"),
				),
			},
		},
	})
}

func TestAccAWSRoute53HealthCheck_withRoutingControlArn(t *testing.T) {
	var check route53.HealthCheck
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_health_check.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(r53rcc.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, route53.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53HealthCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53HealthCheckConfigRoutingControlArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "type", "RECOVERY_CONTROL"),
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

func TestAccAWSRoute53HealthCheck_disappears(t *testing.T) {
	var check route53.HealthCheck
	resourceName := "aws_route53_health_check.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, route53.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53HealthCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53HealthCheckConfigBasic("2", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName, &check),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRoute53HealthCheck(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRoute53HealthCheckDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).r53conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_health_check" {
			continue
		}

		_, err := finder.HealthCheckByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Route53 Health Check %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckRoute53HealthCheckExists(n string, v *route53.HealthCheck) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route53 Health Check ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).r53conn

		output, err := finder.HealthCheckByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccRoute53HealthCheckConfigBasic(thershold string, invert bool) string {
	return fmt.Sprintf(`
resource "aws_route53_health_check" "test" {
  fqdn               = "dev.example.com"
  port               = 80
  type               = "HTTP"
  resource_path      = "/"
  failure_threshold  = %[1]q
  request_interval   = "30"
  measure_latency    = true
  invert_healthcheck = %[2]t
}
`, thershold, invert)
}

func testAccRoute53HealthCheckConfigTags1(tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_route53_health_check" "test" {
  fqdn               = "dev.example.com"
  port               = 80
  type               = "HTTP"
  resource_path      = "/"
  failure_threshold  = "2"
  request_interval   = "30"
  measure_latency    = true
  invert_healthcheck = true

  tags = {
    %[1]q = %[2]q
  }
}
`, tag1Key, tag1Value)
}

func testAccRoute53HealthCheckConfigTags2(tag1Key, tag1Value, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_route53_health_check" "test" {
  fqdn               = "dev.example.com"
  port               = 80
  type               = "HTTP"
  resource_path      = "/"
  failure_threshold  = "2"
  request_interval   = "30"
  measure_latency    = true
  invert_healthcheck = true

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tag1Key, tag1Value, tagKey2, tagValue2)
}

func testAccRoute53HealthCheckIpConfig(ip string) string {
	return fmt.Sprintf(`
resource "aws_route53_health_check" "test" {
  ip_address        = %[1]q
  port              = 80
  type              = "HTTP"
  resource_path     = "/"
  failure_threshold = "2"
  request_interval  = "30"

  tags = {
    Name = "tf-test-health-check"
  }
}
`, ip)
}

const testAccRoute53HealthCheckConfig_withChildHealthChecks = `
resource "aws_route53_health_check" "child1" {
  fqdn              = "child1.example.com"
  port              = 80
  type              = "HTTP"
  resource_path     = "/"
  failure_threshold = "2"
  request_interval  = "30"
}

resource "aws_route53_health_check" "test" {
  type                   = "CALCULATED"
  child_health_threshold = 1
  child_healthchecks     = [aws_route53_health_check.child1.id]

  tags = {
    Name = "tf-test-calculated-health-check"
  }
}
`

func testAccRoute53HealthCheckConfig_withHealthCheckRegions(regions ...string) string {
	return fmt.Sprintf(`
resource "aws_route53_health_check" "test" {
  ip_address        = "1.2.3.4"
  port              = 80
  type              = "HTTP"
  resource_path     = "/"
  failure_threshold = "2"
  request_interval  = "30"

  regions = ["%s"]

  tags = {
    Name = "tf-test-check-with-regions"
  }
}
`, strings.Join(regions, "\", \""))
}

const testAccRoute53HealthCheckCloudWatchAlarm = `
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name          = "cloudwatch-healthcheck-alarm"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = "120"
  statistic           = "Average"
  threshold           = "80"
  alarm_description   = "This metric monitors ec2 cpu utilization"
}

data "aws_region" "current" {}

resource "aws_route53_health_check" "test" {
  type                            = "CLOUDWATCH_METRIC"
  cloudwatch_alarm_name           = aws_cloudwatch_metric_alarm.test.alarm_name
  cloudwatch_alarm_region         = data.aws_region.current.name
  insufficient_data_health_status = "Healthy"
}
`

func testAccRoute53HealthCheckSearchStringConfig(search string, invert bool) string {
	return fmt.Sprintf(`
resource "aws_route53_health_check" "test" {
  fqdn               = "dev.example.com"
  port               = 80
  type               = "HTTP_STR_MATCH"
  resource_path      = "/"
  failure_threshold  = "2"
  request_interval   = "30"
  measure_latency    = true
  invert_healthcheck = %[2]t
  search_string      = %[1]q

  tags = {
    Name = "tf-test-health-check"
  }
}
`, search, invert)
}

const testAccRoute53HealthCheckConfigWithoutSNI = `
resource "aws_route53_health_check" "test" {
  fqdn               = "dev.example.com"
  port               = 443
  type               = "HTTPS"
  resource_path      = "/"
  failure_threshold  = "2"
  request_interval   = "30"
  measure_latency    = true
  invert_healthcheck = true

  tags = {
    Name = "tf-test-health-check"
  }
}
`

func testAccRoute53HealthCheckSNIConfig(enable bool) string {
	return fmt.Sprintf(`
resource "aws_route53_health_check" "test" {
  fqdn               = "dev.example.com"
  port               = 443
  type               = "HTTPS"
  resource_path      = "/"
  failure_threshold  = "2"
  request_interval   = "30"
  measure_latency    = true
  invert_healthcheck = true
  enable_sni         = %[1]t

  tags = {
    Name = "tf-test-health-check"
  }
}
`, enable)
}

func testAccRoute53HealthCheckConfigDisabled(disabled bool) string {
	return fmt.Sprintf(`
resource "aws_route53_health_check" "test" {
  disabled          = %[1]t
  failure_threshold = "2"
  fqdn              = "dev.example.com"
  port              = 80
  request_interval  = "30"
  resource_path     = "/"
  type              = "HTTP"
}
`, disabled)
}

func testAccRoute53HealthCheckConfigRoutingControlArn(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53recoverycontrolconfig_cluster" "test" {
  name = %[1]q
}
resource "aws_route53recoverycontrolconfig_routing_control" "test" {
  name        = %[1]q
  cluster_arn = aws_route53recoverycontrolconfig_cluster.test.arn
}
resource "aws_route53_health_check" "test" {
  type                = "RECOVERY_CONTROL"
  routing_control_arn = aws_route53recoverycontrolconfig_routing_control.test.arn
}
`, rName)
}
