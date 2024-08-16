// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resiliencehub_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehub"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehub/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfresiliencehub "github.com/hashicorp/terraform-provider-aws/internal/service/resiliencehub"
)

func TestAccResilienceHubResiliencyPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policy resiliencehub.DescribeResiliencyPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_resiliency_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
			acctest.PreCheckPartitionHasService(t, names.ResilienceHubServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResiliencyPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResiliencyPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_description", rName),
					resource.TestCheckResourceAttr(resourceName, "tier", "NotApplicable"),
					resource.TestCheckResourceAttr(resourceName, "data_location_constraint", "AnyLocation"),
					resource.TestCheckResourceAttr(resourceName, "policy.az.rpo_in_secs", "3600"),
					resource.TestCheckResourceAttr(resourceName, "policy.az.rto_in_secs", "3600"),
					resource.TestCheckResourceAttr(resourceName, "policy.hardware.rpo_in_secs", "3600"),
					resource.TestCheckResourceAttr(resourceName, "policy.hardware.rto_in_secs", "3600"),
					resource.TestCheckResourceAttr(resourceName, "policy.region.rpo_in_secs", "3600"),
					resource.TestCheckResourceAttr(resourceName, "policy.region.rto_in_secs", "3600"),
					resource.TestCheckResourceAttr(resourceName, "policy.software.rpo_in_secs", "3600"),
					resource.TestCheckResourceAttr(resourceName, "policy.software.rto_in_secs", "3600"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, names.ResilienceHubServiceID, regexache.MustCompile(`resiliency-policy/.+`)),
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

func TestAccResilienceHubResiliencyPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policy1, policy2, policy3, policy4, policy5 resiliencehub.DescribeResiliencyPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_resiliency_policy.test"

	updatedPolicyDescription := "updated policy desecription"
	updatedDataLocationConstraint := "SameCountry"
	updatedTier := "MissionCritical"
	updatedPolicyObjValue := "86400"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
			acctest.PreCheckPartitionHasService(t, names.ResilienceHubServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResiliencyPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResiliencyPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy1),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_description", rName),
					resource.TestCheckResourceAttr(resourceName, "tier", "NotApplicable"),
					resource.TestCheckResourceAttr(resourceName, "data_location_constraint", "AnyLocation"),
					resource.TestCheckResourceAttr(resourceName, "policy.az.rpo_in_secs", "3600"),
					resource.TestCheckResourceAttr(resourceName, "policy.az.rto_in_secs", "3600"),
					resource.TestCheckResourceAttr(resourceName, "policy.hardware.rpo_in_secs", "3600"),
					resource.TestCheckResourceAttr(resourceName, "policy.hardware.rto_in_secs", "3600"),
					resource.TestCheckResourceAttr(resourceName, "policy.region.rpo_in_secs", "3600"),
					resource.TestCheckResourceAttr(resourceName, "policy.region.rto_in_secs", "3600"),
					resource.TestCheckResourceAttr(resourceName, "policy.software.rpo_in_secs", "3600"),
					resource.TestCheckResourceAttr(resourceName, "policy.software.rto_in_secs", "3600"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, names.ResilienceHubServiceID, regexache.MustCompile(`resiliency-policy/.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResiliencyPolicyConfig_updatePolicydescription(rName, updatedPolicyDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy2),
					testAccCheckResiliencyPolicyNotRecreated(&policy1, &policy2),
					resource.TestCheckResourceAttr(resourceName, "policy_description", updatedPolicyDescription),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, names.ResilienceHubServiceID, regexache.MustCompile(`resiliency-policy/.+`)),
				),
			},
			{
				Config: testAccResiliencyPolicyConfig_updateDataLocationConstraint(rName, updatedDataLocationConstraint),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy3),
					testAccCheckResiliencyPolicyNotRecreated(&policy2, &policy3),
					resource.TestCheckResourceAttr(resourceName, "data_location_constraint", updatedDataLocationConstraint),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, names.ResilienceHubServiceID, regexache.MustCompile(`resiliency-policy/.+`)),
				),
			},
			{
				Config: testAccResiliencyPolicyConfig_updateTier(rName, updatedTier),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy4),
					testAccCheckResiliencyPolicyNotRecreated(&policy3, &policy4),
					resource.TestCheckResourceAttr(resourceName, "tier", updatedTier),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, names.ResilienceHubServiceID, regexache.MustCompile(`resiliency-policy/.+`)),
				),
			},
			{
				Config: testAccResiliencyPolicyConfig_updatePolicy(rName, updatedPolicyObjValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy5),
					resource.TestCheckResourceAttr(resourceName, "policy.az.rpo_in_secs", updatedPolicyObjValue),
					resource.TestCheckResourceAttr(resourceName, "policy.az.rto_in_secs", updatedPolicyObjValue),
					resource.TestCheckResourceAttr(resourceName, "policy.hardware.rpo_in_secs", updatedPolicyObjValue),
					resource.TestCheckResourceAttr(resourceName, "policy.hardware.rto_in_secs", updatedPolicyObjValue),
					resource.TestCheckResourceAttr(resourceName, "policy.region.rpo_in_secs", updatedPolicyObjValue),
					resource.TestCheckResourceAttr(resourceName, "policy.region.rto_in_secs", updatedPolicyObjValue),
					resource.TestCheckResourceAttr(resourceName, "policy.software.rpo_in_secs", updatedPolicyObjValue),
					resource.TestCheckResourceAttr(resourceName, "policy.software.rto_in_secs", updatedPolicyObjValue),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, names.ResilienceHubServiceID, regexache.MustCompile(`resiliency-policy/.+`)),
				),
			},
		},
	})
}

func TestAccResilienceHubResiliencyPolicy_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policy1, policy2 resiliencehub.DescribeResiliencyPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_resiliency_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
			acctest.PreCheckPartitionHasService(t, names.ResilienceHubServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResiliencyPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResiliencyPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy1),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Value", "Other"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, names.ResilienceHubServiceID, regexache.MustCompile(`resiliency-policy/.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResiliencyPolicyConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy1),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, names.ResilienceHubServiceID, regexache.MustCompile(`resiliency-policy/.+`)),
				),
			},
			{
				Config: testAccResiliencyPolicyConfig_tag2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy2),
					testAccCheckResiliencyPolicyNotRecreated(&policy1, &policy2),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, names.ResilienceHubServiceID, regexache.MustCompile(`resiliency-policy/.+`)),
				),
			},
		},
	})
}

func TestAccResilienceHubResiliencyPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policy resiliencehub.DescribeResiliencyPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_resiliency_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
			acctest.PreCheckPartitionHasService(t, names.ResilienceHubServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResiliencyPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResiliencyPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResiliencyPolicyExists(ctx, resourceName, &policy),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfresiliencehub.ResourceResiliencyPolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckResiliencyPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ResilienceHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_resiliencehub_resiliency_policy" {
				continue
			}

			input := &resiliencehub.DescribeResiliencyPolicyInput{
				PolicyArn: aws.String(rs.Primary.ID),
			}
			_, err := conn.DescribeResiliencyPolicy(ctx, input)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ResilienceHub, create.ErrActionCheckingDestroyed, tfresiliencehub.ResNameResiliencyPolicy, rs.Primary.ID, err)
			}

			return create.Error(names.ResilienceHub, create.ErrActionCheckingDestroyed, tfresiliencehub.ResNameResiliencyPolicy, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckResiliencyPolicyExists(ctx context.Context, name string, policy *resiliencehub.DescribeResiliencyPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ResilienceHub, create.ErrActionCheckingExistence, tfresiliencehub.ResNameResiliencyPolicy, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ResilienceHub, create.ErrActionCheckingExistence, tfresiliencehub.ResNameResiliencyPolicy, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ResilienceHubClient(ctx)
		resp, err := conn.DescribeResiliencyPolicy(ctx, &resiliencehub.DescribeResiliencyPolicyInput{
			PolicyArn: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.ResilienceHub, create.ErrActionCheckingExistence, tfresiliencehub.ResNameResiliencyPolicy, rs.Primary.ID, err)
		}

		*policy = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ResilienceHubClient(ctx)

	input := &resiliencehub.ListResiliencyPoliciesInput{}
	_, err := conn.ListResiliencyPolicies(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckResiliencyPolicyNotRecreated(before, after *resiliencehub.DescribeResiliencyPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Policy.PolicyArn), aws.ToString(after.Policy.PolicyArn); before != after {
			return create.Error(names.ResilienceHub, create.ErrActionCheckingNotRecreated, tfresiliencehub.ResNameResiliencyPolicy, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccResiliencyPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehub_resiliency_policy" "test" {
  policy_name        = %[1]q

  policy_description = %[1]q

  tier = "NotApplicable"

  data_location_constraint = "AnyLocation"

  policy {
    region {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
    az {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
    hardware {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
    software {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
  }

  tags = {
    Name = %[1]q
	Value = "Other"
  }
}
`, rName)
}

func testAccResiliencyPolicyConfig_updatePolicydescription(rName, resPolicyDescValue string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehub_resiliency_policy" "test" {
  policy_name        = %[1]q

  policy_description = %[2]q

  tier = "NotApplicable"

  data_location_constraint = "AnyLocation"

  policy {
    region {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
    az {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
    hardware {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
    software {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
  }

  tags = {
    Name = %[1]q
	Value = "Other"
  }
}
`, rName, resPolicyDescValue)
}

func testAccResiliencyPolicyConfig_updateDataLocationConstraint(rName, resDataLocConstValue string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehub_resiliency_policy" "test" {
  policy_name        = %[1]q

  policy_description = %[1]q

  tier = "NotApplicable"

  data_location_constraint = %[2]q

  policy {
    region {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
    az {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
    hardware {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
    software {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
  }

  tags = {
    Name = %[1]q
	Value = "Other"
  }
}
`, rName, resDataLocConstValue)
}

func testAccResiliencyPolicyConfig_updateTier(rName, resTierValue string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehub_resiliency_policy" "test" {
  policy_name        = %[1]q

  policy_description = %[1]q

  tier = %[2]q

  data_location_constraint = "AnyLocation"

  policy {
    region {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
    az {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
    hardware {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
    software {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
  }

  tags = {
    Name = %[1]q
	Value = "Other"
  }
}
`, rName, resTierValue)
}

func testAccResiliencyPolicyConfig_updatePolicy(rName, resPolicyObjValue string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehub_resiliency_policy" "test" {
  policy_name        = %[1]q

  policy_description = %[1]q

  tier = "NotApplicable"

  data_location_constraint = "AnyLocation"

  policy {
    region {
      rpo_in_secs = %[2]q
      rto_in_secs = %[2]q
    }
    az {
      rpo_in_secs = %[2]q
      rto_in_secs = %[2]q
    }
    hardware {
      rpo_in_secs = %[2]q
      rto_in_secs = %[2]q
    }
    software {
      rpo_in_secs = %[2]q
      rto_in_secs = %[2]q
    }
  }

  tags = {
    Name = %[1]q
	Value = "Other"
  }
}
`, rName, resPolicyObjValue)
}

func testAccResiliencyPolicyConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehub_resiliency_policy" "test" {
  policy_name        = %[1]q

  policy_description = %[1]q

  tier = "NotApplicable"

  data_location_constraint = "AnyLocation"

  policy {
    region {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
    az {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
    hardware {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
    software {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccResiliencyPolicyConfig_tag2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehub_resiliency_policy" "test" {
  policy_name        = %[1]q

  policy_description = %[1]q

  tier = "NotApplicable"

  data_location_constraint = "AnyLocation"

  policy {
    region {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
    az {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
    hardware {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
    software {
      rpo_in_secs = 3600
      rto_in_secs = 3600
    }
  }

  tags = {
    %[2]q = %[3]q
	%[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
