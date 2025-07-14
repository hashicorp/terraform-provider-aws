// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"strings"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"

	"github.com/hashicorp/terraform-provider-aws/names"

	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"

	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
)

type autonomousVMClusterResourceTest struct {
	exaInfraDisplayNamePrefix            string
	odbNetDisplayNamePrefix              string
	autonomousVmClusterDisplayNamePrefix string
}

var autonomousVMClusterTest = autonomousVMClusterResourceTest{
	exaInfraDisplayNamePrefix:            "Ofake-exa",
	odbNetDisplayNamePrefix:              "odb-net",
	autonomousVmClusterDisplayNamePrefix: "Ofake-avmc",
}

func TestAccODBCloudAutonomousVmClusterCreationBasic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cloudAVMC odbtypes.CloudAutonomousVmCluster
	avmcDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterTest.autonomousVmClusterDisplayNamePrefix)
	odbDisplayNamePrefix := sdkacctest.RandomWithPrefix(autonomousVMClusterTest.odbNetDisplayNamePrefix)
	exaDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterTest.exaInfraDisplayNamePrefix)
	resourceName := "aws_odb_cloud_autonomous_vm_cluster.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			//acctest.PreCheckPartitionHasService(t, names.ODBServiceID)
			autonomousVMClusterTest.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             autonomousVMClusterTest.testAccCheckCloudAutonomousVmClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: autonomousVMClusterTest.avmcWithMandatoryParamsOnly(autonomousVMClusterTest.exaInfra(exaDisplayName), autonomousVMClusterTest.odbNetwork(odbDisplayNamePrefix), avmcDisplayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					autonomousVMClusterTest.checkCloudAutonomousVmClusterExists(ctx, resourceName, &cloudAVMC),
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

func TestAccODBCloudAutonomousVmClusterTagging(t *testing.T) {
	fmt.Println("Update tags test")
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var avmc1, avmc2 odbtypes.CloudAutonomousVmCluster
	avmcDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterTest.autonomousVmClusterDisplayNamePrefix)
	odbDisplayNamePrefix := sdkacctest.RandomWithPrefix(autonomousVMClusterTest.odbNetDisplayNamePrefix)
	exaDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterTest.exaInfraDisplayNamePrefix)
	resourceName := "aws_odb_cloud_autonomous_vm_cluster.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			//acctest.PreCheckPartitionHasService(t, names.ODBServiceID)
			//testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             autonomousVMClusterTest.testAccCheckCloudAutonomousVmClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: autonomousVMClusterTest.avmcWithMandatoryParamsWithTag(autonomousVMClusterTest.exaInfra(exaDisplayName), autonomousVMClusterTest.odbNetwork(odbDisplayNamePrefix), avmcDisplayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					autonomousVMClusterTest.checkCloudAutonomousVmClusterExists(ctx, resourceName, &avmc1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.env", "dev"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: autonomousVMClusterTest.avmcWithMandatoryParamsOnly(autonomousVMClusterTest.exaInfra(exaDisplayName), autonomousVMClusterTest.odbNetwork(odbDisplayNamePrefix), avmcDisplayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					autonomousVMClusterTest.checkCloudAutonomousVmClusterExists(ctx, resourceName, &avmc2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.env", "dev"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						if strings.Compare(*(avmc1.CloudAutonomousVmClusterId), *(avmc2.CloudAutonomousVmClusterId)) != 0 {
							return errors.New("Shouldn't create a new autonomous vm cluster")
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccODBCloudAutonomousVmCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cloudautonomousvmcluster odbtypes.CloudAutonomousVmCluster
	avmcDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterTest.autonomousVmClusterDisplayNamePrefix)
	odbDisplayNamePrefix := sdkacctest.RandomWithPrefix(autonomousVMClusterTest.odbNetDisplayNamePrefix)
	exaDisplayName := sdkacctest.RandomWithPrefix(autonomousVMClusterTest.exaInfraDisplayNamePrefix)
	resourceName := "aws_odb_cloud_autonomous_vm_cluster.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			//acctest.PreCheckPartitionHasService(t, names.ODBServiceID)
			autonomousVMClusterTest.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             autonomousVMClusterTest.testAccCheckCloudAutonomousVmClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: autonomousVMClusterTest.avmcWithMandatoryParamsOnly(autonomousVMClusterTest.exaInfra(exaDisplayName), autonomousVMClusterTest.odbNetwork(odbDisplayNamePrefix), avmcDisplayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					autonomousVMClusterTest.checkCloudAutonomousVmClusterExists(ctx, resourceName, &cloudautonomousvmcluster),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfodb.ResourceCloudAutonomousVMCluster, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func (autonomousVMClusterResourceTest) testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

	input := &odb.ListCloudAutonomousVmClustersInput{}

	_, err := conn.ListCloudAutonomousVmClusters(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
func (autonomousVMClusterResourceTest) testAccCheckCloudAutonomousVmClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_cloud_autonomous_vm_cluster" {
				continue
			}

			_, err := tfodb.FindCloudAutonomousVmClusterByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameCloudAutonomousVmCluster, rs.Primary.ID, err)
			}

			return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameCloudAutonomousVmCluster, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func (autonomousVMClusterResourceTest) checkCloudAutonomousVmClusterExists(ctx context.Context, name string, cloudAutonomousVMCluster *odbtypes.CloudAutonomousVmCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameCloudAutonomousVmCluster, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameCloudAutonomousVmCluster, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
		fmt.Println("")
		resp, err := autonomousVMClusterResourceTest{}.findVMC(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameCloudAutonomousVmCluster, rs.Primary.ID, err)
		}

		*cloudAutonomousVMCluster = *resp

		return nil
	}
}

