package emrserverless_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/emrserverless"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfemrserverless "github.com/hashicorp/terraform-provider-aws/internal/service/emrserverless"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEMRServerlessApplication_basic(t *testing.T) {
	var application emrserverless.Application
	resourceName := "aws_emrserverless_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emrserverless.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &application),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "emr-serverless", regexp.MustCompile(`/applications/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", "hive"),
					resource.TestCheckResourceAttr(resourceName, "release_label", "emr-6.6.0"),
					resource.TestCheckResourceAttr(resourceName, "auto_start_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_start_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_stop_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_stop_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "auto_stop_configuration.0.idle_timeout_minutes", "15"),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.#", "0"),
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

func TestAccEMRServerlessApplication_initialCapacity(t *testing.T) {
	var application emrserverless.Application
	resourceName := "aws_emrserverless_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emrserverless.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_initialCapacity(rName, "2 vCPU"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_type", "HiveDriver"),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_config.0.worker_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_config.0.worker_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_config.0.worker_configuration.0.cpu", "2 vCPU"),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_config.0.worker_configuration.0.memory", "10 GB"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_initialCapacity(rName, "4 vCPU"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_type", "HiveDriver"),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_config.0.worker_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_config.0.worker_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_config.0.worker_configuration.0.cpu", "4 vCPU"),
					resource.TestCheckResourceAttr(resourceName, "initial_capacity.0.initial_capacity_config.0.worker_configuration.0.memory", "10 GB"),
				),
			},
		},
	})
}

func TestAccEMRServerlessApplication_maxCapacity(t *testing.T) {
	var application emrserverless.Application
	resourceName := "aws_emrserverless_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emrserverless.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_maxCapacity(rName, "2 vCPU"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "maximum_capacity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "maximum_capacity.0.cpu", "2 vCPU"),
					resource.TestCheckResourceAttr(resourceName, "maximum_capacity.0.memory", "10 GB"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationConfig_maxCapacity(rName, "4 vCPU"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "maximum_capacity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "maximum_capacity.0.cpu", "4 vCPU"),
					resource.TestCheckResourceAttr(resourceName, "maximum_capacity.0.memory", "10 GB")),
			},
		},
	})
}

func TestAccEMRServerlessApplication_network(t *testing.T) {
	var application emrserverless.Application
	resourceName := "aws_emrserverless_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emrserverless.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_network(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.subnet_ids.#", "2"),
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

func TestAccEMRServerlessApplication_disappears(t *testing.T) {
	var application emrserverless.Application
	resourceName := "aws_emrserverless_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emrserverless.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &application),
					acctest.CheckResourceDisappears(acctest.Provider, tfemrserverless.ResourceApplication(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfemrserverless.ResourceApplication(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEMRServerlessApplication_tags(t *testing.T) {
	var application emrserverless.Application
	resourceName := "aws_emrserverless_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emrserverless.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &application),
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
				Config: testAccApplicationConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccApplicationConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckApplicationExists(resourceName string, application *emrserverless.Application) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRServerlessConn

		output, err := tfemrserverless.FindApplicationByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("EMR Serverless Application (%s) not found", rs.Primary.ID)
		}

		*application = *output

		return nil
	}
}

func testAccCheckApplicationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EMRServerlessConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_emrserverless_application" {
			continue
		}

		_, err := tfemrserverless.FindApplicationByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EMR Serverless Application %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccApplicationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_emrserverless_application" "test" {
  name          = %[1]q
  release_label = "emr-6.6.0"
  type          = "hive"
}
`, rName)
}

func testAccApplicationConfig_initialCapacity(rName, cpu string) string {
	return fmt.Sprintf(`
resource "aws_emrserverless_application" "test" {
  name          = %[1]q
  release_label = "emr-6.6.0"
  type          = "hive"

  initial_capacity {
    initial_capacity_type = "HiveDriver"

    initial_capacity_config {
      worker_count = 1
      worker_configuration {
        cpu    = %[2]q
        memory = "10 GB"
      }
    }
  }
}
`, rName, cpu)
}

func testAccApplicationConfig_maxCapacity(rName, cpu string) string {
	return fmt.Sprintf(`
resource "aws_emrserverless_application" "test" {
  name          = %[1]q
  release_label = "emr-6.6.0"
  type          = "hive"

  maximum_capacity {
    cpu    = %[2]q
    memory = "10 GB"
  }
}
`, rName, cpu)
}

func testAccApplicationConfig_network(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name        = %[1]q
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_emrserverless_application" "test" {
  name          = %[1]q
  release_label = "emr-6.6.0"
  type          = "hive"

  network_configuration {
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = aws_subnet.test[*].id
  }
}
`, rName))
}

func testAccApplicationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_emrserverless_application" "test" {
  name          = %[1]q
  release_label = "emr-6.6.0"
  type          = "hive"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccApplicationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_emrserverless_application" "test" {
  name          = %[1]q
  release_label = "emr-6.6.0"
  type          = "hive"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
