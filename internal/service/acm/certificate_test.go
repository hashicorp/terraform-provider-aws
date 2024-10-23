// Copyright (c) HashiCorp, Inc.
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
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfacm "github.com/hashicorp/terraform-provider-aws/internal/service/acm"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccACMCertificate_emailValidation(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	var v types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic(domain, types.ValidationMethodEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "not_after", ""),
					resource.TestCheckResourceAttr(resourceName, "not_before", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypeAmazonIssued)),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct0),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "validation_emails.#", 0),
					resource.TestMatchResourceAttr(resourceName, "validation_emails.0", regexache.MustCompile(`^[^@]+@.+$`)),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodEmail)),
					resource.TestCheckResourceAttr(resourceName, "validation_option.#", acctest.Ct0),
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

func TestAccACMCertificate_dnsValidation(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	var v types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic(domain, types.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "not_after", ""),
					resource.TestCheckResourceAttr(resourceName, "not_before", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domain),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypeAmazonIssued)),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
					resource.TestCheckResourceAttr(resourceName, "validation_option.#", acctest.Ct0),
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

func TestAccACMCertificate_root(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	var v types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic(rootDomain, types.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
					resource.TestCheckResourceAttr(resourceName, "validation_option.#", acctest.Ct0),
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

func TestAccACMCertificate_validationOptions(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	var v types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_validationOptions(rootDomain, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domain),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "validation_emails.#", 0),
					resource.TestMatchResourceAttr(resourceName, "validation_emails.0", regexache.MustCompile(`^[^@]+@.+$`)),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodEmail)),
					resource.TestCheckResourceAttr(resourceName, "validation_option.#", acctest.Ct1),
				),
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

func TestAccACMCertificate_privateCertificate_renewable(t *testing.T) {
	ctx := acctest.Context(t)
	certificateAuthorityResourceName := "aws_acmpca_certificate_authority.test"
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	var v1, v2, v3, v4 types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_authority_arn", certificateAuthorityResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_after"),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_before"),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "validation_method", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "validation_option.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMClient(ctx)

					_, err := conn.ExportCertificate(ctx, &acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					})
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.ToString(v1.CertificateArn), err)
					}
				},
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v2),
					testAccCheckCertificateNotRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityEligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct0),
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
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMClient(ctx)

					_, err := conn.RenewCertificate(ctx, &acm.RenewCertificateInput{
						CertificateArn: v1.CertificateArn,
					})
					if err != nil {
						t.Fatalf("renewing ACM Certificate (%s): %s", aws.ToString(v1.CertificateArn), err)
					}
				},
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityEligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status", string(types.RenewalStatusPendingAutoRenewal)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status_reason", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "renewal_summary.0.updated_at"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
			},
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMClient(ctx)

					_, err := tfacm.WaitCertificateRenewed(ctx, conn, aws.ToString(v1.CertificateArn), tfacm.CertificateRenewalTimeout)
					if err != nil {
						t.Fatalf("waiting for ACM Certificate (%s) renewal: %s", aws.ToString(v1.CertificateArn), err)
					}
				},
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v4),
					testAccCheckCertificateRenewed(&v3, &v4),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status", string(types.RenewalStatusSuccess)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status_reason", ""),
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

func TestAccACMCertificate_privateCertificate_noRenewalPermission(t *testing.T) {
	ctx := acctest.Context(t)
	certificateAuthorityResourceName := "aws_acmpca_certificate_authority.test"
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	var v1, v2, v3 types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateCertificate_noRenewalPermission(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_authority_arn", certificateAuthorityResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_after"),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_before"),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", certificateDomainName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "validation_method", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "validation_option.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMClient(ctx)

					_, err := conn.ExportCertificate(ctx, &acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					})
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.ToString(v1.CertificateArn), err)
					}
				},
				Config: testAccCertificateConfig_privateCertificate_noRenewalPermission(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityEligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct0),
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
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMClient(ctx)

					_, err := conn.RenewCertificate(ctx, &acm.RenewCertificateInput{
						CertificateArn: v1.CertificateArn,
					})
					if err != nil {
						t.Fatalf("renewing ACM Certificate (%s): %s", aws.ToString(v1.CertificateArn), err)
					}
				},
				Config: testAccCertificateConfig_privateCertificate_noRenewalPermission(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityEligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct1),
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
					testAccCheckCertificateExists(ctx, resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityEligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct1),
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

