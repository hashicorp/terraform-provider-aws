// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package gamelift_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/gamelift/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfgamelift "github.com/hashicorp/terraform-provider-aws/internal/service/gamelift"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGameLiftContainerGroupDefinition_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var containerGroup awstypes.ContainerGroupDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_gamelift_container_group_definition.test"
	imageURI := acctest.SkipIfEnvVarNotSet(t, "AWS_GAMELIFT_MANAGED_CONTAINERS_IMAGE_URI")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GameLiftEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerGroupDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerGroupDefinitionConfig_basic(rName, imageURI, 1024, "initial"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerGroupDefinitionExists(ctx, resourceName, &containerGroup),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "container_group_type", "GAME_SERVER"),
					resource.TestCheckResourceAttr(resourceName, "operating_system", "AMAZON_LINUX_2023"),
					resource.TestCheckResourceAttr(resourceName, "total_memory_limit_mib", "1024"),
					resource.TestCheckResourceAttr(resourceName, "total_vcpu_limit", "1"),
					resource.TestCheckResourceAttr(resourceName, "game_server_container_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "game_server_container_definition.0.container_name", "game-server"),
					resource.TestCheckResourceAttr(resourceName, "game_server_container_definition.0.server_sdk_version", "5.2.0"),
					resource.TestCheckResourceAttr(resourceName, "game_server_container_definition.0.port_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "game_server_container_definition.0.port_configuration.0.container_port_ranges.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "game_server_container_definition.0.port_configuration.0.container_port_ranges.0.from_port", "7777"),
					resource.TestCheckResourceAttr(resourceName, "game_server_container_definition.0.port_configuration.0.container_port_ranges.0.to_port", "7778"),
					resource.TestCheckResourceAttr(resourceName, "game_server_container_definition.0.port_configuration.0.container_port_ranges.0.protocol", "UDP"),
					resource.TestCheckResourceAttr(resourceName, "version_description", "initial"),
					resource.TestCheckResourceAttr(resourceName, "version_number", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`containergroupdefinition/.+:\d+`)),
					resource.TestMatchResourceAttr(resourceName, names.AttrCreationTime, regexache.MustCompile(acctest.RFC3339RegexPattern)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContainerGroupDefinitionConfig_basic(rName, imageURI, 2048, "updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerGroupDefinitionExists(ctx, resourceName, &containerGroup),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "total_memory_limit_mib", "2048"),
					resource.TestCheckResourceAttr(resourceName, "version_description", "updated"),
					resource.TestCheckResourceAttr(resourceName, "version_number", "2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "READY"),
				),
			},
		},
	})
}

func TestAccGameLiftContainerGroupDefinition_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var containerGroup awstypes.ContainerGroupDefinition

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_gamelift_container_group_definition.test"
	imageURI := acctest.SkipIfEnvVarNotSet(t, "AWS_GAMELIFT_MANAGED_CONTAINERS_IMAGE_URI")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GameLiftEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerGroupDefinitionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerGroupDefinitionConfig_tags(rName, imageURI, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerGroupDefinitionExists(ctx, resourceName, &containerGroup),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContainerGroupDefinitionConfig_tags(rName, imageURI, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerGroupDefinitionExists(ctx, resourceName, &containerGroup),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckContainerGroupDefinitionExists(ctx context.Context, n string, v *awstypes.ContainerGroupDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		name, version, err := testAccParseContainerGroupDefinitionID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftClient(ctx)

		output, err := tfgamelift.FindContainerGroupDefinition(ctx, conn, name, version)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckContainerGroupDefinitionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_gamelift_container_group_definition" {
				continue
			}

			name, version, err := testAccParseContainerGroupDefinitionID(rs.Primary.ID)
			if err != nil {
				return err
			}

			definition, err := tfgamelift.FindContainerGroupDefinition(ctx, conn, name, version)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			if definition != nil {
				return fmt.Errorf("GameLift Container Group Definition (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccParseContainerGroupDefinitionID(id string) (string, *int32, error) {
	parts := strings.Split(id, ",")
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("unexpected container group definition ID format: %q", id)
	}

	version, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", nil, fmt.Errorf("parsing version from container group definition ID %q: %w", id, err)
	}

	version32 := int32(version)
	return parts[0], &version32, nil
}

func testAccContainerGroupDefinitionConfig_basic(rName, imageURI string, totalMemory int, versionDescription string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_container_group_definition" "test" {
  name                   = %[1]q
  container_group_type   = "GAME_SERVER"
  operating_system       = "AMAZON_LINUX_2023"
  total_memory_limit_mib = %[2]d
  total_vcpu_limit       = 1
  version_description    = %[3]q

  game_server_container_definition {
    container_name     = "game-server"
    image_uri          = %[4]q
    server_sdk_version = "5.2.0"

    port_configuration {
      container_port_ranges {
        from_port = 7777
        to_port   = 7778
        protocol  = "UDP"
      }
    }
  }
}
`, rName, totalMemory, versionDescription, imageURI)
}

func testAccContainerGroupDefinitionConfig_tags(rName, imageURI, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_container_group_definition" "test" {
  name                   = %[1]q
  container_group_type   = "GAME_SERVER"
  operating_system       = "AMAZON_LINUX_2023"
  total_memory_limit_mib = 1024
  total_vcpu_limit       = 1

  game_server_container_definition {
    container_name     = "game-server"
    image_uri          = %[2]q
    server_sdk_version = "5.2.0"

    port_configuration {
      container_port_ranges {
        from_port = 7777
        to_port   = 7778
        protocol  = "UDP"
      }
    }
  }

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, imageURI, tagKey, tagValue)
}
