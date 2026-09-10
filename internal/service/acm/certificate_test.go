// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package acm_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfacm "github.com/hashicorp/terraform-provider-aws/internal/service/acm"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccACMCertificate_AmazonIssued_emailValidation(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	var v types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic(domain, types.ValidationMethodEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "not_after", ""),
					resource.TestCheckResourceAttr(resourceName, "not_before", ""),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.export", string(types.CertificateExportDisabled)),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypeAmazonIssued)),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "validation_emails.#", 0),
					resource.TestMatchResourceAttr(resourceName, "validation_emails.0", regexache.MustCompile(`^[^@]+@.+$`)),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodEmail)),
					resource.TestCheckResourceAttr(resourceName, "validation_option.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New("pending_renewal")),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("subject_alternative_names"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact(domain),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccACMCertificate_AmazonIssued_dnsValidation(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	var v types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic(domain, types.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "not_after", ""),
					resource.TestCheckResourceAttr(resourceName, "not_before", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypeAmazonIssued)),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
					resource.TestCheckResourceAttr(resourceName, "validation_option.#", "0"),
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

func TestAccACMCertificate_AmazonIssued_root(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	var v types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic(rootDomain, types.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
					resource.TestCheckResourceAttr(resourceName, "validation_option.#", "0"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeAmazonIssued)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"not_after", "not_before", names.AttrStatus},
			},
		},
	})
}

func TestAccACMCertificate_AmazonIssued_validationOptions(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	var v types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_validationOptions(rootDomain, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domain),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "validation_emails.#", 0),
					resource.TestMatchResourceAttr(resourceName, "validation_emails.0", regexache.MustCompile(`^[^@]+@.+$`)),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodEmail)),
					resource.TestCheckResourceAttr(resourceName, "validation_option.#", "1"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeAmazonIssued)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"validation_option"},
			},
		},
	})
}

func TestAccACMCertificate_Private_renewable(t *testing.T) {
	ctx := acctest.Context(t)
	certificateAuthorityResourceName := "aws_acmpca_certificate_authority.test"
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain(t)
	certificateDomainName := commonName.RandomSubdomain(t).String()
	var v1, v2, v3 types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Create potentially renewable certificate
			// Ineligible for renewal because it has not been exported
			// See https://docs.aws.amazon.com/acm/latest/userguide/managed-renewal.html for details on certificate renewal
			{
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_authority_arn", certificateAuthorityResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_after"),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_before"),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "validation_option.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New("pending_renewal")),
					},
				},
			},
			// Step 2: Import
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
			// Step 3: Export to make certificate eligible for renewal
			// Because the early renewal date is unset, the certificate is not pending renewal.
			{
				PreConfig: func() {
					conn := acctest.ProviderMeta(ctx, t).ACMClient(ctx)

					input := acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					}
					_, err := conn.ExportCertificate(ctx, &input)
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.ToString(v1.CertificateArn), err)
					}
				},
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v2),
					testAccCheckCertificateNotRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityEligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("pending_renewal"), knownvalue.Bool(false)),
					},
				},
			},
			// Step 4: Renew the certificate out-of-band
			// This will reset the renewal eligiblity and pending renewal status
			{
				PreConfig: func() {
					conn := acctest.ProviderMeta(ctx, t).ACMClient(ctx)

					input := acm.RenewCertificateInput{
						CertificateArn: v1.CertificateArn,
					}
					_, err := conn.RenewCertificate(ctx, &input)
					if err != nil {
						t.Fatalf("renewing ACM Certificate (%s): %s", aws.ToString(v1.CertificateArn), err)
					}

					_, err = tfacm.WaitCertificateRenewed(ctx, conn, aws.ToString(v1.CertificateArn), tfacm.CertificateRenewalTimeout)
					if err != nil {
						t.Fatalf("waiting for ACM Certificate (%s) renewal: %s", aws.ToString(v1.CertificateArn), err)
					}
				},
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status", string(types.RenewalStatusSuccess)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status_reason", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "renewal_summary.0.updated_at"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("pending_renewal"), knownvalue.Bool(false)),
					},
				},
			},
		},
	})
}

