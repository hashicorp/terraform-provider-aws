package s3control_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// S3 account-level settings must run serialized
// for TeamCity environment
func TestAccS3ControlAccountPublicAccessBlock_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"PublicAccessBlock": {
			"basic":                 testAccAccountPublicAccessBlock_basic,
			"disappears":            testAccAccountPublicAccessBlock_disappears,
			"AccountId":             testAccAccountPublicAccessBlock_AccountID,
			"BlockPublicAcls":       testAccAccountPublicAccessBlock_BlockPublicACLs,
			"BlockPublicPolicy":     testAccAccountPublicAccessBlock_BlockPublicPolicy,
			"IgnorePublicAcls":      testAccAccountPublicAccessBlock_IgnorePublicACLs,
			"RestrictPublicBuckets": testAccAccountPublicAccessBlock_RestrictPublicBuckets,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
					// Explicitly sleep between tests for eventual consistency
					time.Sleep(5 * time.Second)
				})
			}
		})
	}
}

func testAccAccountPublicAccessBlock_basic(t *testing.T) {
	var configuration1 s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountPublicAccessBlockDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(resourceName, &configuration1),
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
	var configuration1 s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountPublicAccessBlockDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(resourceName, &configuration1),
					testAccCheckAccountPublicAccessBlockDisappears(),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAccountPublicAccessBlock_AccountID(t *testing.T) {
	var configuration1 s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountPublicAccessBlockDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(resourceName, &configuration1),
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
	var configuration1, configuration2, configuration3 s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountPublicAccessBlockDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockConfig_acls(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(resourceName, &configuration1),
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
					testAccCheckAccountPublicAccessBlockExists(resourceName, &configuration2),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", "false"),
				),
			},
			{
				Config: testAccAccountPublicAccessBlockConfig_acls(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(resourceName, &configuration3),
					resource.TestCheckResourceAttr(resourceName, "block_public_acls", "true"),
				),
			},
		},
	})
}

func testAccAccountPublicAccessBlock_BlockPublicPolicy(t *testing.T) {
	var configuration1, configuration2, configuration3 s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountPublicAccessBlockDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockConfig_policy(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(resourceName, &configuration1),
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
					testAccCheckAccountPublicAccessBlockExists(resourceName, &configuration2),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", "false"),
				),
			},
			{
				Config: testAccAccountPublicAccessBlockConfig_policy(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(resourceName, &configuration3),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", "true"),
				),
			},
		},
	})
}

func testAccAccountPublicAccessBlock_IgnorePublicACLs(t *testing.T) {
	var configuration1, configuration2, configuration3 s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountPublicAccessBlockDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockConfig_ignoreACLs(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(resourceName, &configuration1),
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
					testAccCheckAccountPublicAccessBlockExists(resourceName, &configuration2),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", "false"),
				),
			},
			{
				Config: testAccAccountPublicAccessBlockConfig_ignoreACLs(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(resourceName, &configuration3),
					resource.TestCheckResourceAttr(resourceName, "ignore_public_acls", "true"),
				),
			},
		},
	})
}

func testAccAccountPublicAccessBlock_RestrictPublicBuckets(t *testing.T) {
	var configuration1, configuration2, configuration3 s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountPublicAccessBlockDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPublicAccessBlockConfig_restrictBuckets(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(resourceName, &configuration1),
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
					testAccCheckAccountPublicAccessBlockExists(resourceName, &configuration2),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", "false"),
				),
			},
			{
				Config: testAccAccountPublicAccessBlockConfig_restrictBuckets(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExists(resourceName, &configuration3),
					resource.TestCheckResourceAttr(resourceName, "restrict_public_buckets", "true"),
				),
			},
		},
	})
}

func testAccCheckAccountPublicAccessBlockExists(resourceName string, configuration *s3control.PublicAccessBlockConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Account Public Access Block ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn

		input := &s3control.GetPublicAccessBlockInput{
			AccountId: aws.String(rs.Primary.ID),
		}

		// Retry for eventual consistency
		var output *s3control.GetPublicAccessBlockOutput
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			var err error
			output, err = conn.GetPublicAccessBlock(input)

			if tfawserr.ErrCodeEquals(err, s3control.ErrCodeNoSuchPublicAccessBlockConfiguration) {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if err != nil {
			return err
		}

		if output == nil || output.PublicAccessBlockConfiguration == nil {
			return fmt.Errorf("S3 Account Public Access Block not found")
		}

		*configuration = *output.PublicAccessBlockConfiguration

		return nil
	}
}

func testAccCheckAccountPublicAccessBlockDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_account_public_access_block" {
			continue
		}

		input := &s3control.GetPublicAccessBlockInput{
			AccountId: aws.String(rs.Primary.ID),
		}

		// Retry for eventual consistency
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			_, err := conn.GetPublicAccessBlock(input)

			if tfawserr.ErrCodeEquals(err, s3control.ErrCodeNoSuchPublicAccessBlockConfiguration) {
				return nil
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return resource.RetryableError(fmt.Errorf("S3 Account Public Access Block (%s) still exists", rs.Primary.ID))
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckAccountPublicAccessBlockDisappears() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn
		accountID := acctest.Provider.Meta().(*conns.AWSClient).AccountID

		deleteInput := &s3control.DeletePublicAccessBlockInput{
			AccountId: aws.String(accountID),
		}

		_, err := conn.DeletePublicAccessBlock(deleteInput)

		if err != nil {
			return err
		}

		getInput := &s3control.GetPublicAccessBlockInput{
			AccountId: aws.String(accountID),
		}

		// Retry for eventual consistency
		return resource.Retry(1*time.Minute, func() *resource.RetryError {
			_, err := conn.GetPublicAccessBlock(getInput)

			if tfawserr.ErrCodeEquals(err, s3control.ErrCodeNoSuchPublicAccessBlockConfiguration) {
				return nil
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return resource.RetryableError(fmt.Errorf("S3 Account Public Access Block (%s) still exists", accountID))
		})
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
  block_public_acls = %t
}
`, blockPublicAcls)
}

func testAccAccountPublicAccessBlockConfig_policy(blockPublicPolicy bool) string {
	return fmt.Sprintf(`
resource "aws_s3_account_public_access_block" "test" {
  block_public_policy = %t
}
`, blockPublicPolicy)
}

func testAccAccountPublicAccessBlockConfig_ignoreACLs(ignorePublicAcls bool) string {
	return fmt.Sprintf(`
resource "aws_s3_account_public_access_block" "test" {
  ignore_public_acls = %t
}
`, ignorePublicAcls)
}

func testAccAccountPublicAccessBlockConfig_restrictBuckets(restrictPublicBuckets bool) string {
	return fmt.Sprintf(`
resource "aws_s3_account_public_access_block" "test" {
  restrict_public_buckets = %t
}
`, restrictPublicBuckets)
}
