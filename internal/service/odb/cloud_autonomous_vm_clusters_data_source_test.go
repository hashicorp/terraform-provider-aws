// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
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

type listAVMCListDSTest struct {
}

func TestAccODBListAutonomousVmClustersDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var avmcListTest = listAVMCListDSTest{}
	var output odb.ListCloudAutonomousVmClustersOutput

	dataSourceName := "data.aws_odb_cloud_autonomous_vm_clusters.test"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			avmcListTest.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: avmcListTest.basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.ComposeTestCheckFunc(func(s *terraform.State) error {
						avmcListTest.count(ctx, dataSourceName, &output)
						resource.TestCheckResourceAttr(dataSourceName, "cloud_autonomous_vm_clusters.#", strconv.Itoa(len(output.CloudAutonomousVmClusters)))
						return nil
					},
					),
				),
			},
		},
	})
}

func (listAVMCListDSTest) basic() string {
	return `data "aws_odb_cloud_autonomous_vm_clusters" "test" {}`
}

func (listAVMCListDSTest) count(ctx context.Context, name string, list *odb.ListCloudAutonomousVmClustersOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.DSNameCloudAutonomousVmClustersList, name, errors.New("not found"))
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
		resp, err := tfodb.ListCloudAutonomousVmClusters(ctx, conn)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.DSNameCloudAutonomousVmClustersList, rs.Primary.ID, err)
		}
		list.CloudAutonomousVmClusters = resp.CloudAutonomousVmClusters
		return nil
	}
}
func (listAVMCListDSTest) testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
	input := odb.ListCloudAutonomousVmClustersInput{}
	_, err := conn.ListCloudAutonomousVmClusters(ctx, &input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
