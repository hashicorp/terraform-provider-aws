// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/directoryservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfds "github.com/hashicorp/terraform-provider-aws/internal/service/ds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccDSRegion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v directoryservice.RegionDescription
	resourceName := "aws_directory_service_region.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, directoryservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckRegionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegionConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRegionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "desired_number_of_domain_controllers", "2"),
					resource.TestCheckResourceAttr(resourceName, "region_name", acctest.AlternateRegion()),
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

func TestAccDSRegion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v directoryservice.RegionDescription
	resourceName := "aws_directory_service_region.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, directoryservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckRegionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegionConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegionExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfds.ResourceRegion(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDSRegion_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v directoryservice.RegionDescription
	resourceName := "aws_directory_service_region.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, directoryservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckRegionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegionConfig_tags1(rName, domainName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegionExists(ctx, resourceName, &v),
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
				Config: testAccRegionConfig_tags2(rName, domainName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRegionConfig_tags1(rName, domainName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccDSRegion_desiredNumberOfDomainControllers(t *testing.T) {
	ctx := acctest.Context(t)
	var v directoryservice.RegionDescription
	resourceName := "aws_directory_service_region.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, directoryservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckRegionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegionConfig_desiredNumberOfDomainControllers(rName, domainName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "desired_number_of_domain_controllers", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRegionConfig_desiredNumberOfDomainControllers(rName, domainName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegionExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "desired_number_of_domain_controllers", "3"),
				),
			},
		},
	})
}

func testAccCheckRegionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_directory_service_region" {
				continue
			}

			directoryID, regionName, err := tfds.RegionParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfds.FindRegion(ctx, conn, directoryID, regionName)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Directory Service Region %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRegionExists(ctx context.Context, n string, v *directoryservice.RegionDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Directory Service Region ID is set")
		}

		directoryID, regionName, err := tfds.RegionParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DSConn(ctx)

		output, err := tfds.FindRegion(ctx, conn, directoryID, regionName)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccRegionConfig_base(rName, domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
data "aws_region" "secondary" {
  provider = awsalternate
}

data "aws_availability_zones" "secondary" {
  provider = awsalternate

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "secondary" {
  provider = awsalternate

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "secondary" {
  provider = awsalternate

  count = 2

  vpc_id            = aws_vpc.secondary.id
  availability_zone = data.aws_availability_zones.secondary.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.secondary.cidr_block, 8, count.index)

  tags = {
    Name = %[1]q
  }
}

resource "aws_directory_service_directory" "test" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}
`, rName, domain))
}

func testAccRegionConfig_basic(rName, domain string) string {
	return acctest.ConfigCompose(testAccRegionConfig_base(rName, domain), `
# The DS Region is provisioned in the primary directory's region
# but references VPC/subnets in the secondary directory's region.
resource "aws_directory_service_region" "test" {
  directory_id = aws_directory_service_directory.test.id
  region_name  = data.aws_region.secondary.name

  vpc_settings {
    vpc_id     = aws_vpc.secondary.id
    subnet_ids = aws_subnet.secondary[*].id
  }
}
`)
}

func testAccRegionConfig_tags1(rName, domain, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccRegionConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_directory_service_region" "test" {
  directory_id = aws_directory_service_directory.test.id
  region_name  = data.aws_region.secondary.name

  vpc_settings {
    vpc_id     = aws_vpc.secondary.id
    subnet_ids = aws_subnet.secondary[*].id
  }

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccRegionConfig_tags2(rName, domain, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccRegionConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_directory_service_region" "test" {
  directory_id = aws_directory_service_directory.test.id
  region_name  = data.aws_region.secondary.name

  vpc_settings {
    vpc_id     = aws_vpc.secondary.id
    subnet_ids = aws_subnet.secondary[*].id
  }

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccRegionConfig_desiredNumberOfDomainControllers(rName, domain string, desiredNumber int) string {
	return acctest.ConfigCompose(testAccRegionConfig_base(rName, domain), fmt.Sprintf(`
resource "aws_directory_service_region" "test" {
  directory_id = aws_directory_service_directory.test.id
  region_name  = data.aws_region.secondary.name

  vpc_settings {
    vpc_id     = aws_vpc.secondary.id
    subnet_ids = aws_subnet.secondary[*].id
  }

  desired_number_of_domain_controllers = %[1]d
}
`, desiredNumber))
}
