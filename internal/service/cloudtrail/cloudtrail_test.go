package cloudtrail_test

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudtrail "github.com/hashicorp/terraform-provider-aws/internal/service/cloudtrail"
)

func TestAccCloudTrail_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Trail": {
			"basic":                 testAcc_basic,
			"cloudwatch":            testAcc_cloudWatch,
			"enableLogging":         testAcc_enableLogging,
			"globalServiceEvents":   testAcc_globalServiceEvents,
			"multiRegion":           testAcc_multiRegion,
			"organization":          testAcc_organization,
			"logValidation":         testAcc_logValidation,
			"kmsKey":                testAcc_kmsKey,
			"tags":                  testAcc_tags,
			"eventSelector":         testAcc_eventSelector,
			"eventSelectorDynamoDB": testAcc_eventSelectorDynamoDB,
			"eventSelectorExclude":  testAcc_eventSelectorExclude,
			"insightSelector":       testAcc_insightSelector,
			"advancedEventSelector": testAcc_advanced_event_selector,
			"disappears":            testAcc_disappears,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAcc_basic(t *testing.T) {
	var trail cloudtrail.Trail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudtrail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "cloudtrail", fmt.Sprintf("trail/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", "true"),
					resource.TestCheckResourceAttr(resourceName, "is_organization_trail", "false"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccModifiedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "s3_key_prefix", "prefix"),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", "false"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
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

func testAcc_cloudWatch(t *testing.T) {
	var trail cloudtrail.Trail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudtrail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudWatchConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttrSet(resourceName, "cloud_watch_logs_group_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "cloud_watch_logs_role_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCloudWatchModifiedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttrSet(resourceName, "cloud_watch_logs_group_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "cloud_watch_logs_role_arn"),
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

func testAcc_enableLogging(t *testing.T) {
	var trail cloudtrail.Trail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudtrail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnableLoggingConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					// AWS will create the trail with logging turned off.
					// Test that "enable_logging" default works.
					testAccCheckCloudTrailLoggingEnabled(resourceName, true),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEnableLoggingConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					testAccCheckCloudTrailLoggingEnabled(resourceName, false),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					testAccCheckCloudTrailLoggingEnabled(resourceName, true),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
		},
	})
}

func testAcc_multiRegion(t *testing.T) {
	var trail cloudtrail.Trail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudtrail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "is_multi_region_trail", "false"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
			{
				Config: testAccMultiRegionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "is_multi_region_trail", "true"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "is_multi_region_trail", "false"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
		},
	})
}

func testAcc_organization(t *testing.T) {
	var trail cloudtrail.Trail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudtrail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "is_organization_trail", "true"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "is_organization_trail", "false"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
		},
	})
}

func testAcc_logValidation(t *testing.T) {
	var trail cloudtrail.Trail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudtrail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLogValidationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", "true"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, true, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLogValidationModifiedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", "true"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
		},
	})
}

func testAcc_kmsKey(t *testing.T) {
	var trail cloudtrail.Trail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_cloudtrail.test"
	kmsResourceName := "aws_kms_key.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudtrail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKMSKeyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "s3_key_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", "true"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsResourceName, "arn"),
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

func testAcc_tags(t *testing.T) {
	var trail cloudtrail.Trail
	var trailTags []*cloudtrail.Tag
	var trailTagsModified []*cloudtrail.Tag
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudtrail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTagsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					testAccCheckCloudTrailLoadTags(&trail, &trailTags),
					resource.TestCheckResourceAttr(resourceName, "tags.Yak", "milk"),
					resource.TestCheckResourceAttr(resourceName, "tags.Fox", "tail"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTagsModifiedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					testAccCheckCloudTrailLoadTags(&trail, &trailTagsModified),
					resource.TestCheckResourceAttr(resourceName, "tags.Yak", "milk"),
					resource.TestCheckResourceAttr(resourceName, "tags.Emu", "toes"),
					resource.TestCheckResourceAttr(resourceName, "tags.Fox", "tail"),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
			{
				Config: testAccTagsModifiedAgainConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccCheckCloudTrailLoadTags(&trail, &trailTagsModified),
					testAccCheckCloudTrailLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
				),
			},
		},
	})
}

