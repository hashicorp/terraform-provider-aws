package mediapackage_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediapackage"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccMediaPackageChannel_basic(t *testing.T) {
	resourceName := "aws_media_package_channel.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mediapackage.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_basic(sdkacctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "mediapackage", regexp.MustCompile(`channels/.+`)),
					resource.TestMatchResourceAttr(resourceName, "hls_ingest.0.ingest_endpoints.0.password", regexp.MustCompile("^[0-9a-f]*$")),
					resource.TestMatchResourceAttr(resourceName, "hls_ingest.0.ingest_endpoints.0.url", regexp.MustCompile("^https://")),
					resource.TestMatchResourceAttr(resourceName, "hls_ingest.0.ingest_endpoints.0.username", regexp.MustCompile("^[0-9a-f]*$")),
					resource.TestMatchResourceAttr(resourceName, "hls_ingest.0.ingest_endpoints.1.password", regexp.MustCompile("^[0-9a-f]*$")),
					resource.TestMatchResourceAttr(resourceName, "hls_ingest.0.ingest_endpoints.1.url", regexp.MustCompile("^https://")),
					resource.TestMatchResourceAttr(resourceName, "hls_ingest.0.ingest_endpoints.1.username", regexp.MustCompile("^[0-9a-f]*$")),
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

func TestAccMediaPackageChannel_description(t *testing.T) {
	resourceName := "aws_media_package_channel.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mediapackage.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccChannelConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccMediaPackageChannel_tags(t *testing.T) {
	resourceName := "aws_media_package_channel.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mediapackage.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccChannelConfig_tags(rName, "Environment", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccChannelConfig_tags(rName, "Environment", "test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "test1"),
				),
			},
			{
				Config: testAccChannelConfig_tags(rName, "Update", "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckChannelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Update", "true"),
				),
			},
		},
	})
}

func testAccCheckChannelDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaPackageConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_media_package_channel" {
			continue
		}

		input := &mediapackage.DescribeChannelInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeChannel(input)
		if err == nil {
			return fmt.Errorf("MediaPackage Channel (%s) not deleted", rs.Primary.ID)
		}

		if !tfawserr.ErrCodeEquals(err, mediapackage.ErrCodeNotFoundException) {
			return err
		}
	}

	return nil
}

func testAccCheckChannelExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaPackageConn

		input := &mediapackage.DescribeChannelInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeChannel(input)

		return err
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaPackageConn

	input := &mediapackage.ListChannelsInput{}

	_, err := conn.ListChannels(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccChannelConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_package_channel" "test" {
  channel_id = "tf_mediachannel_%s"
}
`, rName)
}

func testAccChannelConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_media_package_channel" "test" {
  channel_id  = %q
  description = %q
}
`, rName, description)
}

func testAccChannelConfig_tags(rName, key, value string) string {
	return fmt.Sprintf(`
resource "aws_media_package_channel" "test" {
  channel_id = "%[1]s"

  tags = {
    Name = "%[1]s"

    %[2]s = "%[3]s"
  }
}
`, rName, key, value)
}
