package lightsail_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/lightsail"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
)

const (
	helloWorldImage = "amazon/amazon-lightsail:hello-world"
	redisImage      = "redis:latest"
)

func TestContainerServiceDeploymentVersionParseResourceID(t *testing.T) {
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
		t.Run(testCase.TestName, func(t *testing.T) {
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

func TestAccLightsailContainerServiceDeploymentVersion_Container_Basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	containerName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_container_service_deployment_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_basic(rName, containerName, helloWorldImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "state", lightsail.ContainerServiceDeploymentStateActive),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
					resource.TestCheckResourceAttr(resourceName, "container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container.0.container_name", containerName),
					resource.TestCheckResourceAttr(resourceName, "container.0.image", helloWorldImage),
					resource.TestCheckResourceAttr(resourceName, "container.0.command.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "container.0.environment.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_name", "aws_lightsail_container_service.test", "name"),
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

func TestAccLightsailContainerServiceDeploymentVersion_Container_Multiple(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	containerName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	containerName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_container_service_deployment_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_multiple(rName, containerName1, helloWorldImage, containerName2, redisImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "state", lightsail.ContainerServiceDeploymentStateActive),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
					resource.TestCheckResourceAttr(resourceName, "container.#", "2"),
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

func TestAccLightsailContainerServiceDeploymentVersion_Container_Environment(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	containerName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_container_service_deployment_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_environment1(rName, containerName, "A", "a"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "state", lightsail.ContainerServiceDeploymentStateActive),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
					resource.TestCheckResourceAttr(resourceName, "container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container.0.environment.%", "1"),
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
					testAccCheckContainerServiceDeploymentVersionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "state", lightsail.ContainerServiceDeploymentStateActive),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
					resource.TestCheckResourceAttr(resourceName, "container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container.0.environment.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "container.0.environment.B", "b"),
				),
			},
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_environment2(rName, containerName, "A", "a", "B", "b"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "state", lightsail.ContainerServiceDeploymentStateActive),
					resource.TestCheckResourceAttr(resourceName, "version", "3"),
					resource.TestCheckResourceAttr(resourceName, "container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container.0.environment.%", "2"),
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
					testAccCheckContainerServiceDeploymentVersionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "state", lightsail.ContainerServiceDeploymentStateActive),
					resource.TestCheckResourceAttr(resourceName, "version", "4"),
					resource.TestCheckResourceAttr(resourceName, "container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container.0.environment.%", "0"),
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

func TestAccLightsailContainerServiceDeploymentVersion_Container_Ports(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	containerName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_container_service_deployment_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_ports1(rName, containerName, "80", lightsail.ContainerServiceProtocolHttp),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "state", lightsail.ContainerServiceDeploymentStateActive),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
					resource.TestCheckResourceAttr(resourceName, "container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.80", lightsail.ContainerServiceProtocolHttp),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_ports1(rName, containerName, "90", lightsail.ContainerServiceProtocolTcp),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "state", lightsail.ContainerServiceDeploymentStateActive),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
					resource.TestCheckResourceAttr(resourceName, "container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.90", lightsail.ContainerServiceProtocolTcp),
				),
			},
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_ports2(rName, containerName, "80", lightsail.ContainerServiceProtocolHttp, "90", lightsail.ContainerServiceProtocolTcp),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "state", lightsail.ContainerServiceDeploymentStateActive),
					resource.TestCheckResourceAttr(resourceName, "version", "3"),
					resource.TestCheckResourceAttr(resourceName, "container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.80", lightsail.ContainerServiceProtocolHttp),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.90", lightsail.ContainerServiceProtocolTcp),
				),
			},
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_basic(rName, containerName, helloWorldImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "state", lightsail.ContainerServiceDeploymentStateActive),
					resource.TestCheckResourceAttr(resourceName, "version", "4"),
					resource.TestCheckResourceAttr(resourceName, "container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.%", "0"),
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

func TestAccLightsailContainerServiceDeploymentVersion_Container_PublicEndpoint(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	containerName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	containerName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_lightsail_container_service_deployment_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_publicEndpoint(rName, containerName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "state", lightsail.ContainerServiceDeploymentStateActive),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
					resource.TestCheckResourceAttr(resourceName, "container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "container.0.ports.80", lightsail.ContainerServiceProtocolHttp),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.container_name", containerName1),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.container_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.healthy_threshold", "2"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.interval_seconds", "5"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.path", "/"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.success_codes", "200-499"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.timeout_seconds", "2"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.unhealthy_threshold", "2"),
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
					testAccCheckContainerServiceDeploymentVersionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "state", lightsail.ContainerServiceDeploymentStateActive),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
					resource.TestCheckResourceAttr(resourceName, "container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.container_name", containerName2),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.container_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.interval_seconds", "6"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.path", "/."),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.success_codes", "200"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.timeout_seconds", "3"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.unhealthy_threshold", "3"),
				),
			},
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_publicEndpointMinimalHealthCheck(rName, containerName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "state", lightsail.ContainerServiceDeploymentStateActive),
					resource.TestCheckResourceAttr(resourceName, "version", "3"),
					resource.TestCheckResourceAttr(resourceName, "container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.container_name", containerName2),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.container_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.interval_seconds", "5"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.path", "/"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.success_codes", "200-499"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.timeout_seconds", "2"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.unhealthy_threshold", "2"),
				),
			},
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_publicEndpoint(rName, containerName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "state", lightsail.ContainerServiceDeploymentStateActive),
					resource.TestCheckResourceAttr(resourceName, "version", "4"),
					resource.TestCheckResourceAttr(resourceName, "container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.container_name", containerName2),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.container_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.healthy_threshold", "2"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.interval_seconds", "5"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.path", "/"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.success_codes", "200-499"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.timeout_seconds", "2"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.0.health_check.0.unhealthy_threshold", "2"),
				),
			},
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_basic(rName, containerName1, helloWorldImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "state", lightsail.ContainerServiceDeploymentStateActive),
					resource.TestCheckResourceAttr(resourceName, "version", "5"),
					resource.TestCheckResourceAttr(resourceName, "public_endpoint.#", "0"),
				),
			},
		},
	})
}

func TestAccLightsailContainerServiceDeploymentVersion_Container_EnableService(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	containerName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lightsail_container_service_deployment_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lightsail.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccContainerServiceDeploymentVersionConfig_Container_withDisabledService(rName, containerName, true),
				ExpectError: regexp.MustCompile(`disabled and cannot be deployed to`),
			},
			{
				Config: testAccContainerServiceDeploymentVersionConfig_Container_withDisabledService(rName, containerName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceDeploymentVersionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "state", lightsail.ContainerServiceDeploymentStateActive),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
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

func testAccCheckContainerServiceDeploymentVersionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Lightsail Container Service Deployment Version ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn

		serviceName, version, err := tflightsail.ContainerServiceDeploymentVersionParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = tflightsail.FindContainerServiceDeploymentByVersion(context.TODO(), conn, serviceName, version)
		if err != nil {
			return err
		}

		return nil
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
