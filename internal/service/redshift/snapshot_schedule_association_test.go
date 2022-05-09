package redshift_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
)

func TestAccRedshiftSnapshotScheduleAssociation_basic(t *testing.T) {
	rName := sdkacctest.RandString(8)
	resourceName := "aws_redshift_snapshot_schedule_association.default"
	snapshotScheduleResourceName := "aws_redshift_snapshot_schedule.default"
	clusterResourceName := "aws_redshift_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotScheduleAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleAssociationConfig(rName, "rate(12 hours)"),
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

func testAccCheckSnapshotScheduleAssociationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_snapshot_schedule_association" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn
		clusterIdentifier, scheduleIdentifier, err := tfredshift.SnapshotScheduleAssociationParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		resp, err := conn.DescribeSnapshotSchedules(&redshift.DescribeSnapshotSchedulesInput{
			ScheduleIdentifier: aws.String(scheduleIdentifier),
			ClusterIdentifier:  aws.String(clusterIdentifier),
		})

		if err != nil {
			return err
		}

		if resp != nil && len(resp.SnapshotSchedules) > 0 {
			return fmt.Errorf("Redshift Cluster (%s) Snapshot Schedule (%s) Association still exist", clusterIdentifier, scheduleIdentifier)
		}

		return err
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

		clusterIdentifier, scheduleIdentifier, err := tfredshift.SnapshotScheduleAssociationParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn
		resp, err := conn.DescribeSnapshotSchedules(&redshift.DescribeSnapshotSchedulesInput{
			ScheduleIdentifier: aws.String(scheduleIdentifier),
			ClusterIdentifier:  aws.String(clusterIdentifier),
		})

		if err != nil {
			return err
		}

		if len(resp.SnapshotSchedules) != 0 {
			return nil
		}

		return fmt.Errorf("Redshift Cluster (%s) Snapshot Schedule (%s) Association not found", clusterIdentifier, scheduleIdentifier)
	}
}

func testAccSnapshotScheduleAssociationConfig(rName, definition string) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(rName), testAccSnapshotScheduleConfig(rName, definition), `
resource "aws_redshift_snapshot_schedule_association" "default" {
  schedule_identifier = aws_redshift_snapshot_schedule.default.id
  cluster_identifier  = aws_redshift_cluster.test.id
}
`)
}
