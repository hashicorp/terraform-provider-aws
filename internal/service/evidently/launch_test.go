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

func TestAccEvidentlyLaunch_updateDescription(t *testing.T) {
	var launch cloudwatchevidently.Launch

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05Z")
	originalDescription := "original description"
	updatedDescription := "updated description"
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
				Config: testAccLaunchConfig_description(rName, rName2, rName3, startTime, originalDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "description", originalDescription),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchConfig_description(rName, rName2, rName3, startTime, updatedDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "description", updatedDescription),
				),
			},
		},
	})
}

func TestAccEvidentlyLaunch_updateGroups(t *testing.T) {
	var launch cloudwatchevidently.Launch

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName4 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName5 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
					resource.TestCheckResourceAttr(resourceName, "groups.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "groups.0.feature", "aws_evidently_feature.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "groups.0.name", "Variation1"),
					resource.TestCheckResourceAttr(resourceName, "groups.0.variation", "Variation1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchConfig_twoGroups(rName, rName2, rName3, rName4, rName5, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "groups.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "groups.0.description", "first-group-add-desc"),
					resource.TestCheckResourceAttrPair(resourceName, "groups.0.feature", "aws_evidently_feature.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "groups.0.name", "Variation1UpdatedName"),
					resource.TestCheckResourceAttr(resourceName, "groups.0.variation", "Variation1"),
					resource.TestCheckResourceAttr(resourceName, "groups.1.description", "second-group"),
					resource.TestCheckResourceAttrPair(resourceName, "groups.1.feature", "aws_evidently_feature.test2", "name"),
					resource.TestCheckResourceAttr(resourceName, "groups.1.name", "Variation2OriginalName"),
					resource.TestCheckResourceAttr(resourceName, "groups.1.variation", "Variation2"),
				),
			},
			{
				Config: testAccLaunchConfig_threeGroups(rName, rName2, rName3, rName4, rName5, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "groups.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "groups.0.description", ""),
					resource.TestCheckResourceAttrPair(resourceName, "groups.0.feature", "aws_evidently_feature.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "groups.0.name", "Variation1"),
					resource.TestCheckResourceAttr(resourceName, "groups.0.variation", "Variation1"),
					resource.TestCheckResourceAttr(resourceName, "groups.1.description", "second-group-update-desc"),
					resource.TestCheckResourceAttrPair(resourceName, "groups.1.feature", "aws_evidently_feature.test2", "name"),
					resource.TestCheckResourceAttr(resourceName, "groups.1.name", "Variation2UpdatedName"),
					resource.TestCheckResourceAttr(resourceName, "groups.1.variation", "Variation2a"),
					resource.TestCheckResourceAttr(resourceName, "groups.2.description", "third-group"),
					resource.TestCheckResourceAttrPair(resourceName, "groups.2.feature", "aws_evidently_feature.test3", "name"),
					resource.TestCheckResourceAttr(resourceName, "groups.2.name", "Variation3OriginalName"),
					resource.TestCheckResourceAttr(resourceName, "groups.2.variation", "Variation3"),
				),
			},
			{
				Config: testAccLaunchConfig_basic(rName, rName2, rName3, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "groups.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "groups.0.feature", "aws_evidently_feature.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "groups.0.name", "Variation1"),
					resource.TestCheckResourceAttr(resourceName, "groups.0.variation", "Variation1"),
				),
			},
		},
	})
}

