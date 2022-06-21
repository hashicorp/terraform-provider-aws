package s3_test

// WARNING: This code is DEPRECATED and will be removed in a future release!!
// DO NOT apply fixes or enhancements to this file.
// INSTEAD, apply fixes and enhancements to "objects_data_source_test.go".

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccS3BucketObjectsDataSource_basic(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories:         acctest.ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectsDataSourceConfig_resources(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccBucketObjectsDataSourceConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectsExistsDataSource("data.aws_s3_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.#", "2"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.0", "arch/navajo/north_window"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.1", "arch/navajo/sand_dune"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectsDataSource_basicViaAccessPoint(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories:         acctest.ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectsDataSourceConfig_resourcesPlusAccessPoint(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccBucketObjectsDataSourceConfig_basicViaAccessPoint(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectsExistsDataSource("data.aws_s3_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.#", "2"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.0", "arch/navajo/north_window"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.1", "arch/navajo/sand_dune"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectsDataSource_all(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories:         acctest.ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectsDataSourceConfig_resources(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccBucketObjectsDataSourceConfig_all(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectsExistsDataSource("data.aws_s3_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.#", "7"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.0", "arch/courthouse_towers/landscape"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.1", "arch/navajo/north_window"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.2", "arch/navajo/sand_dune"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.3", "arch/partition/park_avenue"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.4", "arch/rubicon"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.5", "arch/three_gossips/broken"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.6", "arch/three_gossips/turret"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectsDataSource_prefixes(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories:         acctest.ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectsDataSourceConfig_resources(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccBucketObjectsDataSourceConfig_prefixes(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectsExistsDataSource("data.aws_s3_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.#", "1"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.0", "arch/rubicon"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "common_prefixes.#", "4"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "common_prefixes.0", "arch/courthouse_towers/"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "common_prefixes.1", "arch/navajo/"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "common_prefixes.2", "arch/partition/"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "common_prefixes.3", "arch/three_gossips/"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectsDataSource_encoded(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories:         acctest.ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectsDataSourceConfig_extraResource(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccBucketObjectsDataSourceConfig_encoded(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectsExistsDataSource("data.aws_s3_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.#", "2"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.0", "arch/ru+b+ic+on"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.1", "arch/rubicon"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectsDataSource_maxKeys(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories:         acctest.ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectsDataSourceConfig_resources(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccBucketObjectsDataSourceConfig_maxKeys(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectsExistsDataSource("data.aws_s3_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.#", "2"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.0", "arch/courthouse_towers/landscape"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.1", "arch/navajo/north_window"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectsDataSource_startAfter(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories:         acctest.ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectsDataSourceConfig_resources(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccBucketObjectsDataSourceConfig_startAfter(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectsExistsDataSource("data.aws_s3_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.#", "1"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.0", "arch/three_gossips/turret"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectsDataSource_fetchOwner(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		ProviderFactories:         acctest.ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectsDataSourceConfig_resources(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccBucketObjectsDataSourceConfig_owners(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectsExistsDataSource("data.aws_s3_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "keys.#", "2"),
					resource.TestCheckResourceAttr("data.aws_s3_objects.yesh", "owners.#", "2"),
				),
			},
		},
	})
}

func testAccBucketObjectsDataSourceConfig_resources(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "objects_bucket" {
  bucket = "tf-acc-objects-test-bucket-%d"
}

resource "aws_s3_object" "object1" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/three_gossips/turret"
  content = "Delicate"
}

resource "aws_s3_object" "object2" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/three_gossips/broken"
  content = "Dark Angel"
}

resource "aws_s3_object" "object3" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/navajo/north_window"
  content = "Balanced Rock"
}

resource "aws_s3_object" "object4" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/navajo/sand_dune"
  content = "Queen Victoria Rock"
}

resource "aws_s3_object" "object5" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/partition/park_avenue"
  content = "Double-O"
}

resource "aws_s3_object" "object6" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/courthouse_towers/landscape"
  content = "Fiery Furnace"
}

resource "aws_s3_object" "object7" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/rubicon"
  content = "Devils Garden"
}
`, randInt)
}

func testAccBucketObjectsDataSourceConfig_resourcesPlusAccessPoint(randInt int) string {
	return testAccBucketObjectsDataSourceConfig_resources(randInt) + fmt.Sprintf(`
resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.objects_bucket.bucket
  name   = "tf-objects-test-access-point-%[1]d"
}
`, randInt)
}

func testAccBucketObjectsDataSourceConfig_basic(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_s3_objects" "yesh" {
  bucket    = aws_s3_bucket.objects_bucket.id
  prefix    = "arch/navajo/"
  delimiter = "/"
}
`, testAccBucketObjectsDataSourceConfig_resources(randInt))
}

func testAccBucketObjectsDataSourceConfig_basicViaAccessPoint(randInt int) string {
	return testAccBucketObjectsDataSourceConfig_resourcesPlusAccessPoint(randInt) + `
data "aws_s3_objects" "yesh" {
  bucket    = aws_s3_access_point.test.arn
  prefix    = "arch/navajo/"
  delimiter = "/"
}
`
}

func testAccBucketObjectsDataSourceConfig_all(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_s3_objects" "yesh" {
  bucket = aws_s3_bucket.objects_bucket.id
}
`, testAccBucketObjectsDataSourceConfig_resources(randInt))
}

func testAccBucketObjectsDataSourceConfig_prefixes(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_s3_objects" "yesh" {
  bucket    = aws_s3_bucket.objects_bucket.id
  prefix    = "arch/"
  delimiter = "/"
}
`, testAccBucketObjectsDataSourceConfig_resources(randInt))
}

func testAccBucketObjectsDataSourceConfig_extraResource(randInt int) string {
	return fmt.Sprintf(`
%s

resource "aws_s3_object" "object8" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/ru b ic on"
  content = "Goose Island"
}
`, testAccBucketObjectsDataSourceConfig_resources(randInt))
}

func testAccBucketObjectsDataSourceConfig_encoded(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_s3_objects" "yesh" {
  bucket        = aws_s3_bucket.objects_bucket.id
  encoding_type = "url"
  prefix        = "arch/ru"
}
`, testAccBucketObjectsDataSourceConfig_extraResource(randInt))
}

func testAccBucketObjectsDataSourceConfig_maxKeys(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_s3_objects" "yesh" {
  bucket   = aws_s3_bucket.objects_bucket.id
  max_keys = 2
}
`, testAccBucketObjectsDataSourceConfig_resources(randInt))
}

func testAccBucketObjectsDataSourceConfig_startAfter(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_s3_objects" "yesh" {
  bucket      = aws_s3_bucket.objects_bucket.id
  start_after = "arch/three_gossips/broken"
}
`, testAccBucketObjectsDataSourceConfig_resources(randInt))
}

func testAccBucketObjectsDataSourceConfig_owners(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_s3_objects" "yesh" {
  bucket      = aws_s3_bucket.objects_bucket.id
  prefix      = "arch/three_gossips/"
  fetch_owner = true
}
`, testAccBucketObjectsDataSourceConfig_resources(randInt))
}
