package route53_test

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
)

func TestAccRoute53QueryLog_basic(t *testing.T) {
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"
	resourceName := "aws_route53_query_log.test"
	route53ZoneResourceName := "aws_route53_zone.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	var queryLoggingConfig route53.QueryLoggingConfig
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckQueryLog(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueryLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryLogConfig_resourceBasic1(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryLogExists(resourceName, &queryLoggingConfig),
					acctest.MatchResourceAttrGlobalARNNoAccount(resourceName, "arn", "route53", regexp.MustCompile("queryloggingconfig/.+")),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_log_group_arn", cloudwatchLogGroupResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "zone_id", route53ZoneResourceName, "zone_id"),
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

func TestAccRoute53QueryLog_disappears(t *testing.T) {
	resourceName := "aws_route53_query_log.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	var queryLoggingConfig route53.QueryLoggingConfig
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckQueryLog(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueryLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryLogConfig_resourceBasic1(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryLogExists(resourceName, &queryLoggingConfig),
					acctest.CheckResourceDisappears(acctest.Provider, tfroute53.ResourceQueryLog(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53QueryLog_Disappears_hostedZone(t *testing.T) {
	resourceName := "aws_route53_query_log.test"
	route53ZoneResourceName := "aws_route53_zone.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	var queryLoggingConfig route53.QueryLoggingConfig
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckQueryLog(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueryLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryLogConfig_resourceBasic1(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryLogExists(resourceName, &queryLoggingConfig),
					acctest.CheckResourceDisappears(acctest.Provider, tfroute53.ResourceZone(), route53ZoneResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckQueryLogExists(pr string, queryLoggingConfig *route53.QueryLoggingConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProviderRoute53QueryLog.Meta().(*conns.AWSClient).Route53Conn
		rs, ok := s.RootModule().Resources[pr]
		if !ok {
			return fmt.Errorf("Not found: %s", pr)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		out, err := conn.GetQueryLoggingConfig(&route53.GetQueryLoggingConfigInput{
			Id: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}
		if out.QueryLoggingConfig == nil {
			return fmt.Errorf("Route53 query logging configuration does not exist: %q", rs.Primary.ID)
		}

		*queryLoggingConfig = *out.QueryLoggingConfig

		return nil
	}
}

func testAccCheckQueryLogDestroy(s *terraform.State) error {
	conn := testAccProviderRoute53QueryLog.Meta().(*conns.AWSClient).Route53Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_query_log" {
			continue
		}

		out, err := conn.GetQueryLoggingConfig(&route53.GetQueryLoggingConfigInput{
			Id: aws.String(rs.Primary.ID),
		})

		if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchQueryLoggingConfig) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Route 53 Query Logging Configuration (%s): %w", rs.Primary.ID, err)
		}

		if out.QueryLoggingConfig != nil {
			return fmt.Errorf("Route53 query logging configuration exists: %q", rs.Primary.ID)
		}
	}

	return nil
}

func testAccQueryLogConfig_resourceBasic1(rName, domainName string) string {
	return acctest.ConfigCompose(
		testAccQueryLogRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name              = "/aws/route53/${aws_route53_zone.test.name}"
  retention_in_days = 1
}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:${data.aws_partition.current.partition}:logs:*:*:log-group:/aws/route53/*"]

    principals {
      identifiers = ["route53.${data.aws_partition.current.dns_suffix}"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name     = %[1]q
  policy_document = data.aws_iam_policy_document.test.json
}

resource "aws_route53_zone" "test" {
  name = %[2]q
}

resource "aws_route53_query_log" "test" {
  depends_on = [aws_cloudwatch_log_resource_policy.test]

  cloudwatch_log_group_arn = aws_cloudwatch_log_group.test.arn
  zone_id                  = aws_route53_zone.test.zone_id
}
`, rName, domainName))
}

// Route 53 Query Logging can only be enabled with CloudWatch Log Groups in specific regions,

// testAccRoute53QueryLogRegion is the chosen Route 53 Query Logging testing region
//
// Cached to prevent issues should multiple regions become available.
var testAccRoute53QueryLogRegion string

// testAccProviderRoute53QueryLog is the Route 53 Query Logging provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// testAccPreCheckQueryLog(t) must be called before using this provider instance.
var testAccProviderRoute53QueryLog *schema.Provider

// testAccProviderRoute53QueryLogConfigure ensures the provider is only configured once
var testAccProviderRoute53QueryLogConfigure sync.Once

// testAccPreCheckQueryLog verifies AWS credentials and that Route 53 Query Logging is supported
func testAccPreCheckQueryLog(t *testing.T) {
	acctest.PreCheckPartitionHasService(route53.EndpointsID, t)

	region := testAccGetQueryLogRegion()

	if region == "" {
		t.Skip("Route 53 Query Log not available in this AWS Partition")
	}

	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderRoute53QueryLogConfigure.Do(func() {
		testAccProviderRoute53QueryLog = provider.Provider()

		testAccRecordConfig_config := map[string]interface{}{
			"region": region,
		}

		diags := testAccProviderRoute53QueryLog.Configure(context.Background(), terraform.NewResourceConfigRaw(testAccRecordConfig_config))

		if diags != nil && diags.HasError() {
			for _, d := range diags {
				if d.Severity == diag.Error {
					t.Fatalf("error configuring Route 53 Query Logging provider: %s", d.Summary)
				}
			}
		}
	})
}

// testAccQueryLogRegionProviderConfig is the Terraform provider configuration for Route 53 Query Logging region testing
//
// Testing Route 53 Query Logging assumes no other provider configurations
// are necessary and overwrites the "aws" provider configuration.
func testAccQueryLogRegionProviderConfig() string {
	return acctest.ConfigRegionalProvider(testAccGetQueryLogRegion())
}

// testAccGetQueryLogRegion returns the Route 53 Query Logging region for testing
func testAccGetQueryLogRegion() string {
	if testAccRoute53QueryLogRegion != "" {
		return testAccRoute53QueryLogRegion
	}

	// AWS Commercial: https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/query-logs.html
	// AWS GovCloud (US) - only private DNS: https://docs.aws.amazon.com/govcloud-us/latest/UserGuide/govcloud-r53.html
	// AWS China - not available yet: https://docs.amazonaws.cn/en_us/aws/latest/userguide/route53.html
	switch acctest.Partition() {
	case endpoints.AwsPartitionID:
		testAccRoute53QueryLogRegion = endpoints.UsEast1RegionID
	}

	return testAccRoute53QueryLogRegion
}
