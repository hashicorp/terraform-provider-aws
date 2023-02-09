package kms_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kms"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccKMSKeyDataSource_byKeyARN(t *testing.T) {
	resourceName := "aws_kms_key.test"
	dataSourceName := "data.aws_kms_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyDataSourceConfig_byKeyARN(rName),
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

func TestAccKMSKeyDataSource_byKeyID(t *testing.T) {
	resourceName := "aws_kms_key.test"
	dataSourceName := "data.aws_kms_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyDataSourceConfig_byKeyID(rName),
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

func TestAccKMSKeyDataSource_byAliasARN(t *testing.T) {
	resourceName := "aws_kms_key.test"
	dataSourceName := "data.aws_kms_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyDataSourceConfig_byAliasARN(rName),
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

func TestAccKMSKeyDataSource_byAliasID(t *testing.T) {
	resourceName := "aws_kms_key.test"
	dataSourceName := "data.aws_kms_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyDataSourceConfig_byAliasID(rName),
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
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

func TestAccKMSKeyDataSource_multiRegionConfigurationByARN(t *testing.T) {
	resourceName := "aws_kms_key.test"
	dataSourceName := "data.aws_kms_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyDataSourceConfig_multiRegionByARN(rName),
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

func TestAccKMSKeyDataSource_multiRegionConfigurationByID(t *testing.T) {
	resourceName := "aws_kms_key.test"
	dataSourceName := "data.aws_kms_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, kms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyDataSourceConfig_multiRegionByID(rName),
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

func testAccKeyDataSourceConfig_byKeyARN(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

data "aws_kms_key" "test" {
  key_id = aws_kms_key.test.arn
}
`, rName)
}

func testAccKeyDataSourceConfig_byKeyID(rName string) string {
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

func testAccKeyDataSourceConfig_byAliasARN(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "test" {
  name          = "alias/%[1]s"
  target_key_id = aws_kms_key.test.id
}

data "aws_kms_key" "test" {
  key_id = aws_kms_alias.test.arn
}
`, rName)
}

func testAccKeyDataSourceConfig_byAliasID(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "test" {
  name          = "alias/%[1]s"
  target_key_id = aws_kms_key.test.id
}

data "aws_kms_key" "test" {
  key_id = aws_kms_alias.test.id
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

func testAccKeyDataSourceConfig_multiRegionByARN(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  multi_region            = true
}

data "aws_kms_key" "test" {
  key_id = aws_kms_key.test.arn
}
`, rName)
}

func testAccKeyDataSourceConfig_multiRegionByID(rName string) string {
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
