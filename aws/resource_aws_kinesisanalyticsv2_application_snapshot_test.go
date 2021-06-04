package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kinesisanalyticsv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/atest"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kinesisanalyticsv2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
	awsprovider "github.com/terraform-providers/terraform-provider-aws/provider"
)

func TestAccAWSKinesisAnalyticsV2ApplicationSnapshot_basic(t *testing.T) {
	var v kinesisanalyticsv2.SnapshotDetails
	resourceName := "aws_kinesisanalyticsv2_application_snapshot.test"
	applicationResourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { atest.PreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		ErrorCheck:   atest.ErrorCheck(t, kinesisanalyticsv2.EndpointsID),
		Providers:    atest.Providers,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationSnapshotConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationSnapshotExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "application_name", applicationResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "application_version_id", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "snapshot_creation_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_name", rName),
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

func TestAccAWSKinesisAnalyticsV2ApplicationSnapshot_disappears(t *testing.T) {
	var v kinesisanalyticsv2.SnapshotDetails
	resourceName := "aws_kinesisanalyticsv2_application_snapshot.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { atest.PreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		ErrorCheck:   atest.ErrorCheck(t, kinesisanalyticsv2.EndpointsID),
		Providers:    atest.Providers,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationSnapshotConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationSnapshotExists(resourceName, &v),
					atest.CheckDisappears(atest.Provider, resourceAwsKinesisAnalyticsV2ApplicationSnapshot(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsV2ApplicationSnapshot_disappears_Application(t *testing.T) {
	var v kinesisanalyticsv2.SnapshotDetails
	resourceName := "aws_kinesisanalyticsv2_application_snapshot.test"
	applicationResourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { atest.PreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		ErrorCheck:   atest.ErrorCheck(t, kinesisanalyticsv2.EndpointsID),
		Providers:    atest.Providers,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationSnapshotConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationSnapshotExists(resourceName, &v),
					atest.CheckDisappears(atest.Provider, resourceAwsKinesisAnalyticsV2Application(), applicationResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckKinesisAnalyticsV2ApplicationSnapshotDestroy(s *terraform.State) error {
	conn := atest.Provider.Meta().(*awsprovider.AWSClient).KinesisAnalyticsV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kinesisanalyticsv2_application_snapshot" {
			continue
		}

		_, err := finder.SnapshotDetailsByApplicationAndSnapshotNames(conn, rs.Primary.Attributes["application_name"], rs.Primary.Attributes["snapshot_name"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Kinesis Analytics v2 Application Snapshot %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckKinesisAnalyticsV2ApplicationSnapshotExists(n string, v *kinesisanalyticsv2.SnapshotDetails) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kinesis Analytics v2 Application Snapshot ID is set")
		}

		conn := atest.Provider.Meta().(*awsprovider.AWSClient).KinesisAnalyticsV2Conn

		application, err := finder.SnapshotDetailsByApplicationAndSnapshotNames(conn, rs.Primary.Attributes["application_name"], rs.Primary.Attributes["snapshot_name"])

		if err != nil {
			return err
		}

		*v = *application

		return nil
	}
}

func testAccKinesisAnalyticsV2ApplicationSnapshotConfig(rName string) string {
	return testAccKinesisAnalyticsV2ApplicationConfigStartSnapshotableFlinkApplication(rName, "SKIP_RESTORE_FROM_SNAPSHOT", "")
}
