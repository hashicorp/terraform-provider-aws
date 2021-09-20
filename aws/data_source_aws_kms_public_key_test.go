package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kms"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccDataSourceAwsKmsPublicKey_basic(t *testing.T) {
	resourceName := "aws_kms_key.test"
	datasourceName := "data.aws_kms_public_key.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, kms.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsKmsPublicKeyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsKmsPublicKeyCheck(datasourceName),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "customer_master_key_spec", resourceName, "customer_master_key_spec"),
					resource.TestCheckResourceAttrPair(datasourceName, "key_id", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "key_usage", resourceName, "key_usage"),
					resource.TestCheckResourceAttrSet(datasourceName, "public_key"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsKmsPublicKey_encrypt(t *testing.T) {
	resourceName := "aws_kms_key.test"
	datasourceName := "data.aws_kms_public_key.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, kms.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsKmsPublicKeyEncryptConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsKmsPublicKeyCheck(datasourceName),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "key_id", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "customer_master_key_spec", resourceName, "customer_master_key_spec"),
					resource.TestCheckResourceAttrPair(datasourceName, "key_usage", resourceName, "key_usage"),
					resource.TestCheckResourceAttrSet(datasourceName, "public_key"),
				),
			},
		},
	})
}

func testAccDataSourceAwsKmsPublicKeyCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		return nil
	}
}

func testAccDataSourceAwsKmsPublicKeyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description              = %[1]q
  deletion_window_in_days  = 7
  customer_master_key_spec = "RSA_2048"
  key_usage                = "SIGN_VERIFY"
}

data "aws_kms_public_key" "test" {
  key_id = aws_kms_key.test.arn
}
`, rName)
}

func testAccDataSourceAwsKmsPublicKeyEncryptConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description              = %[1]q
  deletion_window_in_days  = 7
  customer_master_key_spec = "RSA_2048"
  key_usage                = "ENCRYPT_DECRYPT"
}

data "aws_kms_public_key" "test" {
  key_id = aws_kms_key.test.arn
}
`, rName)
}
