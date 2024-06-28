// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acmpca_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/acmpca/types"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/acmpca"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfacmpca "github.com/hashicorp/terraform-provider-aws/internal/service/acmpca"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var testAccCheckCertificateAuthorityExists = acctest.CheckACMPCACertificateAuthorityExists

func TestAccACMPCACertificateAuthority_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var certificateAuthority types.CertificateAuthority
	resourceName := "aws_acmpca_certificate_authority.test"
	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateAuthorityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateAuthorityConfig_required(commonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm-pca", regexache.MustCompile(`certificate-authority/.+`)),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority_configuration.0.key_algorithm", "RSA_4096"),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority_configuration.0.signing_algorithm", "SHA512WITHRSA"),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority_configuration.0.subject.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority_configuration.0.subject.0.common_name", commonName),
					resource.TestCheckResourceAttr(resourceName, names.AttrCertificate, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrCertificateChain, ""),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_signing_request"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "key_storage_security_standard", "FIPS_140_2_LEVEL_3_OR_HIGHER"),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_after"),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_before"),
					resource.TestCheckResourceAttr(resourceName, "permanent_deletion_time_in_days", "30"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "serial", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "SUBORDINATE"),
					resource.TestCheckResourceAttr(resourceName, "usage_mode", "SHORT_LIVED_CERTIFICATE"),
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

func TestAccACMPCACertificateAuthority_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var certificateAuthority types.CertificateAuthority
	resourceName := "aws_acmpca_certificate_authority.test"
	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateAuthorityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateAuthorityConfig_required(commonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfacmpca.ResourceCertificateAuthority(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccACMPCACertificateAuthority_enabledDeprecated(t *testing.T) {
	ctx := acctest.Context(t)
	var certificateAuthority types.CertificateAuthority
	resourceName := "aws_acmpca_certificate_authority.test"
	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateAuthorityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateAuthorityConfig_enabled(commonName, acmpca.CertificateAuthorityTypeRoot, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, acmpca.CertificateAuthorityTypeRoot),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					acctest.CheckACMPCACertificateAuthorityActivateRootCA(ctx, &certificateAuthority),
				),
			},
			{
				Config: testAccCertificateAuthorityConfig_enabled(commonName, acmpca.CertificateAuthorityTypeRoot, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, acmpca.CertificateAuthorityTypeRoot),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
			{
				Config: testAccCertificateAuthorityConfig_enabled(commonName, acmpca.CertificateAuthorityTypeRoot, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"permanent_deletion_time_in_days",
				},
			},
		},
	})
}

func TestAccACMPCACertificateAuthority_usageMode(t *testing.T) {
	ctx := acctest.Context(t)
	var certificateAuthority types.CertificateAuthority
	resourceName := "aws_acmpca_certificate_authority.test"
	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateAuthorityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateAuthorityConfig_usageMode(commonName, acmpca.CertificateAuthorityTypeRoot, "SHORT_LIVED_CERTIFICATE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "usage_mode", "SHORT_LIVED_CERTIFICATE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"permanent_deletion_time_in_days",
				},
			},
		},
	})
}

func TestAccACMPCACertificateAuthority_keyStorageSecurityStandard(t *testing.T) {
	ctx := acctest.Context(t)
	var certificateAuthority types.CertificateAuthority
	resourceName := "aws_acmpca_certificate_authority.test"
	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// See https://docs.aws.amazon.com/privateca/latest/userguide/data-protection.html#private-keys.
			acctest.PreCheckRegion(t, endpoints.ApSouth2RegionID, endpoints.ApSoutheast3RegionID, endpoints.ApSoutheast4RegionID, endpoints.EuCentral2RegionID, endpoints.EuSouth2RegionID, endpoints.MeCentral1RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateAuthorityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateAuthorityConfig_keyStorageSecurityStandard(commonName, acmpca.CertificateAuthorityTypeRoot, "FIPS_140_2_LEVEL_2_OR_HIGHER"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "key_storage_security_standard", "FIPS_140_2_LEVEL_2_OR_HIGHER"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"permanent_deletion_time_in_days",
				},
			},
		},
	})
}

