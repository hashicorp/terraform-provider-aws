// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acm

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	sdktypes "github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/types/duration"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// Maximum amount of time for ACM Certificate cross-service reference propagation.
	// Removal of ACM Certificates from API Gateway Custom Domains can take >15 minutes.
	certificateCrossServicePropagationTimeout = 20 * time.Minute

	// Maximum amount of time for ACM Certificate asynchronous DNS validation record assignment.
	// This timeout is unrelated to any creation or validation of those assigned DNS records.
	certificateDNSValidationAssignmentTimeout = 5 * time.Minute

	// CertificateRenewalTimeout is the amount of time to wait for managed renewal of a certificate
	CertificateRenewalTimeout = 1 * time.Minute

	certificateValidationMethodNone = "NONE"
)

// @SDKResource("aws_acm_certificate", name="Certificate")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/acm/types;types.CertificateDetail", tlsKey=true, importIgnore="certificate_body;private_key, generator=false)
func resourceCertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCertificateCreate,
		ReadWithoutTimeout:   resourceCertificateRead,
		UpdateWithoutTimeout: resourceCertificateUpdate,
		DeleteWithoutTimeout: resourceCertificateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_authority_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ValidateFunc:  verify.ValidARN,
				ConflictsWith: []string{"certificate_body", names.AttrPrivateKey, "validation_method"},
			},
			"certificate_body": {
				Type:          schema.TypeString,
				Optional:      true,
				RequiredWith:  []string{names.AttrPrivateKey},
				ConflictsWith: []string{"certificate_authority_arn", names.AttrDomainName, "validation_method"},
			},
			names.AttrCertificateChain: {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"certificate_authority_arn", names.AttrDomainName, "validation_method"},
			},
			names.AttrDomainName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validation.StringDoesNotMatch(regexache.MustCompile(`\.$`), "cannot end with a period"),
				ExactlyOneOf:  []string{names.AttrDomainName, names.AttrPrivateKey},
				ConflictsWith: []string{"certificate_body", names.AttrCertificateChain, names.AttrPrivateKey},
			},
			"domain_validation_options": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDomainName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_record_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_record_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"resource_record_value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Set: domainValidationOptionsHash,
			},
			"early_renewal_duration": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateHybridDuration,
				ConflictsWith:    []string{"certificate_body", names.AttrCertificateChain, names.AttrPrivateKey, "validation_method"},
			},
			"key_algorithm": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.KeyAlgorithm](),
				ConflictsWith:    []string{"certificate_body", names.AttrCertificateChain, names.AttrPrivateKey},
			},
			"not_after": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"not_before": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate_transparency_logging_preference": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          types.CertificateTransparencyLoggingPreferenceEnabled,
							ValidateDiagFunc: enum.Validate[types.CertificateTransparencyLoggingPreference](),
							ConflictsWith:    []string{"certificate_body", names.AttrCertificateChain, names.AttrPrivateKey},
						},
					},
				},
			},
			"pending_renewal": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrPrivateKey: {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ExactlyOneOf: []string{names.AttrDomainName, names.AttrPrivateKey},
			},
			"renewal_eligibility": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"renewal_summary": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"renewal_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"renewal_status_reason": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"updated_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subject_alternative_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 253),
						validation.StringDoesNotMatch(regexache.MustCompile(`\.$`), "cannot end with a period"),
					),
				},
				ConflictsWith: []string{"certificate_body", names.AttrCertificateChain, names.AttrPrivateKey},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"validation_emails": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"validation_method": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.ValidationMethod](),
				ConflictsWith:    []string{"certificate_authority_arn", "certificate_body", names.AttrCertificateChain, names.AttrPrivateKey},
			},
			"validation_option": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDomainName: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"validation_domain": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
				ConflictsWith: []string{"certificate_body", names.AttrCertificateChain, names.AttrPrivateKey},
			},
		},

		CustomizeDiff: customdiff.Sequence(
			func(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
				// Attempt to calculate the domain validation options based on domains present in domain_name and subject_alternative_names
				if diff.Get("validation_method").(string) == string(types.ValidationMethodDns) && (diff.HasChange(names.AttrDomainName) || diff.HasChange("subject_alternative_names")) {
					domainValidationOptionsList := []interface{}{map[string]interface{}{
						names.AttrDomainName: diff.Get(names.AttrDomainName).(string),
					}}

					if sanSet, ok := diff.Get("subject_alternative_names").(*schema.Set); ok {
						for _, sanRaw := range sanSet.List() {
							san, ok := sanRaw.(string)

							if !ok {
								continue
							}

							m := map[string]interface{}{
								names.AttrDomainName: san,
							}

							domainValidationOptionsList = append(domainValidationOptionsList, m)
						}
					}

					if err := diff.SetNew("domain_validation_options", schema.NewSet(domainValidationOptionsHash, domainValidationOptionsList)); err != nil {
						return fmt.Errorf("setting new domain_validation_options diff: %w", err)
					}
				}

				// ACM automatically adds the domain_name value to the list of SANs. Mimic ACM's behavior
				// so that the user doesn't need to explicitly set it themselves.
				if diff.HasChange(names.AttrDomainName) || diff.HasChange("subject_alternative_names") {
					domainName := diff.Get(names.AttrDomainName).(string)

					if sanSet, ok := diff.Get("subject_alternative_names").(*schema.Set); ok {
						sanSet.Add(domainName)
						if err := diff.SetNew("subject_alternative_names", sanSet); err != nil {
							return fmt.Errorf("setting new subject_alternative_names diff: %w", err)
						}
					}
				}

				return nil
			},
			func(_ context.Context, diff *schema.ResourceDiff, _ any) error {
				if diff.Id() == "" {
					return nil
				}

				if diff.HasChange("early_renewal_duration") {
					if duration := diff.Get("early_renewal_duration").(string); duration == "" {
						if err := diff.SetNew("pending_renewal", false); err != nil {
							return err
						}
					} else {
						if err := diff.SetNew("pending_renewal", certificateSetPendingRenewal(diff)); err != nil {
							return err
						}
					}
				} else if diff.Get("pending_renewal").(bool) {
					// Trigger a diff
					if err := diff.SetNewComputed("pending_renewal"); err != nil {
						return err
					}
				}

				return nil
			},
			verify.SetTagsDiff,
		),
	}
}

func resourceCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ACMClient(ctx)

	if _, ok := d.GetOk(names.AttrDomainName); ok {
		_, v1 := d.GetOk("certificate_authority_arn")
		_, v2 := d.GetOk("validation_method")

		if !v1 && !v2 {
			return sdkdiag.AppendErrorf(diags, "`certificate_authority_arn` or `validation_method` must be set when creating an ACM certificate")
		}

		domainName := d.Get(names.AttrDomainName).(string)
		input := &acm.RequestCertificateInput{
			DomainName:       aws.String(domainName),
			IdempotencyToken: aws.String(id.PrefixedUniqueId("tf")), // 32 character limit
			Tags:             getTagsIn(ctx),
		}

		if v, ok := d.GetOk("certificate_authority_arn"); ok {
			input.CertificateAuthorityArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("key_algorithm"); ok {
			input.KeyAlgorithm = types.KeyAlgorithm(v.(string))
		}

		if v, ok := d.GetOk("options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.Options = expandCertificateOptions(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("subject_alternative_names"); ok && v.(*schema.Set).Len() > 0 {
			input.SubjectAlternativeNames = flex.ExpandStringValueSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("validation_method"); ok {
			input.ValidationMethod = types.ValidationMethod(v.(string))
		}

		if v, ok := d.GetOk("validation_option"); ok && v.(*schema.Set).Len() > 0 {
			input.DomainValidationOptions = expandDomainValidationOptions(v.(*schema.Set).List())
		}

		output, err := conn.RequestCertificate(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "requesting ACM Certificate (%s): %s", domainName, err)
		}

		d.SetId(aws.ToString(output.CertificateArn))
	} else {
		input := &acm.ImportCertificateInput{
			Certificate: []byte(d.Get("certificate_body").(string)),
			PrivateKey:  []byte(d.Get(names.AttrPrivateKey).(string)),
			Tags:        getTagsIn(ctx),
		}

		if v, ok := d.GetOk(names.AttrCertificateChain); ok {
			input.CertificateChain = []byte(v.(string))
		}

		output, err := conn.ImportCertificate(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "importing ACM Certificate: %s", err)
		}

		d.SetId(aws.ToString(output.CertificateArn))
	}

	if _, err := waitCertificateDomainValidationsAvailable(ctx, conn, d.Id(), certificateDNSValidationAssignmentTimeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ACM Certificate (%s) to be issued: %s", d.Id(), err)
	}

	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ACMClient(ctx)

	certificate, err := findCertificateByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ACM Certificate %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ACM Certificate (%s): %s", d.Id(), err)
	}

	domainValidationOptions, validationEmails := flattenDomainValidations(certificate.DomainValidationOptions)

	d.Set(names.AttrARN, certificate.CertificateArn)
	d.Set("certificate_authority_arn", certificate.CertificateAuthorityArn)
	d.Set(names.AttrDomainName, certificate.DomainName)
	d.Set("early_renewal_duration", d.Get("early_renewal_duration"))
	if err := d.Set("domain_validation_options", domainValidationOptions); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting domain_validation_options: %s", err)
	}
	keyAlgorithmValue := string(certificate.KeyAlgorithm)
	// ACM DescribeCertificate returns hyphenated string values instead of underscore separated
	// This sets the value to the string in the ACM SDK (i.e. underscore separated)
	for _, v := range enum.Values[types.KeyAlgorithm]() {
		if strings.ReplaceAll(keyAlgorithmValue, "-", "_") == strings.ReplaceAll(v, "-", "_") {
			keyAlgorithmValue = v
			break
		}
	}
	d.Set("key_algorithm", keyAlgorithmValue)
	if certificate.NotAfter != nil {
		d.Set("not_after", aws.ToTime(certificate.NotAfter).Format(time.RFC3339))
	} else {
		d.Set("not_after", nil)
	}
	if certificate.NotBefore != nil {
		d.Set("not_before", aws.ToTime(certificate.NotBefore).Format(time.RFC3339))
	} else {
		d.Set("not_before", nil)
	}
	if certificate.Options != nil {
		if err := d.Set("options", []interface{}{flattenCertificateOptions(certificate.Options)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting options: %s", err)
		}
	} else {
		d.Set("options", nil)
	}
	d.Set("pending_renewal", certificateSetPendingRenewal(d))
	d.Set("renewal_eligibility", certificate.RenewalEligibility)
	if certificate.RenewalSummary != nil {
		if err := d.Set("renewal_summary", []interface{}{flattenRenewalSummary(certificate.RenewalSummary)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting renewal_summary: %s", err)
		}
	} else {
		d.Set("renewal_summary", nil)
	}
	d.Set(names.AttrStatus, certificate.Status)
	d.Set("subject_alternative_names", certificate.SubjectAlternativeNames)
	d.Set(names.AttrType, certificate.Type)
	d.Set("validation_emails", validationEmails)
	d.Set("validation_method", certificateValidationMethod(certificate))

	return diags
}

func resourceCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ACMClient(ctx)

	if d.HasChanges(names.AttrPrivateKey, "certificate_body", names.AttrCertificateChain) {
		oCBRaw, nCBRaw := d.GetChange("certificate_body")
		oCCRaw, nCCRaw := d.GetChange(names.AttrCertificateChain)
		oPKRaw, nPKRaw := d.GetChange(names.AttrPrivateKey)

		if !isChangeNormalizeCertRemoval(oCBRaw, nCBRaw) || !isChangeNormalizeCertRemoval(oCCRaw, nCCRaw) || !isChangeNormalizeCertRemoval(oPKRaw, nPKRaw) {
			input := &acm.ImportCertificateInput{
				Certificate:    []byte(d.Get("certificate_body").(string)),
				CertificateArn: aws.String(d.Get(names.AttrARN).(string)),
				PrivateKey:     []byte(d.Get(names.AttrPrivateKey).(string)),
			}

			if chain, ok := d.GetOk(names.AttrCertificateChain); ok {
				input.CertificateChain = []byte(chain.(string))
			}

			_, err := conn.ImportCertificate(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "importing ACM Certificate (%s): %s", d.Id(), err)
			}
		}
	} else if d.Get("pending_renewal").(bool) {
		_, err := conn.RenewCertificate(ctx, &acm.RenewCertificateInput{
			CertificateArn: aws.String(d.Get(names.AttrARN).(string)),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "renewing ACM Certificate (%s): %s", d.Id(), err)
		}

		if _, err := waitCertificateRenewed(ctx, conn, d.Get(names.AttrARN).(string), CertificateRenewalTimeout); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for ACM Certificate (%s) renewal: %s", d.Id(), err)
		}
	}

	if d.HasChange("options") {
		_, n := d.GetChange("options")
		input := &acm.UpdateCertificateOptionsInput{
			CertificateArn: aws.String(d.Get(names.AttrARN).(string)),
			Options:        expandCertificateOptions(n.([]interface{})[0].(map[string]interface{})),
		}

		_, err := conn.UpdateCertificateOptions(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ACM Certificate options (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ACMClient(ctx)

	log.Printf("[INFO] Deleting ACM Certificate: %s", d.Id())
	_, err := tfresource.RetryWhenIsA[*types.ResourceInUseException](ctx, certificateCrossServicePropagationTimeout,
		func() (interface{}, error) {
			return conn.DeleteCertificate(ctx, &acm.DeleteCertificateInput{
				CertificateArn: aws.String(d.Id()),
			})
		})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ACM Certificate (%s): %s", d.Id(), err)
	}

	return diags
}

func certificateValidationMethod(certificate *types.CertificateDetail) string {
	if certificate.Type == types.CertificateTypeAmazonIssued {
		for _, v := range certificate.DomainValidationOptions {
			return string(v.ValidationMethod)
		}
	}

	return certificateValidationMethodNone
}

func domainValidationOptionsHash(v interface{}) int {
	m, ok := v.(map[string]interface{})

	if !ok {
		return 0
	}

	if v, ok := m[names.AttrDomainName].(string); ok {
		return create.StringHashcode(v)
	}

	return 0
}

type resourceGetter interface {
	Get(key string) any
}

func certificateSetPendingRenewal(d resourceGetter) bool {
	if d.Get("renewal_eligibility") != string(types.RenewalEligibilityEligible) {
		return false
	}

	notAfterRaw := d.Get("not_after")
	if notAfterRaw == nil {
		return false
	}
	notAfter, _ := time.Parse(time.RFC3339, notAfterRaw.(string))

	earlyDuration := d.Get("early_renewal_duration").(string)

	duration, null, err := hybridDurationType(earlyDuration).Value()
	if null || err != nil {
		return false
	}

	earlyExpiration := duration.SubFrom(notAfter)

	return time.Now().After(earlyExpiration)
}

func expandCertificateOptions(tfMap map[string]interface{}) *types.CertificateOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.CertificateOptions{}

	if v, ok := tfMap["certificate_transparency_logging_preference"].(string); ok && v != "" {
		apiObject.CertificateTransparencyLoggingPreference = types.CertificateTransparencyLoggingPreference(v)
	}

	return apiObject
}