func TestAccACMCertificate_privateCertificate_pendingRenewalGoDuration(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	duration := (395 * 24 * time.Hour).String()
	var v1, v2 types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateCertificate_pendingRenewal(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct0),
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
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMClient(ctx)

					_, err := conn.ExportCertificate(ctx, &acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					})
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.ToString(v1.CertificateArn), err)
					}
				},
				// Ideally, we'd have a `RefreshOnly` test step here to validate that `pending_renewal` is true and `renewal_eligibility` is `ELIGIBLE` after exporting
				// before actually performing the renewal.
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/1069
				Config: testAccCertificateConfig_privateCertificate_pendingRenewal(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v2),
					testAccCheckCertificateRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status", string(types.RenewalStatusSuccess)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status_reason", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "renewal_summary.0.updated_at"),
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

func TestAccACMCertificate_privateCertificate_pendingRenewalRFC3339Duration(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	duration := "P395D"
	var v1, v2 types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateCertificate_pendingRenewal(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct0),
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
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMClient(ctx)

					_, err := conn.ExportCertificate(ctx, &acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					})
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.ToString(v1.CertificateArn), err)
					}
				},
				// Ideally, we'd have a `RefreshOnly` test step here to validate that `pending_renewal` is true and `renewal_eligibility` is `ELIGIBLE` after exporting
				// before actually performing the renewal.
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/1069
				Config: testAccCertificateConfig_privateCertificate_pendingRenewal(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v2),
					testAccCheckCertificateRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status", string(types.RenewalStatusSuccess)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status_reason", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "renewal_summary.0.updated_at"),
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

func TestAccACMCertificate_privateCertificate_addEarlyRenewalPast(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	duration := (395 * 24 * time.Hour).String()
	var v1, v2, v3 types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct0),
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
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMClient(ctx)

					_, err := conn.ExportCertificate(ctx, &acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					})
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.ToString(v1.CertificateArn), err)
					}
				},
				// Ideally, we'd have a `RefreshOnly` test step here to validate that `pending_renewal` is false and `renewal_eligibility` is `ELIGIBLE` after exporting.
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/1069
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v2),
					testAccCheckCertificateNotRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityEligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct0),
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
			{
				// Ideally, we'd have a `RefreshOnly` test step here to validate that `pending_renewal` is true after setting `early_renewal_duration`.
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/1069
				Config: testAccCertificateConfig_privateCertificate_pendingRenewal(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v3),
					testAccCheckCertificateRenewed(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status", string(types.RenewalStatusSuccess)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.0.renewal_status_reason", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "renewal_summary.0.updated_at"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
			},
		},
	})
}

func TestAccACMCertificate_privateCertificate_addEarlyRenewalPastIneligible(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	duration := (395 * 24 * time.Hour).String()
	var v1, v2 types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct0),
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
			{
				// Ideally, we'd have a `RefreshOnly` test step here to validate that `pending_renewal` is true after setting `early_renewal_duration`.
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/1069
				Config: testAccCertificateConfig_privateCertificate_pendingRenewal(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v2),
					testAccCheckCertificateNotRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
			},
		},
	})
}

func TestAccACMCertificate_privateCertificate_addEarlyRenewalFuture(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	duration := (90 * 24 * time.Hour).String()
	var v1, v2, v3 types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct0),
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
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMClient(ctx)

					_, err := conn.ExportCertificate(ctx, &acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					})
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.ToString(v1.CertificateArn), err)
					}
				},
				// Ideally, we'd have a `RefreshOnly` test step here to validate that `pending_renewal` is false and `renewal_eligibility` is `ELIGIBLE` after exporting.
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/1069
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v2),
					testAccCheckCertificateNotRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityEligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct0),
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
			{
				// Ideally, we'd have a `RefreshOnly` test step here to validate that `pending_renewal` is true after setting `early_renewal_duration`.
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/1069
				Config: testAccCertificateConfig_privateCertificate_pendingRenewal(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v3),
					testAccCheckCertificateNotRenewed(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityEligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.CertificateTypePrivate)),
				),
			},
		},
	})
}

