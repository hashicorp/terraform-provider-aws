package glue_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/glue"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccGlueDevEndpoint_basic(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDevEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("devEndpoint/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "arguments.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "number_of_nodes", "5"),
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

func TestAccGlueDevEndpoint_arguments(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDevEndpointConfig_arguments(rName, "--arg1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "arguments.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "arguments.--arg1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDevEndpointConfig_arguments2(rName, "--arg1", "value1updated", "--arg2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "arguments.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "arguments.--arg1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "arguments.--arg2", "value2"),
				),
			},
			{
				Config: testAccDevEndpointConfig_arguments(rName, "--arg2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "arguments.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "arguments.--arg2", "value2"),
				),
			},
		},
	})
}

func TestAccGlueDevEndpoint_extraJarsS3Path(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	extraJarsS3Path := "foo"
	extraJarsS3PathUpdated := "bar"
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDevEndpointConfig_extraJarsS3Path(rName, extraJarsS3Path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "extra_jars_s3_path", extraJarsS3Path),
				),
			},
			{
				Config: testAccDevEndpointConfig_extraJarsS3Path(rName, extraJarsS3PathUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
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

func TestAccGlueDevEndpoint_extraPythonLibsS3Path(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	extraPythonLibsS3Path := "foo"
	extraPythonLibsS3PathUpdated := "bar"
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDevEndpointConfig_extraPythonLibsS3Path(rName, extraPythonLibsS3Path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "extra_python_libs_s3_path", extraPythonLibsS3Path),
				),
			},
			{
				Config: testAccDevEndpointConfig_extraPythonLibsS3Path(rName, extraPythonLibsS3PathUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
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

func TestAccGlueDevEndpoint_glueVersion(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccDevEndpointConfig_version(rName, "1"),
				ExpectError: regexp.MustCompile(`must match version pattern X.X`),
			},
			{
				Config: testAccDevEndpointConfig_version(rName, "1.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "glue_version", "1.0"),
				),
			},
			{
				Config: testAccDevEndpointConfig_version(rName, "0.9"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "glue_version", "0.9"),
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

func TestAccGlueDevEndpoint_numberOfNodes(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccDevEndpointConfig_numberOfNodes(rName, 1),
				ExpectError: regexp.MustCompile(`expected number_of_nodes to be at least`),
			},
			{
				Config: testAccDevEndpointConfig_numberOfNodes(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "number_of_nodes", "2"),
				),
			},
			{
				Config: testAccDevEndpointConfig_numberOfNodes(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "number_of_nodes", "5"),
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

func TestAccGlueDevEndpoint_numberOfWorkers(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccDevEndpointConfig_numberOfWorkers(rName, 1),
				ExpectError: regexp.MustCompile(`expected number_of_workers to be at least`),
			},
			{
				Config: testAccDevEndpointConfig_numberOfWorkers(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "number_of_workers", "2"),
				),
			},
			{
				Config: testAccDevEndpointConfig_numberOfWorkers(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "number_of_workers", "5"),
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

func TestAccGlueDevEndpoint_publicKey(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	publicKey1, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}
	publicKey2, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDevEndpointConfig_publicKey(rName, publicKey1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "public_key", publicKey1),
				),
			},
			{
				Config: testAccDevEndpointConfig_publicKey(rName, publicKey2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
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

func TestAccGlueDevEndpoint_publicKeys(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	publicKey1, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}
	publicKey2, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}
	publicKey3, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}
	publicKey4, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDevEndpointConfig_publicKeys2(rName, publicKey1, publicKey2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "public_keys.#", "2"),
				),
			},
			{
				Config: testAccDevEndpointConfig_publicKeys3(rName, publicKey1, publicKey3, publicKey4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "public_keys.#", "3"),
				),
			},
			{
				Config: testAccDevEndpointConfig_publicKeys4(rName, publicKey1, publicKey1, publicKey3, publicKey4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
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

func TestAccGlueDevEndpoint_security(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDevEndpointConfig_securityConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
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
func TestAccGlueDevEndpoint_SubnetID_securityGroupIDs(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDevEndpointConfig_SubnetID_securityGroupIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
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

func TestAccGlueDevEndpoint_tags(t *testing.T) {
	var endpoint1, endpoint2, endpoint3 glue.DevEndpoint

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDevEndpointConfig_Tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint1),
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
				Config: testAccDevEndpointConfig_Tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDevEndpointConfig_Tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccGlueDevEndpoint_workerType(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDevEndpointConfig_WorkerType_standard(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "worker_type", glue.WorkerTypeStandard),
				),
			},
			{
				Config: testAccDevEndpointConfig_workerType(rName, glue.WorkerTypeG1x),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "worker_type", glue.WorkerTypeG1x),
				),
			},
			{
				Config: testAccDevEndpointConfig_workerType(rName, glue.WorkerTypeG2x),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
					resource.TestCheckResourceAttr(resourceName, "worker_type", glue.WorkerTypeG2x),
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

func TestAccGlueDevEndpoint_disappears(t *testing.T) {
	var endpoint glue.DevEndpoint

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_dev_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDevEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDevEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDevEndpointExists(resourceName, &endpoint),
					acctest.CheckResourceDisappears(acctest.Provider, tfglue.ResourceDevEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDevEndpointExists(resourceName string, v *glue.DevEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Glue Dev Endpoint ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn

		output, err := tfglue.FindDevEndpointByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckDevEndpointDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_dev_endpoint" {
			continue
		}

		_, err := tfglue.FindDevEndpointByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Glue Dev Endpoint %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccDevEndpointConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = "AWSGlueServiceRole-%[1]s"
  assume_role_policy = data.aws_iam_policy_document.service.json
}

data "aws_iam_policy_document" "service" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["glue.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "glue_service_role" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSGlueServiceRole"
  role       = aws_iam_role.test.name
}

data "aws_partition" "current" {}
`, rName)
}

func testAccDevEndpointConfig_basic(rName string) string {
	return testAccDevEndpointConfig_base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name     = %q
  role_arn = aws_iam_role.test.arn
}
`, rName)
}

func testAccDevEndpointConfig_arguments(rName, argKey, argValue string) string {
	return testAccDevEndpointConfig_base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  arguments = {
    %[2]q = %[3]q
  }
}
`, rName, argKey, argValue)
}

func testAccDevEndpointConfig_arguments2(rName, argKey1, argValue1, argKey2, argValue2 string) string {
	return testAccDevEndpointConfig_base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn
  arguments = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, argKey1, argValue1, argKey2, argValue2)
}

func testAccDevEndpointConfig_extraJarsS3Path(rName string, extraJarsS3Path string) string {
	return testAccDevEndpointConfig_base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name               = %q
  role_arn           = aws_iam_role.test.arn
  extra_jars_s3_path = %q
}
`, rName, extraJarsS3Path)
}

