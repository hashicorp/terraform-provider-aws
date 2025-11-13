// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlueFederatedCatalog_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var federatedcatalog glue.GetCatalogOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_federated_catalog.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GlueEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFederatedCatalogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFederatedCatalogConfig_s3Tables(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFederatedCatalogExists(ctx, resourceName, &federatedcatalog),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "s3tablescatalog"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test S3 Tables federated catalog"),
					resource.TestCheckResourceAttr(resourceName, "federated_catalog.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "federated_catalog.0.identifier"),
					resource.TestCheckResourceAttr(resourceName, "federated_catalog.0.connection_name", "aws:s3tables"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "glue", regexache.MustCompile(`catalog/.+$`)),
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

func TestAccGlueFederatedCatalog_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var federatedcatalog glue.GetCatalogOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_federated_catalog.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GlueEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFederatedCatalogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFederatedCatalogConfig_s3Tables(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFederatedCatalogExists(ctx, resourceName, &federatedcatalog),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfglue.ResourceFederatedCatalog, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGlueFederatedCatalog_catalogProperties(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var federatedcatalog glue.GetCatalogOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_federated_catalog.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GlueEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFederatedCatalogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFederatedCatalogConfig_s3Tables(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFederatedCatalogExists(ctx, resourceName, &federatedcatalog),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "s3tablescatalog"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Test S3 Tables federated catalog"),
					resource.TestCheckResourceAttr(resourceName, "federated_catalog.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "federated_catalog.0.identifier"),
					resource.TestCheckResourceAttr(resourceName, "federated_catalog.0.connection_name", "aws:s3tables"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "glue", regexache.MustCompile(`catalog/.+$`)),
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

func TestAccGlueFederatedCatalog_configurationError(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GlueEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFederatedCatalogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccFederatedCatalogConfig_missingConfiguration(rName),
				ExpectError: regexache.MustCompile("Missing Required Configuration"),
			},
		},
	})
}

func TestAccGlueFederatedCatalog_catalogPropertiesDisappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var federatedcatalog glue.GetCatalogOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_federated_catalog.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GlueEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFederatedCatalogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFederatedCatalogConfig_s3Tables(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFederatedCatalogExists(ctx, resourceName, &federatedcatalog),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfglue.ResourceFederatedCatalog, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFederatedCatalogDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_federated_catalog" {
				continue
			}

			_, err := tfglue.FindFederatedCatalogByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return smarterr.NewError(err)
			}

			return smarterr.NewError(errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckFederatedCatalogExists(ctx context.Context, name string, federatedcatalog *glue.GetCatalogOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return smarterr.NewError(errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return smarterr.NewError(errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

		resp, err := tfglue.FindFederatedCatalogByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return smarterr.NewError(err)
		}

		*federatedcatalog = glue.GetCatalogOutput{
			Catalog: resp,
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

	if conn == nil {
		t.Fatal("Glue client is not configured")
	}
}

func testAccFederatedCatalogConfig_missingConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_federated_catalog" "test" {
  name        = %[1]q
  description = "Test federated catalog without required configuration"
}
`, rName)
}

func testAccFederatedCatalogConfig_s3Tables(rName string) string {
	return acctest.ConfigCompose(
		testAccFederatedCatalogConfig_s3TablesBase(rName), `
resource "aws_glue_federated_catalog" "test" {
  name        = "s3tablescatalog"
  description = "Test S3 Tables federated catalog"

  federated_catalog {
    identifier      = "arn:${data.aws_partition.current.partition}:s3tables:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:bucket/*"
    connection_name = "aws:s3tables"
  }

  depends_on = [aws_lakeformation_resource.test]
}
`,
	)
}

func testAccFederatedCatalogConfig_s3TablesBase(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

# IAM role for Lake Formation data access
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "lakeformation.amazonaws.com"
        }
        Action = [
          "sts:AssumeRole",
          "sts:SetSourceIdentity",
          "sts:SetContext"
        ]
      }
    ]
  })
}

# IAM policy for S3 Tables permissions
resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "LakeFormationPermissionsForS3ListTableBucket"
        Effect = "Allow"
        Action = [
          "s3tables:ListTableBuckets"
        ]
        Resource = ["*"]
      },
      {
        Sid    = "LakeFormationDataAccessPermissionsForS3TableBucket"
        Effect = "Allow"
        Action = [
          "s3tables:CreateTableBucket",
          "s3tables:GetTableBucket",
          "s3tables:CreateNamespace",
          "s3tables:GetNamespace",
          "s3tables:ListNamespaces",
          "s3tables:DeleteNamespace",
          "s3tables:DeleteTableBucket",
          "s3tables:CreateTable",
          "s3tables:DeleteTable",
          "s3tables:GetTable",
          "s3tables:ListTables",
          "s3tables:RenameTable",
          "s3tables:UpdateTableMetadataLocation",
          "s3tables:GetTableMetadataLocation",
          "s3tables:GetTableData",
          "s3tables:PutTableData"
        ]
        Resource = [
          "arn:${data.aws_partition.current.partition}:s3tables:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:bucket/*"
        ]
      }
    ]
  })
}

# Register the S3 Tables location with Lake Formation
resource "aws_lakeformation_resource" "test" {
  arn      = "arn:${data.aws_partition.current.partition}:s3tables:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:bucket/*"
  role_arn = aws_iam_role.test.arn
}
`, rName)
}
