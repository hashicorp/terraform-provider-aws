package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_lightsail_container_service", &resource.Sweeper{
		Name: "aws_lightsail_container_service",
		F:    testSweepLightsailContainerService,
	})
}

func testSweepLightsailContainerService(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).lightsailconn
	resp, err := conn.GetContainerServices(&lightsail.GetContainerServicesInput{})

	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Lightsail Container Service skipping sweep for region (%s): %s", region, err)
			return nil
		}
		return fmt.Errorf("error getting Lightsail Container Services: %w", err)
	}

	var sweeperErrs *multierror.Error
	for _, containerService := range resp.ContainerServices {
		name := containerService.ContainerServiceName

		_, err := conn.DeleteContainerService(&lightsail.DeleteContainerServiceInput{
			ServiceName: name,
		})

		if err != nil {
			sweeperErr := fmt.Errorf("error deleting Lightsail Container Service (%s): %w", aws.StringValue(name), err)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
		}
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSLightsailContainerService_basic(t *testing.T) {
	var cs lightsail.ContainerService
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:        testAccErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSLightsailContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailContainerServiceConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
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
		},
	})

}

func TestAccAWSLightsailContainerService_disappears(t *testing.T) {
	var cs lightsail.ContainerService
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:        testAccErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSLightsailContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailContainerServiceConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsLightsailContainerService(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLightsailContainerService_Name(t *testing.T) {
	var cs lightsail.ContainerService
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:        testAccErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSLightsailContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailContainerServiceConfigBasic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				Config: testAccAWSLightsailContainerServiceConfigBasic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func TestAccAWSLightsailContainerService_DeploymentContainerBasic(t *testing.T) {
	var cs lightsail.ContainerService
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:        testAccErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSLightsailContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailContainerServiceConfigDeploymentContainer1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
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
				Config: testAccAWSLightsailContainerServiceConfigDeploymentContainer2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
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
				Config: testAccAWSLightsailContainerServiceConfigDeploymentContainer3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
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
				Config:      testAccAWSLightsailContainerServiceConfigBasic(rName),
				ExpectError: regexp.MustCompile("a container service's deployment cannot be removed once a previous deployment was successful. You must specify a deployment now."),
			},
		},
	})
}

func TestAccAWSLightsailContainerService_DeploymentContainerEnvironment(t *testing.T) {
	var cs lightsail.ContainerService
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:        testAccErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSLightsailContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailContainerServiceConfigDeploymentContainerEnvironment1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
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
				Config: testAccAWSLightsailContainerServiceConfigDeploymentContainerEnvironment2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
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
				Config: testAccAWSLightsailContainerServiceConfigDeploymentContainerEnvironment3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
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
				Config: testAccAWSLightsailContainerServiceConfigDeploymentContainerEnvironment4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
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

func TestAccAWSLightsailContainerService_DeploymentContainerPort(t *testing.T) {
	var cs lightsail.ContainerService
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:        testAccErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSLightsailContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailContainerServiceConfigDeploymentContainerPort1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
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
				Config: testAccAWSLightsailContainerServiceConfigDeploymentContainerPort2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
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
				Config: testAccAWSLightsailContainerServiceConfigDeploymentContainerPort3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
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
				Config: testAccAWSLightsailContainerServiceConfigDeploymentContainerPort4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
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

func TestAccAWSLightsailContainerService_DeploymentPublicEndpoint(t *testing.T) {
	var cs lightsail.ContainerService
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:        testAccErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSLightsailContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailContainerServiceConfigDeploymentPublicEndpoint1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
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
				Config: testAccAWSLightsailContainerServiceConfigDeploymentPublicEndpoint2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
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
				Config: testAccAWSLightsailContainerServiceConfigDeploymentPublicEndpoint3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
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
				Config: testAccAWSLightsailContainerServiceConfigDeploymentPublicEndpoint4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
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
				Config: testAccAWSLightsailContainerServiceConfigDeploymentPublicEndpoint5(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
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

func TestAccAWSLightsailContainerService_IsDisabled(t *testing.T) {
	var cs lightsail.ContainerService
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:        testAccErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSLightsailContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailContainerServiceConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "is_disabled", "false"),
				),
			},
			{
				Config: testAccAWSLightsailContainerServiceConfigIsDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "is_disabled", "true"),
				),
			},
			{
				Config: testAccAWSLightsailContainerServiceConfigIsDisabledWithDeployment1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "is_disabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "power", "nano"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test1"),
				),
			},
			{
				Config: testAccAWSLightsailContainerServiceConfigIsDisabledWithDeployment2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "is_disabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "power", "micro"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test1"),
				),
			},
			{
				Config: testAccAWSLightsailContainerServiceConfigIsDisabledWithDeployment3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "is_disabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "power", "micro"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "2"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test2"),
				),
			},
			{
				Config: testAccAWSLightsailContainerServiceConfigIsDisabledWithDeployment4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "is_disabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "power", "nano"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "3"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test1"),
				),
			},
			{
				Config: testAccAWSLightsailContainerServiceConfigIsDisabledWithDeployment5(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "is_disabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "power", "nano"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.version", "3"),
					resource.TestCheckResourceAttr(resourceName, "deployment.0.container.0.container_name", "test1"),
				),
			},
		},
	})
}

