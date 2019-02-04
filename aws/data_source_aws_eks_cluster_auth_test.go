package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/kubernetes-sigs/aws-iam-authenticator/pkg/token"
)

func TestAccAWSEksClusterAuthDataSource_basic(t *testing.T) {
	dataSourceResourceName := "data.aws_eks_cluster_auth.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEksClusterAuthConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEksClusterAuthToken(dataSourceResourceName),
				),
			},
		},
	})
}

func testAccCheckAwsEksClusterAuthToken(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find EKS Cluster Auth resource: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("EKS Cluster Auth resource ID not set")
		}

		name := rs.Primary.Attributes["name"]
		if expected := "foobar"; name != expected {
			return fmt.Errorf("Incorrect EKS cluster name: expected %q, got %q", expected, name)
		}

		tok := rs.Primary.Attributes["token"]
		if tok == "" {
			return fmt.Errorf("Token expected to not be nil")
		}

		verifier := token.NewVerifier(name)
		identity, err := verifier.Verify(tok)
		if err != nil {
			return fmt.Errorf("Error verifying token for cluster %q: %v", name, err)
		}
		if identity.ARN == "" {
			return fmt.Errorf("Received unexpected blank ARN for token identity")
		}

		return nil
	}
}

const testAccCheckAwsEksClusterAuthConfig_basic = `
data "aws_eks_cluster_auth" "test" {
	name = "foobar"
}
`
