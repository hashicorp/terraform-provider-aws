package iam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIAMOpenidConnectProviderDataSource_basic(t *testing.T) {
	rString := sdkacctest.RandString(5)
	dataSourceName := "data.aws_iam_openid_connect_provider.test"
	resourceName := "aws_iam_openid_connect_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenIDConnectProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenIDConnectProviderDataSourceConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenIDConnectProvider(resourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "url", resourceName, "url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "client_id_list", resourceName, "client_id_list"),
					resource.TestCheckResourceAttrPair(dataSourceName, "thumbprint_list", resourceName, "thumbprint_list"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccIAMOpenidConnectProviderDataSource_url(t *testing.T) {
	rString := sdkacctest.RandString(5)
	dataSourceName := "data.aws_iam_openid_connect_provider.test"
	resourceName := "aws_iam_openid_connect_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenIDConnectProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenIDConnectProviderDataSourceConfig_url(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenIDConnectProvider(resourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "url", resourceName, "url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "client_id_list", resourceName, "client_id_list"),
					resource.TestCheckResourceAttrPair(dataSourceName, "thumbprint_list", resourceName, "thumbprint_list"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccIAMOpenidConnectProviderDataSource_tags(t *testing.T) {
	rString := sdkacctest.RandString(5)
	dataSourceName := "data.aws_iam_openid_connect_provider.test"
	resourceName := "aws_iam_openid_connect_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenIDConnectProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenIDConnectProviderDataSourceConfig_tags(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenIDConnectProvider(resourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "url", resourceName, "url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "client_id_list", resourceName, "client_id_list"),
					resource.TestCheckResourceAttrPair(dataSourceName, "thumbprint_list", resourceName, "thumbprint_list"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.tag1", "test-value1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.tag2", "test-value2")),
			},
		},
	})
}

func testAccOpenIDConnectProviderDataSourceConfig_basic(rString string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%s"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = []
}

data "aws_iam_openid_connect_provider" "test" {
  arn = aws_iam_openid_connect_provider.test.arn
}
`, rString)
}

func testAccOpenIDConnectProviderDataSourceConfig_url(rString string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%s"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = []
}

data "aws_iam_openid_connect_provider" "test" {
  url = "https://${aws_iam_openid_connect_provider.test.url}"
}
`, rString)
}

func testAccOpenIDConnectProviderDataSourceConfig_tags(rString string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%s"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = []

  tags = {
    tag1 = "test-value1"
    tag2 = "test-value2"
  }
}

data "aws_iam_openid_connect_provider" "test" {
  arn = aws_iam_openid_connect_provider.test.arn
}
`, rString)
}
