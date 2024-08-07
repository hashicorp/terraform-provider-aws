// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudtrail_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudtrail "github.com/hashicorp/terraform-provider-aws/internal/service/cloudtrail"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.CloudTrailServiceID, testAccErrorCheckSkip)
}

// testAccErrorCheckSkip skips CloudTrail tests that have error messages indicating unsupported features
func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"AccessDeniedException:",
	)
}

func TestAccCloudTrail_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Trail": {
			acctest.CtBasic:         testAccTrail_basic,
			"cloudwatch":            testAccTrail_cloudWatch,
			"enableLogging":         testAccTrail_enableLogging,
			"globalServiceEvents":   testAccTrail_globalServiceEvents,
			"multiRegion":           testAccTrail_multiRegion,
			"organization":          testAccTrail_organization,
			"logValidation":         testAccTrail_logValidation,
			"kmsKey":                testAccTrail_kmsKey,
			"tags":                  testAccTrail_tags,
			"eventSelector":         testAccTrail_eventSelector,
			"eventSelectorDynamoDB": testAccTrail_eventSelectorDynamoDB,
			"eventSelectorExclude":  testAccTrail_eventSelectorExclude,
			"insightSelector":       testAccTrail_insightSelector,
			"advancedEventSelector": testAccTrail_advancedEventSelector,
			acctest.CtDisappears:    testAccTrail_disappears,
			"migrateV0":             testAccTrail_migrateV0,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccTrail_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var trail types.Trail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrailDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudTrailConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "cloudtrail", fmt.Sprintf("trail/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "is_organization_trail", acctest.CtFalse),
					testAccCheckLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCloudTrailConfig_modified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, names.AttrS3KeyPrefix, names.AttrPrefix),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", acctest.CtFalse),
					testAccCheckLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
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

func testAccTrail_cloudWatch(t *testing.T) {
	ctx := acctest.Context(t)
	var trail types.Trail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrailDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudTrailConfig_cloudWatch(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
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
				Config: testAccCloudTrailConfig_cloudWatchModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
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

func testAccTrail_enableLogging(t *testing.T) {
	ctx := acctest.Context(t)
	var trail types.Trail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrailDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudTrailConfig_enableLogging(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
					// AWS will create the trail with logging turned off.
					// Test that "enable_logging" default works.
					testAccCheckLoggingEnabled(ctx, resourceName, true),
					testAccCheckLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCloudTrailConfig_enableLogging(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
					testAccCheckLoggingEnabled(ctx, resourceName, false),
					testAccCheckLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCloudTrailConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
					testAccCheckLoggingEnabled(ctx, resourceName, true),
					testAccCheckLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
				),
			},
		},
	})
}

func testAccTrail_multiRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var trail types.Trail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrailDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudTrailConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "is_multi_region_trail", acctest.CtFalse),
					testAccCheckLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
				),
			},
			{
				Config: testAccCloudTrailConfig_multiRegion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "is_multi_region_trail", acctest.CtTrue),
					testAccCheckLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCloudTrailConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "is_multi_region_trail", acctest.CtFalse),
					testAccCheckLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
				),
			},
		},
	})
}

func testAccTrail_organization(t *testing.T) {
	ctx := acctest.Context(t)
	var trail types.Trail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationManagementAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrailDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudTrailConfig_organization(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "is_organization_trail", acctest.CtTrue),
					testAccCheckLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCloudTrailConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "is_organization_trail", acctest.CtFalse),
					testAccCheckLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
				),
			},
		},
	})
}

func testAccTrail_logValidation(t *testing.T) {
	ctx := acctest.Context(t)
	var trail types.Trail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrailDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudTrailConfig_logValidation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, names.AttrS3KeyPrefix, ""),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", acctest.CtTrue),
					testAccCheckLogValidationEnabled(resourceName, true, &trail),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCloudTrailConfig_logValidationModified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, names.AttrS3KeyPrefix, ""),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", acctest.CtTrue),
					testAccCheckLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
				),
			},
		},
	})
}

