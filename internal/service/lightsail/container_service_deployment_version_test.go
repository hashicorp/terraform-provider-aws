// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	helloWorldImage = "amazon/amazon-lightsail:hello-world"
	redisImage      = "redis:latest"
)

func TestContainerServiceDeploymentVersionParseResourceID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName            string
		Input               string
		ExpectedServiceName string
		ExpectedVersion     int
		Error               bool
	}{
		{
			TestName:            "empty",
			Input:               "",
			ExpectedServiceName: "",
			ExpectedVersion:     0,
			Error:               true,
		},
		{
			TestName:            "Invalid ID",
			Input:               "abcdefg12345678/",
			ExpectedServiceName: "",
			ExpectedVersion:     0,
			Error:               true,
		},
		{
			TestName:            "Invalid ID separator",
			Input:               "abcdefg12345678:1",
			ExpectedServiceName: "",
			ExpectedVersion:     0,
			Error:               true,
		},
		{
			TestName:            "Invalid ID with more than 1 separator",
			Input:               "abcdefg12345678/qwerty09876/1",
			ExpectedServiceName: "",
			ExpectedVersion:     0,
			Error:               true,
		},
		{
			TestName:            "Valid ID",
			Input:               "abcdefg12345678/1",
			ExpectedServiceName: "abcdefg12345678",
			ExpectedVersion:     1,
			Error:               false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			gotServiceName, gotVersion, err := tflightsail.ContainerServiceDeploymentVersionParseResourceID(testCase.Input)

			if err != nil && !testCase.Error {
				t.Errorf("got error (%s), expected no error", err)
			}

			if err == nil && testCase.Error {
				t.Errorf("got (ServiceName: %s, Version: %d) and no error, expected error", gotServiceName, gotVersion)
			}

			if gotServiceName != testCase.ExpectedServiceName {
				t.Errorf("got %s, expected %s", gotServiceName, testCase.ExpectedServiceName)
			}

			if gotVersion != testCase.ExpectedVersion {
				t.Errorf("got %d, expected %d", gotVersion, testCase.ExpectedVersion)
			}
		})
	}
}

