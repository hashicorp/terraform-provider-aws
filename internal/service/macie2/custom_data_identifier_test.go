// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmacie2 "github.com/hashicorp/terraform-provider-aws/internal/service/macie2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccCustomDataIdentifier_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetCustomDataIdentifierOutput
	resourceName := "aws_macie2_custom_data_identifier.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	regex := "[0-9]{3}-[0-9]{2}-[0-9]{4}"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDataIdentifierDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDataIdentifierConfig_basic(rName, regex),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDataIdentifierExists(ctx, resourceName, &macie2Output),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "macie2", regexache.MustCompile(`custom-data-identifier/.+`)),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttr(resourceName, "regex", regex),
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

func testAccCustomDataIdentifier_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetCustomDataIdentifierOutput
	resourceName := "aws_macie2_custom_data_identifier.test"
	regex := "[0-9]{3}-[0-9]{2}-[0-9]{4}"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDataIdentifierDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDataIdentifierConfig_nameGenerated(regex),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDataIdentifierExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, id.UniqueIdPrefix),
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

func testAccCustomDataIdentifier_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetCustomDataIdentifierOutput
	resourceName := "aws_macie2_custom_data_identifier.test"
	regex := "[0-9]{3}-[0-9]{2}-[0-9]{4}"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDataIdentifierDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDataIdentifierConfig_namePrefix("tf-acc-test-prefix-", regex),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDataIdentifierExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
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

func testAccCustomDataIdentifier_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetCustomDataIdentifierOutput
	resourceName := "aws_macie2_custom_data_identifier.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	regex := "[0-9]{3}-[0-9]{2}-[0-9]{4}"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDataIdentifierDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDataIdentifierConfig_basic(rName, regex),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDataIdentifierExists(ctx, resourceName, &macie2Output),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfmacie2.ResourceCustomDataIdentifier(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCustomDataIdentifier_WithClassificationJob(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetCustomDataIdentifierOutput
	resourceName := "aws_macie2_custom_data_identifier.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	regex := "[0-9]{3}-[0-9]{2}-[0-9]{4}"
	description := "this is a description"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDataIdentifierDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDataIdentifierConfig_complete(rName, regex, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDataIdentifierExists(ctx, resourceName, &macie2Output),
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

func testAccCustomDataIdentifier_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var macie2Output macie2.GetCustomDataIdentifierOutput
	resourceName := "aws_macie2_custom_data_identifier.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	regex := "[0-9]{3}-[0-9]{2}-[0-9]{4}"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDataIdentifierDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Macie2ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDataIdentifierConfig_tags1(rName, regex, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDataIdentifierExists(ctx, resourceName, &macie2Output),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCustomDataIdentifierConfig_tags2(rName, regex, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDataIdentifierExists(ctx, resourceName, &macie2Output),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccCustomDataIdentifierConfig_tags1(rName, regex, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDataIdentifierExists(ctx, resourceName, &macie2Output),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccCheckCustomDataIdentifierExists(ctx context.Context, n string, v *macie2.GetCustomDataIdentifierOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Client(ctx)

		output, err := tfmacie2.FindCustomDataIdentifierByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckCustomDataIdentifierDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Macie2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_macie2_custom_data_identifier" {
				continue
			}

			_, err := tfmacie2.FindCustomDataIdentifierByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Macie Custom Data Identifier %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCustomDataIdentifierConfig_basic(rName, regex string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {}

resource "aws_macie2_custom_data_identifier" "test" {
  name  = %[1]q
  regex = %[2]q

  depends_on = [aws_macie2_account.test]
}
`, rName, regex)
}

func testAccCustomDataIdentifierConfig_nameGenerated(regex string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {}

resource "aws_macie2_custom_data_identifier" "test" {
  regex = %[1]q

  depends_on = [aws_macie2_account.test]
}
`, regex)
}

func testAccCustomDataIdentifierConfig_namePrefix(name, regex string) string {
	return fmt.Sprintf(`
resource "aws_macie2_account" "test" {}

resource "aws_macie2_custom_data_identifier" "test" {
  name_prefix = %[1]q
  regex       = %[2]q

  depends_on = [aws_macie2_account.test]
}
`, name, regex)
}

func testAccCustomDataIdentifierConfig_complete(rName, regex, description string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_macie2_account" "test" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_macie2_custom_data_identifier" "test" {
  name                   = %[1]q
  regex                  = %[2]q
  description            = %[3]q
  maximum_match_distance = 10
  keywords               = ["test"]
  ignore_words           = ["not testing"]

  depends_on = [aws_macie2_account.test]
}

resource "aws_macie2_classification_job" "test" {
  name                       = %[1]q
  custom_data_identifier_ids = [aws_macie2_custom_data_identifier.test.id]
  job_type                   = "SCHEDULED"
  s3_job_definition {
    bucket_definitions {
      account_id = data.aws_caller_identity.current.account_id
      buckets    = [aws_s3_bucket.test.bucket]
    }
  }
  schedule_frequency {
    daily_schedule = true
  }
  sampling_percentage = 100
  description         = "test"
  initial_run         = true
}
`, rName, regex, description)
}

func testAccCustomDataIdentifierConfig_tags1(rName, regex, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_macie2_account" "test" {}

resource "aws_macie2_custom_data_identifier" "test" {
  name                   = %[1]q
  regex                  = %[2]q
  description            = "this a description"
  maximum_match_distance = 10
  keywords               = ["test"]
  ignore_words           = ["not testing"]

  tags = {
    %[3]q = %[4]q
  }

  depends_on = [aws_macie2_account.test]
}
`, rName, regex, tag1Key, tag1Value)
}

func testAccCustomDataIdentifierConfig_tags2(rName, regex, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_macie2_account" "test" {}

resource "aws_macie2_custom_data_identifier" "test" {
  name                   = %[1]q
  regex                  = %[2]q
  description            = "this a description"
  maximum_match_distance = 10
  keywords               = ["test"]
  ignore_words           = ["not testing"]

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }

  depends_on = [aws_macie2_account.test]
}
`, rName, regex, tag1Key, tag1Value, tag2Key, tag2Value)
}
