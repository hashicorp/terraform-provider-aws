package lightsail_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflightsail "github.com/hashicorp/terraform-provider-aws/internal/service/lightsail"
)

func TestAccContainerService_basic(t *testing.T) {
	var cs lightsail.ContainerService
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

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
				Config: testAccContainerServiceConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "power", "nano"),
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
				Config: testAccContainerServiceConfigScale(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "scale", "2"),
				),
			},
		},
	})
}

func TestAccContainerService_disappears(t *testing.T) {
	var cs lightsail.ContainerService
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

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
				Config: testAccContainerServiceConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					acctest.CheckResourceDisappears(acctest.Provider, tflightsail.ResourceContainerService(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccContainerService_Name(t *testing.T) {
	var cs lightsail.ContainerService
	rName1 := sdkacctest.RandomWithPrefix("tf-acc-test")
	rName2 := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

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
				Config: testAccContainerServiceConfigBasic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				Config: testAccContainerServiceConfigBasic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func TestAccContainerService_DeploymentContainerBasic(t *testing.T) {
	var cs lightsail.ContainerService
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

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
				Config: testAccContainerServiceConfigDeploymentContainer1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.state", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.command.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.environment.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.port.#", "0"),
				),
			},
			{
				Config: testAccContainerServiceConfigDeploymentContainer2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.state", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "2"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.command.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.environment.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.port.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.1.container_name", "test2"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.1.image", "redis:latest"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.1.command.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.1.environment.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.1.port.#", "0"),
				),
			},
			{
				Config: testAccContainerServiceConfigDeploymentContainer3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.state", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "3"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test3"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.image", "nginx:latest"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.command.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.environment.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.port.#", "0"),
				),
			},
			{
				Config:      testAccContainerServiceConfigBasic(rName),
				ExpectError: regexp.MustCompile("a container service's deployment cannot be removed once a previous deployment was successful. You must specify a deployment now."),
			},
		},
	})
}

func TestAccContainerService_DeploymentContainerEnvironment(t *testing.T) {
	var cs lightsail.ContainerService
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

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
				Config: testAccContainerServiceConfigDeploymentContainerEnvironment1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.state", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.command.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.environment.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "deployment.0.container.0.environment.*",
						map[string]string{
							"key":   "A",
							"value": "a",
						}),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.port.#", "0"),
				),
			},
			{
				Config: testAccContainerServiceConfigDeploymentContainerEnvironment2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.state", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "2"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.command.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.environment.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "deployment.0.container.0.environment.*",
						map[string]string{
							"key":   "B",
							"value": "b",
						}),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.port.#", "0"),
				),
			},
			{
				Config: testAccContainerServiceConfigDeploymentContainerEnvironment3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.state", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "3"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.command.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.environment.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "deployment.0.container.0.environment.*",
						map[string]string{
							"key":   "A",
							"value": "a",
						}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "deployment.0.container.0.environment.*",
						map[string]string{
							"key":   "B",
							"value": "b",
						}),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.port.#", "0"),
				),
			},
			{
				Config: testAccContainerServiceConfigDeploymentContainerEnvironment4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.state", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "4"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.command.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.environment.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.port.#", "0"),
				),
			},
		},
	})
}

func TestAccContainerService_DeploymentContainerPort(t *testing.T) {
	var cs lightsail.ContainerService
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

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
				Config: testAccContainerServiceConfigDeploymentContainerPort1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.state", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.command.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.environment.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.port.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "deployment.0.container.0.port.*",
						map[string]string{
							"port_number": "80",
							"protocol":    "HTTP",
						}),
				),
			},
			{
				Config: testAccContainerServiceConfigDeploymentContainerPort2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.state", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "2"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.command.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.environment.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.port.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "deployment.0.container.0.port.*",
						map[string]string{
							"port_number": "90",
							"protocol":    "TCP",
						}),
				),
			},
			{
				Config: testAccContainerServiceConfigDeploymentContainerPort3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.state", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "3"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.command.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.environment.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.port.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "deployment.0.container.0.port.*",
						map[string]string{
							"port_number": "80",
							"protocol":    "HTTP",
						}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "deployment.0.container.0.port.*",
						map[string]string{
							"port_number": "90",
							"protocol":    "TCP",
						}),
				),
			},
			{
				Config: testAccContainerServiceConfigDeploymentContainerPort4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.state", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "4"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.command.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.environment.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.port.#", "0"),
				),
			},
		},
	})
}

