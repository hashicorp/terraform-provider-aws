// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayDomainName_certificateARN(t *testing.T) {
	ctx := acctest.Context(t)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	var domainName apigateway.GetDomainNameOutput
	acmCertificateResourceName := "aws_acm_certificate.test"
	resourceName := "aws_api_gateway_domain_name.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, endpoints.UsEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_certificateARN(rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &domainName),
					testAccCheckResourceAttrRegionalARNEdgeDomainName(resourceName, names.AttrARN, "apigateway", domain),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, acmCertificateResourceName, names.AttrARN),
					resource.TestMatchResourceAttr(resourceName, "cloudfront_domain_name", regexache.MustCompile(`[0-9a-z]+.cloudfront.net`)),
					resource.TestCheckResourceAttr(resourceName, "cloudfront_zone_id", acctest.ProviderMeta(ctx, t).CloudFrontDistributionHostedZoneID(ctx)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDomainName, acmCertificateResourceName, names.AttrDomainName),
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

func TestAccAPIGatewayDomainName_certificateName(t *testing.T) {
	ctx := acctest.Context(t)
	certificateBody := os.Getenv("AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_BODY")
	if certificateBody == "" {
		t.Skip(
			"Environment variable AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_BODY is not set. " +
				"This environment variable must be set to any non-empty value " +
				"with a publicly trusted certificate body to enable the test.")
	}

	certificateChain := os.Getenv("AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_CHAIN")
	if certificateChain == "" {
		t.Skip(
			"Environment variable AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_CHAIN is not set. " +
				"This environment variable must be set to any non-empty value " +
				"with a chain certificate acceptable for the certificate to enable the test.")
	}

	certificatePrivateKey := os.Getenv("AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_PRIVATE_KEY")
	if certificatePrivateKey == "" {
		t.Skip(
			"Environment variable AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_PRIVATE_KEY is not set. " +
				"This environment variable must be set to any non-empty value " +
				"with a private key of a publicly trusted certificate to enable the test.")
	}

	domainName := os.Getenv("AWS_API_GATEWAY_DOMAIN_NAME_DOMAIN_NAME")
	if domainName == "" {
		t.Skip(
			"Environment variable AWS_API_GATEWAY_DOMAIN_NAME_DOMAIN_NAME is not set. " +
				"This environment variable must be set to any non-empty value " +
				"with a domain name acceptable for the certificate to enable the test.")
	}
	var conf apigateway.GetDomainNameOutput
	resourceName := "aws_api_gateway_domain_name.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_certificate(domainName, certificatePrivateKey, certificateBody, certificateChain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/domainnames/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "certificate_name", "tf-acc-apigateway-domain-name"),
					resource.TestCheckResourceAttrSet(resourceName, "cloudfront_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "cloudfront_zone_id", acctest.ProviderMeta(ctx, t).CloudFrontDistributionHostedZoneID(ctx)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domainName),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_upload_date"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"certificate_body", names.AttrCertificateChain, "certificate_private_key"},
			},
		},
	})
}

func TestAccAPIGatewayDomainName_regionalCertificateARN(t *testing.T) {
	ctx := acctest.Context(t)
	var domainName apigateway.GetDomainNameOutput
	resourceName := "aws_api_gateway_domain_name.test"
	rName := acctest.RandomSubdomain()
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_regionalCertificateARN(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &domainName),
					testAccCheckResourceAttrRegionalARNRegionalDomainName(resourceName, names.AttrARN, "apigateway", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_id", ""),
					acctest.MatchResourceAttrRegionalHostname(resourceName, "regional_domain_name", "execute-api", regexache.MustCompile(`d-[0-9a-z]+`)),
					resource.TestMatchResourceAttr(resourceName, "regional_zone_id", regexache.MustCompile(`^Z`)),
				),
			},
		},
	})
}

