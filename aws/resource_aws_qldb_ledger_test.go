package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/qldb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
)

func init() {
	resource.AddTestSweepers("aws_qldb_ledger", &resource.Sweeper{
		Name: "aws_qldb_ledger",
		F:    testSweepQLDBLedgers,
	})
}

func testSweepQLDBLedgers(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).qldbconn
	req := &qldb.DescribeLedgerInput{}
	resp, err := conn.DescribeLedger(req)
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping QLDB Ledger sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing QLDB Ledgers: %s", err)
	}

	if len(aws.StringValue(resp.Name)) == 0 {
		log.Print("[DEBUG] No aws QLDB Kedgers to sweep")
		return nil
	}

	return nil
}

func TestAccAWSQLDBLedger_basic(t *testing.T) {
	var qldbCluster qldb.DescribeLedgerOutput
	rInt := acctest.RandInt()
	resourceName := "aws_qldb_ledger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSQLDBLedgerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQLDBLedgerConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSQLDBLedgerExists(resourceName, &qldbCluster),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(`^arn:[^:]+:qldb:[^:]+:\d{12}:ledger/.+`)),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("test-ledger-[0-9]+")),
					resource.TestCheckResourceAttr(resourceName, "permissions_mode", "ALLOW_ALL"),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"apply_immediately",
				},
			},
		},
	})
}

func testAccCheckAWSQLDBLedgerDestroy(s *terraform.State) error {
	err := testAccCheckAWSClusterDestroyWithProvider(s, testAccProvider)
	if err != nil {
		qldberr := err.(awserr.Error)
		if qldberr.Code() == "ResourceInUseException" {
			log.Println("Ledger is in creating state. Retrying in 10 seconds...")
			time.Sleep(time.Second * 10)
			return testAccCheckAWSQLDBLedgerDestroy(s)
		}
	}

	return err
}

func testAccCheckAWSQLDBLedgerExists(n string, v *qldb.DescribeLedgerOutput) resource.TestCheckFunc {
	return testAccCheckAWSQLDBLedgerExistsWithProvider(n, v, func() *schema.Provider { return testAccProvider })
}

func testAccCheckAWSQLDBLedgerExistsWithProvider(n string, v *qldb.DescribeLedgerOutput, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No QLDB Ledger ID is set")
		}

		provider := providerF()
		conn := provider.Meta().(*AWSClient).qldbconn
		resp, err := conn.DescribeLedger(&qldb.DescribeLedgerInput{
			Name: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.Name == rs.Primary.ID {
			*v = *resp
			return nil
		}

		return fmt.Errorf("QLDB Ledger (%s) not found", rs.Primary.ID)
	}
}

func testAccAWSQLDBLedgerConfig(n int) string {
	return fmt.Sprintf(`
resource "aws_qldb_ledger" "test" {
  name                            = "test-ledger-%d"
  deletion_protection             = false
}
`, n)
}

func TestAccAWSQLDBLedger_Tags(t *testing.T) {
	var cluster1, cluster2, cluster3 qldb.DescribeLedgerOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_qldb_ledger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQLDBLedgerConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSQLDBLedgerExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSQLDBLedgerConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSQLDBLedgerExists(resourceName, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSQLDBLedgerConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSQLDBLedgerExists(resourceName, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccAWSQLDBLedgerConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_qldb_ledger" "test" {
  name                = %[1]q
  deletion_protection = false

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSQLDBLedgerConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_qldb_ledger" "test" {
  name                = %[1]q
  deletion_protection = false

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
