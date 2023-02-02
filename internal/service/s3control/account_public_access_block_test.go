package s3control_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// S3 account-level settings must run serialized
// for TeamCity environment
func TestAccS3ControlAccountPublicAccessBlock_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"PublicAccessBlock": {
			"basic":                 testAccAccountPublicAccessBlock_basic,
			"disappears":            testAccAccountPublicAccessBlock_disappears,
			"AccountId":             testAccAccountPublicAccessBlock_AccountID,
			"BlockPublicAcls":       testAccAccountPublicAccessBlock_BlockPublicACLs,
			"BlockPublicPolicy":     testAccAccountPublicAccessBlock_BlockPublicPolicy,
			"IgnorePublicAcls":      testAccAccountPublicAccessBlock_IgnorePublicACLs,
			"RestrictPublicBuckets": testAccAccountPublicAccessBlock_RestrictPublicBuckets,
			"DataSourceBasic":       testAccAccountPublicAccessBlockDataSource_basic,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 5*time.Second)
}

func testAccAccountPublicAccessBlock_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3control.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountPublicAccessBlockDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, "account_id"),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", "false"),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", "false"),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", "false"),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", "false"),
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

func testAccAccountPublicAccessBlock_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3control.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountPublicAccessBlockDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfs3control.ResourceAccountPublicAccessBlock(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAccountPublicAccessBlock_AccountID(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3control.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountPublicAccessBlockDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, "account_id"),
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

func testAccAccountPublicAccessBlock_BlockPublicACLs(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3control.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountPublicAccessBlockDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockConfig_acls(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountPublicAccessBlockConfig_acls(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", "false"),
				),
			},
			{
				Config: testAccAccountPublicAccessBlockConfig_acls(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", "true"),
				),
			},
		},
	})
}

func testAccAccountPublicAccessBlock_BlockPublicPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3control.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountPublicAccessBlockDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockConfig_policy(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountPublicAccessBlockConfig_policy(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", "false"),
				),
			},
			{
				Config: testAccAccountPublicAccessBlockConfig_policy(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", "true"),
				),
			},
		},
	})
}

func testAccAccountPublicAccessBlock_IgnorePublicACLs(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3control.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountPublicAccessBlockDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockConfig_ignoreACLs(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountPublicAccessBlockConfig_ignoreACLs(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", "false"),
				),
			},
			{
				Config: testAccAccountPublicAccessBlockConfig_ignoreACLs(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", "true"),
				),
			},
		},
	})
}

func testAccAccountPublicAccessBlock_RestrictPublicBuckets(t *testing.T) {
	ctx := acctest.Context(t)
	var v s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3control.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountPublicAccessBlockDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockConfig_restrictBuckets(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountPublicAccessBlockConfig_restrictBuckets(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", "false"),
				),
			},
			{
				Config: testAccAccountPublicAccessBlockConfig_restrictBuckets(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", "true"),
				),
			},
		},
	})
}

func testAccCheckAccountPublicAccessBlockExists(ctx context.Context, n string, v *s3control.PublicAccessBlockConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Account Public Access Block ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn()

		output, err := tfs3control.FindPublicAccessBlockByAccountID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAccountPublicAccessBlockDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3_account_public_access_block" {
				continue
			}

			_, err := tfs3control.FindPublicAccessBlockByAccountID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Account Public Access Block %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAccountPublicAccessBlockConfig_basic() string {
	return `resource "aws_s3_account_public_access_block" "test" {}`
}

func testAccAccountPublicAccessBlockConfig_id() string {
	return `
data "aws_caller_identity" "test" {}

resource "aws_s3_account_public_access_block" "test" {
  account_id = data.aws_caller_identity.test.account_id
}
`
}

func testAccAccountPublicAccessBlockConfig_acls(blockPublicAcls bool) string {
	return fmt.Sprintf(`
resource "aws_s3_account_public_access_block" "test" {
  block_public_acls = %[1]t
}
`, blockPublicAcls)
}

func testAccAccountPublicAccessBlockConfig_policy(blockPublicPolicy bool) string {
	return fmt.Sprintf(`
resource "aws_s3_account_public_access_block" "test" {
  block_public_policy = %[1]t
}
`, blockPublicPolicy)
}

func testAccAccountPublicAccessBlockConfig_ignoreACLs(ignorePublicAcls bool) string {
	return fmt.Sprintf(`
resource "aws_s3_account_public_access_block" "test" {
  ignore_public_acls = %[1]t
}
`, ignorePublicAcls)
}

func testAccAccountPublicAccessBlockConfig_restrictBuckets(restrictPublicBuckets bool) string {
	return fmt.Sprintf(`
resource "aws_s3_account_public_access_block" "test" {
  restrict_public_buckets = %[1]t
}
`, restrictPublicBuckets)
}