func TestAccAPIGatewayDomainName_regionalCertificateName(t *testing.T) {
	ctx := acctest.Context(t)
	// For now, use an environment variable to limit running this test
	// BadRequestException: Uploading certificates is not supported for REGIONAL.
	// See Remarks section of https://docs.aws.amazon.com/apigateway/api-reference/link-relation/domainname-create/
	// which suggests this configuration should be possible somewhere, e.g. AWS China?
	regionalCertificateArn := os.Getenv("AWS_API_GATEWAY_DOMAIN_NAME_REGIONAL_CERTIFICATE_NAME_ENABLED")
	if regionalCertificateArn == "" {
		t.Skip(
			"Environment variable AWS_API_GATEWAY_DOMAIN_NAME_REGIONAL_CERTIFICATE_NAME_ENABLED is not set. " +
				"This environment variable must be set to any non-empty value " +
				"in a region where uploading REGIONAL certificates is allowed " +
				"to enable the test.")
	}
	var domainName apigateway.GetDomainNameOutput
	resourceName := "aws_api_gateway_domain_name.test"
	domain := acctest.RandomDomainName()
	domainWildcard := fmt.Sprintf("*.%s", domain)
	rName := fmt.Sprintf("%s.%s", sdkacctest.RandString(8), domain)
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, caKey)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509LocallySignedCertificatePEM(t, caKey, caCertificate, key, domainWildcard)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_regionalCertificate(rName, key, certificate, caCertificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &domainName),
					testAccCheckResourceAttrRegionalARNRegionalDomainName(resourceName, names.AttrARN, "apigateway", rName),
					resource.TestCheckResourceAttr(resourceName, "certificate_body", certificate),
					resource.TestCheckResourceAttr(resourceName, names.AttrCertificateChain, caCertificate),
					resource.TestCheckResourceAttr(resourceName, "certificate_name", "tf-acc-apigateway-domain-name"),
					resource.TestCheckResourceAttr(resourceName, "certificate_private_key", key),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_upload_date"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rName),
					acctest.MatchResourceAttrRegionalHostname(resourceName, "regional_domain_name", "execute-api", regexache.MustCompile(`d-[0-9a-z]+`)),
					resource.TestMatchResourceAttr(resourceName, "regional_zone_id", regexache.MustCompile(`^Z`)),
				),
			},
		},
	})
}

