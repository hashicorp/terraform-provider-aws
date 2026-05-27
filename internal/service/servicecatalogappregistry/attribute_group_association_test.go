// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicecatalogappregistry_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfservicecatalogappregistry "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalogappregistry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceCatalogAppRegistryAttributeGroupAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_servicecatalogappregistry_attribute_group_association.test"
	applicationResourceName := "aws_servicecatalogappregistry_application.test"
	attributeGroupResourceName := "aws_servicecatalogappregistry_attribute_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceCatalogAppRegistryEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttributeGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAttributeGroupAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttributeGroupAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrApplicationID, applicationResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "attribute_group_id", attributeGroupResourceName, names.AttrID),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccCheckAttributeGroupAssociationImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "attribute_group_id",
			},
		},
	})
}

func TestAccServiceCatalogAppRegistryAttributeGroupAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_servicecatalogappregistry_attribute_group_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceCatalogAppRegistryEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttributeGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAttributeGroupAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttributeGroupAssociationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfservicecatalogappregistry.ResourceAttributeGroupAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAttributeGroupAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ServiceCatalogAppRegistryClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalogappregistry_attribute_group_association" {
				continue
			}

			applicationId := rs.Primary.Attributes[names.AttrApplicationID]
			attributeGroupId := rs.Primary.Attributes["attribute_group_id"]

			_, err := tfservicecatalogappregistry.FindAttributeGroupAssociationByTwoPartKey(ctx, conn, applicationId, attributeGroupId)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingDestroyed, tfservicecatalogappregistry.ResNameAttributeGroupAssociation, attributeGroupId, err)
			}

			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingDestroyed, tfservicecatalogappregistry.ResNameAttributeGroupAssociation, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAttributeGroupAssociationExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, tfservicecatalogappregistry.ResNameAttributeGroupAssociation, name, errors.New("not found"))
		}

		applicationId := rs.Primary.Attributes[names.AttrApplicationID]
		attributeGroupId := rs.Primary.Attributes["attribute_group_id"]
		if applicationId == "" {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, tfservicecatalogappregistry.ResNameAttributeGroupAssociation, name, errors.New("application_id not set"))
		}
		if attributeGroupId == "" {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, tfservicecatalogappregistry.ResNameAttributeGroupAssociation, name, errors.New("attribute_group_id not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).ServiceCatalogAppRegistryClient(ctx)
		_, err := tfservicecatalogappregistry.FindAttributeGroupAssociationByTwoPartKey(ctx, conn, applicationId, attributeGroupId)
		if err != nil {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, tfservicecatalogappregistry.ResNameAttributeGroupAssociation, attributeGroupId, err)
		}

		return nil
	}
}

func testAccCheckAttributeGroupAssociationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes[names.AttrApplicationID], rs.Primary.Attributes["attribute_group_id"]), nil
	}
}

func testAccAttributeGroupAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalogappregistry_application" "test" {
  name = %[1]q
}

resource "aws_servicecatalogappregistry_attribute_group" "test" {
  name = %[1]q

  attributes = jsonencode({
    a = "1"
    b = "2"
  })
}

resource "aws_servicecatalogappregistry_attribute_group_association" "test" {
  application_id     = aws_servicecatalogappregistry_application.test.id
  attribute_group_id = aws_servicecatalogappregistry_attribute_group.test.id
}
`, rName)
}
