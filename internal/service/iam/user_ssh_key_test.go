package iam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccIAMUserSSHKey_basic(t *testing.T) {
	var conf iam.GetSSHPublicKeyOutput
	resourceName := "aws_iam_user_ssh_key.user"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	publicKey, _, err := RandSSHKeyPairSize(2048, acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSSHKeyConfig_sshEncoding(rName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserSSHKeyExists(resourceName, "Inactive", &conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccUserSSHKeyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIAMUserSSHKey_pemEncoding(t *testing.T) {
	var conf iam.GetSSHPublicKeyOutput

	ri := sdkacctest.RandInt()
	config := fmt.Sprintf(testAccSSHKeyConfig_pemEncoding, ri)
	resourceName := "aws_iam_user_ssh_key.user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserSSHKeyExists(resourceName, "Active", &conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccUserSSHKeyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckUserSSHKeyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_user_ssh_key" {
			continue
		}

		username := rs.Primary.Attributes["username"]
		encoding := rs.Primary.Attributes["encoding"]
		_, err := conn.GetSSHPublicKey(&iam.GetSSHPublicKeyInput{
			SSHPublicKeyId: aws.String(rs.Primary.ID),
			UserName:       aws.String(username),
			Encoding:       aws.String(encoding),
		})
		if err == nil {
			return fmt.Errorf("still exist.")
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "NoSuchEntity" {
			return err
		}
	}

	return nil
}

func testAccCheckUserSSHKeyExists(n, status string, res *iam.GetSSHPublicKeyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSHPublicKeyID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		username := rs.Primary.Attributes["username"]
		encoding := rs.Primary.Attributes["encoding"]
		resp, err := conn.GetSSHPublicKey(&iam.GetSSHPublicKeyInput{
			SSHPublicKeyId: aws.String(rs.Primary.ID),
			UserName:       aws.String(username),
			Encoding:       aws.String(encoding),
		})
		if err != nil {
			return err
		}

		*res = *resp

		keyStruct := resp.SSHPublicKey

		if *keyStruct.Status != status {
			return fmt.Errorf("Key status has wrong status should be %s is %s", status, *keyStruct.Status)
		}

		return nil
	}
}

func testAccUserSSHKeyImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		username := rs.Primary.Attributes["username"]
		sshPublicKeyId := rs.Primary.Attributes["ssh_public_key_id"]
		encoding := rs.Primary.Attributes["encoding"]

		return fmt.Sprintf("%s:%s:%s", username, sshPublicKeyId, encoding), nil
	}
}

func testAccSSHKeyConfig_sshEncoding(rName, publicKey string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user" {
  name = %[1]q
  path = "/"
}

resource "aws_iam_user_ssh_key" "user" {
  username   = aws_iam_user.user.name
  encoding   = "SSH"
  public_key = %[2]q
  status     = "Inactive"
}
`, rName, publicKey)
}

const testAccSSHKeyConfig_pemEncoding = `
resource "aws_iam_user" "user" {
  name = "test-user-%d"
  path = "/"
}

resource "aws_iam_user_ssh_key" "user" {
  username = aws_iam_user.user.name
  encoding = "PEM"

  public_key = <<EOF
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA9xercjxBRM1dC191/AbF
3TLEM9cdnBIpCgxGNGiI+NaoMTAj/4rXp3ql0iBWQaeb4sz72qCEd1JvcSuzxqFv
IIrqRp/hD7sSAOHAzOL8zqjpIjD4c+VytMIRI5Fc06OPktKbrw2bsCLHYlvZsYSX
O7YATS9HGJVkmFZM+Bv37JTX0T1uZmADOPX+H4bcT2+aJOENi4PXTylRzvwYHruc
KDHO0WNKdXo+g+AihROpcpkgyaVtGB1/8KhPfnHxGroe8WXBtKvbdrWuhen5l9Go
L6RcmaPGhW13lAa+6LEgiTYL2r1mzP9Op4lqzr2F9scFnYV5l0q21/GW2m1aIQSu
NQIDAQAB
-----END PUBLIC KEY-----
EOF
}
`
