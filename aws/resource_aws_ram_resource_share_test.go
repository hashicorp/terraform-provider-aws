package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
)

func TestAccAwsRamResourceShare_basic(t *testing.T) {
	var resourceShare ram.ResourceShare
	resourceName := "aws_ram_resource_share.example"
	shareName := fmt.Sprintf("tf-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRamResourceShareConfig_basic(shareName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRamResourceShareExists(resourceName, &resourceShare),
					resource.TestCheckResourceAttr(resourceName, "name", shareName),
					resource.TestCheckResourceAttr(resourceName, "allow_external_principals", "true"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "Production"),
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

func testAccCheckAwsRamResourceShareExists(resourceName string, resourceShare *ram.ResourceShare) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ramconn

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		request := &ram.GetResourceSharesInput{
			ResourceShareArns: []*string{aws.String(rs.Primary.ID)},
			ResourceOwner:     aws.String(ram.ResourceOwnerSelf),
		}

		output, err := conn.GetResourceShares(request)
		if err != nil {
			return err
		}

		if len(output.ResourceShares) == 0 {
			return fmt.Errorf("No RAM resource share found")
		}

		resourceShare = output.ResourceShares[0]

		if aws.StringValue(resourceShare.Status) != ram.ResourceShareStatusActive {
			return fmt.Errorf("RAM resource share (%s) delet(ing|ed)", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAwsRamResourceShareDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ramconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ram_resource_share" {
			continue
		}

		request := &ram.GetResourceSharesInput{
			ResourceShareArns: []*string{aws.String(rs.Primary.ID)},
			ResourceOwner:     aws.String(ram.ResourceOwnerSelf),
		}

		output, err := conn.GetResourceShares(request)
		if err != nil {
			return err
		}

		if len(output.ResourceShares) > 0 {
			resourceShare := output.ResourceShares[0]
			if aws.StringValue(resourceShare.Status) != ram.ResourceShareStatusDeleted {
				return fmt.Errorf("RAM resource share (%s) still exists", rs.Primary.ID)
			}
			return fmt.Errorf("No RAM resource share found")
		}
	}

	return nil
}

func testAccAwsRamResourceShareConfig_basic(shareName string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "example" {
  name                      = "%s"
  allow_external_principals = true

  tags {
    Environment = "Production"
  }
}
`, shareName)
}
