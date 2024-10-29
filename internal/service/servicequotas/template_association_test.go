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
	tfservicequotas "github.com/hashicorp/terraform-provider-aws/internal/service/servicequotas"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTemplateAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicequotas_template_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID)
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
			acctest.PreCheckRegion(t, names.USEast1RegionID)
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
			acctest.PreCheckRegion(t, names.USEast1RegionID)
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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateAssociationAssociated(ctx), // verify association is still live on the remote
				),
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

			out, err := conn.GetAssociationForServiceQuotaTemplate(ctx, &servicequotas.GetAssociationForServiceQuotaTemplateInput{})
			if out != nil && out.ServiceQuotaTemplateAssociationStatus == types.ServiceQuotaTemplateAssociationStatusDisassociated {
				return nil
			}
			if err != nil {
				return create.Error(names.ServiceQuotas, create.ErrActionCheckingDestroyed, tfservicequotas.ResNameTemplateAssociation, rs.Primary.ID, err)
			}

			return create.Error(names.ServiceQuotas, create.ErrActionCheckingDestroyed, tfservicequotas.ResNameTemplateAssociation, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTemplateAssociationExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ServiceQuotas, create.ErrActionCheckingExistence, tfservicequotas.ResNameTemplateAssociation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ServiceQuotas, create.ErrActionCheckingExistence, tfservicequotas.ResNameTemplateAssociation, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceQuotasClient(ctx)
		out, err := conn.GetAssociationForServiceQuotaTemplate(ctx, &servicequotas.GetAssociationForServiceQuotaTemplateInput{})
		if err != nil {
			return create.Error(names.ServiceQuotas, create.ErrActionCheckingExistence, tfservicequotas.ResNameTemplateAssociation, rs.Primary.ID, err)
		}
		if out != nil && out.ServiceQuotaTemplateAssociationStatus == types.ServiceQuotaTemplateAssociationStatusDisassociated {
			return create.Error(names.ServiceQuotas, create.ErrActionCheckingExistence, tfservicequotas.ResNameTemplateAssociation, rs.Primary.ID, fmt.Errorf("unexpected status: %s", out.ServiceQuotaTemplateAssociationStatus))
		}

		return nil
	}
}

// testAccCheckTemplateAssociationAssociated is a helper function for verifying a
// template association remains in place when the skip_destroy argument is set to true
// and the association resource is removed
func testAccCheckTemplateAssociationAssociated(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceQuotasClient(ctx)

		out, err := conn.GetAssociationForServiceQuotaTemplate(ctx, &servicequotas.GetAssociationForServiceQuotaTemplateInput{})
		if err != nil {
			return create.Error(names.ServiceQuotas, create.ErrActionCheckingExistence, tfservicequotas.ResNameTemplateAssociation, "", err)
		}
		if out == nil || out.ServiceQuotaTemplateAssociationStatus != types.ServiceQuotaTemplateAssociationStatusAssociated {
			return create.Error(names.ServiceQuotas, create.ErrActionCheckingExistence, tfservicequotas.ResNameTemplateAssociation, "", fmt.Errorf("unexpected status: %s", out.ServiceQuotaTemplateAssociationStatus))
		}

		return nil
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
