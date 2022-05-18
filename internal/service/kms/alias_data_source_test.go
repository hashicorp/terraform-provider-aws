package kms_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/kms"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccKMSAliasDataSource_service(t *testing.T) {
	rName := "alias/aws/s3"
	resourceName := "data.aws_kms_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAliasDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "kms", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "target_key_arn", "kms", regexp.MustCompile(`key/[a-z0-9]{8}-([a-z0-9]{4}-){3}[a-z0-9]{12}`)),
					resource.TestMatchResourceAttr(resourceName, "target_key_id", regexp.MustCompile("^[a-z0-9]{8}-([a-z0-9]{4}-){3}[a-z0-9]{12}$")),
				),
			},
		},
	})
}

func TestAccKMSAliasDataSource_cmk(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	aliasResourceName := "aws_kms_alias.test"
	datasourceAliasResourceName := "data.aws_kms_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAliasCMKDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceAliasResourceName, "arn", aliasResourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceAliasResourceName, "target_key_arn", aliasResourceName, "target_key_arn"),
					resource.TestCheckResourceAttrPair(datasourceAliasResourceName, "target_key_id", aliasResourceName, "target_key_id"),
				),
			},
		},
	})
}

func testAccAliasDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_kms_alias" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAliasCMKDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "test" {
  name          = "alias/%[1]s"
  target_key_id = aws_kms_key.test.key_id
}

data "aws_kms_alias" "test" {
  name = aws_kms_alias.test.name
}
`, rName)
}