func TestAccContainerService_DeploymentPublicEndpoint(t *testing.T) {
	var cs lightsail.ContainerService
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

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
				Config: testAccContainerServiceConfigDeploymentPublicEndpoint1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.state", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.port.0.port_number", "80"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.port.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.container_name", "test1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.container_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.healthy_threshold", "2"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.interval_seconds", "5"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.path", "/"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.success_codes", "200-499"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.timeout_seconds", "2"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.unhealthy_threshold", "2"),
				),
			},
			{
				Config: testAccContainerServiceConfigDeploymentPublicEndpoint2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.state", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "2"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test2"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.port.0.port_number", "80"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.port.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.container_name", "test2"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.container_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.interval_seconds", "6"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.path", "/."),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.success_codes", "200"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.timeout_seconds", "3"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.unhealthy_threshold", "3"),
				),
			},
			{
				Config: testAccContainerServiceConfigDeploymentPublicEndpoint3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.state", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "3"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test2"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.port.0.port_number", "80"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.port.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.container_name", "test2"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.container_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.healthy_threshold", "4"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.interval_seconds", "5"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.path", "/"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.success_codes", "200-499"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.timeout_seconds", "2"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.unhealthy_threshold", "2"),
				),
			},
			{
				Config: testAccContainerServiceConfigDeploymentPublicEndpoint4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.state", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "4"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test2"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.port.0.port_number", "80"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.port.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.container_name", "test2"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.container_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.healthy_threshold", "2"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.interval_seconds", "5"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.path", "/"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.success_codes", "200-499"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.timeout_seconds", "2"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.0.health_check.0.unhealthy_threshold", "2"),
				),
			},
			{
				Config: testAccContainerServiceConfigDeploymentPublicEndpoint5(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.state", "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "5"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test2"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.image", "amazon/amazon-lightsail:hello-world"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.port.0.port_number", "80"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.port.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.public_endpoint.#", "0"),
				),
			},
		},
	})
}

func TestAccContainerService_IsDisabled(t *testing.T) {
	var cs lightsail.ContainerService
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

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
				Config: testAccContainerServiceConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "is_disabled", "false"),
				),
			},
			{
				Config: testAccContainerServiceConfigIsDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "is_disabled", "true"),
				),
			},
			{
				Config: testAccContainerServiceConfigIsDisabledWithDeployment1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "is_disabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "power", "nano"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test1"),
				),
			},
			{
				Config: testAccContainerServiceConfigIsDisabledWithDeployment2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "is_disabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "power", "micro"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test1"),
				),
			},
			{
				Config: testAccContainerServiceConfigIsDisabledWithDeployment3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "is_disabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "power", "micro"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "2"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test2"),
				),
			},
			{
				Config: testAccContainerServiceConfigIsDisabledWithDeployment4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "is_disabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "power", "nano"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "3"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test1"),
				),
			},
			{
				Config: testAccContainerServiceConfigIsDisabledWithDeployment5(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "is_disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "power", "nano"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "3"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test1"),
				),
			},
		},
	})
}

func TestAccContainerService_Power(t *testing.T) {
	var cs lightsail.ContainerService
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

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
				Config: testAccContainerServiceConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "power", "nano"),
				),
			},
			{
				Config: testAccContainerServiceConfigPower(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "power", "micro"),
				),
			},
		},
	})
}

func TestAccContainerService_PublicDomainNames(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

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
				Config:      testAccContainerServiceConfigPublicDomainNames(rName),
				ExpectError: regexp.MustCompile(`do not exist`),
			},
		},
	})
}

func TestAccContainerService_Scale(t *testing.T) {
	var cs lightsail.ContainerService
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

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
				Config: testAccContainerServiceConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "scale", "1"),
				),
			},
			{
				Config: testAccContainerServiceConfigScale(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "scale", "2"),
				),
			},
		},
	})

}

