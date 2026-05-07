// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package neptunegraph_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/neptunegraph"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfneptunegraph "github.com/hashicorp/terraform-provider-aws/internal/service/neptunegraph"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNeptuneGraphPrivateGraphEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var privategraphendpoint neptunegraph.GetPrivateGraphEndpointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_neptunegraph_private_graph_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneGraphServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateGraphEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateGraphEndpointConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPrivateGraphEndpointExists(ctx, t, resourceName, &privategraphendpoint),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVPCEndpointID),
					resource.TestCheckResourceAttrSet(resourceName, "private_graph_endpoint_identifier"),
					resource.TestCheckResourceAttrPair(resourceName, "graph_identifier", "aws_neptunegraph_graph.test", names.AttrID),
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

func TestAccNeptuneGraphPrivateGraphEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var endpoint neptunegraph.GetPrivateGraphEndpointOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_neptunegraph_private_graph_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneGraphServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPrivateGraphEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateGraphEndpointConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPrivateGraphEndpointExists(ctx, t, resourceName, &endpoint),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfneptunegraph.ResourcePrivateGraphEndpoint, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPrivateGraphEndpointDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NeptuneGraphClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_neptunegraph_private_graph_endpoint" {
				continue
			}

			_, err := tfneptunegraph.FindPrivateGraphEndpointByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return err
			}
			return fmt.Errorf("Neptune Analytics Private Graph Endpoint %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPrivateGraphEndpointExists(ctx context.Context, t *testing.T, n string, v *neptunegraph.GetPrivateGraphEndpointOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NeptuneGraphClient(ctx)

		output, err := tfneptunegraph.FindPrivateGraphEndpointByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPrivateGraphEndpointConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_neptunegraph_graph" "test" {
  graph_name          = %[1]q
  provisioned_memory  = 16
  deletion_protection = false
}

resource "aws_neptunegraph_private_graph_endpoint" "test" {
  graph_identifier = aws_neptunegraph_graph.test.id
  vpc_id           = aws_vpc.test.id
  subnet_ids       = aws_subnet.test[*].id
}
`, rName))
}
