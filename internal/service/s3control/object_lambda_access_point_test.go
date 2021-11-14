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

func TestAccS3ControlObjectLambdaAccessPoint_basic(t *testing.T) {
	var v s3control.GetAccessPointForObjectLambdaOutput
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	accessPointName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_object_lambda_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3control.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckObjectLambdaAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectLambdaAccessPointConfig_basic(bucketName, accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointExists(resourceName, &v),
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

func TestAccS3ControlObjectLambdaAccessPoint_disappears(t *testing.T) {
	var v s3control.GetAccessPointForObjectLambdaOutput
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	accessPointName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_object_lambda_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3control.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckObjectLambdaAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectLambdaAccessPointConfig_basic(bucketName, accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3control.ResourceObjectLambdaAccessPoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3ControlObjectLambdaAccessPoint_disappears_Bucket(t *testing.T) {
	var v s3control.GetAccessPointForObjectLambdaOutput
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	accessPointName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_s3control_object_lambda_access_point.test"
	bucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, s3control.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckObjectLambdaAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectLambdaAccessPointConfig_basic(bucketName, accessPointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectLambdaAccessPointExists(resourceName, &v),
					testAccCheckDestroyBucket(bucketResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckObjectLambdaAccessPointDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3control_object_lambda_access_point" {
			continue
		}

		accountId, name, err := tfs3control.ObjectLambdaAccessPointParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.GetAccessPointForObjectLambda(&s3control.GetAccessPointForObjectLambdaInput{
			AccountId: aws.String(accountId),
			Name:      aws.String(name),
		})
		if err == nil {
			return fmt.Errorf("S3 Access Point still exists")
		}
	}
	return nil
}

func testAccCheckObjectLambdaAccessPointExists(n string, output *s3control.GetAccessPointForObjectLambdaOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Access Point ID is set")
		}

		accountId, name, err := tfs3control.ObjectLambdaAccessPointParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn

		resp, err := conn.GetAccessPointForObjectLambda(&s3control.GetAccessPointForObjectLambdaInput{
			AccountId: aws.String(accountId),
			Name:      aws.String(name),
		})
		if err != nil {
			return err
		}

		*output = *resp

		return nil
	}
}

func testAccObjectLambdaAccessPointConfig_basic(bucketName, accessPointName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3control_object_lambda_access_point" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[2]q
}
`, bucketName, accessPointName)
}