func TestAccContainerService_Tags(t *testing.T) {
	var cs lightsail.ContainerService
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

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
				Config: testAccContainerServiceConfigTag1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccContainerServiceConfigTag2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "new_value1"),
				),
			},
		},
	})
}

func testAccCheckContainerServiceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn

	for _, r := range s.RootModule().Resources {
		if r.Type != "aws_lightsail_container_service" {
			continue
		}

		input := lightsail.GetContainerServicesInput{
			ServiceName: aws.String(r.Primary.ID),
		}

		_, err := conn.GetContainerServices(&input)
		if err == nil {
			return fmt.Errorf("container service still exists: %s", r.Primary.ID)
		}

		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == lightsail.ErrCodeNotFoundException {
				return nil
			}
		}
		return err
	}

	return nil
}

func testAccCheckContainerServiceExists(resourceName string, cs *lightsail.ContainerService) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not finding Lightsail Container Service (%s)", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Lightsail Container Service ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LightsailConn
		resp, err := conn.GetContainerServices(&lightsail.GetContainerServicesInput{
			ServiceName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}
		*cs = *resp.ContainerServices[0]

		return nil
	}
}

func testAccContainerServiceConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1
}
`, rName)
}

func testAccContainerServiceConfigDeploymentContainer1(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1

  deployment {
    container {
      container_name = "test1"
      image = "amazon/amazon-lightsail:hello-world"
    }
  }
}
`, rName)
}

func testAccContainerServiceConfigDeploymentContainer2(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1

  deployment {
    container {
      container_name = "test1"
      image = "amazon/amazon-lightsail:hello-world"
    }

    container {
      container_name = "test2"
      image = "redis:latest"
    }
  }
}
`, rName)
}

func testAccContainerServiceConfigDeploymentContainer3(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1

  deployment {
    container {
      container_name = "test3"
      image = "nginx:latest"
    }
  }
}
`, rName)
}

func testAccContainerServiceConfigDeploymentContainerEnvironment1(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1

  deployment {
    container {
      container_name = "test"
      image = "amazon/amazon-lightsail:hello-world"
      environment {
        key = "A"
        value = "a"
      }
    }
  }
}
`, rName)
}

func testAccContainerServiceConfigDeploymentContainerEnvironment2(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1

  deployment {
    container {
      container_name = "test"
      image = "amazon/amazon-lightsail:hello-world"
      environment {
        key = "B"
        value = "b"
      }
    }
  }
}
`, rName)
}

func testAccContainerServiceConfigDeploymentContainerEnvironment3(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1

  deployment {
    container {
      container_name = "test"
      image = "amazon/amazon-lightsail:hello-world"
      environment {
        key = "A"
        value = "a"
      }
      environment {
        key = "B"
        value = "b"
      }
    }
  }
}
`, rName)
}

func testAccContainerServiceConfigDeploymentContainerEnvironment4(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1

  deployment {
    container {
      container_name = "test"
      image = "amazon/amazon-lightsail:hello-world"
    }
  }
}
`, rName)
}

func testAccContainerServiceConfigDeploymentContainerPort1(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1

  deployment {
    container {
      container_name = "test"
      image = "amazon/amazon-lightsail:hello-world"
      port {
        port_number = 80
        protocol = "HTTP"
      }
    }
  }
}
`, rName)
}

func testAccContainerServiceConfigDeploymentContainerPort2(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1

  deployment {
    container {
      container_name = "test"
      image = "amazon/amazon-lightsail:hello-world"
      port {
        port_number = 90
        protocol = "TCP"
      }
    }
  }
}
`, rName)
}

func testAccContainerServiceConfigDeploymentContainerPort3(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1

  deployment {
    container {
      container_name = "test"
      image = "amazon/amazon-lightsail:hello-world"
      port {
        port_number = 80
        protocol = "HTTP"
      }
      port {
        port_number = 90
        protocol = "TCP"
      }
    }
  }
}
`, rName)
}

func testAccContainerServiceConfigDeploymentContainerPort4(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1

  deployment {
    container {
      container_name = "test"
      image = "amazon/amazon-lightsail:hello-world"
    }
  }
}
`, rName)
}

