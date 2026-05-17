// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
	providerslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestReadLFTagID(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         string
		catalogID   string
		tagKey      string
		expectError bool
	}

	tests := map[string]testCase{
		"empty_string": {
			expectError: true,
		},
		"invalid_id": {
			val:         "test",
			expectError: true,
		},
		"valid_key_simple": {
			val:       "123344556:tagKey",
			catalogID: "123344556",
			tagKey:    "tagKey",
		},
		"valid_key_complex": {
			val:       "123344556:keyPrefix:tagKey",
			catalogID: "123344556",
			tagKey:    "keyPrefix:tagKey",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			catalogID, tagKey, err := tflakeformation.LFTagParseResourceID(test.val)

			if err == nil && test.expectError {
				t.Fatal("expected error")
			}

			if err != nil && !test.expectError {
				t.Fatalf("got unexpected error: %s", err)
			}

			if test.catalogID != catalogID || test.tagKey != tagKey {
				t.Fatalf("expected catalogID (%s), tagKey (%s), got catalogID (%s), tagKey (%s)", test.catalogID, test.tagKey, catalogID, tagKey)
			}
		})
	}
}

func testAccLFTag_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lakeformation_lf_tag.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLFTagsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLFTagConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, rName),
					resource.TestCheckResourceAttr(resourceName, "values.0", names.AttrValue),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrCatalogID),
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

func testAccLFTag_TagKey_complex(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lakeformation_lf_tag.test"
	rName := fmt.Sprintf("%s:%s", acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "subKey")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLFTagsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLFTagConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, rName),
					resource.TestCheckResourceAttr(resourceName, "values.0", names.AttrValue),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrCatalogID),
				),
			},
		},
	})
}

func testAccLFTag_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lakeformation_lf_tag.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLFTagsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLFTagConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tflakeformation.ResourceLFTag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccLFTag_Values(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lakeformation_lf_tag.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLFTagsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:  testAccLFTagConfig_values(rName, []string{acctest.CtValue1, acctest.CtValue2}),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, rName),
					resource.TestCheckResourceAttr(resourceName, "values.0", acctest.CtValue1),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrCatalogID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test an update that adds, removes and retains a tag value
				Config: testAccLFTagConfig_values(rName, []string{acctest.CtValue1, "value3"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, rName),
					resource.TestCheckResourceAttr(resourceName, "values.0", acctest.CtValue1),
					resource.TestCheckTypeSetElemAttr(resourceName, "values.*", "value3"),
					testAccCheckLFTagValuesLen(ctx, t, resourceName, 2),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrCatalogID),
				),
			},
		},
	})
}

func testAccLFTag_Values_overFifty(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lakeformation_lf_tag.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	generatedValues := generateLFTagValueList(1, 52)
	generatedValues2 := slices.Clone(generatedValues)
	generatedValues2 = append(generatedValues2, generateLFTagValueList(53, 120)...)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLFTagsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLFTagConfig_values(rName, generatedValues),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, rName),
					resource.TestCheckResourceAttr(resourceName, "values.0", acctest.CtValue1),
					testAccCheckLFTagValuesLen(ctx, t, resourceName, len(generatedValues)),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrCatalogID),
				),
			},
			{
				Config: testAccLFTagConfig_values(rName, generatedValues2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, rName),
					resource.TestCheckResourceAttr(resourceName, "values.0", acctest.CtValue1),
					testAccCheckLFTagValuesLen(ctx, t, resourceName, len(generatedValues2)),
					resource.TestCheckTypeSetElemAttr(resourceName, "values.*", "value59"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrCatalogID),
				),
			},
			{
				Config: testAccLFTagConfig_values(rName, providerslices.RemoveAll(generatedValues, "value36")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrKey, rName),
					resource.TestCheckResourceAttr(resourceName, "values.0", acctest.CtValue1),
					testAccCheckLFTagValuesLen(ctx, t, resourceName, len(generatedValues)-1),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrCatalogID),
				),
			},
		},
	})
}

func testAccCheckLFTagsDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LakeFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lakeformation_lf_tag" {
				continue
			}

			catalogID, tagKey, err := tflakeformation.LFTagParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			input := &lakeformation.GetLFTagInput{
				CatalogId: aws.String(catalogID),
				TagKey:    aws.String(tagKey),
			}

			if _, err := conn.GetLFTag(ctx, input); err != nil {
				if errs.IsA[*awstypes.EntityNotFoundException](err) || errs.IsA[*awstypes.AccessDeniedException](err) {
					continue
				}

				return err
			}
			return fmt.Errorf("Lake Formation LF-Tag (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLFTagExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		catalogID, tagKey, err := tflakeformation.LFTagParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &lakeformation.GetLFTagInput{
			CatalogId: aws.String(catalogID),
			TagKey:    aws.String(tagKey),
		}

		conn := acctest.ProviderMeta(ctx, t).LakeFormationClient(ctx)
		_, err = conn.GetLFTag(ctx, input)

		return err
	}
}

func testAccCheckLFTagValuesLen(ctx context.Context, t *testing.T, name string, expectedLength int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		catalogID, tagKey, err := tflakeformation.LFTagParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &lakeformation.GetLFTagInput{
			CatalogId: aws.String(catalogID),
			TagKey:    aws.String(tagKey),
		}

		conn := acctest.ProviderMeta(ctx, t).LakeFormationClient(ctx)
		output, err := conn.GetLFTag(ctx, input)

		if len(output.TagValues) != expectedLength {
			return fmt.Errorf("expected %d values, got %d", expectedLength, len(output.TagValues))
		}

		return err
	}
}

func testAccLFTagConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_lf_tag" "test" {
  key    = %[1]q
  values = ["value"]
  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName)
}

func testAccLFTagConfig_values(rName string, values []string) string {
	quotedValues := make([]string, len(values))
	for i, v := range values {
		quotedValues[i] = strconv.Quote(v)
	}

	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [data.aws_iam_session_context.current.issuer_arn]
}

resource "aws_lakeformation_lf_tag" "test" {
  key    = %[1]q
  values = [%[2]s]
  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName, strings.Join(quotedValues, ","))
}

func generateLFTagValueList(start, end int) []string {
	var out []string
	for i := start; i <= end; i++ {
		out = append(out, fmt.Sprintf("value%d", i))
	}

	return out
}