func testAccTrail_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	var trail types.Trail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_cloudtrail.test"
	kmsResourceName := "aws_kms_key.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrailDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudTrailConfig_kmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, names.AttrS3KeyPrefix, ""),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", acctest.CtTrue),
					testAccCheckLogValidationEnabled(resourceName, false, &trail),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsResourceName, names.AttrARN),
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

func testAccTrail_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var trail types.Trail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrailDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudTrailConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCloudTrailConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCloudTrailConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccTrail_globalServiceEvents(t *testing.T) {
	ctx := acctest.Context(t)
	var trail types.Trail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrailDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudTrailConfig_globalServiceEvents(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
					resource.TestCheckResourceAttr(resourceName, "include_global_service_events", acctest.CtFalse),
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

func testAccTrail_eventSelector(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrailDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudTrailConfig_eventSelector(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.type", "AWS::S3::Object"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.values.#", acctest.Ct2),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.0.data_resource.0.values.0", "s3", fmt.Sprintf("%s-2/isen", rName)),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.0.data_resource.0.values.1", "s3", fmt.Sprintf("%s-2/ko", rName)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.include_management_events", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.read_write_type", "ReadOnly"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.exclude_management_event_sources.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCloudTrailConfig_eventSelectorReadWriteType(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.include_management_events", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.read_write_type", "WriteOnly"),
				),
			},
			{
				Config: testAccCloudTrailConfig_eventSelectorModified(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.type", "AWS::S3::Object"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.values.#", acctest.Ct2),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.0.data_resource.0.values.0", "s3", fmt.Sprintf("%s-2/isen", rName)),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.0.data_resource.0.values.1", "s3", fmt.Sprintf("%s-2/ko", rName)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.include_management_events", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.read_write_type", "ReadOnly"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.0.type", "AWS::S3::Object"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.0.values.#", acctest.Ct2),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.1.data_resource.0.values.0", "s3", fmt.Sprintf("%s-2/tf1", rName)),
					acctest.CheckResourceAttrGlobalARNNoAccount(resourceName, "event_selector.1.data_resource.0.values.1", "s3", fmt.Sprintf("%s-2/tf2", rName)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.1.type", "AWS::Lambda::Function"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.data_resource.1.values.#", acctest.Ct1),
					acctest.CheckResourceAttrRegionalARN(resourceName, "event_selector.1.data_resource.1.values.0", "lambda", fmt.Sprintf("function:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.exclude_management_event_sources.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.include_management_events", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "event_selector.1.read_write_type", "All"),
				),
			},
			{
				Config: testAccCloudTrailConfig_eventSelectorNone(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccTrail_eventSelectorDynamoDB(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrailDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudTrailConfig_eventSelectorDynamoDB(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.type", "AWS::DynamoDB::Table"),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.data_resource.0.values.#", acctest.Ct1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "event_selector.0.data_resource.0.values.0", "dynamodb", regexache.MustCompile(`table/tf-acc-test-.+`)),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.include_management_events", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.read_write_type", "All"),
				),
			},
		},
	})
}

