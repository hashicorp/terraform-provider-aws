// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSClusterEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var customReaderEndpoint types.DBClusterEndpoint
	var customEndpoint types.DBClusterEndpoint
	readerResourceName := "aws_rds_cluster_endpoint.reader"
	defaultResourceName := "aws_rds_cluster_endpoint.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterEndpointConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterEndpointExists(ctx, readerResourceName, &customReaderEndpoint),
					testAccCheckClusterEndpointExists(ctx, defaultResourceName, &customEndpoint),
					acctest.MatchResourceAttrRegionalARN(readerResourceName, names.AttrARN, "rds", regexache.MustCompile(`cluster-endpoint:.+`)),
					resource.TestCheckResourceAttrSet(readerResourceName, names.AttrEndpoint),
					acctest.MatchResourceAttrRegionalARN(defaultResourceName, names.AttrARN, "rds", regexache.MustCompile(`cluster-endpoint:.+`)),
					resource.TestCheckResourceAttrSet(defaultResourceName, names.AttrEndpoint),
					resource.TestCheckResourceAttr(defaultResourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(readerResourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      readerResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},

			{
				ResourceName:      defaultResourceName,
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

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var customReaderEndpoint types.DBClusterEndpoint
	resourceName := "aws_rds_cluster_endpoint.reader"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterEndpointConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterEndpointExists(ctx, resourceName, &customReaderEndpoint),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterEndpointConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterEndpointExists(ctx, resourceName, &customReaderEndpoint),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccClusterEndpointConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterEndpointExists(ctx, resourceName, &customReaderEndpoint),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckClusterEndpointDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

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

func testAccCheckClusterEndpointExists(ctx context.Context, n string, v *types.DBClusterEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSClient(ctx)

		output, err := tfrds.FindDBClusterEndpointByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccClusterEndpointConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                     = aws_rds_cluster.default.engine
  engine_version             = aws_rds_cluster.default.engine_version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t3.medium"]
}

resource "aws_rds_cluster" "default" {
  cluster_identifier = %[1]q
  availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1],
    data.aws_availability_zones.available.names[2]
  ]

  database_name       = "test"
  engine              = %[2]q
  master_username     = "tfacctest"
  master_password     = "avoid-plaintext-passwords"
  skip_final_snapshot = true
}

resource "aws_rds_cluster_instance" "test1" {
  apply_immediately  = true
  cluster_identifier = aws_rds_cluster.default.id
  identifier         = "%[1]s-1"
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
  engine             = aws_rds_cluster.default.engine
}

resource "aws_rds_cluster_instance" "test2" {
  apply_immediately  = true
  cluster_identifier = aws_rds_cluster.default.id
  identifier         = "%[1]s-2"
  instance_class     = data.aws_rds_orderable_db_instance.test.instance_class
  engine             = aws_rds_cluster.default.engine
}
`, rName, tfrds.ClusterEngineAuroraMySQL))
}

func testAccClusterEndpointConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterEndpointConfig_base(rName), fmt.Sprintf(`
resource "aws_rds_cluster_endpoint" "reader" {
  cluster_identifier          = aws_rds_cluster.default.id
  cluster_endpoint_identifier = "%[1]s-reader"
  custom_endpoint_type        = "READER"

  static_members = [aws_rds_cluster_instance.test2.id]
}

resource "aws_rds_cluster_endpoint" "default" {
  cluster_identifier          = aws_rds_cluster.default.id
  cluster_endpoint_identifier = "%[1]s-default"
  custom_endpoint_type        = "ANY"

  excluded_members = [aws_rds_cluster_instance.test2.id]
}
`, rName))
}

func testAccClusterEndpointConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccClusterEndpointConfig_base(rName), fmt.Sprintf(`
resource "aws_rds_cluster_endpoint" "reader" {
  cluster_identifier          = aws_rds_cluster.default.id
  cluster_endpoint_identifier = "%[1]s-reader"
  custom_endpoint_type        = "READER"

  static_members = [aws_rds_cluster_instance.test2.id]

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccClusterEndpointConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccClusterEndpointConfig_base(rName), fmt.Sprintf(`
resource "aws_rds_cluster_endpoint" "reader" {
  cluster_identifier          = aws_rds_cluster.default.id
  cluster_endpoint_identifier = "%[1]s-reader"
  custom_endpoint_type        = "READER"

  static_members = [aws_rds_cluster_instance.test2.id]

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
