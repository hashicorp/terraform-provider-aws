// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dax_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dax"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dax/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfdax "github.com/hashicorp/terraform-provider-aws/internal/service/dax"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDAXCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dc awstypes.Cluster
	rString := sdkacctest.RandString(10)
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_dax_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DAXServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dc),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "dax", regexache.MustCompile("cache/.+")),
					resource.TestCheckResourceAttr(
						resourceName, "cluster_endpoint_encryption_type", "NONE"),
					resource.TestMatchResourceAttr(
						resourceName, names.AttrClusterName, regexache.MustCompile(`^tf-\w+$`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrIAMRoleARN, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(
						resourceName, "node_type", "dax.t3.small"),
					resource.TestCheckResourceAttr(
						resourceName, "replication_factor", acctest.Ct1),
					resource.TestCheckResourceAttr(
						resourceName, names.AttrDescription, "test cluster"),
					resource.TestMatchResourceAttr(
						resourceName, names.AttrParameterGroupName, regexache.MustCompile(`^default.dax`)),
					resource.TestMatchResourceAttr(
						resourceName, "maintenance_window", regexache.MustCompile(`^\w{3}:\d{2}:\d{2}-\w{3}:\d{2}:\d{2}$`)),
					resource.TestCheckResourceAttr(
						resourceName, "subnet_group_name", "default"),
					resource.TestMatchResourceAttr(
						resourceName, "nodes.0.id", regexache.MustCompile(`^tf-[\w-]+$`)),
					resource.TestMatchResourceAttr(
						resourceName, "configuration_endpoint", regexache.MustCompile(`:\d+$`)),
					resource.TestCheckResourceAttrSet(
						resourceName, "cluster_address"),
					resource.TestMatchResourceAttr(
						resourceName, names.AttrPort, regexache.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(
						resourceName, "server_side_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(
						resourceName, "server_side_encryption.0.enabled", acctest.CtFalse),
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
	ctx := acctest.Context(t)
	var dc awstypes.Cluster
	rString := sdkacctest.RandString(10)
	resourceName := "aws_dax_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DAXServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_resizeSingleNode(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dc),
					resource.TestCheckResourceAttr(
						resourceName, "replication_factor", acctest.Ct1),
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
					testAccCheckClusterExists(ctx, resourceName, &dc),
					resource.TestCheckResourceAttr(
						resourceName, "replication_factor", acctest.Ct2),
				),
			},
			{
				Config: testAccClusterConfig_resizeSingleNode(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dc),
					resource.TestCheckResourceAttr(
						resourceName, "replication_factor", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccDAXCluster_Encryption_disabled(t *testing.T) {
	ctx := acctest.Context(t)
	var dc awstypes.Cluster
	rString := sdkacctest.RandString(10)
	resourceName := "aws_dax_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DAXServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encryption(rString, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dc),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", acctest.CtFalse),
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
	ctx := acctest.Context(t)
	var dc awstypes.Cluster
	rString := sdkacctest.RandString(10)
	resourceName := "aws_dax_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DAXServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_encryption(rString, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dc),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", acctest.CtTrue),
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
	ctx := acctest.Context(t)
	var dc awstypes.Cluster
	rString := sdkacctest.RandString(10)
	resourceName := "aws_dax_cluster.test"
	clusterEndpointEncryptionType := "NONE"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DAXServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_endpointEncryption(rString, clusterEndpointEncryptionType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dc),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var dc awstypes.Cluster
	rString := sdkacctest.RandString(10)
	resourceName := "aws_dax_cluster.test"
	clusterEndpointEncryptionType := "TLS"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DAXServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_endpointEncryption(rString, clusterEndpointEncryptionType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dc),
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

func TestAccDAXCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var dc awstypes.Cluster
	rString := sdkacctest.RandString(10)
	resourceName := "aws_dax_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DAXServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &dc),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdax.ResourceCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DAXClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dax_cluster" {
				continue
			}
			res, err := conn.DescribeClusters(ctx, &dax.DescribeClustersInput{
				ClusterNames: []string{rs.Primary.ID},
			})
			if err != nil {
				// Verify the error is what we want
				if errs.IsA[*awstypes.ClusterNotFoundFault](err) {
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
}

func testAccCheckClusterExists(ctx context.Context, n string, v *awstypes.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DAX cluster ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DAXClient(ctx)
		resp, err := conn.DescribeClusters(ctx, &dax.DescribeClustersInput{
			ClusterNames: []string{rs.Primary.ID},
		})
		if err != nil {
			return fmt.Errorf("DAX error: %v", err)
		}

		for _, c := range resp.Clusters {
			if aws.ToString(c.ClusterName) == rs.Primary.ID {
				*v = c
			}
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DAXClient(ctx)

	input := &dax.DescribeClustersInput{}

	_, err := conn.DescribeClusters(ctx, input)
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
  node_type          = "dax.t3.small"
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
  node_type          = "dax.t3.small"
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
  node_type                        = "dax.t3.small"
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
  node_type          = "dax.r5.large"
  replication_factor = 1
}
`, baseConfig, rString)
}

func testAccClusterConfig_resizeMultiNode(rString string) string {
	return fmt.Sprintf(`%s
resource "aws_dax_cluster" "test" {
  cluster_name       = "tf-%s"
  iam_role_arn       = aws_iam_role.test.arn
  node_type          = "dax.r5.large"
  replication_factor = 2
}
`, baseConfig, rString)
}
