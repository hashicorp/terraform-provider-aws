// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalogappregistry_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	flex2 "github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfservicecatalogappregistry "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalogappregistry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceCatalogAppRegistryApplicationAttributeGroupAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "Some description"
	resourceName := "aws_servicecatalogappregistry_application_attribute_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceCatalogAppRegistryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationAttributeGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationAttributeGroupAssociationConfig_basic(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationAttributeGroupAssociationExists(ctx, resourceName),
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

func TestAccServiceCatalogAppRegistryApplicationAttributeGroupAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "Some description"
	resourceName := "aws_servicecatalogappregistry_application_attribute_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceCatalogAppRegistryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationAttributeGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationAttributeGroupAssociationConfig_basic(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationAttributeGroupAssociationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfservicecatalogappregistry.ResourceApplicationAttributeGroupAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckApplicationAttributeGroupAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogAppRegistryClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalogappregistry_application_attribute_group_association" {
				continue
			}

			parts, err := flex2.ExpandResourceId(rs.Primary.ID, 2, false)
			applicationId := parts[0]
			attributeGroupId := parts[1]

			if err != nil {
				return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, tfservicecatalogappregistry.ResNameApplicationAttributeGroupAssociation, rs.Primary.ID, err)
			}

			resp, err := conn.ListAssociatedAttributeGroups(ctx, &servicecatalogappregistry.ListAssociatedAttributeGroupsInput{
				Application: aws.String(applicationId),
			})

			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}

			if err != nil {
				return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, tfservicecatalogappregistry.ResNameApplicationAttributeGroupAssociation, rs.Primary.ID, errors.New("error listing associations"))
			}

			for _, groupId := range resp.AttributeGroups {
				if groupId == attributeGroupId {
					return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingDestroyed, tfservicecatalogappregistry.ResNameApplicationAttributeGroupAssociation, rs.Primary.ID, errors.New("not destroyed"))
				}
			}
		}

		return nil
	}
}

func testAccCheckApplicationAttributeGroupAssociationExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, tfservicecatalogappregistry.ResNameApplicationAttributeGroupAssociation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, tfservicecatalogappregistry.ResNameApplicationAttributeGroupAssociation, name, errors.New("not set"))
		}

		parts, err := flex2.ExpandResourceId(rs.Primary.ID, 2, false)
		applicationId := parts[0]
		attributeGroupId := parts[1]

		if err != nil {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, tfservicecatalogappregistry.ResNameApplicationAttributeGroupAssociation, rs.Primary.ID, err)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogAppRegistryClient(ctx)
		resp, err := conn.ListAssociatedAttributeGroups(ctx, &servicecatalogappregistry.ListAssociatedAttributeGroupsInput{
			Application: aws.String(applicationId),
		})

		if err != nil {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, tfservicecatalogappregistry.ResNameApplicationAttributeGroupAssociation, rs.Primary.ID, errors.New("error listing attribute groups"))
		}

		for _, groupId := range resp.AttributeGroups {
			if groupId == attributeGroupId {
				return nil
			}
		}

		return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, tfservicecatalogappregistry.ResNameApplicationAttributeGroupAssociation, rs.Primary.ID, errors.New("missing"))
	}
}

func testAccApplicationAttributeGroupAssociationConfig_basic(rName, description string) string {
	return acctest.ConfigCompose(
		testAccAttributeGroupConfig_basic(rName, description),
		testAccApplicationConfig_description(rName, description),
		`
resource "aws_servicecatalogappregistry_application_attribute_group_association" "test" {
  application_id     = aws_servicecatalogappregistry_application.test.id
  attribute_group_id = aws_servicecatalogappregistry_attribute_group.test.id
}
`)
}
