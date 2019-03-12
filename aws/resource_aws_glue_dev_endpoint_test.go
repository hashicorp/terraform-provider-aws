package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const (
	GlueDevEndpointResourcePrefix = "tf-acc-test"
	publicKey1                    = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 foo1@bar.com"
	publicKey2                    = "ssh-rsa AAAAB3NzaC1yc2EAAAABJQAAAQEAq6U3HQYC4g8WzU147gZZ7CKQH8TgYn3chZGRPxaGmHW1RUwsyEs0nmombmIhwxudhJ4ehjqXsDLoQpd6+c7BuLgTMvbv8LgE9LX53vnljFe1dsObsr/fYLvpU9LTlo8HgHAqO5ibNdrAUvV31ronzCZhms/Gyfdaue88Fd0/YnsZVGeOZPayRkdOHSpqme2CBrpa8myBeL1CWl0LkDG4+YCURjbaelfyZlIApLYKy3FcCan9XQFKaL32MJZwCgzfOvWIMtYcU8QtXMgnA3/I3gXk8YDUJv5P4lj0s/PJXuTM8DygVAUtebNwPuinS7wwonm5FXcWMuVGsVpG5K7FGQ== foo2@bar.com"
	publicKey3                    = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCtCk0lzMj1gPEOjdfQ37AIxCyETqJBubaMWuB4bgvGHp8LvEghr2YDl2bml1JrE1EOcZhPnIwgyucryXKA959sTUlgbvaFN7vmpVze56Q9tVU6BJQxOdaRoy5FcQMET9LB6SdbXk+V4CkDMsQNaFXezpg98HgCj+V7+bBWsfI6U63IESlWKK7kraCom8EWxkQk4mk9fizE2I+KrtiqN4xcah02LFG6IMnS+Xy3CDhcpZeYzWOV6zhcf675UJOdg/pLgQbUhhiwTOJFgRo8IcvE3iBrRMz508ppx6vLLr8J+3B8ujykc+/3ZSGfQfx6rO+OuSskhG5FLI6icbQBtBzf foo3@bar.com"
	publicKey4                    = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 foo4@bar.com"
)

func init() {
	resource.AddTestSweepers("aws_glue_dev_endpoint", &resource.Sweeper{
		Name: "aws_glue_dev_endpoint",
		F:    testSweepGlueDevEndpoint,
	})
}