func TestAccACMCertificate_Private_noRenewalPermission(t *testing.T) {
	ctx := acctest.Context(t)
	certificateAuthorityResourceName := "aws_acmpca_certificate_authority.test"
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain(t)
	certificateDomainName := commonName.RandomSubdomain(t).String()
	var v1, v2, v3 types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateCertificate_noRenewalPermission(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_authority_arn", certificateAuthorityResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_after"),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_before"),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "validation_option.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				PreConfig: func() {
					conn := acctest.ProviderMeta(ctx, t).ACMClient(ctx)

					input := acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					}
					_, err := conn.ExportCertificate(ctx, &input)
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.ToString(v1.CertificateArn), err)
					}
				},
				Config: testAccCertificateConfig_privateCertificate_noRenewalPermission(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityEligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				PreConfig: func() {
					conn := acctest.ProviderMeta(ctx, t).ACMClient(ctx)

					input := acm.RenewCertificateInput{
						CertificateArn: v1.CertificateArn,
					}
					_, err := conn.RenewCertificate(ctx, &input)
					if err != nil {
						t.Fatalf("renewing ACM Certificate (%s): %s", aws.ToString(v1.CertificateArn), err)
					}
				},
				Config: testAccCertificateConfig_privateCertificate_noRenewalPermission(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityEligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status", string(types.RenewalStatusPendingAutoRenewal)),
					acctest.CheckResourceAttrRFC3339(resourceName, "renewal_summary.0.updated_at"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
			},
			{
				PreConfig: func() {
					time.Sleep(tfacm.CertificateRenewalTimeout)
				},
				Config: testAccCertificateConfig_privateCertificate_noRenewalPermission(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityEligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status", string(types.RenewalStatusPendingAutoRenewal)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status_reason", string(types.FailureReasonPcaAccessDenied)),
					acctest.CheckResourceAttrRFC3339(resourceName, "renewal_summary.0.updated_at"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
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

func TestAccACMCertificate_Private_pendingRenewalGoDuration(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain(t)
	certificateDomainName := commonName.RandomSubdomain(t).String()
	duration := (395 * 24 * time.Hour).String()
	var v1, v2 types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Create with `early_renewal_duration`
			// Ineligible for renewal because it has not been exported
			// See https://docs.aws.amazon.com/acm/latest/userguide/managed-renewal.html for details on certificate renewal
			{
				Config: testAccCertificateConfig_privateCertificate_earlyRenewalDuration(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New("pending_renewal")),
					},
				},
			},
			// Step 2: Import
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
			// Step 3: Export to make certificate eligible for renewal
			// Plan is non-empty to trigger Update on subsequent apply
			{
				PreConfig: func() {
					conn := acctest.ProviderMeta(ctx, t).ACMClient(ctx)

					input := acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					}
					_, err := conn.ExportCertificate(ctx, &input)
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.ToString(v1.CertificateArn), err)
					}
				},
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityEligible)),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
				RefreshPlanChecks: resource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New("pending_renewal")),
					},
				},
				ExpectNonEmptyPlan: true,
			},
			// Step 4: Update renews the certificate as `pending_renewal` is true
			// This will reset the renewal eligiblity and pending renewal status
			{
				Config: testAccCertificateConfig_privateCertificate_earlyRenewalDuration(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v2),
					testAccCheckCertificateRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status", string(types.RenewalStatusSuccess)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status_reason", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "renewal_summary.0.updated_at"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New("pending_renewal")),
					},
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
		},
	})
}

func TestAccACMCertificate_Private_pendingRenewalRFC3339Duration(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain(t)
	certificateDomainName := commonName.RandomSubdomain(t).String()
	duration := "P395D"
	var v1, v2 types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Create with `early_renewal_duration`
			// Ineligible for renewal because it has not been exported
			// See https://docs.aws.amazon.com/acm/latest/userguide/managed-renewal.html for details on certificate renewal
			{
				Config: testAccCertificateConfig_privateCertificate_earlyRenewalDuration(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New("pending_renewal")),
					},
				},
			},
			// Step 2: Import
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
			// Step 3: Export to make certificate eligible for renewal
			// Plan is non-empty to trigger Update on subsequent apply
			{
				PreConfig: func() {
					conn := acctest.ProviderMeta(ctx, t).ACMClient(ctx)

					input := acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					}
					_, err := conn.ExportCertificate(ctx, &input)
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.ToString(v1.CertificateArn), err)
					}
				},
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityEligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
				RefreshPlanChecks: resource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New("pending_renewal")),
					},
				},
				ExpectNonEmptyPlan: true,
			},
			// Step 4: Update renews the certificate as `pending_renewal` is true
			// This will reset the renewal eligiblity and pending renewal status
			{
				Config: testAccCertificateConfig_privateCertificate_earlyRenewalDuration(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v2),
					testAccCheckCertificateRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status", string(types.RenewalStatusSuccess)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status_reason", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "renewal_summary.0.updated_at"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New("pending_renewal")),
					},
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
		},
	})
}

