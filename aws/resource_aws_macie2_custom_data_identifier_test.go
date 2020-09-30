package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSMacie2CustomDataIdentifier_basic(t *testing.T) {
	resourceName := "aws_macie2_custom_data_identifier.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMacie2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSMacie2CustomDataIdentifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSMacie2CustomDataIdentifierConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-%d", rInt)),
					resource.TestCheckResourceAttr(resourceName, "regex", "\\d"),
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

func TestAccAWSMacie2CustomDataIdentifier_full(t *testing.T) {
	resourceName := "aws_macie2_custom_data_identifier.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMacie2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSMacie2CustomDataIdentifierDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSMacie2CustomDataIdentifierConfig_full(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-%d", rInt)),
					resource.TestCheckResourceAttr(resourceName, "regex", "\\d"),
					resource.TestCheckResourceAttr(resourceName, "description", "Terraform test"),
					resource.TestCheckResourceAttr(resourceName, "ignore_words.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "keywords.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "maximum_match_distance", "20"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.baz", "qux"),
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

func testAccCheckAWSMacie2CustomDataIdentifierDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).macie2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie2_custom_data_identifier" {
			continue
		}

		req := &macie2.GetCustomDataIdentifierInput{
			Id: aws.String(rs.Primary.ID),
		}

		out, err := conn.GetCustomDataIdentifier(req)
		if err != nil {
			return err
		}
		if !*out.Deleted {
			return fmt.Errorf("Custom Data Identifier %s is not deleted", rs.Primary.ID)
		}
	}
	return nil
}

func testAccPreCheckAWSMacie2(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).macie2conn

	input := &macie2.GetMacieSessionInput{}

	_, err := conn.GetMacieSession(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSMacie2CustomDataIdentifierConfig_basic(randInt int) string {
	return fmt.Sprintf(`
resource "aws_macie2_custom_data_identifier" "test" {
  name  = "tf-%d"
  regex = "\\d"
}
`, randInt)
}

func testAccAWSMacie2CustomDataIdentifierConfig_full(randInt int) string {
	return fmt.Sprintf(`
resource "aws_macie2_custom_data_identifier" "test" {
  name                   = "tf-%d"
  regex                  = "\\d"
  description            = "Terraform test"
  ignore_words           = ["test", "ignore", "words"]
  keywords               = ["important", "things"]
  maximum_match_distance = 20
  tags = {
    foo = "bar"
    baz = "qux"
  }
}
`, randInt)
}
