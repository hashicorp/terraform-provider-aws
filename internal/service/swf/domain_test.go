package swf_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/swf"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccPreCheckDomainTestingEnabled(t *testing.T) {
	if os.Getenv("SWF_DOMAIN_TESTING_ENABLED") == "" {
		t.Skip(
			"Environment variable SWF_DOMAIN_TESTING_ENABLED is not set. " +
				"SWF limits domains per region and the API does not support " +
				"deletions. Set the environment variable to any value to enable.")
	}
}

func TestAccSWFDomain_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_swf_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckDomainTestingEnabled(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, swf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_Name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "swf", regexp.MustCompile(`/domain/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccSWFDomain_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_swf_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckDomainTestingEnabled(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, swf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(resourceName),
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
				Config: testAccDomainTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDomainTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSWFDomain_namePrefix(t *testing.T) {
	resourceName := "aws_swf_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckDomainTestingEnabled(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, swf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_NamePrefix,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile(`^tf-acc-test`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"}, // this line is only necessary if the test configuration is using name_prefix
			},
		},
	})
}

func TestAccSWFDomain_generatedName(t *testing.T) {
	resourceName := "aws_swf_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckDomainTestingEnabled(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, swf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_GeneratedName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(resourceName),
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

func TestAccSWFDomain_description(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_swf_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckDomainTestingEnabled(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, swf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_Description(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
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

func testAccCheckDomainDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SWFConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_swf_domain" {
			continue
		}

		name := rs.Primary.ID
		input := &swf.DescribeDomainInput{
			Name: aws.String(name),
		}

		resp, err := conn.DescribeDomain(input)
		if err != nil {
			return err
		}

		if *resp.DomainInfo.Status != swf.RegistrationStatusDeprecated {
			return fmt.Errorf(`SWF Domain %s status is %s instead of %s. Failing!`, name, *resp.DomainInfo.Status, swf.RegistrationStatusDeprecated)
		}
	}

	return nil
}

func testAccCheckDomainExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SWF Domain not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SWF Domain name not set")
		}

		name := rs.Primary.ID
		conn := acctest.Provider.Meta().(*conns.AWSClient).SWFConn

		input := &swf.DescribeDomainInput{
			Name: aws.String(name),
		}

		resp, err := conn.DescribeDomain(input)
		if err != nil {
			return fmt.Errorf("SWF Domain %s not found in AWS", name)
		}

		if *resp.DomainInfo.Status != swf.RegistrationStatusRegistered {
			return fmt.Errorf(`SWF Domain %s status is %s instead of %s. Failing!`, name, *resp.DomainInfo.Status, swf.RegistrationStatusRegistered)
		}
		return nil
	}
}

func testAccDomainConfig_Description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_swf_domain" "test" {
  description                                 = %q
  name                                        = %q
  workflow_execution_retention_period_in_days = 1
}
`, description, rName)
}

const testAccDomainConfig_GeneratedName = `
resource "aws_swf_domain" "test" {
  workflow_execution_retention_period_in_days = 1
}
`

func testAccDomainConfig_Name(rName string) string {
	return fmt.Sprintf(`
resource "aws_swf_domain" "test" {
  name                                        = %q
  workflow_execution_retention_period_in_days = 1
}
`, rName)
}

func testAccDomainTags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_swf_domain" "test" {
  name                                        = %[1]q
  workflow_execution_retention_period_in_days = 1

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDomainTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_swf_domain" "test" {
  name                                        = %[1]q
  workflow_execution_retention_period_in_days = 1

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

const testAccDomainConfig_NamePrefix = `
resource "aws_swf_domain" "test" {
  name_prefix                                 = "tf-acc-test"
  workflow_execution_retention_period_in_days = 1
}
`
