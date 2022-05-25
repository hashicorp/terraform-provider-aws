package kms_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kms"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccKMSKeyDataSource_basic(t *testing.T) {
	resourceName := "aws_kms_key.test"
	dataSourceName := "data.aws_kms_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "aws_account_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "creation_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "customer_master_key_spec", resourceName, "customer_master_key_spec"),
					resource.TestCheckNoResourceAttr(dataSourceName, "deletion_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enabled", resourceName, "is_enabled"),
					resource.TestCheckResourceAttr(dataSourceName, "expiration_model", ""),
					resource.TestCheckResourceAttr(dataSourceName, "key_manager", "CUSTOMER"),
					resource.TestCheckResourceAttr(dataSourceName, "key_state", "Enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "key_usage", resourceName, "key_usage"),
					resource.TestCheckResourceAttrPair(dataSourceName, "multi_region", resourceName, "multi_region"),
					resource.TestCheckResourceAttr(dataSourceName, "multi_region_configuration.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "origin", "AWS_KMS"),
					resource.TestCheckNoResourceAttr(dataSourceName, "valid_to"),
				),
			},
		},
	})
}

func TestAccKMSKeyDataSource_grantToken(t *testing.T) {
	resourceName := "aws_kms_key.test"
	dataSourceName := "data.aws_kms_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyDataSourceConfig_grantToken(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "aws_account_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "creation_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "customer_master_key_spec", resourceName, "customer_master_key_spec"),
					resource.TestCheckNoResourceAttr(dataSourceName, "deletion_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enabled", resourceName, "is_enabled"),
					resource.TestCheckResourceAttr(dataSourceName, "expiration_model", ""),
					resource.TestCheckResourceAttr(dataSourceName, "key_manager", "CUSTOMER"),
					resource.TestCheckResourceAttr(dataSourceName, "key_state", "Enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "key_usage", resourceName, "key_usage"),
					resource.TestCheckResourceAttrPair(dataSourceName, "multi_region", resourceName, "multi_region"),
					resource.TestCheckResourceAttr(dataSourceName, "multi_region_configuration.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "origin", "AWS_KMS"),
					resource.TestCheckNoResourceAttr(dataSourceName, "valid_to"),
				),
			},
		},
	})
}

func TestAccKMSKeyDataSource_multiRegionConfiguration(t *testing.T) {
	resourceName := "aws_kms_key.test"
	dataSourceName := "data.aws_kms_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyDataSourceConfig_multiRegion(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "aws_account_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "creation_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "customer_master_key_spec", resourceName, "customer_master_key_spec"),
					resource.TestCheckNoResourceAttr(dataSourceName, "deletion_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enabled", resourceName, "is_enabled"),
					resource.TestCheckResourceAttr(dataSourceName, "expiration_model", ""),
					resource.TestCheckResourceAttr(dataSourceName, "key_manager", "CUSTOMER"),
					resource.TestCheckResourceAttr(dataSourceName, "key_state", "Enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "key_usage", resourceName, "key_usage"),
					resource.TestCheckResourceAttrPair(dataSourceName, "multi_region", resourceName, "multi_region"),
					resource.TestCheckResourceAttr(dataSourceName, "multi_region_configuration.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "multi_region_configuration.0.multi_region_key_type", "PRIMARY"),
					resource.TestCheckResourceAttr(dataSourceName, "multi_region_configuration.0.primary_key.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "multi_region_configuration.0.primary_key.0.arn", resourceName, "arn"),
					resource.TestCheckResourceAttr(dataSourceName, "multi_region_configuration.0.primary_key.0.region", acctest.Region()),
					resource.TestCheckResourceAttr(dataSourceName, "multi_region_configuration.0.replica_keys.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "origin", "AWS_KMS"),
					resource.TestCheckNoResourceAttr(dataSourceName, "valid_to"),
				),
			},
		},
	})
}

func testAccKeyDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

data "aws_kms_key" "test" {
  key_id = aws_kms_key.test.key_id
}
`, rName)
}

func testAccKeyDataSourceConfig_grantToken(rName string) string {
	return acctest.ConfigCompose(testAccGrantBaseConfig(rName), fmt.Sprintf(`
resource "aws_kms_grant" "test" {
  name              = %[1]q
  key_id            = aws_kms_key.test.key_id
  grantee_principal = aws_iam_role.test.arn
  operations        = ["DescribeKey"]
}

data "aws_kms_key" "test" {
  key_id       = aws_kms_key.test.key_id
  grant_tokens = [aws_kms_grant.test.grant_token]
}
`, rName))
}

func testAccKeyDataSourceConfig_multiRegion(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  multi_region            = true
}

data "aws_kms_key" "test" {
  key_id = aws_kms_key.test.key_id
}
`, rName)
}
