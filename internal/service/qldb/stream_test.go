package qldb

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/qldb"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccAWSQLDBStream_basic(t *testing.T) {
	var qldbCluster qldb.DescribeJournalKinesisStreamOutput

	ledgerName := fmt.Sprintf("test-ledger-%s", sdkacctest.RandString(10))
	streamName := fmt.Sprintf("test-stream-%s", sdkacctest.RandString(10))
	kinesisStreamName := fmt.Sprintf("test-kinesis-stream-%s", sdkacctest.RandString(10))
	roleName := fmt.Sprintf("test-role-%s", sdkacctest.RandString(10))

	resourceName := "aws_qldb_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(qldb.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, qldb.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSQLDBStreamCancel,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQLDBStreamDependenciesConfig(ledgerName, kinesisStreamName, roleName, streamName),
			},
			{
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSQLDBStreamExists(resourceName, &qldbCluster),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "qldb", regexp.MustCompile(`stream/.+`)),
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

func testAccCheckAWSQLDBStreamCancel(s *terraform.State) error {
	return testAccCheckAWSQLDBStreamCancelWithProvider(s, acctest.Provider)
}

func testAccCheckAWSQLDBStreamCancelWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*conns.AWSClient).QLDBConn

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
		if tfawserr.ErrMessageContains(err, qldb.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckAWSQLDBStreamExists(n string, v *qldb.DescribeJournalKinesisStreamOutput) resource.TestCheckFunc {
	log.Printf("Checking for QLDB stream's existence... %s", n)
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			log.Printf("Not found: %s", n)
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			log.Printf("No QLDB Stream ID is set")
			return fmt.Errorf("No QLDB Stream ID is set")
		}

		ledgerName, ok := rs.Primary.Attributes["ledger_name"]
		if !ok {
			log.Printf("No ledger name value has been set")
			return fmt.Errorf("No ledger name value has been set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QLDBConn
		resp, err := conn.DescribeJournalKinesisStream(&qldb.DescribeJournalKinesisStreamInput{
			LedgerName: aws.String(ledgerName),
			StreamId:   aws.String(rs.Primary.ID),
		})

		if err != nil {
			log.Printf("Error: %s", err.Error())
			return err
		}

		if *resp.Stream.StreamId == rs.Primary.ID {
			*v = *resp
			return nil
		}

		log.Printf("QLDB Stream (%s) not found", rs.Primary.ID)
		return fmt.Errorf("QLDB Stream (%s) not found", rs.Primary.ID)
	}
}

func testAccAWSQLDBStreamDependenciesConfig(rLedgerName, rKinesisStreamName, rRoleName, rStreamName string) string {
	return fmt.Sprintf(`
resource "aws_qldb_ledger" "test" {
	name                = "%s"
	deletion_protection = false
}

resource "aws_kinesis_stream" "test" {
	name             = "%s"
	shard_count      = 1
	retention_period = 24
}

resource "aws_iam_role" "test" {
	name = "%s"

	assume_role_policy = jsonencode({
		Version = "2012-10-17"
		Statement = [
			{
				Action = "sts:AssumeRole"
				Effect = "Allow"
				Sid    = ""
				Principal = {
				Service = "qldb.amazonaws.com"
				}
			},
		]
	})

	inline_policy {
		name = "test-qldb-policy"
		policy = jsonencode({
		Version = "2012-10-17"
		Statement = [
			{
				Action = [
					"kinesis:PutRecord*",
					"kinesis:DescribeStream",
					"kinesis:ListShards",
				]
				Effect   = "Allow"
				Resource = aws_kinesis_stream.test.arn
			},
		]
		})
	}
}

resource "time_sleep" "wait_seconds" {
	depends_on = [
		aws_iam_role.tf_test,
	]

	create_duration = "10s"
}
  
resource "aws_qldb_stream" "test" {
	stream_name          = "%s"
	ledger_name          = aws_qldb_ledger.test.id
	inclusive_start_time = "2021-01-01T00:00:00Z"

	role_arn = aws_iam_role.test.arn

	kinesis_configuration = {
		aggregation_enabled = false
		stream_arn          = aws_kinesis_stream.test.arn
	}
}
`, rLedgerName, rKinesisStreamName, rRoleName, rStreamName)
}
