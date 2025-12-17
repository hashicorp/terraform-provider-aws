// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftLogging_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var log redshift.DescribeLoggingStatusOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_logging.test"
	clusterResourceName := "aws_redshift_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RedshiftEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingExists(ctx, resourceName, &log),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClusterIdentifier, clusterResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", string(awstypes.LogDestinationTypeCloudwatch)),
					resource.TestCheckResourceAttr(resourceName, "log_exports.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "log_exports.*", string(tfredshift.LogExportsConnectionLog)),
					resource.TestCheckTypeSetElemAttr(resourceName, "log_exports.*", string(tfredshift.LogExportsUserActivityLog)),
					resource.TestCheckTypeSetElemAttr(resourceName, "log_exports.*", string(tfredshift.LogExportsUserLog)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"log_destination_type",
				},
			},
		},
	})
}

func TestAccRedshiftLogging_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var log redshift.DescribeLoggingStatusOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_logging.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RedshiftEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingExists(ctx, resourceName, &log),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfredshift.ResourceLogging, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedshiftLogging_disappears_Cluster(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var log redshift.DescribeLoggingStatusOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_logging.test"
	clusterResourceName := "aws_redshift_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RedshiftEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingExists(ctx, resourceName, &log),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshift.ResourceCluster(), clusterResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedshiftLogging_s3(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var log redshift.DescribeLoggingStatusOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_logging.test"
	clusterResourceName := "aws_redshift_cluster.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.RedshiftEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLoggingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLoggingConfig_s3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoggingExists(ctx, resourceName, &log),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClusterIdentifier, clusterResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrBucketName, bucketResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", string(awstypes.LogDestinationTypeS3)),
					resource.TestCheckResourceAttr(resourceName, names.AttrS3KeyPrefix, "testprefix/"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"log_destination_type",
				},
			},
		},
	})
}

func testAccCheckLoggingDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_logging" {
				continue
			}

			_, err := tfredshift.FindLoggingStatusByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Logging %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLoggingExists(ctx context.Context, n string, v *redshift.DescribeLoggingStatusOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftClient(ctx)

		output, err := tfredshift.FindLoggingStatusByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccLoggingConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_cluster" "test" {
  cluster_identifier    = %[1]q
  database_name         = "mydb"
  master_username       = "foo_test"
  master_password       = "Mustbe8characters"
  multi_az              = false
  node_type             = "ra3.large"
  allow_version_upgrade = false
  skip_final_snapshot   = true
}
`, rName)
}

func testAccLoggingConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccLoggingConfigBase(rName),
		`
resource "aws_redshift_logging" "test" {
  cluster_identifier   = aws_redshift_cluster.test.cluster_identifier
  log_destination_type = "cloudwatch"
  log_exports          = ["connectionlog", "useractivitylog", "userlog"]
}
`)
}

func testAccLoggingConfig_s3(rName string) string {
	return acctest.ConfigCompose(
		testAccLoggingConfigBase(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  policy = <<EOF
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "redshift.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "s3:PutObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}/*"
    },
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "redshift.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}"
    }
  ]
}
EOF
}

resource "aws_redshift_logging" "test" {
  depends_on = [aws_s3_bucket_policy.test]

  cluster_identifier   = aws_redshift_cluster.test.cluster_identifier
  bucket_name          = aws_s3_bucket.test.bucket
  s3_key_prefix        = "testprefix/"
  log_destination_type = "s3"
}
`, rName))
}
