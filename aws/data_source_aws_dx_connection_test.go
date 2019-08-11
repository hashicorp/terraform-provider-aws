package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"math/rand"
	"os"
	"regexp"
	"testing"
	"time"
)

func TestAccDataSourceAwsDxConnection_Basic(t *testing.T) {
	connectionName := fmt.Sprintf("tf-dx-%s", acctest.RandString(5))
	resourceName := "aws_dx_connection.test"
	datasourceName := "data.aws_dx_connection.test"

	dxLocation, _ := testAccAwsDxConnectionLocation()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsDxConnectionConfig_NonExistent,
				ExpectError: regexp.MustCompile(`Direct Connect Connection not found`),
			},
			{
				Config: testAccDataSourceAwsDxConnectionConfig_Name(connectionName, aws.StringValue(dxLocation)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "location", resourceName, "location"),
					resource.TestCheckResourceAttrPair(datasourceName, "jumbo_frame_capable", resourceName, "jumbo_frame_capable"),
					resource.TestCheckResourceAttrPair(datasourceName, "bandwidth", resourceName, "bandwidth"),
					resource.TestCheckResourceAttr(datasourceName, "state", "requested"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsDxConnection_tags(t *testing.T) {
	connectionName := fmt.Sprintf("tf-dx-%s", acctest.RandString(5))
	resourceName := "aws_dx_connection.test"
	datasourceName := "data.aws_dx_connection.test"

	dxLocation, _ := testAccAwsDxConnectionLocation()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsDxConnectionConfig_tags(connectionName, aws.StringValue(dxLocation)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "location", resourceName, "location"),
					resource.TestCheckResourceAttrPair(datasourceName, "jumbo_frame_capable", resourceName, "jumbo_frame_capable"),
					resource.TestCheckResourceAttrPair(datasourceName, "bandwidth", resourceName, "bandwidth"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(datasourceName, "tags.Name", connectionName),
					resource.TestCheckResourceAttr(datasourceName, "tags.Location", aws.StringValue(dxLocation)),
				),
			},
		},
	})
}

func testAccDataSourceAwsDxConnectionConfig_Name(connectionName, dxLocation string) string {
	return fmt.Sprintf(`
resource "aws_dx_connection" "wrong" {
  name            = "%[1]s-wrong"
  bandwidth       = "1Gbps"
  location        = %[2]q
}
resource "aws_dx_connection" "test" {
  name            = %[1]q
  bandwidth       = "1Gbps"
  location        = %[2]q
}

data "aws_dx_connection" "test" {
  name = "${aws_dx_connection.test.name}"
}
`, connectionName, dxLocation)
}

func testAccDataSourceAwsDxConnectionConfig_tags(connectionName, dxLocation string) string {
	return fmt.Sprintf(`
resource "aws_dx_connection" "wrong-tags" {
  name            = %[1]q
  bandwidth       = "1Gbps"
  location        = %[2]q

  tags = {
    Name = "%[1]s-Wrong"
    Location = %[2]q
  }
}

resource "aws_dx_connection" "test" {
  name            = %[1]q
  bandwidth       = "1Gbps"
  location        = %[2]q
  
  tags = {
    Name = %[1]q
    Location = %[2]q
  }
}

data "aws_dx_connection" "test" {
  name = "${aws_dx_connection.test.name}"
  tags = {
    Name = %[1]q
    Location = %[2]q
  }
}
`, connectionName, dxLocation)
}

const testAccDataSourceAwsDxConnectionConfig_NonExistent = `
data "aws_dx_connection" "test" {
  name = "tf-acc-test-does-not-exist"
}
`

func testAccAwsDxConnectionLocation() (*string, error) {
	var region string

	if _, ok := os.LookupEnv("AWS_DEFAULT_REGION"); ok {
		region = os.Getenv("AWS_DEFAULT_REGION")
	} else {
		region = "us-west-2"
	}

	client, err := sharedClientForRegion(region)

	if err != nil {
		return nil, fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).dxconn
	input := &directconnect.DescribeLocationsInput{}

	resp, err := conn.DescribeLocations(input)

	if err != nil {
		fmt.Println("Error Describing DX Locations")
	}

	rand.Seed(time.Now().Unix())

	dxLocation := resp.Locations[rand.Intn(len(resp.Locations))].LocationCode

	fmt.Printf("Testing Connections in DX Location: %s", aws.StringValue(dxLocation))

	return dxLocation, nil
}
