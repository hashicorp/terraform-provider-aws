package ram_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfram "github.com/hashicorp/terraform-provider-aws/internal/service/ram"
)

func TestAccRAMResourceAssociation_basic(t *testing.T) {
	var resourceShareAssociation1 ram.ResourceShareAssociation
	resourceName := "aws_ram_resource_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ram.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAssociationExists(resourceName, &resourceShareAssociation1),
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

func TestAccRAMResourceAssociation_disappears(t *testing.T) {
	var resourceShareAssociation1 ram.ResourceShareAssociation
	resourceName := "aws_ram_resource_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ram.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAssociationExists(resourceName, &resourceShareAssociation1),
					testAccCheckResourceAssociationDisappears(&resourceShareAssociation1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckResourceAssociationDisappears(resourceShareAssociation *ram.ResourceShareAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RAMConn

		input := &ram.DisassociateResourceShareInput{
			ResourceArns:     []*string{resourceShareAssociation.AssociatedEntity},
			ResourceShareArn: resourceShareAssociation.ResourceShareArn,
		}

		_, err := conn.DisassociateResourceShare(input)
		if err != nil {
			return err
		}

		return tfram.WaitForResourceShareResourceDisassociation(conn, aws.StringValue(resourceShareAssociation.ResourceShareArn), aws.StringValue(resourceShareAssociation.AssociatedEntity))
	}
}

func testAccCheckResourceAssociationExists(resourceName string, association *ram.ResourceShareAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RAMConn

		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		resourceShareARN, resourceARN, err := tfram.DecodeResourceAssociationID(rs.Primary.ID)

		if err != nil {
			return err
		}

		resourceShareAssociation, err := tfram.GetResourceShareAssociation(conn, resourceShareARN, resourceARN)

		if err != nil {
			return fmt.Errorf("error reading RAM Resource Share (%s) Resource Association (%s): %s", resourceShareARN, resourceARN, err)
		}

		if resourceShareAssociation == nil {
			return fmt.Errorf("RAM Resource Share (%s) Resource Association (%s) not found", resourceShareARN, resourceARN)
		}

		if aws.StringValue(resourceShareAssociation.Status) != ram.ResourceShareAssociationStatusAssociated {
			return fmt.Errorf("RAM Resource Share (%s) Resource Association (%s) not associated: %s", resourceShareARN, resourceARN, aws.StringValue(resourceShareAssociation.Status))
		}

		*association = *resourceShareAssociation

		return nil
	}
}

func testAccCheckResourceAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ram_resource_association" {
			continue
		}

		resourceShareARN, resourceARN, err := tfram.DecodeResourceAssociationID(rs.Primary.ID)

		if err != nil {
			return err
		}

		resourceShareAssociation, err := tfram.GetResourceShareAssociation(conn, resourceShareARN, resourceARN)

		if err != nil {
			return err
		}

		if resourceShareAssociation != nil && aws.StringValue(resourceShareAssociation.Status) != ram.ResourceShareAssociationStatusDisassociated {
			return fmt.Errorf("RAM Resource Share (%s) Resource Association (%s) not disassociated: %s", resourceShareARN, resourceARN, aws.StringValue(resourceShareAssociation.Status))
		}
	}

	return nil
}

func testAccResourceAssociationConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ram-resource-association"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ram-resource-association"
  }
}

resource "aws_ram_resource_share" "test" {
  name = %q
}

resource "aws_ram_resource_association" "test" {
  resource_arn       = aws_subnet.test.arn
  resource_share_arn = aws_ram_resource_share.test.id
}
`, rName)
}
