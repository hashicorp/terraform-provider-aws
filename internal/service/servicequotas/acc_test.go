// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicequotas_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

const (
	setQuotaServiceCode = "vpc"
	setQuotaQuotaCode   = "L-F678F1CE"
	setQuotaQuotaName   = "VPCs per Region"

	unsetQuotaServiceCode = "s3"
	unsetQuotaQuotaCode   = "L-FAABEEBA"
	unsetQuotaQuotaName   = "Access Points"

	hasUsageMetricServiceCode = "autoscaling"
	hasUsageMetricQuotaCode   = "L-CDE20ADC"
	hasUsageMetricQuotaName   = "Auto Scaling groups per region"
)

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceQuotasClient(ctx)

	input := &servicequotas.ListServicesInput{}

	_, err := conn.ListServices(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

// nosemgrep:ci.servicequotas-in-func-name
func testAccPreCheckServiceQuotaSet(ctx context.Context, t *testing.T, serviceCode, quotaCode string) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceQuotasClient(ctx)

	input := &servicequotas.GetServiceQuotaInput{
		QuotaCode:   aws.String(quotaCode),
		ServiceCode: aws.String(serviceCode),
	}

	_, err := conn.GetServiceQuota(ctx, input)
	var nsr *types.NoSuchResourceException
	if errors.As(err, &nsr) {
		t.Fatalf("The Service Quota (%s/%s) has never been set. This test can only be run with a quota that has previously been set. Please update the test to check a new quota.", serviceCode, quotaCode)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error getting Service Quota (%s/%s) : %s", serviceCode, quotaCode, err)
	}
}

func testAccPreCheckServiceQuotaUnset(ctx context.Context, t *testing.T, serviceCode, quotaCode string) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceQuotasClient(ctx)

	input := &servicequotas.GetServiceQuotaInput{
		QuotaCode:   aws.String(quotaCode),
		ServiceCode: aws.String(serviceCode),
	}

	_, err := conn.GetServiceQuota(ctx, input)
	if err == nil {
		t.Fatalf("The Service Quota (%s/%s) has been set. This test can only be run with a quota that has never been set. Please update the test to check a new quota.", serviceCode, quotaCode)
	}
	var nsr *types.NoSuchResourceException
	if !errors.As(err, &nsr) {
		t.Fatalf("unexpected PreCheck error getting Service Quota (%s/%s) : %s", serviceCode, quotaCode, err)
	}
}

func testAccPreCheckServiceQuotaHasUsageMetric(ctx context.Context, t *testing.T, serviceCode, quotaCode string) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceQuotasClient(ctx)

	input := &servicequotas.GetAWSDefaultServiceQuotaInput{
		QuotaCode:   aws.String(quotaCode),
		ServiceCode: aws.String(serviceCode),
	}

	quota, err := conn.GetAWSDefaultServiceQuota(ctx, input)
	if err != nil {
		t.Fatalf("unexpected PreCheck error getting Service Quota (%s/%s) : %s", serviceCode, quotaCode, err)
	}
	if quota.Quota.UsageMetric == nil || quota.Quota.UsageMetric.MetricName == nil {
		t.Fatalf("The Service Quota (%s/%s) does not have a usage metric. This test can only be run with a quota that has a usage metric. Please update the test to check a new quota.", serviceCode, quotaCode)
	}
}