func testAcc_globalServiceEvents(t *testing.T) {
	var trail cloudtrail.Trail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudtrail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalServiceEventsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", "false"),
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

func testAcc_eventSelector(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudtrail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSelectorConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.type", "AWS::S3::Object"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.values.#", "2"),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.0.data_resource.0.values.0", "s3", fmt.Sprintf("%s-2/isen", rName)),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.0.data_resource.0.values.1", "s3", fmt.Sprintf("%s-2/ko", rName)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.include_management_events", "false"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.read_write_type", "ReadOnly"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.exclude_management_event_sources.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventSelectorReadWriteTypeConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.include_management_events", "true"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.read_write_type", "WriteOnly"),
				),
			},
			{
				Config: testAccEventSelectorModifiedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.type", "AWS::S3::Object"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.values.#", "2"),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.0.data_resource.0.values.0", "s3", fmt.Sprintf("%s-2/isen", rName)),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.0.data_resource.0.values.1", "s3", fmt.Sprintf("%s-2/ko", rName)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.include_management_events", "true"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.read_write_type", "ReadOnly"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.0.type", "AWS::S3::Object"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.0.values.#", "2"),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.1.data_resource.0.values.0", "s3", fmt.Sprintf("%s-2/tf1", rName)),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.1.data_resource.0.values.1", "s3", fmt.Sprintf("%s-2/tf2", rName)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.1.type", "AWS::Lambda::Function"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.1.values.#", "1"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "event_selector.1.data_resource.1.values.0", "lambda", fmt.Sprintf("function:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.exclude_management_event_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.include_management_events", "false"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.read_write_type", "All"),
				),
			},
			{
				Config: testAccEventSelectorNoneConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", "0"),
				),
			},
		},
	})
}

func testAcc_eventSelectorDynamoDB(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudtrail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSelectorDynamoDBConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.type", "AWS::DynamoDB::Table"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.values.#", "1"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "event_selector.0.data_resource.0.values.0", "dynamodb", regexp.MustCompile(`table/tf-acc-test-.+`)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.include_management_events", "true"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.read_write_type", "All"),
				),
			},
		},
	})
}

func testAcc_eventSelectorExclude(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudtrail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEventSelectorExcludeKMSConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.include_management_events", "true"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.exclude_management_event_sources.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_selector.0.exclude_management_event_sources.*", "kms.amazonaws.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventSelectorExcludeKMSAndRDSDataConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.include_management_events", "true"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.exclude_management_event_sources.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_selector.0.exclude_management_event_sources.*", "kms.amazonaws.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_selector.0.exclude_management_event_sources.*", "rdsdata.amazonaws.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventSelectorNoneConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", "0"),
				),
			},
		},
	})
}

func testAcc_insightSelector(t *testing.T) {
	resourceName := "aws_cloudtrail.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudtrail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInsightSelectorConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "insight_selector.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "insight_selector.0.insight_type", "ApiCallRateInsight"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInsightSelectorMultiConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "insight_selector.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "insight_selector.0.insight_type", "ApiCallRateInsight"),
					resource.TestCheckResourceAttr(resourceName, "insight_selector.1.insight_type", "ApiErrorRateInsight"),
				),
			},
			{
				Config: testAccInsightSelectorConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "insight_selector.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "insight_selector.0.insight_type", "ApiCallRateInsight"),
				),
			},
			{
				Config: testAccConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "insight_selector.#", "0"),
				),
			},
		},
	})
}

