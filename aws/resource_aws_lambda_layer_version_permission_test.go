package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_lambda_layer_version_permission", &resource.Sweeper{
		Name: "aws_lambda_layer_version_permission",
		F:    testSweepLambdaLayerVersionPermission,
	})
}

func testSweepLambdaLayerVersionPermission(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	lambdaconn := client.(*AWSClient).lambdaconn
	resp, err := lambdaconn.ListLayers(&lambda.ListLayersInput{})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Lambda Layer sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Lambda layers: %s", err)
	}

	if len(resp.Layers) == 0 {
		log.Print("[DEBUG] No aws lambda layers to sweep")
		return nil
	}

	for _, l := range resp.Layers {
		versionResp, err := lambdaconn.ListLayerVersions(&lambda.ListLayerVersionsInput{
			LayerName: l.LayerName,
		})
		if err != nil {
			return fmt.Errorf("Error retrieving versions for lambda layer: %s", err)
		}

		for _, v := range versionResp.LayerVersions {
			_, err := lambdaconn.DeleteLayerVersion(&lambda.DeleteLayerVersionInput{
				LayerName:     l.LayerName,
				VersionNumber: v.Version,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccAWSLambdaLayerVersionPermission_basic(t *testing.T) {
	resourceName := "aws_lambda_layer_version_permission.lambda_layer_permission"
	layerName := fmt.Sprintf("tf_acc_lambda_layer_version_permission_testing_%s", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaLayerVersionDestroy2,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaLayerVersionPermission(layerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaLayerVersionExists("aws_lambda_layer_version.lambda_layer", layerName),
					resource.TestCheckResourceAttr(resourceName, "action", "lambda:GetLayerVersion"),
					resource.TestCheckResourceAttr(resourceName, "principal", "*"),
					resource.TestCheckResourceAttr(resourceName, "statement_id", "xaccount"),
				),
			},
		},
	})
}

func testAccAWSLambdaLayerVersionPermission(layerName string) string {
	return fmt.Sprintf(`
resource "aws_lambda_layer_version" "lambda_layer" {
	filename   = "test-fixtures/lambdatest.zip"
	layer_name = "%s"
}

resource "aws_lambda_layer_version_permission" "lambda_layer_permission" {
	layer_arn = aws_lambda_layer_version.lambda_layer.layer_arn
	layer_version = aws_lambda_layer_version.lambda_layer.version
	action = "lambda:GetLayerVersion"
	statement_id = "xaccount"
	principal = "*"
}
`, layerName)
}

func testAccCheckLambdaLayerVersionDestroy2(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lambdaconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lambda_layer_version" {
			continue
		}

		layerName, version, err := resourceAwsLambdaLayerVersionParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.GetLayerVersion(&lambda.GetLayerVersionInput{
			LayerName:     aws.String(layerName),
			VersionNumber: aws.Int64(version),
		})
		if isAWSErr(err, lambda.ErrCodeResourceNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("Lambda Layer Version (%s) still exists", rs.Primary.ID)
	}

	return nil
}
