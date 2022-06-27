package dax_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dax"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccDAXCluster_basic(t *testing.T) {
	var dc dax.Cluster
	rString := sdkacctest.RandString(10)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_dax_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dax.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dc),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "dax", regexp.MustCompile("cache/.+")),
					resource.TestCheckResourceAttr(
						resourceName, "cluster_endpoint_encryption_type", "NONE"),
					resource.TestMatchResourceAttr(
						resourceName, "cluster_name", regexp.MustCompile(`^tf-\w+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "iam_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(
						resourceName, "node_type", "dax.t2.small"),
					resource.TestCheckResourceAttr(
						resourceName, "replication_factor", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "description", "test cluster"),
					resource.TestMatchResourceAttr(
						resourceName, "parameter_group_name", regexp.MustCompile(`^default.dax`)),
					resource.TestMatchResourceAttr(
						resourceName, "maintenance_window", regexp.MustCompile(`^\w{3}:\d{2}:\d{2}-\w{3}:\d{2}:\d{2}$`)),
					resource.TestCheckResourceAttr(
						resourceName, "subnet_group_name", "default"),
					resource.TestMatchResourceAttr(
						resourceName, "nodes.0.id", regexp.MustCompile(`^tf-[\w-]+$`)),
					resource.TestMatchResourceAttr(
						resourceName, "configuration_endpoint", regexp.MustCompile(`:\d+$`)),
					resource.TestCheckResourceAttrSet(
						resourceName, "cluster_address"),
					resource.TestMatchResourceAttr(
						resourceName, "port", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(
						resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "server_side_encryption.0.enabled", "false"),
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

func TestAccDAXCluster_resize(t *testing.T) {
	var dc dax.Cluster
	rString := sdkacctest.RandString(10)
	resourceName := "aws_dax_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dax.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_resizeSingleNode(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dc),
					resource.TestCheckResourceAttr(
						resourceName, "replication_factor", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterConfig_resizeMultiNode(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dc),
					resource.TestCheckResourceAttr(
						resourceName, "replication_factor", "2"),
				),
			},
			{
				Config: testAccClusterConfig_resizeSingleNode(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dc),
					resource.TestCheckResourceAttr(
						resourceName, "replication_factor", "1"),
				),
			},
		},
	})
}

func TestAccDAXCluster_Encryption_disabled(t *testing.T) {
	var dc dax.Cluster
	rString := sdkacctest.RandString(10)
	resourceName := "aws_dax_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dax.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encryption(rString, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dc),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Ensure it shows no difference when removing server_side_encryption configuration
			{
				Config:             testAccClusterConfig_basic(rString),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccDAXCluster_Encryption_enabled(t *testing.T) {
	var dc dax.Cluster
	rString := sdkacctest.RandString(10)
	resourceName := "aws_dax_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dax.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encryption(rString, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dc),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Ensure it shows a difference when removing server_side_encryption configuration
			{
				Config:             testAccClusterConfig_basic(rString),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDAXCluster_EndpointEncryption_disabled(t *testing.T) {
	var dc dax.Cluster
	rString := sdkacctest.RandString(10)
	resourceName := "aws_dax_cluster.test"
	clusterEndpointEncryptionType := "NONE"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dax.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_endpointEncryption(rString, clusterEndpointEncryptionType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dc),
					resource.TestCheckResourceAttr(resourceName, "cluster_endpoint_encryption_type", clusterEndpointEncryptionType),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Ensure it shows no difference when removing cluster_endpoint_encryption_type configuration
			{
				Config:             testAccClusterConfig_basic(rString),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccDAXCluster_EndpointEncryption_enabled(t *testing.T) {
	var dc dax.Cluster
	rString := sdkacctest.RandString(10)
	resourceName := "aws_dax_cluster.test"
	clusterEndpointEncryptionType := "TLS"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dax.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_endpointEncryption(rString, clusterEndpointEncryptionType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &dc),
					resource.TestCheckResourceAttr(resourceName, "cluster_endpoint_encryption_type", clusterEndpointEncryptionType),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Ensure it shows a difference when removing cluster_endpoint_encryption_type configuration
			{
				Config:             testAccClusterConfig_basic(rString),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckClusterDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DAXConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dax_cluster" {
			continue
		}
		res, err := conn.DescribeClusters(&dax.DescribeClustersInput{
			ClusterNames: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			// Verify the error is what we want
			if tfawserr.ErrCodeEquals(err, dax.ErrCodeClusterNotFoundFault) {
				continue
			}
			return err
		}
		if len(res.Clusters) > 0 {
			return fmt.Errorf("still exist.")
		}
	}
	return nil
}

func testAccCheckClusterExists(n string, v *dax.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DAX cluster ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DAXConn
		resp, err := conn.DescribeClusters(&dax.DescribeClustersInput{
			ClusterNames: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return fmt.Errorf("DAX error: %v", err)
		}

		for _, c := range resp.Clusters {
			if *c.ClusterName == rs.Primary.ID {
				*v = *c
			}
		}

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DAXConn

	input := &dax.DescribeClustersInput{}

	_, err := conn.DescribeClusters(input)

	if acctest.PreCheckSkipError(err) || tfawserr.ErrMessageContains(err, "InvalidParameterValueException", "Access Denied to API Version: DAX_V3") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

const baseConfig = `
resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
       {
            "Effect": "Allow",
            "Principal": {
                "Service": "dax.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.id

  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "dynamodb:*",
            "Effect": "Allow",
            "Resource": "*"
        }
    ]
}
EOF
}
`

func testAccClusterConfig_basic(rString string) string {
	return fmt.Sprintf(`%s
resource "aws_dax_cluster" "test" {
  cluster_name       = "tf-%s"
  iam_role_arn       = aws_iam_role.test.arn
  node_type          = "dax.t2.small"
  replication_factor = 1
  description        = "test cluster"

  tags = {
    foo = "bar"
  }
}
`, baseConfig, rString)
}

func testAccClusterConfig_encryption(rString string, enabled bool) string {
	return fmt.Sprintf(`%s
resource "aws_dax_cluster" "test" {
  cluster_name       = "tf-%s"
  iam_role_arn       = aws_iam_role.test.arn
  node_type          = "dax.t2.small"
  replication_factor = 1
  description        = "test cluster"

  tags = {
    foo = "bar"
  }

  server_side_encryption {
    enabled = %t
  }
}
`, baseConfig, rString, enabled)
}

func testAccClusterConfig_endpointEncryption(rString string, encryptionType string) string {
	return fmt.Sprintf(`%s
resource "aws_dax_cluster" "test" {
  cluster_name                     = "tf-%s"
  cluster_endpoint_encryption_type = "%s"
  iam_role_arn                     = aws_iam_role.test.arn
  node_type                        = "dax.t2.small"
  replication_factor               = 1
  description                      = "test cluster"

  tags = {
    foo = "bar"
  }
}
`, baseConfig, rString, encryptionType)
}

func testAccClusterConfig_resizeSingleNode(rString string) string {
	return fmt.Sprintf(`%s
resource "aws_dax_cluster" "test" {
  cluster_name       = "tf-%s"
  iam_role_arn       = aws_iam_role.test.arn
  node_type          = "dax.r3.large"
  replication_factor = 1
}
`, baseConfig, rString)
}

func testAccClusterConfig_resizeMultiNode(rString string) string {
	return fmt.Sprintf(`%s
resource "aws_dax_cluster" "test" {
  cluster_name       = "tf-%s"
  iam_role_arn       = aws_iam_role.test.arn
  node_type          = "dax.r3.large"
  replication_factor = 2
}
`, baseConfig, rString)
}