func TestAccAPIGatewayDomainName_securityPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var domainName apigateway.GetDomainNameOutput
	resourceName := "aws_api_gateway_domain_name.test"
	rName := acctest.RandomSubdomain()
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_securityPolicy(rName, key, certificate, string(types.SecurityPolicyTls12)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, "security_policy", string(types.SecurityPolicyTls12)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_access_mode", ""),
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

func TestAccAPIGatewayDomainName_securityPolicyEnhanced(t *testing.T) {
	ctx := acctest.Context(t)
	var domainName apigateway.GetDomainNameOutput
	resourceName := "aws_api_gateway_domain_name.test"
	rName := acctest.RandomSubdomain()
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_securityPolicyWithEndpointAccessMode(rName, key, certificate, string(types.SecurityPolicySecurityPolicyTls1312PfsPq202509), string(types.EndpointAccessModeBasic)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, "security_policy", string(types.SecurityPolicySecurityPolicyTls1312PfsPq202509)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_access_mode", string(types.EndpointAccessModeBasic)),
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

func TestAccAPIGatewayDomainName_securityPolicyEnhancedUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var domainName apigateway.GetDomainNameOutput
	resourceName := "aws_api_gateway_domain_name.test"
	rName := acctest.RandomSubdomain()
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_securityPolicy(rName, key, certificate, string(types.SecurityPolicyTls12)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, "security_policy", string(types.SecurityPolicyTls12)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_access_mode", ""),
				),
			},
			{
				Config: testAccDomainNameConfig_securityPolicyWithEndpointAccessMode(rName, key, certificate, string(types.SecurityPolicySecurityPolicyTls1312PfsPq202509), string(types.EndpointAccessModeBasic)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, "security_policy", string(types.SecurityPolicySecurityPolicyTls1312PfsPq202509)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_access_mode", string(types.EndpointAccessModeBasic)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainNameConfig_securityPolicyWithEndpointAccessMode(rName, key, certificate, string(types.SecurityPolicySecurityPolicyTls1313202509), string(types.EndpointAccessModeStrict)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, "security_policy", string(types.SecurityPolicySecurityPolicyTls1313202509)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_access_mode", string(types.EndpointAccessModeStrict)),
				),
			},
			{
				Config: testAccDomainNameConfig_securityPolicy(rName, key, certificate, string(types.SecurityPolicyTls12)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, "security_policy", string(types.SecurityPolicyTls12)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_access_mode", ""),
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

func TestAccAPIGatewayDomainName_private(t *testing.T) {
	ctx := acctest.Context(t)
	var domainName apigateway.GetDomainNameOutput
	resourceName := "aws_api_gateway_domain_name.test"
	rName := acctest.RandomSubdomain()
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_private(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &domainName),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_id"),
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

func TestAccAPIGatewayDomainName_privatePolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var domainName apigateway.GetDomainNameOutput
	resourceName := "aws_api_gateway_domain_name.test"
	rName := acctest.RandomSubdomain()
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_privatePolicy(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &domainName),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
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

func TestAccAPIGatewayDomainName_privatePolicy_added(t *testing.T) {
	ctx := acctest.Context(t)
	var domainName apigateway.GetDomainNameOutput
	resourceName := "aws_api_gateway_domain_name.test"
	rName := acctest.RandomSubdomain()
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_private(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &domainName),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_id"),
				),
			},
			{
				Config: testAccDomainNameConfig_privatePolicy(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &domainName),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
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

func TestAccAPIGatewayDomainName_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var domainName apigateway.GetDomainNameOutput
	resourceName := "aws_api_gateway_domain_name.test"
	rName := acctest.RandomSubdomain()
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_regionalCertificateARN(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &domainName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfapigateway.ResourceDomainName(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAPIGatewayDomainName_MutualTLSAuthentication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := fmt.Sprintf("%s.%s", acctest.RandomSubdomain(), rootDomain)
	var v apigateway.GetDomainNameOutput
	resourceName := "aws_api_gateway_domain_name.test"
	acmCertificateResourceName := "aws_acm_certificate.test"
	s3ObjectResourceName := "aws_s3_object.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_mutualTLSAuthentication(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/domainnames/.+$`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDomainName, acmCertificateResourceName, names.AttrDomainName),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.0.truststore_uri", fmt.Sprintf("s3://%s/%s", rName, rName)),
					resource.TestCheckResourceAttrPair(resourceName, "mutual_tls_authentication.0.truststore_version", s3ObjectResourceName, "version_id"),
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
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDomainName, acmCertificateResourceName, names.AttrDomainName),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "0"),
				),
			},
			{
				Config: testAccDomainNameConfig_mutualTLSAuthentication(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/domainnames/.+$`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDomainName, acmCertificateResourceName, names.AttrDomainName),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.0.truststore_uri", fmt.Sprintf("s3://%s/%s", rName, rName)),
					resource.TestCheckResourceAttrPair(resourceName, "mutual_tls_authentication.0.truststore_version", s3ObjectResourceName, "version_id"),
				),
			},
		},
	})
}

func TestAccAPIGatewayDomainName_MutualTLSAuthentication_ownership(t *testing.T) {
	ctx := acctest.Context(t)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := fmt.Sprintf("%s.%s", acctest.RandomSubdomain(), rootDomain)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, domain)
	var v apigateway.GetDomainNameOutput
	resourceName := "aws_api_gateway_domain_name.test"
	publicAcmCertificateResourceName := "aws_acm_certificate.test"
	s3ObjectResourceName := "aws_s3_object.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_mutualTLSOwnership(rName, rootDomain, domain, certificate, key),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, names.AttrARN, "apigateway", regexache.MustCompile(`/domainnames/.+$`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDomainName, publicAcmCertificateResourceName, names.AttrDomainName),
					resource.TestCheckResourceAttrPair(resourceName, "ownership_verification_certificate_arn", publicAcmCertificateResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.0.truststore_uri", fmt.Sprintf("s3://%s/%s", rName, rName)),
					resource.TestCheckResourceAttrPair(resourceName, "mutual_tls_authentication.0.truststore_version", s3ObjectResourceName, "version_id"),
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

func TestAccAPIGatewayDomainName_updateIDFormat(t *testing.T) {
	ctx := acctest.Context(t)
	var domainName apigateway.GetDomainNameOutput
	resourceName := "aws_api_gateway_domain_name.test"
	rName := acctest.RandomSubdomain()
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.APIGatewayServiceID),
		CheckDestroy: testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.80.0",
					},
				},
				Config: testAccDomainNameConfig_regionalCertificateARN(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &domainName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoValue(resourceName, tfjsonpath.New("domain_name_id")),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccDomainNameConfig_regionalCertificateARN(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &domainName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("domain_name_id"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func TestAccAPIGatewayDomainName_ipAddressType(t *testing.T) {
	ctx := acctest.Context(t)
	var domainName apigateway.GetDomainNameOutput
	resourceName := "aws_api_gateway_domain_name.test"
	rName := acctest.RandomSubdomain()
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_ipAddressType(rName, key, certificate, "ipv4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &domainName),
					testAccCheckResourceAttrRegionalARNRegionalDomainName(resourceName, names.AttrARN, "apigateway", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_id", ""),
					acctest.MatchResourceAttrRegionalHostname(resourceName, "regional_domain_name", "execute-api", regexache.MustCompile(`d-[0-9a-z]+`)),
					resource.TestMatchResourceAttr(resourceName, "regional_zone_id", regexache.MustCompile(`^Z`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.ip_address_type", "ipv4"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainNameConfig_ipAddressType(rName, key, certificate, "dualstack"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &domainName),
					testAccCheckResourceAttrRegionalARNRegionalDomainName(resourceName, names.AttrARN, "apigateway", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, rName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_id", ""),
					acctest.MatchResourceAttrRegionalHostname(resourceName, "regional_domain_name", "execute-api", regexache.MustCompile(`d-[0-9a-z]+`)),
					resource.TestMatchResourceAttr(resourceName, "regional_zone_id", regexache.MustCompile(`^Z`)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_configuration.0.ip_address_type", "dualstack"),
				),
			},
		},
	})
}

func TestAccAPIGatewayDomainName_routingMode(t *testing.T) {
	ctx := acctest.Context(t)
	var domainName apigateway.GetDomainNameOutput
	resourceName := "aws_api_gateway_domain_name.test"
	rName := acctest.RandomSubdomain()
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameConfig_routingMode(rName, key, certificate, types.RoutingModeRoutingRuleOnly),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &domainName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("routing_mode"), tfknownvalue.StringExact(types.RoutingModeRoutingRuleOnly)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainNameConfig_routingMode(rName, key, certificate, types.RoutingModeRoutingRuleThenBasePathMapping),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameExists(ctx, t, resourceName, &domainName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("routing_mode"), tfknownvalue.StringExact(types.RoutingModeRoutingRuleThenBasePathMapping)),
				},
			},
		},
	})
}

func testAccCheckDomainNameExists(ctx context.Context, t *testing.T, n string, v *apigateway.GetDomainNameOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).APIGatewayClient(ctx)

		output, err := tfapigateway.FindDomainNameByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrDomainName], rs.Primary.Attributes["domain_name_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckDomainNameDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_domain_name" {
				continue
			}

			_, err := tfapigateway.FindDomainNameByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrDomainName], rs.Primary.Attributes["domain_name_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("API Gateway Domain Name %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

// testAccCheckResourceAttrRegionalARNEdgeDomainName ensures the Terraform state exactly matches the expected API Gateway Edge Domain Name format.
func testAccCheckResourceAttrRegionalARNEdgeDomainName(resourceName, attributeName, arnService string, domain string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			Partition: acctest.Partition(),
			Region:    acctest.Region(),
			Resource:  fmt.Sprintf("/domainnames/%s", domain),
			Service:   arnService,
		}.String()

		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// testAccCheckResourceAttrRegionalARNRegionalDomainName ensures the Terraform state exactly matches the expected API Gateway Regional Domain Name format.
func testAccCheckResourceAttrRegionalARNRegionalDomainName(resourceName, attributeName, arnService string, domain string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			Partition: acctest.Partition(),
			Region:    acctest.Region(),
			Resource:  fmt.Sprintf("/domainnames/%s", domain),
			Service:   arnService,
		}.String()

		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

func testAccDomainNameConfig_basePublicCert(rootDomain, domain string) string {
	return fmt.Sprintf(`
data "aws_route53_zone" "test" {
  name         = %[1]q
  private_zone = false
}

resource "aws_acm_certificate" "test" {
  domain_name       = %[2]q
  validation_method = "DNS"
}

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

func testAccDomainNameConfig_certificateARN(rootDomain string, domain string) string {
	return acctest.ConfigCompose(testAccDomainNameConfig_basePublicCert(rootDomain, domain), `
resource "aws_api_gateway_domain_name" "test" {
  domain_name     = aws_acm_certificate.test.domain_name
  certificate_arn = aws_acm_certificate_validation.test.certificate_arn

  endpoint_configuration {
    types = ["EDGE"]
  }
}
`)
}

func testAccDomainNameConfig_certificate(domainName, key, certificate, chainCertificate string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_domain_name" "test" {
  domain_name             = %[1]q
  certificate_body        = "%[2]s"
  certificate_chain       = "%[3]s"
  certificate_name        = "tf-acc-apigateway-domain-name"
  certificate_private_key = "%[4]s"
}
`, domainName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(chainCertificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccDomainNameConfig_regionalCertificateARN(domainName, key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_api_gateway_domain_name" "test" {
  domain_name              = %[1]q
  regional_certificate_arn = aws_acm_certificate.test.arn

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}
`, domainName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccDomainNameConfig_regionalCertificate(domainName, key, certificate, chainCertificate string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_domain_name" "test" {
  certificate_body          = "%[2]s"
  certificate_chain         = "%[3]s"
  certificate_private_key   = "%[4]s"
  domain_name               = %[1]q
  regional_certificate_name = "tf-acc-apigateway-domain-name"

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}
`, domainName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(chainCertificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccDomainNameConfig_securityPolicy(domainName, key, certificate, securityPolicy string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_api_gateway_domain_name" "test" {
  domain_name              = %[1]q
  regional_certificate_arn = aws_acm_certificate.test.arn
  security_policy          = %[4]q

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}
`, domainName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key), securityPolicy)
}