func TestAccLightsailContainerServiceDeploymentVersion_container_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	containerName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_container_service_deployment_version.test"

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
				Config: testAccContainerServiceDeploymentVersionConfig_Container_basic(rName, containerName, helloWorldImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.ContainerServiceDeploymentStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "container.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "container.0.container_name", containerName),
					resource.TestCheckResourceAttr(resourceName, "container.0.image", helloWorldImage),
					resource.TestCheckResourceAttr(resourceName, "container.0.command.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "container.0.environment.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.%", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceName, "aws_lightsail_container_service.test", names.AttrName),
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

func TestAccLightsailContainerServiceDeploymentVersion_container_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	containerName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	containerName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_container_service_deployment_version.test"

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
				Config: testAccContainerServiceDeploymentVersionConfig_Container_multiple(rName, containerName1, helloWorldImage, containerName2, redisImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.ContainerServiceDeploymentStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "container.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "container.0.container_name", containerName1),
					resource.TestCheckResourceAttr(resourceName, "container.0.image", helloWorldImage),
					resource.TestCheckResourceAttr(resourceName, "container.1.container_name", containerName2),
					resource.TestCheckResourceAttr(resourceName, "container.1.image", redisImage),
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

func TestAccLightsailContainerServiceDeploymentVersion_container_environment(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	containerName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_container_service_deployment_version.test"

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
				Config: testAccContainerServiceDeploymentVersionConfig_Container_environment1(rName, containerName, "A", "a"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.ContainerServiceDeploymentStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "container.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "container.0.environment.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "container.0.environment.A", "a"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_environment1(rName, containerName, "B", "b"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.ContainerServiceDeploymentStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "container.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "container.0.environment.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "container.0.environment.B", "b"),
				),
			},
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_environment2(rName, containerName, "A", "a", "B", "b"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.ContainerServiceDeploymentStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "container.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "container.0.environment.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "container.0.environment.A", "a"),
					resource.TestCheckResourceAttr(resourceName, "container.0.environment.B", "b"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_basic(rName, containerName, helloWorldImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.ContainerServiceDeploymentStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "container.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "container.0.environment.%", acctest.Ct0),
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

func TestAccLightsailContainerServiceDeploymentVersion_container_ports(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	containerName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_container_service_deployment_version.test"

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
				Config: testAccContainerServiceDeploymentVersionConfig_Container_ports1(rName, containerName, "80", string(types.ContainerServiceProtocolHttp)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.ContainerServiceDeploymentStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "container.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.80", string(types.ContainerServiceProtocolHttp)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_ports1(rName, containerName, "90", string(types.ContainerServiceProtocolTcp)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.ContainerServiceDeploymentStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "container.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.90", string(types.ContainerServiceProtocolTcp)),
				),
			},
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_ports2(rName, containerName, "80", string(types.ContainerServiceProtocolHttp), "90", string(types.ContainerServiceProtocolTcp)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.ContainerServiceDeploymentStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "container.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.80", string(types.ContainerServiceProtocolHttp)),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.90", string(types.ContainerServiceProtocolTcp)),
				),
			},
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_basic(rName, containerName, helloWorldImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.ContainerServiceDeploymentStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "container.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.%", acctest.Ct0),
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

func TestAccLightsailContainerServiceDeploymentVersion_container_publicEndpoint(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	containerName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	containerName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_lightsail_container_service_deployment_version.test"

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
				Config: testAccContainerServiceDeploymentVersionConfig_Container_publicEndpoint(rName, containerName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.ContainerServiceDeploymentStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "container.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.80", string(types.ContainerServiceProtocolHttp)),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.container_name", containerName1),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.container_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.healthy_threshold", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.interval_seconds", "5"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.path", "/"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.success_codes", "200-499"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.timeout_seconds", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.unhealthy_threshold", acctest.Ct2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_publicEndpointCompleteHealthCheck(rName, containerName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.ContainerServiceDeploymentStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "container.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.container_name", containerName2),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.container_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.healthy_threshold", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.interval_seconds", "6"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.path", "/."),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.success_codes", "200"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.timeout_seconds", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.unhealthy_threshold", acctest.Ct3),
				),
			},
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_publicEndpointMinimalHealthCheck(rName, containerName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.ContainerServiceDeploymentStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "container.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.container_name", containerName2),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.container_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.healthy_threshold", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.interval_seconds", "5"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.path", "/"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.success_codes", "200-499"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.timeout_seconds", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.unhealthy_threshold", acctest.Ct2),
				),
			},
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_publicEndpoint(rName, containerName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.ContainerServiceDeploymentStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "container.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.container_name", containerName2),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.container_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.healthy_threshold", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.interval_seconds", "5"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.path", "/"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.success_codes", "200-499"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.timeout_seconds", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.unhealthy_threshold", acctest.Ct2),
				),
			},
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_basic(rName, containerName1, helloWorldImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.ContainerServiceDeploymentStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "5"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccLightsailContainerServiceDeploymentVersion_Container_enableService(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	containerName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_container_service_deployment_version.test"

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
				Config:      testAccContainerServiceDeploymentVersionConfig_Container_withDisabledService(rName, containerName, true),
				ExpectError: regexache.MustCompile(`disabled and cannot be deployed to`),
			},
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_withDisabledService(rName, containerName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(types.ContainerServiceDeploymentStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "container.0.container_name", containerName),
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

func testAccCheckContainerServiceDeploymentVersionExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Lightsail Container Service Deployment Version ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailClient(ctx)

		serviceName, version, err := tflightsail.ContainerServiceDeploymentVersionParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = tflightsail.FindContainerServiceDeploymentByVersion(ctx, conn, serviceName, version)

		return err
	}
}

func testAccContainerServiceDeploymentVersionBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name  = %[1]q
  power = "nano"
  scale = 1
}
`, rName)
}

func testAccContainerServiceDeploymentVersionConfig_Container_basic(rName, containerName, image string) string {
	return acctest.ConfigCompose(
		testAccContainerServiceDeploymentVersionBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_lightsail_container_service_deployment_version" "test" {
  container {
    container_name = %[1]q
    image          = %[2]q
  }

  service_name = aws_lightsail_container_service.test.name
}
`, containerName, image))
}

func testAccContainerServiceDeploymentVersionConfig_Container_multiple(rName, containerName1, image1, containerName2, image2 string) string {
	return acctest.ConfigCompose(
		testAccContainerServiceDeploymentVersionBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_lightsail_container_service_deployment_version" "test" {
  container {
    container_name = %[1]q
    image          = %[2]q
  }

  container {
    container_name = %[3]q
    image          = %[4]q
  }

  service_name = aws_lightsail_container_service.test.name
}
`, containerName1, image1, containerName2, image2))
}

func testAccContainerServiceDeploymentVersionConfig_Container_environment1(rName, containerName, envKey, envValue string) string {
	return acctest.ConfigCompose(
		testAccContainerServiceDeploymentVersionBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_lightsail_container_service_deployment_version" "test" {
  container {
    container_name = %[1]q
    image          = "amazon/amazon-lightsail:hello-world"
    environment = {
      %[2]q = %[3]q
    }
  }

  service_name = aws_lightsail_container_service.test.name
}
`, containerName, envKey, envValue))
}

