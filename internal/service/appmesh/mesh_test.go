// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/appmesh"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappmesh "github.com/hashicorp/terraform-provider-aws/internal/service/appmesh"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccMesh_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var mesh appmesh.MeshData
	resourceName := "aws_appmesh_mesh.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMeshDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMeshConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMeshExists(ctx, resourceName, &mesh),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "appmesh", regexp.MustCompile(`mesh/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					acctest.CheckResourceAttrAccountID(resourceName, "mesh_owner"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "resource_owner"),
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
	var mesh appmesh.MeshData
	resourceName := "aws_appmesh_mesh.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMeshDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMeshConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMeshExists(ctx, resourceName, &mesh),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappmesh.ResourceMesh(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccMesh_egressFilter(t *testing.T) {
	ctx := acctest.Context(t)
	var mesh appmesh.MeshData
	resourceName := "aws_appmesh_mesh.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMeshDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMeshConfig_egressFilter(rName, "ALLOW_ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMeshExists(ctx, resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, "spec.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.egress_filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spec.0.egress_filter.0.type", "ALLOW_ALL"),
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
					testAccCheckMeshExists(ctx, resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "spec.0.egress_filter.0.type", "DROP_ALL"),
				),
			},
			{
				PlanOnly: true,
				Config:   testAccMeshConfig_basic(rName),
			},
		},
	})
}

func testAccMesh_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var mesh appmesh.MeshData
	resourceName := "aws_appmesh_mesh.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, appmesh.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMeshDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccMeshConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMeshExists(ctx, resourceName, &mesh),
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
				Config: testAccMeshConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMeshExists(ctx, resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccMeshConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMeshExists(ctx, resourceName, &mesh),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckMeshDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appmesh_mesh" {
				continue
			}

			_, err := tfappmesh.FindMeshByTwoPartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["mesh_owner"])

			if tfresource.NotFound(err) {
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

func testAccCheckMeshExists(ctx context.Context, n string, v *appmesh.MeshData) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppMeshConn(ctx)

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

func testAccMeshConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccMeshConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