func TestAccACMPCACertificateAuthority_deleteFromActiveState(t *testing.T) {
	ctx := acctest.Context(t)
	var certificateAuthority types.CertificateAuthority
	resourceName := "aws_acmpca_certificate_authority.test"
	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateAuthorityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateAuthorityConfig_root(commonName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, acmpca.CertificateAuthorityTypeRoot),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccACMPCACertificateAuthority_RevocationConfiguration_empty(t *testing.T) {
	ctx := acctest.Context(t)
	var certificateAuthority types.CertificateAuthority
	resourceName := "aws_acmpca_certificate_authority.test"
	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateAuthorityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateAuthorityConfig_revocationConfigurationEmpty(commonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "acm-pca", regexache.MustCompile(`certificate-authority/.+`)),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority_configuration.0.key_algorithm", "RSA_4096"),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority_configuration.0.signing_algorithm", "SHA512WITHRSA"),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority_configuration.0.subject.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority_configuration.0.subject.0.common_name", commonName),
					resource.TestCheckResourceAttr(resourceName, names.AttrCertificate, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrCertificateChain, ""),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_signing_request"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "key_storage_security_standard", "FIPS_140_2_LEVEL_3_OR_HIGHER"),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_after"),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_before"),
					resource.TestCheckResourceAttr(resourceName, "permanent_deletion_time_in_days", "30"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "serial", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "SUBORDINATE"),
					resource.TestCheckResourceAttr(resourceName, "usage_mode", "SHORT_LIVED_CERTIFICATE"),
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

