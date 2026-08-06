// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mailmanager_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/mailmanager"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfmailmanager "github.com/hashicorp/terraform-provider-aws/internal/service/mailmanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMailManagerIngressPoint_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:                testAccMailManagerIngressPoint_basic,
		acctest.CtDisappears:           testAccMailManagerIngressPoint_disappears,
		"update":                       testAccMailManagerIngressPoint_update,
		"tlsPolicy":                    testAccMailManagerIngressPoint_tlsPolicy,
		"type":                         testAccMailManagerIngressPoint_type,
		"networkConfiguration_public":  testAccMailManagerIngressPoint_networkConfiguration_public,
		"networkConfiguration_private": testAccMailManagerIngressPoint_networkConfiguration_private,
		"ingressPointConfiguration_smtpPasswordWO": testAccMailManagerIngressPoint_ingressPointConfiguration_smtpPasswordWO,
		"ingressPointConfiguration_tlsAuth":        testAccMailManagerIngressPoint_ingressPointConfiguration_tlsAuth,
		"Identity":                                 testAccMailManagerIngressPoint_identitySerial,
		"Tags":                                     testAccMailManagerIngressPoint_tagsSerial,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccMailManagerIngressPoint_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_ingress_point.test"
	trafficPolicyName := "aws_mailmanager_traffic_policy.test"
	ruleSetName := "aws_mailmanager_rule_set.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckIngressPoint(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngressPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIngressPointConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIngressPointExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ses", regexache.MustCompile(`mailmanager-ingress-point/.+`)),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_timestamp"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_updated_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "rule_set_id", ruleSetName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttrSet(resourceName, "a_record"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "OPEN"),
					resource.TestCheckResourceAttrPair(resourceName, "traffic_policy_id", trafficPolicyName, names.AttrID),
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

func testAccMailManagerIngressPoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_ingress_point.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckIngressPoint(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngressPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIngressPointConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIngressPointExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfmailmanager.ResourceIngressPoint, resourceName),
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

func testAccMailManagerIngressPoint_update(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_ingress_point.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckIngressPoint(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngressPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIngressPointConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIngressPointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "OPEN"),
				),
			},
			{
				Config: testAccIngressPointConfig_basic(rNameUpdated),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIngressPointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
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

func testAccMailManagerIngressPoint_tlsPolicy(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_ingress_point.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckIngressPoint(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngressPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIngressPointConfig_tlsPolicy(rName, "REQUIRED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIngressPointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tls_policy", "REQUIRED"),
				),
			},
			{
				Config: testAccIngressPointConfig_tlsPolicy(rName, "OPTIONAL"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIngressPointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tls_policy", "OPTIONAL"),
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

func testAccMailManagerIngressPoint_type(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_ingress_point.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckIngressPoint(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngressPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIngressPointConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIngressPointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "OPEN"),
				),
			},
			{
				// Changing type requires replacement, not an in-place update.
				Config: testAccIngressPointConfig_smtpPasswordWO(rName, "Password1!", 1),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIngressPointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "AUTH"),
				),
			},
		},
	})
}

func testAccMailManagerIngressPoint_networkConfiguration_public(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_ingress_point.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckIngressPoint(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngressPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIngressPointConfig_publicNetworkConfiguration(rName, "DUAL_STACK"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIngressPointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.public_network_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.public_network_configuration.0.ip_type", "DUAL_STACK"),
				),
			},
			{
				Config: testAccIngressPointConfig_publicNetworkConfiguration(rName, "IPV4"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIngressPointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.public_network_configuration.0.ip_type", "IPV4"),
				),
			},
		},
	})
}

func testAccMailManagerIngressPoint_networkConfiguration_private(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_ingress_point.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckIngressPoint(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngressPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIngressPointConfig_privateNetworkConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIngressPointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.private_network_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "network_configuration.0.private_network_configuration.0.vpc_endpoint_id"),
				),
			},
		},
	})
}

