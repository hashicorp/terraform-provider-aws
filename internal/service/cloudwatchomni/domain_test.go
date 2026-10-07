// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudwatchomni "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatchomni"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccCloudWatchOmniDomain_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Domain
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatchomni_domain.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("cloudwatch", regexache.MustCompile(`domain/d-[0-9a-z]+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("custom_endpoint_urls"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("domain_endpoint_url"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("domain_id"), knownvalue.StringRegexp(regexache.MustCompile(`^d-[0-9a-z]+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("identity_center_application_arn"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("identity_provider_configuration"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("identity_providers"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact(string(awstypes.IdentityProviderIam)),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "domain_id"),
				ImportStateVerifyIdentifierAttribute: "domain_id",
			},
		},
	})
}

func testAccCloudWatchOmniDomain_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Domain
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatchomni_domain.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudwatchomni.ResourceDomain, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCloudWatchOmniDomain_name(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.Domain
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatchomni_domain.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, t, resourceName, &v1),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName1)),
				},
			},
			{
				Config: testAccDomainConfig_basic(rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, t, resourceName, &v2),
					testAccCheckDomainNotRecreated(&v1, &v2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("custom_endpoint_urls"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringRegexp(regexache.MustCompile(`^https://` + rName2 + `\.`)),
					})),
				},
			},
		},
	})
}

func testAccCloudWatchOmniDomain_identityCenter(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.Domain
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatchomni_domain.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, t, resourceName, &v1),
				),
			},
			{
				Config: testAccDomainConfig_identityCenter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDomainExists(ctx, t, resourceName, &v2),
					testAccCheckDomainNotRecreated(&v1, &v2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("identity_center_application_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("identity_provider_configuration"), knownvalue.ListSizeExact(1)),
					statecheck.CompareValuePairs(
						resourceName, tfjsonpath.New("identity_provider_configuration").AtSliceIndex(0).AtMapKey("identity_center_configuration").AtSliceIndex(0).AtMapKey("identity_center_instance_arn"),
						"data.aws_ssoadmin_instances.test", tfjsonpath.New(names.AttrARNs).AtSliceIndex(0),
						compare.ValuesSame(),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("identity_providers"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact(string(awstypes.IdentityProviderIam)),
						knownvalue.StringExact(string(awstypes.IdentityProviderIdc)),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "domain_id"),
				ImportStateVerifyIdentifierAttribute: "domain_id",
			},
		},
	})
}

func testAccCheckDomainDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudWatchOmniClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatchomni_domain" {
				continue
			}

			_, err := tfcloudwatchomni.FindDomainByID(ctx, conn, rs.Primary.Attributes["domain_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Omni Domain %s still exists", rs.Primary.Attributes["domain_id"])
		}

		return nil
	}
}

func testAccCheckDomainExists(ctx context.Context, t *testing.T, n string, v *awstypes.Domain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudWatchOmniClient(ctx)

		output, err := tfcloudwatchomni.FindDomainByID(ctx, conn, rs.Primary.Attributes["domain_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckDomainNotRecreated(before, after *awstypes.Domain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := before.DomainId, after.DomainId; *before != *after {
			return fmt.Errorf("CloudWatch Omni Domain recreated: %s -> %s", *before, *after)
		}

		return nil
	}
}

func testAccDomainConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatchomni_domain" "test" {
  name               = %[1]q
  identity_providers = ["IAM"]
}
`, rName)
}

func testAccDomainConfig_identityCenter(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_cloudwatchomni_domain" "test" {
  name               = %[1]q
  identity_providers = ["IAM", "IDC"]

  identity_provider_configuration {
    identity_center_configuration {
      identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
    }
  }
}
`, rName)
}