func testAcc_advanced_event_selector(t *testing.T) {
	resourceName := "aws_cloudtrail.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudtrail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfig_advancedEventSelector(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.#", "5"),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.0.name", "s3Custom"),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.0.field_selector.#", "5"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.0.field_selector.*", map[string]string{
						"field":    "eventCategory",
						"equals.#": "1",
						"equals.0": "Data",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.0.field_selector.*", map[string]string{
						"field":    "eventName",
						"equals.#": "1",
						"equals.0": "DeleteObject",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.0.field_selector.*", map[string]string{
						"field":    "resources.ARN",
						"equals.#": "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.0.field_selector.*", map[string]string{
						"field":    "readOnly",
						"equals.#": "1",
						"equals.0": "false",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.0.field_selector.*", map[string]string{
						"field":    "resources.type",
						"equals.#": "1",
						"equals.0": "AWS::S3::Object",
					}),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.1.name", "lambdaLogAllEvents"),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.1.field_selector.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.1.field_selector.*", map[string]string{
						"field":    "eventCategory",
						"equals.#": "1",
						"equals.0": "Data",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.1.field_selector.*", map[string]string{
						"field":    "resources.type",
						"equals.#": "1",
						"equals.0": "AWS::Lambda::Function",
					}),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.2.name", "dynamoDbReadOnlyEvents"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.2.field_selector.*", map[string]string{
						"field":    "readOnly",
						"equals.#": "1",
						"equals.0": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.2.field_selector.*", map[string]string{
						"field":    "resources.type",
						"equals.#": "1",
						"equals.0": "AWS::DynamoDB::Table",
					}),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.3.name", "s3OutpostsWriteOnlyEvents"),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.3.field_selector.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.3.field_selector.*", map[string]string{
						"field":    "eventCategory",
						"equals.#": "1",
						"equals.0": "Data",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.3.field_selector.*", map[string]string{
						"field":    "readOnly",
						"equals.#": "1",
						"equals.0": "false",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.3.field_selector.*", map[string]string{
						"field":    "resources.type",
						"equals.#": "1",
						"equals.0": "AWS::S3Outposts::Object",
					}),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.4.name", "managementEventsSelector"),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.4.field_selector.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.4.field_selector.*", map[string]string{
						"field":    "eventCategory",
						"equals.#": "1",
						"equals.0": "Management",
					}),
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

func testAcc_disappears(t *testing.T) {
	var trail cloudtrail.Trail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudtrail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists(resourceName, &trail),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudtrail.ResourceCloudTrail(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudtrail.ResourceCloudTrail(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCloudTrailExists(n string, trail *cloudtrail.Trail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudTrailConn
		params := cloudtrail.DescribeTrailsInput{
			TrailNameList: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeTrails(&params)
		if err != nil {
			return err
		}
		if len(resp.TrailList) == 0 {
			return fmt.Errorf("Trail not found")
		}
		*trail = *resp.TrailList[0]

		return nil
	}
}

func testAccCheckCloudTrailLoggingEnabled(n string, desired bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudTrailConn
		params := cloudtrail.GetTrailStatusInput{
			Name: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetTrailStatus(&params)

		if err != nil {
			return err
		}

		isLog := aws.BoolValue(resp.IsLogging)
		if isLog != desired {
			return fmt.Errorf("Expected logging status %t, given %t", desired, isLog)
		}

		return nil
	}
}

func testAccCheckCloudTrailLogValidationEnabled(n string, desired bool, trail *cloudtrail.Trail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if trail.LogFileValidationEnabled == nil {
			return fmt.Errorf("No LogFileValidationEnabled attribute present in trail: %s", trail)
		}

		logValid := aws.BoolValue(trail.LogFileValidationEnabled)
		if logValid != desired {
			return fmt.Errorf("Expected log validation status %t, given %t", desired, logValid)
		}

		// local state comparison
		enabled, ok := rs.Primary.Attributes["enable_log_file_validation"]
		if !ok {
			return fmt.Errorf("No enable_log_file_validation attribute defined for %s, expected %t",
				n, desired)
		}
		desiredInString := fmt.Sprintf("%t", desired)
		if enabled != desiredInString {
			return fmt.Errorf("Expected log validation status %s, saved %s", desiredInString, enabled)
		}

		return nil
	}
}

func testAccCheckDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudTrailConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudtrail" {
			continue
		}

		params := cloudtrail.DescribeTrailsInput{
			TrailNameList: []*string{aws.String(rs.Primary.ID)},
		}

		resp, err := conn.DescribeTrails(&params)

		if err == nil {
			if len(resp.TrailList) != 0 &&
				aws.StringValue(resp.TrailList[0].Name) == rs.Primary.ID {
				return fmt.Errorf("CloudTrail still exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckCloudTrailLoadTags(trail *cloudtrail.Trail, tags *[]*cloudtrail.Tag) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudTrailConn
		input := cloudtrail.ListTagsInput{
			ResourceIdList: []*string{trail.TrailARN},
		}
		out, err := conn.ListTags(&input)
		if err != nil {
			return err
		}
		log.Printf("[DEBUG] Received CloudTrail tags during test: %s", out)
		if len(out.ResourceTagList) > 0 {
			*tags = out.ResourceTagList[0].TagsList
		}
		log.Printf("[DEBUG] Loading CloudTrail tags into a var: %s", *tags)
		return nil
	}
}

func testAccBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AWSCloudTrailAclCheck"
        Effect    = "Allow"
        Principal = "*"
        Action    = "s3:GetBucketAcl"
        Resource  = "arn:${data.aws_partition.current.partition}:s3:::%[1]s"
      },
      {
        Sid       = "AWSCloudTrailWrite"
        Effect    = "Allow"
        Principal = "*"
        Action    = "s3:PutObject"
        Resource  = "arn:${data.aws_partition.current.partition}:s3:::%[1]s/*"
        Condition = {
          StringEquals = {
            "s3:x-amz-acl" = "bucket-owner-full-control"
          }
        }
      }
    ]
  })
}
`, rName)
}

func testAccConfig(rName string) string {
	return acctest.ConfigCompose(testAccBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id
}
`, rName))
}

