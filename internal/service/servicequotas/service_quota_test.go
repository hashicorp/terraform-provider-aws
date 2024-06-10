// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicequotas_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// This resource is different than many since quotas are pre-existing
// and the resource is only designed to help with increases.
// In the basic case, we test that the resource can match the existing quota
// without unexpected changes.
func TestAccServiceQuotasServiceQuota_basic(t *testing.T) {
	ctx := acctest.Context(t)
	const dataSourceName = "data.aws_servicequotas_service_quota.test"
	const resourceName = "aws_servicequotas_service_quota.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckServiceQuotaSet(ctx, t, setQuotaServiceCode, setQuotaQuotaCode)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceQuotaConfig_sameValue(setQuotaServiceCode, setQuotaQuotaCode),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "adjustable", dataSourceName, "adjustable"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDefaultValue, dataSourceName, names.AttrDefaultValue),
					resource.TestCheckResourceAttrPair(resourceName, "quota_code", dataSourceName, "quota_code"),
					resource.TestCheckResourceAttrPair(resourceName, "quota_name", dataSourceName, "quota_name"),
					resource.TestCheckResourceAttrPair(resourceName, "service_code", dataSourceName, "service_code"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceName, dataSourceName, names.AttrServiceName),
					resource.TestCheckResourceAttrPair(resourceName, "usage_metric", dataSourceName, "usage_metric"),
					resource.TestCheckNoResourceAttr(resourceName, "usage_metric.0.metric_name"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrValue, dataSourceName, names.AttrValue),
					resource.TestCheckNoResourceAttr(resourceName, "request_id"),
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

func TestAccServiceQuotasServiceQuota_basic_Unset(t *testing.T) {
	ctx := acctest.Context(t)
	const dataSourceName = "data.aws_servicequotas_service_quota.test"
	const resourceName = "aws_servicequotas_service_quota.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckServiceQuotaUnset(ctx, t, unsetQuotaServiceCode, unsetQuotaQuotaCode)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceQuotaConfig_sameValue(unsetQuotaServiceCode, unsetQuotaQuotaCode),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "adjustable", dataSourceName, "adjustable"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDefaultValue, dataSourceName, names.AttrDefaultValue),
					resource.TestCheckResourceAttrPair(resourceName, "quota_code", dataSourceName, "quota_code"),
					resource.TestCheckResourceAttrPair(resourceName, "quota_name", dataSourceName, "quota_name"),
					resource.TestCheckResourceAttrPair(resourceName, "service_code", dataSourceName, "service_code"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceName, dataSourceName, names.AttrServiceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrValue, dataSourceName, names.AttrValue),
					resource.TestCheckResourceAttrPair(resourceName, "usage_metric", dataSourceName, "usage_metric"),
					resource.TestCheckNoResourceAttr(resourceName, "usage_metric.0.metric_name"),
					resource.TestCheckNoResourceAttr(resourceName, "request_id"),
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

func TestAccServiceQuotasServiceQuota_basic_hasUsageMetric(t *testing.T) {
	ctx := acctest.Context(t)
	const dataSourceName = "data.aws_servicequotas_service_quota.test"
	const resourceName = "aws_servicequotas_service_quota.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			testAccPreCheckServiceQuotaHasUsageMetric(ctx, t, hasUsageMetricServiceCode, hasUsageMetricQuotaCode)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceQuotaConfig_sameValue(hasUsageMetricServiceCode, hasUsageMetricQuotaCode),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "adjustable", dataSourceName, "adjustable"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDefaultValue, dataSourceName, names.AttrDefaultValue),
					resource.TestCheckResourceAttrPair(resourceName, "quota_code", dataSourceName, "quota_code"),
					resource.TestCheckResourceAttrPair(resourceName, "quota_name", dataSourceName, "quota_name"),
					resource.TestCheckResourceAttrPair(resourceName, "service_code", dataSourceName, "service_code"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceName, dataSourceName, names.AttrServiceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrValue, dataSourceName, names.AttrValue),
					resource.TestCheckResourceAttrPair(resourceName, "usage_metric", dataSourceName, "usage_metric"),
					resource.TestCheckResourceAttrPair(resourceName, "usage_metric.0.metric_name", dataSourceName, "usage_metric.0.metric_name"),
					resource.TestCheckResourceAttr(resourceName, "usage_metric.0.metric_dimensions.#", acctest.Ct1),
					resource.TestCheckNoResourceAttr(resourceName, "request_id"),
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

func TestAccServiceQuotasServiceQuota_Value_increaseOnCreate(t *testing.T) {
	ctx := acctest.Context(t)
	quotaCode := os.Getenv("SERVICEQUOTAS_INCREASE_ON_CREATE_QUOTA_CODE")
	if quotaCode == "" {
		t.Skip(
			"Environment variable SERVICEQUOTAS_INCREASE_ON_CREATE_QUOTA_CODE is not set. " +
				"WARNING: This test will submit a real service quota increase!")
	}

	serviceCode := os.Getenv("SERVICEQUOTAS_INCREASE_ON_CREATE_SERVICE_CODE")
	if serviceCode == "" {
		t.Skip(
			"Environment variable SERVICEQUOTAS_INCREASE_ON_CREATE_SERVICE_CODE is not set. " +
				"WARNING: This test will submit a real service quota increase!")
	}

	value := os.Getenv("SERVICEQUOTAS_INCREASE_ON_CREATE_VALUE")
	if value == "" {
		t.Skip(
			"Environment variable SERVICEQUOTAS_INCREASE_ON_CREATE_VALUE is not set. " +
				"WARNING: This test will submit a real service quota increase!")
	}

	resourceName := "aws_servicequotas_service_quota.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceQuotaConfig_value(serviceCode, quotaCode, value),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "quota_code", quotaCode),
					resource.TestCheckResourceAttr(resourceName, "service_code", serviceCode),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, value),
					resource.TestCheckResourceAttrSet(resourceName, "request_id"),
				),
			},
		},
	})
}