// This certificate will be eligible because it is exported
// See https://docs.aws.amazon.com/acm/latest/userguide/managed-renewal.html for details on certificate renewal
func TestAccACMCertificate_Private_addEarlyRenewalPast(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain(t)
	certificateDomainName := commonName.RandomSubdomain(t).String()
	duration := (395 * 24 * time.Hour).String()
	var v1, v2, v3 types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Create potentially renewable certificate
			// Ineligible for renewal because it has not been exported
			// See https://docs.aws.amazon.com/acm/latest/userguide/managed-renewal.html for details on certificate renewal
			{
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						// plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("pending_renewal"), knownvalue.Bool(false)),
					},
				},
			},
			// Step 2: Import
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
			// Step 3: Export to make certificate eligible for renewal
			{
				PreConfig: func() {
					conn := acctest.ProviderMeta(ctx, t).ACMClient(ctx)

					input := acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					}
					_, err := conn.ExportCertificate(ctx, &input)
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.ToString(v1.CertificateArn), err)
					}
				},
				// Ideally, we'd have a `RefreshOnly` test step here to validate that `pending_renewal` is false and `renewal_eligibility` is `ELIGIBLE` after exporting.
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/1069
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v2),
					testAccCheckCertificateNotRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityEligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("pending_renewal"), knownvalue.Bool(false)),
					},
				},
			},
			// Step 4: Update renews the certificate as `pending_renewal` is true
			// This will reset the renewal eligiblity and pending renewal status
			{
				Config: testAccCertificateConfig_privateCertificate_earlyRenewalDuration(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v3),
					testAccCheckCertificateRenewed(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status", string(types.RenewalStatusSuccess)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status_reason", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "renewal_summary.0.updated_at"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("pending_renewal"), knownvalue.Bool(true)),
					},
				},
			},
		},
	})
}

// This certificate will be ineligible because it has not been exported or otherwise made eligible
// See https://docs.aws.amazon.com/acm/latest/userguide/managed-renewal.html for details on certificate renewal
func TestAccACMCertificate_Private_addEarlyRenewalPastIneligible(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain(t)
	certificateDomainName := commonName.RandomSubdomain(t).String()
	duration := (395 * 24 * time.Hour).String()
	var v1, v2 types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Create potentially renewable certificate
			// Ineligible for renewal because it has not been exported
			// See https://docs.aws.amazon.com/acm/latest/userguide/managed-renewal.html for details on certificate renewal
			{
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
			},
			// Step 2: Update does not trigger renewal because the certificate is not eligible for renewal
			{
				Config: testAccCertificateConfig_privateCertificate_earlyRenewalDuration(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v2),
					testAccCheckCertificateNotRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
			},
		},
	})
}

func TestAccACMCertificate_Private_addEarlyRenewalFuture(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain(t)
	certificateDomainName := commonName.RandomSubdomain(t).String()
	duration := (90 * 24 * time.Hour).String()
	var v1, v2, v3 types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Create potentially renewable certificate
			// Ineligible for renewal because it has not been exported
			// See https://docs.aws.amazon.com/acm/latest/userguide/managed-renewal.html for details on certificate renewal
			{
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
			},
			// Step 2: Import
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
			// Step 3: Export to make certificate eligible for renewal
			// Because the early renewal date is unset, the certificate is not pending renewal.
			{
				PreConfig: func() {
					conn := acctest.ProviderMeta(ctx, t).ACMClient(ctx)

					input := acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					}
					_, err := conn.ExportCertificate(ctx, &input)
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.ToString(v1.CertificateArn), err)
					}
				},
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v2),
					testAccCheckCertificateNotRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityEligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
				RefreshPlanChecks: resource.RefreshPlanChecks{
					PostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("pending_renewal"), knownvalue.Bool(false)),
					},
				},
				ExpectNonEmptyPlan: true,
			},
			// Step 4: The renewal window is in the future, so it is not renewed
			{
				Config: testAccCertificateConfig_privateCertificate_earlyRenewalDuration(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v3),
					testAccCheckCertificateNotRenewed(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityEligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("pending_renewal"), knownvalue.Bool(false)),
					},
				},
			},
		},
	})
}

