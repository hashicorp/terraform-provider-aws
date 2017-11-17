package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsCognitoUserPoolDomain_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCognitoUserPoolDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCognitoUserPoolDomainConfig(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCognitoUserPoolDomainExists("aws_cognito_user_pool_domain.test"),
				),
			},
		},
	})
}

func testAccCheckAwsCognitoUserPoolDomainDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_user_pool_domain" {
			continue
		}
		domainStateConf := &resource.StateChangeConf{
			Pending:    []string{"ACTIVE", "DELETING"},
			Target:     []string{},
			Refresh:    cognitoUserPoolDomainStateRefreshFunc(rs.Primary.ID, conn),
			Timeout:    10 * time.Minute,
			Delay:      3 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err := domainStateConf.WaitForState()

		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case cognitoidentityprovider.ErrCodeResourceNotFoundException:
					return nil
				default:
					return err
				}
			}
			return err
		}
	}

	return nil

}

func testAccCheckAwsCognitoUserPoolDomainExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccAwsCognitoUserPoolDomainConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = "tf-user-pool-%s"
}

resource "aws_cognito_user_pool_domain" "test" {
  user_pool_id = "${aws_cognito_user_pool.test.id}"
  domain = "tf-user-pool-domain-%s"
}
`, rName, rName)
}
