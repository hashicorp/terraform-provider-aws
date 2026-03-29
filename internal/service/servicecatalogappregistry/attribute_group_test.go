// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicecatalogappregistry_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalogappregistry/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfservicecatalogappregistry "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalogappregistry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceCatalogAppRegistryAttributeGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var attributegroup servicecatalogappregistry.GetAttributeGroupOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_servicecatalogappregistry_attribute_group.test"
	description := "Simple Description"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceCatalogAppRegistryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttributeGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAttributeGroupConfig_basic(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttributeGroupExists(ctx, t, resourceName, &attributegroup),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "servicecatalog", regexache.MustCompile(`/attribute-groups/+.`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
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

func TestAccServiceCatalogAppRegistryAttributeGroup_update(t *testing.T) {
	ctx := acctest.Context(t)

	var attributegroup1, attributegroup2 servicecatalogappregistry.GetAttributeGroupOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_servicecatalogappregistry_attribute_group.test"
	description := "Simple Description"
	expectJsonV1 := `{"a":"1","b":"2"}`
	expectJsonV2 := `{"b":"3","c":"4"}`

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceCatalogAppRegistryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttributeGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAttributeGroupConfig_basic(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttributeGroupExists(ctx, t, resourceName, &attributegroup1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "servicecatalog", regexache.MustCompile(`/attribute-groups/+.`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, names.AttrAttributes, expectJsonV1),
				),
			},
			{
				Config: testAccAttributeGroupConfig_update(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttributeGroupExists(ctx, t, resourceName, &attributegroup2),
					testAccCheckAttributeGroupNotRecreated(&attributegroup1, &attributegroup2),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "servicecatalog", regexache.MustCompile(`/attribute-groups/+.`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, names.AttrAttributes, expectJsonV2),
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

func TestAccServiceCatalogAppRegistryAttributeGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var attributegroup servicecatalogappregistry.GetAttributeGroupOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	description := "Simple Description"
	resourceName := "aws_servicecatalogappregistry_attribute_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceCatalogAppRegistryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogAppRegistryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAttributeGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAttributeGroupConfig_basic(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAttributeGroupExists(ctx, t, resourceName, &attributegroup),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfservicecatalogappregistry.ResourceAttributeGroup, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAttributeGroupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ServiceCatalogAppRegistryClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalogappregistry_attribute_group" {
				continue
			}

			_, err := tfservicecatalogappregistry.FindAttributeGroupByID(ctx, conn, rs.Primary.ID)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingDestroyed, tfservicecatalogappregistry.ResNameAttributeGroup, rs.Primary.ID, err)
			}

			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingDestroyed, tfservicecatalogappregistry.ResNameAttributeGroup, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAttributeGroupExists(ctx context.Context, t *testing.T, name string, attributegroup *servicecatalogappregistry.GetAttributeGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, tfservicecatalogappregistry.ResNameAttributeGroup, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, tfservicecatalogappregistry.ResNameAttributeGroup, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).ServiceCatalogAppRegistryClient(ctx)
		resp, err := tfservicecatalogappregistry.FindAttributeGroupByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingExistence, tfservicecatalogappregistry.ResNameAttributeGroup, rs.Primary.ID, err)
		}

		*attributegroup = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ServiceCatalogAppRegistryClient(ctx)

	input := &servicecatalogappregistry.ListAttributeGroupsInput{}
	_, err := conn.ListAttributeGroups(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckAttributeGroupNotRecreated(before, after *servicecatalogappregistry.GetAttributeGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Id), aws.ToString(after.Id); before != after {
			return create.Error(names.ServiceCatalogAppRegistry, create.ErrActionCheckingNotRecreated, tfservicecatalogappregistry.ResNameAttributeGroup, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccAttributeGroupConfig_basic(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalogappregistry_attribute_group" "test" {
  name        = %[1]q
  description = %[2]q
  attributes = jsonencode({
    a = "1"
    b = "2"
  })
}
`, rName, description)
}

func testAccAttributeGroupConfig_update(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalogappregistry_attribute_group" "test" {
  name        = %[1]q
  description = %[2]q
  attributes = jsonencode({
    b = "3"
    c = "4"
  })
}
`, rName, description)
}