func TestAccACMCertificate_Private_updateEarlyRenewalFuture(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain(t)
	certificateDomainName := commonName.RandomSubdomain(t).String()
	duration := (395 * 24 * time.Hour).String()
	durationUpdated := (90 * 24 * time.Hour).String()
	var v1, v2 types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateCertificate_earlyRenewalDuration(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
			// Export to make certificate eligible for renewal
			{
				PreConfig: func() {
					conn := acctest.ProviderMeta(ctx, t).ACMClient(ctx)

					input := acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					}
					_, err := conn.ExportCertificate(ctx, &input)
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.ToString(v1.CertificateArn), err)
					}
				},
				// Ideally, we'd have a `RefreshOnly` test step here to validate that `pending_renewal` is false and `renewal_eligibility` is `ELIGIBLE` after exporting.
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/1069
				Config: testAccCertificateConfig_privateCertificate_earlyRenewalDuration(commonName.String(), certificateDomainName, durationUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v2),
					testAccCheckCertificateNotRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", durationUpdated),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityEligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
		},
	})
}

func TestAccACMCertificate_Private_removeEarlyRenewal(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain(t)
	certificateDomainName := commonName.RandomSubdomain(t).String()
	duration := (395 * 24 * time.Hour).String()
	var v1, v2 types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateCertificate_earlyRenewalDuration(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
			// Export to make certificate eligible for renewal
			{
				PreConfig: func() {
					conn := acctest.ProviderMeta(ctx, t).ACMClient(ctx)

					input := acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					}
					_, err := conn.ExportCertificate(ctx, &input)
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.ToString(v1.CertificateArn), err)
					}
				},
				// Ideally, we'd have a `RefreshOnly` test step here to validate that `pending_renewal` is false and `renewal_eligibility` is `ELIGIBLE` after exporting.
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/1069
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v2),
					testAccCheckCertificateNotRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityEligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"early_renewal_duration"},
			},
		},
	})
}

// TestAccACMCertificate_AmazonIssued_Root_trailingPeriod updated in 3.0 to account for domain_name plan-time validation
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13510
func TestAccACMCertificate_AmazonIssued_Root_trailingPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := fmt.Sprintf("%s.", rootDomain)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccCertificateConfig_basic(domain, types.ValidationMethodDns),
				ExpectError: regexache.MustCompile(`invalid value for domain_name \(cannot end with a period\)`),
			},
		},
	})
}

func TestAccACMCertificate_AmazonIssued_rootAndWildcardSan(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)
	var v types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(rootDomain, strconv.Quote(wildcardDomain), types.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   wildcardDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", wildcardDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeAmazonIssued)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"not_after", "not_before", names.AttrStatus},
			},
		},
	})
}

func TestAccACMCertificate_AmazonIssued_SubjectAlternativeNames_emptyString(t *testing.T) {
	ctx := acctest.Context(t)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccCertificateConfig_subjectAlternativeNames(domain, strconv.Quote(""), types.ValidationMethodDns),
				ExpectError: regexache.MustCompile(`expected length`),
			},
		},
	})
}

func TestAccACMCertificate_AmazonIssued_San_single(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	sanDomain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	sanDomainUpdated := acctest.ACMCertificateRandomSubDomain(rootDomain)
	var v types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(domain, strconv.Quote(sanDomain), types.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   sanDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeAmazonIssued)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("subject_alternative_names"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact(domain),
						knownvalue.StringExact(sanDomain),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(domain, strconv.Quote(sanDomainUpdated), types.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   sanDomainUpdated,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("subject_alternative_names"), knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact(domain),
							knownvalue.StringExact(sanDomainUpdated),
						})),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeAmazonIssued)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("subject_alternative_names"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact(domain),
						knownvalue.StringExact(sanDomainUpdated),
					})),
				},
			},
		},
	})
}

func TestAccACMCertificate_AmazonIssued_San_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	sanDomain1 := acctest.ACMCertificateRandomSubDomain(rootDomain)
	sanDomain2 := acctest.ACMCertificateRandomSubDomain(rootDomain)
	var v types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(domain, fmt.Sprintf("%q, %q", sanDomain1, sanDomain2), types.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   sanDomain1,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   sanDomain2,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domain),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", sanDomain1),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", sanDomain2),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeAmazonIssued)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccACMCertificate_AmazonIssued_San_trailingPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	sanDomain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	resourceName := "aws_acm_certificate.test"
	var v types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(domain, strconv.Quote(sanDomain), types.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile(`certificate/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   strings.TrimSuffix(sanDomain, "."),
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domain),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", strings.TrimSuffix(sanDomain, ".")),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeAmazonIssued)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccACMCertificate_AmazonIssued_San_matches_domain(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	sanDomain := rootDomain
	var v types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(domain, strconv.Quote(sanDomain), types.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile(`certificate/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   sanDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domain),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", sanDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeAmazonIssued)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccACMCertificate_AmazonIssued_wildcard(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)
	var v types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic(wildcardDomain, types.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, wildcardDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   wildcardDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", wildcardDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeAmazonIssued)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"not_after", "not_before", names.AttrStatus},
			},
		},
	})
}

