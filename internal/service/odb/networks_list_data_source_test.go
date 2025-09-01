//Copyright (c) 2025, Oracle and/or its affiliates. All rights reserved.

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"testing"

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

func TestAccListOdbNetworksDataSource(t *testing.T) {
	ctx := acctest.Context(t)
	var networkListTest = odbNetworksListTestDS{}
	var output odb.ListOdbNetworksOutput

	dataSourceName := "data.aws_odb_networks_list.test"
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
						resource.TestCheckResourceAttr(dataSourceName, "aws_odb_networks_list.#", strconv.Itoa(len(output.OdbNetworks)))
						i := 0
						for i < len(output.OdbNetworks) {
							key := fmt.Sprintf("aws_odb_networks_list.%q.id", i)
							resource.TestMatchResourceAttr(dataSourceName, key, regexp.MustCompile(pattern))
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
	config := fmt.Sprintf(`


data "aws_odb_cloud_autonomous_vm_clusters_list" "test" {

}
`)
	return config
}

func (odbNetworksListTestDS) count(ctx context.Context, name string, list *odb.ListOdbNetworksOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.DSNameNetworksList, name, errors.New("not found"))
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
		resp, err := conn.ListOdbNetworks(ctx, &odb.ListOdbNetworksInput{})
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.DSNameNetworksList, rs.Primary.ID, err)
		}
		list.OdbNetworks = resp.OdbNetworks

		return nil
	}
}
func (odbNetworksListTestDS) testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

	input := &odb.ListOdbNetworksInput{}

	_, err := conn.ListOdbNetworks(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
