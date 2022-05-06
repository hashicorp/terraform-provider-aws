package iam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
)

func TestAccIAMOpenIDConnectProvider_basic(t *testing.T) {
	rString := sdkacctest.RandString(5)
	url := "accounts.testle.com/" + rString
	resourceName := "aws_iam_openid_connect_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenIDConnectProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenIDConnectProviderConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenIDConnectProvider(resourceName),
					acctest.CheckResourceAttrGlobalARN(resourceName, "arn", "iam", fmt.Sprintf("oidc-provider/%s", url)),
					resource.TestCheckResourceAttr(resourceName, "url", url),
					resource.TestCheckResourceAttr(resourceName, "client_id_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_id_list.0",
						"266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com"),
					resource.TestCheckResourceAttr(resourceName, "thumbprint_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOpenIDConnectProviderConfig_modified(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenIDConnectProvider(resourceName),
					resource.TestCheckResourceAttr(resourceName, "url", url),
					resource.TestCheckResourceAttr(resourceName, "client_id_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "client_id_list.0",
						"266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com"),
					resource.TestCheckResourceAttr(resourceName, "thumbprint_list.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "thumbprint_list.0", "cf23df2207d99a74fbe169e3eba035e633b65d94"),
					resource.TestCheckResourceAttr(resourceName, "thumbprint_list.1", "c784713d6f9cb67b55dd84f4e4af7832d42b8f55"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccIAMOpenIDConnectProvider_tags(t *testing.T) {
	rString := sdkacctest.RandString(5)
	resourceName := "aws_iam_openid_connect_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenIDConnectProviderConfig_tags1(rString, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenIDConnectProvider(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
			{
				Config: testAccOpenIDConnectProviderConfig_tags2(rString, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenIDConnectProvider(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccOpenIDConnectProviderConfig_tags1(rString, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenIDConnectProvider(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccIAMOpenIDConnectProvider_disappears(t *testing.T) {
	rString := sdkacctest.RandString(5)
	resourceName := "aws_iam_openid_connect_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenIDConnectProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenIDConnectProviderConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenIDConnectProvider(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfiam.ResourceOpenIDConnectProvider(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckOpenIDConnectProviderDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_openid_connect_provider" {
			continue
		}

		input := &iam.GetOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: aws.String(rs.Primary.ID),
		}
		out, err := conn.GetOpenIDConnectProvider(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
				// none found, that's good
				return nil
			}
			return fmt.Errorf("Error reading IAM OpenID Connect Provider, out: %s, err: %w", out, err)
		}

		if out != nil {
			return fmt.Errorf("Found IAM OpenID Connect Provider, expected none: %s", out)
		}
	}

	return nil
}

func testAccCheckOpenIDConnectProvider(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("Not Found: %s", id)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn
		_, err := conn.GetOpenIDConnectProvider(&iam.GetOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccOpenIDConnectProviderConfig_basic(rString string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%s"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = []
}
`, rString)
}

func testAccOpenIDConnectProviderConfig_modified(rString string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%s"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = ["cf23df2207d99a74fbe169e3eba035e633b65d94", "c784713d6f9cb67b55dd84f4e4af7832d42b8f55"]
}
`, rString)
}

func testAccOpenIDConnectProviderConfig_tags1(rString, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%s"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = []

  tags = {
    %[2]q = %[3]q
  }
}
`, rString, tagKey1, tagValue1)
}

func testAccOpenIDConnectProviderConfig_tags2(rString, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/%s"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = []

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rString, tagKey1, tagValue1, tagKey2, tagValue2)
}