func TestAccACMCertificate_AmazonIssued_wildcardAndRootSan(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)
	var v types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(wildcardDomain, strconv.Quote(rootDomain), types.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, wildcardDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   wildcardDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", wildcardDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeAmazonIssued)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"not_after", "not_before", names.AttrStatus},
			},
		},
	})
}

func TestAccACMCertificate_AmazonIssued_keyAlgorithm(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	var v types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_keyAlgorithm(rootDomain, types.ValidationMethodDns, types.KeyAlgorithmEcPrime256v1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
					resource.TestCheckResourceAttr(resourceName, "key_algorithm", string(types.KeyAlgorithmEcPrime256v1)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeAmazonIssued)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"not_after", "not_before", names.AttrStatus},
			},
		},
	})
}

func TestAccACMCertificate_AmazonIssued_disableCTLogging(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	var v types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_disableCTLogging(rootDomain, types.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.certificate_transparency_logging_preference", string(types.CertificateTransparencyLoggingPreferenceDisabled)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeAmazonIssued)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"not_after", "not_before", names.AttrStatus},
			},
		},
	})
}

func TestAccACMCertificate_AmazonIssued_disableReenableCTLogging(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	var v types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_optionsWithValidation(rootDomain, types.ValidationMethodDns, "ENABLED"),
				Check:  testAccCheckCertificateExists(ctx, t, resourceName, &v),
			},
			// Check the certificate's attributes once the validation has been applied.
			{
				Config: testAccCertificateConfig_optionsWithValidation(rootDomain, types.ValidationMethodDns, "ENABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.certificate_transparency_logging_preference", string(types.CertificateTransparencyLoggingPreferenceEnabled)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeAmazonIssued)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCertificateConfig_optionsWithValidation(rootDomain, types.ValidationMethodDns, "DISABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.certificate_transparency_logging_preference", string(types.CertificateTransparencyLoggingPreferenceDisabled)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeAmazonIssued)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCertificateConfig_optionsWithValidation(rootDomain, types.ValidationMethodDns, "ENABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.certificate_transparency_logging_preference", string(types.CertificateTransparencyLoggingPreferenceEnabled)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeAmazonIssued)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// lintignore:AT002
func TestAccACMCertificate_Imported_PrivateKeyWo(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	commonName := "example.com"
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, caKey)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509LocallySignedCertificatePEM(t, caKey, caCertificate, key, commonName)
	newCaKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	newCaCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, newCaKey)
	newKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	newCertificate := acctest.TLSRSAX509LocallySignedCertificatePEM(t, newCaKey, newCaCertificate, newKey, commonName)
	var v1, v2 types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateKeyWo(certificate, key),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v1),
					resource.TestCheckNoResourceAttr(resourceName, "private_key_wo"),
					resource.TestCheckResourceAttr(resourceName, "private_key_wo_version", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New("pending_renewal")),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeImported)),
				},
			},
			{
				Config: testAccCertificateConfig_privateKeyWoUpdate(newCertificate, newKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v2),
					testAccCheckCertificateNotRecreated(&v1, &v2),
					resource.TestCheckNoResourceAttr(resourceName, "private_key_wo"),
					resource.TestCheckResourceAttr(resourceName, "private_key_wo_version", "2"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("pending_renewal"), knownvalue.Bool(false)),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeImported)),
				},
			},
		},
	})
}

