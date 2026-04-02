// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appmesh_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appmesh/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfappmesh "github.com/hashicorp/terraform-provider-aws/internal/service/appmesh"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccMesh_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var mesh awstypes.MeshData
	resourceName := "aws_appmesh_mesh.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppMeshEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceMeshDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMeshConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceMeshExists(ctx, t, resourceName, &mesh),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appmesh", regexache.MustCompile(`mesh/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrLastUpdatedDate),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, acctest.CtResourceOwner),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
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

func testAccMesh_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var mesh awstypes.MeshData
	resourceName := "aws_appmesh_mesh.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppMeshEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceMeshDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMeshConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceMeshExists(ctx, t, resourceName, &mesh),
					acctest.CheckSDKResourceDisappears(ctx, t, tfappmesh.ResourceMesh(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccMesh_egressFilter(t *testing.T) {
	ctx := acctest.Context(t)
	var mesh awstypes.MeshData
	resourceName := "aws_appmesh_mesh.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppMeshEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceMeshDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMeshConfig_egressFilter(rName, "ALLOW_ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceMeshExists(ctx, t, resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.egress_filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.egress_filter.0.type", "ALLOW_ALL"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMeshConfig_egressFilter(rName, "DROP_ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceMeshExists(ctx, t, resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.egress_filter.0.type", "DROP_ALL"),
				),
			},
		},
	})
}

func testAccMesh_serviceDiscovery(t *testing.T) {
	ctx := acctest.Context(t)
	var mesh awstypes.MeshData
	resourceName := "aws_appmesh_mesh.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppMeshEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceMeshDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMeshConfig_serviceDiscovery(rName, "IPv6_PREFERRED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceMeshExists(ctx, t, resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.ip_preference", "IPv6_PREFERRED"),
				),
			},
			{
				Config: testAccMeshConfig_serviceDiscovery(rName, "IPv4_PREFERRED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceMeshExists(ctx, t, resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.ip_preference", "IPv4_PREFERRED"),
				),
			},
			{
				Config: testAccMeshConfig_serviceDiscovery(rName, "IPv4_ONLY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceMeshExists(ctx, t, resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.ip_preference", "IPv4_ONLY"),
				),
			},
			{
				Config: testAccMeshConfig_serviceDiscovery(rName, "IPv6_ONLY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceMeshExists(ctx, t, resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.service_discovery.0.ip_preference", "IPv6_ONLY"),
				),
			},
		},
	})
}

func testAccCheckServiceMeshDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AppMeshClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appmesh_mesh" {
				continue
			}

			_, err := tfappmesh.FindMeshByTwoPartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["mesh_owner"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("App Mesh Service Mesh %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckServiceMeshExists(ctx context.Context, t *testing.T, n string, v *awstypes.MeshData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AppMeshClient(ctx)

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Mesh Service Mesh ID is set")
		}

		output, err := tfappmesh.FindMeshByTwoPartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["mesh_owner"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccMeshConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}
`, rName)
}

func testAccMeshConfig_egressFilter(rName, egressFilterType string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q

  spec {
    egress_filter {
      type = %[2]q
    }
  }
}
`, rName, egressFilterType)
}

func testAccMeshConfig_serviceDiscovery(rName, serviceDiscoveryIpPreference string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q

  spec {
    service_discovery {
      ip_preference = %[2]q
    }
  }
}
`, rName, serviceDiscoveryIpPreference)
}
