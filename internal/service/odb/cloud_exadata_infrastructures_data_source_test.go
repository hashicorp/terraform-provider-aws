// Copyright IBM Corp. 2014, 2025
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

type listExaInfraTest struct {
}

func TestAccODBListCloudExadataInfrastructuresDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var listExaInfraDSTest = listExaInfraTest{}
	var infraList odb.ListCloudExadataInfrastructuresOutput
	dataSourceName := "data.aws_odb_cloud_exadata_infrastructures.test"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			listExaInfraDSTest.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: listExaInfraDSTest.basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.ComposeTestCheckFunc(func(s *terraform.State) error {
						listExaInfraDSTest.countExaInfrastructures(ctx, dataSourceName, &infraList)
						resource.TestCheckResourceAttr(dataSourceName, "cloud_exadata_infrastructures.#", strconv.Itoa(len(infraList.CloudExadataInfrastructures)))
						return nil
					},
					),
				),
			},
		},
	})
}

func (listExaInfraTest) basic() string {
	return `data "aws_odb_cloud_exadata_infrastructures" "test" {}`
}

func (listExaInfraTest) countExaInfrastructures(ctx context.Context, name string, listOfInfra *odb.ListCloudExadataInfrastructuresOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.DSNameCloudExadataInfrastructuresList, name, errors.New("not found"))
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
		resp, err := tfodb.ListCloudExadataInfrastructures(ctx, conn)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.DSNameCloudExadataInfrastructuresList, rs.Primary.ID, err)
		}
		listOfInfra.CloudExadataInfrastructures = resp.CloudExadataInfrastructures
		return nil
	}
}
func (listExaInfraTest) testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
	input := odb.ListCloudExadataInfrastructuresInput{}
	_, err := conn.ListCloudExadataInfrastructures(ctx, &input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