func TestAccACMPCACertificateAuthority_RevocationCrl_customCNAME(t *testing.T) {
	ctx := acctest.Context(t)
	var certificateAuthority types.CertificateAuthority
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_acmpca_certificate_authority.test"
	domain := acctest.RandomDomain()
	commonName := domain.String()
	customCName := domain.Subdomain("crl").String()
	customCName2 := domain.Subdomain("crl2").String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateAuthorityDestroy(ctx),
		Steps: []resource.TestStep{
			// Test creating revocation configuration on resource creation
			{
				Config: testAccCertificateAuthorityConfig_revocationConfigurationCrlConfigurationCustomCNAME(rName, commonName, customCName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", customCName),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test importing revocation configuration
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"permanent_deletion_time_in_days",
				},
			},
			// Test updating revocation configuration
			{
				Config: testAccCertificateAuthorityConfig_revocationConfigurationCrlConfigurationCustomCNAME(rName, commonName, customCName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", customCName2),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test removing custom cname on resource update
			{
				Config: testAccCertificateAuthorityConfig_revocationConfigurationCrlConfigurationEnabled(rName, commonName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", ""),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test adding custom cname on resource update
			{
				Config: testAccCertificateAuthorityConfig_revocationConfigurationCrlConfigurationCustomCNAME(rName, commonName, customCName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", customCName),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test removing revocation configuration on resource update
			{
				Config: testAccCertificateAuthorityConfig_required(commonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccACMPCACertificateAuthority_RevocationCrl_enabled(t *testing.T) {
	ctx := acctest.Context(t)
	var certificateAuthority types.CertificateAuthority
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_acmpca_certificate_authority.test"
	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateAuthorityDestroy(ctx),
		Steps: []resource.TestStep{
			// Test creating revocation configuration on resource creation
			{
				Config: testAccCertificateAuthorityConfig_revocationConfigurationCrlConfigurationEnabled(rName, commonName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", ""),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test importing revocation configuration
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"permanent_deletion_time_in_days",
				},
			},
			// Test disabling revocation configuration
			{
				Config: testAccCertificateAuthorityConfig_revocationConfigurationCrlConfigurationEnabled(rName, commonName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", acctest.CtFalse),
				),
			},
			// Test enabling revocation configuration
			{
				Config: testAccCertificateAuthorityConfig_revocationConfigurationCrlConfigurationEnabled(rName, commonName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", ""),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test removing revocation configuration on resource update
			{
				Config: testAccCertificateAuthorityConfig_required(commonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccACMPCACertificateAuthority_RevocationCrl_expirationInDays(t *testing.T) {
	ctx := acctest.Context(t)
	var certificateAuthority types.CertificateAuthority
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_acmpca_certificate_authority.test"
	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateAuthorityDestroy(ctx),
		Steps: []resource.TestStep{
			// Test creating revocation configuration on resource creation
			{
				Config: testAccCertificateAuthorityConfig_revocationConfigurationCrlConfigurationExpirationInDays(rName, commonName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", ""),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_object_acl", "PUBLIC_READ"),
				),
			},
			// Test importing revocation configuration
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"permanent_deletion_time_in_days",
				},
			},
			// Test updating revocation configuration
			{
				Config: testAccCertificateAuthorityConfig_revocationConfigurationCrlConfigurationExpirationInDays(rName, commonName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test removing revocation configuration on resource update
			{
				Config: testAccCertificateAuthorityConfig_required(commonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccACMPCACertificateAuthority_RevocationCrl_s3ObjectACL(t *testing.T) {
	ctx := acctest.Context(t)
	var certificateAuthority types.CertificateAuthority
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_acmpca_certificate_authority.test"
	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateAuthorityDestroy(ctx),
		Steps: []resource.TestStep{
			// Test creating revocation configuration on resource creation
			{
				Config: testAccCertificateAuthorityConfig_revocationConfigurationCrlConfigurationS3ObjectACL(rName, commonName, "BUCKET_OWNER_FULL_CONTROL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_object_acl", "BUCKET_OWNER_FULL_CONTROL"),
				),
			},
			// Test importing revocation configuration
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"permanent_deletion_time_in_days",
				},
			},
			// Test updating revocation configuration
			{
				Config: testAccCertificateAuthorityConfig_revocationConfigurationCrlConfigurationS3ObjectACL(rName, commonName, "PUBLIC_READ"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_object_acl", "PUBLIC_READ"),
				),
			},
		},
	})
}

func TestAccACMPCACertificateAuthority_RevocationOcsp_enabled(t *testing.T) {
	ctx := acctest.Context(t)
	var certificateAuthority types.CertificateAuthority
	resourceName := "aws_acmpca_certificate_authority.test"
	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateAuthorityDestroy(ctx),
		Steps: []resource.TestStep{
			// Test creating OCSP revocation configuration on resource creation
			{
				Config: testAccCertificateAuthorityConfig_revocationConfigurationOcspConfigurationEnabled(commonName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.0.ocsp_custom_cname", ""),
				),
			},
			// Test importing OCSP revocation configuration
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"permanent_deletion_time_in_days",
				},
			},
			// Test disabling OCSP revocation configuration
			{
				Config: testAccCertificateAuthorityConfig_revocationConfigurationOcspConfigurationEnabled(commonName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.0.enabled", acctest.CtFalse),
				),
			},
			// Test enabling OCSP revocation configuration
			{
				Config: testAccCertificateAuthorityConfig_revocationConfigurationOcspConfigurationEnabled(commonName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.0.ocsp_custom_cname", ""),
				),
			},
			// Test removing revocation configuration on resource update
			{
				Config: testAccCertificateAuthorityConfig_required(commonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.0.enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccACMPCACertificateAuthority_RevocationOcsp_customCNAME(t *testing.T) {
	ctx := acctest.Context(t)
	var certificateAuthority types.CertificateAuthority
	resourceName := "aws_acmpca_certificate_authority.test"
	domain := acctest.RandomDomain()
	commonName := domain.String()
	customCName := domain.Subdomain("ocspl").String()
	customCName2 := domain.Subdomain("ocsp2").String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ACMPCAServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCertificateAuthorityDestroy(ctx),
		Steps: []resource.TestStep{
			// Test creating revocation configuration on resource creation
			{
				Config: testAccCertificateAuthorityConfig_revocationConfigurationOcspConfigurationCustomCNAME(commonName, customCName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.0.ocsp_custom_cname", customCName),
				),
			},
			// Test importing OCSP revocation configuration
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"permanent_deletion_time_in_days",
				},
			},
			// Test updating OCSP revocation configuration
			{
				Config: testAccCertificateAuthorityConfig_revocationConfigurationOcspConfigurationCustomCNAME(commonName, customCName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.0.ocsp_custom_cname", customCName2),
				),
			},
			// Test removing OCSP custom cname on resource update
			{
				Config: testAccCertificateAuthorityConfig_revocationConfigurationOcspConfigurationEnabled(commonName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.0.ocsp_custom_cname", ""),
				),
			},
			// Test adding OCSP custom cname on resource update
			{
				Config: testAccCertificateAuthorityConfig_revocationConfigurationOcspConfigurationCustomCNAME(commonName, customCName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.0.ocsp_custom_cname", customCName),
				),
			},
			// Test removing revocation configuration on resource update
			{
				Config: testAccCertificateAuthorityConfig_required(commonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateAuthorityExists(ctx, resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.ocsp_configuration.0.enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCheckCertificateAuthorityDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ACMPCAClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_acmpca_certificate_authority" {
				continue
			}

			_, err := tfacmpca.FindCertificateAuthorityByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ACM PCA Certificate Authority %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCertificateAuthorityConfig_enabled(commonName, certificateAuthorityType string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  enabled                         = %[1]t
  usage_mode                      = "SHORT_LIVED_CERTIFICATE"
  permanent_deletion_time_in_days = 7
  type                            = %[2]q

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[3]q
    }
  }
}
`, enabled, certificateAuthorityType, commonName)
}

func testAccCertificateAuthorityConfig_usageMode(commonName, certificateAuthorityType string, usageMode string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  enabled                         = true
  usage_mode                      = %[1]q
  permanent_deletion_time_in_days = 7
  type                            = %[2]q

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[3]q
    }
  }
}
`, usageMode, certificateAuthorityType, commonName)
}

func testAccCertificateAuthorityConfig_root(commonName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  usage_mode                      = "SHORT_LIVED_CERTIFICATE"
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

resource "aws_acmpca_certificate_authority_certificate" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn

  certificate       = aws_acmpca_certificate.test.certificate
  certificate_chain = aws_acmpca_certificate.test.certificate_chain
}

resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 1
  }
}

data "aws_partition" "current" {}
`, commonName)
}

func testAccCertificateAuthorityConfig_required(commonName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  usage_mode = "SHORT_LIVED_CERTIFICATE"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}
`, commonName)
}

func testAccCertificateAuthorityConfig_revocationConfigurationEmpty(commonName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  usage_mode = "SHORT_LIVED_CERTIFICATE"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }

  revocation_configuration {
  }
}
`, commonName)
}

func testAccCertificateAuthorityConfig_revocationConfigurationCrlConfigurationCustomCNAME(rName, commonName, customCname string) string {
	return acctest.ConfigCompose(
		testAccCertificateAuthorityConfig_S3Bucket(rName),
		fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  usage_mode                      = "SHORT_LIVED_CERTIFICATE"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }

  revocation_configuration {
    crl_configuration {
      custom_cname       = %[2]q
      enabled            = true
      expiration_in_days = 1
      s3_bucket_name     = aws_s3_bucket.test.id
    }
  }

  depends_on = [
    aws_s3_bucket_policy.test,
    aws_s3_bucket_public_access_block.test,
    aws_s3_bucket_ownership_controls.test,
  ]
}
`, commonName, customCname))
}

func testAccCertificateAuthorityConfig_revocationConfigurationCrlConfigurationEnabled(rName, commonName string, enabled bool) string {
	return acctest.ConfigCompose(
		testAccCertificateAuthorityConfig_S3Bucket(rName),
		fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  usage_mode                      = "SHORT_LIVED_CERTIFICATE"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }

  revocation_configuration {
    crl_configuration {
      enabled            = %[2]t
      expiration_in_days = 1
      s3_bucket_name     = aws_s3_bucket.test.id
    }
  }

  depends_on = [
    aws_s3_bucket_policy.test,
    aws_s3_bucket_public_access_block.test,
    aws_s3_bucket_ownership_controls.test,
  ]
}
`, commonName, enabled))
}

func testAccCertificateAuthorityConfig_revocationConfigurationCrlConfigurationExpirationInDays(rName, commonName string, expirationInDays int) string {
	return acctest.ConfigCompose(
		testAccCertificateAuthorityConfig_S3Bucket(rName),
		fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  usage_mode                      = "SHORT_LIVED_CERTIFICATE"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }

  revocation_configuration {
    crl_configuration {
      enabled            = true
      expiration_in_days = %[2]d
      s3_bucket_name     = aws_s3_bucket.test.id
    }
  }

  depends_on = [
    aws_s3_bucket_policy.test,
    aws_s3_bucket_public_access_block.test,
    aws_s3_bucket_ownership_controls.test,
  ]
}
`, commonName, expirationInDays))
}

func testAccCertificateAuthorityConfig_revocationConfigurationCrlConfigurationS3ObjectACL(rName, commonName, s3ObjectAcl string) string {
	return acctest.ConfigCompose(
		testAccCertificateAuthorityConfig_S3Bucket(rName),
		fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  usage_mode                      = "SHORT_LIVED_CERTIFICATE"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }

  revocation_configuration {
    crl_configuration {
      enabled            = true
      expiration_in_days = 1
      s3_bucket_name     = aws_s3_bucket.test.id
      s3_object_acl      = %[2]q
    }
  }

  depends_on = [
    aws_s3_bucket_policy.test,
    aws_s3_bucket_public_access_block.test,
    aws_s3_bucket_ownership_controls.test,
  ]
}
`, commonName, s3ObjectAcl))
}

func testAccCertificateAuthorityConfig_revocationConfigurationOcspConfigurationCustomCNAME(commonName, customCname string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  usage_mode                      = "SHORT_LIVED_CERTIFICATE"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }

  revocation_configuration {
    ocsp_configuration {
      enabled           = true
      ocsp_custom_cname = %[2]q
    }
  }
}
`, commonName, customCname)
}

func testAccCertificateAuthorityConfig_revocationConfigurationOcspConfigurationEnabled(commonName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  usage_mode                      = "SHORT_LIVED_CERTIFICATE"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }

  revocation_configuration {
    ocsp_configuration {
      enabled = %[2]t
    }
  }
}
`, commonName, enabled)
}

func testAccCertificateAuthorityConfig_S3Bucket(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_ownership_controls" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

data "aws_iam_policy_document" "acmpca_bucket_access" {
  statement {
    actions = [
      "s3:GetBucketAcl",
      "s3:GetBucketLocation",
      "s3:PutObject",
      "s3:PutObjectAcl",
    ]

    resources = [
      aws_s3_bucket.test.arn,
      "${aws_s3_bucket.test.arn}/*",
    ]

    principals {
      identifiers = ["acm-pca.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  policy = data.aws_iam_policy_document.acmpca_bucket_access.json
}
`, rName)
}

func testAccCertificateAuthorityConfig_keyStorageSecurityStandard(commonName, certificateAuthorityType, keyStorageSecurityStandard string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  enabled                         = true
  usage_mode                      = "SHORT_LIVED_CERTIFICATE"
  key_storage_security_standard   = %[1]q
  permanent_deletion_time_in_days = 7
  type                            = %[2]q

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[3]q
    }
  }
}
`, keyStorageSecurityStandard, certificateAuthorityType, commonName)
}
