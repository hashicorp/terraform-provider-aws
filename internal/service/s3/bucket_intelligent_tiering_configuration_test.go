package s3_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccS3BucketIntelligentTieringConfiguration_basic(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketIntelligentTieringConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", bucketResourceName, "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "tiering.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tiering.0.access_tier", "DEEP_ARCHIVE_ACCESS"),
					resource.TestCheckResourceAttr(resourceName, "tiering.0.days", "180"),
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

func TestAccS3BucketIntelligentTieringConfiguration_disappears(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketIntelligentTieringConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketIntelligentTieringConfigurationExists(resourceName, &itc),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucketIntelligentTieringConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketIntelligentTieringConfiguration_Filter(t *testing.T) {
	var itc s3.IntelligentTieringConfiguration
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3_bucket_intelligent_tiering_configuration.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBucketIntelligentTieringConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketIntelligentTieringConfigurationConfig_filterPrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", bucketResourceName, "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", "p1/"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", "Disabled"),
					resource.TestCheckResourceAttr(resourceName, "tiering.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tiering.*", map[string]string{
						"access_tier": "DEEP_ARCHIVE_ACCESS",
						"days":        "180",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBucketIntelligentTieringConfigurationConfig_filterPrefixAndTag(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", bucketResourceName, "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", "p2/"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.Environment", "test"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "tiering.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tiering.*", map[string]string{
						"access_tier": "ARCHIVE_ACCESS",
						"days":        "90",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tiering.*", map[string]string{
						"access_tier": "DEEP_ARCHIVE_ACCESS",
						"days":        "360",
					}),
				),
			},
			{
				Config: testAccBucketIntelligentTieringConfigurationConfig_filterTag(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", bucketResourceName, "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.Environment", "acctest"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", "Disabled"),
					resource.TestCheckResourceAttr(resourceName, "tiering.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tiering.*", map[string]string{
						"access_tier": "DEEP_ARCHIVE_ACCESS",
						"days":        "270",
					}),
				),
			},
			{
				Config: testAccBucketIntelligentTieringConfigurationConfig_filterPrefixAndTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", bucketResourceName, "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", "p3/"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.Environment1", "test"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.Environment2", "acctest"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "tiering.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tiering.*", map[string]string{
						"access_tier": "ARCHIVE_ACCESS",
						"days":        "365",
					}),
				),
			},
			{
				Config: testAccBucketIntelligentTieringConfigurationConfig_filterTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketIntelligentTieringConfigurationExists(resourceName, &itc),
					resource.TestCheckResourceAttrPair(resourceName, "bucket", bucketResourceName, "bucket"),
					resource.TestCheckResourceAttr(resourceName, "filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.Environment1", "acctest"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.tags.Environment2", "test"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", "Enabled"),
					resource.TestCheckResourceAttr(resourceName, "tiering.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "tiering.*", map[string]string{
						"access_tier": "DEEP_ARCHIVE_ACCESS",
						"days":        "365",
					}),
				),
			},
		},
	})
}

func testAccBucketIntelligentTieringConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  tiering {
    access_tier = "DEEP_ARCHIVE_ACCESS"
    days        = 180
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}

func testAccBucketIntelligentTieringConfigurationConfig_filterPrefix(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  status = "Disabled"

  filter {
    prefix = "p1/"
  }

  tiering {
    access_tier = "DEEP_ARCHIVE_ACCESS"
    days        = 180
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}

func testAccBucketIntelligentTieringConfigurationConfig_filterPrefixAndTag(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  status = "Enabled"

  filter {
    prefix = "p2/"

    tags = {
      Environment = "test"
    }
  }

  tiering {
    access_tier = "ARCHIVE_ACCESS"
    days        = 90
  }

  tiering {
    access_tier = "DEEP_ARCHIVE_ACCESS"
    days        = 360
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}

func testAccBucketIntelligentTieringConfigurationConfig_filterTag(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  status = "Disabled"

  filter {
    tags = {
      Environment = "acctest"
    }
  }

  tiering {
    access_tier = "DEEP_ARCHIVE_ACCESS"
    days        = 270
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}

func testAccBucketIntelligentTieringConfigurationConfig_filterPrefixAndTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  filter {
    prefix = "p3/"

    tags = {
      Environment1 = "test"
      Environment2 = "acctest"
    }
  }

  tiering {
    access_tier = "ARCHIVE_ACCESS"
    days        = 365
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}

func testAccBucketIntelligentTieringConfigurationConfig_filterTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket_intelligent_tiering_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q

  filter {
    tags = {
      Environment1 = "acctest"
      Environment2 = "test"
    }
  }

  tiering {
    access_tier = "DEEP_ARCHIVE_ACCESS"
    days        = 365
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}
`, rName)
}

func testAccCheckBucketIntelligentTieringConfigurationExists(n string, v *s3.IntelligentTieringConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Intelligent-Tiering Configuration ID is set")
		}

		bucketName, configurationName, err := tfs3.BucketIntelligentTieringConfigurationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

		output, err := tfs3.FindBucketIntelligentTieringConfiguration(conn, bucketName, configurationName)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckBucketIntelligentTieringConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_intelligent_tiering_configuration" {
			continue
		}

		bucketName, configurationName, err := tfs3.BucketIntelligentTieringConfigurationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfs3.FindBucketIntelligentTieringConfiguration(conn, bucketName, configurationName)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("S3 Intelligent-Tiering Configuration %s still exists", rs.Primary.ID)
	}

	return nil

}
