// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalogappregistry_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfservicecatalogappregistry "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalogappregistry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceCatalogAppRegistryApplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_servicecatalogappregistry_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceCatalogAppRegistryEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "servicecatalog", regexache.MustCompile(`/applications/+.`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_servicecatalogappregistry_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceCatalogAppRegistryEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfservicecatalogappregistry.ResourceApplication, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServiceCatalogAppRegistryApplication_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "text1"
	descriptionUpdated := "text2"
	resourceName := "aws_servicecatalogappregistry_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceCatalogAppRegistryEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_description(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName),
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
					testAccCheckApplicationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionUpdated),
				),
			},
			{
				Config: testAccApplicationConfig_description(rNameUpdated, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
				),
			},
		},
	})
}

func testAccCheckApplicationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogAppRegistryClient(ctx)

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

func testAccCheckApplicationExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, tfservicecatalogappregistry.ResNameApplication, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, tfservicecatalogappregistry.ResNameApplication, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogAppRegistryClient(ctx)

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
