package s3control_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
)

func TestAccS3ControlMultiRegionAccessPointPolicy_basic(t *testing.T) {
	var v s3control.MultiRegionAccessPointPolicyDocument
	resourceName := "aws_s3control_multi_region_access_point_policy.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	multiRegionAccessPointName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	if acctest.Partition() == "aws-us-gov" {
		t.Skip("S3 Multi-Region Access Point is not supported in GovCloud partition")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		// Multi-Region Access Point Policy cannot be deleted once applied.
		// Ensure parent resource is destroyed instead.
		CheckDestroy: testAccCheckMultiRegionAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionAccessPointPolicyConfig_basic(bucketName, multiRegionAccessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionAccessPointPolicyExists(resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, "account_id"),
					resource.TestCheckResourceAttr(resourceName, "details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "details.0.name", multiRegionAccessPointName),
					resource.TestCheckResourceAttrSet(resourceName, "details.0.policy"),
					resource.TestCheckResourceAttrSet(resourceName, "established"),
					resource.TestCheckResourceAttrSet(resourceName, "proposed"),
					resource.TestCheckResourceAttrPair(resourceName, "details.0.policy", resourceName, "proposed"),
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

func TestAccS3ControlMultiRegionAccessPointPolicy_disappears_MultiRegionAccessPoint(t *testing.T) {
	var v s3control.MultiRegionAccessPointReport
	parentResourceName := "aws_s3control_multi_region_access_point.test"
	resourceName := "aws_s3control_multi_region_access_point_policy.test"
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	if acctest.Partition() == "aws-us-gov" {
		t.Skip("S3 Multi-Region Access Point is not supported in GovCloud partition")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		// Multi-Region Access Point Policy cannot be deleted once applied.
		// Ensure parent resource is destroyed instead.
		CheckDestroy: testAccCheckMultiRegionAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionAccessPointPolicyConfig_basic(bucketName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionAccessPointExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3control.ResourceMultiRegionAccessPoint(), parentResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlMultiRegionAccessPointPolicy_details_policy(t *testing.T) {
	var v1, v2 s3control.MultiRegionAccessPointPolicyDocument
	resourceName := "aws_s3control_multi_region_access_point_policy.test"
	multiRegionAccessPointName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	if acctest.Partition() == "aws-us-gov" {
		t.Skip("S3 Multi-Region Access Point is not supported in GovCloud partition")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		// Multi-Region Access Point Policy cannot be deleted once applied.
		// Ensure parent resource is destroyed instead.
		CheckDestroy: testAccCheckMultiRegionAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionAccessPointPolicyConfig_basic(bucketName, multiRegionAccessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionAccessPointPolicyExists(resourceName, &v1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMultiRegionAccessPointPolicyConfig_updatedStatement(bucketName, multiRegionAccessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionAccessPointPolicyExists(resourceName, &v2),
					testAccCheckMultiRegionAccessPointPolicyChanged(&v1, &v2),
				),
			},
		},
	})
}

func TestAccS3ControlMultiRegionAccessPointPolicy_details_name(t *testing.T) {
	var v1, v2 s3control.MultiRegionAccessPointPolicyDocument
	resourceName := "aws_s3control_multi_region_access_point_policy.test"
	multiRegionAccessPointName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	multiRegionAccessPointName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	if acctest.Partition() == "aws-us-gov" {
		t.Skip("S3 Multi-Region Access Point is not supported in GovCloud partition")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		// Multi-Region Access Point Policy cannot be deleted once applied.
		// Ensure parent resource is destroyed instead.
		CheckDestroy: testAccCheckMultiRegionAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMultiRegionAccessPointPolicyConfig_basic(bucketName, multiRegionAccessPointName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionAccessPointPolicyExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "details.0.name", multiRegionAccessPointName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMultiRegionAccessPointPolicyConfig_basic(bucketName, multiRegionAccessPointName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMultiRegionAccessPointPolicyExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "details.0.name", multiRegionAccessPointName2),
				),
			},
		},
	})
}

func testAccCheckMultiRegionAccessPointPolicyExists(n string, v *s3control.MultiRegionAccessPointPolicyDocument) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Multi-Region Access Point Policy ID is set")
		}

		accountID, name, err := tfs3control.MultiRegionAccessPointParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn, err := tfs3control.ConnForMRAP(acctest.Provider.Meta().(*conns.AWSClient))

		if err != nil {
			return err
		}

		output, err := tfs3control.FindMultiRegionAccessPointPolicyDocumentByAccountIDAndName(conn, accountID, name)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckMultiRegionAccessPointPolicyChanged(i, j *s3control.MultiRegionAccessPointPolicyDocument) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.Proposed.Policy) == aws.StringValue(j.Proposed.Policy) {
			return fmt.Errorf("S3 Multi-Region Access Point Policy did not change")
		}

		return nil
	}
}

func testAccMultiRegionAccessPointPolicyConfig_basic(bucketName, multiRegionAccessPointName string) string {
	return acctest.ConfigCompose(
		testAccMultiRegionAccessPointConfig_basic(bucketName, multiRegionAccessPointName),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_s3control_multi_region_access_point_policy" "test" {
  details {
    name = %[1]q
    policy = jsonencode({
      "Version" : "2012-10-17",
      "Statement" : [
        {
          "Sid" : "Test",
          "Effect" : "Allow",
          "Principal" : {
            "AWS" : data.aws_caller_identity.current.account_id
          },
          "Action" : "s3:GetObject",
          "Resource" : "arn:${data.aws_partition.current.partition}:s3::${data.aws_caller_identity.current.account_id}:accesspoint/${aws_s3control_multi_region_access_point.test.alias}/object/*"
        }
      ]
    })
  }
}
`, multiRegionAccessPointName))
}

func testAccMultiRegionAccessPointPolicyConfig_updatedStatement(bucketName, multiRegionAccessPointName string) string {
	return acctest.ConfigCompose(
		testAccMultiRegionAccessPointConfig_basic(bucketName, multiRegionAccessPointName),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_s3control_multi_region_access_point_policy" "test" {
  details {
    name = %[1]q
    policy = jsonencode({
      "Version" : "2012-10-17",
      "Statement" : [
        {
          "Sid" : "Test",
          "Effect" : "Allow",
          "Principal" : {
            "AWS" : data.aws_caller_identity.current.account_id
          },
          "Action" : "s3:PutObject",
          "Resource" : "arn:${data.aws_partition.current.partition}:s3::${data.aws_caller_identity.current.account_id}:accesspoint/${aws_s3control_multi_region_access_point.test.alias}/object/*"
        }
      ]
    })
  }
}
`, multiRegionAccessPointName))
}
