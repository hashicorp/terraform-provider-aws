package codestarnotifications_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarnotifications"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// For PreCheck, using acctest.PreCheckPartitionHasService does not work for
// codestarnotifications because it gives false positives always saying the
// partition (aws or GovCloud) does not support the service

func TestAccCodeStarNotificationsNotificationRule_basic(t *testing.T) {
	resourceName := "aws_codestarnotifications_notification_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codestarnotifications.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNotificationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationRuleBasicConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "codestar-notifications", regexp.MustCompile("notificationrule/.+")),
					resource.TestCheckResourceAttr(resourceName, "detail_type", codestarnotifications.DetailTypeBasic),
					resource.TestCheckResourceAttr(resourceName, "event_type_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", codestarnotifications.NotificationRuleStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "target.#", "1"),
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

func TestAccCodeStarNotificationsNotificationRule_status(t *testing.T) {
	resourceName := "aws_codestarnotifications_notification_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codestarnotifications.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNotificationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationRuleStatusConfig(rName, codestarnotifications.NotificationRuleStatusDisabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "status", codestarnotifications.NotificationRuleStatusDisabled),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotificationRuleStatusConfig(rName, codestarnotifications.NotificationRuleStatusEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "status", codestarnotifications.NotificationRuleStatusEnabled),
				),
			},
			{
				Config: testAccNotificationRuleStatusConfig(rName, codestarnotifications.NotificationRuleStatusDisabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "status", codestarnotifications.NotificationRuleStatusDisabled),
				),
			},
		},
	})
}

func TestAccCodeStarNotificationsNotificationRule_targets(t *testing.T) {
	resourceName := "aws_codestarnotifications_notification_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codestarnotifications.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNotificationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationRuleTargets1Config(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "target.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotificationRuleTargets2Config(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "target.#", "2"),
				),
			},
			{
				Config: testAccNotificationRuleTargets1Config(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "target.#", "1"),
				),
			},
		},
	})
}

func TestAccCodeStarNotificationsNotificationRule_tags(t *testing.T) {
	resourceName := "aws_codestarnotifications_notification_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codestarnotifications.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNotificationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationRuleTags1Config(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.TestTag1", "123456"),
					resource.TestCheckResourceAttr(resourceName, "tags.TestTag2", "654321"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotificationRuleTags2Config(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.TestTag2", "654321"),
					resource.TestCheckResourceAttr(resourceName, "tags.TestTag3", "asdfgh"),
				),
			},
			{
				Config: testAccNotificationRuleTags1Config(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.TestTag1", "123456"),
					resource.TestCheckResourceAttr(resourceName, "tags.TestTag2", "654321"),
				),
			},
		},
	})
}

func TestAccCodeStarNotificationsNotificationRule_eventTypeIDs(t *testing.T) {
	resourceName := "aws_codestarnotifications_notification_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codestarnotifications.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckNotificationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationRuleEventTypeIds1Config(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_type_ids.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotificationRuleEventTypeIds2Config(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_type_ids.#", "2"),
				),
			},
			{
				Config: testAccNotificationRuleEventTypeIds3Config(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_type_ids.#", "1"),
				),
			},
		},
	})
}

func testAccCheckNotificationRuleDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeStarNotificationsConn

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "aws_codestarnotifications_notification_rule":
			_, err := conn.DescribeNotificationRule(&codestarnotifications.DescribeNotificationRuleInput{
				Arn: aws.String(rs.Primary.ID),
			})

			if err != nil && !tfawserr.ErrCodeEquals(err, codestarnotifications.ErrCodeResourceNotFoundException) {
				return err
			}
		case "aws_sns_topic":
			res, err := conn.ListTargets(&codestarnotifications.ListTargetsInput{
				Filters: []*codestarnotifications.ListTargetsFilter{
					{
						Name:  aws.String("TARGET_ADDRESS"),
						Value: aws.String(rs.Primary.ID),
					},
					{
						Name:  aws.String("TARGET_TYPE"),
						Value: aws.String("SNS"),
					},
				},
				MaxResults: aws.Int64(1),
			})
			if err != nil {
				return err
			}
			if len(res.Targets) > 0 {
				return fmt.Errorf("codestar notification target (%s) is not removed", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeStarNotificationsConn

	input := &codestarnotifications.ListTargetsInput{
		MaxResults: aws.Int64(1),
	}

	_, err := conn.ListTargets(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccNotificationRuleBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_codecommit_repository" "test" {
  repository_name = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}
`, rName)
}

func testAccNotificationRuleBasicConfig(rName string) string {
	return testAccNotificationRuleBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]
  name           = %[1]q
  resource       = aws_codecommit_repository.test.arn
  status         = "ENABLED"

  tags = {
    TestTag = "123456"
  }

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName)
}

func testAccNotificationRuleStatusConfig(rName, status string) string {
	return testAccNotificationRuleBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]
  name           = %[1]q
  resource       = aws_codecommit_repository.test.arn
  status         = %[2]q

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName, status)
}

func testAccNotificationRuleTargets1Config(rName string) string {
	return testAccNotificationRuleBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]
  name           = %[1]q
  resource       = aws_codecommit_repository.test.arn

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName)
}

func testAccNotificationRuleTargets2Config(rName string) string {
	return testAccNotificationRuleBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sns_topic" "test2" {
  name = "%[1]s2"
}

resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]
  name           = %[1]q
  resource       = aws_codecommit_repository.test.arn

  target {
    address = aws_sns_topic.test.arn
  }

  target {
    address = aws_sns_topic.test2.arn
  }
}
`, rName)
}

func testAccNotificationRuleTags1Config(rName string) string {
	return testAccNotificationRuleBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]
  name           = %[1]q
  resource       = aws_codecommit_repository.test.arn
  status         = "ENABLED"

  tags = {
    TestTag1 = "123456"
    TestTag2 = "654321"
  }

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName)
}

func testAccNotificationRuleTags2Config(rName string) string {
	return testAccNotificationRuleBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]
  name           = %[1]q
  resource       = aws_codecommit_repository.test.arn
  status         = "ENABLED"

  tags = {
    TestTag2 = "654321"
    TestTag3 = "asdfgh"
  }

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName)
}

func testAccNotificationRuleEventTypeIds1Config(rName string) string {
	return testAccNotificationRuleBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type = "BASIC"
  event_type_ids = [
    "codecommit-repository-comments-on-commits",
  ]
  name     = %[1]q
  resource = aws_codecommit_repository.test.arn
  status   = "ENABLED"

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName)
}

func testAccNotificationRuleEventTypeIds2Config(rName string) string {
	return testAccNotificationRuleBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type = "BASIC"
  event_type_ids = [
    "codecommit-repository-comments-on-commits",
    "codecommit-repository-pull-request-created",
  ]
  name     = %[1]q
  resource = aws_codecommit_repository.test.arn
  status   = "ENABLED"

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName)
}

func testAccNotificationRuleEventTypeIds3Config(rName string) string {
	return testAccNotificationRuleBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type = "BASIC"
  event_type_ids = [
    "codecommit-repository-pull-request-created",
  ]
  name     = %[1]q
  resource = aws_codecommit_repository.test.arn
  status   = "ENABLED"

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName)
}
