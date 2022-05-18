package servicediscovery_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfservicediscovery "github.com/hashicorp/terraform-provider-aws/internal/service/servicediscovery"
)

func TestAccServiceDiscoveryPrivateDNSNamespace_basic(t *testing.T) {
	resourceName := "aws_service_discovery_private_dns_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(servicediscovery.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, servicediscovery.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPrivateDNSNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateDNSNamespaceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateDNSNamespaceExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "servicediscovery", regexp.MustCompile(`namespace/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "hosted_zone"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPrivateDNSNamespaceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccServiceDiscoveryPrivateDNSNamespace_disappears(t *testing.T) {
	resourceName := "aws_service_discovery_private_dns_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(servicediscovery.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, servicediscovery.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPrivateDNSNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateDNSNamespaceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateDNSNamespaceExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfservicediscovery.ResourcePrivateDNSNamespace(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServiceDiscoveryPrivateDNSNamespace_description(t *testing.T) {
	resourceName := "aws_service_discovery_private_dns_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(servicediscovery.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, servicediscovery.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPrivateDNSNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateDNSNamespaceConfig_description(rName, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateDNSNamespaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
				),
			},
		},
	})
}

// This acceptance test ensures we properly send back error messaging. References:
//  * https://github.com/hashicorp/terraform-provider-aws/issues/2830
//  * https://github.com/hashicorp/terraform-provider-aws/issues/5532
func TestAccServiceDiscoveryPrivateDNSNamespace_Error_overlap(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(servicediscovery.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, servicediscovery.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPrivateDNSNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccPrivateDNSNamespaceConfig_overlapping(rName),
				ExpectError: regexp.MustCompile(`ConflictingDomainExists`),
			},
		},
	})
}

func TestAccServiceDiscoveryPrivateDNSNamespace_tags(t *testing.T) {
	resourceName := "aws_service_discovery_private_dns_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(servicediscovery.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, servicediscovery.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPrivateDNSNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateDNSNamespaceConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateDNSNamespaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccPrivateDNSNamespaceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccPrivateDNSNamespaceConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateDNSNamespaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccPrivateDNSNamespaceConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateDNSNamespaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckPrivateDNSNamespaceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceDiscoveryConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_service_discovery_private_dns_namespace" {
			continue
		}

		input := &servicediscovery.GetNamespaceInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetNamespace(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, servicediscovery.ErrCodeNamespaceNotFound) {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccCheckPrivateDNSNamespaceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceDiscoveryConn

	input := &servicediscovery.ListNamespacesInput{}

	_, err := conn.ListNamespaces(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPrivateDNSNamespaceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name = "%[1]s.tf"
  vpc  = aws_vpc.test.id
}
`, rName)
}

func testAccPrivateDNSNamespaceConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[2]q
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  description = %[1]q
  name        = "%[2]s.tf"
  vpc         = aws_vpc.test.id
}
`, description, rName)
}

func testAccPrivateDNSNamespaceConfig_overlapping(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_service_discovery_private_dns_namespace" "top" {
  name = "%[1]s.tf"
  vpc  = aws_vpc.test.id
}

# Ensure ordering after first namespace
resource "aws_service_discovery_private_dns_namespace" "subdomain" {
  name = aws_service_discovery_private_dns_namespace.top.name
  vpc  = aws_service_discovery_private_dns_namespace.top.vpc
}
`, rName)
}

func testAccPrivateDNSNamespaceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name = "%[1]s.tf"
  vpc  = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccPrivateDNSNamespaceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name = "%[1]s.tf"
  vpc  = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccPrivateDNSNamespaceImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return fmt.Sprintf("%s:%s", rs.Primary.ID, rs.Primary.Attributes["vpc"]), nil
	}
}
