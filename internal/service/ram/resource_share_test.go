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
	tfram "github.com/hashicorp/terraform-provider-aws/internal/service/ram"
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
				Config: testAccResourceShareConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(resourceName, &resourceShare),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ram", regexp.MustCompile(`resource-share/.+`)),
					resource.TestCheckResourceAttr(resourceName, "allow_external_principals", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "permission_arns.#", "0"),
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

func TestAccRAMResourceShare_permission(t *testing.T) {
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
				Config: testAccResourceShareConfig_namePermission(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(resourceName, &resourceShare),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ram", regexp.MustCompile(`resource-share/.+`)),
					resource.TestCheckResourceAttr(resourceName, "allow_external_principals", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "permission_arns.#", "1"),
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
				Config: testAccResourceShareConfig_allowExternalPrincipals(rName, false),
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
				Config: testAccResourceShareConfig_allowExternalPrincipals(rName, true),
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
				Config: testAccResourceShareConfig_name(rName),
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
				Config: testAccResourceShareConfig_name(rName2),
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
				Config: testAccResourceShareConfig_tags1(rName, "key1", "value1"),
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
				Config: testAccResourceShareConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(resourceName, &resourceShare2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccResourceShareConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(resourceName, &resourceShare3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccRAMResourceShare_disappears(t *testing.T) {
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
				Config: testAccResourceShareConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(resourceName, &resourceShare),
					acctest.CheckResourceDisappears(acctest.Provider, tfram.ResourceResourceShare(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckResourceShareExists(resourceName string, v *ram.ResourceShare) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RAMConn

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		output, err := tfram.FindResourceShareOwnerSelfByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if aws.StringValue(output.Status) != ram.ResourceShareStatusActive {
			return fmt.Errorf("RAM resource share (%s) delet(ing|ed)", rs.Primary.ID)
		}

		*v = *output

		return nil
	}
}

func testAccCheckResourceShareDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ram_resource_share" {
			continue
		}

		resourceShare, err := tfram.FindResourceShareOwnerSelfByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if aws.StringValue(resourceShare.Status) != ram.ResourceShareStatusDeleted {
			return fmt.Errorf("RAM resource share (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccResourceShareConfig_allowExternalPrincipals(rName string, allowExternalPrincipals bool) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  allow_external_principals = %[1]t
  name                      = %[2]q
}
`, allowExternalPrincipals, rName)
}

func testAccResourceShareConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  name = %[1]q
}
`, rName)
}

func testAccResourceShareConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccResourceShareConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccResourceShareConfig_namePermission(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_ram_resource_share" "test" {
  name            = %[1]q
  permission_arns = ["arn:${data.aws_partition.current.partition}:ram::aws:permission/AWSRAMBlankEndEntityCertificateAPICSRPassthroughIssuanceCertificateAuthority"]
}
`, rName)
}
