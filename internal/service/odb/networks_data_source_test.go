// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type odbNetworksListTestDS struct {
}

func TestAccODBListNetworksDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var networkListTest = odbNetworksListTestDS{}
	var output odb.ListOdbNetworksOutput

	dataSourceName := "data.aws_odb_networks.test"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			networkListTest.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: networkListTest.basic(),
				Check: resource.ComposeAggregateTestCheckFunc(

					resource.ComposeTestCheckFunc(func(s *terraform.State) error {
						pattern := `^odbnet_`
						networkListTest.count(ctx, dataSourceName, &output)
						resource.TestCheckResourceAttr(dataSourceName, "aws_odb_networks.#", strconv.Itoa(len(output.OdbNetworks)))
						i := 0
						for i < len(output.OdbNetworks) {
							key := fmt.Sprintf("aws_odb_networks.%q.id", i)
							resource.TestMatchResourceAttr(dataSourceName, key, regexache.MustCompile(pattern))
						}
						return nil
					},
					),
				),
			},
		},
	})
}

func (odbNetworksListTestDS) basic() string {
	return `data "aws_odb_networks" "test" {}`
}

func (odbNetworksListTestDS) count(ctx context.Context, name string, list *odb.ListOdbNetworksOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.DSNameNetworksList, name, errors.New("not found"))
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
		resp, err := tfodb.ListOracleDBNetworks(ctx, conn)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.DSNameNetworksList, rs.Primary.ID, err)
		}
		list.OdbNetworks = resp.OdbNetworks

		return nil
	}
}
func (odbNetworksListTestDS) testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
	input := odb.ListOdbNetworksInput{}
	_, err := conn.ListOdbNetworks(ctx, &input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
