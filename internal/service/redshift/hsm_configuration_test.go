package redshift_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/redshift"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRedshiftHSMConfiguration_basic(t *testing.T) {
	resourceName := "aws_redshift_hsm_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckHSMConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHSMConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHSMConfigurationExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "redshift", regexp.MustCompile(`hsmconfiguration:.+`)),
					resource.TestCheckResourceAttr(resourceName, "hsm_configuration_identifier", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"hsm_partition_password", "hsm_server_public_certificate"},
			},
		},
	})
}

func TestAccRedshiftHSMConfiguration_tags(t *testing.T) {
	resourceName := "aws_redshift_hsm_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckHSMConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHSMConfigurationConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHSMConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"hsm_partition_password", "hsm_server_public_certificate"},
			},
			{
				Config: testAccHSMConfigurationConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHSMConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			}, {
				Config: testAccHSMConfigurationConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHSMConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccRedshiftHSMConfiguration_disappears(t *testing.T) {
	resourceName := "aws_redshift_hsm_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckHSMConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccHSMConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHSMConfigurationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfredshift.ResourceHSMConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckHSMConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_hsm_configuration" {
			continue
		}

		_, err := tfredshift.FindHSMConfigurationByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Redshift Hsm Configuration %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckHSMConfigurationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Redshift Hsm Configuration is not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

		_, err := tfredshift.FindHSMConfigurationByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccHSMConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_hsm_configuration" "test" {
  description                   = %[1]q
  hsm_configuration_identifier  = %[1]q
  hsm_ip_address                = "10.0.0.1"
  hsm_partition_name            = "aws"
  hsm_partition_password        = %[1]q
  hsm_server_public_certificate = %[1]q
}
`, rName)
}

func testAccHSMConfigurationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_redshift_hsm_configuration" "test" {
  description                   = %[1]q
  hsm_configuration_identifier  = %[1]q
  hsm_ip_address                = "10.0.0.1"
  hsm_partition_name            = "aws"
  hsm_partition_password        = %[1]q
  hsm_server_public_certificate = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccHSMConfigurationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_redshift_hsm_configuration" "test" {
  description                   = %[1]q
  hsm_configuration_identifier  = %[1]q
  hsm_ip_address                = "10.0.0.1"
  hsm_partition_name            = "aws"
  hsm_partition_password        = %[1]q
  hsm_server_public_certificate = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
