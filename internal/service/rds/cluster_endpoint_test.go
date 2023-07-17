// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRDSClusterEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rInt := sdkacctest.RandInt()
	var customReaderEndpoint rds.DBClusterEndpoint
	var customEndpoint rds.DBClusterEndpoint
	readerResourceName := "aws_rds_cluster_endpoint.reader"
	defaultResourceName := "aws_rds_cluster_endpoint.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterEndpointConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterEndpointExists(ctx, readerResourceName, &customReaderEndpoint),
					testAccCheckClusterEndpointAttributes(&customReaderEndpoint),
					testAccCheckClusterEndpointExists(ctx, defaultResourceName, &customEndpoint),
					testAccCheckClusterEndpointAttributes(&customEndpoint),
					acctest.MatchResourceAttrRegionalARN(readerResourceName, "arn", "rds", regexp.MustCompile(`cluster-endpoint:.+`)),
					resource.TestCheckResourceAttrSet(readerResourceName, "endpoint"),
					acctest.MatchResourceAttrRegionalARN(defaultResourceName, "arn", "rds", regexp.MustCompile(`cluster-endpoint:.+`)),
					resource.TestCheckResourceAttrSet(defaultResourceName, "endpoint"),
					resource.TestCheckResourceAttr(defaultResourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(readerResourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      "aws_rds_cluster_endpoint.reader",
				ImportState:       true,
				ImportStateVerify: true,
			},

			{
				ResourceName:      "aws_rds_cluster_endpoint.default",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRDSClusterEndpoint_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rInt := sdkacctest.RandInt()
	var customReaderEndpoint rds.DBClusterEndpoint
	resourceName := "aws_rds_cluster_endpoint.reader"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterEndpointConfig_tags1(rInt, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterEndpointExists(ctx, resourceName, &customReaderEndpoint),
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
				Config: testAccClusterEndpointConfig_tags2(rInt, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterEndpointExists(ctx, resourceName, &customReaderEndpoint),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccClusterEndpointConfig_tags1(rInt, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterEndpointExists(ctx, resourceName, &customReaderEndpoint),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckClusterEndpointAttributes(v *rds.DBClusterEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(v.Endpoint) == "" {
			return fmt.Errorf("empty endpoint domain")
		}

		if aws.StringValue(v.CustomEndpointType) != "READER" &&
			aws.StringValue(v.CustomEndpointType) != "ANY" {
			return fmt.Errorf("Incorrect endpoint type: expected: READER or ANY, got: %s", aws.StringValue(v.CustomEndpointType))
		}

		if len(v.StaticMembers) == 0 && len(v.ExcludedMembers) == 0 {
			return fmt.Errorf("Empty members")
		}

		for _, m := range aws.StringValueSlice(v.StaticMembers) {
			if !strings.HasPrefix(m, "tf-aurora-cluster-instance") {
				return fmt.Errorf("Incorrect StaticMember Cluster Instance Identifier prefix:\nexpected: %s\ngot: %s", "tf-aurora-cluster-instance", m)
			}
		}

		for _, m := range aws.StringValueSlice(v.ExcludedMembers) {
			if !strings.HasPrefix(m, "tf-aurora-cluster-instance") {
				return fmt.Errorf("Incorrect ExcludeMember Cluster Instance Identifier prefix:\nexpected: %s\ngot: %s", "tf-aurora-cluster-instance", m)
			}
		}

		return nil
	}
}

func testAccCheckClusterEndpointDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rds_cluster_endpoint" {
				continue
			}

			_, err := tfrds.FindDBClusterEndpointByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RDS Cluster Endpoint %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClusterEndpointExists(ctx context.Context, n string, v *rds.DBClusterEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No RDS Cluster Endpoint ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn(ctx)

		output, err := tfrds.FindDBClusterEndpointByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccClusterEndpointBaseConfig(n int) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.default.engine
  engine_version             = aws_rds_cluster.default.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster" "default" {
  cluster_identifier = "tf-aurora-cluster-%[1]d"
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]
  database_name                   = "mydb"
  master_username                 = "foo"
  master_password                 = "mustbeeightcharaters"
  db_cluster_parameter_group_name = "default.aurora5.6"
  skip_final_snapshot             = true
}

resource "aws_rds_cluster_instance" "test1" {
  apply_immediately  = true
  cluster_identifier = aws_rds_cluster.default.id
  identifier         = "tf-aurora-cluster-instance-test1-%[1]d"
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
}

resource "aws_rds_cluster_instance" "test2" {
  apply_immediately  = true
  cluster_identifier = aws_rds_cluster.default.id
  identifier         = "tf-aurora-cluster-instance-test2-%[1]d"
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
}
`, n))
}

func testAccClusterEndpointConfig_basic(n int) string {
	return acctest.ConfigCompose(
		testAccClusterEndpointBaseConfig(n),
		fmt.Sprintf(`
resource "aws_rds_cluster_endpoint" "reader" {
  cluster_identifier          = aws_rds_cluster.default.id
  cluster_endpoint_identifier = "reader-%[1]d"
  custom_endpoint_type        = "READER"

  static_members = [aws_rds_cluster_instance.test2.id]
}

resource "aws_rds_cluster_endpoint" "default" {
  cluster_identifier          = aws_rds_cluster.default.id
  cluster_endpoint_identifier = "default-%[1]d"
  custom_endpoint_type        = "ANY"

  excluded_members = [aws_rds_cluster_instance.test2.id]
}
`, n))
}

func testAccClusterEndpointConfig_tags1(n int, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccClusterEndpointBaseConfig(n),
		fmt.Sprintf(`
resource "aws_rds_cluster_endpoint" "reader" {
  cluster_identifier          = aws_rds_cluster.default.id
  cluster_endpoint_identifier = "reader-%[1]d"
  custom_endpoint_type        = "READER"

  static_members = [aws_rds_cluster_instance.test2.id]

  tags = {
    %[2]q = %[3]q
  }
}
`, n, tagKey1, tagValue1))
}

func testAccClusterEndpointConfig_tags2(n int, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccClusterEndpointBaseConfig(n),
		fmt.Sprintf(`
resource "aws_rds_cluster_endpoint" "reader" {
  cluster_identifier          = aws_rds_cluster.default.id
  cluster_endpoint_identifier = "reader-%[1]d"
  custom_endpoint_type        = "READER"

  static_members = [aws_rds_cluster_instance.test2.id]

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, n, tagKey1, tagValue1, tagKey2, tagValue2))
}
