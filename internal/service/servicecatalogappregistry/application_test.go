// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicecatalogappregistry_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfservicecatalogappregistry "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalogappregistry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceCatalogAppRegistryApplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_servicecatalogappregistry_application.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceCatalogAppRegistryEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "servicecatalog", regexache.MustCompile(`/applications/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccServiceCatalogAppRegistryApplication_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_servicecatalogappregistry_application.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceCatalogAppRegistryEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfservicecatalogappregistry.ResourceApplication, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServiceCatalogAppRegistryApplication_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	description := "text1"
	descriptionUpdated := "text2"
	resourceName := "aws_servicecatalogappregistry_application.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceCatalogAppRegistryEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_description(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_description(rName, descriptionUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionUpdated),
				),
			},
			{
				Config: testAccApplicationConfig_description(rNameUpdated, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
				),
			},
		},
	})
}

func testAccCheckApplicationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ServiceCatalogAppRegistryClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalogappregistry_application" {
				continue
			}

			_, err := tfservicecatalogappregistry.FindApplicationByID(ctx, conn, rs.Primary.ID)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingDestroyed, tfservicecatalogappregistry.ResNameApplication, rs.Primary.ID, err)
			}

			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingDestroyed, tfservicecatalogappregistry.ResNameApplication, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckApplicationExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, tfservicecatalogappregistry.ResNameApplication, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, tfservicecatalogappregistry.ResNameApplication, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).ServiceCatalogAppRegistryClient(ctx)

		_, err := tfservicecatalogappregistry.FindApplicationByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccApplicationConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalogappregistry_application" "test" {
  name = %[1]q
}

resource "aws_ssm_parameter" "test" {
  name  = %[1]q
  type  = "String"
  value = "test"
  tags  = aws_servicecatalogappregistry_application.test.application_tag
}
`, name)
}

func testAccApplicationConfig_description(name, description string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalogappregistry_application" "test" {
  name        = %[1]q
  description = %[2]q
}
`, name, description)
}