func testAccTrail_eventSelectorExclude(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrailDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudTrailConfig_eventSelectorExcludeKMS(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.include_management_events", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.exclude_management_event_sources.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_selector.0.exclude_management_event_sources.*", "kms.amazonaws.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCloudTrailConfig_eventSelectorExcludeKMSAndRDSData(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.include_management_events", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "event_selector.0.exclude_management_event_sources.#", acctest.Ct2),
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
				Config: testAccCloudTrailConfig_eventSelectorNone(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_selector.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccTrail_insightSelector(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudtrail.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrailDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudTrailConfig_insightSelector(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "insight_selector.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "insight_selector.0.insight_type", "ApiCallRateInsight"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCloudTrailConfig_insightSelectorMulti(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "insight_selector.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "insight_selector.*", map[string]string{
						"insight_type": "ApiCallRateInsight",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "insight_selector.*", map[string]string{
						"insight_type": "ApiErrorRateInsight",
					}),
				),
			},
			{
				Config: testAccCloudTrailConfig_insightSelector(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "insight_selector.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "insight_selector.0.insight_type", "ApiCallRateInsight"),
				),
			},
			{
				Config: testAccCloudTrailConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "insight_selector.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccTrail_advancedEventSelector(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudtrail.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrailDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudTrailConfig_advancedEventSelector(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.#", "5"),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.0.name", "s3Custom"),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.0.field_selector.#", "5"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.0.field_selector.*", map[string]string{
						names.AttrField: "eventCategory",
						"equals.#":      acctest.Ct1,
						"equals.0":      "Data",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.0.field_selector.*", map[string]string{
						names.AttrField: "eventName",
						"equals.#":      acctest.Ct1,
						"equals.0":      "DeleteObject",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.0.field_selector.*", map[string]string{
						names.AttrField: "resources.ARN",
						"equals.#":      acctest.Ct2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.0.field_selector.*", map[string]string{
						names.AttrField: "readOnly",
						"equals.#":      acctest.Ct1,
						"equals.0":      acctest.CtFalse,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.0.field_selector.*", map[string]string{
						names.AttrField: "resources.type",
						"equals.#":      acctest.Ct1,
						"equals.0":      "AWS::S3::Object",
					}),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.1.name", "lambdaLogAllEvents"),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.1.field_selector.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.1.field_selector.*", map[string]string{
						names.AttrField: "eventCategory",
						"equals.#":      acctest.Ct1,
						"equals.0":      "Data",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.1.field_selector.*", map[string]string{
						names.AttrField: "resources.type",
						"equals.#":      acctest.Ct1,
						"equals.0":      "AWS::Lambda::Function",
					}),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.2.name", "dynamoDbReadOnlyEvents"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.2.field_selector.*", map[string]string{
						names.AttrField: "readOnly",
						"equals.#":      acctest.Ct1,
						"equals.0":      acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.2.field_selector.*", map[string]string{
						names.AttrField: "resources.type",
						"equals.#":      acctest.Ct1,
						"equals.0":      "AWS::DynamoDB::Table",
					}),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.3.name", "s3OutpostsWriteOnlyEvents"),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.3.field_selector.#", acctest.Ct3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.3.field_selector.*", map[string]string{
						names.AttrField: "eventCategory",
						"equals.#":      acctest.Ct1,
						"equals.0":      "Data",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.3.field_selector.*", map[string]string{
						names.AttrField: "readOnly",
						"equals.#":      acctest.Ct1,
						"equals.0":      acctest.CtFalse,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.3.field_selector.*", map[string]string{
						names.AttrField: "resources.type",
						"equals.#":      acctest.Ct1,
						"equals.0":      "AWS::S3Outposts::Object",
					}),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.4.name", "managementEventsSelector"),
					resource.TestCheckResourceAttr(resourceName, "advanced_event_selector.4.field_selector.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "advanced_event_selector.4.field_selector.*", map[string]string{
						names.AttrField: "eventCategory",
						"equals.#":      acctest.Ct1,
						"equals.0":      "Management",
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

func testAccTrail_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var trail types.Trail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudTrailServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrailDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudTrailConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudtrail.ResourceTrail(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudtrail.ResourceTrail(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTrail_migrateV0(t *testing.T) {
	ctx := acctest.Context(t)
	var trail types.Trail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudtrail.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.CloudTrailServiceID),
		CheckDestroy: testAccCheckTrailDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.24.0",
					},
				},
				Config: testAccCloudTrailConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "cloudtrail", fmt.Sprintf("trail/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrName),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccCloudTrailConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrailExists(ctx, resourceName, &trail),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "cloudtrail", fmt.Sprintf("trail/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrARN),
				),
			},
		},
	})
}