func testAccDomainNameConfig_securityPolicyWithEndpointAccessMode(domainName, key, certificate, securityPolicy, endpointAccessMode string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_api_gateway_domain_name" "test" {
  domain_name              = %[1]q
  regional_certificate_arn = aws_acm_certificate.test.arn
  security_policy          = %[4]q
  endpoint_access_mode     = %[5]q

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}
`, domainName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key), securityPolicy, endpointAccessMode)
}

func testAccDomainNameConfig_private(domainName, key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_api_gateway_domain_name" "test" {
  domain_name     = %[1]q
  certificate_arn = aws_acm_certificate.test.arn

  endpoint_configuration {
    types = ["PRIVATE"]
  }
}
`, domainName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccDomainNameConfig_privatePolicy(domainName, key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_api_gateway_domain_name" "test" {
  domain_name     = %[1]q
  certificate_arn = aws_acm_certificate.test.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Action   = "execute-api:Invoke"
      Resource = "*"
    }]
  })

  endpoint_configuration {
    types = ["PRIVATE"]
  }
}
`, domainName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccDomainNameConfig_mutualTLSAuthentication(rName, rootDomain, domain string) string {
	return acctest.ConfigCompose(
		testAccDomainNameConfig_basePublicCert(rootDomain, domain),
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

resource "aws_api_gateway_domain_name" "test" {
  domain_name              = aws_acm_certificate.test.domain_name
  regional_certificate_arn = aws_acm_certificate_validation.test.certificate_arn
  security_policy          = "TLS_1_2"

  endpoint_configuration {
    types = ["REGIONAL"]
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
		testAccDomainNameConfig_basePublicCert(rootDomain, domain),
		`
