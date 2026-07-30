// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package acm_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfacm "github.com/hashicorp/terraform-provider-aws/internal/service/acm"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccACMACMEEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.AcmeEndpoint
	resourceName := "aws_acm_acme_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckACMEEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccACMEEndpointConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckACMEEndpointExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile(`acme-endpoint/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "authorization_behavior", string(awstypes.AcmeAuthorizationBehaviorPreApproved)),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority.0.public_certificate_authority.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority.0.public_certificate_authority.0.allowed_key_algorithms.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "certificate_authority.0.public_certificate_authority.0.allowed_key_algorithms.*", string(awstypes.PublicKeyAlgorithmRsa2048)),
					resource.TestCheckNoResourceAttr(resourceName, "certificate_tags.%"),
					resource.TestCheckResourceAttrSet(resourceName, "contact"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint_url"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(awstypes.AcmeEndpointStatusActive)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
			},
		},
	})
}

func TestAccACMACMEEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.AcmeEndpoint
	resourceName := "aws_acm_acme_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckACMEEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccACMEEndpointConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckACMEEndpointExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfacm.ResourceACMEEndpoint, resourceName),
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

func TestAccACMACMEEndpoint_contact(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.AcmeEndpoint
	resourceName := "aws_acm_acme_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckACMEEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccACMEEndpointConfig_contact(string(awstypes.AcmeContactRequired)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckACMEEndpointExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "contact", string(awstypes.AcmeContactRequired)),
				),
			},
			{
				Config: testAccACMEEndpointConfig_contact(string(awstypes.AcmeContactNotRequired)),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckACMEEndpointExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "contact", string(awstypes.AcmeContactNotRequired)),
				),
			},
		},
	})
}

func TestAccACMACMEEndpoint_allowedKeyAlgorithms(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.AcmeEndpoint
	resourceName := "aws_acm_acme_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckACMEEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccACMEEndpointConfig_allowedKeyAlgorithms(`["RSA_2048"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckACMEEndpointExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority.0.public_certificate_authority.0.allowed_key_algorithms.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "certificate_authority.0.public_certificate_authority.0.allowed_key_algorithms.*", string(awstypes.PublicKeyAlgorithmRsa2048)),
				),
			},
			{
				Config: testAccACMEEndpointConfig_allowedKeyAlgorithms(`["RSA_2048", "EC_prime256v1"]`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckACMEEndpointExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority.0.public_certificate_authority.0.allowed_key_algorithms.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "certificate_authority.0.public_certificate_authority.0.allowed_key_algorithms.*", string(awstypes.PublicKeyAlgorithmRsa2048)),
					resource.TestCheckTypeSetElemAttr(resourceName, "certificate_authority.0.public_certificate_authority.0.allowed_key_algorithms.*", string(awstypes.PublicKeyAlgorithmEcPrime256V1)),
				),
			},
			{
				// allowed_key_algorithms is Optional+Computed, so omitting it retains the configured value rather than clearing it.
				Config: testAccACMEEndpointConfig_noAllowedKeyAlgorithms(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckACMEEndpointExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority.0.public_certificate_authority.0.allowed_key_algorithms.#", "2"),
				),
			},
		},
	})
}

func TestAccACMACMEEndpoint_certificateTags(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.AcmeEndpoint
	resourceName := "aws_acm_acme_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckACMEEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccACMEEndpointConfig_certificateTags(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckACMEEndpointExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "certificate_tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "certificate_tags."+acctest.CtKey1, acctest.CtValue1),
				),
			},
			{
				// certificate_tags cannot be modified in place: UpdateAcmeEndpoint does not accept it.
				Config: testAccACMEEndpointConfig_certificateTags(acctest.CtKey1, acctest.CtValue1Updated),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckACMEEndpointExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "certificate_tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "certificate_tags."+acctest.CtKey1, acctest.CtValue1Updated),
				),
			},
			{
				Config: testAccACMEEndpointConfig_basic(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckACMEEndpointExists(ctx, t, resourceName, &v),
					resource.TestCheckNoResourceAttr(resourceName, "certificate_tags.%"),
				),
			},
		},
	})
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	t.Helper()

	conn := acctest.ProviderMeta(ctx, t).ACMClient(ctx)

	input := acm.ListAcmeEndpointsInput{}
	_, err := conn.ListAcmeEndpoints(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckACMEEndpointDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ACMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_acm_acme_endpoint" {
				continue
			}

			arn := rs.Primary.Attributes[names.AttrARN]

			_, err := tfacm.FindACMEEndpointByARN(ctx, conn, arn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return create.Error(names.ACM, create.ErrActionCheckingDestroyed, tfacm.ResNameACMEEndpoint, arn, err)
			}

			return create.Error(names.ACM, create.ErrActionCheckingDestroyed, tfacm.ResNameACMEEndpoint, arn, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckACMEEndpointExists(ctx context.Context, t *testing.T, n string, v *awstypes.AcmeEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return create.Error(names.ACM, create.ErrActionCheckingExistence, tfacm.ResNameACMEEndpoint, n, errors.New("not found"))
		}

		arn := rs.Primary.Attributes[names.AttrARN]
		if arn == "" {
			return create.Error(names.ACM, create.ErrActionCheckingExistence, tfacm.ResNameACMEEndpoint, n, errors.New("no ARN is set"))
		}

		conn := acctest.ProviderMeta(ctx, t).ACMClient(ctx)

		output, err := tfacm.FindACMEEndpointByARN(ctx, conn, arn)
		if err != nil {
			return create.Error(names.ACM, create.ErrActionCheckingExistence, tfacm.ResNameACMEEndpoint, arn, err)
		}

		*v = *output

		return nil
	}
}

func testAccACMEEndpointConfig_basic() string {
	return `
resource "aws_acm_acme_endpoint" "test" {
  authorization_behavior = "PRE_APPROVED"

  certificate_authority {
    public_certificate_authority {
      allowed_key_algorithms = ["RSA_2048"]
    }
  }
}
`
}

func testAccACMEEndpointConfig_contact(contact string) string {
	return fmt.Sprintf(`
resource "aws_acm_acme_endpoint" "test" {
  authorization_behavior = "PRE_APPROVED"
  contact                = %[1]q

  certificate_authority {
    public_certificate_authority {
      allowed_key_algorithms = ["RSA_2048"]
    }
  }
}
`, contact)
}

func testAccACMEEndpointConfig_allowedKeyAlgorithms(algorithms string) string {
	return fmt.Sprintf(`
resource "aws_acm_acme_endpoint" "test" {
  authorization_behavior = "PRE_APPROVED"

  certificate_authority {
    public_certificate_authority {
      allowed_key_algorithms = %[1]s
    }
  }
}
`, algorithms)
}

func testAccACMEEndpointConfig_noAllowedKeyAlgorithms() string {
	return `
resource "aws_acm_acme_endpoint" "test" {
  authorization_behavior = "PRE_APPROVED"

  certificate_authority {
    public_certificate_authority {}
  }
}
`
}

func testAccACMEEndpointConfig_certificateTags(tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_acm_acme_endpoint" "test" {
  authorization_behavior = "PRE_APPROVED"

  certificate_tags = {
    %[1]q = %[2]q
  }

  certificate_authority {
    public_certificate_authority {
      allowed_key_algorithms = ["RSA_2048"]
    }
  }
}
`, tagKey, tagValue)
}
