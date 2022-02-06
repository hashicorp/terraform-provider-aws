package iam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccIAMOpenidConnectProviderClientID_basic(t *testing.T) {
	rString := sdkacctest.RandString(5)
	resourceName := "aws_iam_openid_connect_provider_client_id.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMOpenIDConnectProviderClientIDDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMOpenIDConnectClientIDProviderConfig(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMOpenIDConnectProviderClientID(resourceName),
					resource.TestCheckResourceAttr(resourceName, "arn", "arn:aws:iam::0123456789012:oidc-provider/oidc-provider-name.com"),
					resource.TestCheckResourceAttr(resourceName, "client_id",
						"266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com"),
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

func testAccCheckIAMOpenIDConnectProviderClientID(id string) resource.TestCheckFunc {
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

func testAccCheckIAMOpenIDConnectProviderClientIDDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_openid_connect_provider_client_id" {
			continue
		}

		input := &iam.GetOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: aws.String(rs.Primary.ID),
		}
		out, _ := conn.GetOpenIDConnectProvider(input)

		if out.ClientIDList[0] != nil {
			return fmt.Errorf("Found ClientID entry in IAM OpenID Connect Provider, expected none: %s", *out.ClientIDList[0])
		}
	}

	return nil
}

func testAccIAMOpenIDConnectClientIDProviderConfig(rString string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider_client_id" "test" {
  arn 		= "arn:aws:iam::0123456789012:oidc-provider/oidc-provider-name.com"
  client_id = "0123456789012-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com"
}
`, rString)
}
