// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfapigatewayv2 "github.com/hashicorp/terraform-provider-aws/internal/service/apigatewayv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayV2DomainName_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigatewayv2.GetDomainNameOutput
	resourceName := "aws_apigatewayv2_domain_name.test"
	certResourceName := "aws_acm_certificate.test.0"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	domainName := fmt.Sprintf("%s.example.com", rName)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domainName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_basic(rName, certificate, key, 1, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", fmt.Sprintf("/domainnames/%s.example.com", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", certResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.ip_address_type", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccAPIGatewayV2DomainName_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigatewayv2.GetDomainNameOutput
	resourceName := "aws_apigatewayv2_domain_name.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	domainName := fmt.Sprintf("%s.example.com", rName)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domainName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_basic(rName, certificate, key, 1, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfapigatewayv2.ResourceDomainName(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayV2DomainName_updateCertificate(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigatewayv2.GetDomainNameOutput
	resourceName := "aws_apigatewayv2_domain_name.test"
	certResourceName0 := "aws_acm_certificate.test.0"
	certResourceName1 := "aws_acm_certificate.test.1"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	domainName := fmt.Sprintf("%s.example.com", rName)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domainName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_basic(rName, certificate, key, 2, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", fmt.Sprintf("/domainnames/%s.example.com", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", certResourceName0, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				Config: testAccDomainNameConfig_basic(rName, certificate, key, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", fmt.Sprintf("/domainnames/%s.example.com", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", certResourceName1, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				Config: testAccDomainNameConfig_tags(rName, certificate, key, 2, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", fmt.Sprintf("/domainnames/%s.example.com", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", certResourceName0, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
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

func TestAccAPIGatewayV2DomainName_MutualTLSAuthentication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	var v apigatewayv2.GetDomainNameOutput
	resourceName := "aws_apigatewayv2_domain_name.test"
	acmCertificateResourceName := "aws_acm_certificate.test"
	s3ObjectResourceName := "aws_s3_object.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_mutualTLSAuthenticationObjectVersion(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", "/domainnames/"+domain),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDomainName, acmCertificateResourceName, names.AttrDomainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", acmCertificateResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.0.truststore_uri", fmt.Sprintf("s3://%s/%s", rName, rName)),
					resource.TestCheckResourceAttrPair(resourceName, "mutual_tls_authentication.0.truststore_version", s3ObjectResourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				Config: testAccDomainNameConfig_mutualTLSAuthenticationObjectVersion(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", "/domainnames/"+domain),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDomainName, acmCertificateResourceName, names.AttrDomainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", acmCertificateResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.0.truststore_uri", fmt.Sprintf("s3://%s/%s", rName, rName)),
					resource.TestCheckResourceAttrPair(resourceName, "mutual_tls_authentication.0.truststore_version", s3ObjectResourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Test disabling mutual TLS authentication.
			{
				Config: testAccDomainNameConfig_mutualTLSAuthenticationMissing(rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", "/domainnames/"+domain),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDomainName, acmCertificateResourceName, names.AttrDomainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", acmCertificateResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				Config: testAccDomainNameConfig_mutualTLSAuthenticationObjectVersion(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", "/domainnames/"+domain),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDomainName, acmCertificateResourceName, names.AttrDomainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", acmCertificateResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.0.truststore_uri", fmt.Sprintf("s3://%s/%s", rName, rName)),
					resource.TestCheckResourceAttrPair(resourceName, "mutual_tls_authentication.0.truststore_version", s3ObjectResourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2DomainName_MutualTLSAuthentication_noVersion(t *testing.T) {
	ctx := acctest.Context(t)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	var v apigatewayv2.GetDomainNameOutput
	resourceName := "aws_apigatewayv2_domain_name.test"
	acmCertificateResourceName := "aws_acm_certificate.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_mutualTLSAuthenticationNoObjectVersion(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", "/domainnames/"+domain),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDomainName, acmCertificateResourceName, names.AttrDomainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", acmCertificateResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.0.truststore_uri", fmt.Sprintf("s3://%s/%s", rName, rName)),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.0.truststore_version", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccAPIGatewayV2DomainName_MutualTLSAuthentication_ownership(t *testing.T) {
	ctx := acctest.Context(t)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)
	var v apigatewayv2.GetDomainNameOutput
	resourceName := "aws_apigatewayv2_domain_name.test"
	publicAcmCertificateResourceName := "aws_acm_certificate.test"
	importedAcmCertificateResourceName := "aws_acm_certificate.imported"
	s3BucketObjectResourceName := "aws_s3_object.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_mutualTLSAuthenticationOwnershipVerificationCert(rName, rootDomain, domain, certificate, key),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", "/domainnames/"+domain),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDomainName, publicAcmCertificateResourceName, names.AttrDomainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", importedAcmCertificateResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.ownership_verification_certificate_arn", publicAcmCertificateResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.0.truststore_uri", fmt.Sprintf("s3://%s/%s", rName, rName)),
					resource.TestCheckResourceAttrPair(resourceName, "mutual_tls_authentication.0.truststore_version", s3BucketObjectResourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccAPIGatewayV2DomainName_ipAddressType(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigatewayv2.GetDomainNameOutput
	resourceName := "aws_apigatewayv2_domain_name.test"
	certResourceName := "aws_acm_certificate.test.0"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	domainName := fmt.Sprintf("%s.example.com", rName)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domainName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_ipAddressType(rName, "ipv4", certificate, key, 1, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", fmt.Sprintf("/domainnames/%s.example.com", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", certResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.ip_address_type", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainNameConfig_ipAddressType(rName, "dualstack", certificate, key, 1, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", fmt.Sprintf("/domainnames/%s.example.com", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", certResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.ip_address_type", "dualstack"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccAPIGatewayV2DomainName_routingMode(t *testing.T) {
	ctx := acctest.Context(t)
	var v apigatewayv2.GetDomainNameOutput
	resourceName := "aws_apigatewayv2_domain_name.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	domainName := fmt.Sprintf("%s.example.com", rName)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domainName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_routingMode(rName, certificate, key, awstypes.RoutingModeRoutingRuleThenApiMapping, 1, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("routing_mode"), tfknownvalue.StringExact(awstypes.RoutingModeRoutingRuleThenApiMapping)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainNameConfig_routingMode(rName, certificate, key, awstypes.RoutingModeApiMappingOnly, 1, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("routing_mode"), tfknownvalue.StringExact(awstypes.RoutingModeApiMappingOnly)),
				},
			},
		},
	})
}

func testAccCheckDomainNameDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).APIGatewayV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apigatewayv2_domain_name" {
				continue
			}

			_, err := tfapigatewayv2.FindDomainName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway v2 Domain Name %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDomainNameExists(ctx context.Context, t *testing.T, n string, v *apigatewayv2.GetDomainNameOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).APIGatewayV2Client(ctx)

		output, err := tfapigatewayv2.FindDomainName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDomainNameImportedCertsConfig(rName, certificate, key string, count int) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  count = %[4]d

  certificate_body = %[2]q
  private_key      = %[3]q

  tags = {
    Name = %[1]q
  }
}
`, rName, certificate, key, count)
}

func testAccDomainNamePublicCertConfig(rootDomain, domain string) string {
	return fmt.Sprintf(`
data "aws_route53_zone" "test" {
  name         = %[1]q
  private_zone = false
}

resource "aws_acm_certificate" "test" {
  domain_name       = %[2]q
  validation_method = "DNS"
}

#
# for_each acceptance testing requires:
# https://github.com/hashicorp/terraform-plugin-sdk/issues/536
#
# resource "aws_route53_record" "test" {
#   for_each = {
#     for dvo in aws_acm_certificate.test.domain_validation_options: dvo.domain_name => {
#       name   = dvo.resource_record_name
#       record = dvo.resource_record_value
#       type   = dvo.resource_record_type
#     }
#   }
#   allow_overwrite = true
#   name            = each.value.name
#   records         = [each.value.record]
#   ttl             = 60
#   type            = each.value.type
#   zone_id         = data.aws_route53_zone.test.zone_id
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
  certificate_arn         = aws_acm_certificate.test.arn
  validation_record_fqdns = [aws_route53_record.test.fqdn]
}
`, rootDomain, domain)
}

func testAccDomainNameConfig_basic(rName, certificate, key string, count, index int) string {
	return acctest.ConfigCompose(
		testAccDomainNameImportedCertsConfig(rName, certificate, key, count),
		fmt.Sprintf(`
resource "aws_apigatewayv2_domain_name" "test" {
  domain_name = "%[1]s.example.com"

  domain_name_configuration {
    certificate_arn = aws_acm_certificate.test[%[2]d].arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }
}
`, rName, index))
}

func testAccDomainNameConfig_tags(rName, certificate, key string, count, index int) string {
	return acctest.ConfigCompose(
		testAccDomainNameImportedCertsConfig(rName, certificate, key, count),
		fmt.Sprintf(`
resource "aws_apigatewayv2_domain_name" "test" {
  domain_name = "%[1]s.example.com"

  domain_name_configuration {
    certificate_arn = aws_acm_certificate.test[%[2]d].arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }

  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
  }
}
`, rName, index))
}

func testAccDomainNameConfig_mutualTLSAuthenticationNoObjectVersion(rName, rootDomain, domain string) string {
	return acctest.ConfigCompose(
		testAccDomainNamePublicCertConfig(rootDomain, domain),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = %[1]q
  source = "test-fixtures/apigateway-domain-name-truststore-1.pem"
}

resource "aws_apigatewayv2_domain_name" "test" {
  domain_name = aws_acm_certificate.test.domain_name

  domain_name_configuration {
    certificate_arn = aws_acm_certificate_validation.test.certificate_arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }

  mutual_tls_authentication {
    truststore_uri = "s3://${aws_s3_object.test.bucket}/${aws_s3_object.test.key}"
  }
}
`, rName))
}

