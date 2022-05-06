package ram_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccRAMResourceShare_basic(t *testing.T) {
	var resourceShare ram.ResourceShare
	resourceName := "aws_ram_resource_share.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ram.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(resourceName, &resourceShare),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ram", regexp.MustCompile(`resource-share/.+`)),
					resource.TestCheckResourceAttr(resourceName, "allow_external_principals", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccRAMResourceShare_allowExternalPrincipals(t *testing.T) {
	var resourceShare1, resourceShare2 ram.ResourceShare
	resourceName := "aws_ram_resource_share.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ram.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareAllowExternalPrincipalsConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(resourceName, &resourceShare1),
					resource.TestCheckResourceAttr(resourceName, "allow_external_principals", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceShareAllowExternalPrincipalsConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(resourceName, &resourceShare2),
					resource.TestCheckResourceAttr(resourceName, "allow_external_principals", "true"),
				),
			},
		},
	})
}

func TestAccRAMResourceShare_name(t *testing.T) {
	var resourceShare1, resourceShare2 ram.ResourceShare
	resourceName := "aws_ram_resource_share.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ram.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(resourceName, &resourceShare1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceShareNameConfig(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(resourceName, &resourceShare2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func TestAccRAMResourceShare_tags(t *testing.T) {
	var resourceShare1, resourceShare2, resourceShare3 ram.ResourceShare
	resourceName := "aws_ram_resource_share.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ram.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckResourceShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(resourceName, &resourceShare1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceShareTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(resourceName, &resourceShare2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccResourceShareTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(resourceName, &resourceShare3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckResourceShareExists(resourceName string, resourceShare *ram.ResourceShare) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RAMConn

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

func testAccCheckResourceShareDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RAMConn

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
		}
	}

	return nil
}

func testAccResourceShareAllowExternalPrincipalsConfig(rName string, allowExternalPrincipals bool) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  allow_external_principals = %[1]t
  name                      = %[2]q
}
`, allowExternalPrincipals, rName)
}

func testAccResourceShareNameConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  name = %[1]q
}
`, rName)
}

func testAccResourceShareTags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccResourceShareTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