resource "aws_api_gateway_domain_name" "test" {
  domain_name              = aws_acm_certificate.test.domain_name
  regional_certificate_arn = aws_acm_certificate_validation.test.certificate_arn
  security_policy          = "TLS_1_2"

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}
`)
}

func testAccDomainNameConfig_mutualTLSOwnership(rName, rootDomain, domain, certificate, key string) string {
	return acctest.ConfigCompose(
		testAccDomainNameConfig_basePublicCert(rootDomain, domain),
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

resource "aws_acm_certificate" "imported" {
  certificate_body = %[2]q
  private_key      = %[3]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_api_gateway_domain_name" "test" {
  domain_name                            = aws_acm_certificate.test.domain_name
  regional_certificate_arn               = aws_acm_certificate.imported.arn
  security_policy                        = "TLS_1_2"
  ownership_verification_certificate_arn = aws_acm_certificate_validation.test.certificate_arn

  endpoint_configuration {
    types = ["REGIONAL"]
  }

  mutual_tls_authentication {
    truststore_uri     = "s3://${aws_s3_object.test.bucket}/${aws_s3_object.test.key}"
    truststore_version = aws_s3_object.test.version_id
  }
}
`, rName, certificate, key))
}

func testAccDomainNameConfig_ipAddressType(domainName, key, certificate, ipAddressType string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_api_gateway_domain_name" "test" {
  domain_name              = %[1]q
  regional_certificate_arn = aws_acm_certificate.test.arn

  endpoint_configuration {
    types           = ["REGIONAL"]
    ip_address_type = %[4]q
  }
}
`, domainName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key), ipAddressType)
}

func testAccDomainNameConfig_routingMode(domainName, key, certificate string, routingMode types.RoutingMode) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_api_gateway_domain_name" "test" {
  domain_name     = %[1]q
  certificate_arn = aws_acm_certificate.test.arn

  endpoint_configuration {
    types = ["PRIVATE"]
  }

  routing_mode = %[4]q
}
`, domainName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key), routingMode)
}