func testSweepGlueDevEndpoint(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).glueconn

	input := &glue.GetDevEndpointsInput{}
	err = conn.GetDevEndpointsPages(input, func(page *glue.GetDevEndpointsOutput, lastPage bool) bool {
		if len(page.DevEndpoints) == 0 {
			log.Printf("[INFO] No Glue Dev Endpoints to sweep")
			return false
		}
		for _, endpoint := range page.DevEndpoints {
			name := aws.StringValue(endpoint.EndpointName)
			if !strings.HasPrefix(name, GlueDevEndpointResourcePrefix) {
				log.Printf("[INFO] Skipping Glue Dev Endpoint: %s", name)
				continue
			}

			log.Printf("[INFO] Deleting Glue Dev Endpoint: %s", name)
			_, err := conn.DeleteDevEndpoint(&glue.DeleteDevEndpointInput{
				EndpointName: aws.String(name),
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete Glue Dev Endpoint %s: %s", name, err)
			}
		}
		return !lastPage
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Glue Dev Endpoint sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error retrieving Glue Dev Endpoint: %s", err)
	}

	return nil
}

func TestAccGlueDevEndpoint_Basic(t *testing.T) {
	t.Skip()

	var endpoint glue.DevEndpoint
	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),

					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestMatchResourceAttr(resourceName, "role_arn", regexp.MustCompile(`^arn:[^:]+:iam::[^:]+:role/AWSGlueServiceRole-tf-acc-test-[0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
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

func TestAccGlueDevEndpoint_ExtraJarsS3Path(t *testing.T) {
	t.Skip()

	var endpoint glue.DevEndpoint
	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	extraJarsS3Path := "foo"
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_ExtraJarsS3Path(rName, extraJarsS3Path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),

					resource.TestCheckResourceAttr(resourceName, "extra_jars_s3_path", extraJarsS3Path),
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

func TestAccGlueDevEndpoint_ExtraPythonLibsS3Path(t *testing.T) {
	t.Skip()

	var endpoint glue.DevEndpoint
	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	extraPythonLibsS3Path := "foo"
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_ExtraPythonLibsS3Path(rName, extraPythonLibsS3Path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),

					resource.TestCheckResourceAttr(resourceName, "extra_python_libs_s3_path", extraPythonLibsS3Path),
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

func TestAccGlueDevEndpoint_NumberOfNodes(t *testing.T) {
	t.Skip()

	var endpoint glue.DevEndpoint
	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_NumberOfNodes(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),

					resource.TestCheckResourceAttr(resourceName, "number_of_nodes", "2"),
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

func TestAccGlueDevEndpoint_PublicKey(t *testing.T) {
	t.Skip()

	var endpoint glue.DevEndpoint
	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_PublicKey(rName, publicKey1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),

					resource.TestCheckResourceAttr(resourceName, "public_key", publicKey1),
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

func TestAccGlueDevEndpoint_PublicKeys(t *testing.T) {
	t.Skip()

	var endpoint glue.DevEndpoint
	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_PublicKeys(rName, publicKey1, publicKey2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),

					resource.TestCheckResourceAttr(resourceName, "public_keys.#", "2"),
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

func TestAccGlueDevEndpoint_SecurityConfiguration(t *testing.T) {
	t.Skip()

	var endpoint glue.DevEndpoint
	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_SecurityConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),

					resource.TestCheckResourceAttr(resourceName, "security_configuration", rName),
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

// Note: Either none or both of subnetId and securityGroupIds must be specified.
func TestAccGlueDevEndpoint_SubnetID_SecurityGroupIDs(t *testing.T) {
	var endpoint glue.DevEndpoint
	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_SubnetID_SecurityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),

					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", "data.aws_security_group.default", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", "aws_subnet.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", "aws_subnet.test", "availability_zone"),
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

func TestAccGlueDevEndpoint_Update_PublicKey(t *testing.T) {
	t.Skip()

	var endpoint glue.DevEndpoint
	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_PublicKey(rName, publicKey1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),

					resource.TestCheckResourceAttr(resourceName, "public_key", publicKey1),
				),
			},
			{
				Config: testAccGlueDevEndpointConfig_PublicKey(rName, publicKey2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),

					resource.TestCheckResourceAttr(resourceName, "public_key", publicKey2),
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

func TestAccGlueDevEndpoint_Update_ExtraJarsS3Path(t *testing.T) {
	var endpoint glue.DevEndpoint
	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	extraJarsS3Path := "foo"
	extraJarsS3PathUpdated := "bar"
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_ExtraJarsS3Path(rName, extraJarsS3Path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),

					resource.TestCheckResourceAttr(resourceName, "extra_jars_s3_path", extraJarsS3Path),
				),
			},
			{
				Config: testAccGlueDevEndpointConfig_ExtraJarsS3Path(rName, extraJarsS3PathUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),

					resource.TestCheckResourceAttr(resourceName, "extra_jars_s3_path", extraJarsS3PathUpdated),
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

func TestAccGlueDevEndpoint_Update_ExtraPythonLibsS3Path(t *testing.T) {
	t.Skip()

	var endpoint glue.DevEndpoint
	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	extraPythonLibsS3Path := "foo"
	extraPythonLibsS3PathUpdated := "bar"

	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_ExtraPythonLibsS3Path(rName, extraPythonLibsS3Path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),

					resource.TestCheckResourceAttr(resourceName, "extra_python_libs_s3_path", extraPythonLibsS3Path),
				),
			},
			{
				Config: testAccGlueDevEndpointConfig_ExtraPythonLibsS3Path(rName, extraPythonLibsS3PathUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),

					resource.TestCheckResourceAttr(resourceName, "extra_python_libs_s3_path", extraPythonLibsS3PathUpdated),
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

func TestAccGlueDevEndpoint_Update_PublicKeys(t *testing.T) {
	var endpoint glue.DevEndpoint
	rName := acctest.RandomWithPrefix(GlueDevEndpointResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlueDevEndpointConfig_PublicKeys(rName, publicKey1, publicKey2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),

					resource.TestCheckResourceAttr(resourceName, "public_keys.#", "2"),
				),
			},
			{
				Config: testAccGlueDevEndpointConfig_Update_PublicKeys(rName, publicKey1, publicKey3, publicKey4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueDevEndpointExists(resourceName, &endpoint),

					resource.TestCheckResourceAttr(resourceName, "public_keys.#", "3"),
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

func testAccCheckAWSGlueDevEndpointExists(resourceName string, endpoint *glue.DevEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).glueconn
		output, err := conn.GetDevEndpoint(&glue.GetDevEndpointInput{
			EndpointName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("no Glue Dev Endpoint")
		}

		*endpoint = *output.DevEndpoint

		return nil
	}
}

func testAccCheckAWSGlueDevEndpointDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_dev_endpoint" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).glueconn
		output, err := conn.GetDevEndpoint(&glue.GetDevEndpointInput{
			EndpointName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
				return nil
			}
			return err
		}

		endpoint := output.DevEndpoint
		if endpoint != nil && aws.StringValue(endpoint.EndpointName) == rs.Primary.ID {
			return fmt.Errorf("the Glue Dev Endpoint %s still exists", rs.Primary.ID)
		}

		return nil
	}

	return nil
}

func testAccGlueDevEndpointConfig_Base(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "AWSGlueServiceRole-%s"
  assume_role_policy = "${data.aws_iam_policy_document.test.json}"
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["glue.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "test-AWSGlueServiceRole" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSGlueServiceRole"
  role       = "${aws_iam_role.test.name}"
}
`, rName)
}

func testAccGlueDevEndpointConfig_Basic(rName string) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name = %q
  role_arn = "${aws_iam_role.test.arn}"
}
`, rName)
}

func testAccGlueDevEndpointConfig_ExtraJarsS3Path(rName string, extraJarsS3Path string) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name = %q
  role_arn = "${aws_iam_role.test.arn}"
  extra_jars_s3_path = %q
}
`, rName, extraJarsS3Path)
}