func TestAccAWSLightsailContainerService_Power(t *testing.T) {
	var cs lightsail.ContainerService
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:        testAccErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSLightsailContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailContainerServiceConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "power", "nano"),
				),
			},
			{
				Config: testAccAWSLightsailContainerServiceConfigPower(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "power", "micro"),
				),
			},
		},
	})
}

func TestAccAWSLightsailContainerService_PublicDomainNames(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:        testAccErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSLightsailContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSLightsailContainerServiceConfigPublicDomainNames(rName),
				ExpectError: regexp.MustCompile("NotFoundException: The specified certificate does not exist"),
			},
		},
	})
}

func TestAccAWSLightsailContainerService_Scale(t *testing.T) {
	var cs lightsail.ContainerService
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:        testAccErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSLightsailContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailContainerServiceConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "scale", "1"),
				),
			},
			{
				Config: testAccAWSLightsailContainerServiceConfigScale(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "scale", "2"),
				),
			},
		},
	})

}

func TestAccAWSLightsailContainerService_Tags(t *testing.T) {
	var cs lightsail.ContainerService
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lightsail_container_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lightsail.EndpointsID, t)
			testAccPreCheckAWSLightsail(t)
		},
		ErrorCheck:        testAccErrorCheck(t, lightsail.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSLightsailContainerServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLightsailContainerServiceConfigTag1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSLightsailContainerServiceConfigTag2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLightsailContainerServiceExists(resourceName, &cs),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "new_value1"),
				),
			},
		},
	})
}

func testAccCheckAWSLightsailContainerServiceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lightsailconn

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

func testAccCheckAWSLightsailContainerServiceExists(resourceName string, cs *lightsail.ContainerService) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not finding Lightsail Container Service (%s)", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Lightsail Container Service ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).lightsailconn
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

func testAccAWSLightsailContainerServiceConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1
}
`, rName)
}

func testAccAWSLightsailContainerServiceConfigDeploymentContainer1(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigDeploymentContainer2(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigDeploymentContainer3(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigDeploymentContainerEnvironment1(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigDeploymentContainerEnvironment2(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigDeploymentContainerEnvironment3(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigDeploymentContainerEnvironment4(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigDeploymentContainerPort1(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigDeploymentContainerPort2(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigDeploymentContainerPort3(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigDeploymentContainerPort4(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigDeploymentPublicEndpoint1(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigDeploymentPublicEndpoint2(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigDeploymentPublicEndpoint3(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigDeploymentPublicEndpoint4(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigDeploymentPublicEndpoint5(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigIsDisabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 1
  is_disabled = true
}
`, rName)
}

func testAccAWSLightsailContainerServiceConfigIsDisabledWithDeployment1(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigIsDisabledWithDeployment2(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigIsDisabledWithDeployment3(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigIsDisabledWithDeployment4(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigIsDisabledWithDeployment5(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigPower(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "micro"
  scale = 1
}
`, rName)
}

func testAccAWSLightsailContainerServiceConfigPublicDomainNames(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigScale(rName string) string {
	return fmt.Sprintf(`
resource "aws_lightsail_container_service" "test" {
  name = %q
  power = "nano"
  scale = 2
}
`, rName)
}

func testAccAWSLightsailContainerServiceConfigTag1(rName string) string {
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

func testAccAWSLightsailContainerServiceConfigTag2(rName string) string {
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
