package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/qldb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_qldb_stream", &resource.Sweeper{
		Name: "aws_qldb_stream",
		F:    testSweepQLDBStreams,
	})
}

func testSweepQLDBStreams(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).qldbconn
	input := &qldb.ListLedgersInput{}
	page, err := conn.ListLedgers(input)

	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping QLDB Stream sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error listing QLDB Ledgers for QLDB Stream Sweep: %s", err)
	}

	for _, item := range page.Ledgers {
		ledgerName := aws.StringValue(item.Name)
		err := testSweepQLDBLedgerStreams(conn, region, ledgerName)
		if err != nil {
			log.Printf("[ERROR] Failed to delete QLDB Stream for Ledger %s: %s", ledgerName, err)
			continue
		}
	}

	return nil
}

func testSweepQLDBLedgerStreams(conn *qldb.QLDB, region string, ledgerName string) error {
	input := &qldb.ListJournalKinesisStreamsForLedgerInput{
		LedgerName: aws.String(ledgerName),
	}
	page, err := conn.ListJournalKinesisStreamsForLedger(input)

	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping QLDB Stream sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error listing QLDB Streams: %s", err)
	}

	for _, item := range page.Streams {
		input := &qldb.CancelJournalKinesisStreamInput{
			LedgerName: item.LedgerName,
			StreamId:   item.StreamId,
		}

		ledgerName := aws.StringValue(item.LedgerName)
		streamID := aws.StringValue(item.StreamId)

		log.Printf("[INFO] Cancelling QLDB Stream: (%s, %s)", ledgerName, streamID)
		_, err = conn.CancelJournalKinesisStream(input)

		if err != nil {
			log.Printf("[ERROR] Failed to cancel QLDB Ledger: (%s, %s): %s", ledgerName, streamID, err)
			continue
		}

		if err := waitForQLDBStreamCancellation(conn, ledgerName, streamID); err != nil {
			log.Printf("[ERROR] Error waiting for QLDB Ledger (%s, %s) deletion: %s", ledgerName, streamID, err)
		}
	}

	return nil
}

func waitForQLDBStreamCancellation(conn *qldb.QLDB, ledgerName string, streamID string) error {
	stateConf := resource.StateChangeConf{
		Pending: []string{
			qldb.StreamStatusActive,
			qldb.StreamStatusImpaired,
		},
		Target: []string{
			qldb.StreamStatusCanceled,
			qldb.StreamStatusCompleted,
			qldb.StreamStatusFailed,
		},
		Timeout:    5 * time.Minute,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeJournalKinesisStream(&qldb.DescribeJournalKinesisStreamInput{
				LedgerName: aws.String(ledgerName),
				StreamId:   aws.String(streamID),
			})

			if isAWSErr(err, qldb.ErrCodeResourceNotFoundException, "") {
				return 1, "", nil
			}

			if err != nil {
				return nil, qldb.ErrCodeResourceInUseException, err
			}

			return resp, aws.StringValue(resp.Stream.Status), nil
		},
	}

	_, err := stateConf.WaitForState()

	return err
}