// lintignore:AT002
func TestAccACMCertificate_Imported_domainName(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	commonName := "example.com"
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, caKey)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509LocallySignedCertificatePEM(t, caKey, caCertificate, key, commonName)
	newCaKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	newCaCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, newCaKey)
	newCertificate := acctest.TLSRSAX509LocallySignedCertificatePEM(t, newCaKey, newCaCertificate, key, commonName)
	withoutChainDomain := acctest.RandomDomainName(t)
	var v1, v2, v3 types.CertificateDetail

	arnNoChange := statecheck.CompareValue(compare.ValuesSame())
	certificateBodyExpectChange := statecheck.CompareValue(compare.ValuesDiffer())
	privateKeyExpectChange := statecheck.CompareValue(compare.ValuesDiffer())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateKey(certificate, key, caCertificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v1),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_after"),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_before"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New("pending_renewal")),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("acm", regexache.MustCompile("certificate/.+$"))),
					arnNoChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("certificate_authority_arn"), knownvalue.StringExact("")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("certificate_body"), knownvalue.StringExact(certificate)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrCertificateChain), knownvalue.StringExact(caCertificate)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDomainName), knownvalue.StringExact(commonName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("domain_validation_options"), knownvalue.SetSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("early_renewal_duration"), knownvalue.StringExact("")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_algorithm"), tfknownvalue.StringExact(types.KeyAlgorithmRsa2048)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("options"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"certificate_transparency_logging_preference": tfknownvalue.StringExact(types.CertificateTransparencyLoggingPreferenceDisabled),
							"export": tfknownvalue.StringExact(types.CertificateExportDisabled),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("pending_renewal"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPrivateKey), knownvalue.StringExact(key)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("private_key_wo"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("private_key_wo_version"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("renewal_eligibility"), tfknownvalue.StringExact(types.RenewalEligibilityIneligible)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("renewal_summary"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(types.CertificateStatusIssued)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("subject_alternative_names"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact(commonName),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeImported)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_emails"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_method"), tfknownvalue.StringExact("NONE")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("validation_option"), knownvalue.SetSizeExact(0)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// These are not returned by the API
				ImportStateVerifyIgnore: []string{names.AttrPrivateKey, "certificate_body", names.AttrCertificateChain},
			},
			{
				Config: testAccCertificateConfig_privateKey(newCertificate, key, newCaCertificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New(names.AttrDomainName)),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("pending_renewal"), knownvalue.Bool(false)),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New("subject_alternative_names")),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					arnNoChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("certificate_body"), knownvalue.StringExact(newCertificate)),
					certificateBodyExpectChange.AddStateValue(resourceName, tfjsonpath.New("certificate_body")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrCertificateChain), knownvalue.StringExact(newCaCertificate)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDomainName), knownvalue.StringExact(commonName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPrivateKey), knownvalue.StringExact(key)),
					privateKeyExpectChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrPrivateKey)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(types.CertificateStatusIssued)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("subject_alternative_names"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact(commonName),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeImported)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// These are not returned by the API
				ImportStateVerifyIgnore: []string{names.AttrPrivateKey, "certificate_body", names.AttrCertificateChain},
			},
			{
				Config: testAccCertificateConfig_privateKeyNoChain(t, withoutChainDomain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v3),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New(names.AttrDomainName)),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("pending_renewal"), knownvalue.Bool(false)),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New("subject_alternative_names")),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					arnNoChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrARN)),
					certificateBodyExpectChange.AddStateValue(resourceName, tfjsonpath.New("certificate_body")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrCertificateChain), knownvalue.StringExact("")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDomainName), knownvalue.StringExact(withoutChainDomain)),
					privateKeyExpectChange.AddStateValue(resourceName, tfjsonpath.New(names.AttrPrivateKey)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(types.CertificateStatusIssued)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("subject_alternative_names"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact(withoutChainDomain),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeImported)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// These are not returned by the API
				ImportStateVerifyIgnore: []string{names.AttrPrivateKey, "certificate_body", names.AttrCertificateChain},
			},
		},
	})
}

// lintignore:AT002
func TestAccACMCertificate_Imported_validityDates(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	commonName := "example.com"
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, caKey)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509LocallySignedCertificatePEM(t, caKey, caCertificate, key, commonName)

	var v types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateKey(certificate, key, caCertificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_after"),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_before"),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypeImported)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// These are not returned by the API
				ImportStateVerifyIgnore: []string{names.AttrPrivateKey, "certificate_body", names.AttrCertificateChain},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7103
