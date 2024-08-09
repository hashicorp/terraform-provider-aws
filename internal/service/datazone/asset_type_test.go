// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/datazone"
	"github.com/aws/aws-sdk-go-v2/service/datazone/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfdatazone "github.com/hashicorp/terraform-provider-aws/internal/service/datazone"
)

func TestAccDataZoneAssetType_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var assettype datazone.GetAssetTypeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_asset_type.test"
	projectName := "aws_datazone_project.test"
	domainName := "aws_datazone_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssetTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssetTypeConfig_basic(rName, pName, dName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssetTypeExists(ctx, resourceName, &assettype),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttrSet(resourceName, "revision"),
					resource.TestCheckResourceAttr(resourceName, "description", "desc"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "domain_id", domainName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "owning_project_id", projectName, "id"),
					resource.TestCheckResourceAttr(resourceName, "description", "origin_domain_id"),
					/// ???
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAuthorizerAssetTypeImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccDataZoneAssetType_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var assettype datazone.GetAssetTypeOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_asset_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssetTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssetTypeConfig_basic(rName, pName, dName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssetTypeExists(ctx, resourceName, &assettype),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdatazone.ResourceAssetType, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAssetTypeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datazone_asset_type" {
				continue
			}

			_, err := tfdatazone.FindAssetTypeByID(ctx, conn, rs.Primary.Attributes["domain_id"], rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes["revision"])
			if errs.IsA[*types.ResourceNotFoundException](err) || errs.IsA[*types.AccessDeniedException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameAssetType, rs.Primary.ID, err)
			}

			return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameAssetType, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAssetTypeExists(ctx context.Context, name string, assettype *datazone.GetAssetTypeOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameAssetType, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameAssetType, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)
		resp, err := tfdatazone.FindAssetTypeByID(ctx, conn, rs.Primary.Attributes["domain_id"], rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes["revision"])

		if err != nil {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameAssetType, rs.Primary.ID, err)
		}

		*assettype = *resp

		return nil
	}
}

func testAccAuthorizerAssetTypeImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return strings.Join([]string{rs.Primary.Attributes["domain_identifier"], rs.Primary.Attributes["name"], rs.Primary.Attributes["revision"]}, ","), nil
	}
}

func testAccAssetTypeConfig_basic(rName, pName, dName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_basic(pName, dName), fmt.Sprint(`
resource "aws_datazone_form_type" "test" {
  description               = "desc"
  name                      = "SageMakerModelFormType"
  domain_identifier         = aws_datazone_domain.test.id
  owning_project_identifier = aws_datazone_project.test.id
  status                    = "ENABLED"
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

resource "aws_datazone_asset_type" "test" {
	description = "desc"
	domain_identifier = aws_datazone_domain.test.id
	name = "hi"
	owning_project_identifier = aws_datazone_project.test.id
	forms_input = {
		"first" = {
				required = true
				type_identifier = aws_datazone_form_type.test.name
				type_revision = aws_datazone_form_type.test.revision
		}
	}
}
`))
}
