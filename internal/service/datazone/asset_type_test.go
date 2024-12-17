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
	tfdatazone "github.com/hashicorp/terraform-provider-aws/internal/service/datazone"
	"github.com/hashicorp/terraform-provider-aws/names"
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
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttrSet(resourceName, "revision"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, strings.ReplaceAll(rName, "-", "_")),
					resource.TestCheckResourceAttrPair(resourceName, "domain_identifier", domainName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "owning_project_identifier", projectName, names.AttrID),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateIdFunc:                    testAccAssetTypeImportStateIdFunc(resourceName),
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

			_, err := tfdatazone.FindAssetTypeByID(ctx, conn, rs.Primary.Attributes["domain_identifier"], rs.Primary.Attributes[names.AttrName])
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
		resp, err := tfdatazone.FindAssetTypeByID(ctx, conn, rs.Primary.Attributes["domain_identifier"], rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameAssetType, rs.Primary.ID, err)
		}

		*assettype = *resp

		return nil
	}
}

func testAccAssetTypeImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return strings.Join([]string{rs.Primary.Attributes["domain_identifier"], rs.Primary.Attributes[names.AttrName]}, ","), nil
	}
}

func testAccAssetTypeConfig_basic(rName, pName, dName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_basic(pName, dName), fmt.Sprintf(`
resource "aws_datazone_asset_type" "test" {
  description               = %[1]q
  domain_identifier         = aws_datazone_domain.test.id
  name                      = %[2]q
  owning_project_identifier = aws_datazone_project.test.id
}
`, rName, strings.ReplaceAll(rName, "-", "_")))
}
