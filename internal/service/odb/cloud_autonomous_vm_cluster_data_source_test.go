// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccODBCloudAutonomousVmClusterDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	dataSourceName := "basicExaInfraDataSource.aws_odb_cloud_autonomous_vm_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			//acctest.PreCheckPartitionHasService(t, names.ODBEndpointID)
			//testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		//CheckDestroy:             testAccCheckCloudAutonomousVmClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: basicHardcodedAVmCluster("avmc_yhqyltpw6m"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						fmt.Println(state)
						return nil
					}),
					resource.TestCheckResourceAttr(dataSourceName, "display_name", "Real-AVMC-CONN-001"),
					//acctest.MatchResourceAttrRegionalARN(ctx, dataSourceName, names.AttrARN, "odb", regexache.MustCompile(`cloudautonomousvmcluster:.+$`)),
				),
			},
		},
	})
}

func basicHardcodedAVmCluster(id string) string {
	avmcDataSource := fmt.Sprintf(`

basicExaInfraDataSource "aws_odb_cloud_autonomous_vm_cluster" "test" {
  id             = %[1]q

}
`, id)
	fmt.Println(avmcDataSource)
	return avmcDataSource
}
