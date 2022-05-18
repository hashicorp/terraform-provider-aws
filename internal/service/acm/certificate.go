package acm

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// Maximum amount of time for ACM Certificate cross-service reference propagation.
	// Removal of ACM Certificates from API Gateway Custom Domains can take >15 minutes.
	AcmCertificateCrossServicePropagationTimeout = 20 * time.Minute

	// Maximum amount of time for ACM Certificate asynchronous DNS validation record assignment.
	// This timeout is unrelated to any creation or validation of those assigned DNS records.
	AcmCertificateDnsValidationAssignmentTimeout = 5 * time.Minute

	certificateValidationMethodNone = "NONE"
)

func ResourceCertificate() *schema.Resource {
	return &schema.Resource{
		Create: resourceCertificateCreate,
		Read:   resourceCertificateRead,
		Update: resourceCertificateUpdate,
		Delete: resourceCertificateDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_authority_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ValidateFunc:  verify.ValidARN,
				ConflictsWith: []string{"certificate_body", "private_key", "validation_method"},
			},
			"certificate_body": {
				Type:          schema.TypeString,
				Optional:      true,
				RequiredWith:  []string{"private_key"},
				ConflictsWith: []string{"certificate_authority_arn", "domain_name", "validation_method"},
			},
			"certificate_chain": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"certificate_authority_arn", "domain_name", "validation_method"},
			},
			"domain_name": {
				// AWS Provider 3.0.0 aws_route53_zone references no longer contain a
				// trailing period, no longer requiring a custom StateFunc
				// to prevent ACM API error
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validation.StringDoesNotMatch(regexp.MustCompile(`\.$`), "cannot end with a period"),
				ExactlyOneOf:  []string{"domain_name", "private_key"},
				ConflictsWith: []string{"certificate_body", "certificate_chain", "private_key"},
			},
			"domain_validation_options": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain_name": {
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
				Set: acmDomainValidationOptionsHash,
			},
			"options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate_transparency_logging_preference": {
							Type:          schema.TypeString,
							Optional:      true,
							ForceNew:      true,
							Default:       acm.CertificateTransparencyLoggingPreferenceEnabled,
							ValidateFunc:  validation.StringInSlice(acm.CertificateTransparencyLoggingPreference_Values(), false),
							ConflictsWith: []string{"certificate_body", "certificate_chain", "private_key"},
						},
					},
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if _, ok := d.GetOk("private_key"); ok {
						// ignore diffs for imported certs; they have a different logging preference
						// default to requested certs which can't be changed by the ImportCertificate API
						return true
					}
					// behave just like verify.SuppressMissingOptionalConfigurationBlock() for requested certs
					return old == "1" && new == "0"
				},
			},
			"private_key": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ExactlyOneOf: []string{"domain_name", "private_key"},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subject_alternative_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Schema{
					// AWS Provider 3.0.0 aws_route53_zone references no longer contain a
					// trailing period, no longer requiring a custom StateFunc
					// to prevent ACM API error
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 253),
						validation.StringDoesNotMatch(regexp.MustCompile(`\.$`), "cannot end with a period"),
					),
				},
				ConflictsWith: []string{"certificate_body", "certificate_chain", "private_key"},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"validation_emails": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"validation_method": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validation.StringInSlice(append(acm.ValidationMethod_Values(), certificateValidationMethodNone), false),
				ConflictsWith: []string{"certificate_authority_arn", "certificate_body", "certificate_chain", "private_key"},
			},
			"validation_option": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain_name": {
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
				ConflictsWith: []string{"certificate_body", "certificate_chain", "private_key"},
			},
		},

		CustomizeDiff: customdiff.Sequence(
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				// Attempt to calculate the domain validation options based on domains present in domain_name and subject_alternative_names
				if diff.Get("validation_method").(string) == acm.ValidationMethodDns && (diff.HasChange("domain_name") || diff.HasChange("subject_alternative_names")) {
					domainValidationOptionsList := []interface{}{map[string]interface{}{
						// AWS Provider 3.0 -- plan-time validation prevents "domain_name"
						// argument to accept a string with trailing period; thus, trim of trailing period
						// no longer required here
						"domain_name": diff.Get("domain_name").(string),
					}}

					if sanSet, ok := diff.Get("subject_alternative_names").(*schema.Set); ok {
						for _, sanRaw := range sanSet.List() {
							san, ok := sanRaw.(string)

							if !ok {
								continue
							}

							m := map[string]interface{}{
								// AWS Provider 3.0 -- plan-time validation prevents "subject_alternative_names"
								// argument to accept strings with trailing period; thus, trim of trailing period
								// no longer required here
								"domain_name": san,
							}

							domainValidationOptionsList = append(domainValidationOptionsList, m)
						}
					}

					if err := diff.SetNew("domain_validation_options", schema.NewSet(acmDomainValidationOptionsHash, domainValidationOptionsList)); err != nil {
						return fmt.Errorf("error setting new domain_validation_options diff: %w", err)
					}
				}

				// ACM automatically adds the domain_name value to the list of SANs. Mimic ACM's behavior
				// so that the user doesn't need to explicitly set it themselves.
				if diff.HasChange("domain_name") || diff.HasChange("subject_alternative_names") {
					domain_name := diff.Get("domain_name").(string)

					if sanSet, ok := diff.Get("subject_alternative_names").(*schema.Set); ok {
						sanSet.Add(domain_name)
						if err := diff.SetNew("subject_alternative_names", sanSet); err != nil {
							return fmt.Errorf("error setting new subject_alternative_names diff: %w", err)
						}
					}
				}

				return nil
			},
			verify.SetTagsDiff,
		),
	}
}

func resourceCertificateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if _, ok := d.GetOk("domain_name"); ok {
		_, v1 := d.GetOk("certificate_authority_arn")
		_, v2 := d.GetOk("validation_method")

		if !v1 && !v2 {
			return errors.New("`certificate_authority_arn` or `validation_method` must be set when creating an ACM certificate")
		}

		domainName := d.Get("domain_name").(string)
		input := &acm.RequestCertificateInput{
			DomainName:       aws.String(domainName),
			IdempotencyToken: aws.String(resource.PrefixedUniqueId("tf")), // 32 character limit
		}

		if v, ok := d.GetOk("certificate_authority_arn"); ok {
			input.CertificateAuthorityArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.Options = expandCertificateOptions(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("subject_alternative_names"); ok {
			for _, v := range v.(*schema.Set).List() {
				input.SubjectAlternativeNames = append(input.SubjectAlternativeNames, aws.String(v.(string)))
			}
		}

		if v, ok := d.GetOk("validation_method"); ok {
			input.ValidationMethod = aws.String(v.(string))
		}

		if v, ok := d.GetOk("validation_option"); ok && v.(*schema.Set).Len() > 0 {
			input.DomainValidationOptions = expandDomainValidationOptions(v.(*schema.Set).List())
		}

		if len(tags) > 0 {
			input.Tags = Tags(tags.IgnoreAWS())
		}

		log.Printf("[DEBUG] Requesting ACM Certificate: %s", input)
		output, err := conn.RequestCertificate(input)

		if err != nil {
			return fmt.Errorf("requesting ACM Certificate (%s): %w", domainName, err)
		}

		d.SetId(aws.StringValue(output.CertificateArn))
	} else {
		input := &acm.ImportCertificateInput{
			Certificate: []byte(d.Get("certificate_body").(string)),
			PrivateKey:  []byte(d.Get("private_key").(string)),
		}

		if v, ok := d.GetOk("certificate_chain"); ok {
			input.CertificateChain = []byte(v.(string))
		}

		if len(tags) > 0 {
			input.Tags = Tags(tags.IgnoreAWS())
		}

		output, err := conn.ImportCertificate(input)

		if err != nil {
			return fmt.Errorf("importing ACM Certificate: %w", err)
		}

		d.SetId(aws.StringValue(output.CertificateArn))
	}

	if _, err := waitCertificateDomainValidationsAvailable(conn, d.Id(), AcmCertificateDnsValidationAssignmentTimeout); err != nil {
		return fmt.Errorf("waiting for ACM Certificate (%s) to be issued: %w", d.Id(), err)
	}

	return resourceCertificateRead(d, meta)
}

func resourceCertificateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	certificate, err := FindCertificateByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ACM Certificate %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading ACM Certificate (%s): %w", d.Id(), err)
	}

	domainValidationOptions, validationEmails := flattenDomainValidations(certificate.DomainValidationOptions)

	d.Set("arn", certificate.CertificateArn)
	d.Set("certificate_authority_arn", certificate.CertificateAuthorityArn)
	d.Set("domain_name", certificate.DomainName)
	if err := d.Set("domain_validation_options", domainValidationOptions); err != nil {
		return fmt.Errorf("error setting domain_validation_options: %w", err)
	}
	if certificate.Options != nil {
		if err := d.Set("options", []interface{}{flattenCertificateOptions(certificate.Options)}); err != nil {
			return fmt.Errorf("error setting options: %w", err)
		}
	} else {
		d.Set("options", nil)
	}
	d.Set("status", certificate.Status)
	d.Set("subject_alternative_names", aws.StringValueSlice(certificate.SubjectAlternativeNames))
	d.Set("validation_emails", validationEmails)
	d.Set("validation_method", certificateValidationMethod(certificate))

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("listing tags for ACM Certificate (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceCertificateUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMConn

	if d.HasChanges("private_key", "certificate_body", "certificate_chain") {
		// Prior to version 3.0.0 of the Terraform AWS Provider, these attributes were stored in state as hashes.
		// If the changes to these attributes are only changes only match updating the state value, then skip the API call.
		oCBRaw, nCBRaw := d.GetChange("certificate_body")
		oCCRaw, nCCRaw := d.GetChange("certificate_chain")
		oPKRaw, nPKRaw := d.GetChange("private_key")

		if !isChangeNormalizeCertRemoval(oCBRaw, nCBRaw) || !isChangeNormalizeCertRemoval(oCCRaw, nCCRaw) || !isChangeNormalizeCertRemoval(oPKRaw, nPKRaw) {
			input := &acm.ImportCertificateInput{
				Certificate:    []byte(d.Get("certificate_body").(string)),
				CertificateArn: aws.String(d.Get("arn").(string)),
				PrivateKey:     []byte(d.Get("private_key").(string)),
			}

			if chain, ok := d.GetOk("certificate_chain"); ok {
				input.CertificateChain = []byte(chain.(string))
			}

			_, err := conn.ImportCertificate(input)

			if err != nil {
				return fmt.Errorf("re-importing ACM Certificate (%s): %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceCertificateRead(d, meta)
}

func resourceCertificateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMConn

	log.Printf("[INFO] Deleting ACM Certificate: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(AcmCertificateCrossServicePropagationTimeout,
		func() (interface{}, error) {
			return conn.DeleteCertificate(&acm.DeleteCertificateInput{
				CertificateArn: aws.String(d.Id()),
			})
		}, acm.ErrCodeResourceInUseException)

	if tfawserr.ErrCodeEquals(err, acm.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting ACM Certificate (%s): %w", d.Id(), err)
	}

	return nil
}

func certificateValidationMethod(certificate *acm.CertificateDetail) string {
	if aws.StringValue(certificate.Type) == acm.CertificateTypeAmazonIssued {
		for _, v := range certificate.DomainValidationOptions {
			if v.ValidationMethod != nil {
				return aws.StringValue(v.ValidationMethod)
			}
		}
	}

	return certificateValidationMethodNone
}

func acmDomainValidationOptionsHash(v interface{}) int {
	m, ok := v.(map[string]interface{})

	if !ok {
		return 0
	}

	if v, ok := m["domain_name"].(string); ok {
		return create.StringHashcode(v)
	}

	return 0
}

func expandCertificateOptions(tfMap map[string]interface{}) *acm.CertificateOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &acm.CertificateOptions{}

	if v, ok := tfMap["certificate_transparency_logging_preference"].(string); ok && v != "" {
		apiObject.CertificateTransparencyLoggingPreference = aws.String(v)
	}

	return apiObject
}

func flattenCertificateOptions(apiObject *acm.CertificateOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CertificateTransparencyLoggingPreference; v != nil {
		tfMap["certificate_transparency_logging_preference"] = aws.StringValue(v)
	}

	return tfMap
}

func expandDomainValidationOption(tfMap map[string]interface{}) *acm.DomainValidationOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &acm.DomainValidationOption{}

	if v, ok := tfMap["domain_name"].(string); ok && v != "" {
		apiObject.DomainName = aws.String(v)
	}

	if v, ok := tfMap["validation_domain"].(string); ok && v != "" {
		apiObject.ValidationDomain = aws.String(v)
	}

	return apiObject
}

func expandDomainValidationOptions(tfList []interface{}) []*acm.DomainValidationOption {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*acm.DomainValidationOption

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandDomainValidationOption(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenDomainValidation(apiObject *acm.DomainValidation) (map[string]interface{}, []string) {
	if apiObject == nil {
		return nil, nil
	}

	tfMap := map[string]interface{}{}
	var tfStrings []string

	if v := apiObject.ResourceRecord; v != nil {
		if v := apiObject.DomainName; v != nil {
			tfMap["domain_name"] = aws.StringValue(v)
		}

		if v := v.Name; v != nil {
			tfMap["resource_record_name"] = aws.StringValue(v)
		}

		if v := v.Type; v != nil {
			tfMap["resource_record_type"] = aws.StringValue(v)
		}

		if v := v.Value; v != nil {
			tfMap["resource_record_value"] = aws.StringValue(v)
		}
	}

	tfStrings = aws.StringValueSlice(apiObject.ValidationEmails)

	return tfMap, tfStrings
}

func flattenDomainValidations(apiObjects []*acm.DomainValidation) ([]interface{}, []string) {
	if len(apiObjects) == 0 {
		return nil, nil
	}

	var tfList []interface{}
	var tfStrings []string

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

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

func findCertificate(conn *acm.ACM, input *acm.DescribeCertificateInput) (*acm.CertificateDetail, error) {
	output, err := conn.DescribeCertificate(input)

	if tfawserr.ErrCodeEquals(err, acm.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
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

func FindCertificateByARN(conn *acm.ACM, arn string) (*acm.CertificateDetail, error) {
	input := &acm.DescribeCertificateInput{
		CertificateArn: aws.String(arn),
	}

	output, err := findCertificate(conn, input)

	if err != nil {
		return nil, err
	}

	if status := aws.StringValue(output.Status); status == acm.CertificateStatusValidationTimedOut {
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}

func statusCertificateDomainValidationsAvailable(conn *acm.ACM, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		certificate, err := FindCertificateByARN(conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		domainValidationsAvailable := true

		switch aws.StringValue(certificate.Type) {
		case acm.CertificateTypeAmazonIssued:
			domainValidationsAvailable = false

			for _, v := range certificate.DomainValidationOptions {
				if v.ResourceRecord != nil || len(v.ValidationEmails) > 0 || (aws.StringValue(v.ValidationStatus) == acm.DomainStatusSuccess) {
					domainValidationsAvailable = true

					break
				}
			}

		case acm.CertificateTypePrivate:
			// While ACM PRIVATE certificates do not need to be validated, there is a slight delay for
			// the API to fill in all certificate details, which is during the PENDING_VALIDATION status.
			if aws.StringValue(certificate.Status) == acm.DomainStatusPendingValidation {
				domainValidationsAvailable = false
			}
		}

		return certificate, strconv.FormatBool(domainValidationsAvailable), nil
	}
}

func waitCertificateDomainValidationsAvailable(conn *acm.ACM, arn string, timeout time.Duration) (*acm.CertificateDetail, error) {
	stateConf := &resource.StateChangeConf{
		Target:  []string{strconv.FormatBool(true)},
		Refresh: statusCertificateDomainValidationsAvailable(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*acm.CertificateDetail); ok {
		return output, err
	}

	return nil, err
}
