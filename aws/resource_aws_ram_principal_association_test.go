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

func TestAccAwsRamPrincipalAssociation_basic(t *testing.T) {
	var association ram.ResourceShareAssociation
	resourceName := "aws_ram_principal_association.example"
	shareName := fmt.Sprintf("tf-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	principal := "111111111111"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRamPrincipalAssociationConfig_basic(shareName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRamPrincipalAssociationExists(resourceName, &association),
					resource.TestCheckResourceAttr(resourceName, "name", shareName),
					resource.TestCheckResourceAttr(resourceName, "allow_external_principals", "true"),
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

func testAccCheckAwsRamPrincipalAssociationExists(resourceName string, resourceShare *ram.ResourceShareAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ramconn

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		request := &ram.GetResourceShareAssociationsInput{
			ResourceShareArns: []*string{aws.String(rs.Primary.ID)},
			ResourceOwner:     aws.String(ram.ResourceOwnerSelf),
		}

		output, err := conn.GetResourceShareAssociations(request)
		if err != nil {
			return err
		}

		if len(output.ResourceShareAssociations) == 0 {
			return fmt.Errorf("No RAM resource share found")
		}

		resourceShare = output.ResourceShareAssociations[0]

		if aws.StringValue(resourceShare.Status) != ram.ResourceShareAssociationStatusActive {
			return fmt.Errorf("RAM resource share (%s) delet(ing|ed)", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAwsRamPrincipalAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ramconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ram_resource_share" {
			continue
		}

		request := &ram.GetResourceShareAssociationsInput{
			ResourceShareArns: []*string{aws.String(rs.Primary.ID)},
			ResourceOwner:     aws.String(ram.ResourceOwnerSelf),
		}

		output, err := conn.GetResourceShareAssociations(request)
		if err != nil {
			return err
		}

		if len(output.ResourceShareAssociations) > 0 {
			resourceShare := output.ResourceShareAssociations[0]
			if aws.StringValue(resourceShare.Status) != ram.ResourceShareAssociationStatusDeleted {
				return fmt.Errorf("RAM resource share (%s) still exists", rs.Primary.ID)
			}
			return fmt.Errorf("No RAM resource share found")
		}
	}

	return nil
}

func testAccAwsRamPrincipalAssociationConfig_basic(shareName, principal string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "example" {
  name                      = "%s"
  allow_external_principals = true

  tags {
    Environment = "Production"
  }
}

resource "aws_ram_principal_association" "example" {
  resource_share_arn = "${aws_ram_resource_share.example.id}"
  principal          = "%s"
}
`, shareName)
}