func (autonomousVMClusterResourceTest) findVMC(ctx context.Context, conn *odb.Client, id string) (*odbtypes.CloudAutonomousVmCluster, error) {
	input := odb.GetCloudAutonomousVmClusterInput{
		CloudAutonomousVmClusterId: aws.String(id),
	}
	out, err := conn.GetCloudAutonomousVmCluster(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}
		return nil, err
	}

	if out == nil || out.CloudAutonomousVmCluster == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out.CloudAutonomousVmCluster, nil
}

/*func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

	input := &odb.ListCloudAutonomousVmClustersInput{}

	_, err := conn.ListCloudAutonomousVmClusters(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}*/

/*func testAccCheckCloudAutonomousVmClusterNotRecreated(before, after *odb.DescribeCloudAutonomousVmClusterResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.CloudAutonomousVmClusterId), aws.ToString(after.CloudAutonomousVmClusterId); before != after {
			return create.Error(names.ODB, create.ErrActionCheckingNotRecreated, tfodb.ResNameCloudAutonomousVmCluster, aws.ToString(before.CloudAutonomousVmClusterId), errors.New("recreated"))
		}

		return nil
	}
}*/

// exa_rkmma5b09a
func testAccCloudAutonomousVmClusterConfig_basic_with_db_servers(rName string) string {
	res := fmt.Sprintf(`


resource "aws_odb_cloud_autonomous_vm_cluster" "test" {
  display_name             = %[1]q
  cloud_exadata_infrastructure_id              = "exa_ky7jabi90t"
  autonomous_data_storage_size_in_tbs          = 5
  is_mtls_enabled_vm_cluster                   = false
  license_model                                = "LICENSE_INCLUDED"
  memory_per_oracle_compute_unit_in_gbs        = 2
  odb_network_id                               = "odbnet_fjey4b8oth"
  total_container_databases                    = 1
  cpu_core_count_per_node                      = 40
  db_servers								   = ["dbs_7zuqmh4045","dbs_o1p7nape7g"]	
  tags = {
    "env"= "dev"
  }




}
`, rName)
	fmt.Println(res)
	return res
}

func (autonomousVMClusterResourceTest) avmcWithMandatoryParamsOnly(exaInfra, odbNetwork, avmcDisplayName string) string {
	res := fmt.Sprintf(`
%s

%s

resource "aws_odb_cloud_autonomous_vm_cluster" "test" {
  	display_name             				= %[1]q
    cloud_exadata_infrastructure_id         = "exa_ky7jabi90t"
  	odb_network_id                          = "odbnet_fjey4b8oth"
  	autonomous_data_storage_size_in_tbs     = 5
  	memory_per_oracle_compute_unit_in_gbs   = 2
  	total_container_databases               = 1
  	cpu_core_count_per_node                 = 4

}
`, exaInfra, odbNetwork, avmcDisplayName)

	return res
}

func (autonomousVMClusterResourceTest) avmcWithMandatoryParamsWithTag(exaInfra, odbNetwork, avmcDisplayName string) string {
	res := fmt.Sprintf(`
%s

%s

resource "aws_odb_cloud_autonomous_vm_cluster" "test" {
  	display_name             				= %[1]q
    cloud_exadata_infrastructure_id         = "exa_ky7jabi90t"
  	odb_network_id                          = "odbnet_fjey4b8oth"
  	autonomous_data_storage_size_in_tbs     = 5
  	memory_per_oracle_compute_unit_in_gbs   = 2
  	total_container_databases               = 1
  	cpu_core_count_per_node                 = 4
 	tags = {
    	"env"= "dev"
  	}

}
`, exaInfra, odbNetwork, avmcDisplayName)

	return res
}

func (autonomousVMClusterResourceTest) exaInfra(exaDisplayName string) string {
	resource := fmt.Sprintf(`
resource "aws_odb_cloud_exadata_infrastructure" "test" {
  display_name          = "%[1]s"
  shape             	= "Exadata.X9M"
  storage_count      	= 3
  compute_count         = 2
  availability_zone_id 	= "use1-az6"
  customer_contacts_to_send_to_oci = ["abc@example.com"]
  
}
`, exaDisplayName)

	return resource
}
func (autonomousVMClusterResourceTest) odbNetwork(odbNetDisplayName string) string {
	return fmt.Sprintf(`
	resource "aws_odb_network" "test" {
  		display_name          = %[1]q
  		availability_zone_id = "use1-az6"
  		client_subnet_cidr   = "10.2.0.0/24"
  		backup_subnet_cidr   = "10.2.1.0/24"
  		s3_access = "DISABLED"
  		zero_etl_access = "DISABLED"
	}
`, odbNetDisplayName)

}
