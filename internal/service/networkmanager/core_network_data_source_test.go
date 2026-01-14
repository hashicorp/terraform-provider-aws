// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package networkmanager_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkManagerCoreNetworkDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_networkmanager_core_network.test"
	resourceName := "aws_networkmanager_core_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCoreNetworkDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "core_network_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_network_id", resourceName, "global_network_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrState, resourceName, names.AttrState),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrTags, resourceName, names.AttrTags),
					acctest.MatchResourceAttrGlobalARN(ctx, dataSourceName, names.AttrARN, "networkmanager", regexache.MustCompile(`core-network/.+`)),
					resource.TestCheckResourceAttr(dataSourceName, "edges.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "edges.0.edge_location", acctest.Region()),
					resource.TestCheckResourceAttr(dataSourceName, "edges.0.asn", "64555"),
					resource.TestCheckResourceAttr(dataSourceName, "edges.0.inside_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "edges.1.edge_location", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(dataSourceName, "edges.1.asn", "64556"),
					resource.TestCheckResourceAttr(dataSourceName, "edges.1.inside_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "network_function_groups.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "segments.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "segments.0.name", "segment1"),
					resource.TestCheckResourceAttr(dataSourceName, "segments.0.edge_locations.#", "2"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "segments.0.edge_locations.*", acctest.Region()),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "segments.0.edge_locations.*", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(dataSourceName, "segments.0.shared_segments.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "segments.1.name", "segment2"),
					resource.TestCheckResourceAttr(dataSourceName, "segments.1.edge_locations.#", "1"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "segments.1.edge_locations.*", acctest.Region()),
					resource.TestCheckResourceAttr(dataSourceName, "segments.1.shared_segments.#", "0"),
				),
			},
		},
	})
}

func testAccCoreNetworkDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkmanager_global_network" "test" {
  description = %[1]q

  tags = {
    Name = %[1]q
  }
}

data "aws_networkmanager_core_network_policy_document" "test" {
  core_network_configuration {
    asn_ranges = ["64512-65534"]

    edge_locations {
      location = %[2]q
      asn      = 64555
    }

    edge_locations {
      location = %[3]q
      asn      = 64556
    }
  }

  segments {
    name           = "segment1"
    edge_locations = [%[2]q, %[3]q]
  }

  segments {
    name           = "segment2"
    edge_locations = [%[2]q]
  }
}

resource "aws_networkmanager_core_network" "test" {
  global_network_id    = aws_networkmanager_global_network.test.id
  description          = %[1]q
  create_base_policy   = true
  base_policy_document = data.aws_networkmanager_core_network_policy_document.test.json

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkmanager_core_network_policy_attachment" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  policy_document = data.aws_networkmanager_core_network_policy_document.test.json
}

data "aws_networkmanager_core_network" "test" {
  core_network_id = aws_networkmanager_core_network.test.id
  depends_on      = [aws_networkmanager_core_network_policy_attachment.test]
}
`, rName, acctest.Region(), acctest.AlternateRegion())
}
