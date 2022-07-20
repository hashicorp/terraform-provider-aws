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

func TestAccRedshiftSnapshotSchedule_basic(t *testing.T) {
	var v redshift.SnapshotSchedule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_snapshot_schedule.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleConfig_basic(rName, "rate(12 hours)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "definitions.#", "1"),
				),
			},
			{
				Config: testAccSnapshotScheduleConfig_basic(rName, "cron(30 12 *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "definitions.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy",
				},
			},
		},
	})
}

func TestAccRedshiftSnapshotSchedule_withMultipleDefinition(t *testing.T) {
	var v redshift.SnapshotSchedule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_snapshot_schedule.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleConfig_multipleDefinition(rName, "cron(30 12 *)", "cron(15 6 *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "definitions.#", "2"),
				),
			},
			{
				Config: testAccSnapshotScheduleConfig_multipleDefinition(rName, "cron(30 8 *)", "cron(15 10 *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "definitions.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy",
				},
			},
		},
	})

}

func TestAccRedshiftSnapshotSchedule_withIdentifierPrefix(t *testing.T) {
	var v redshift.SnapshotSchedule
	resourceName := "aws_redshift_snapshot_schedule.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleConfig_identifierPrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleExists(resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"identifier_prefix",
					"force_destroy",
				},
			},
		},
	})
}

func TestAccRedshiftSnapshotSchedule_withDescription(t *testing.T) {
	var v redshift.SnapshotSchedule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_snapshot_schedule.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleConfig_description(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "description", "Test Schedule"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy",
				},
			},
		},
	})
}

func TestAccRedshiftSnapshotSchedule_withTags(t *testing.T) {
	var v redshift.SnapshotSchedule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_snapshot_schedule.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
				),
			},
			{
				Config: testAccSnapshotScheduleConfig_tagsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar2"),
					resource.TestCheckResourceAttr(resourceName, "tags.good", "bad"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy",
				},
			},
		},
	})
}

func TestAccRedshiftSnapshotSchedule_withForceDestroy(t *testing.T) {
	var snapshotSchedule redshift.SnapshotSchedule
	var cluster redshift.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_redshift_snapshot_schedule.default"
	clusterResourceName := "aws_redshift_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSnapshotScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotScheduleConfig_forceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotScheduleExists(resourceName, &snapshotSchedule),
					testAccCheckClusterExists(clusterResourceName, &cluster),
					testAccCheckSnapshotScheduleCreateSnapshotScheduleAssociation(&cluster, &snapshotSchedule),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy",
				},
			},
		},
	})
}

func testAccCheckSnapshotScheduleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_snapshot_schedule" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn
		resp, err := conn.DescribeSnapshotSchedules(&redshift.DescribeSnapshotSchedulesInput{
			ScheduleIdentifier: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if len(resp.SnapshotSchedules) != 0 {
				for _, s := range resp.SnapshotSchedules {
					if *s.ScheduleIdentifier == rs.Primary.ID {
						return fmt.Errorf("Redshift Cluster Snapshot Schedule %s still exists", rs.Primary.ID)
					}
				}
			}
		}

		return err
	}

	return nil
}

func testAccCheckSnapshotScheduleExists(n string, v *redshift.SnapshotSchedule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Cluster Snapshot Schedule ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn
		resp, err := conn.DescribeSnapshotSchedules(&redshift.DescribeSnapshotSchedulesInput{
			ScheduleIdentifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		for _, s := range resp.SnapshotSchedules {
			if *s.ScheduleIdentifier == rs.Primary.ID {
				*v = *s
				return nil
			}
		}

		return fmt.Errorf("Redshift Snapshot Schedule (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckSnapshotScheduleCreateSnapshotScheduleAssociation(cluster *redshift.Cluster, snapshotSchedule *redshift.SnapshotSchedule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

		if _, err := conn.ModifyClusterSnapshotSchedule(&redshift.ModifyClusterSnapshotScheduleInput{
			ClusterIdentifier:    cluster.ClusterIdentifier,
			ScheduleIdentifier:   snapshotSchedule.ScheduleIdentifier,
			DisassociateSchedule: aws.Bool(false),
		}); err != nil {
			return fmt.Errorf("Error associate Redshift Cluster and Snapshot Schedule: %s", err)
		}

		id := fmt.Sprintf("%s/%s", aws.StringValue(cluster.ClusterIdentifier), aws.StringValue(snapshotSchedule.ScheduleIdentifier))
		if _, err := tfredshift.WaitScheduleAssociationActive(conn, id); err != nil {
			return err
		}

		return nil
	}
}

const testAccSnapshotScheduleConfig_identifierPrefix = `
resource "aws_redshift_snapshot_schedule" "default" {
  identifier_prefix = "tf-acc-test"
  definitions = [
    "rate(12 hours)",
  ]
}
`

func testAccSnapshotScheduleConfig_basic(rName, definition string) string {
	return fmt.Sprintf(`
resource "aws_redshift_snapshot_schedule" "default" {
  identifier = %[1]q
  definitions = [
    "%[2]s",
  ]
}
`, rName, definition)
}

func testAccSnapshotScheduleConfig_multipleDefinition(rName, definition1, definition2 string) string {
	return fmt.Sprintf(`
resource "aws_redshift_snapshot_schedule" "default" {
  identifier = %[1]q
  definitions = [
    "%[2]s",
    "%[3]s",
  ]
}
`, rName, definition1, definition2)
}

func testAccSnapshotScheduleConfig_description(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_snapshot_schedule" "default" {
  identifier  = %[1]q
  description = "Test Schedule"
  definitions = [
    "rate(12 hours)",
  ]
}
`, rName)
}

func testAccSnapshotScheduleConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_snapshot_schedule" "default" {
  identifier = %[1]q
  definitions = [
    "rate(12 hours)",
  ]

  tags = {
    foo  = "bar"
    fizz = "buzz"
  }
}
`, rName)
}

func testAccSnapshotScheduleConfig_tagsUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_snapshot_schedule" "default" {
  identifier = %[1]q
  definitions = [
    "rate(12 hours)",
  ]

  tags = {
    foo  = "bar2"
    good = "bad"
  }
}
`, rName)
}

func testAccSnapshotScheduleConfig_forceDestroy(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_basic(rName), fmt.Sprintf(`
resource "aws_redshift_snapshot_schedule" "default" {
  identifier  = %[1]q
  description = "Test Schedule"
  definitions = [
    "rate(12 hours)",
  ]
  force_destroy = true
}
`, rName))
}