func TestAccACMCertificate_privateCertificate_updateEarlyRenewalFuture(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	duration := (395 * 24 * time.Hour).String()
	durationUpdated := (90 * 24 * time.Hour).String()
	var v1, v2 types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateCertificate_pendingRenewal(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct0),
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
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMClient(ctx)

					_, err := conn.ExportCertificate(ctx, &acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					})
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.ToString(v1.CertificateArn), err)
					}
				},
				// Ideally, we'd have a `RefreshOnly` test step here to validate that `pending_renewal` is false and `renewal_eligibility` is `ELIGIBLE` after exporting.
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/1069
				Config: testAccCertificateConfig_privateCertificate_pendingRenewal(commonName.String(), certificateDomainName, durationUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v2),
					testAccCheckCertificateNotRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", durationUpdated),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityEligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct0),
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

func TestAccACMCertificate_privateCertificate_removeEarlyRenewal(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	commonName := acctest.RandomDomain()
	certificateDomainName := commonName.RandomSubdomain().String()
	duration := (395 * 24 * time.Hour).String()
	var v1, v2 types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateCertificate_pendingRenewal(commonName.String(), certificateDomainName, duration),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", duration),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct0),
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
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).ACMClient(ctx)

					_, err := conn.ExportCertificate(ctx, &acm.ExportCertificateInput{
						CertificateArn: v1.CertificateArn,
						Passphrase:     []byte("passphrase"),
					})
					if err != nil {
						t.Fatalf("exporting ACM Certificate (%s): %s", aws.ToString(v1.CertificateArn), err)
					}
				},
				// Ideally, we'd have a `RefreshOnly` test step here to validate that `pending_renewal` is false and `renewal_eligibility` is `ELIGIBLE` after exporting.
				// https://github.com/hashicorp/terraform-plugin-sdk/issues/1069
				Config: testAccCertificateConfig_privateCertificate_renewable(commonName.String(), certificateDomainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v2),
					testAccCheckCertificateNotRenewed(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "early_renewal_duration", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityEligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct0),
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

// TestAccACMCertificate_Root_trailingPeriod updated in 3.0 to account for domain_name plan-time validation
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13510
func TestAccACMCertificate_Root_trailingPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := fmt.Sprintf("%s.", rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccCertificateConfig_basic(domain, types.ValidationMethodDns),
				ExpectError: regexache.MustCompile(`invalid value for domain_name \(cannot end with a period\)`),
			},
		},
	})
}

func TestAccACMCertificate_rootAndWildcardSan(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)
	var v types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(rootDomain, strconv.Quote(wildcardDomain), types.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   wildcardDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", wildcardDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
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

func TestAccACMCertificate_SubjectAlternativeNames_emptyString(t *testing.T) {
	ctx := acctest.Context(t)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccCertificateConfig_subjectAlternativeNames(domain, strconv.Quote(""), types.ValidationMethodDns),
				ExpectError: regexache.MustCompile(`expected length`),
			},
		},
	})
}

