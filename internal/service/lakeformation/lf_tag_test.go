package lakeformation_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
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
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			catalogID, tagKey, err := tflakeformation.ReadLFTagID(test.val)

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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, lakeformation.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLFTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLFTagConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", rName),
					resource.TestCheckResourceAttr(resourceName, "values.0", "value"),
					acctest.CheckResourceAttrAccountID(resourceName, "catalog_id"),
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

<<<<<<< HEAD
func testAccLFTag_many_values(t *testing.T) {
	resourceName := "aws_lakeformation_lf_tag.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:               acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLFTagsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLFTagConfig_manyvalues(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", rName),
					resource.TestCheckResourceAttr(resourceName, "values.0", "value1"),
					testAccCheckLFTagValuesLen(resourceName, 52),
					acctest.CheckResourceAttrAccountID(resourceName, "catalog_id"),
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

func testAccLFTag_disappears(t *testing.T) {
=======
func testAccLFTag_TagKey_complex(t *testing.T) {
	ctx := acctest.Context(t)
>>>>>>> main
	resourceName := "aws_lakeformation_lf_tag.test"
	rName := fmt.Sprintf("%s:%s", sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "subKey")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, lakeformation.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLFTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLFTagConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", rName),
					resource.TestCheckResourceAttr(resourceName, "values.0", "value"),
					acctest.CheckResourceAttrAccountID(resourceName, "catalog_id"),
				),
			},
		},
	})
}

func testAccLFTag_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lakeformation_lf_tag.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, lakeformation.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLFTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLFTagConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflakeformation.ResourceLFTag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccLFTag_values(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_lakeformation_lf_tag.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, lakeformation.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLFTagsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccLFTagConfig_values(rName, []string{"value1", "value2"}),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", rName),
					resource.TestCheckResourceAttr(resourceName, "values.0", "value1"),
					acctest.CheckResourceAttrAccountID(resourceName, "catalog_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test an update that adds, removes and retains a tag value
				Config: testAccLFTagConfig_values(rName, []string{"value1", "value3", "value4", "value5", "value6", "value7", "value8", "value9", "value10", "value11", "value12", "value13", "value14", "value15", "value16", "value17", "value18", "value19", "value20", "value21", "value22", "value23", "value24", "value25", "value26", "value27", "value28", "value29", "value30", "value31", "value32", "value33", "value34", "value35", "value36", "value37", "value38", "value39", "value40", "value41", "value42", "value43", "value44", "value45", "value46", "value47", "value48", "value49", "value50", "value51", "value52", "value53", "value54", "value55", "value56", "value57", "value58", "value59", "value60", "value61", "value62", "value63", "value64", "value65", "value66", "value67", "value68", "value69", "value70", "value71", "value72", "value73", "value74", "value75", "value76", "value77", "value78", "value79", "value80", "value81", "value82", "value83", "value84", "value85", "value86", "value87", "value88", "value89", "value90", "value91", "value92", "value93", "value94", "value95", "value96", "value97", "value98", "value99", "value100", "value101", "value102", "value103", "value104", "value105", "value106", "value107", "value108", "value109", "value110", "value111", "value112", "value113", "value114", "value115", "value116", "value117", "value118", "value119", "value120", "value121", "value122", "value123", "value124", "value125", "value126", "value127", "value128", "value129", "value130", "value131", "value132", "value133", "value134", "value135", "value136", "value137", "value138", "value139", "value140", "value141"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", rName),
					resource.TestCheckResourceAttr(resourceName, "values.0", "value1"),
					testAccCheckLFTagValuesLen(resourceName, 140),
					acctest.CheckResourceAttrAccountID(resourceName, "catalog_id"),
				),
			},
			{
				// Test an update that adds, removes and retains many values
				Config: testAccLFTagConfig_values(rName, []string{"value1", "value3", "value4", "value5", "value6", "value7", "value8", "value9", "value10", "value160", "value161", "value162", "value163", "value164", "value165", "value166", "value167", "value168", "value169", "value170", "value171", "value172", "value173", "value174", "value175", "value176", "value177", "value178", "value179", "value180", "value181", "value182", "value183", "value184", "value185", "value186", "value187", "value188", "value189", "value190", "value191", "value192", "value193", "value194", "value195", "value196", "value197", "value198", "value199", "value200", "value201", "value202", "value203", "value204", "value205", "value206", "value207", "value208", "value209", "value210", "value211", "value212", "value213", "value214", "value215", "value216", "value217", "value218", "value219", "value220"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", rName),
					resource.TestCheckResourceAttr(resourceName, "values.0", "value1"),
					testAccCheckLFTagValuesLen(resourceName, 70),
					acctest.CheckResourceAttrAccountID(resourceName, "catalog_id"),
				),
			},
		},
	})
}

func testAccCheckLFTagsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lakeformation_lf_tag" {
				continue
			}

			catalogID, tagKey, err := tflakeformation.ReadLFTagID(rs.Primary.ID)
			if err != nil {
				return err
			}

			input := &lakeformation.GetLFTagInput{
				CatalogId: aws.String(catalogID),
				TagKey:    aws.String(tagKey),
			}

			if _, err := conn.GetLFTagWithContext(ctx, input); err != nil {
				if tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeEntityNotFoundException) {
					continue
				}
				// If the lake formation admin has been revoked, there will be access denied instead of entity not found
				if tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeAccessDeniedException) {
					continue
				}
				return err
			}
			return fmt.Errorf("Lake Formation LF-Tag (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLFTagExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		catalogID, tagKey, err := tflakeformation.ReadLFTagID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &lakeformation.GetLFTagInput{
			CatalogId: aws.String(catalogID),
			TagKey:    aws.String(tagKey),
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationConn()
		_, err = conn.GetLFTagWithContext(ctx, input)

		return err
	}
}

func testAccCheckLFTagValuesLen(name string, expected_len int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		catalogID, tagKey, err := tflakeformation.ReadLFTagID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &lakeformation.GetLFTagInput{
			CatalogId: aws.String(catalogID),
			TagKey:    aws.String(tagKey),
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationConn
		output, err := conn.GetLFTag(input)

		if len(output.TagValues) != expected_len {
			return fmt.Errorf("expected %d values, got %d", expected_len, len(output.TagValues))
		}

		if err != nil {
			return err
		}

		return nil
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

func testAccLFTagConfig_manyvalues(rName string) string {
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
  values = ["value1", "value2", "value3", "value4", "value5", "value6", "value7", "value8", "value9", "value10", "value11", "value12", "value13", "value14", "value15", "value16", "value17", "value18", "value19", "value20", "value21", "value22", "value23", "value24", "value25", "value26", "value27", "value28", "value29", "value30", "value31", "value32", "value33", "value34", "value35", "value36", "value37", "value38", "value39", "value40", "value41", "value42", "value43", "value44", "value45", "value46", "value47", "value48", "value49", "value50", "value51", "value52"]
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
