package lakeformation_test

import (
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

func testAccLFTag_basic(t *testing.T) {
	resourceName := "aws_lakeformation_lf_tag.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLFTagsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLFTagConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExists(resourceName),
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

func testAccLFTag_disappears(t *testing.T) {
	resourceName := "aws_lakeformation_lf_tag.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLFTagsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLFTagConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tflakeformation.ResourceLFTag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccLFTag_values(t *testing.T) {
	resourceName := "aws_lakeformation_lf_tag.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(lakeformation.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, lakeformation.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLFTagsDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccLFTagConfig_values(rName, []string{"value1", "value2"}),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExists(resourceName),
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
				Config: testAccLFTagConfig_values(rName, []string{"value1", "value3"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLFTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", rName),
					resource.TestCheckResourceAttr(resourceName, "values.0", "value1"),
					resource.TestCheckResourceAttr(resourceName, "values.1", "value3"),
					acctest.CheckResourceAttrAccountID(resourceName, "catalog_id"),
				),
			},
		},
	})
}

func testAccCheckLFTagsDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LakeFormationConn

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

		if _, err := conn.GetLFTag(input); err != nil {
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

func testAccCheckLFTagExists(name string) resource.TestCheckFunc {
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
		_, err = conn.GetLFTag(input)

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
  values = [%s]
  # for consistency, ensure that admins are setup before testing
  depends_on = [aws_lakeformation_data_lake_settings.test]
}
`, rName, strings.Join(quotedValues, ","))
}