func testAccContainerServiceConfigDeploymentPublicEndpoint1(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1

  deployment {
    container {
      container_name = "test1"
      image = "amazon/amazon-lightsail:hello-world"
      port {
        port_number = 80
        protocol = "HTTP"
      }
    }

    public_endpoint {
      container_name = "test1"
      container_port = 80
      health_check {}
    }
  }
}
`, rName)
}

func testAccContainerServiceConfigDeploymentPublicEndpoint2(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1

  deployment {
    container {
      container_name = "test2"
      image = "amazon/amazon-lightsail:hello-world"
      port {
        port_number = 80
        protocol = "HTTP"
      }
    }

    public_endpoint {
      container_name = "test2"
      container_port = 80
      health_check {
        healthy_threshold = 3
        interval_seconds = 6
        path = "/."
        success_codes = "200"
        timeout_seconds = 3
        unhealthy_threshold = 3
      }
    }
  }
}
`, rName)
}

func testAccContainerServiceConfigDeploymentPublicEndpoint3(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1

  deployment {
    container {
      container_name = "test2"
      image = "amazon/amazon-lightsail:hello-world"
      port {
        port_number = 80
        protocol = "HTTP"
      }
    }

    public_endpoint {
      container_name = "test2"
      container_port = 80
      health_check {
        healthy_threshold = 4
      }
    }
  }
}
`, rName)
}

func testAccContainerServiceConfigDeploymentPublicEndpoint4(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1

  deployment {
    container {
      container_name = "test2"
      image = "amazon/amazon-lightsail:hello-world"
      port {
        port_number = 80
        protocol = "HTTP"
      }
    }

    public_endpoint {
      container_name = "test2"
      container_port = 80
      health_check {}
    }
  }
}
`, rName)
}

func testAccContainerServiceConfigDeploymentPublicEndpoint5(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1

  deployment {
    container {
      container_name = "test2"
      image = "amazon/amazon-lightsail:hello-world"
      port {
        port_number = 80
        protocol = "HTTP"
      }
    }
  }
}
`, rName)
}

func testAccContainerServiceConfigIsDisabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1
  is_disabled = true
}
`, rName)
}

func testAccContainerServiceConfigIsDisabledWithDeployment1(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1
  is_disabled = true

  deployment {
    container {
      container_name = "test1"
      image = "amazon/amazon-lightsail:hello-world"
    }
  }
}
`, rName)
}

func testAccContainerServiceConfigIsDisabledWithDeployment2(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "micro"
  scale = 1
  is_disabled = true

  deployment {
    container {
      container_name = "test1"
      image = "amazon/amazon-lightsail:hello-world"
    }
  }
}
`, rName)
}

func testAccContainerServiceConfigIsDisabledWithDeployment3(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "micro"
  scale = 1
  is_disabled = true

  deployment {
    container {
      container_name = "test2"
      image = "amazon/amazon-lightsail:hello-world"
    }
  }
}
`, rName)
}

func testAccContainerServiceConfigIsDisabledWithDeployment4(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1
  is_disabled = true

  deployment {
    container {
      container_name = "test1"
      image = "amazon/amazon-lightsail:hello-world"
    }
  }
}
`, rName)
}

func testAccContainerServiceConfigIsDisabledWithDeployment5(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1

  deployment {
    container {
      container_name = "test1"
      image = "amazon/amazon-lightsail:hello-world"
    }
  }
}
`, rName)
}

func testAccContainerServiceConfigPower(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "micro"
  scale = 1
}
`, rName)
}

func testAccContainerServiceConfigPublicDomainNames(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
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

func testAccContainerServiceConfigScale(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 2
}
`, rName)
}

func testAccContainerServiceConfigTag1(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1
  
  tags = {
    key1 = "value1"
	key2 = "value2"
  }
}
`, rName)
}

func testAccContainerServiceConfigTag2(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1

  tags = {
    key1 = "new_value1"
  }
}
`, rName)
}
