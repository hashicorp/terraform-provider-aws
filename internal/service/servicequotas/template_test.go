// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicequotas_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
	var template types.ServiceQuotaIncreaseRequestInTemplate
	resourceName := "aws_servicequotas_template.test"
	regionDataSourceName := "data.aws_region.current"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_basic(lambdaStorageQuotaCode, lambdaServiceCode, lambdaStorageValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, resourceName, &template),
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
	var template types.ServiceQuotaIncreaseRequestInTemplate
	resourceName := "aws_servicequotas_template.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_basic(lambdaConcurrentExecQuotaCode, lambdaServiceCode, lambdaConcurrentExecValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, resourceName, &template),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfservicequotas.ResourceTemplate, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTemplate_value(t *testing.T) {
	ctx := acctest.Context(t)
	var template types.ServiceQuotaIncreaseRequestInTemplate
	resourceName := "aws_servicequotas_template.test"
	regionDataSourceName := "data.aws_region.current"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_basic(lambdaENIQuotaCode, lambdaServiceCode, lambdaENIValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, resourceName, &template),
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
					testAccCheckTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRegion, regionDataSourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "quota_code", lambdaENIQuotaCode),
					resource.TestCheckResourceAttr(resourceName, "service_code", lambdaServiceCode),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, lambdaENIValueUpdated),
				),
			},
		},
	})
}

func testAccCheckTemplateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceQuotasClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicequotas_template" {
				continue
			}

			_, err := tfservicequotas.FindTemplateByID(ctx, conn, rs.Primary.ID)
			if errs.IsA[*types.NoSuchResourceException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ServiceQuotas, create.ErrActionCheckingDestroyed, tfservicequotas.ResNameTemplate, rs.Primary.ID, err)
			}

			return create.Error(names.ServiceQuotas, create.ErrActionCheckingDestroyed, tfservicequotas.ResNameTemplate, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTemplateExists(ctx context.Context, name string, template *types.ServiceQuotaIncreaseRequestInTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ServiceQuotas, create.ErrActionCheckingExistence, tfservicequotas.ResNameTemplate, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ServiceQuotas, create.ErrActionCheckingExistence, tfservicequotas.ResNameTemplate, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceQuotasClient(ctx)
		resp, err := tfservicequotas.FindTemplateByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.ServiceQuotas, create.ErrActionCheckingExistence, tfservicequotas.ResNameTemplate, rs.Primary.ID, err)
		}

		*template = *resp

		return nil
	}
}

func testAccPreCheckTemplate(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceQuotasClient(ctx)

	input := &servicequotas.ListServiceQuotaIncreaseRequestsInTemplateInput{}
	_, err := conn.ListServiceQuotaIncreaseRequestsInTemplate(ctx, input)

	// Request must come from organization owner account
	if errs.IsAErrorMessageContains[*types.AccessDeniedException](err, "The request was called by a member account.") {
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
  region       = data.aws_region.current.name
  quota_code   = %[1]q
  service_code = %[2]q
  value        = %[3]s
}
`, quotaCode, serviceCode, value)
}