func testAccDomainNameConfig_mutualTLSAuthenticationObjectVersion(rName, rootDomain, domain string) string {
	return acctest.ConfigCompose(
		testAccDomainNamePublicCertConfig(rootDomain, domain),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  force_destroy = true
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket_versioning.test.bucket
  key    = %[1]q
  source = "test-fixtures/apigateway-domain-name-truststore-1.pem"
}

resource "aws_apigatewayv2_domain_name" "test" {
  domain_name = aws_acm_certificate.test.domain_name

  domain_name_configuration {
    certificate_arn = aws_acm_certificate_validation.test.certificate_arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }

  mutual_tls_authentication {
    truststore_uri     = "s3://${aws_s3_object.test.bucket}/${aws_s3_object.test.key}"
    truststore_version = aws_s3_object.test.version_id
  }
}
`, rName))
}

func testAccDomainNameConfig_mutualTLSAuthenticationMissing(rootDomain, domain string) string {
	return acctest.ConfigCompose(
		testAccDomainNamePublicCertConfig(rootDomain, domain),
		`
resource "aws_apigatewayv2_domain_name" "test" {
  domain_name = aws_acm_certificate.test.domain_name

  domain_name_configuration {
    certificate_arn = aws_acm_certificate_validation.test.certificate_arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }
}
`)
}

func testAccDomainNameConfig_mutualTLSAuthenticationOwnershipVerificationCert(rName, rootDomain, domain, certificate, key string) string {
	return acctest.ConfigCompose(
		testAccDomainNamePublicCertConfig(rootDomain, domain),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = %[1]q
  source = "test-fixtures/apigateway-domain-name-truststore-1.pem"
}

resource "aws_acm_certificate" "imported" {
  certificate_body = %[2]q
  private_key      = %[3]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_apigatewayv2_domain_name" "test" {
  domain_name = aws_acm_certificate.imported.domain_name

  domain_name_configuration {
    certificate_arn                        = aws_acm_certificate.imported.arn
    endpoint_type                          = "REGIONAL"
    security_policy                        = "TLS_1_2"
    ownership_verification_certificate_arn = aws_acm_certificate_validation.test.certificate_arn
  }

  mutual_tls_authentication {
    truststore_uri = "s3://${aws_s3_object.test.bucket}/${aws_s3_object.test.key}"
  }
}
`, rName, certificate, key))
}

func testAccDomainNameConfig_ipAddressType(rName, ipAddressType, certificate, key string, count, index int) string {
	return acctest.ConfigCompose(
		testAccDomainNameImportedCertsConfig(rName, certificate, key, count),
		fmt.Sprintf(`
resource "aws_apigatewayv2_domain_name" "test" {
  domain_name = "%[1]s.example.com"

  domain_name_configuration {
    certificate_arn = aws_acm_certificate.test[%[2]d].arn
    endpoint_type   = "REGIONAL"
    ip_address_type = "%[3]s"
    security_policy = "TLS_1_2"
  }
}
`, rName, index, ipAddressType))
}

func testAccDomainNameConfig_routingMode(rName, certificate, key string, routingMode awstypes.RoutingMode, count, index int) string {
	return acctest.ConfigCompose(
		testAccDomainNameImportedCertsConfig(rName, certificate, key, count),
		fmt.Sprintf(`
resource "aws_apigatewayv2_domain_name" "test" {
  domain_name = "%[1]s.example.com"

  domain_name_configuration {
    certificate_arn = aws_acm_certificate.test[%[2]d].arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }

  routing_mode = %[3]q
}
`, rName, index, routingMode))
}