func testAccContainerServiceDeploymentVersionConfig_Container_environment2(rName, containerName, envKey1, envValue1, envKey2, envValue2 string) string {
	return acctest.ConfigCompose(
		testAccContainerServiceDeploymentVersionBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_lightsail_container_service_deployment_version" "test" {
  container {
    container_name = %[1]q
    image          = "amazon/amazon-lightsail:hello-world"
    environment = {
      %[2]q = %[3]q
      %[4]q = %[5]q
    }
  }

  service_name = aws_lightsail_container_service.test.name
}
`, containerName, envKey1, envValue1, envKey2, envValue2))
}

func testAccContainerServiceDeploymentVersionConfig_Container_ports1(rName, containerName, portKey, portValue string) string {
	return acctest.ConfigCompose(
		testAccContainerServiceDeploymentVersionBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_lightsail_container_service_deployment_version" "test" {
  container {
    container_name = %[1]q
    image          = "amazon/amazon-lightsail:hello-world"
    ports = {
      %[2]q = %[3]q
    }
  }

  service_name = aws_lightsail_container_service.test.name
}
`, containerName, portKey, portValue))
}

func testAccContainerServiceDeploymentVersionConfig_Container_ports2(rName, containerName, portKey1, portValue1, portKey2, portValue2 string) string {
	return acctest.ConfigCompose(
		testAccContainerServiceDeploymentVersionBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_lightsail_container_service_deployment_version" "test" {
  container {
    container_name = %[1]q
    image          = "amazon/amazon-lightsail:hello-world"
    ports = {
      %[2]q = %[3]q
      %[4]q = %[5]q
    }
  }

  service_name = aws_lightsail_container_service.test.name
}
`, containerName, portKey1, portValue1, portKey2, portValue2))
}

func testAccContainerServiceDeploymentVersionConfig_Container_publicEndpoint(rName, containerName string) string {
	return acctest.ConfigCompose(
		testAccContainerServiceDeploymentVersionBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_lightsail_container_service_deployment_version" "test" {
  container {
    container_name = %[1]q
    image          = "amazon/amazon-lightsail:hello-world"
    ports = {
      80 = "HTTP"
    }
  }

  public_endpoint {
    container_name = %[1]q
    container_port = 80
    health_check {}
  }

  service_name = aws_lightsail_container_service.test.name
}
`, containerName))
}

func testAccContainerServiceDeploymentVersionConfig_Container_publicEndpointCompleteHealthCheck(rName, containerName string) string {
	return acctest.ConfigCompose(
		testAccContainerServiceDeploymentVersionBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_lightsail_container_service_deployment_version" "test" {
  container {
    container_name = %[1]q
    image          = "amazon/amazon-lightsail:hello-world"
    ports = {
      80 = "HTTP"
    }
  }

  public_endpoint {
    container_name = %[1]q
    container_port = 80
    health_check {
      healthy_threshold   = 3
      interval_seconds    = 6
      path                = "/."
      success_codes       = "200"
      timeout_seconds     = 3
      unhealthy_threshold = 3
    }
  }

  service_name = aws_lightsail_container_service.test.name
}
`, containerName))
}

func testAccContainerServiceDeploymentVersionConfig_Container_publicEndpointMinimalHealthCheck(rName, containerName string) string {
	return acctest.ConfigCompose(
		testAccContainerServiceDeploymentVersionBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_lightsail_container_service_deployment_version" "test" {
  container {
    container_name = %[1]q
    image          = "amazon/amazon-lightsail:hello-world"
    ports = {
      80 = "HTTP"
    }
  }
  public_endpoint {
    container_name = %[1]q
    container_port = 80
    health_check {
      healthy_threshold = 4
    }
  }

  service_name = aws_lightsail_container_service.test.name
}
`, containerName))
}

func testAccContainerServiceDeploymentVersionConfig_Container_withDisabledService(rName, containerName string, isDisabled bool) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name        = %[1]q
  power       = "nano"
  scale       = 1
  is_disabled = %[2]t
}

resource "aws_lightsail_container_service_deployment_version" "test" {
  container {
    container_name = %[3]q
    image          = "amazon/amazon-lightsail:hello-world"
  }

  service_name = aws_lightsail_container_service.test.name
}
`, rName, isDisabled, containerName)
}
