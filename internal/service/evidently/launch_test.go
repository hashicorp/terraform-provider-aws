package evidently_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatchevidently"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatchevidently "github.com/hashicorp/terraform-provider-aws/internal/service/evidently"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEvidentlyLaunch_basic(t *testing.T) {
	var launch cloudwatchevidently.Launch

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05Z")
	resourceName := "aws_evidently_launch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(cloudwatchevidently.EndpointsID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchevidently.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfig_basic(rName, rName2, rName3, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(resourceName, &launch),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "evidently", fmt.Sprintf("project/%s/launch/%s", rName, rName3)),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
					// not returned at create time
					// resource.TestCheckResourceAttr(resourceName, "execution.#", "1"),
					// resource.TestCheckResourceAttr(resourceName, "execution.0.started_time", startTime),
					resource.TestCheckResourceAttr(resourceName, "groups.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "groups.0.feature", "aws_evidently_feature.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "groups.0.name", "Variation1"),
					resource.TestCheckResourceAttr(resourceName, "groups.0.variation", "Variation1"),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_updated_time"),
					resource.TestCheckResourceAttr(resourceName, "name", rName3),
					resource.TestCheckResourceAttrPair(resourceName, "project", "aws_evidently_project.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "randomization_salt", rName3), // set to name if not specified
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.group_weights.Variation1", "0"),
					resource.TestCheckResourceAttr(resourceName, "scheduled_splits_config.0.steps.0.start_time", startTime),
					resource.TestCheckResourceAttr(resourceName, "status", cloudwatchevidently.LaunchStatusCreated),
					resource.TestCheckResourceAttr(resourceName, "type", cloudwatchevidently.LaunchTypeAwsEvidentlySplits),
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

func testAccCheckLaunchDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EvidentlyConn()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_evidently_launch" {
			continue
		}

		launchName, projectNameOrARN, err := tfcloudwatchevidently.LaunchParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfcloudwatchevidently.FindLaunchWithProjectNameorARN(context.Background(), conn, launchName, projectNameOrARN)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("CloudWatch Evidently Launch %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckLaunchExists(n string, v *cloudwatchevidently.Launch) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudWatch Evidently Launch ID is set")
		}

		launchName, projectNameOrARN, err := tfcloudwatchevidently.LaunchParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EvidentlyConn()

		output, err := tfcloudwatchevidently.FindLaunchWithProjectNameorARN(context.Background(), conn, launchName, projectNameOrARN)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccLaunchConfigBase(rName, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_evidently_project" "test" {
  name = %[1]q
}

resource "aws_evidently_feature" "test" {
  name    = %[2]q
  project = aws_evidently_project.test.name

  variations {
    name = "Variation1"
    value {
      string_value = "test"
    }
  }

  variations {
    name = "Variation1b"
    value {
      string_value = "test1b"
    }
  }
}
`, rName, rName2)
}

func testAccLaunchConfig_basic(rName, rName2, rName3, startTime string) string {
	return acctest.ConfigCompose(
		testAccLaunchConfigBase(rName, rName2),
		fmt.Sprintf(`
resource "aws_evidently_launch" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  groups {
    feature   = aws_evidently_feature.test.name
    name      = "Variation1"
    variation = "Variation1"
  }

  scheduled_splits_config {
    steps {
      group_weights = {
        "Variation1" = 0
      }
      start_time = %[2]q
    }
  }
}
`, rName3, startTime))
}