func TestAccACMCertificate_San_single(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	sanDomain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	var v types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(domain, strconv.Quote(sanDomain), types.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   sanDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domain),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", sanDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
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

func TestAccACMCertificate_San_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	sanDomain1 := acctest.ACMCertificateRandomSubDomain(rootDomain)
	sanDomain2 := acctest.ACMCertificateRandomSubDomain(rootDomain)
	var v types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(domain, fmt.Sprintf("%q, %q", sanDomain1, sanDomain2), types.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", acctest.Ct3),
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
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", acctest.Ct3),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domain),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", sanDomain1),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", sanDomain2),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
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

func TestAccACMCertificate_San_trailingPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	sanDomain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	resourceName := "aws_acm_certificate.test"
	var v types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(domain, strconv.Quote(sanDomain), types.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile(`certificate/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   strings.TrimSuffix(sanDomain, "."),
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domain),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", strings.TrimSuffix(sanDomain, ".")),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
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

func TestAccACMCertificate_San_matches_domain(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	sanDomain := rootDomain
	var v types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(domain, strconv.Quote(sanDomain), types.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile(`certificate/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   sanDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", domain),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", sanDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
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

func TestAccACMCertificate_wildcard(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)
	var v types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_basic(wildcardDomain, types.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, wildcardDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   wildcardDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", wildcardDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
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

func TestAccACMCertificate_wildcardAndRootSan(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)
	var v types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_subjectAlternativeNames(wildcardDomain, strconv.Quote(rootDomain), types.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, wildcardDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   wildcardDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", wildcardDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
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

func TestAccACMCertificate_keyAlgorithm(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	var v types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_keyAlgorithm(rootDomain, types.ValidationMethodDns, types.KeyAlgorithmEcPrime256v1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
					resource.TestCheckResourceAttr(resourceName, "key_algorithm", string(types.KeyAlgorithmEcPrime256v1)),
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

func TestAccACMCertificate_disableCTLogging(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	var v types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_disableCTLogging(rootDomain, types.ValidationMethodDns),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusPendingValidation)),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.certificate_transparency_logging_preference", string(types.CertificateTransparencyLoggingPreferenceDisabled)),
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

func TestAccACMCertificate_disableReenableCTLogging(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	var v types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_optionsWithValidation(rootDomain, types.ValidationMethodDns, "ENABLED"),
				Check:  testAccCheckCertificateExists(ctx, resourceName, &v),
			},
			// Check the certificate's attributes once the validation has been applied.
			{
				Config: testAccCertificateConfig_optionsWithValidation(rootDomain, types.ValidationMethodDns, "ENABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.certificate_transparency_logging_preference", string(types.CertificateTransparencyLoggingPreferenceEnabled)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCertificateConfig_optionsWithValidation(rootDomain, types.ValidationMethodDns, "DISABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.certificate_transparency_logging_preference", string(types.CertificateTransparencyLoggingPreferenceDisabled)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCertificateConfig_optionsWithValidation(rootDomain, types.ValidationMethodDns, "ENABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm", regexache.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						names.AttrDomainName:   rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "validation_method", string(types.ValidationMethodDns)),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.certificate_transparency_logging_preference", string(types.CertificateTransparencyLoggingPreferenceEnabled)),
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
	withoutChainDomain := acctest.RandomDomainName()
	var v1, v2, v3 types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateKey(certificate, key, caCertificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, commonName),
				),
			},
			{
				Config: testAccCertificateConfig_privateKey(newCertificate, key, newCaCertificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v2),
					testAccCheckCertficateNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, commonName),
				),
			},
			{
				Config: testAccCertificateConfig_privateKeyNoChain(t, withoutChainDomain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v3),
					testAccCheckCertficateNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, withoutChainDomain),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateKey(certificate, key, caCertificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_after"),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_before"),
					resource.TestCheckResourceAttr(resourceName, "pending_renewal", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "renewal_eligibility", string(types.RenewalEligibilityIneligible)),
					resource.TestCheckResourceAttr(resourceName, "renewal_summary.#", acctest.Ct0),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig_privateKeyNoChain(t, "1.2.3.4"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.CertificateStatusIssued)),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", acctest.Ct0),
				),
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
func TestAccACMCertificate_PrivateKey_ReimportWithTags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_acm_certificate.test"
	privateKeyPEM1 := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificatePEM1 := acctest.TLSRSAX509SelfSignedCertificatePEM(t, privateKeyPEM1, "1.2.3.4")

	privateKeyPEM2 := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificatePEM2 := acctest.TLSRSAX509SelfSignedCertificatePEM(t, privateKeyPEM2, "5.6.7.8")
	var v types.CertificateDetail

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateDestroy(ctx),
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
					testAccCheckCertificateExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
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
					testAccCheckCertificateExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
					})),
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

func testAccCheckCertificateExists(ctx context.Context, n string, v *types.CertificateDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ACM Certificate ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ACMClient(ctx)

		output, err := tfacm.FindCertificateByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckCertificateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ACMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_acm_certificate" {
				continue
			}

			_, err := tfacm.FindCertificateByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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

func testAccCheckCertficateNotRecreated(v1, v2 *types.CertificateDetail) resource.TestCheckFunc {
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

func testAccCertificateConfig_privateCertificate_pendingRenewal(commonName, certificateDomainName, duration string) string {
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