func testAccDevEndpointConfig_extraPythonLibsS3Path(rName string, extraPythonLibsS3Path string) string {
	return testAccDevEndpointConfig_base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name                      = %q
  role_arn                  = aws_iam_role.test.arn
  extra_python_libs_s3_path = %q
}
`, rName, extraPythonLibsS3Path)
}

func testAccDevEndpointConfig_version(rName string, glueVersion string) string {
	return testAccDevEndpointConfig_base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name         = %[1]q
  role_arn     = aws_iam_role.test.arn
  glue_version = %[2]q
}
`, rName, glueVersion)
}

func testAccDevEndpointConfig_numberOfNodes(rName string, numberOfNodes int) string {
	return testAccDevEndpointConfig_base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name            = %q
  role_arn        = aws_iam_role.test.arn
  number_of_nodes = %d
}
`, rName, numberOfNodes)
}

func testAccDevEndpointConfig_numberOfWorkers(rName string, numberOfWorkers int) string {
	return testAccDevEndpointConfig_base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name              = %q
  role_arn          = aws_iam_role.test.arn
  worker_type       = "G.1X"
  number_of_workers = %d
}
`, rName, numberOfWorkers)
}

func testAccDevEndpointConfig_publicKey(rName string, publicKey string) string {
	return testAccDevEndpointConfig_base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name       = %q
  role_arn   = aws_iam_role.test.arn
  public_key = "%s"
}
`, rName, publicKey)
}

func testAccDevEndpointConfig_publicKeys2(rName string, publicKey1 string, publicKey2 string) string {
	return testAccDevEndpointConfig_base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name        = %[1]q
  role_arn    = aws_iam_role.test.arn
  public_keys = [%[2]q, %[3]q]
}
`, rName, publicKey1, publicKey2)
}

