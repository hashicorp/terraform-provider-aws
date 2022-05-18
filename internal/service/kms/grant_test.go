package kms_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkms "github.com/hashicorp/terraform-provider-aws/internal/service/kms"
)

func TestAccKMSGrant_basic(t *testing.T) {
	resourceName := "aws_kms_grant.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGrant_Basic(rName, "\"Encrypt\", \"Decrypt\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "operations.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "operations.*", "Encrypt"),
					resource.TestCheckTypeSetElemAttr(resourceName, "operations.*", "Decrypt"),
					resource.TestCheckResourceAttrSet(resourceName, "grantee_principal"),
					resource.TestCheckResourceAttrSet(resourceName, "key_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"grant_token", "retire_on_delete"},
			},
		},
	})
}

func TestAccKMSGrant_withConstraints(t *testing.T) {
	resourceName := "aws_kms_grant.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGrant_withConstraints(rName, "encryption_context_equals", `foo = "bar"
                        baz = "kaz"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "constraints.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "constraints.*", map[string]string{
						"encryption_context_equals.%":   "2",
						"encryption_context_equals.baz": "kaz",
						"encryption_context_equals.foo": "bar",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"grant_token", "retire_on_delete"},
			},
			{
				Config: testAccGrant_withConstraints(rName, "encryption_context_subset", `foo = "bar"
			            baz = "kaz"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "constraints.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "constraints.*", map[string]string{
						"encryption_context_subset.%":   "2",
						"encryption_context_subset.baz": "kaz",
						"encryption_context_subset.foo": "bar",
					}),
				),
			},
		},
	})
}

func TestAccKMSGrant_withRetiringPrincipal(t *testing.T) {
	resourceName := "aws_kms_grant.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGrant_withRetiringPrincipal(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "retiring_principal"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"grant_token", "retire_on_delete"},
			},
		},
	})
}

func TestAccKMSGrant_bare(t *testing.T) {
	resourceName := "aws_kms_grant.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGrant_bare(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantExists(resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "name"),
					resource.TestCheckNoResourceAttr(resourceName, "constraints.#"),
					resource.TestCheckNoResourceAttr(resourceName, "retiring_principal"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"grant_token", "retire_on_delete"},
			},
		},
	})
}

func TestAccKMSGrant_arn(t *testing.T) {
	resourceName := "aws_kms_grant.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGrant_ARN(rName, "\"Encrypt\", \"Decrypt\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "operations.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "operations.*", "Encrypt"),
					resource.TestCheckTypeSetElemAttr(resourceName, "operations.*", "Decrypt"),
					resource.TestCheckResourceAttrSet(resourceName, "grantee_principal"),
					resource.TestCheckResourceAttrSet(resourceName, "key_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"grant_token", "retire_on_delete"},
			},
		},
	})
}

func TestAccKMSGrant_asymmetricKey(t *testing.T) {
	resourceName := "aws_kms_grant.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGrant_AsymmetricKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"grant_token", "retire_on_delete"},
			},
		},
	})
}

func TestAccKMSGrant_disappears(t *testing.T) {
	resourceName := "aws_kms_grant.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGrant_Basic(rName, "\"Encrypt\", \"Decrypt\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGrantExists(resourceName),
					testAccCheckGrantDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckGrantDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kms_grant" {
			continue
		}

		err := tfkms.WaitForGrantToBeRevoked(conn, rs.Primary.Attributes["key_id"], rs.Primary.ID)
		return err
	}

	return nil
}

func testAccCheckGrantExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		return nil
	}
}

func testAccCheckGrantDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn

		revokeInput := kms.RevokeGrantInput{
			GrantId: aws.String(rs.Primary.Attributes["grant_id"]),
			KeyId:   aws.String(rs.Primary.Attributes["key_id"]),
		}

		_, err := conn.RevokeGrant(&revokeInput)
		return err
	}
}

func testAccGrantBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

data "aws_iam_policy_document" "test" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/service-role/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}
`, rName)
}

func testAccGrant_Basic(rName string, operations string) string {
	return acctest.ConfigCompose(testAccGrantBaseConfig(rName), fmt.Sprintf(`
resource "aws_kms_grant" "test" {
  name              = %[1]q
  key_id            = aws_kms_key.test.key_id
  grantee_principal = aws_iam_role.test.arn
  operations        = [%[2]s]
}
`, rName, operations))
}

func testAccGrant_withConstraints(rName string, constraintName string, encryptionContext string) string {
	return acctest.ConfigCompose(testAccGrantBaseConfig(rName), fmt.Sprintf(`
resource "aws_kms_grant" "test" {
  name              = %[1]q
  key_id            = aws_kms_key.test.key_id
  grantee_principal = aws_iam_role.test.arn
  operations        = ["RetireGrant", "DescribeKey"]

  constraints {
    %[2]s = {
      %[3]s
    }
  }
}
`, rName, constraintName, encryptionContext))
}

func testAccGrant_withRetiringPrincipal(rName string) string {
	return acctest.ConfigCompose(testAccGrantBaseConfig(rName), fmt.Sprintf(`
resource "aws_kms_grant" "test" {
  name               = %[1]q
  key_id             = aws_kms_key.test.key_id
  grantee_principal  = aws_iam_role.test.arn
  operations         = ["ReEncryptTo", "CreateGrant"]
  retiring_principal = aws_iam_role.test.arn
}
`, rName))
}

func testAccGrant_bare(rName string) string {
	return acctest.ConfigCompose(testAccGrantBaseConfig(rName), `
resource "aws_kms_grant" "test" {
  key_id            = aws_kms_key.test.key_id
  grantee_principal = aws_iam_role.test.arn
  operations        = ["ReEncryptTo", "CreateGrant"]
}
`)
}

func testAccGrant_ARN(rName string, operations string) string {
	return acctest.ConfigCompose(testAccGrantBaseConfig(rName), fmt.Sprintf(`
resource "aws_kms_grant" "test" {
  name              = %[1]q
  key_id            = aws_kms_key.test.arn
  grantee_principal = aws_iam_role.test.arn
  operations        = [%[2]s]
}
`, rName, operations))
}

func testAccGrant_AsymmetricKey(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_grant" "test" {
  name              = %[1]q
  key_id            = aws_kms_key.test.key_id
  grantee_principal = aws_iam_role.test.arn
  operations        = ["GetPublicKey", "Sign", "Verify"]
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  key_usage                = "SIGN_VERIFY"
  customer_master_key_spec = "RSA_2048"
}

data "aws_iam_policy_document" "test" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/service-role/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}
`, rName)
}
