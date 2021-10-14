package gamelift_test

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const testAccGameliftBuildPrefix = "tf_acc_build_"

func TestAccGameLiftBuild_basic(t *testing.T) {
	var conf gamelift.Build

	rString := sdkacctest.RandString(8)
	resourceName := "aws_gamelift_build.test"

	buildName := fmt.Sprintf("%s_%s", testAccGameliftBuildPrefix, rString)
	uBuildName := fmt.Sprintf("%s_updated_%s", testAccGameliftBuildPrefix, rString)

	region := acctest.Region()
	g, err := testAccSampleGame(region)

	if tfresource.NotFound(err) {
		t.Skip(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	loc := g.Location
	bucketName := *loc.Bucket
	roleArn := *loc.RoleArn
	key := *loc.Key

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, gamelift.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBuildDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildBasicConfig(buildName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", buildName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`build/build-.+`)),
					resource.TestCheckResourceAttr(resourceName, "operating_system", "WINDOWS_2012"),
					resource.TestCheckResourceAttr(resourceName, "storage_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.key", key),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.role_arn", roleArn),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccBuildBasicConfig(uBuildName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", uBuildName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`build/build-.+`)),
					resource.TestCheckResourceAttr(resourceName, "operating_system", "WINDOWS_2012"),
					resource.TestCheckResourceAttr(resourceName, "storage_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.key", key),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.role_arn", roleArn),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccGameLiftBuild_tags(t *testing.T) {
	var conf gamelift.Build

	rString := sdkacctest.RandString(8)
	resourceName := "aws_gamelift_build.test"

	buildName := fmt.Sprintf("%s_%s", testAccGameliftBuildPrefix, rString)
	region := acctest.Region()
	g, err := testAccSampleGame(region)

	if tfresource.NotFound(err) {
		t.Skip(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	loc := g.Location
	bucketName := *loc.Bucket
	roleArn := *loc.RoleArn
	key := *loc.Key

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, gamelift.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBuildDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildBasicTags1Config(buildName, bucketName, key, roleArn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccBuildBasicTags2Config(buildName, bucketName, key, roleArn, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccBuildBasicTags1Config(buildName, bucketName, key, roleArn, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccGameLiftBuild_disappears(t *testing.T) {
	var conf gamelift.Build

	rString := sdkacctest.RandString(8)
	resourceName := "aws_gamelift_build.test"

	buildName := fmt.Sprintf("%s_%s", testAccGameliftBuildPrefix, rString)

	region := acctest.Region()
	g, err := testAccSampleGame(region)

	if tfresource.NotFound(err) {
		t.Skip(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	loc := g.Location
	bucketName := *loc.Bucket
	roleArn := *loc.RoleArn
	key := *loc.Key

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, gamelift.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBuildDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBuildBasicConfig(buildName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBuildExists(resourceName, &conf),
					testAccCheckBuildDisappears(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBuildExists(n string, res *gamelift.Build) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Gamelift Build ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn

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

func testAccCheckBuildDisappears(res *gamelift.Build) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn

		input := &gamelift.DeleteBuildInput{BuildId: res.BuildId}

		_, err := conn.DeleteBuild(input)
		return err
	}
}

func testAccCheckBuildDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn

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
		if tfawserr.ErrMessageContains(err, gamelift.ErrCodeNotFoundException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn

	input := &gamelift.ListBuildsInput{}

	_, err := conn.ListBuilds(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccBuildBasicConfig(buildName, bucketName, key, roleArn string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_build" "test" {
  name             = "%s"
  operating_system = "WINDOWS_2012"

  storage_location {
    bucket   = "%s"
    key      = "%s"
    role_arn = "%s"
  }
}
`, buildName, bucketName, key, roleArn)
}

func testAccBuildBasicTags1Config(buildName, bucketName, key, roleArn, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_build" "test" {
  name             = %[1]q
  operating_system = "WINDOWS_2012"

  storage_location {
    bucket   = %[2]q
    key      = %[3]q
    role_arn = %[4]q
  }

  tags = {
    %[5]q = %[6]q
  }
}
`, buildName, bucketName, key, roleArn, tagKey1, tagValue1)
}

func testAccBuildBasicTags2Config(buildName, bucketName, key, roleArn, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_build" "test" {
  name             = %[1]q
  operating_system = "WINDOWS_2012"

  storage_location {
    bucket   = %[2]q
    key      = %[3]q
    role_arn = %[4]q
  }

  tags = {
    %[5]q = %[6]q
    %[7]q = %[8]q
  }
}
`, buildName, bucketName, key, roleArn, tagKey1, tagValue1, tagKey2, tagValue2)
}