func testAccGlueDevEndpointConfig_ExtraPythonLibsS3Path(rName string, extraPythonLibsS3Path string) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name = %q
  role_arn = "${aws_iam_role.test.arn}"
  extra_python_libs_s3_path = %q
}
`, rName, extraPythonLibsS3Path)
}

func testAccGlueDevEndpointConfig_NumberOfNodes(rName string, numberOfNodes int) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name = %q
  role_arn = "${aws_iam_role.test.arn}"
  number_of_nodes = %d
}
`, rName, numberOfNodes)
}

func testAccGlueDevEndpointConfig_PublicKey(rName string, publicKey string) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name = %q
  role_arn = "${aws_iam_role.test.arn}"
  public_key = "%s"
}
`, rName, publicKey)
}

func testAccGlueDevEndpointConfig_PublicKeys(rName string, publicKey1 string, publicKey2 string) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name = %q
  role_arn = "${aws_iam_role.test.arn}"
  public_keys = ["%s", "%s"]
}
`, rName, publicKey1, publicKey2)
}

func testAccGlueDevEndpointConfig_Update_PublicKeys(rName string, publicKey1 string, publicKey2 string, publicKey3 string) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name = %q
  role_arn = "${aws_iam_role.test.arn}"
  public_keys = ["%s", "%s", "%s"]
}
`, rName, publicKey1, publicKey2, publicKey3)
}

func testAccGlueDevEndpointConfig_SecurityConfiguration(rName string) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name = %q
  role_arn = "${aws_iam_role.test.arn}"
  security_configuration = "${aws_glue_security_configuration.test.name}"
}

resource "aws_glue_security_configuration" "test" {
  name = %q

  encryption_configuration {
    cloudwatch_encryption {
      cloudwatch_encryption_mode = "DISABLED"
    }

    job_bookmarks_encryption {
      job_bookmarks_encryption_mode = "DISABLED"
    }

    s3_encryption {
      s3_encryption_mode = "DISABLED"
    }
  }
}
`, rName, rName)
}

func testAccGlueDevEndpointConfig_SubnetID_SecurityGroupIDs(rName string) string {
	return testAccGlueDevEndpointConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name = %q
  role_arn = "${aws_iam_role.test.arn}"
  subnet_id = "${aws_subnet.test.id}"
  security_group_ids = ["${data.aws_security_group.default.id}"]
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = %q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = %q
  }
}

data "aws_security_group" "default" {
  vpc_id = "${aws_vpc.test.id}"
  name   = "default"
}

data "aws_vpc_endpoint_service" "s3" {
  service = "s3"
}

resource "aws_vpc_endpoint" "test" {
  service_name = "${data.aws_vpc_endpoint_service.s3.service_name}"
  vpc_id       = "${aws_vpc.test.id}"
}

resource "aws_internet_gateway" "test" {
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_eip" "test" {
  vpc = true
}

resource "aws_nat_gateway" "test" {
  allocation_id = "${aws_eip.test.id}"
  subnet_id = "${aws_subnet.test.id}"

  tags = {
    Name = %q
  }

  depends_on = ["aws_internet_gateway.test"]
}
`, rName, rName, rName, rName)
}
