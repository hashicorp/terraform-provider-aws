// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/appmesh"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccMeshDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appmesh_mesh.test"
	dataSourceName := "data.aws_appmesh_mesh.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMeshDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCreatedDate, dataSourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLastUpdatedDate, dataSourceName, names.AttrLastUpdatedDate),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtResourceOwner, dataSourceName, acctest.CtResourceOwner),
					resource.TestCheckResourceAttrPair(resourceName, "spec.#", dataSourceName, "spec.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.egress_filter.#", dataSourceName, "spec.0.egress_filter.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.egress_filter.0.type", dataSourceName, "spec.0.egress_filter.0.type"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func testAccMeshDataSource_meshOwner(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appmesh_mesh.test"
	dataSourceName := "data.aws_appmesh_mesh.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMeshDataSourceConfig_meshOwner(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCreatedDate, dataSourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLastUpdatedDate, dataSourceName, names.AttrLastUpdatedDate),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtResourceOwner, dataSourceName, acctest.CtResourceOwner),
					resource.TestCheckResourceAttrPair(resourceName, "spec.#", dataSourceName, "spec.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.egress_filter.#", dataSourceName, "spec.0.egress_filter.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.egress_filter.0.type", dataSourceName, "spec.0.egress_filter.0.type"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func testAccMeshDataSource_specSet(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appmesh_mesh.test"
	dataSourceName := "data.aws_appmesh_mesh.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMeshDataSourceConfig_specSet(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCreatedDate, dataSourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLastUpdatedDate, dataSourceName, names.AttrLastUpdatedDate),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtResourceOwner, dataSourceName, acctest.CtResourceOwner),
					resource.TestCheckResourceAttrPair(resourceName, "spec.#", dataSourceName, "spec.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.egress_filter.#", dataSourceName, "spec.0.egress_filter.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.egress_filter.0.type", dataSourceName, "spec.0.egress_filter.0.type"),
				),
			},
		},
	})
}

func testAccMeshDataSource_shared(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appmesh_mesh.test"
	dataSourceName := "data.aws_appmesh_mesh.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			acctest.PreCheckPartitionHasService(t, appmesh.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppMeshServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMeshDataSourceConfig_shared(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCreatedDate, dataSourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrLastUpdatedDate, dataSourceName, names.AttrLastUpdatedDate),
					resource.TestCheckResourceAttrPair(resourceName, "mesh_owner", dataSourceName, "mesh_owner"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtResourceOwner, dataSourceName, acctest.CtResourceOwner),
					resource.TestCheckResourceAttrPair(resourceName, "spec.#", dataSourceName, "spec.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.egress_filter.#", dataSourceName, "spec.0.egress_filter.#"),
					resource.TestCheckResourceAttrPair(resourceName, "spec.0.egress_filter.0.type", dataSourceName, "spec.0.egress_filter.0.type"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func testAccMeshDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

data "aws_appmesh_mesh" "test" {
  name = aws_appmesh_mesh.test.name
}
`, rName)
}

func testAccMeshDataSourceConfig_meshOwner(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_appmesh_mesh" "test" {
  name = %[1]q
}

data "aws_appmesh_mesh" "test" {
  name       = aws_appmesh_mesh.test.name
  mesh_owner = data.aws_caller_identity.current.account_id
}
`, rName)
}

func testAccMeshDataSourceConfig_specSet(rName string) string {
	return fmt.Sprintf(`
resource "aws_appmesh_mesh" "test" {
  name = %[1]q

  spec {
    egress_filter {
      type = "DROP_ALL"
    }
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_appmesh_mesh" "test" {
  name = aws_appmesh_mesh.test.name
}
`, rName)
}

func testAccMeshDataSourceConfig_shared(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_caller_identity" "source" {}

data "aws_caller_identity" "target" {
  provider = "awsalternate"
}

resource "aws_appmesh_mesh" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_ram_resource_share" "test" {
  name                      = %[1]q
  allow_external_principals = false
}

resource "aws_ram_resource_association" "test" {
  resource_arn       = aws_appmesh_mesh.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "aws_ram_principal_association" "test" {
  principal          = data.aws_caller_identity.target.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

data "aws_appmesh_mesh" "test" {
  provider = "awsalternate"

  name       = aws_appmesh_mesh.test.name
  mesh_owner = data.aws_caller_identity.source.account_id

  depends_on = [aws_ram_resource_association.test, aws_ram_principal_association.test]
}
`, rName))
}