// lintignore:AT002
func TestAccACMCertificate_Imported_ipAddress(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	var v types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateKeyNoChain(t, "1.2.3.4"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("certificate_authority_arn"), knownvalue.StringExact("")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("certificate_body"), knownvalue.StringRegexp(regexache.MustCompile(`.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrCertificateChain), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDomainName), knownvalue.StringExact("1.2.3.4")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPrivateKey), knownvalue.StringRegexp(regexache.MustCompile(`.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(types.CertificateStatusIssued)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("subject_alternative_names"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("1.2.3.4"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeImported)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// These are not returned by the API
				ImportStateVerifyIgnore: []string{names.AttrPrivateKey, "certificate_body"},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/15055
// lintignore:AT002
func TestAccACMCertificate_Imported_ReimportWithTags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	privateKeyPEM1 := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificatePEM1 := acctest.TLSRSAX509SelfSignedCertificatePEM(t, privateKeyPEM1, "1.2.3.4")

	privateKeyPEM2 := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificatePEM2 := acctest.TLSRSAX509SelfSignedCertificatePEM(t, privateKeyPEM2, "5.6.7.8")
	var v types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Certificate/tags/"),
				ConfigVariables: config.Variables{
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
					acctest.CtCertificatePEM: config.StringVariable(certificatePEM1),
					acctest.CtPrivateKeyPEM:  config.StringVariable(privateKeyPEM1),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeImported)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Certificate/tags/"),
				ConfigVariables: config.Variables{
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
					acctest.CtCertificatePEM: config.StringVariable(certificatePEM1),
					acctest.CtPrivateKeyPEM:  config.StringVariable(privateKeyPEM1),
				},
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"certificate_body", names.AttrPrivateKey,
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Certificate/tags/"),
				ConfigVariables: config.Variables{
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1Updated),
					}),
					acctest.CtCertificatePEM: config.StringVariable(certificatePEM2),
					acctest.CtPrivateKeyPEM:  config.StringVariable(privateKeyPEM2),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeImported)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Certificate/tags/"),
				ConfigVariables: config.Variables{
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1Updated),
					}),
					acctest.CtCertificatePEM: config.StringVariable(certificatePEM2),
					acctest.CtPrivateKeyPEM:  config.StringVariable(privateKeyPEM2),
				},
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"certificate_body", names.AttrPrivateKey,
				},
			},
		},
	})
}

func TestAccACMCertificate_AmazonIssued_optionExport(t *testing.T) {
	// Issuing an exportable ACM Certificate is expensive.
	// Skip the test by default and only run if the environment variable is set.
	acctest.SkipIfEnvVarNotSet(t, "ACM_TEST_CERTIFICATE_EXPORT")
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	var v types.CertificateDetail

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_optionExport(rootDomain, types.ValidationMethodDns, types.CertificateExportEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.export", string(types.CertificateExportEnabled)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(types.CertificateTypeAmazonIssued)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckCertificateExists(ctx context.Context, t *testing.T, n string, v *types.CertificateDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ACMClient(ctx)

		output, err := tfacm.FindCertificateByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckCertificateDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ACMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_acm_certificate" {
				continue
			}

			_, err := tfacm.FindCertificateByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ACM Certificate %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCertificateNotRecreated(v1, v2 *types.CertificateDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(v1.CertificateArn) != aws.ToString(v2.CertificateArn) {
			return fmt.Errorf("ACM Certificate recreated")
		}
		return nil
	}
}

func testAccCheckCertificateRenewed(i, j *types.CertificateDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.ToTime(j.NotAfter).After(aws.ToTime(i.NotAfter)) {
			return fmt.Errorf("ACM Certificate not renewed: i.NotAfter=%q, j.NotAfter=%q", aws.ToTime(i.NotAfter), aws.ToTime(j.NotAfter))
		}

		return nil
	}
}

func testAccCheckCertificateNotRenewed(i, j *types.CertificateDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.ToTime(j.NotAfter).Equal(aws.ToTime(i.NotAfter)) {
			return fmt.Errorf("ACM Certificate renewed: i.NotAfter=%q, j.NotAfter=%q", aws.ToTime(i.NotAfter), aws.ToTime(j.NotAfter))
		}

		return nil
	}
}

func testAccCertificateConfig_basic(domainName string, validationMethod types.ValidationMethod) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name       = %[1]q
  validation_method = %[2]q
}
`, domainName, validationMethod)
}

func testAccCertificateConfig_validationOptions(rootDomainName, domainName string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name       = %[2]q
  validation_method = "EMAIL"

  validation_option {
    domain_name       = %[2]q
    validation_domain = %[1]q
  }
}
`, rootDomainName, domainName)
}

func testAccCertificateConfig_privateCertificateBase(commonName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 2
  }
}

resource "aws_acmpca_certificate_authority_certificate" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn

  certificate       = aws_acmpca_certificate.test.certificate
  certificate_chain = aws_acmpca_certificate.test.certificate_chain
}

