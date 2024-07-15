// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfservicediscovery "github.com/hashicorp/terraform-provider-aws/internal/service/servicediscovery"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceDiscoveryInstance_private(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_service_discovery_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_private(rName, domainName, "AWS_INSTANCE_IPV4 = \"10.0.0.1\" \n    AWS_INSTANCE_IPV6 = \"2001:0db8:85a3:0000:0000:abcd:0001:2345\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, rName),
					resource.TestCheckResourceAttr(resourceName, "attributes.AWS_INSTANCE_IPV4", "10.0.0.1"),
					resource.TestCheckResourceAttr(resourceName, "attributes.AWS_INSTANCE_IPV6", "2001:0db8:85a3:0000:0000:abcd:0001:2345"),
				),
			},
			{
				Config: testAccInstanceConfig_private(rName, domainName, "AWS_INSTANCE_IPV4 = \"10.0.0.2\" \n    AWS_INSTANCE_IPV6 = \"2001:0db8:85a3:0000:0000:abcd:0001:2345\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, rName),
					resource.TestCheckResourceAttr(resourceName, "attributes.AWS_INSTANCE_IPV4", "10.0.0.2"),
					resource.TestCheckResourceAttr(resourceName, "attributes.AWS_INSTANCE_IPV6", "2001:0db8:85a3:0000:0000:abcd:0001:2345"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccInstanceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccServiceDiscoveryInstance_public(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_service_discovery_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_public(rName, domainName, "AWS_INSTANCE_IPV4 = \"52.18.0.2\" \n    CUSTOM_KEY = \"this is a custom value\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, rName),
					resource.TestCheckResourceAttr(resourceName, "attributes.AWS_INSTANCE_IPV4", "52.18.0.2"),
					resource.TestCheckResourceAttr(resourceName, "attributes.CUSTOM_KEY", "this is a custom value"),
				),
			},
			{
				Config: testAccInstanceConfig_public(rName, domainName, "AWS_INSTANCE_IPV4 = \"52.18.0.2\" \n    CUSTOM_KEY = \"this is a custom value updated\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, rName),
					resource.TestCheckResourceAttr(resourceName, "attributes.AWS_INSTANCE_IPV4", "52.18.0.2"),
					resource.TestCheckResourceAttr(resourceName, "attributes.CUSTOM_KEY", "this is a custom value updated"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccInstanceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccServiceDiscoveryInstance_http(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_service_discovery_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ServiceDiscoveryEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceDiscoveryServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_http(rName, domainName, "AWS_EC2_INSTANCE_ID = aws_instance.test.id"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, rName),
					resource.TestCheckResourceAttrSet(resourceName, "attributes.AWS_EC2_INSTANCE_ID"),
				),
			},
			{
				Config: testAccInstanceConfig_http(rName, domainName, "AWS_INSTANCE_IPV4 = \"172.18.0.12\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, rName),
					resource.TestCheckResourceAttr(resourceName, "attributes.AWS_INSTANCE_IPV4", "172.18.0.12"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccInstanceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckInstanceExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceDiscoveryClient(ctx)

		_, err := tfservicediscovery.FindInstanceByTwoPartKey(ctx, conn, rs.Primary.Attributes["service_id"], rs.Primary.Attributes[names.AttrInstanceID])

		return err
	}
}

func testAccInstanceImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["service_id"], rs.Primary.Attributes[names.AttrInstanceID]), nil
	}
}

func testAccCheckInstanceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceDiscoveryClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_service_discovery_instance" {
				continue
			}

			_, err := tfservicediscovery.FindInstanceByTwoPartKey(ctx, conn, rs.Primary.Attributes["service_id"], rs.Primary.Attributes[names.AttrInstanceID])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Service Discovery Instance %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccInstanceConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}`, rName)
}

func testAccInstanceConfig_private(rName, domainName, attributes string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_base(rName),
		testAccInstanceConfig_privateNamespace(rName, domainName),
		testAccInstanceConfig_basic(rName, attributes),
	)
}

func testAccInstanceConfig_public(rName, domainName, attributes string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_base(rName),
		testAccInstanceConfig_publicNamespace(rName, domainName),
		testAccInstanceConfig_basic(rName, attributes),
	)
}

func testAccInstanceConfig_http(rName, domainName, attributes string) string {
	return acctest.ConfigCompose(
		testAccInstanceConfig_base(rName),
		testAccInstanceConfig_httpNamespace(rName, domainName),
		testAccInstanceConfig_basic(rName, attributes),
	)
}

func testAccInstanceConfig_privateNamespace(rName, domainName string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_private_dns_namespace" "test" {
  name        = %[2]q
  description = %[1]q
  vpc         = aws_vpc.test.id
}

resource "aws_service_discovery_service" "test" {
  name = %[1]q

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.test.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {
    failure_threshold = 1
  }
}`, rName, domainName)
}

func testAccInstanceConfig_publicNamespace(rName, domainName string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_public_dns_namespace" "test" {
  name = %[2]q
}

resource "aws_service_discovery_service" "test" {
  name = %[1]q

  dns_config {
    namespace_id = aws_service_discovery_public_dns_namespace.test.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {
    failure_threshold = 1
  }
}`, rName, domainName)
}

func testAccInstanceConfig_httpNamespace(rName, domainName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_instance" "test" {
  instance_type = "t2.micro"
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_service_discovery_http_namespace" "test" {
  name = %[2]q
}

resource "aws_service_discovery_service" "test" {
  name         = %[1]q
  namespace_id = aws_service_discovery_http_namespace.test.id
}`, rName, domainName))
}

func testAccInstanceConfig_basic(instanceID string, attributes string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_instance" "test" {
  service_id  = aws_service_discovery_service.test.id
  instance_id = %[1]q

  attributes = {
    %[2]s
  }
}`, instanceID, attributes)
}