func TestAccEvidentlyLaunch_updateMetricMonitors(t *testing.T) {
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
				Config: testAccLaunchConfig_oneMetricMonitor(rName, rName2, rName3, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.entity_id_key", "entity_id_key1"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.event_pattern", "{\"Price\":[{\"numeric\":[\">\",10,\"<=\",20]}]}"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.name", "name1"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.unit_label", "unit_label1"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.value_key", "value_key1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchConfig_twoMetricMonitors(rName, rName2, rName3, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.entity_id_key", "entity_id_key1a"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.event_pattern", "{\"Price\":[{\"numeric\":[\">\",11,\"<=\",22]}]}"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.name", "name1a"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.unit_label", "unit_label1a"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.value_key", "value_key1a"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.0.entity_id_key", "entity_id_key2"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.0.event_pattern", "{\"Price\":[{\"numeric\":[\">\",9,\"<=\",19]}]}"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.0.name", "name2"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.0.unit_label", "unit_label2"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.0.value_key", "value_key2"),
				),
			},
			{
				Config: testAccLaunchConfig_threeMetricMonitors(rName, rName2, rName3, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.entity_id_key", "entity_id_key1b"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.event_pattern", "{\"Price\":[{\"numeric\":[\">\",15,\"<=\",25]}]}"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.name", "name1b"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.unit_label", "unit_label1b"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.0.metric_definition.0.value_key", "value_key1b"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.0.entity_id_key", "entity_id_key2a"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.0.event_pattern", "{\"Price\":[{\"numeric\":[\">\",8,\"<=\",18]}]}"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.0.name", "name2a"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.0.unit_label", "unit_label2a"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.1.metric_definition.0.value_key", "value_key2a"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.2.metric_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.2.metric_definition.0.entity_id_key", "entity_id_key3"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.2.metric_definition.0.name", "name3"),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.2.metric_definition.0.value_key", "value_key3"),
				),
			},
			{
				Config: testAccLaunchConfig_basic(rName, rName2, rName3, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "metric_monitors.#", "0"),
				),
			},
		},
	})
}