data "aws_partition" "current" {}
`, commonName)
}

func testAccCertificateConfig_privateCertificate_renewable(commonName, certificateDomainName string) string {
	return acctest.ConfigCompose(testAccCertificateConfig_privateCertificateBase(commonName), fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name               = %[1]q
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn

  depends_on = [
    aws_acmpca_certificate_authority_certificate.test,
    aws_acmpca_permission.test,
  ]
}

resource "aws_acmpca_permission" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
  principal                 = "acm.amazonaws.com"
  actions                   = ["IssueCertificate", "GetCertificate", "ListPermissions"]
}
`, certificateDomainName))
}

func testAccCertificateConfig_privateCertificate_noRenewalPermission(commonName, certificateDomainName string) string {
	return acctest.ConfigCompose(testAccCertificateConfig_privateCertificateBase(commonName), fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name               = %[1]q
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn

  depends_on = [
    aws_acmpca_certificate_authority_certificate.test,
  ]
}
`, certificateDomainName))
}

func testAccCertificateConfig_privateCertificate_earlyRenewalDuration(commonName, certificateDomainName, duration string) string {
	return acctest.ConfigCompose(testAccCertificateConfig_privateCertificateBase(commonName), fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name               = %[1]q
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn

  early_renewal_duration = %[2]q

  depends_on = [
    aws_acmpca_certificate_authority_certificate.test,
    aws_acmpca_permission.test,
  ]
}

resource "aws_acmpca_permission" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
  principal                 = "acm.amazonaws.com"
  actions                   = ["IssueCertificate", "GetCertificate", "ListPermissions"]
}
`, certificateDomainName, duration))
}

func testAccCertificateConfig_subjectAlternativeNames(domainName, subjectAlternativeNames string, validationMethod types.ValidationMethod) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name               = %[1]q
  subject_alternative_names = [%[2]s]
  validation_method         = %[3]q
}
`, domainName, subjectAlternativeNames, validationMethod)
}

func testAccCertificateConfig_privateKeyNoChain(t *testing.T, commonName string) string {
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, commonName)

	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"
}
`, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccCertificateConfig_privateKey(certificate, privateKey, chain string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body  = "%[1]s"
  private_key       = "%[2]s"
  certificate_chain = "%[3]s"
}
`, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(privateKey), acctest.TLSPEMEscapeNewlines(chain))
}

func testAccCertificateConfig_privateKeyWo(certificate, privateKey string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body       = "%[1]s"
  private_key_wo         = "%[2]s"
  private_key_wo_version = 1
}
`, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(privateKey))
}

func testAccCertificateConfig_privateKeyWoUpdate(certificate, privateKey string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body       = "%[1]s"
  private_key_wo         = "%[2]s"
  private_key_wo_version = 2
}
`, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(privateKey))
}

func testAccCertificateConfig_disableCTLogging(domainName string, validationMethod types.ValidationMethod) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name       = %[1]q
  validation_method = %[2]q

  options {
    certificate_transparency_logging_preference = "DISABLED"
  }
}
`, domainName, validationMethod)
}

func testAccCertificateConfig_optionsWithValidation(domainName string, validationMethod types.ValidationMethod, loggingPreference string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name       = %[1]q
  validation_method = %[2]q

  options {
    certificate_transparency_logging_preference = %[3]q
  }
}

data "aws_route53_zone" "test" {
  name         = %[1]q
  private_zone = false
}

# for_each acceptance testing requires SDKv2
#
# resource "aws_route53_record" "test" {
# 	for_each = {
# 	  for dvo in aws_acm_certificate.test.domain_validation_options : dvo.domain_name => {
# 		name   = dvo.resource_record_name
# 		record = dvo.resource_record_value
# 		type   = dvo.resource_record_type
# 	  }
# 	}

# 	allow_overwrite = true
# 	name            = each.value.name
# 	records         = [each.value.record]
# 	ttl             = 60
# 	type            = each.value.type
# 	zone_id         = data.aws_route53_zone.test.zone_id
# }

resource "aws_route53_record" "test" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_name
  records         = [tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_acm_certificate_validation" "test" {
  depends_on      = [aws_route53_record.test]
  certificate_arn = aws_acm_certificate.test.arn
}
`, domainName, validationMethod, loggingPreference)
}

func testAccCertificateConfig_keyAlgorithm(domainName string, validationMethod types.ValidationMethod, keyAlgorithm types.KeyAlgorithm) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name       = %[1]q
  validation_method = %[2]q
  key_algorithm     = %[3]q
}
`, domainName, validationMethod, keyAlgorithm)
}

func testAccCertificateConfig_optionExport(domainName string, validationMethod types.ValidationMethod, export types.CertificateExport) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name       = %[1]q
  validation_method = %[2]q

  options {
    export = %[3]q
  }
}
`, domainName, validationMethod, export)
}
