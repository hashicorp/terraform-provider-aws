// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53HealthCheck_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var check awstypes.HealthCheck
	resourceName := "aws_route53_health_check.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHealthCheckDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHealthCheckConfig_basic("2", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					acctest.MatchResourceAttrGlobalARNNoAccount(resourceName, names.AttrARN, "route53", regexache.MustCompile("healthcheck/.+")),
					resource.TestCheckResourceAttr(resourceName, "measure_latency", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "80"),
					resource.TestCheckResourceAttr(resourceName, "failure_threshold", "2"),
					resource.TestCheckResourceAttr(resourceName, "invert_healthcheck", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHealthCheckConfig_basic("5", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "failure_threshold", "5"),
					resource.TestCheckResourceAttr(resourceName, "invert_healthcheck", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRoute53HealthCheck_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var check awstypes.HealthCheck
	resourceName := "aws_route53_health_check.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHealthCheckDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHealthCheckConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHealthCheckConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccHealthCheckConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRoute53HealthCheck_withSearchString(t *testing.T) {
	ctx := acctest.Context(t)
	var check awstypes.HealthCheck
	resourceName := "aws_route53_health_check.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHealthCheckDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHealthCheckConfig_searchString("OK", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "invert_healthcheck", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "search_string", "OK"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHealthCheckConfig_searchString("FAILED", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "invert_healthcheck", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "search_string", "FAILED"),
				),
			},
		},
	})
}

func TestAccRoute53HealthCheck_withChildHealthChecks(t *testing.T) {
	ctx := acctest.Context(t)
	var check awstypes.HealthCheck
	resourceName := "aws_route53_health_check.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHealthCheckDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHealthCheckConfig_childs(1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.HealthCheckTypeCalculated)),
					resource.TestCheckResourceAttr(resourceName, "child_health_threshold", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHealthCheckConfig_childs(0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.HealthCheckTypeCalculated)),
					resource.TestCheckResourceAttr(resourceName, "child_health_threshold", "0"),
				),
			},
		},
	})
}

func TestAccRoute53HealthCheck_withChildHealthChecksThresholdZero(t *testing.T) {
	ctx := acctest.Context(t)
	var check awstypes.HealthCheck
	resourceName := "aws_route53_health_check.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHealthCheckDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHealthCheckConfig_childs(0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.HealthCheckTypeCalculated)),
					resource.TestCheckResourceAttr(resourceName, "child_health_threshold", "0"),
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

func TestAccRoute53HealthCheck_withHealthCheckRegions(t *testing.T) {
	ctx := acctest.Context(t)
	var check awstypes.HealthCheck
	resourceName := "aws_route53_health_check.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartition(t, endpoints.AwsPartitionID) }, // GovCloud has 2 regions, test requires 3
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHealthCheckDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHealthCheckConfig_regions(
					string(awstypes.HealthCheckRegionUsWest2),
					string(awstypes.HealthCheckRegionUsEast1),
					string(awstypes.HealthCheckRegionEuWest1),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "regions.*", string(awstypes.HealthCheckRegionUsWest2)),
					resource.TestCheckTypeSetElemAttr(resourceName, "regions.*", string(awstypes.HealthCheckRegionUsEast1)),
					resource.TestCheckTypeSetElemAttr(resourceName, "regions.*", string(awstypes.HealthCheckRegionEuWest1)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHealthCheckConfig_regions(
					string(awstypes.HealthCheckRegionUsWest2),
					string(awstypes.HealthCheckRegionUsEast1),
					string(awstypes.HealthCheckRegionEuWest1),
					string(awstypes.HealthCheckRegionApNortheast1),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "4"),
					resource.TestCheckTypeSetElemAttr(resourceName, "regions.*", string(awstypes.HealthCheckRegionUsWest2)),
					resource.TestCheckTypeSetElemAttr(resourceName, "regions.*", string(awstypes.HealthCheckRegionUsEast1)),
					resource.TestCheckTypeSetElemAttr(resourceName, "regions.*", string(awstypes.HealthCheckRegionEuWest1)),
					resource.TestCheckTypeSetElemAttr(resourceName, "regions.*", string(awstypes.HealthCheckRegionApNortheast1)),
				),
			},
		},
	})
}

func TestAccRoute53HealthCheck_ip(t *testing.T) {
	ctx := acctest.Context(t)
	var check awstypes.HealthCheck
	resourceName := "aws_route53_health_check.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHealthCheckDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHealthCheckConfig_ip("1.2.3.4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddress, "1.2.3.4"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHealthCheckConfig_ip("1.2.3.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, names.AttrIPAddress, "1.2.3.5"),
				),
			},
		},
	})
}

