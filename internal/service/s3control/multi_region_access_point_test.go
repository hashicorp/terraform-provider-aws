package s3control_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
)

func TestAccS3ControlMultiRegionAccessPoint_basic(t *testing.T) {
	var v s3control.MultiRegionAccessPointReport
	resourceName := "aws_s3control_multi_region_access_point.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckRegion(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:   acctest.ErrorCheck(t, s3control.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckMultiRegionAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionAccessPointConfig_basic(bucketName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionAccessPointExists(resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, "account_id"),
					resource.TestMatchResourceAttr(resourceName, "alias", regexp.MustCompile(`^[a-z][a-z0-9]*[.]mrap$`)),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "s3", regexp.MustCompile(`accesspoint\/[a-z][a-z0-9]*[.]mrap$`)),
					acctest.MatchResourceAttrGlobalHostname(resourceName, "domain_name", "accesspoint.s3-global", regexp.MustCompile(`^[a-z][a-z0-9]*[.]mrap`)),
					resource.TestCheckResourceAttr(resourceName, "details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "details.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "details.0.public_access_block.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "details.0.public_access_block.0.block_public_acls", "true"),
					resource.TestCheckResourceAttr(resourceName, "details.0.public_access_block.0.block_public_policy", "true"),
					resource.TestCheckResourceAttr(resourceName, "details.0.public_access_block.0.ignore_public_acls", "true"),
					resource.TestCheckResourceAttr(resourceName, "details.0.public_access_block.0.restrict_public_buckets", "true"),
					resource.TestCheckResourceAttr(resourceName, "details.0.region.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "details.0.region.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "status", s3control.MultiRegionAccessPointStatusReady),
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

func TestAccS3ControlMultiRegionAccessPoint_disappears(t *testing.T) {
	var v s3control.MultiRegionAccessPointReport
	resourceName := "aws_s3control_multi_region_access_point.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckRegion(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:   acctest.ErrorCheck(t, s3control.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckMultiRegionAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionAccessPointConfig_basic(bucketName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionAccessPointExists(resourceName, &v),
					testAccCheckMultiRegionAccessPointDisappears(acctest.Provider, tfs3control.ResourceMultiRegionAccessPoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlMultiRegionAccessPoint_PublicAccessBlock(t *testing.T) {
	var v s3control.MultiRegionAccessPointReport
	resourceName := "aws_s3control_multi_region_access_point.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckRegion(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:   acctest.ErrorCheck(t, s3control.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckMultiRegionAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionAccessPointConfig_publicAccessBlock(bucketName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionAccessPointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "details.0.public_access_block.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "details.0.public_access_block.0.block_public_acls", "false"),
					resource.TestCheckResourceAttr(resourceName, "details.0.public_access_block.0.block_public_policy", "false"),
					resource.TestCheckResourceAttr(resourceName, "details.0.public_access_block.0.ignore_public_acls", "false"),
					resource.TestCheckResourceAttr(resourceName, "details.0.public_access_block.0.restrict_public_buckets", "false"),
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

func TestAccS3ControlMultiRegionAccessPoint_name(t *testing.T) {
	var v1, v2 s3control.MultiRegionAccessPointReport
	resourceName := "aws_s3control_multi_region_access_point.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckRegion(t, endpoints.UsWest2RegionID)
		},
		ErrorCheck:   acctest.ErrorCheck(t, s3control.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckMultiRegionAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionAccessPointConfig_basic(bucketName, rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionAccessPointExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "details.0.name", rName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMultiRegionAccessPointConfig_basic(bucketName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionAccessPointExists(resourceName, &v2),
					testAccCheckMultiRegionAccessPointRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "details.0.name", rName2),
				),
			},
		},
	})
}

func testAccCheckMultiRegionAccessPointDisappears(provider *schema.Provider, resource *schema.Resource, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if resourceState.Primary.ID == "" {
			return fmt.Errorf("No S3 Multi-Region Access Point ID is set")
		}

		return acctest.DeleteResource(resource, resource.Data(resourceState.Primary), provider.Meta())
	}
}

func testAccCheckMultiRegionAccessPointDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3control_multi_region_access_point" {
			continue
		}

		accountId, name, err := tfs3control.MultiRegionAccessPointParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		resp, err := conn.GetMultiRegionAccessPoint(&s3control.GetMultiRegionAccessPointInput{
			AccountId: aws.String(accountId),
			Name:      aws.String(name),
		})

		if tfawserr.ErrCodeEquals(err, tfs3control.ErrCodeNoSuchMultiRegionAccessPoint) {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil && resp.AccessPoint != nil && fmt.Sprintf("%s:%s", accountId, aws.StringValue(resp.AccessPoint.Name)) == rs.Primary.ID {
			return fmt.Errorf("S3 Multi-Region Access Point with ID %v still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckMultiRegionAccessPointExists(n string, m *s3control.MultiRegionAccessPointReport) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		accountId, name, err := tfs3control.MultiRegionAccessPointParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn

		multiRegionAccessPoint, err := tfs3control.FindMultiRegionAccessPointByName(conn, accountId, name)

		if err != nil {
			return err
		}

		if multiRegionAccessPoint != nil {
			*m = *multiRegionAccessPoint
			return nil
		}

		return fmt.Errorf("Multi-Region Access Point not found")
	}
}

// Multi-Region Access Point aliases are unique throughout time and aren’t based on the name or configuration of a Multi-Region Access Point.
// If you create a Multi-Region Access Point, and then delete it and create another one with the same name and configuration, the
// second Multi-Region Access Point will have a different alias than the first. (https://docs.aws.amazon.com/AmazonS3/latest/userguide/CreatingMultiRegionAccessPoints.html#multi-region-access-point-naming)
func testAccCheckMultiRegionAccessPointRecreated(before, after *s3control.MultiRegionAccessPointReport) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.Alias), aws.StringValue(after.Alias); before == after {
			return fmt.Errorf("S3 Multi-Region Access Point (%s) not recreated", before)
		}

		return nil
	}
}

func testAccMultiRegionAccessPointConfig_basic(bucketName, multiRegionAccessPointName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3control_multi_region_access_point" "test" {
  details {
    name = %[2]q

    region {
      bucket = aws_s3_bucket.test.id
    }
  }
}
`, bucketName, multiRegionAccessPointName)
}

func testAccMultiRegionAccessPointConfig_publicAccessBlock(bucketName, multiRegionAccessPointName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3control_multi_region_access_point" "test" {
  details {
    name = %[2]q

    public_access_block {
      block_public_acls       = false
      block_public_policy     = false
      ignore_public_acls      = false
      restrict_public_buckets = false
    }

    region {
      bucket = aws_s3_bucket.test.id
    }
  }
}
`, bucketName, multiRegionAccessPointName)
}
