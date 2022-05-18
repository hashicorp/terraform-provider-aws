package transfer_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
)

func testAccSSHKey_basic(t *testing.T) {
	var conf transfer.SshPublicKey
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_transfer_ssh_key.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, transfer.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSSHKeyConfig_basic(rName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSSHKeyExists(resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "server_id", "aws_transfer_server.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "user_name", "aws_transfer_user.test", "user_name"),
					resource.TestCheckResourceAttr(resourceName, "body", publicKey),
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

func testAccCheckSSHKeyExists(n string, res *transfer.SshPublicKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transfer Ssh Public Key ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn
		serverID, userName, sshKeyID, err := tftransfer.DecodeSSHKeyID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing Transfer SSH Public Key ID: %s", err)
		}

		describe, err := conn.DescribeUser(&transfer.DescribeUserInput{
			ServerId: aws.String(serverID),
			UserName: aws.String(userName),
		})

		if err != nil {
			return err
		}

		for _, sshPublicKey := range describe.User.SshPublicKeys {
			if sshKeyID == *sshPublicKey.SshPublicKeyId {
				*res = *sshPublicKey
				return nil
			}
		}

		return fmt.Errorf("Transfer Ssh Public Key doesn't exists.")
	}
}

func testAccCheckSSHKeyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_transfer_ssh_key" {
			continue
		}
		serverID, userName, sshKeyID, err := tftransfer.DecodeSSHKeyID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing Transfer SSH Public Key ID: %w", err)
		}

		describe, err := conn.DescribeUser(&transfer.DescribeUserInput{
			UserName: aws.String(userName),
			ServerId: aws.String(serverID),
		})

		if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		for _, sshPublicKey := range describe.User.SshPublicKeys {
			if sshKeyID == *sshPublicKey.SshPublicKeyId {
				return fmt.Errorf("Transfer SSH Public Key still exists")
			}
		}
	}

	return nil
}

func testAccSSHKeyConfig_basic(rName, publicKey string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  identity_provider_type = "SERVICE_MANAGED"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "transfer.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowFullAccesstoS3",
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_transfer_user" "test" {
  server_id = aws_transfer_server.test.id
  user_name = "tftestuser"
  role      = aws_iam_role.test.arn
}

resource "aws_transfer_ssh_key" "test" {
  server_id = aws_transfer_server.test.id
  user_name = aws_transfer_user.test.user_name
  body      = "%[2]s"
}
`, rName, publicKey)
}