func testAccDevEndpointConfig_publicKeys3(rName string, publicKey1 string, publicKey2 string, publicKey3 string) string {
	return testAccDevEndpointConfig_base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name        = %[1]q
  role_arn    = aws_iam_role.test.arn
  public_keys = [%[2]q, %[3]q, %[4]q]
}
`, rName, publicKey1, publicKey2, publicKey3)
}

func testAccDevEndpointConfig_publicKeys4(rName string, publicKey1 string, publicKey2 string, publicKey3 string, publicKey4 string) string {
	return testAccDevEndpointConfig_base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name        = %[1]q
  role_arn    = aws_iam_role.test.arn
  public_keys = [%[2]q, %[3]q, %[4]q, %[5]q]
}
`, rName, publicKey1, publicKey2, publicKey3, publicKey4)
}

func testAccDevEndpointConfig_securityConfiguration(rName string) string {
	return testAccDevEndpointConfig_base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name                   = %[1]q
  role_arn               = aws_iam_role.test.arn
  security_configuration = aws_glue_security_configuration.test.name
}

resource "aws_glue_security_configuration" "test" {
  name = %[1]q

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
`, rName)
}

func testAccDevEndpointConfig_SubnetID_securityGroupIDs(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), testAccDevEndpointConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name               = %[1]q
  role_arn           = aws_iam_role.test.arn
  subnet_id          = aws_subnet.test.id
  security_group_ids = [aws_security_group.test.id]
}

resource "aws_vpc_endpoint" "s3" {
  vpc_id       = aws_vpc.test.id
  service_name = data.aws_vpc_endpoint_service.s3.service_name
}

data "aws_vpc_endpoint_service" "s3" {
  service      = "s3"
  service_type = "Gateway"
}

resource "aws_vpc_endpoint_route_table_association" "test" {
  vpc_endpoint_id = aws_vpc_endpoint.s3.id
  route_table_id  = aws_vpc.test.main_route_table_id
}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }

  timeouts {
    delete = "40m"
  }
  depends_on = [aws_iam_role_policy_attachment.glue_service_role]
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  timeouts {
    delete = "40m"
  }
  depends_on = [aws_iam_role_policy_attachment.glue_service_role]
}
`, rName))
}

func testAccDevEndpointConfig_Tags1(rName, tagKey1, tagValue1 string) string {
	return testAccJobConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDevEndpointConfig_Tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccJobConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccDevEndpointConfig_workerType(rName, workerType string) string {
	return testAccDevEndpointConfig_base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name              = %[1]q
  role_arn          = aws_iam_role.test.arn
  worker_type       = %[2]q
  number_of_workers = 2
}
`, rName, workerType)
}

func testAccDevEndpointConfig_WorkerType_standard(rName string) string {
	return testAccDevEndpointConfig_base(rName) + fmt.Sprintf(`
resource "aws_glue_dev_endpoint" "test" {
  name              = %[1]q
  role_arn          = aws_iam_role.test.arn
  worker_type       = "Standard"
  number_of_workers = 2
}
`, rName)
}
