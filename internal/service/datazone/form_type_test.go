// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package datazone_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/datazone"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdatazone "github.com/hashicorp/terraform-provider-aws/internal/service/datazone"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataZoneFormType_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var formtype datazone.GetFormTypeOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resourceName := "aws_datazone_form_type.test"
	domainName := "aws_datazone_domain.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFormTypeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFormTypeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFormTypeExists(ctx, t, resourceName, &formtype),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "desc"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_identifier", domainName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "model.#"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "SageMakerModelFormType"),
					resource.TestCheckResourceAttrSet(resourceName, "revision"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "DISABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "imports.#"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateIdFunc:                    testAccAuthorizerImportStateUserProfileFunc(resourceName),
				ImportStateVerifyIgnore:              []string{"model"},
			},
		},
	})
}

func TestAccDataZoneFormType_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var formtype datazone.GetFormTypeOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resourceName := "aws_datazone_form_type.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFormTypeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFormTypeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFormTypeExists(ctx, t, resourceName, &formtype),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdatazone.ResourceFormType, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFormTypeDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DataZoneClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datazone_form_type" {
				continue
			}

			_, err := tfdatazone.FindFormTypeByID(ctx, conn, rs.Primary.Attributes["domain_identifier"], rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes["revision"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameFormType, rs.Primary.ID, err)
			}

			return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameFormType, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckFormTypeExists(ctx context.Context, t *testing.T, name string, formtype *datazone.GetFormTypeOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameFormType, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameFormType, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).DataZoneClient(ctx)

		resp, err := tfdatazone.FindFormTypeByID(ctx, conn, rs.Primary.Attributes["domain_identifier"], rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes["revision"])

		if err != nil {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameFormType, rs.Primary.ID, err)
		}

		*formtype = *resp

		return nil
	}
}

func testAccAuthorizerImportStateUserProfileFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return strings.Join([]string{rs.Primary.Attributes["domain_identifier"], rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes["revision"]}, ","), nil
	}
}

func testAccFormTypeConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_basic(rName), `
resource "aws_datazone_form_type" "test" {
  description               = "desc"
  name                      = "SageMakerModelFormType"
  domain_identifier         = aws_datazone_domain.test.id
  owning_project_identifier = aws_datazone_project.test.id
  status                    = "DISABLED"
  model {
    smithy = <<EOF
	structure SageMakerModelFormType {
			@required
			@amazon.datazone#searchable
			modelName: String

			@required
			modelArn: String

			@required
			creationTime: String
			}
		EOF
  }
}
`)
}
