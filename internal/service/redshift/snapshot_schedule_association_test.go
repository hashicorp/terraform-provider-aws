package redshift_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/redshift"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRedshiftSnapshotScheduleAssociation_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_snapshot_schedule_association.test"
	snapshotScheduleResourceName := "aws_redshift_snapshot_schedule.default"
	clusterResourceName := "aws_redshift_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotScheduleAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleAssociationConfig_basic(rName, "rate(12 hours)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "schedule_identifier", snapshotScheduleResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_identifier", clusterResourceName, "id"),
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

func TestAccRedshiftSnapshotScheduleAssociation_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_snapshot_schedule_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotScheduleAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleAssociationConfig_basic(rName, "rate(12 hours)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfredshift.ResourceSnapshotScheduleAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedshiftSnapshotScheduleAssociation_disappears_cluster(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_snapshot_schedule_association.test"
	clusterResourceName := "aws_redshift_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotScheduleAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleAssociationConfig_basic(rName, "rate(12 hours)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleAssociationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfredshift.ResourceCluster(), clusterResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSnapshotScheduleAssociationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_snapshot_schedule_association" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

		_, _, err := tfredshift.FindScheduleAssociationById(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Redshift Schedule Association %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckSnapshotScheduleAssociationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Cluster Snapshot Schedule Association ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

		_, _, err := tfredshift.FindScheduleAssociationById(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccSnapshotScheduleAssociationConfig_basic(rName, definition string) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(rName), testAccSnapshotScheduleConfig_basic(rName, definition), `
resource "aws_redshift_snapshot_schedule_association" "test" {
  schedule_identifier = aws_redshift_snapshot_schedule.default.id
  cluster_identifier  = aws_redshift_cluster.test.id
}
`)
}