func testAccModifiedConfig(rName string) string {
	return acctest.ConfigCompose(testAccBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name                          = %[1]q
  s3_bucket_name                = aws_s3_bucket.test.id
  s3_key_prefix                 = "prefix"
  include_global_service_events = false
}
`, rName))
}

func testAccEnableLoggingConfig(rName string, enableLogging bool) string {
	return acctest.ConfigCompose(testAccBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name                          = %[1]q
  s3_bucket_name                = aws_s3_bucket.test.id
  s3_key_prefix                 = "prefix"
  include_global_service_events = false
  enable_logging                = %[2]t
}
`, rName, enableLogging))
}

func testAccCloudWatchConfig(rName string) string {
	return acctest.ConfigCompose(testAccBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id

  cloud_watch_logs_group_arn = "${aws_cloudwatch_log_group.test.arn}:*"
  cloud_watch_logs_role_arn  = aws_iam_role.test.arn
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"

    Statement = [{
      Sid    = ""
      Effect = "Allow"
      Action = "sts:AssumeRole"

      Principal = {
        Service = "cloudtrail.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"

    Statement = [{
      Sid      = "AWSCloudTrailCreateLogStream"
      Effect   = "Allow"
      Resource = "${aws_cloudwatch_log_group.test.arn}:*"

      Action = [
        "logs:CreateLogStream",
        "logs:PutLogEvents",
      ]
    }]
  })
}
`, rName))
}

func testAccCloudWatchModifiedConfig(rName string) string {
	return acctest.ConfigCompose(testAccBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id

  cloud_watch_logs_group_arn = "${aws_cloudwatch_log_group.test2.arn}:*"
  cloud_watch_logs_role_arn  = aws_iam_role.test.arn
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_group" "test2" {
  name = "%[1]s-2"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"

    Statement = [{
      Sid    = ""
      Effect = "Allow"
      Action = "sts:AssumeRole"

      Principal = {
        Service = "cloudtrail.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"

    Statement = [{
      Sid      = "AWSCloudTrailCreateLogStream"
      Effect   = "Allow"
      Resource = "${aws_cloudwatch_log_group.test2.arn}:*"

      Action = [
        "logs:CreateLogStream",
        "logs:PutLogEvents",
      ]
    }]
  })
}
`, rName))
}