func testAccMailManagerIngressPoint_ingressPointConfiguration_smtpPasswordWO(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_ingress_point.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckIngressPoint(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngressPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIngressPointConfig_smtpPasswordWO(rName, "initialPassword1!", 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIngressPointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "AUTH"),
					resource.TestCheckResourceAttr(resourceName, "ingress_point_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ingress_point_configuration.0.smtp_password_wo_version", "1"),
				),
			},
			{
				// Rotate the password: bumping smtp_password_wo_version triggers an in-place update.
				// After rotation, the previous version fields are populated.
				Config: testAccIngressPointConfig_smtpPasswordWO(rName, "rotatedPassword2!", 2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIngressPointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ingress_point_configuration.0.smtp_password_wo_version", "2"),
				),
			},
		},
	})
}

func testAccMailManagerIngressPoint_ingressPointConfiguration_tlsAuth(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_mailmanager_ingress_point.test"
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCert := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, caKey)
	crl := acctest.TLSRSAX509CertificateRevocationListPEM(t, caKey, caCert)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckIngressPoint(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.MailManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngressPointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIngressPointConfig_tlsAuth(rName, caCert),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIngressPointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MTLS"),
					resource.TestCheckResourceAttr(resourceName, "ingress_point_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ingress_point_configuration.0.tls_auth_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ingress_point_configuration.0.tls_auth_configuration.0.trust_store.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "ingress_point_configuration.0.tls_auth_configuration.0.trust_store.0.ca_content"),
				),
			},
			{
				// Update: add crl_content to the existing trust store.
				Config: testAccIngressPointConfig_tlsAuthWithCRL(rName, caCert, crl),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIngressPointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "ingress_point_configuration.0.tls_auth_configuration.0.trust_store.0.ca_content"),
					resource.TestCheckResourceAttrSet(resourceName, "ingress_point_configuration.0.tls_auth_configuration.0.trust_store.0.crl_content"),
				),
			},
		},
	})
}

func testAccCheckIngressPointDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).MailManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_mailmanager_ingress_point" {
				continue
			}

			_, err := tfmailmanager.FindIngressPointByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.MailManager, create.ErrActionCheckingDestroyed, tfmailmanager.ResNameIngressPoint, rs.Primary.ID, err)
			}

			return create.Error(names.MailManager, create.ErrActionCheckingDestroyed, tfmailmanager.ResNameIngressPoint, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckIngressPointExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.MailManager, create.ErrActionCheckingExistence, tfmailmanager.ResNameIngressPoint, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.MailManager, create.ErrActionCheckingExistence, tfmailmanager.ResNameIngressPoint, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).MailManagerClient(ctx)

		_, err := tfmailmanager.FindIngressPointByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.MailManager, create.ErrActionCheckingExistence, tfmailmanager.ResNameIngressPoint, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccPreCheckIngressPoint(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).MailManagerClient(ctx)

	var input mailmanager.ListIngressPointsInput

	_, err := conn.ListIngressPoints(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccIngressPointConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_mailmanager_traffic_policy" "test" {
  default_action = "ALLOW"
  name           = %[1]q

  policy_statement {
    action = "DENY"

    condition {
      ip_expression {
        operator = "CIDR_MATCHES"
        values   = ["192.0.2.0/24"]

        evaluate {
          attribute = "SENDER_IP"
        }
      }
    }
  }
}

resource "aws_mailmanager_rule_set" "test" {
  name = %[1]q

  rule {
    action {
      add_header {
        header_name  = "X-Test"
        header_value = "example"
      }
    }
  }
}
`, rName)
}

func testAccIngressPointConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccIngressPointConfigBase(rName),
		fmt.Sprintf(`
resource "aws_mailmanager_ingress_point" "test" {
  name              = %[1]q
  type              = "OPEN"
  rule_set_id       = aws_mailmanager_rule_set.test.id
  traffic_policy_id = aws_mailmanager_traffic_policy.test.id
}
`, rName))
}

func testAccIngressPointConfig_tlsPolicy(rName, tlsPolicy string) string {
	return acctest.ConfigCompose(
		testAccIngressPointConfigBase(rName),
		fmt.Sprintf(`