func TestAccAWSQLDBStream_basic(t *testing.T) {
	var qldbCluster qldb.DescribeJournalKinesisStreamOutput
	rInt := acctest.RandInt()
	resourceName := "aws_qldb_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(qldb.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, qldb.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSQLDBStreamCancel,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQLDBLedgerConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSQLDBStreamExists(resourceName, &qldbCluster),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "qldb", regexp.MustCompile(`stream/.+`)),
					resource.TestMatchResourceAttr(resourceName, "stream_name", regexp.MustCompile("test-stream-[0-9]+")),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

// reference: testAccCheckAWSLedgerDestroyWithProvider
func testAccCheckAWSQLDBStreamCancel(s *terraform.State) error {
	return testAccCheckAWSQLDBStreamCancelWithProvider(s, testAccProvider)
}

func testAccCheckAWSQLDBStreamCancelWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*AWSClient).qldbconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_qldb_ledger" {
			continue
		}

		// Try to find the Group
		var err error
		resp, err := conn.DescribeLedger(
			&qldb.DescribeLedgerInput{
				Name: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(aws.StringValue(resp.Name)) != 0 && aws.StringValue(resp.Name) == rs.Primary.ID {
				return fmt.Errorf("QLDB Ledger %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the cluster is already destroyed
		if isAWSErr(err, qldb.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckAWSQLDBStreamExists(n string, v *qldb.DescribeJournalKinesisStreamOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No QLDB Stream ID is set")
		}

		ledgerName, ok := rs.Primary.Attributes["ledger_name"]
		if !ok {
			return fmt.Errorf("No ledger name value has been set")
		}

		conn := testAccProvider.Meta().(*AWSClient).qldbconn
		resp, err := conn.DescribeJournalKinesisStream(&qldb.DescribeJournalKinesisStreamInput{
			LedgerName: aws.String(ledgerName),
			StreamId:   aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.Stream.StreamId == rs.Primary.ID {
			*v = *resp
			return nil
		}

		return fmt.Errorf("QLDB Stream (%s) not found", rs.Primary.ID)
	}
}

func testAccAWSQLDBStreamConfig(n int) string {
	return fmt.Sprintf(`
resource "aws_qldb_stream" "test" {
  name                = "test-stream-%d"
  deletion_protection = false
}
`, n)
}

func TestAccAWSQLDBStream_Tags(t *testing.T) {
	var cluster1 qldb.DescribeLedgerOutput
	var cluster2, cluster3, cluster4 qldb.DescribeJournalKinesisStreamOutput
	rLedgerName := acctest.RandomWithPrefix("tf-acc-test")
	rStreamName := acctest.RandomWithPrefix("tf-acc-test")

	resourceNameForQLDBLedger := "aws_qldb_ledger.test"
	resourceNameForQLDBStream := "aws_qldb_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(qldb.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, qldb.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSQLDBStreamCancel,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQLDBLedgerForStreamsConfig(rLedgerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSQLDBLedgerExists(resourceNameForQLDBLedger, &cluster1),
				),
			},
			{
				Config: testAccAWSQLDBStreamConfigTags1(rLedgerName, rStreamName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSQLDBStreamExists(resourceNameForQLDBStream, &cluster2),
					resource.TestCheckResourceAttr(resourceNameForQLDBStream, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceNameForQLDBStream, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceNameForQLDBStream,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSQLDBStreamConfigTags2(rLedgerName, rStreamName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSQLDBStreamExists(resourceNameForQLDBStream, &cluster3),
					resource.TestCheckResourceAttr(resourceNameForQLDBStream, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceNameForQLDBStream, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceNameForQLDBStream, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSQLDBStreamConfigTags1(rLedgerName, rStreamName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSQLDBStreamExists(resourceNameForQLDBStream, &cluster4),
					resource.TestCheckResourceAttr(resourceNameForQLDBStream, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceNameForQLDBStream, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccAWSQLDBLedgerForStreamsConfig(rLedgerName string) string {
	return fmt.Sprintf(`
	resource "aws_qldb_ledger" "test" {
		name                = "%s"
		deletion_protection = false
	}
	`, rLedgerName)
}

func testAccAWSQLDBStreamConfigTags1(rLedgerName, rStreamName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
	resource "aws_qldb_stream" "test" {
		stream_name          = "%[1]q"
		ledger_name          = "%[2]q"
		inclusive_start_time = "2021-01-01T00:00:00Z"
		deletion_protection  = false
	
		role_arn = "arn:aws:iam::xxxxxxxxxxxx:role/service-role/test-qldb-role"
	
		kinesis_configuration = {
			aggregation_enabled = false
			stream_arn          = "arn:aws:kinesis:us-east-1:xxxxxxxxxxxx:stream/test-kinesis-stream"
		}
	
		tags = {
			%[3]q = %[4]q
		}
	}
	`, rLedgerName, rStreamName, tagKey1, tagValue1)
}

func testAccAWSQLDBStreamConfigTags2(rLedgerName, rStreamName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_qldb_stream" "test" {
	stream_name          = "%[1]q"
	ledger_name          = "%[2]q"
	inclusive_start_time = "2021-01-01T00:00:00Z"
	deletion_protection  = false

	role_arn = "arn:aws:iam::xxxxxxxxxxxx:role/service-role/test-qldb-role"

	kinesis_configuration = {
		aggregation_enabled = false
		stream_arn          = "arn:aws:kinesis:us-east-1:xxxxxxxxxxxx:stream/test-kinesis-stream"
	}

	tags = {
		%[3]q = %[4]q
		%[5]q = %[6]q
	}
}

`, rLedgerName, rStreamName, tagKey1, tagValue1, tagKey2, tagValue2)
}
