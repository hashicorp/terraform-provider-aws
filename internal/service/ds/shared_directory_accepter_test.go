package ds_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfds "github.com/hashicorp/terraform-provider-aws/internal/service/ds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDSSharedDirectoryAccepter_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_shared_directory_accepter.test"

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, directoryservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(t),
		CheckDestroy:             testAccCheckSharedDirectoryAccepterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSharedDirectoryAccepterConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSharedDirectoryAccepterExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "method", directoryservice.ShareMethodHandshake),
					resource.TestCheckResourceAttr(resourceName, "notes", "There were hints and allegations"),
					resource.TestCheckResourceAttrPair(resourceName, "owner_account_id", "data.aws_caller_identity.current", "account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "owner_directory_id"),
					resource.TestCheckResourceAttrSet(resourceName, "shared_directory_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"notes",
				},
			},
		},
	})

}

func testAccCheckSharedDirectoryAccepterExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DS, create.ErrActionCheckingExistence, tfds.ResNameSharedDirectoryAccepter, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DS, create.ErrActionCheckingExistence, tfds.ResNameSharedDirectoryAccepter, name, errors.New("no ID is set"))
		}

		ownerId := rs.Primary.Attributes["owner_directory_id"]
		sharedId := rs.Primary.Attributes["shared_directory_id"]

		conn := acctest.Provider.Meta().(*conns.AWSClient).DSConn
		out, err := conn.DescribeSharedDirectories(&directoryservice.DescribeSharedDirectoriesInput{
			OwnerDirectoryId:   aws.String(ownerId),
			SharedDirectoryIds: aws.StringSlice([]string{sharedId}),
		})

		if err != nil {
			return create.Error(names.DS, create.ErrActionCheckingExistence, tfds.ResNameSharedDirectoryAccepter, name, err)
		}

		if len(out.SharedDirectories) < 1 {
			return create.Error(names.DS, create.ErrActionCheckingExistence, tfds.ResNameSharedDirectoryAccepter, name, errors.New("not found"))
		}

		if aws.StringValue(out.SharedDirectories[0].SharedDirectoryId) != sharedId {
			return create.Error(names.DS, create.ErrActionCheckingExistence, tfds.ResNameSharedDirectoryAccepter, rs.Primary.ID, fmt.Errorf("shared directory ID mismatch - existing: %q, state: %q", aws.StringValue(out.SharedDirectories[0].SharedDirectoryId), sharedId))
		}

		if aws.StringValue(out.SharedDirectories[0].OwnerDirectoryId) != ownerId {
			return create.Error(names.DS, create.ErrActionCheckingExistence, tfds.ResNameSharedDirectoryAccepter, rs.Primary.ID, fmt.Errorf("owner directory ID mismatch - existing: %q, state: %q", aws.StringValue(out.SharedDirectories[0].OwnerDirectoryId), ownerId))
		}

		return nil
	}

}

func testAccCheckSharedDirectoryAccepterDestroy(s *terraform.State) error {
	// cannot be destroyed from consumer account
	return nil
}

func testAccSharedDirectoryAccepterConfig_basic(rName, domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		testAccDirectoryConfig_microsoftStandard(rName, domain),
		`
data "aws_caller_identity" "current" {}

resource "aws_directory_service_shared_directory" "test" {
  directory_id = aws_directory_service_directory.test.id
  notes        = "There were hints and allegations"

  target {
    id = data.aws_caller_identity.consumer.account_id
  }
}

data "aws_caller_identity" "consumer" {
  provider = "awsalternate"
}

resource "aws_directory_service_shared_directory_accepter" "test" {
  provider = "awsalternate"

  shared_directory_id = aws_directory_service_shared_directory.test.shared_directory_id
}
`)
}
