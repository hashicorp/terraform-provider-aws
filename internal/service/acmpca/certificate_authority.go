// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acmpca

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acmpca"
	awstypes "github.com/aws/aws-sdk-go-v2/service/acmpca/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	certificateAuthorityPermanentDeletionTimeInDaysMin     = 7
	certificateAuthorityPermanentDeletionTimeInDaysMax     = 30
	certificateAuthorityPermanentDeletionTimeInDaysDefault = certificateAuthorityPermanentDeletionTimeInDaysMax
)

// @SDKResource("aws_acmpca_certificate_authority", name="Certificate Authority")
// @Tags(identifierAttribute="id")
// @Testing(existsType="github.com/aws/aws-sdk-go/service/acmpca.CertificateAuthority", generator="acctest.RandomDomainName()", importIgnore="permanent_deletion_time_in_days")
func ResourceCertificateAuthority() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceCertificateAuthorityCreate,
		ReadWithoutTimeout:   resourceCertificateAuthorityRead,
		UpdateWithoutTimeout: resourceCertificateAuthorityUpdate,
		DeleteWithoutTimeout: resourceCertificateAuthorityDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set(
					"permanent_deletion_time_in_days",
					certificateAuthorityPermanentDeletionTimeInDaysDefault,
				)

				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
		},

		MigrateState:  resourceCertificateAuthorityMigrateState,
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// https://docs.aws.amazon.com/privateca/latest/APIReference/API_CertificateAuthorityConfiguration.html
			"certificate_authority_configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key_algorithm": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.KeyAlgorithm](),
						},
						"signing_algorithm": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.SigningAlgorithm](),
						},
						// https://docs.aws.amazon.com/privateca/latest/APIReference/API_ASN1Subject.html
						"subject": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"common_name": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 64),
									},
									"country": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 2),
									},
									"distinguished_name_qualifier": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 64),
									},
									"generation_qualifier": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 3),
									},
									"given_name": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 16),
									},
									"initials": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 5),
									},
									"locality": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 128),
									},
									"organization": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 64),
									},
									"organizational_unit": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 64),
									},
									"pseudonym": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 128),
									},
									"state": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 128),
									},
									"surname": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 40),
									},
									"title": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(0, 64),
									},
								},
							},
						},
					},
				},
			},
			"certificate_chain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_signing_request": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"key_storage_security_standard": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.KeyStorageSecurityStandard](),
			},
			"not_after": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"not_before": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permanent_deletion_time_in_days": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  certificateAuthorityPermanentDeletionTimeInDaysDefault,
				ValidateFunc: validation.IntBetween(
					certificateAuthorityPermanentDeletionTimeInDaysMin,
					certificateAuthorityPermanentDeletionTimeInDaysMax,
				),
			},
			// https://docs.aws.amazon.com/privateca/latest/APIReference/API_RevocationConfiguration.html
			"revocation_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "1" && new == "0" {
						return true
					}
					return false
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// https://docs.aws.amazon.com/privateca/latest/APIReference/API_CrlConfiguration.html
						"crl_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if old == "1" && new == "0" {
									return true
								}
								return false
							},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"custom_cname": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 253),
										DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
											// Ignore attributes if CRL configuration is not enabled
											if d.Get("revocation_configuration.0.crl_configuration.0.enabled").(bool) {
												return old == new
											}
											return true
										},
									},
									"enabled": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"expiration_in_days": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 5000),
										DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
											// Ignore attributes if CRL configuration is not enabled
											if d.Get("revocation_configuration.0.crl_configuration.0.enabled").(bool) {
												return old == new
											}
											return true
										},
									},
									"s3_bucket_name": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(3, 255),
										DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
											// Ignore attributes if CRL configuration is not enabled
											if d.Get("revocation_configuration.0.crl_configuration.0.enabled").(bool) {
												return old == new
											}
											return true
										},
									},
									"s3_object_acl": {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										ValidateDiagFunc: enum.Validate[awstypes.S3ObjectAcl](),
										DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
											// Ignore attributes if CRL configuration is not enabled
											if d.Get("revocation_configuration.0.crl_configuration.0.enabled").(bool) {
												return old == new
											}
											return true
										},
									},
								},
							},
						},
						// https://docs.aws.amazon.com/privateca/latest/APIReference/API_OcspConfiguration.html
						"ocsp_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if old == "1" && new == "0" {
									return true
								}
								return false
							},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"ocsp_custom_cname": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 253),
									},
								},
							},
						},
					},
				},
			},
			"serial": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"type": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.CertificateAuthorityTypeSubordinate,
				ValidateDiagFunc: enum.Validate[awstypes.CertificateAuthorityType](),
			},
			"usage_mode": {
				Type:             schema.TypeString,
				Computed:         true,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.CertificateAuthorityUsageMode](),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCertificateAuthorityCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAClient(ctx)

	input := &acmpca.CreateCertificateAuthorityInput{
		CertificateAuthorityConfiguration: expandCertificateAuthorityConfiguration(d.Get("certificate_authority_configuration").([]interface{})),
		CertificateAuthorityType:          awstypes.CertificateAuthorityType(d.Get("type").(string)),
		IdempotencyToken:                  aws.String(id.UniqueId()),
		RevocationConfiguration:           expandRevocationConfiguration(d.Get("revocation_configuration").([]interface{})),
		Tags:                              getTagsIn(ctx),
	}

	if v, ok := d.GetOk("key_storage_security_standard"); ok {
		input.KeyStorageSecurityStandard = awstypes.KeyStorageSecurityStandard(v.(string))
	}

	if v, ok := d.GetOk("usage_mode"); ok {
		input.UsageMode = awstypes.CertificateAuthorityUsageMode(v.(string))
	}

	// ValidationException: The ACM Private CA service account 'acm-pca-prod-pdx' requires getBucketAcl permissions for your S3 bucket 'tf-acc-test-5224996536060125340'. Check your S3 bucket permissions and try again.
	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, 1*time.Minute, func() (interface{}, error) {
		return conn.CreateCertificateAuthority(ctx, input)
	}, "ValidationException", "Check your S3 bucket permissions and try again")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ACM PCA Certificate Authority: %s", err)
	}

	d.SetId(aws.ToString(outputRaw.(*acmpca.CreateCertificateAuthorityOutput).CertificateAuthorityArn))

	if _, err := waitCertificateAuthorityCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ACM PCA Certificate Authority (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceCertificateAuthorityRead(ctx, d, meta)...)
}

func resourceCertificateAuthorityRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAClient(ctx)

	certificateAuthority, err := FindCertificateAuthorityByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ACM PCA Certificate Authority (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate Authority (%s): %s", d.Id(), err)
	}

	d.Set("arn", certificateAuthority.Arn)
	if err := d.Set("certificate_authority_configuration", flattenCertificateAuthorityConfiguration(certificateAuthority.CertificateAuthorityConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting certificate_authority_configuration: %s", err)
	}
	d.Set("enabled", (certificateAuthority.Status != awstypes.CertificateAuthorityStatusDisabled))
	d.Set("key_storage_security_standard", certificateAuthority.KeyStorageSecurityStandard)
	d.Set("not_after", aws.ToTime(certificateAuthority.NotAfter).Format(time.RFC3339))
	d.Set("not_before", aws.ToTime(certificateAuthority.NotBefore).Format(time.RFC3339))
	if err := d.Set("revocation_configuration", flattenRevocationConfiguration(certificateAuthority.RevocationConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting revocation_configuration: %s", err)
	}
	d.Set("serial", certificateAuthority.Serial)
	d.Set("type", certificateAuthority.Type)
	d.Set("usage_mode", certificateAuthority.UsageMode)

	getCertificateAuthorityCertificateInput := &acmpca.GetCertificateAuthorityCertificateInput{
		CertificateAuthorityArn: aws.String(d.Id()),
	}

	getCertificateAuthorityCertificateOutput, err := conn.GetCertificateAuthorityCertificate(ctx, getCertificateAuthorityCertificateInput)

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] ACM PCA Certificate Authority (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	// Returned when in PENDING_CERTIFICATE status
	// InvalidStateException: The certificate authority XXXXX is not in the correct state to have a certificate signing request.
	if err != nil && !errs.IsA[*awstypes.InvalidStateException](err) {
		return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate Authority (%s) Certificate: %s", d.Id(), err)
	}

	d.Set("certificate", "")
	d.Set("certificate_chain", "")
	if getCertificateAuthorityCertificateOutput != nil {
		d.Set("certificate", getCertificateAuthorityCertificateOutput.Certificate)
		d.Set("certificate_chain", getCertificateAuthorityCertificateOutput.CertificateChain)
	}

	getCertificateAuthorityCsrInput := &acmpca.GetCertificateAuthorityCsrInput{
		CertificateAuthorityArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading ACM PCA Certificate Authority Certificate Signing Request: %s", getCertificateAuthorityCsrInput)

	getCertificateAuthorityCsrOutput, err := conn.GetCertificateAuthorityCsr(ctx, getCertificateAuthorityCsrInput)

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] ACM PCA Certificate Authority (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	// Returned when in PENDING_CERTIFICATE status
	// InvalidStateException: The certificate authority XXXXX is not in the correct state to have a certificate signing request.
	if err != nil && !errs.IsA[*awstypes.InvalidStateException](err) {
		return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate Authority (%s) Certificate Signing Request: %s", d.Id(), err)
	}

	d.Set("certificate_signing_request", "")
	if getCertificateAuthorityCsrOutput != nil {
		d.Set("certificate_signing_request", getCertificateAuthorityCsrOutput.Csr)
	}

	return diags
}

func resourceCertificateAuthorityUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAClient(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &acmpca.UpdateCertificateAuthorityInput{
			CertificateAuthorityArn: aws.String(d.Id()),
		}

		if d.HasChange("enabled") {
			input.Status = awstypes.CertificateAuthorityStatusActive
			if !d.Get("enabled").(bool) {
				input.Status = awstypes.CertificateAuthorityStatusDisabled
			}
		}

		if d.HasChange("revocation_configuration") {
			input.RevocationConfiguration = expandRevocationConfiguration(d.Get("revocation_configuration").([]interface{}))
		}

		_, err := conn.UpdateCertificateAuthority(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ACM PCA Certificate Authority (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceCertificateAuthorityRead(ctx, d, meta)...)
}

func resourceCertificateAuthorityDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAClient(ctx)

	// The Certificate Authority must be in PENDING_CERTIFICATE or DISABLED state before deleting.
	updateInput := &acmpca.UpdateCertificateAuthorityInput{
		CertificateAuthorityArn: aws.String(d.Id()),
		Status:                  awstypes.CertificateAuthorityStatusDisabled,
	}
	_, err := conn.UpdateCertificateAuthority(ctx, updateInput)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}
	if err != nil && !errs.IsAErrorMessageContains[*awstypes.InvalidStateException](err, "The certificate authority must be in the ACTIVE or DISABLED state to be updated") {
		return sdkdiag.AppendErrorf(diags, "setting ACM PCA Certificate Authority (%s) to DISABLED status before deleting: %s", d.Id(), err)
	}

	deleteInput := &acmpca.DeleteCertificateAuthorityInput{
		CertificateAuthorityArn: aws.String(d.Id()),
	}

	if v, exists := d.GetOk("permanent_deletion_time_in_days"); exists {
		deleteInput.PermanentDeletionTimeInDays = aws.Int32(int32(v.(int)))
	}

	log.Printf("[INFO] Deleting ACM PCA Certificate Authority: %s", d.Id())
	_, err = conn.DeleteCertificateAuthority(ctx, deleteInput)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ACM PCA Certificate Authority (%s): %s", d.Id(), err)
	}

	return diags
}

