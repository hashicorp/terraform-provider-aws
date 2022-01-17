package groundstation_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Allieway/terraform-provider-aws/internal/acctest"
	"github.com/Allieway/terraform-provider-aws/internal/conns"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/groundstation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform/helper/schema"
)

var testAccProvider *schema.Provider

func testAccMissionProfile(t *testing.T) {
	resourceName := "aws_groundstation_mission_profile.test"

	var randArns []string
	var randDurationSeconds []string
	var randStrings []string

	for i := 1; i <= 10; i++ {
		randArn := arn.ARN{
			Partition: testAccProvider.Meta().(*conns.AWSClient).partition,
			Service:   "groundstation",
			Region:    testAccProvider.Meta().(*conns.AWSClient).region,
			AccountID: testAccProvider.Meta().(*conns.AWSClient).accountid,
			Resource:  fmt.Sprintf("%s-1234", acctest.RandString(10)),
		}
		randStrings = append(randStrings, acctest.RandString(10))
		randDurationSeconds = append(randDurationSeconds, fmt.Sprintf("%d", acctest.RandIntRange(1, 21600)))
		randArns = append(randArns, randArn.String())
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, groundstation.ErrCodeResourceNotFoundException, ""),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckMissionProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroundStationMissionProfileConfig1(randDurationSeconds[0], randArns[0]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMissionProfileExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "minimum_viable_contact_duration_seconds", randDurationSeconds[0]),
					resource.TestCheckResourceAttr(resourceName, "tracking_config_arn", randArns[0]),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGroundStationMissionProfileConfig2(randDurationSeconds[1], randDurationSeconds[2], randDurationSeconds[3], randArns[1], randArns[2], randArns[3], randStrings[0], randStrings[1]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMissionProfileExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "minimum_viable_contact_duration_seconds", randDurationSeconds[1]),
					resource.TestCheckResourceAttr(resourceName, "contact_pre_pass_duration_seconds", randDurationSeconds[2]),
					resource.TestCheckResourceAttr(resourceName, "contact_post_pass_duration_seconds", randDurationSeconds[3]),
					resource.TestCheckResourceAttr(resourceName, "tracking_config_arn", randArns[1]),
					resource.TestCheckResourceAttr(resourceName, "dataflow_edges.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dataflow_edges.0.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "dataflow_edges.0.0", randArns[2]),
					resource.TestCheckResourceAttr(resourceName, "dataflow_edges.0.1", randArns[3]),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("tags.%s", randStrings[0]), randStrings[1]),
				),
			},
			{
				Config: testAccGroundStationMissionProfileConfig3(randDurationSeconds[4], randDurationSeconds[5], randDurationSeconds[6], randArns[4], randArns[5], randArns[6], randArns[7], randArns[8], randStrings[2], randStrings[3], randStrings[4], randStrings[5], randStrings[6], randStrings[7], randStrings[8], randStrings[9]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMissionProfileExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "minimum_viable_contact_duration_seconds", randDurationSeconds[4]),
					resource.TestCheckResourceAttr(resourceName, "contact_pre_pass_duration_seconds", randDurationSeconds[5]),
					resource.TestCheckResourceAttr(resourceName, "contact_post_pass_duration_seconds", randDurationSeconds[6]),
					resource.TestCheckResourceAttr(resourceName, "tracking_config_arn", randArns[4]),
					resource.TestCheckResourceAttr(resourceName, "dataflow_edges.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "dataflow_edges.0.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "dataflow_edges.0.0", randArns[5]),
					resource.TestCheckResourceAttr(resourceName, "dataflow_edges.0.1", randArns[6]),
					resource.TestCheckResourceAttr(resourceName, "dataflow_edges.1.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "dataflow_edges.1.0", randArns[7]),
					resource.TestCheckResourceAttr(resourceName, "dataflow_edges.1.1", randArns[8]),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("tags.%s", randStrings[2]), randStrings[3]),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("tags.%s", randStrings[4]), randStrings[5]),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("tags.%s", randStrings[6]), randStrings[7]),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("tags.%s", randStrings[8]), randStrings[9]),
				),
			},
		},
	})
}

func testAccCheckMissionProfileDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).groundstationconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_groundstation_mission_profile" {
			continue
		}

		_, err := conn.GetMissionProfile(&groundstation.GetMissionProfileInput{
			MissionProfileId: aws.String(rs.Primary.ID),
		})

		if isAWSErr(err, groundstation.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Mission Profile %q still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckMissionProfileExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Mission Profile ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).groundstationconn

		_, err := conn.GetMissionProfile(&groundstation.GetMissionProfileInput{
			MissionProfileId: aws.String(rs.Primary.ID),
		})

		if isAWSErr(err, groundstation.ErrCodeResourceNotFoundException, "") {
			return fmt.Errorf("Mission Profile %q not found", rs.Primary.ID)
		}

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccGroundStationMissionProfileConfig1(minViableContactDurationSeconds string, trackingConfigArn string) string {
	return fmt.Sprintf(
		`
resource "aws_groundstation_mission_profile" "test" {
	  name                 = "tf-test-profile-%d"
	  minimum_viable_contact_duration_seconds = %s
	  tracking_config_arn = %q
}
`,
		time.Now().Unix(),
		minViableContactDurationSeconds,
		trackingConfigArn,
	)
}

func testAccGroundStationMissionProfileConfig2(minViableContactDurationSeconds string, contactPrePassDurationSeconds string, contactPostPassDurationSeconds string, trackingConfigArn string, dataflowEdge1Arn string, dataflowEdge2Arn string, tagKey string, tagValue string) string {
	return fmt.Sprintf(
		`
resource "aws_groundstation_mission_profile" "test" {
	  name                 = "tf-test-profile-%d"
	  minimum_viable_contact_duration_seconds = %s
	  contact_pre_pass_duration_seconds = %s
	  contact_post_pass_duration_seconds = %s
	  tracking_config_arn = %q
	  dataflow_edges = [
		  [
			  %q,
			  %q
		  ]
	  ]	
	  tags = {
		  %s = %s
	  }
}
`,
		time.Now().Unix(),
		minViableContactDurationSeconds,
		contactPrePassDurationSeconds,
		contactPostPassDurationSeconds,
		trackingConfigArn,
		dataflowEdge1Arn,
		dataflowEdge2Arn,
		tagKey,
		tagValue,
	)
}

func testAccGroundStationMissionProfileConfig3(minViableContactDurationSeconds string, contactPrePassDurationSeconds string, contactPostPassDurationSeconds string, trackingConfigArn string, dataflowEdge1Arn string, dataflowEdge2Arn string, dataflowEdge3Arn string, dataflowEdge4Arn string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string, tagKey3 string, tagValue3 string, tagKey4 string, tagValue4 string) string {
	return fmt.Sprintf(
		`
resource "aws_groundstation_mission_profile" "test" {
	  name                 = "tf-test-profile-%d"
	  minimum_viable_contact_duration_seconds = %s
	  contact_pre_pass_duration_seconds = %s
	  contact_post_pass_duration_seconds = %s
	  tracking_config_arn = %q
	  dataflow_edges = [
		  [
			  %q,
			  %q
		  ],
		  [
			  %q,
			  %q
		  ]
	  ]	
	  tags = {
		  %s = %s
		  %s = %s
		  %s = %s
		  %s = %s
	  }
}
`,
		time.Now().Unix(),
		minViableContactDurationSeconds,
		contactPrePassDurationSeconds,
		contactPostPassDurationSeconds,
		trackingConfigArn,
		dataflowEdge1Arn,
		dataflowEdge2Arn,
		dataflowEdge3Arn,
		dataflowEdge4Arn,
		tagKey1,
		tagValue1,
		tagKey2,
		tagValue2,
		tagKey3,
		tagValue3,
		tagKey4,
		tagValue4,
	)
}

func isAWSErr(err error, code string, message string) bool {
	if err, ok := err.(awserr.Error); ok && err.Code() == code && err.Message() == message {
		return true
	}
	return false
}
