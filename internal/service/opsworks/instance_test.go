package opsworks_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccOpsWorksInstance_basic(t *testing.T) {
	var opsinst opsworks.Instance
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_instance.test"
	dataSourceName := "data.aws_availability_zones.available"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &opsinst),
					testAccCheckInstanceAttributes(&opsinst),
					resource.TestCheckResourceAttr(resourceName, "hostname", "tf-acc1"),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "state", "stopped"),
					resource.TestCheckResourceAttr(resourceName, "layer_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "install_updates_on_boot", "true"),
					resource.TestCheckResourceAttr(resourceName, "architecture", "x86_64"),
					resource.TestCheckResourceAttr(resourceName, "tenancy", "default"),
					resource.TestCheckResourceAttr(resourceName, "os", "Amazon Linux 2016.09"),                       // inherited from opsworks_stack_test
					resource.TestCheckResourceAttr(resourceName, "root_device_type", "ebs"),                          // inherited from opsworks_stack_test
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", dataSourceName, "names.0"), // inherited from opsworks_stack_test
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"state"}, //state is something we pass to the API and get back as status :(
			},
			{
				Config: testAccInstanceConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &opsinst),
					testAccCheckInstanceAttributes(&opsinst),
					resource.TestCheckResourceAttr(resourceName, "hostname", "tf-acc1"),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.small"),
					resource.TestCheckResourceAttr(resourceName, "layer_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "os", "Amazon Linux 2015.09"),
					resource.TestCheckResourceAttr(resourceName, "tenancy", "default"),
				),
			},
		},
	})
}

func TestAccOpsWorksInstance_updateHostNameForceNew(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_instance.test"
	var before, after opsworks.Instance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_create(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "hostname", "tf-acc1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"state"},
			},
			{
				Config: testAccInstanceConfig_updateHostName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "hostname", "tf-acc2"),
					testAccCheckInstanceRecreated(t, &before, &after),
				),
			},
		},
	})
}

func testAccCheckInstanceRecreated(t *testing.T,
	before, after *opsworks.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.InstanceId == *after.InstanceId {
			t.Fatalf("Expected change of OpsWorks Instance IDs, but both were %s", *before.InstanceId)
		}
		return nil
	}
}

func testAccCheckInstanceExists(
	n string, opsinst *opsworks.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Opsworks Instance is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn

		params := &opsworks.DescribeInstancesInput{
			InstanceIds: []*string{&rs.Primary.ID},
		}
		resp, err := conn.DescribeInstances(params)

		if err != nil {
			return err
		}

		if v := len(resp.Instances); v != 1 {
			return fmt.Errorf("Expected 1 request returned, got %d", v)
		}

		*opsinst = *resp.Instances[0]

		return nil
	}
}

func testAccCheckInstanceAttributes(
	opsinst *opsworks.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Depending on the timing, the state could be requested or stopped
		if *opsinst.Status != "stopped" && *opsinst.Status != "requested" {
			return fmt.Errorf("Unexpected request status: %s", *opsinst.Status)
		}
		if *opsinst.Architecture != "x86_64" {
			return fmt.Errorf("Unexpected architecture: %s", *opsinst.Architecture)
		}
		if *opsinst.Tenancy != "default" {
			return fmt.Errorf("Unexpected tenancy: %s", *opsinst.Tenancy)
		}
		if *opsinst.InfrastructureClass != "ec2" {
			return fmt.Errorf("Unexpected infrastructure class: %s", *opsinst.InfrastructureClass)
		}
		if *opsinst.RootDeviceType != "ebs" {
			return fmt.Errorf("Unexpected root device type: %s", *opsinst.RootDeviceType)
		}
		if *opsinst.VirtualizationType != "hvm" {
			return fmt.Errorf("Unexpected virtualization type: %s", *opsinst.VirtualizationType)
		}
		return nil
	}
}

func testAccCheckInstanceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_opsworks_instance" {
			continue
		}
		req := &opsworks.DescribeInstancesInput{
			InstanceIds: aws.StringSlice([]string{rs.Primary.ID}),
		}

		output, err := conn.DescribeInstances(req)

		if tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil && len(output.Instances) > 0 {
			for _, instance := range output.Instances {
				if aws.StringValue(instance.InstanceId) != rs.Primary.ID {
					continue
				}
				return fmt.Errorf("Expected OpsWorks instance (%s) to be gone, but was still found", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccInstanceConfig_updateHostName(rName string) string {
	return acctest.ConfigCompose(
		testAccStackConfig_vpcCreate(rName),
		fmt.Sprintf(`
resource "aws_security_group" "tf-ops-acc-web" {
  name = "%[1]s-web"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "tf-ops-acc-php" {
  name = "%[1]s-php"

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_opsworks_static_web_layer" "test" {
  stack_id = aws_opsworks_stack.test.id

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-web.id,
  ]
}

resource "aws_opsworks_php_app_layer" "test" {
  stack_id = aws_opsworks_stack.test.id

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-php.id,
  ]
}

resource "aws_opsworks_instance" "test" {
  stack_id = aws_opsworks_stack.test.id

  layer_ids = [
    aws_opsworks_static_web_layer.test.id,
  ]

  instance_type = "t2.micro"
  state         = "stopped"
  hostname      = "tf-acc2"
}
`, rName))
}

func testAccInstanceConfig_create(rName string) string {
	return acctest.ConfigCompose(
		testAccStackConfig_vpcCreate(rName),
		fmt.Sprintf(`
resource "aws_security_group" "tf-ops-acc-web" {
  name = "%[1]s-web"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "tf-ops-acc-php" {
  name = "%[1]s-php"

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_opsworks_static_web_layer" "test" {
  stack_id = aws_opsworks_stack.test.id

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-web.id,
  ]
}

resource "aws_opsworks_php_app_layer" "test" {
  stack_id = aws_opsworks_stack.test.id

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-php.id,
  ]
}

resource "aws_opsworks_instance" "test" {
  stack_id = aws_opsworks_stack.test.id

  layer_ids = [
    aws_opsworks_static_web_layer.test.id,
  ]

  instance_type = "t2.micro"
  state         = "stopped"
  hostname      = "tf-acc1"
}
`, rName))
}

func testAccInstanceConfig_update(rName string) string {
	return acctest.ConfigCompose(
		testAccStackConfig_vpcCreate(rName),
		fmt.Sprintf(`
resource "aws_security_group" "tf-ops-acc-web" {
  name = "%[1]s-web"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "tf-ops-acc-php" {
  name = "%[1]s-php"

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_opsworks_static_web_layer" "test" {
  stack_id = aws_opsworks_stack.test.id

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-web.id,
  ]
}

resource "aws_opsworks_php_app_layer" "test" {
  stack_id = aws_opsworks_stack.test.id

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-php.id,
  ]
}

resource "aws_opsworks_instance" "test" {
  stack_id = aws_opsworks_stack.test.id

  layer_ids = [
    aws_opsworks_static_web_layer.test.id,
    aws_opsworks_php_app_layer.test.id,
  ]

  instance_type = "t2.small"
  state         = "stopped"
  hostname      = "tf-acc1"
  os            = "Amazon Linux 2015.09"

  timeouts {
    update = "15s"
  }
}
`, rName))
}