func testAccCheckTrailExists(ctx context.Context, n string, v *types.Trail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudTrailClient(ctx)

		output, err := tfcloudtrail.FindTrailByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckLoggingEnabled(ctx context.Context, n string, want bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudTrailClient(ctx)

		output, err := conn.GetTrailStatus(ctx, &cloudtrail.GetTrailStatusInput{
			Name: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if got := aws.ToBool(output.IsLogging); got != want {
			return fmt.Errorf("Expected logging status %t, given %t", want, got)
		}

		return nil
	}
}

func testAccCheckLogValidationEnabled(n string, desired bool, trail *types.Trail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if trail.LogFileValidationEnabled == nil {
			return fmt.Errorf("No LogFileValidationEnabled attribute present")
		}

		logValid := aws.ToBool(trail.LogFileValidationEnabled)
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

func testAccCheckTrailDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudTrailClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudtrail" {
				continue
			}

			_, err := tfcloudtrail.FindTrailByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudTrail Trail (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCloudTrailConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

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
        Sid    = "AWSCloudTrailAclCheck"
        Effect = "Allow"
        Principal = {
          Service = "cloudtrail.amazonaws.com"
        }
        Action   = "s3:GetBucketAcl"
        Resource = "arn:${data.aws_partition.current.partition}:s3:::%[1]s"
        Condition = {
          StringEquals = {
            "aws:SourceArn" = "arn:${data.aws_partition.current.partition}:cloudtrail:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:trail/%[1]s"
          }
        }
      },
      {
        Sid    = "AWSCloudTrailWrite"
        Effect = "Allow"
        Principal = {
          Service = "cloudtrail.amazonaws.com"
        }
        Action   = "s3:PutObject"
        Resource = "arn:${data.aws_partition.current.partition}:s3:::%[1]s/*"
        Condition = {
          StringEquals = {
            "s3:x-amz-acl"  = "bucket-owner-full-control"
            "aws:SourceArn" = "arn:${data.aws_partition.current.partition}:cloudtrail:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:trail/%[1]s"
          }
        }
      }
    ]
  })
}
`, rName)
}

func testAccCloudTrailConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id
}
`, rName))
}

func testAccCloudTrailConfig_modified(rName string) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
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

func testAccCloudTrailConfig_enableLogging(rName string, enableLogging bool) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
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

func testAccCloudTrailConfig_cloudWatch(rName string) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
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

func testAccCloudTrailConfig_cloudWatchModified(rName string) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
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

func testAccCloudTrailConfig_multiRegion(rName string) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name                  = %[1]q
  s3_bucket_name        = aws_s3_bucket.test.id
  is_multi_region_trail = true
}
`, rName))
}

func testAccCloudTrailConfig_organization(rName string) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  is_organization_trail = true
  name                  = %[1]q
  s3_bucket_name        = aws_s3_bucket.test.id
}
`, rName))
}

func testAccCloudTrailConfig_logValidation(rName string) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
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

func testAccCloudTrailConfig_logValidationModified(rName string) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name                          = %[1]q
  s3_bucket_name                = aws_s3_bucket.test.id
  include_global_service_events = true
}
`, rName))
}

func testAccCloudTrailConfig_kmsKey(rName string) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
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

func testAccCloudTrailConfig_globalServiceEvents(rName string) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name                          = %[1]q
  s3_bucket_name                = aws_s3_bucket.test.id
  include_global_service_events = false
}
`, rName))
}

func testAccCloudTrailConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccCloudTrailConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccCloudTrailConfig_eventSelector(rName string) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
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

func testAccCloudTrailConfig_eventSelectorReadWriteType(rName string) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
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

func testAccCloudTrailConfig_eventSelectorModified(rName string) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
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
  runtime       = "nodejs16.x"
}
`, rName))
}

func testAccCloudTrailConfig_eventSelectorNone(rName string) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id
}
`, rName))
}

func testAccCloudTrailConfig_eventSelectorDynamoDB(rName string) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
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

func testAccCloudTrailConfig_eventSelectorExcludeKMS(rName string) string {
	return acctest.ConfigCompose(
		testAccCloudTrailConfig_base(rName),
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

func testAccCloudTrailConfig_eventSelectorExcludeKMSAndRDSData(rName string) string {
	return acctest.ConfigCompose(
		testAccCloudTrailConfig_base(rName),
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

func testAccCloudTrailConfig_insightSelector(rName string) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
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

func testAccCloudTrailConfig_insightSelectorMulti(rName string) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
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

func testAccCloudTrailConfig_advancedEventSelector(rName string) string {
	return acctest.ConfigCompose(testAccCloudTrailConfig_base(rName), fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  # Must have bucket policy attached first
  depends_on = [aws_s3_bucket_policy.test]

  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.id

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

resource "aws_s3_bucket" "test2" {
  bucket        = "%[1]s-2"
  force_destroy = true
}
`, rName))
}