func FindCertificateAuthorityByARN(ctx context.Context, conn *acmpca.Client, arn string) (*awstypes.CertificateAuthority, error) {
	input := &acmpca.DescribeCertificateAuthorityInput{
		CertificateAuthorityArn: aws.String(arn),
	}

	output, err := conn.DescribeCertificateAuthority(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.CertificateAuthority == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if output.CertificateAuthority.Status == awstypes.CertificateAuthorityStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(output.CertificateAuthority.Status),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.CertificateAuthority.Arn) != arn {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output.CertificateAuthority, nil
}

func statusCertificateAuthority(ctx context.Context, conn *acmpca.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindCertificateAuthorityByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitCertificateAuthorityCreated(ctx context.Context, conn *acmpca.Client, arn string, timeout time.Duration) (*awstypes.CertificateAuthority, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CertificateAuthorityStatusCreating),
		Target:  enum.Slice(awstypes.CertificateAuthorityStatusActive, awstypes.CertificateAuthorityStatusPendingCertificate),
		Refresh: statusCertificateAuthority(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.CertificateAuthority); ok {
		if output.Status == awstypes.CertificateAuthorityStatusFailed {
			tfresource.SetLastError(err, errors.New(string(output.FailureReason)))
		}

		return output, err
	}

	return nil, err
}

const (
	certificateAuthorityActiveTimeout = 1 * time.Minute
)

func expandASN1Subject(l []interface{}) *awstypes.ASN1Subject {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	subject := &awstypes.ASN1Subject{}
	if v, ok := m["common_name"]; ok && v.(string) != "" {
		subject.CommonName = aws.String(v.(string))
	}
	if v, ok := m["country"]; ok && v.(string) != "" {
		subject.Country = aws.String(v.(string))
	}
	if v, ok := m["distinguished_name_qualifier"]; ok && v.(string) != "" {
		subject.DistinguishedNameQualifier = aws.String(v.(string))
	}
	if v, ok := m["generation_qualifier"]; ok && v.(string) != "" {
		subject.GenerationQualifier = aws.String(v.(string))
	}
	if v, ok := m["given_name"]; ok && v.(string) != "" {
		subject.GivenName = aws.String(v.(string))
	}
	if v, ok := m["initials"]; ok && v.(string) != "" {
		subject.Initials = aws.String(v.(string))
	}
	if v, ok := m["locality"]; ok && v.(string) != "" {
		subject.Locality = aws.String(v.(string))
	}
	if v, ok := m["organization"]; ok && v.(string) != "" {
		subject.Organization = aws.String(v.(string))
	}
	if v, ok := m["organizational_unit"]; ok && v.(string) != "" {
		subject.OrganizationalUnit = aws.String(v.(string))
	}
	if v, ok := m["pseudonym"]; ok && v.(string) != "" {
		subject.Pseudonym = aws.String(v.(string))
	}
	if v, ok := m["state"]; ok && v.(string) != "" {
		subject.State = aws.String(v.(string))
	}
	if v, ok := m["surname"]; ok && v.(string) != "" {
		subject.Surname = aws.String(v.(string))
	}
	if v, ok := m["title"]; ok && v.(string) != "" {
		subject.Title = aws.String(v.(string))
	}

	return subject
}

func expandCertificateAuthorityConfiguration(l []interface{}) *awstypes.CertificateAuthorityConfiguration {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &awstypes.CertificateAuthorityConfiguration{
		KeyAlgorithm:     awstypes.KeyAlgorithm(m["key_algorithm"].(string)),
		SigningAlgorithm: awstypes.SigningAlgorithm(m["signing_algorithm"].(string)),
		Subject:          expandASN1Subject(m["subject"].([]interface{})),
	}

	return config
}

func expandCrlConfiguration(l []interface{}) *awstypes.CrlConfiguration {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	crlEnabled := m["enabled"].(bool)

	config := &awstypes.CrlConfiguration{
		Enabled: aws.Bool(crlEnabled),
	}

	if crlEnabled {
		if v, ok := m["custom_cname"]; ok && v.(string) != "" {
			config.CustomCname = aws.String(v.(string))
		}
		if v, ok := m["expiration_in_days"]; ok && v.(int) > 0 {
			config.ExpirationInDays = aws.Int32(int32(v.(int)))
		}
		if v, ok := m["s3_bucket_name"]; ok && v.(string) != "" {
			config.S3BucketName = aws.String(v.(string))
		}
		if v, ok := m["s3_object_acl"]; ok && v.(string) != "" {
			config.S3ObjectAcl = awstypes.S3ObjectAcl(v.(string))
		}
	}

	return config
}

func expandOcspConfiguration(l []interface{}) *awstypes.OcspConfiguration {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &awstypes.OcspConfiguration{
		Enabled: aws.Bool(m["enabled"].(bool)),
	}

	if v, ok := m["ocsp_custom_cname"]; ok && v.(string) != "" {
		config.OcspCustomCname = aws.String(v.(string))
	}

	return config
}

func expandRevocationConfiguration(l []interface{}) *awstypes.RevocationConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &awstypes.RevocationConfiguration{
		CrlConfiguration:  expandCrlConfiguration(m["crl_configuration"].([]interface{})),
		OcspConfiguration: expandOcspConfiguration(m["ocsp_configuration"].([]interface{})),
	}

	return config
}

