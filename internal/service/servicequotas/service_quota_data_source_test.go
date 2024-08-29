// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicequotas_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceQuotasServiceQuotaDataSource_quotaCode(t *testing.T) {
	ctx := acctest.Context(t)
	const dataSourceName = "data.aws_servicequotas_service_quota.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheckServiceQuotaSet(ctx, t, setQuotaServiceCode, setQuotaQuotaCode)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceQuotaDataSourceConfig_code(setQuotaServiceCode, setQuotaQuotaCode),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "adjustable", acctest.CtTrue),
					acctest.CheckResourceAttrRegionalARN(dataSourceName, names.AttrARN, "servicequotas", fmt.Sprintf("%s/%s", setQuotaServiceCode, setQuotaQuotaCode)),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDefaultValue, "5"),
					resource.TestCheckResourceAttr(dataSourceName, "global_quota", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, "quota_code", setQuotaQuotaCode),
					resource.TestCheckResourceAttr(dataSourceName, "quota_name", "VPCs per Region"),
					resource.TestCheckResourceAttr(dataSourceName, "service_code", setQuotaServiceCode),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrServiceName, "Amazon Virtual Private Cloud (Amazon VPC)"),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.#", acctest.Ct0),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrValue, regexache.MustCompile(`^\d+$`)),
				),
			},
		},
	})
}

func TestAccServiceQuotasServiceQuotaDataSource_quotaCode_Unset(t *testing.T) {
	ctx := acctest.Context(t)
	const dataSourceName = "data.aws_servicequotas_service_quota.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheckServiceQuotaUnset(ctx, t, unsetQuotaServiceCode, unsetQuotaQuotaCode)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceQuotaDataSourceConfig_code(unsetQuotaServiceCode, unsetQuotaQuotaCode),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrRegionalARNNoAccount(dataSourceName, names.AttrARN, "servicequotas", fmt.Sprintf("%s/%s", unsetQuotaServiceCode, unsetQuotaQuotaCode)),
					resource.TestCheckResourceAttr(dataSourceName, "adjustable", acctest.CtTrue),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrDefaultValue, regexache.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(dataSourceName, "global_quota", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, "quota_code", unsetQuotaQuotaCode),
					resource.TestCheckResourceAttr(dataSourceName, "quota_name", unsetQuotaQuotaName),
					resource.TestCheckResourceAttr(dataSourceName, "service_code", unsetQuotaServiceCode),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrServiceName, "Amazon Simple Storage Service (Amazon S3)"),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.#", acctest.Ct0),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrValue, regexache.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrValue, dataSourceName, names.AttrDefaultValue),
				),
			},
		},
	})
}

func TestAccServiceQuotasServiceQuotaDataSource_quotaCode_hasUsageMetric(t *testing.T) {
	ctx := acctest.Context(t)
	const dataSourceName = "data.aws_servicequotas_service_quota.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheckServiceQuotaHasUsageMetric(ctx, t, hasUsageMetricServiceCode, hasUsageMetricQuotaCode)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceQuotaDataSourceConfig_code(hasUsageMetricServiceCode, hasUsageMetricQuotaCode),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrRegionalARN(dataSourceName, names.AttrARN, "servicequotas", fmt.Sprintf("%s/%s", hasUsageMetricServiceCode, hasUsageMetricQuotaCode)),
					resource.TestCheckResourceAttr(dataSourceName, "adjustable", acctest.CtTrue),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDefaultValue, "500"),
					resource.TestCheckResourceAttr(dataSourceName, "global_quota", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, "quota_code", hasUsageMetricQuotaCode),
					resource.TestCheckResourceAttr(dataSourceName, "quota_name", hasUsageMetricQuotaName),
					resource.TestCheckResourceAttr(dataSourceName, "service_code", hasUsageMetricServiceCode),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrServiceName, "Amazon EC2 Auto Scaling"),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.0.metric_namespace", "AWS/Usage"),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.0.metric_name", "ResourceCount"),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.0.metric_statistic_recommendation", "Maximum"),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.0.metric_dimensions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.0.metric_dimensions.0.service", "AutoScaling"),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.0.metric_dimensions.0.class", "None"),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.0.metric_dimensions.0.type", "Resource"),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.0.metric_dimensions.0.resource", "NumberOfAutoScalingGroup"),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrValue, regexache.MustCompile(`^\d+$`)),
				),
			},
		},
	})
}

func TestAccServiceQuotasServiceQuotaDataSource_PermissionError_quotaCode(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckAssumeRoleARN(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceQuotaDataSourceConfig_permissionErrorCode("elasticloadbalancing", "L-53DA6B97"),
				ExpectError: regexache.MustCompile(`DEPENDENCY_ACCESS_DENIED_ERROR`),
			},
		},
	})
}

func TestAccServiceQuotasServiceQuotaDataSource_quotaName(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_servicequotas_service_quota.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheckServiceQuotaSet(ctx, t, setQuotaServiceCode, setQuotaQuotaCode)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceQuotaDataSourceConfig_name("vpc", setQuotaQuotaName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "adjustable", acctest.CtTrue),
					acctest.CheckResourceAttrRegionalARN(dataSourceName, names.AttrARN, "servicequotas", fmt.Sprintf("%s/%s", setQuotaServiceCode, setQuotaQuotaCode)),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDefaultValue, "5"),
					resource.TestCheckResourceAttr(dataSourceName, "global_quota", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, "quota_code", setQuotaQuotaCode),
					resource.TestCheckResourceAttr(dataSourceName, "quota_name", setQuotaQuotaName),
					resource.TestCheckResourceAttr(dataSourceName, "service_code", setQuotaServiceCode),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrServiceName, "Amazon Virtual Private Cloud (Amazon VPC)"),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.#", acctest.Ct0),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrValue, regexache.MustCompile(`^\d+$`)),
				),
			},
		},
	})
}