resource "aws_mailmanager_ingress_point" "test" {
  name              = %[1]q
  type              = "OPEN"
  rule_set_id       = aws_mailmanager_rule_set.test.id
  traffic_policy_id = aws_mailmanager_traffic_policy.test.id
  tls_policy        = %[2]q
}
`, rName, tlsPolicy))
}

func testAccIngressPointConfig_publicNetworkConfiguration(rName, ipType string) string {
	return acctest.ConfigCompose(
		testAccIngressPointConfigBase(rName),
		fmt.Sprintf(`
resource "aws_mailmanager_ingress_point" "test" {
  name              = %[1]q
  type              = "OPEN"
  rule_set_id       = aws_mailmanager_rule_set.test.id
  traffic_policy_id = aws_mailmanager_traffic_policy.test.id

  network_configuration {
    public_network_configuration {
      ip_type = %[2]q
    }
  }
}
`, rName, ipType))
}

func testAccIngressPointConfig_smtpPasswordWO(rName, password string, version int) string {
	return acctest.ConfigCompose(
		testAccIngressPointConfigBase(rName),
		fmt.Sprintf(`
resource "aws_mailmanager_ingress_point" "test" {
  name              = %[1]q
  type              = "AUTH"
  rule_set_id       = aws_mailmanager_rule_set.test.id
  traffic_policy_id = aws_mailmanager_traffic_policy.test.id

  ingress_point_configuration {
    smtp_password_wo         = %[2]q
    smtp_password_wo_version = %[3]d
  }
}
`, rName, password, version))
}

func testAccIngressPointConfig_privateNetworkConfiguration(rName string) string {
	return acctest.ConfigCompose(
		testAccIngressPointConfigBase(rName),
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_vpc_endpoint" "test" {
  vpc_id            = aws_vpc.test.id
  service_name      = "com.amazonaws.${data.aws_region.current.name}.mail-manager-smtp.open"
  vpc_endpoint_type = "Interface"

  subnet_ids = aws_subnet.test[*].id

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_mailmanager_ingress_point" "test" {
  name              = %[1]q
  type              = "OPEN"
  rule_set_id       = aws_mailmanager_rule_set.test.id
  traffic_policy_id = aws_mailmanager_traffic_policy.test.id
  tls_policy        = "REQUIRED"

  network_configuration {
    private_network_configuration {
      vpc_endpoint_id = aws_vpc_endpoint.test.id
    }
  }
}
`, rName))
}

func testAccIngressPointConfig_tlsAuth(rName, caCert string) string {
	return acctest.ConfigCompose(
		testAccIngressPointConfigBase(rName),
		fmt.Sprintf(`
resource "aws_mailmanager_ingress_point" "test" {
  name              = %[1]q
  type              = "MTLS"
  rule_set_id       = aws_mailmanager_rule_set.test.id
  traffic_policy_id = aws_mailmanager_traffic_policy.test.id

  ingress_point_configuration {
    tls_auth_configuration {
      trust_store {
        ca_content = %[2]q
      }
    }
  }
}
`, rName, caCert))
}

func testAccIngressPointConfig_tlsAuthWithCRL(rName, caCert, crl string) string {
	return acctest.ConfigCompose(
		testAccIngressPointConfigBase(rName),
		fmt.Sprintf(`
resource "aws_mailmanager_ingress_point" "test" {
  name              = %[1]q
  type              = "MTLS"
  rule_set_id       = aws_mailmanager_rule_set.test.id
  traffic_policy_id = aws_mailmanager_traffic_policy.test.id

  ingress_point_configuration {
    tls_auth_configuration {
      trust_store {
        ca_content  = %[2]q
        crl_content = %[3]q
      }
    }
  }
}
`, rName, caCert, crl))
}