func flattenCertificateOptions(apiObject *types.CertificateOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["certificate_transparency_logging_preference"] = apiObject.CertificateTransparencyLoggingPreference

	return tfMap
}

func expandDomainValidationOption(tfMap map[string]interface{}) *types.DomainValidationOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.DomainValidationOption{}

	if v, ok := tfMap[names.AttrDomainName].(string); ok && v != "" {
		apiObject.DomainName = aws.String(v)
	}

	if v, ok := tfMap["validation_domain"].(string); ok && v != "" {
		apiObject.ValidationDomain = aws.String(v)
	}

	return apiObject
}

func expandDomainValidationOptions(tfList []interface{}) []types.DomainValidationOption {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.DomainValidationOption

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandDomainValidationOption(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenDomainValidation(apiObject types.DomainValidation) (map[string]interface{}, []string) {
	tfMap := map[string]interface{}{}
	var tfStrings []string

	if v := apiObject.ResourceRecord; v != nil {
		if v := apiObject.DomainName; v != nil {
			tfMap[names.AttrDomainName] = aws.ToString(v)
		}

		if v := v.Name; v != nil {
			tfMap["resource_record_name"] = aws.ToString(v)
		}

		tfMap["resource_record_type"] = v.Type

		if v := v.Value; v != nil {
			tfMap["resource_record_value"] = aws.ToString(v)
		}
	}

	tfStrings = apiObject.ValidationEmails

	return tfMap, tfStrings
}

func flattenDomainValidations(apiObjects []types.DomainValidation) ([]interface{}, []string) {
	if len(apiObjects) == 0 {
		return nil, nil
	}

	var tfList []interface{}
	var tfStrings []string

	for _, apiObject := range apiObjects {
		v1, v2 := flattenDomainValidation(apiObject)

		if len(v1) > 0 {
			tfList = append(tfList, v1)
		}
		if len(v2) > 0 {
			tfStrings = append(tfStrings, v2...)
		}
	}

	return tfList, tfStrings
}

func flattenRenewalSummary(apiObject *types.RenewalSummary) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["renewal_status"] = apiObject.RenewalStatus
	tfMap["renewal_status_reason"] = apiObject.RenewalStatusReason

	if v := apiObject.UpdatedAt; v != nil {
		tfMap["updated_at"] = aws.ToTime(v).Format(time.RFC3339)
	}

	return tfMap
}

func isChangeNormalizeCertRemoval(oldRaw, newRaw interface{}) bool {
	old, ok := oldRaw.(string)

	if !ok {
		return false
	}

	new, ok := newRaw.(string)

	if !ok {
		return false
	}

	// strip CRs from raw literals. Lifted from go/scanner/scanner.go
	// See https://github.com/golang/go/blob/release-branch.go1.6/src/go/scanner/scanner.go#L479
	stripCR := func(b []byte) []byte {
		c := make([]byte, len(b))
		i := 0
		for _, ch := range b {
			if ch != '\r' {
				c[i] = ch
				i++
			}
		}
		return c[:i]
	}

	newCleanVal := sha1.Sum(stripCR([]byte(strings.TrimSpace(new))))
	return hex.EncodeToString(newCleanVal[:]) == old
}

func findCertificate(ctx context.Context, conn *acm.Client, input *acm.DescribeCertificateInput) (*types.CertificateDetail, error) {
	output, err := conn.DescribeCertificate(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Certificate == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Certificate, nil
}

func findCertificateByARN(ctx context.Context, conn *acm.Client, arn string) (*types.CertificateDetail, error) {
	input := &acm.DescribeCertificateInput{
		CertificateArn: aws.String(arn),
	}

	output, err := findCertificate(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.Status; status == types.CertificateStatusValidationTimedOut {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func findCertificateRenewalByARN(ctx context.Context, conn *acm.Client, arn string) (*types.RenewalSummary, error) {
	certificate, err := findCertificateByARN(ctx, conn, arn)

	if err != nil {
		return nil, err
	}

	if certificate.RenewalSummary == nil {
		return nil, tfresource.NewEmptyResultError(arn)
	}

	return certificate.RenewalSummary, nil
}

func statusCertificateDomainValidationsAvailable(ctx context.Context, conn *acm.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		certificate, err := findCertificateByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		domainValidationsAvailable := true

		switch certificate.Type {
		case types.CertificateTypeAmazonIssued:
			domainValidationsAvailable = false

			for _, v := range certificate.DomainValidationOptions {
				if v.ResourceRecord != nil || len(v.ValidationEmails) > 0 || (v.ValidationStatus == types.DomainStatusSuccess) {
					domainValidationsAvailable = true

					break
				}
			}

		case types.CertificateTypePrivate:
			// While ACM PRIVATE certificates do not need to be validated, there is a slight delay for
			// the API to fill in all certificate details, which is during the PENDING_VALIDATION status.
			if certificate.Status == types.CertificateStatusPendingValidation {
				domainValidationsAvailable = false
			}
		}

		return certificate, strconv.FormatBool(domainValidationsAvailable), nil
	}
}

func waitCertificateDomainValidationsAvailable(ctx context.Context, conn *acm.Client, arn string, timeout time.Duration) (*types.CertificateDetail, error) {
	stateConf := &retry.StateChangeConf{
		Target:  []string{strconv.FormatBool(true)},
		Refresh: statusCertificateDomainValidationsAvailable(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.CertificateDetail); ok {
		return output, err
	}

	return nil, err
}

func statusCertificateRenewal(ctx context.Context, conn *acm.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findCertificateRenewalByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.RenewalStatus), nil
	}
}

func waitCertificateRenewed(ctx context.Context, conn *acm.Client, arn string, timeout time.Duration) (*types.RenewalSummary, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.RenewalStatusPendingAutoRenewal),
		Target:  enum.Slice(types.RenewalStatusSuccess),
		Refresh: statusCertificateRenewal(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.RenewalSummary); ok {
		if output.RenewalStatus == types.RenewalStatusFailed {
			tfresource.SetLastError(err, errors.New(string(output.RenewalStatusReason)))
		}

		return output, err
	}

	return nil, err
}

var validateHybridDuration = validation.AnyDiag(
	sdktypes.ValidateDuration,
	sdktypes.ValidateRFC3339Duration,
)

type hybridDurationType string

func (d hybridDurationType) IsNull() bool {
	return d == ""
}

func (d hybridDurationType) Value() (hybridDurationValue, bool, error) {
	if d.IsNull() {
		return nil, true, nil
	}

	value, err := parseHybridDuration(string(d))
	if err != nil {
		return nil, false, err
	}
	return value, false, nil
}

type hybridDurationValue interface {
	SubFrom(time.Time) time.Time
}

func parseHybridDuration(s string) (hybridDurationValue, error) {
	if duration, err := duration.Parse(s); err == nil {
		return rfc3339HybridDurationValue{d: duration}, nil
	}
	if duration, err := time.ParseDuration(s); err == nil {
		return goHybridDurationValue{d: duration}, nil
	}
	return nil, fmt.Errorf("unable to parse: %q", s)
}

type rfc3339HybridDurationValue struct {
	d duration.Duration
}

func (v rfc3339HybridDurationValue) SubFrom(t time.Time) time.Time {
	return duration.Sub(t, v.d)
}

type goHybridDurationValue struct {
	d time.Duration
}

func (v goHybridDurationValue) SubFrom(t time.Time) time.Time {
	return t.Add(-v.d)
}
