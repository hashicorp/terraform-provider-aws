// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicequotas_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfservicequotas "github.com/hashicorp/terraform-provider-aws/internal/service/servicequotas"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	lambdaServiceCode = "lambda"

	lambdaStorageQuotaCode = "L-2ACBD22F" // Function and layer storage
	lambdaStorageValue     = "80"         // Default is 75 GB

	lambdaConcurrentExecQuotaCode = "L-B99A9384" // Concurrent executions
	lambdaConcurrentExecValue     = "1500"       // Default is 1000

	lambdaENIQuotaCode    = "L-9FEE3D26" // Elastic network interfaces per VPC
	lambdaENIValue        = "275"        // Default is 250
	lambdaENIValueUpdated = "300"        // Default is 250
)

func testAccTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.ServiceQuotaIncreaseRequestInTemplate
	resourceName := "aws_servicequotas_template.test"
	regionDataSourceName := "data.aws_region.current"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_basic(lambdaStorageQuotaCode, lambdaServiceCode, lambdaStorageValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &template),
					resource.TestCheckResourceAttrPair(resourceName, "aws_region", regionDataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRegion, regionDataSourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "quota_code", lambdaStorageQuotaCode),
					resource.TestCheckResourceAttr(resourceName, "service_code", lambdaServiceCode),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, lambdaStorageValue),
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

func testAccTemplate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.ServiceQuotaIncreaseRequestInTemplate
	resourceName := "aws_servicequotas_template.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_basic(lambdaConcurrentExecQuotaCode, lambdaServiceCode, lambdaConcurrentExecValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &template),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfservicequotas.ResourceTemplate, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTemplate_value(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.ServiceQuotaIncreaseRequestInTemplate
	resourceName := "aws_servicequotas_template.test"
	regionDataSourceName := "data.aws_region.current"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_basic(lambdaENIQuotaCode, lambdaServiceCode, lambdaENIValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &template),
					resource.TestCheckResourceAttrPair(resourceName, "aws_region", regionDataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRegion, regionDataSourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "quota_code", lambdaENIQuotaCode),
					resource.TestCheckResourceAttr(resourceName, "service_code", lambdaServiceCode),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, lambdaENIValue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTemplateConfig_basic(lambdaENIQuotaCode, lambdaServiceCode, lambdaENIValueUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &template),
					resource.TestCheckResourceAttrPair(resourceName, "aws_region", regionDataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRegion, regionDataSourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "quota_code", lambdaENIQuotaCode),
					resource.TestCheckResourceAttr(resourceName, "service_code", lambdaServiceCode),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, lambdaENIValueUpdated),
				),
			},
		},
	})
}

func testAccTemplate_region(t *testing.T) {
	ctx := acctest.Context(t)
	var template awstypes.ServiceQuotaIncreaseRequestInTemplate
	resourceName := "aws_servicequotas_template.test"
	regionDataSourceName := "data.aws_region.current"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_region(lambdaStorageQuotaCode, lambdaServiceCode, lambdaStorageValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &template),
					resource.TestCheckResourceAttrPair(resourceName, "aws_region", regionDataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRegion, regionDataSourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "quota_code", lambdaStorageQuotaCode),
					resource.TestCheckResourceAttr(resourceName, "service_code", lambdaServiceCode),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, lambdaStorageValue),
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

func testAccCheckTemplateDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ServiceQuotasClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicequotas_template" {
				continue
			}

			_, err := tfservicequotas.FindTemplateByThreePartKey(ctx, conn, rs.Primary.Attributes["aws_region"], rs.Primary.Attributes["quota_code"], rs.Primary.Attributes["service_code"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Service Quotas Template still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTemplateExists(ctx context.Context, t *testing.T, n string, v *awstypes.ServiceQuotaIncreaseRequestInTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ServiceQuotasClient(ctx)

		output, err := tfservicequotas.FindTemplateByThreePartKey(ctx, conn, rs.Primary.Attributes["aws_region"], rs.Primary.Attributes["quota_code"], rs.Primary.Attributes["service_code"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheckTemplate(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ServiceQuotasClient(ctx)

	input := &servicequotas.ListServiceQuotaIncreaseRequestsInTemplateInput{}
	_, err := conn.ListServiceQuotaIncreaseRequestsInTemplate(ctx, input)

	// Request must come from organization owner account
	if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "The request was called by a member account.") {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccTemplateConfig_basic(quotaCode, serviceCode, value string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_servicequotas_template" "test" {
  aws_region   = data.aws_region.current.region
  quota_code   = %[1]q
  service_code = %[2]q
  value        = %[3]s
}
`, quotaCode, serviceCode, value)
}

func testAccTemplateConfig_region(quotaCode, serviceCode, value string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_servicequotas_template" "test" {
  region       = data.aws_region.current.region
  quota_code   = %[1]q
  service_code = %[2]q
  value        = %[3]s
}
`, quotaCode, serviceCode, value)
}
