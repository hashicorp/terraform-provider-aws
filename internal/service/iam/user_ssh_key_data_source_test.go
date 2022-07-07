package iam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIAMUserSSHKeyDataSource_basic(t *testing.T) {
	resourceName := "aws_iam_user_ssh_key.test"
	dataSourceName := "data.aws_iam_user_ssh_key.test"

	username := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	publicKey, _, err := RandSSHKeyPairSize(2048, acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserSSHKeyDataSourceConfig_basic(username, publicKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "encoding", resourceName, "encoding"),
					resource.TestCheckResourceAttrPair(dataSourceName, "fingerprint", resourceName, "fingerprint"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_key", resourceName, "public_key"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ssh_public_key_id", resourceName, "ssh_public_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "username", resourceName, "username"),
				),
			},
		},
	})
}

func testAccUserSSHKeyDataSourceConfig_basic(username, publicKey string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
  path = "/"
}

resource "aws_iam_user_ssh_key" "test" {
  username   = aws_iam_user.test.name
  encoding   = "SSH"
  public_key = %[2]q
  status     = "Inactive"
}

data "aws_iam_user_ssh_key" "test" {
  username          = aws_iam_user.test.name
  encoding          = "SSH"
  ssh_public_key_id = aws_iam_user_ssh_key.test.ssh_public_key_id
}
`, username, publicKey)
}
