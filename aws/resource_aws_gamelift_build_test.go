package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const testAccGameliftPrefix = "tf_acc_build_"

func init() {
	resource.AddTestSweepers("aws_gamelift_build", &resource.Sweeper{
		Name: "aws_gamelift_build",
		F:    testSweepGameliftBuilds,
	})
}

func testSweepGameliftBuilds(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).gameliftconn

	resp, err := conn.ListBuilds(&gamelift.ListBuildsInput{})
	if err != nil {
		return fmt.Errorf("Error listing Gamelift Builds: %s", err)
	}

	if len(resp.Builds) == 0 {
		log.Print("[DEBUG] No Gamelift Builds to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d Gamelift Builds", len(resp.Builds))

	for _, build := range resp.Builds {
		if !strings.HasPrefix(*build.Name, testAccGameliftPrefix) {
			continue
		}

		log.Printf("[INFO] Deleting Gamelift Build %q", *build.BuildId)
		_, err := conn.DeleteBuild(&gamelift.DeleteBuildInput{
			BuildId: build.BuildId,
		})
		if err != nil {
			return fmt.Errorf("Error deleting Gamelift Build (%s): %s",
				*build.BuildId, err)
		}
	}

	return nil
}

func TestAccAWSGameliftBuild_basic(t *testing.T) {
	var conf gamelift.Build

	rString := acctest.RandString(8)

	buildName := fmt.Sprintf("%s_%s", testAccGameliftPrefix, rString)
	uBuildName := fmt.Sprintf("%s_updated_%s", testAccGameliftPrefix, rString)

	region := testAccGetRegion()
	loc, err := testAccAWSGameliftSampleGameLocation(region)
	if err != nil {
		t.Fatal(err)
	}

	bucketName := *loc.Bucket
	roleArn := *loc.RoleArn
	key := *loc.Key

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftBuildDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftBuildBasicConfig(buildName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftBuildExists("aws_gamelift_build.test", &conf),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "name", buildName),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "operating_system", "WINDOWS_2012"),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "storage_location.#", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "storage_location.0.bucket", bucketName),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "storage_location.0.key", key),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "storage_location.0.role_arn", roleArn),
				),
			},
			{
				Config: testAccAWSGameliftBuildBasicConfig(uBuildName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftBuildExists("aws_gamelift_build.test", &conf),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "name", uBuildName),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "operating_system", "WINDOWS_2012"),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "storage_location.#", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "storage_location.0.bucket", bucketName),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "storage_location.0.key", key),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "storage_location.0.role_arn", roleArn),
				),
			},
		},
	})
}

func testAccCheckAWSGameliftBuildExists(n string, res *gamelift.Build) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Gamelift Build ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).gameliftconn

		req := &gamelift.DescribeBuildInput{
			BuildId: aws.String(rs.Primary.ID),
		}
		out, err := conn.DescribeBuild(req)
		if err != nil {
			return err
		}

		b := out.Build

		if *b.BuildId != rs.Primary.ID {
			return fmt.Errorf("Gamelift Build not found")
		}

		*res = *b

		return nil
	}
}

func testAccCheckAWSGameliftBuildDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).gameliftconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_gamelift_build" {
			continue
		}

		req := gamelift.DescribeBuildInput{
			BuildId: aws.String(rs.Primary.ID),
		}
		out, err := conn.DescribeBuild(&req)
		if err == nil {
			if *out.Build.BuildId == rs.Primary.ID {
				return fmt.Errorf("Gamelift Build still exists")
			}
		}
		if isAWSErr(err, gamelift.ErrCodeNotFoundException, "") {
			return nil
		}

		return err
	}

	return nil
}

// Location found from CloudTrail event after finishing tutorial
// e.g. https://us-west-2.console.aws.amazon.com/gamelift/home?region=us-west-2#/r/fleets/sample
func testAccAWSGameliftSampleGameLocation(region string) (*gamelift.S3Location, error) {
	version := "v1.2.0.0"
	accId, err := testAccGameliftAccountIdByRegion(region)
	if err != nil {
		return nil, err
	}
	bucket := fmt.Sprintf("gamelift-sample-builds-prod-%s", region)
	key := fmt.Sprintf("%s/server/sample_build_%s", version, version)
	roleArn := fmt.Sprintf("arn:aws:iam::%s:role/sample-build-upload-role-%s", accId, region)

	return &gamelift.S3Location{
		Bucket:  aws.String(bucket),
		Key:     aws.String(key),
		RoleArn: aws.String(roleArn),
	}, nil
}

// Account ID found from CloudTrail event (role ARN) after finishing tutorial in given region
func testAccGameliftAccountIdByRegion(region string) (string, error) {
	m := map[string]string{
		"ap-northeast-1": "120069834884",
		"ap-northeast-2": "805673136642",
		"ap-south-1":     "134975661615",
		"ap-southeast-1": "077577004113",
		"ap-southeast-2": "112188327105",
		"ca-central-1":   "800535022691",
		"eu-central-1":   "797584052317",
		"eu-west-1":      "319803218673",
		"eu-west-2":      "937342764187",
		"sa-east-1":      "028872612690",
		"us-east-1":      "783764748367",
		"us-east-2":      "415729564621",
		"us-west-1":      "715879310420",
		"us-west-2":      "741061592171",
	}

	if accId, ok := m[region]; ok {
		return accId, nil
	}

	return "", fmt.Errorf("Account ID not found for region %q", region)
}

func testAccAWSGameliftBuildBasicConfig(buildName, bucketName, key, roleArn string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_build" "test" {
  name = "%s"
  operating_system = "WINDOWS_2012"
  storage_location {
    bucket = "%s"
    key = "%s"
    role_arn = "%s"
  }
}
`, buildName, bucketName, key, roleArn)
}
