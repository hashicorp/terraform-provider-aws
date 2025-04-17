// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicequotas_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfservicequotas "github.com/hashicorp/terraform-provider-aws/internal/service/servicequotas"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTemplateAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicequotas_template_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.ServiceQuotaTemplateAssociationStatusAssociated)),
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

func testAccTemplateAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicequotas_template_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateAssociationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfservicequotas.ResourceTemplateAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTemplateAssociation_skipDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicequotas_template_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			acctest.PreCheckPartitionHasService(t, names.ServiceQuotasEndpointID)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceQuotasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateAssociationConfig_skipDestroy(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.ServiceQuotaTemplateAssociationStatusAssociated)),
				),
			},
			{
				// aws_servicequotas_template_association resource is removed from config
				Config: testAccTemplateConfig_basic(lambdaENIQuotaCode, lambdaServiceCode, lambdaENIValue),
			},
			{
				// Use the basic config to remove association on destroy
				Config: testAccTemplateAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.ServiceQuotaTemplateAssociationStatusAssociated)),
				),
			},
		},
	})
}

func testAccCheckTemplateAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceQuotasClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicequotas_template_association" {
				continue
			}

			_, err := tfservicequotas.FindTemplateAssociation(ctx, conn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Service Quotas Template Association still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTemplateAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceQuotasClient(ctx)

		_, err := tfservicequotas.FindTemplateAssociation(ctx, conn)

		return err
	}
}

func testAccTemplateAssociationConfig_basic() string {
	return acctest.ConfigCompose(
		testAccTemplateConfig_basic(lambdaENIQuotaCode, lambdaServiceCode, lambdaENIValue),
		`
resource "aws_servicequotas_template_association" "test" {
  depends_on = ["aws_servicequotas_template.test"]
}
`)
}

func testAccTemplateAssociationConfig_skipDestroy() string {
	return acctest.ConfigCompose(
		testAccTemplateConfig_basic(lambdaENIQuotaCode, lambdaServiceCode, lambdaENIValue),
		`
resource "aws_servicequotas_template_association" "test" {
  depends_on = ["aws_servicequotas_template.test"]

  skip_destroy = true
}
`)
}
