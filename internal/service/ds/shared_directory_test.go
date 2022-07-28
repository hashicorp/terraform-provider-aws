package ds_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccDSSharedDirectory_basic(t *testing.T) {
	var sharedDirectory directoryservice.SharedDirectory
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_shared_directory.test"

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, directoryservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(t),
		CheckDestroy:             testAccCheckSharedDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSharedDirectoryConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "shared_directory_id"),
					testAccCheckSharedDirectoryExists(resourceName, &sharedDirectory),
					resource.TestCheckResourceAttr(resourceName, "method", "HANDSHAKE"),
					resource.TestCheckResourceAttr(resourceName, "notes", "test"),
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

func testAccCheckSharedDirectoryExists(name string, share *directoryservice.SharedDirectory) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		ownerId := rs.Primary.Attributes["directory_id"]
		sharedId := rs.Primary.Attributes["shared_directory_id"]

		conn := acctest.Provider.Meta().(*conns.AWSClient).DSConn
		out, err := conn.DescribeSharedDirectories(&directoryservice.DescribeSharedDirectoriesInput{
			OwnerDirectoryId:   aws.String(ownerId),
			SharedDirectoryIds: aws.StringSlice([]string{sharedId}),
		})
		if err != nil {
			return err
		}

		if len(out.SharedDirectories) < 1 {
			return fmt.Errorf("No Shared Directory found")
		}

		if *out.SharedDirectories[0].SharedDirectoryId != sharedId {
			return fmt.Errorf("Shared Directory mismatch - existing: %q, state: %q",
				*out.SharedDirectories[0].SharedDirectoryId, sharedId)
		}

		if *out.SharedDirectories[0].OwnerDirectoryId != ownerId {
			return fmt.Errorf("Owner Directory ID mismatch - existing: %q, state: %q",
				*out.SharedDirectories[0].OwnerDirectoryId, ownerId)
		}

		*share = *out.SharedDirectories[0]

		return nil
	}

}

func testAccCheckSharedDirectoryDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_directory_service_shared_directory" {
			continue
		}

		ownerId := rs.Primary.Attributes["directory_id"]
		sharedId := rs.Primary.Attributes["shared_directory_id"]

		input := directoryservice.DescribeSharedDirectoriesInput{
			OwnerDirectoryId:   aws.String(ownerId),
			SharedDirectoryIds: []*string{aws.String(sharedId)},
		}
		out, err := conn.DescribeSharedDirectories(&input)

		if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) {
			continue
		}

		if err != nil {
			return err
		}

		if out != nil && len(out.SharedDirectories) > 0 {
			return fmt.Errorf("Expected AWS Directory Service Shared Directory to be gone, but was still found")
		}
	}

	return nil
}

func testAccSharedDirectoryConfig_basic(rName, domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		testAccDirectoryConfig_microsoftStandard(rName, domain),
		`
resource "aws_directory_service_shared_directory" "test" {
  directory_id = aws_directory_service_directory.test.id
  notes        = "test"

  target {
    id = data.aws_caller_identity.receiver.account_id
  }
}

data "aws_caller_identity" "receiver" {
  provider = "awsalternate"
}
`)
}
