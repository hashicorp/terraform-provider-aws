package iam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
)

func TestAccIAMSamlProvider_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())
	idpEntityIdModified := fmt.Sprintf("https://%s", acctest.RandomDomainName())
	resourceName := "aws_iam_saml_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMSamlProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMSamlProviderConfig(rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMSamlProviderExists(resourceName),
					acctest.CheckResourceAttrGlobalARN(resourceName, "arn", "iam", fmt.Sprintf("saml-provider/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "saml_metadata_document"),
					resource.TestCheckResourceAttrSet(resourceName, "valid_until"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccIAMSamlProviderConfigUpdate(rName, idpEntityIdModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMSamlProviderExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "saml_metadata_document"),
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

func TestAccIAMSamlProvider_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())
	resourceName := "aws_iam_saml_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMSamlProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMSamlProviderConfigTags1(rName, idpEntityId, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMSamlProviderExists(resourceName),
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
				Config: testAccIAMSamlProviderConfigTags2(rName, idpEntityId, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMSamlProviderExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccIAMSamlProviderConfigTags1(rName, idpEntityId, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMSamlProviderExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccIAMSamlProvider_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())
	resourceName := "aws_iam_saml_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMSamlProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMSamlProviderConfig(rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMSamlProviderExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfiam.ResourceSamlProvider(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIAMSamlProviderDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_saml_provider" {
			continue
		}

		input := &iam.GetSAMLProviderInput{
			SAMLProviderArn: aws.String(rs.Primary.ID),
		}
		out, err := conn.GetSAMLProvider(input)

		if tfawserr.ErrMessageContains(err, iam.ErrCodeNoSuchEntityException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if out != nil {
			return fmt.Errorf("IAM SAML Provider (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckIAMSamlProviderExists(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("Not Found: %s", id)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn
		_, err := conn.GetSAMLProvider(&iam.GetSAMLProviderInput{
			SAMLProviderArn: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccIAMSamlProviderConfig(rName, idpEntityId string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })
}
`, rName, idpEntityId)
}

func testAccIAMSamlProviderConfigUpdate(rName, idpEntityIdModified string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata-modified.xml.tpl", { entity_id_modified = %[2]q })
}
`, rName, idpEntityIdModified)
}

func testAccIAMSamlProviderConfigTags1(rName, idpEntityId, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, idpEntityId, tagKey1, tagValue1)
}

func testAccIAMSamlProviderConfigTags2(rName, idpEntityId, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, idpEntityId, tagKey1, tagValue1, tagKey2, tagValue2)
}
