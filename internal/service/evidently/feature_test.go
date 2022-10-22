package evidently_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatchevidently"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatchevidently "github.com/hashicorp/terraform-provider-aws/internal/service/evidently"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEvidentlyFeature_basic(t *testing.T) {
	var feature cloudwatchevidently.Feature

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_evidently_feature.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(cloudwatchevidently.EndpointsID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchevidently.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureConfig_basic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(resourceName, &feature),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "evidently", fmt.Sprintf("project/%s/feature/%s", rName, rName2)),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "default_variation", "Variation1"),
					resource.TestCheckResourceAttr(resourceName, "entity_overrides.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_rules.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "evaluation_strategy", cloudwatchevidently.FeatureEvaluationStrategyAllRules),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_time"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "project", rName),
					resource.TestCheckResourceAttr(resourceName, "status", cloudwatchevidently.FeatureStatusAvailable),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "value_type", cloudwatchevidently.VariationValueTypeString),
					resource.TestCheckResourceAttr(resourceName, "variations.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "variations.*", map[string]string{
						"name":                 "Variation1",
						"value.#":              "1",
						"value.0.string_value": "test",
					}),
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

func TestAccEvidentlyFeature_disappears(t *testing.T) {
	var feature cloudwatchevidently.Feature

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_evidently_feature.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchevidently.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureConfig_basic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureExists(resourceName, &feature),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudwatchevidently.ResourceFeature(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFeatureDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EvidentlyConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_evidently_feature" {
			continue
		}

		featureName, projectNameOrARN, err := tfcloudwatchevidently.FeatureParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfcloudwatchevidently.FindFeatureWithProjectNameorARN(context.Background(), conn, featureName, projectNameOrARN)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("CloudWatch Evidently Feature %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckFeatureExists(n string, v *cloudwatchevidently.Feature) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudWatch Evidently Feature ID is set")
		}

		featureName, projectNameOrARN, err := tfcloudwatchevidently.FeatureParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EvidentlyConn

		output, err := tfcloudwatchevidently.FindFeatureWithProjectNameorARN(context.Background(), conn, featureName, projectNameOrARN)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccFeatureConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_evidently_project" "test" {
  name = %[1]q
}
`, rName)
}

func testAccFeatureConfig_basic(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccFeatureConfigBase(rName),
		fmt.Sprintf(`
resource "aws_evidently_feature" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  variations {
    name = "Variation1"
    value {
      string_value = "test"
    }
  }
}
`, rName2))
}