func testAccMultiRegionConfig(rName string) string {
	return acctest.ConfigCompose(testAccBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name                  = %[1]q
  s3_bucket_name        = aws_s3_bucket.test.id
  is_multi_region_trail = true
}
`, rName))
}

func testAccOrganizationConfig(rName string) string {
	return acctest.ConfigCompose(testAccBaseConfig(rName), fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["cloudtrail.${data.aws_partition.current.dns_suffix}"]
}

resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  is_organization_trail = true
  name                  = %[1]q
  s3_bucket_name        = aws_s3_bucket.test.id
}
`, rName))
}

func testAccLogValidationConfig(rName string) string {
	return acctest.ConfigCompose(testAccBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name                          = %[1]q
  s3_bucket_name                = aws_s3_bucket.test.id
  is_multi_region_trail         = true
  include_global_service_events = true
  enable_log_file_validation    = true
}
`, rName))
}

func testAccLogValidationModifiedConfig(rName string) string {
	return acctest.ConfigCompose(testAccBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name                          = %[1]q
  s3_bucket_name                = aws_s3_bucket.test.id
  include_global_service_events = true
}
`, rName))
}

func testAccKMSKeyConfig(rName string) string {
	return acctest.ConfigCompose(testAccBaseConfig(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = %[1]q

    Statement = [{
      Sid      = "Enable IAM User Permissions"
      Effect   = "Allow"
      Action   = "kms:*"
      Resource = "*"

      Principal = {
        AWS = "*"
      }
    }]
  })
}

resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name                          = %[1]q
  s3_bucket_name                = aws_s3_bucket.test.id
  include_global_service_events = true
  kms_key_id                    = aws_kms_key.test.arn
}
`, rName))
}

func testAccGlobalServiceEventsConfig(rName string) string {
	return acctest.ConfigCompose(testAccBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name                          = %[1]q
  s3_bucket_name                = aws_s3_bucket.test.id
  include_global_service_events = false
}
`, rName))
}

func testAccTagsConfig(rName string) string {
	return acctest.ConfigCompose(testAccBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id

  tags = {
    Yak = "milk"
    Fox = "tail"
  }
}
`, rName))
}

func testAccTagsModifiedConfig(rName string) string {
	return acctest.ConfigCompose(testAccBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id

  tags = {
    Yak = "milk"
    Fox = "tail"
    Emu = "toes"
  }
}
`, rName))
}

func testAccTagsModifiedAgainConfig(rName string) string {
	return acctest.ConfigCompose(testAccBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id
}
`, rName))
}

func testAccEventSelectorConfig(rName string) string {
	return acctest.ConfigCompose(testAccBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id

  event_selector {
    read_write_type           = "ReadOnly"
    include_management_events = false

    data_resource {
      type = "AWS::S3::Object"

      values = [
        "${aws_s3_bucket.test2.arn}/isen",
        "${aws_s3_bucket.test2.arn}/ko",
      ]
    }
  }
}

resource "aws_s3_bucket" "test2" {
  bucket        = "%[1]s-2"
  force_destroy = true
}
`, rName))
}

func testAccEventSelectorReadWriteTypeConfig(rName string) string {
	return acctest.ConfigCompose(testAccBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id

  event_selector {
    read_write_type           = "WriteOnly"
    include_management_events = true
  }
}
`, rName))
}

func testAccEventSelectorModifiedConfig(rName string) string {
	return acctest.ConfigCompose(testAccBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id

  event_selector {
    read_write_type           = "ReadOnly"
    include_management_events = true

    data_resource {
      type = "AWS::S3::Object"

      values = [
        "${aws_s3_bucket.test2.arn}/isen",
        "${aws_s3_bucket.test2.arn}/ko",
      ]
    }
  }

  event_selector {
    read_write_type           = "All"
    include_management_events = false

    data_resource {
      type = "AWS::S3::Object"

      values = [
        "${aws_s3_bucket.test2.arn}/tf1",
        "${aws_s3_bucket.test2.arn}/tf2",
      ]
    }

    data_resource {
      type = "AWS::Lambda::Function"

      values = [
        aws_lambda_function.test.arn,
      ]
    }
  }
}

resource "aws_s3_bucket" "test2" {
  bucket        = "%[1]s-2"
  force_destroy = true
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"

    Statement = [{
      Sid    = ""
      Effect = "Allow"
      Action = "sts:AssumeRole"

      Principal = {
        Service = "lambda.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}
`, rName))
}