func TestAccRoute53HealthCheck_ipv6(t *testing.T) {
	ctx := acctest.Context(t)
	var check awstypes.HealthCheck
	resourceName := "aws_route53_health_check.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHealthCheckDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHealthCheckConfig_ip("1234:5678:9abc:6811:0:0:0:4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrIPAddress), knownvalue.StringExact("1234:5678:9abc:6811:0:0:0:4")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHealthCheckConfig_ip("1234:5678:9abc:6811:0:0:0:4"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
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

func TestAccRoute53HealthCheck_cloudWatchAlarmCheck(t *testing.T) {
	ctx := acctest.Context(t)
	var check awstypes.HealthCheck
	resourceName := "aws_route53_health_check.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHealthCheckDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHealthCheckConfig_cloudWatchAlarm,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
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

func TestAccRoute53HealthCheck_triggers(t *testing.T) {
	ctx := acctest.Context(t)
	var check awstypes.HealthCheck
	resourceName := "aws_route53_health_check.test"
	alarmResourceName := "aws_cloudwatch_metric_alarm.test"
	threshold := "80"
	thresholdUpdated := "90"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHealthCheckDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHealthCheckConfig_triggers(threshold),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_alarm_name", alarmResourceName, "alarm_name"),
					resource.TestCheckResourceAttrPair(resourceName, "triggers.threshold", alarmResourceName, "threshold"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrTriggers},
			},
			{
				Config: testAccHealthCheckConfig_triggers(thresholdUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_alarm_name", alarmResourceName, "alarm_name"),
					resource.TestCheckResourceAttrPair(resourceName, "triggers.threshold", alarmResourceName, "threshold"),
				),
			},
		},
	})
}

func TestAccRoute53HealthCheck_withSNI(t *testing.T) {
	ctx := acctest.Context(t)
	var check awstypes.HealthCheck
	resourceName := "aws_route53_health_check.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHealthCheckDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHealthCheckConfig_noSNI,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "enable_sni", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHealthCheckConfig_sni(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "enable_sni", acctest.CtFalse),
				),
			},
			{
				Config: testAccHealthCheckConfig_sni(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "enable_sni", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRoute53HealthCheck_disabled(t *testing.T) {
	ctx := acctest.Context(t)
	var check awstypes.HealthCheck
	resourceName := "aws_route53_health_check.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHealthCheckDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHealthCheckConfig_disabled(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "disabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHealthCheckConfig_disabled(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "disabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccHealthCheckConfig_disabled(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, "disabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRoute53HealthCheck_withRoutingControlARN(t *testing.T) {
	ctx := acctest.Context(t)
	var check awstypes.HealthCheck
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_health_check.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Route53RecoveryControlConfigEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHealthCheckDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHealthCheckConfig_routingControlARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "RECOVERY_CONTROL"),
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

func TestAccRoute53HealthCheck_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var check awstypes.HealthCheck
	resourceName := "aws_route53_health_check.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHealthCheckDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHealthCheckConfig_basic("2", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHealthCheckExists(ctx, t, resourceName, &check),
					acctest.CheckSDKResourceDisappears(ctx, t, tfroute53.ResourceHealthCheck(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckHealthCheckDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).Route53Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_health_check" {
				continue
			}

			_, err := tfroute53.FindHealthCheckByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route53 Health Check %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckHealthCheckExists(ctx context.Context, t *testing.T, n string, v *awstypes.HealthCheck) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).Route53Client(ctx)

		output, err := tfroute53.FindHealthCheckByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccHealthCheckConfig_basic(thershold string, invert bool) string {
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

func testAccHealthCheckConfig_tags1(tag1Key, tag1Value string) string {
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

func testAccHealthCheckConfig_tags2(tag1Key, tag1Value, tagKey2, tagValue2 string) string {
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

func testAccHealthCheckConfig_ip(ip string) string {
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

func testAccHealthCheckConfig_childs(threshold int) string {
	return fmt.Sprintf(`
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
  child_health_threshold = %[1]d
  child_healthchecks     = [aws_route53_health_check.child1.id]

  tags = {
    Name = "tf-test-calculated-health-check"
  }
}
`, threshold)
}

func testAccHealthCheckConfig_regions(regions ...string) string {
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

const testAccHealthCheckConfig_cloudWatchAlarm = `
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
  cloudwatch_alarm_region         = data.aws_region.current.region
  insufficient_data_health_status = "Healthy"
}
`

func testAccHealthCheckConfig_triggers(threshold string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name          = "cloudwatch-healthcheck-alarm"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = "120"
  statistic           = "Average"
  threshold           = %[1]q
  alarm_description   = "This metric monitors ec2 cpu utilization"
}

data "aws_region" "current" {}

resource "aws_route53_health_check" "test" {
  type                            = "CLOUDWATCH_METRIC"
  cloudwatch_alarm_name           = aws_cloudwatch_metric_alarm.test.alarm_name
  cloudwatch_alarm_region         = data.aws_region.current.region
  insufficient_data_health_status = "Healthy"

  triggers = {
    threshold = aws_cloudwatch_metric_alarm.test.threshold
  }
}
`, threshold)
}

func testAccHealthCheckConfig_searchString(search string, invert bool) string {
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

const testAccHealthCheckConfig_noSNI = `
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

func testAccHealthCheckConfig_sni(enable bool) string {
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

func testAccHealthCheckConfig_disabled(disabled bool) string {
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

func testAccHealthCheckConfig_routingControlARN(rName string) string {
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
