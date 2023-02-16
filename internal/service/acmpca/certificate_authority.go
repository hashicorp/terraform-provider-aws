package acmpca

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	certificateAuthorityPermanentDeletionTimeInDaysMin     = 7
	certificateAuthorityPermanentDeletionTimeInDaysMax     = 30
	certificateAuthorityPermanentDeletionTimeInDaysDefault = certificateAuthorityPermanentDeletionTimeInDaysMax
)

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
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(acmpca.KeyAlgorithm_Values(), false),
						},
						"signing_algorithm": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(acmpca.SigningAlgorithm_Values(), false),
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
									},
									"enabled": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									// ValidationException: 1 validation error detected: Value null or empty at 'expirationInDays' failed to satisfy constraint: Member must not be null or empty.
									// InvalidParameter: 1 validation error(s) found. minimum field value of 1, CreateCertificateAuthorityInput.RevocationConfiguration.CrlConfiguration.ExpirationInDays.
									"expiration_in_days": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(1, 5000),
									},
									"s3_bucket_name": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 255),
									},
									"s3_object_acl": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.StringInSlice(acmpca.S3ObjectAcl_Values(), false),
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
			// See https://github.com/hashicorp/terraform-provider-aws/issues/17832 for deprecation / removal status
			"status": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "The reported value of the \"status\" attribute is often inaccurate. Use the resource's \"enabled\" attribute to explicitly set status.",
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      acmpca.CertificateAuthorityTypeSubordinate,
				ValidateFunc: validation.StringInSlice(acmpca.CertificateAuthorityType_Values(), false),
			},
			"usage_mode": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(acmpca.CertificateAuthorityUsageMode_Values(), false),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCertificateAuthorityCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &acmpca.CreateCertificateAuthorityInput{
		CertificateAuthorityConfiguration: expandCertificateAuthorityConfiguration(d.Get("certificate_authority_configuration").([]interface{})),
		CertificateAuthorityType:          aws.String(d.Get("type").(string)),
		IdempotencyToken:                  aws.String(resource.UniqueId()),
		RevocationConfiguration:           expandRevocationConfiguration(d.Get("revocation_configuration").([]interface{})),
	}

	if v, ok := d.GetOk("usage_mode"); ok {
		input.UsageMode = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating ACM PCA Certificate Authority: %s", input)
	var output *acmpca.CreateCertificateAuthorityOutput
	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		var err error
		output, err = conn.CreateCertificateAuthorityWithContext(ctx, input)
		if err != nil {
			// ValidationException: The ACM Private CA service account 'acm-pca-prod-pdx' requires getBucketAcl permissions for your S3 bucket 'tf-acc-test-5224996536060125340'. Check your S3 bucket permissions and try again.
			if tfawserr.ErrMessageContains(err, "ValidationException", "Check your S3 bucket permissions and try again") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		output, err = conn.CreateCertificateAuthorityWithContext(ctx, input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ACM PCA Certificate Authority: %s", err)
	}

	d.SetId(aws.StringValue(output.CertificateAuthorityArn))

	_, err = waitCertificateAuthorityCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ACM PCA Certificate Authority %q to be active or pending certificate: %s", d.Id(), err)
	}

	return append(diags, resourceCertificateAuthorityRead(ctx, d, meta)...)
}

func resourceCertificateAuthorityRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	certificateAuthority, err := FindCertificateAuthorityByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] ACM PCA Certificate Authority (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate Authority (%s): %s", d.Id(), err)
	}

	if certificateAuthority == nil || aws.StringValue(certificateAuthority.Status) == acmpca.CertificateAuthorityStatusDeleted {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate Authority (%s): not found or deleted", d.Id())
		}

		log.Printf("[WARN] ACM PCA Certificate Authority (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("arn", certificateAuthority.Arn)
	if err := d.Set("certificate_authority_configuration", flattenCertificateAuthorityConfiguration(certificateAuthority.CertificateAuthorityConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting certificate_authority_configuration: %s", err)
	}
	d.Set("enabled", (aws.StringValue(certificateAuthority.Status) != acmpca.CertificateAuthorityStatusDisabled))
	d.Set("not_after", aws.TimeValue(certificateAuthority.NotAfter).Format(time.RFC3339))
	d.Set("not_before", aws.TimeValue(certificateAuthority.NotBefore).Format(time.RFC3339))
	if err := d.Set("revocation_configuration", flattenRevocationConfiguration(certificateAuthority.RevocationConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting revocation_configuration: %s", err)
	}
	d.Set("serial", certificateAuthority.Serial)
	d.Set("status", certificateAuthority.Status)
	d.Set("type", certificateAuthority.Type)
	d.Set("usage_mode", certificateAuthority.UsageMode)

	getCertificateAuthorityCertificateInput := &acmpca.GetCertificateAuthorityCertificateInput{
		CertificateAuthorityArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading ACM PCA Certificate Authority Certificate: %s", getCertificateAuthorityCertificateInput)

	getCertificateAuthorityCertificateOutput, err := conn.GetCertificateAuthorityCertificateWithContext(ctx, getCertificateAuthorityCertificateInput)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] ACM PCA Certificate Authority (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	// Returned when in PENDING_CERTIFICATE status
	// InvalidStateException: The certificate authority XXXXX is not in the correct state to have a certificate signing request.
	if err != nil && !tfawserr.ErrCodeEquals(err, acmpca.ErrCodeInvalidStateException) {
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

	getCertificateAuthorityCsrOutput, err := conn.GetCertificateAuthorityCsrWithContext(ctx, getCertificateAuthorityCsrInput)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] ACM PCA Certificate Authority (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	// Returned when in PENDING_CERTIFICATE status
	// InvalidStateException: The certificate authority XXXXX is not in the correct state to have a certificate signing request.
	if err != nil && !tfawserr.ErrCodeEquals(err, acmpca.ErrCodeInvalidStateException) {
		return sdkdiag.AppendErrorf(diags, "reading ACM PCA Certificate Authority (%s) Certificate Signing Request: %s", d.Id(), err)
	}

	d.Set("certificate_signing_request", "")
	if getCertificateAuthorityCsrOutput != nil {
		d.Set("certificate_signing_request", getCertificateAuthorityCsrOutput.Csr)
	}

	tags, err := ListTags(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for ACM PCA Certificate Authority (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceCertificateAuthorityUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAConn()
	updateCertificateAuthority := false

	input := &acmpca.UpdateCertificateAuthorityInput{
		CertificateAuthorityArn: aws.String(d.Id()),
	}

	if d.HasChange("enabled") {
		input.Status = aws.String(acmpca.CertificateAuthorityStatusActive)
		if !d.Get("enabled").(bool) {
			input.Status = aws.String(acmpca.CertificateAuthorityStatusDisabled)
		}
		updateCertificateAuthority = true
	}

	if d.HasChange("revocation_configuration") {
		input.RevocationConfiguration = expandRevocationConfiguration(d.Get("revocation_configuration").([]interface{}))
		updateCertificateAuthority = true
	}

	if updateCertificateAuthority {
		log.Printf("[DEBUG] Updating ACM PCA Certificate Authority: %s", input)
		_, err := conn.UpdateCertificateAuthorityWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ACM PCA Certificate Authority: %s", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ACM PCA Certificate Authority (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceCertificateAuthorityRead(ctx, d, meta)...)
}

func resourceCertificateAuthorityDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAConn()

	// The Certificate Authority must be in PENDING_CERTIFICATE or DISABLED state before deleting.
	updateInput := &acmpca.UpdateCertificateAuthorityInput{
		CertificateAuthorityArn: aws.String(d.Id()),
		Status:                  aws.String(acmpca.CertificateAuthorityStatusDisabled),
	}
	_, err := conn.UpdateCertificateAuthorityWithContext(ctx, updateInput)
	if tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) {
		return diags
	}
	if err != nil && !tfawserr.ErrMessageContains(err, acmpca.ErrCodeInvalidStateException, "The certificate authority must be in the ACTIVE or DISABLED state to be updated") {
		return sdkdiag.AppendErrorf(diags, "setting ACM PCA Certificate Authority (%s) to DISABLED status before deleting: %s", d.Id(), err)
	}

	deleteInput := &acmpca.DeleteCertificateAuthorityInput{
		CertificateAuthorityArn: aws.String(d.Id()),
	}

	if v, exists := d.GetOk("permanent_deletion_time_in_days"); exists {
		deleteInput.PermanentDeletionTimeInDays = aws.Int64(int64(v.(int)))
	}

	log.Printf("[INFO] Deleting ACM PCA Certificate Authority: %s", d.Id())
	_, err = conn.DeleteCertificateAuthorityWithContext(ctx, deleteInput)
	if tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ACM PCA Certificate Authority (%s): %s", d.Id(), err)
	}

	return diags
}

func expandASN1Subject(l []interface{}) *acmpca.ASN1Subject {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	subject := &acmpca.ASN1Subject{}
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

func expandCertificateAuthorityConfiguration(l []interface{}) *acmpca.CertificateAuthorityConfiguration {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &acmpca.CertificateAuthorityConfiguration{
		KeyAlgorithm:     aws.String(m["key_algorithm"].(string)),
		SigningAlgorithm: aws.String(m["signing_algorithm"].(string)),
		Subject:          expandASN1Subject(m["subject"].([]interface{})),
	}

	return config
}

func expandCrlConfiguration(l []interface{}) *acmpca.CrlConfiguration {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &acmpca.CrlConfiguration{
		Enabled: aws.Bool(m["enabled"].(bool)),
	}

	if v, ok := m["custom_cname"]; ok && v.(string) != "" {
		config.CustomCname = aws.String(v.(string))
	}
	if v, ok := m["expiration_in_days"]; ok && v.(int) > 0 {
		config.ExpirationInDays = aws.Int64(int64(v.(int)))
	}
	if v, ok := m["s3_bucket_name"]; ok && v.(string) != "" {
		config.S3BucketName = aws.String(v.(string))
	}
	if v, ok := m["s3_object_acl"]; ok && v.(string) != "" {
		config.S3ObjectAcl = aws.String(v.(string))
	}

	return config
}

func expandOcspConfiguration(l []interface{}) *acmpca.OcspConfiguration {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &acmpca.OcspConfiguration{
		Enabled: aws.Bool(m["enabled"].(bool)),
	}

	if v, ok := m["ocsp_custom_cname"]; ok && v.(string) != "" {
		config.OcspCustomCname = aws.String(v.(string))
	}

	return config
}

func expandRevocationConfiguration(l []interface{}) *acmpca.RevocationConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &acmpca.RevocationConfiguration{
		CrlConfiguration:  expandCrlConfiguration(m["crl_configuration"].([]interface{})),
		OcspConfiguration: expandOcspConfiguration(m["ocsp_configuration"].([]interface{})),
	}

	return config
}

func flattenASN1Subject(subject *acmpca.ASN1Subject) []interface{} {
	if subject == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"common_name":                  aws.StringValue(subject.CommonName),
		"country":                      aws.StringValue(subject.Country),
		"distinguished_name_qualifier": aws.StringValue(subject.DistinguishedNameQualifier),
		"generation_qualifier":         aws.StringValue(subject.GenerationQualifier),
		"given_name":                   aws.StringValue(subject.GivenName),
		"initials":                     aws.StringValue(subject.Initials),
		"locality":                     aws.StringValue(subject.Locality),
		"organization":                 aws.StringValue(subject.Organization),
		"organizational_unit":          aws.StringValue(subject.OrganizationalUnit),
		"pseudonym":                    aws.StringValue(subject.Pseudonym),
		"state":                        aws.StringValue(subject.State),
		"surname":                      aws.StringValue(subject.Surname),
		"title":                        aws.StringValue(subject.Title),
	}

	return []interface{}{m}
}

func flattenCertificateAuthorityConfiguration(config *acmpca.CertificateAuthorityConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"key_algorithm":     aws.StringValue(config.KeyAlgorithm),
		"signing_algorithm": aws.StringValue(config.SigningAlgorithm),
		"subject":           flattenASN1Subject(config.Subject),
	}

	return []interface{}{m}
}

func flattenCrlConfiguration(config *acmpca.CrlConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"custom_cname":       aws.StringValue(config.CustomCname),
		"enabled":            aws.BoolValue(config.Enabled),
		"expiration_in_days": int(aws.Int64Value(config.ExpirationInDays)),
		"s3_bucket_name":     aws.StringValue(config.S3BucketName),
		"s3_object_acl":      aws.StringValue(config.S3ObjectAcl),
	}

	return []interface{}{m}
}

func flattenOcspConfiguration(config *acmpca.OcspConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"enabled":           aws.BoolValue(config.Enabled),
		"ocsp_custom_cname": aws.StringValue(config.OcspCustomCname),
	}

	return []interface{}{m}
}

func flattenRevocationConfiguration(config *acmpca.RevocationConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"crl_configuration":  flattenCrlConfiguration(config.CrlConfiguration),
		"ocsp_configuration": flattenOcspConfiguration(config.OcspConfiguration),
	}

	return []interface{}{m}
}