func testAccEventSelectorNoneConfig(rName string) string {
	return acctest.ConfigCompose(testAccBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id
}
`, rName))
}

func testAccEventSelectorDynamoDBConfig(rName string) string {
	return acctest.ConfigCompose(testAccBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id

  event_selector {
    read_write_type           = "All"
    include_management_events = true

    data_resource {
      type = "AWS::DynamoDB::Table"

      values = [
        aws_dynamodb_table.test.arn,
      ]
    }
  }
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1
  hash_key       = %[1]q

  attribute {
    name = %[1]q
    type = "S"
  }
}
`, rName))
}

func testAccEventSelectorExcludeKMSConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id

  event_selector {
    exclude_management_event_sources = ["kms.${data.aws_partition.current.dns_suffix}"]
  }
}
`, rName))
}

func testAccEventSelectorExcludeKMSAndRDSDataConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id

  event_selector {
    exclude_management_event_sources = [
      "kms.${data.aws_partition.current.dns_suffix}",
      "rdsdata.${data.aws_partition.current.dns_suffix}"
    ]
  }
}
`, rName))
}

func testAccInsightSelectorConfig(rName string) string {
	return acctest.ConfigCompose(testAccBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id


  insight_selector {
    insight_type = "ApiCallRateInsight"
  }
}
`, rName))
}

func testAccInsightSelectorMultiConfig(rName string) string {
	return acctest.ConfigCompose(testAccBaseConfig(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id


  insight_selector {
    insight_type = "ApiCallRateInsight"
  }

  insight_selector {
    insight_type = "ApiErrorRateInsight"
  }
}
`, rName))
}

func testAccConfig_advancedEventSelector(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test1.id

  advanced_event_selector {
    name = "s3Custom"
    field_selector {
      field  = "eventCategory"
      equals = ["Data"]
    }

    field_selector {
      field  = "eventName"
      equals = ["DeleteObject"]
    }

    field_selector {
      field = "resources.ARN"
      equals = [
        "${aws_s3_bucket.test2.arn}/foobar",
        "${aws_s3_bucket.test2.arn}/bar",
      ]
    }

    field_selector {
      field  = "readOnly"
      equals = ["false"]
    }

    field_selector {
      field  = "resources.type"
      equals = ["AWS::S3::Object"]
    }
  }
  advanced_event_selector {
    name = "lambdaLogAllEvents"
    field_selector {
      field  = "eventCategory"
      equals = ["Data"]
    }

    field_selector {
      field  = "resources.type"
      equals = ["AWS::Lambda::Function"]
    }
  }

  advanced_event_selector {
    name = "dynamoDbReadOnlyEvents"
    field_selector {
      field  = "eventCategory"
      equals = ["Data"]
    }

    field_selector {
      field  = "readOnly"
      equals = ["true"]
    }

    field_selector {
      field  = "resources.type"
      equals = ["AWS::DynamoDB::Table"]
    }
  }

  advanced_event_selector {
    name = "s3OutpostsWriteOnlyEvents"
    field_selector {
      field  = "eventCategory"
      equals = ["Data"]
    }

    field_selector {
      field  = "readOnly"
      equals = ["false"]
    }

    field_selector {
      field  = "resources.type"
      equals = ["AWS::S3Outposts::Object"]
    }
  }

  advanced_event_selector {
    name = "managementEventsSelector"
    field_selector {
      field  = "eventCategory"
      equals = ["Management"]
    }
  }
}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "test1" {
  bucket        = "%[1]s-1"
  force_destroy = true
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test1.id
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailAclCheck",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::%[1]s-1"
    },
    {
      "Sid": "AWSCloudTrailWrite",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::%[1]s-1/*",
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "test2" {
  bucket        = "%[1]s-2"
  force_destroy = true
}
`, rName)
}
