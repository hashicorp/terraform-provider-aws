// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3tables_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	tfs3tables "github.com/hashicorp/terraform-provider-aws/internal/service/s3tables"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3TablesNamespace_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var namespace s3tables.GetNamespaceOutput
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := strings.ReplaceAll(bucketName, "-", "_")
	resourceName := "aws_s3tables_namespace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfig_basic(rName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNamespaceExists(ctx, resourceName, &namespace),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "created_by"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamespace, rName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerAccountID),
					resource.TestCheckResourceAttrPair(resourceName, "table_bucket_arn", "aws_s3tables_table_bucket.test", names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ";", "table_bucket_arn", names.AttrNamespace),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrNamespace,
			},
		},
	})
}

func TestAccS3TablesNamespace_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var namespace s3tables.GetNamespaceOutput
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := strings.ReplaceAll(bucketName, "-", "_")
	resourceName := "aws_s3tables_namespace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfig_basic(rName, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNamespaceExists(ctx, resourceName, &namespace),
					acctest.CheckFrameworkResourceDisappearsWithStateFunc(ctx, acctest.Provider, tfs3tables.ResourceNamespace, resourceName, namespaceDisappearsStateFunc),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckNamespaceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3TablesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3tables_namespace" {
				continue
			}

			_, err := tfs3tables.FindNamespaceByTwoPartKey(ctx, conn, rs.Primary.Attributes["table_bucket_arn"], rs.Primary.Attributes[names.AttrNamespace])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Tables Namespace %s still exists", rs.Primary.Attributes[names.AttrNamespace])
		}

		return nil
	}
}

func testAccCheckNamespaceExists(ctx context.Context, n string, v *s3tables.GetNamespaceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3TablesClient(ctx)

		output, err := tfs3tables.FindNamespaceByTwoPartKey(ctx, conn, rs.Primary.Attributes["table_bucket_arn"], rs.Primary.Attributes[names.AttrNamespace])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func namespaceDisappearsStateFunc(ctx context.Context, state *tfsdk.State, is *terraform.InstanceState) error {
	v, ok := is.Attributes[names.AttrNamespace]
	if !ok {
		return errors.New(`Identifying attribute "namespace" not defined`)
	}

	if err := fwdiag.DiagnosticsError(state.SetAttribute(ctx, path.Root(names.AttrNamespace), v)); err != nil {
		return err
	}

	v, ok = is.Attributes["table_bucket_arn"]
	if !ok {
		return errors.New(`Identifying attribute "table_bucket_arn" not defined`)
	}

	if err := fwdiag.DiagnosticsError(state.SetAttribute(ctx, path.Root("table_bucket_arn"), v)); err != nil {
		return err
	}

	return nil
}

func testAccNamespaceConfig_basic(rName, bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3tables_namespace" "test" {
  namespace        = %[1]q
  table_bucket_arn = aws_s3tables_table_bucket.test.arn
}

resource "aws_s3tables_table_bucket" "test" {
  name = %[2]q
}
`, rName, bucketName)
}
