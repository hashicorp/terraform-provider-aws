// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccLightsailContainerService_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_container_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerServiceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "power", string(types.ContainerServicePowerNameNano)),
					resource.TestCheckResourceAttr(resourceName, "scale", "1"),
					resource.TestCheckResourceAttr(resourceName, "is_disabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone"),
					resource.TestCheckResourceAttrSet(resourceName, "power_id"),
					resource.TestCheckResourceAttrSet(resourceName, "principal_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "private_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "resource_type", "ContainerService"),
					resource.TestCheckResourceAttr(resourceName, "state", "READY"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContainerServiceConfig_scale(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "scale", "2"),
				),
			},
		},
	})
}

func TestAccLightsailContainerService_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_container_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerServiceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflightsail.ResourceContainerService(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLightsailContainerService_name(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_container_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerServiceConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				Config: testAccContainerServiceConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func TestAccLightsailContainerService_isDisabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_container_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerServiceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "is_disabled", "false"),
				),
			},
			{
				Config: testAccContainerServiceConfig_disabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "is_disabled", "true"),
				),
			},
		},
	})
}

func TestAccLightsailContainerService_power(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_container_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerServiceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "power", string(types.ContainerServicePowerNameNano)),
				),
			},
			{
				Config: testAccContainerServiceConfig_power(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "power", string(types.ContainerServicePowerNameMicro)),
				),
			},
		},
	})
}

func TestAccLightsailContainerService_publicDomainNames(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccContainerServiceConfig_publicDomainNames(rName),
				ExpectError: regexp.MustCompile(`do not exist`),
			},
		},
	})
}

func TestAccLightsailContainerService_privateRegistryAccess(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_container_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerServiceConfig_privateRegistryAccess(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "private_registry_access.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "private_registry_access.0.ecr_image_puller_role.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "private_registry_access.0.ecr_image_puller_role.0.is_active", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "private_registry_access.0.ecr_image_puller_role.0.principal_arn"),
				),
			},
		},
	})
}

func TestAccLightsailContainerService_scale(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_container_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerServiceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "scale", "1"),
				),
			},
			{
				Config: testAccContainerServiceConfig_scale(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "scale", "2"),
				),
			},
		},
	})
}

func TestAccLightsailContainerService_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_container_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(lightsail.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(lightsail.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerServiceConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContainerServiceExists(ctx, resourceName),
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
				Config: testAccContainerServiceConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContainerServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccContainerServiceConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContainerServiceExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckContainerServiceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)

		for _, r := range s.RootModule().Resources {
			if r.Type != "aws_lightsail_container_service" {
				continue
			}

			_, err := tflightsail.FindContainerServiceByName(ctx, conn, r.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccCheckContainerServiceExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not finding Lightsail Container Service (%s)", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Lightsail Container Service ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)

		_, err := tflightsail.FindContainerServiceByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccContainerServiceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name  = %q
  power = "nano"
  scale = 1
}
`, rName)
}

func testAccContainerServiceConfig_disabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name        = %q
  power       = "nano"
  scale       = 1
  is_disabled = true
}
`, rName)
}

func testAccContainerServiceConfig_power(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name  = %q
  power = "micro"
  scale = 1
}
`, rName)
}

func testAccContainerServiceConfig_publicDomainNames(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name  = %q
  power = "nano"
  scale = 1
  public_domain_names {
    certificate {
      certificate_name = "NonExsitingCertificate"
      domain_names = [
        "nonexisting1.com",
        "nonexisting2.com",
      ]
    }
  }
}
`, rName)
}

func testAccContainerServiceConfig_privateRegistryAccess(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name  = %q
  power = "micro"
  scale = 1

  private_registry_access {
    ecr_image_puller_role {
      is_active = true
    }
  }
}
`, rName)
}

func testAccContainerServiceConfig_scale(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name  = %q
  power = "nano"
  scale = 2
}
`, rName)
}

func testAccContainerServiceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name  = %[1]q
  power = "nano"
  scale = 1
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccContainerServiceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name  = %q
  power = "nano"
  scale = 1
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
