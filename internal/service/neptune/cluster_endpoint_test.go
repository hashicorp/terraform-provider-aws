// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package neptune_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/neptune/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfneptune "github.com/hashicorp/terraform-provider-aws/internal/service/neptune"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNeptuneClusterEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster awstypes.DBClusterEndpoint
	rName := acctest.RandomWithPrefix(t, "tf-acc")
	resourceName := "aws_neptune_cluster_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterEndpointExists(ctx, t, resourceName, &dbCluster),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "rds", regexache.MustCompile(`cluster-endpoint:.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEndpointType, "READER"),
					resource.TestCheckResourceAttr(resourceName, "cluster_endpoint_identifier", rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClusterIdentifier, "aws_neptune_cluster.test", names.AttrClusterIdentifier),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "static_members.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "excluded_members.#", "0"),
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

func TestAccNeptuneClusterEndpoint_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.DBClusterEndpoint
	rName := acctest.RandomWithPrefix(t, "tf-acc")
	resourceName := "aws_neptune_cluster_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterEndpointConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterEndpointExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
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
					testAccCheckClusterEndpointExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccClusterEndpointConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterEndpointExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccNeptuneClusterEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster awstypes.DBClusterEndpoint
	rName := acctest.RandomWithPrefix(t, "tf-acc")
	resourceName := "aws_neptune_cluster_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterEndpointExists(ctx, t, resourceName, &dbCluster),
					acctest.CheckSDKResourceDisappears(ctx, t, tfneptune.ResourceClusterEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNeptuneClusterEndpoint_Disappears_cluster(t *testing.T) {
	ctx := acctest.Context(t)
	var dbCluster awstypes.DBClusterEndpoint
	rName := acctest.RandomWithPrefix(t, "tf-acc")
	resourceName := "aws_neptune_cluster_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterEndpointExists(ctx, t, resourceName, &dbCluster),
					acctest.CheckSDKResourceDisappears(ctx, t, tfneptune.ResourceCluster(), "aws_neptune_cluster.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckClusterEndpointDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NeptuneClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_neptune_cluster_endpoint" {
				continue
			}

			_, err := tfneptune.FindClusterEndpointByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrClusterIdentifier], rs.Primary.Attributes["cluster_endpoint_identifier"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Neptune Cluster Endpoint %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClusterEndpointExists(ctx context.Context, t *testing.T, n string, v *awstypes.DBClusterEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NeptuneClient(ctx)

		output, err := tfneptune.FindClusterEndpointByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrClusterIdentifier], rs.Primary.Attributes["cluster_endpoint_identifier"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccClusterEndpointConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
locals {
  availability_zone_names = slice(data.aws_availability_zones.available.names, 0, min(3, length(data.aws_availability_zones.available.names)))
}

resource "aws_neptune_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zones                   = local.availability_zone_names
  engine                               = "neptune"
  neptune_cluster_parameter_group_name = "default.neptune1.3"
  skip_final_snapshot                  = true
}
`, rName))
}

func testAccClusterEndpointConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterEndpointConfig_base(rName), fmt.Sprintf(`
resource "aws_neptune_cluster_endpoint" "test" {
  cluster_identifier          = aws_neptune_cluster.test.cluster_identifier
  cluster_endpoint_identifier = %[1]q
  endpoint_type               = "READER"
}
`, rName))
}

func testAccClusterEndpointConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccClusterEndpointConfig_base(rName), fmt.Sprintf(`
resource "aws_neptune_cluster_endpoint" "test" {
  cluster_identifier          = aws_neptune_cluster.test.cluster_identifier
  cluster_endpoint_identifier = %[1]q
  endpoint_type               = "READER"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccClusterEndpointConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccClusterEndpointConfig_base(rName), fmt.Sprintf(`
resource "aws_neptune_cluster_endpoint" "test" {
  cluster_identifier          = aws_neptune_cluster.test.cluster_identifier
  cluster_endpoint_identifier = %[1]q
  endpoint_type               = "READER"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