func TestAccServiceQuotasServiceQuotaDataSource_quotaName_Unset(t *testing.T) {
	ctx := acctest.Context(t)
	const dataSourceName = "data.aws_servicequotas_service_quota.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheckServiceQuotaUnset(ctx, t, unsetQuotaServiceCode, unsetQuotaQuotaCode)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceQuotaDataSourceConfig_name(unsetQuotaServiceCode, unsetQuotaQuotaName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrRegionalARNNoAccount(dataSourceName, names.AttrARN, "servicequotas", fmt.Sprintf("%s/%s", unsetQuotaServiceCode, unsetQuotaQuotaCode)),
					resource.TestCheckResourceAttr(dataSourceName, "adjustable", acctest.CtTrue),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrDefaultValue, regexache.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(dataSourceName, "global_quota", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, "quota_code", unsetQuotaQuotaCode),
					resource.TestCheckResourceAttr(dataSourceName, "quota_name", unsetQuotaQuotaName),
					resource.TestCheckResourceAttr(dataSourceName, "service_code", unsetQuotaServiceCode),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrServiceName, "Amazon Simple Storage Service (Amazon S3)"),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.#", acctest.Ct0),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrValue, regexache.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrValue, dataSourceName, names.AttrDefaultValue),
				),
			},
		},
	})
}

func TestAccServiceQuotasServiceQuotaDataSource_quotaName_hasUsageMetric(t *testing.T) {
	ctx := acctest.Context(t)
	const dataSourceName = "data.aws_servicequotas_service_quota.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheckServiceQuotaHasUsageMetric(ctx, t, hasUsageMetricServiceCode, hasUsageMetricQuotaCode)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceQuotaDataSourceConfig_name(hasUsageMetricServiceCode, hasUsageMetricQuotaName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrRegionalARN(dataSourceName, names.AttrARN, "servicequotas", fmt.Sprintf("%s/%s", hasUsageMetricServiceCode, hasUsageMetricQuotaCode)),
					resource.TestCheckResourceAttr(dataSourceName, "adjustable", acctest.CtTrue),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDefaultValue, "500"),
					resource.TestCheckResourceAttr(dataSourceName, "global_quota", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, "quota_code", hasUsageMetricQuotaCode),
					resource.TestCheckResourceAttr(dataSourceName, "quota_name", hasUsageMetricQuotaName),
					resource.TestCheckResourceAttr(dataSourceName, "service_code", hasUsageMetricServiceCode),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrServiceName, "Amazon EC2 Auto Scaling"),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.0.metric_namespace", "AWS/Usage"),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.0.metric_name", "ResourceCount"),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.0.metric_statistic_recommendation", "Maximum"),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.0.metric_dimensions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.0.metric_dimensions.0.service", "AutoScaling"),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.0.metric_dimensions.0.class", "None"),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.0.metric_dimensions.0.type", "Resource"),
					resource.TestCheckResourceAttr(dataSourceName, "usage_metric.0.metric_dimensions.0.resource", "NumberOfAutoScalingGroup"),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrValue, regexache.MustCompile(`^\d+$`)),
				),
			},
		},
	})
}

func TestAccServiceQuotasServiceQuotaDataSource_PermissionError_quotaName(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckAssumeRoleARN(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceQuotaDataSourceConfig_permissionErrorName("elasticloadbalancing", "Application Load Balancers per Region"),
				ExpectError: regexache.MustCompile(`DEPENDENCY_ACCESS_DENIED_ERROR`),
			},
		},
	})
}

func testAccServiceQuotaDataSourceConfig_code(serviceCode, quotaCode string) string {
	return fmt.Sprintf(`
data "aws_servicequotas_service_quota" "test" {
  quota_code   = %[1]q
  service_code = %[2]q
}
`, quotaCode, serviceCode)
}

func testAccServiceQuotaDataSourceConfig_permissionErrorCode(serviceCode, quotaCode string) string {
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
data "aws_servicequotas_service_quota" "test" {
  service_code = %[1]q
  quota_code   = %[2]q
}
`, serviceCode, quotaCode))
}

func testAccServiceQuotaDataSourceConfig_name(serviceCode, quotaName string) string {
	return fmt.Sprintf(`
data "aws_servicequotas_service_quota" "test" {
  quota_name   = %[1]q
  service_code = %[2]q
}
`, quotaName, serviceCode)
}

func testAccServiceQuotaDataSourceConfig_permissionErrorName(serviceCode, quotaName string) string {
	policy := `{
  "Version": "2012-10-17",
  "Statement": [
    {
  	  "Effect": "Allow",
  	  "Action": [
  	    "servicequotas:ListServiceQuotas"
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
data "aws_servicequotas_service_quota" "test" {
  service_code = %[1]q
  quota_name   = %[2]q
}
`, serviceCode, quotaName))
}