func TestAccEvidentlyLaunch_tags(t *testing.T) {
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
				Config: testAccLaunchConfig_tags1(rName, rName2, rName3, startTime, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLaunchConfig_tags2(rName, rName2, rName3, startTime, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLaunchConfig_tags1(rName, rName2, rName3, startTime, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(resourceName, &launch),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEvidentlyLaunch_disappears(t *testing.T) {
	var launch cloudwatchevidently.Launch

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	startTime := time.Now().AddDate(0, 0, 2).Format("2006-01-02T15:04:05Z")
	resourceName := "aws_evidently_launch.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, cloudwatchevidently.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLaunchDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfig_basic(rName, rName2, rName3, startTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLaunchExists(resourceName, &launch),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudwatchevidently.ResourceLaunch(), resourceName),
				),
				ExpectNonEmptyPlan: true,
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

func testAccLaunchConfig_description(rName, rName2, rName3, startTime, description string) string {
	return acctest.ConfigCompose(
		testAccLaunchConfigBase(rName, rName2),
		fmt.Sprintf(`
resource "aws_evidently_launch" "test" {
  name        = %[1]q
  project     = aws_evidently_project.test.name
  description = %[3]q

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
`, rName3, startTime, description))
}

func testAccLaunchConfigGroupsBase(rName, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_evidently_feature" "test2" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  variations {
    name = "Variation2"
    value {
      string_value = "test2"
    }
  }

  variations {
    name = "Variation2a"
    value {
      string_value = "test2a"
    }
  }
}

resource "aws_evidently_feature" "test3" {
  name    = %[2]q
  project = aws_evidently_project.test.name

  variations {
    name = "Variation3"
    value {
      string_value = "test3"
    }
  }
}
`, rName, rName2)
}

func testAccLaunchConfig_twoGroups(rName, rName2, rName3, rName4, rName5, startTime string) string {
	return acctest.ConfigCompose(
		testAccLaunchConfigBase(rName, rName2),
		testAccLaunchConfigGroupsBase(rName3, rName4),
		fmt.Sprintf(`
resource "aws_evidently_launch" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  groups {
    feature     = aws_evidently_feature.test.name
    name        = "Variation1UpdatedName"
    variation   = "Variation1"
    description = "first-group-add-desc"
  }

  groups {
    feature     = aws_evidently_feature.test2.name
    name        = "Variation2OriginalName"
    variation   = "Variation2"
    description = "second-group"
  }

  scheduled_splits_config {
    steps {
      group_weights = {
        "Variation1UpdatedName"  = 0
        "Variation2OriginalName" = 0
      }
      start_time = %[2]q
    }
  }
}
`, rName5, startTime))
}

func testAccLaunchConfig_threeGroups(rName, rName2, rName3, rName4, rName5, startTime string) string {
	return acctest.ConfigCompose(
		testAccLaunchConfigBase(rName, rName2),
		testAccLaunchConfigGroupsBase(rName3, rName4),
		fmt.Sprintf(`
resource "aws_evidently_launch" "test" {
  name    = %[1]q
  project = aws_evidently_project.test.name

  groups {
    feature   = aws_evidently_feature.test.name
    name      = "Variation1"
    variation = "Variation1"
  }

  groups {
    feature     = aws_evidently_feature.test2.name
    name        = "Variation2UpdatedName"
    variation   = "Variation2a"
    description = "second-group-update-desc"
  }

  groups {
    feature     = aws_evidently_feature.test3.name
    name        = "Variation3OriginalName"
    variation   = "Variation3"
    description = "third-group"
  }

  scheduled_splits_config {
    steps {
      group_weights = {
        "Variation1"             = 0
        "Variation2UpdatedName"  = 0
        "Variation3OriginalName" = 0
      }
      start_time = %[2]q
    }
  }
}
`, rName5, startTime))
}

func testAccLaunchConfig_oneMetricMonitor(rName, rName2, rName3, startTime string) string {
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

  metric_monitors {
    metric_definition {
      entity_id_key = "entity_id_key1"
      event_pattern = "{\"Price\":[{\"numeric\":[\">\",10,\"<=\",20]}]}"
      name          = "name1"
      unit_label    = "unit_label1"
      value_key     = "value_key1"
    }
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

func testAccLaunchConfig_twoMetricMonitors(rName, rName2, rName3, startTime string) string {
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

  metric_monitors {
    metric_definition {
      entity_id_key = "entity_id_key1a"
      event_pattern = "{\"Price\":[{\"numeric\":[\">\",11,\"<=\",22]}]}"
      name          = "name1a"
      unit_label    = "unit_label1a"
      value_key     = "value_key1a"
    }
  }

  metric_monitors {
    metric_definition {
      entity_id_key = "entity_id_key2"
      event_pattern = "{\"Price\":[{\"numeric\":[\">\",9,\"<=\",19]}]}"
      name          = "name2"
      unit_label    = "unit_label2"
      value_key     = "value_key2"
    }
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

func testAccLaunchConfig_threeMetricMonitors(rName, rName2, rName3, startTime string) string {
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

  metric_monitors {
    metric_definition {
      entity_id_key = "entity_id_key1b"
      event_pattern = "{\"Price\":[{\"numeric\":[\">\",15,\"<=\",25]}]}"
      name          = "name1b"
      unit_label    = "unit_label1b"
      value_key     = "value_key1b"
    }
  }

  metric_monitors {
    metric_definition {
      entity_id_key = "entity_id_key2a"
      event_pattern = "{\"Price\":[{\"numeric\":[\">\",8,\"<=\",18]}]}"
      name          = "name2a"
      unit_label    = "unit_label2a"
      value_key     = "value_key2a"
    }
  }

  metric_monitors {
    metric_definition {
      entity_id_key = "entity_id_key3"
      name          = "name3"
      value_key     = "value_key3"
    }
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

func testAccLaunchConfig_tags1(rName, rName2, rName3, startTime, tag, value string) string {
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

  tags = {
    %[3]q = %[4]q
  }
}
`, rName3, startTime, tag, value))
}

func testAccLaunchConfig_tags2(rName, rName2, rName3, startTime, tag1, value1, tag2, value2 string) string {
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

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName3, startTime, tag1, value1, tag2, value2))
}