func TestAccServiceQuotasServiceQuota_Value_increaseOnUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	quotaCode := os.Getenv("SERVICEQUOTAS_INCREASE_ON_UPDATE_QUOTA_CODE")
	if quotaCode == "" {
		t.Skip(
			"Environment variable SERVICEQUOTAS_INCREASE_ON_UPDATE_QUOTA_CODE is not set. " +
				"WARNING: This test will submit a real service quota increase!")
	}

	serviceCode := os.Getenv("SERVICEQUOTAS_INCREASE_ON_UPDATE_SERVICE_CODE")
	if serviceCode == "" {
		t.Skip(
			"Environment variable SERVICEQUOTAS_INCREASE_ON_UPDATE_SERVICE_CODE is not set. " +
				"WARNING: This test will submit a real service quota increase!")
	}

	value := os.Getenv("SERVICEQUOTAS_INCREASE_ON_UPDATE_VALUE")
	if value == "" {
		t.Skip(
			"Environment variable SERVICEQUOTAS_INCREASE_ON_UPDATE_VALUE is not set. " +
				"WARNING: This test will submit a real service quota increase!")
	}

	dataSourceName := "aws_servicequotas_service_quota.test"
	resourceName := "aws_servicequotas_service_quota.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceQuotaConfig_sameValue(serviceCode, quotaCode),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "quota_code", quotaCode),
					resource.TestCheckResourceAttr(resourceName, "service_code", serviceCode),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrValue, dataSourceName, names.AttrValue),
					resource.TestCheckNoResourceAttr(resourceName, "request_id"),
				),
			},
			{
				Config: testAccServiceQuotaConfig_value(serviceCode, quotaCode, value),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "quota_code", quotaCode),
					resource.TestCheckResourceAttr(resourceName, "service_code", serviceCode),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, value),
					resource.TestCheckResourceAttrSet(resourceName, "request_id"),
				),
			},
		},
	})
}

func TestAccServiceQuotasServiceQuota_permissionError(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t); acctest.PreCheckAssumeRoleARN(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceQuotaConfig_permissionError("elasticloadbalancing", "L-53DA6B97"),
				ExpectError: regexache.MustCompile(`DEPENDENCY_ACCESS_DENIED_ERROR`),
			},
		},
	})
}

// nosemgrep:ci.servicequotas-in-func-name
func testAccServiceQuotaConfig_sameValue(serviceCode, quotaCode string) string {
	return fmt.Sprintf(`
data "aws_servicequotas_service_quota" "test" {
  quota_code   = %[1]q
  service_code = %[2]q
}

resource "aws_servicequotas_service_quota" "test" {
  quota_code   = data.aws_servicequotas_service_quota.test.quota_code
  service_code = data.aws_servicequotas_service_quota.test.service_code
  value        = data.aws_servicequotas_service_quota.test.value
}
`, quotaCode, serviceCode)
}

func testAccServiceQuotaConfig_value(serviceCode, quotaCode, value string) string {
	return fmt.Sprintf(`
resource "aws_servicequotas_service_quota" "test" {
  quota_code   = %[1]q
  service_code = %[2]q
  value        = %[3]s
}
`, quotaCode, serviceCode, value)
}

func testAccServiceQuotaConfig_permissionError(serviceCode, quotaCode string) string {
	policy := `{
  "Version": "2012-10-17",
  "Statement": [
    {
  	  "Effect": "Allow",
  	  "Action": [
  	    "servicequotas:GetServiceQuota"
  	  ],
  	  "Resource": "*"
    },
    {
  	  "Effect": "Deny",
  	  "Action": [
  	    "elasticloadbalancing:*"
  	  ],
  	  "Resource": "*"
    }
  ]
}`

	return acctest.ConfigCompose(
		acctest.ConfigAssumeRolePolicy(policy),
		fmt.Sprintf(`
resource "aws_servicequotas_service_quota" "test" {
  service_code = %[1]q
  quota_code   = %[2]q
  value        = 1
}
`, serviceCode, quotaCode))
}