func flattenASN1Subject(subject *awstypes.ASN1Subject) []interface{} {
	if subject == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"common_name":                  aws.ToString(subject.CommonName),
		"country":                      aws.ToString(subject.Country),
		"distinguished_name_qualifier": aws.ToString(subject.DistinguishedNameQualifier),
		"generation_qualifier":         aws.ToString(subject.GenerationQualifier),
		"given_name":                   aws.ToString(subject.GivenName),
		"initials":                     aws.ToString(subject.Initials),
		"locality":                     aws.ToString(subject.Locality),
		"organization":                 aws.ToString(subject.Organization),
		"organizational_unit":          aws.ToString(subject.OrganizationalUnit),
		"pseudonym":                    aws.ToString(subject.Pseudonym),
		"state":                        aws.ToString(subject.State),
		"surname":                      aws.ToString(subject.Surname),
		"title":                        aws.ToString(subject.Title),
	}

	return []interface{}{m}
}

func flattenCertificateAuthorityConfiguration(config *awstypes.CertificateAuthorityConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"key_algorithm":     string(config.KeyAlgorithm),
		"signing_algorithm": string(config.SigningAlgorithm),
		"subject":           flattenASN1Subject(config.Subject),
	}

	return []interface{}{m}
}

func flattenCrlConfiguration(config *awstypes.CrlConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"custom_cname":       aws.ToString(config.CustomCname),
		"enabled":            aws.ToBool(config.Enabled),
		"expiration_in_days": int(aws.ToInt32(config.ExpirationInDays)),
		"s3_bucket_name":     aws.ToString(config.S3BucketName),
		"s3_object_acl":      string(config.S3ObjectAcl),
	}

	return []interface{}{m}
}

func flattenOcspConfiguration(config *awstypes.OcspConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"enabled":           aws.ToBool(config.Enabled),
		"ocsp_custom_cname": aws.ToString(config.OcspCustomCname),
	}

	return []interface{}{m}
}

func flattenRevocationConfiguration(config *awstypes.RevocationConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"crl_configuration":  flattenCrlConfiguration(config.CrlConfiguration),
		"ocsp_configuration": flattenOcspConfiguration(config.OcspConfiguration),
	}

	return []interface{}{m}
}
