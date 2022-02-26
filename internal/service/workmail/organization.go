package workmail

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workmail"
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
	// Maximum amount of time for ACM Organization cross-service reference propagation.
	// Removal of ACM Organizations from API Gateway Custom Domains can take >15 minutes.
	AcmOrganizationCrossServicePropagationTimeout = 20 * time.Minute

	// Maximum amount of time for ACM Organization asynchronous DNS validation record assignment.
	// This timeout is unrelated to any creation or validation of those assigned DNS records.
	AcmOrganizationDnsValidationAssignmentTimeout = 5 * time.Minute
)

	// The organization alias.
	//
	// Alias is a required field
	Alias *string `min:"1" type:"string" required:"true"`

	// The idempotency token associated with the request.
	ClientToken *string `min:"1" type:"string" idempotencyToken:"true"`

	// The AWS Directory Service directory ID.
	DirectoryId *string `min:"12" type:"string"`

	// The email domains to associate with the organization.
	Domains []*Domain `type:"list"`

	// When true, allows organization interoperability between Amazon WorkMail and
	// Microsoft Exchange. Can only be set to true if an AD Connector directory
	// ID is included in the request.
	EnableInteroperability *bool `type:"boolean"`

	// The Amazon Resource Name (ARN) of a customer managed master key from AWS
	// KMS.
	KmsKeyArn *string `min:"20" type:"string"`

func ResourceOrganization() *schema.Resource {
	return &schema.Resource{
		Create: resourceOrganizationCreate,
		Read:   resourceOrganizationRead,
		Update: resourceOrganizationUpdate,
		Delete: resourceOrganizationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"alias": {
				Type:     schema.TypeString,
				Required: true,
			},
			"client_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"certificate_authority_arn": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"domains": {
				// AWS Provider 3.0.0 aws_route53_zone references no longer contain a
				// trailing period, no longer requiring a custom StateFunc
				// to prevent ACM API error
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"private_key", "certificate_body", "certificate_chain"},
				ValidateFunc:  validation.StringDoesNotMatch(regexp.MustCompile(`\.$`), "cannot end with a period"),
			},
			"enable_interoperability": {
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
				Set:           schema.HashString,
				ConflictsWith: []string{"private_key", "certificate_body", "certificate_chain"},
			},
			"validation_method": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"private_key", "certificate_body", "certificate_chain", "certificate_authority_arn"},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domains": {
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
				Set: workmailDomainValidationOptionsHash,
			},
			"kms_key_arn": {
				Type:     schema.TypeList,
				Optional: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceOrganizationCreate(d *schema.ResourceData, meta interface{}) error {
	if _, ok := d.GetOk("domain_name"); ok {
		if _, ok := d.GetOk("certificate_authority_arn"); ok {
			return resourceOrganizationCreateRequested(d, meta)
		}

		if _, ok := d.GetOk("validation_method"); !ok {
			return errors.New("validation_method must be set when creating a certificate")
		}
		return resourceOrganizationCreateRequested(d, meta)
	} else if _, ok := d.GetOk("private_key"); ok {
		if _, ok := d.GetOk("certificate_body"); !ok {
			return errors.New("certificate_body must be set when importing a certificate with private_key")
		}
		return resourceOrganizationCreateImported(d, meta)
	}
	return errors.New("certificate must be imported (private_key) or created (domain_name)")
}

func resourceOrganizationCreateImported(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &workmail.ImportOrganizationInput{
		Organization: []byte(d.Get("certificate_body").(string)),
		PrivateKey:   []byte(d.Get("private_key").(string)),
	}

	if v, ok := d.GetOk("certificate_chain"); ok {
		input.OrganizationChain = []byte(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.ImportOrganization(input)

	if err != nil {
		return fmt.Errorf("error importing ACM Organization: %w", err)
	}

	d.SetId(aws.StringValue(output.OrganizationArn))

	return resourceOrganizationRead(d, meta)
}

func resourceOrganizationCreateRequested(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	params := &workmail.RequestOrganizationInput{
		DomainName:       aws.String(d.Get("domain_name").(string)),
		IdempotencyToken: aws.String(resource.PrefixedUniqueId("tf")), // 32 character limit
		Options:          expandAcmOrganizationOptions(d.Get("options").([]interface{})),
	}

	if len(tags) > 0 {
		params.Tags = Tags(tags.IgnoreAWS())
	}

	if caARN, ok := d.GetOk("certificate_authority_arn"); ok {
		params.OrganizationAuthorityArn = aws.String(caARN.(string))
	}

	if sans, ok := d.GetOk("subject_alternative_names"); ok {
		subjectAlternativeNames := make([]*string, len(sans.(*schema.Set).List()))
		for i, sanRaw := range sans.(*schema.Set).List() {
			subjectAlternativeNames[i] = aws.String(sanRaw.(string))
		}
		params.SubjectAlternativeNames = subjectAlternativeNames
	}

	if v, ok := d.GetOk("validation_method"); ok {
		params.ValidationMethod = aws.String(v.(string))
	}

	log.Printf("[DEBUG] ACM Organization Request: %#v", params)
	resp, err := conn.RequestOrganization(params)

	if err != nil {
		return fmt.Errorf("Error requesting certificate: %s", err)
	}

	d.SetId(aws.StringValue(resp.OrganizationArn))

	return resourceOrganizationRead(d, meta)
}

func resourceOrganizationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	params := &workmail.DescribeOrganizationInput{
		OrganizationArn: aws.String(d.Id()),
	}

	return resource.Retry(AcmOrganizationDnsValidationAssignmentTimeout, func() *resource.RetryError {
		resp, err := conn.DescribeOrganization(params)

		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, workmail.ErrCodeResourceNotFoundException) {
			log.Printf("[WARN] ACM Organization (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("error reading ACM Organization (%s): %w", d.Id(), err))
		}

		if resp == nil || resp.Organization == nil {
			return resource.NonRetryableError(fmt.Errorf("error reading ACM Organization (%s): empty response", d.Id()))
		}

		if !d.IsNewResource() && aws.StringValue(resp.Organization.Status) == workmail.OrganizationStatusValidationTimedOut {
			log.Printf("[WARN] ACM Organization (%s) validation timed out, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		d.Set("domain_name", resp.Organization.DomainName)
		d.Set("arn", resp.Organization.OrganizationArn)
		d.Set("certificate_authority_arn", resp.Organization.OrganizationAuthorityArn)

		if err := d.Set("subject_alternative_names", cleanUpSubjectAlternativeNames(resp.Organization)); err != nil {
			return resource.NonRetryableError(err)
		}

		domainValidationOptions, emailValidationOptions, err := convertValidationOptions(resp.Organization)

		if err != nil {
			return resource.RetryableError(err)
		}

		if err := d.Set("domain_validation_options", domainValidationOptions); err != nil {
			return resource.NonRetryableError(err)
		}
		if err := d.Set("validation_emails", emailValidationOptions); err != nil {
			return resource.NonRetryableError(err)
		}

		d.Set("validation_method", resourceOrganizationValidationMethod(resp.Organization))

		if err := d.Set("options", flattenAcmOrganizationOptions(resp.Organization.Options)); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error setting certificate options: %s", err))
		}

		d.Set("status", resp.Organization.Status)

		tags, err := ListTags(conn, d.Id())

		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("error listing tags for ACM Organization (%s): %s", d.Id(), err))
		}

		tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

		//lintignore:AWSR002
		if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error setting tags: %w", err))
		}

		if err := d.Set("tags_all", tags.Map()); err != nil {
			return resource.NonRetryableError(fmt.Errorf("error setting tags_all: %w", err))
		}

		return nil
	})
}

func resourceOrganizationValidationMethod(certificate *workmail.OrganizationDetail) string {
	if aws.StringValue(certificate.Type) == workmail.OrganizationTypeAmazonIssued {
		for _, domainValidation := range certificate.DomainValidationOptions {
			if domainValidation.ValidationMethod != nil {
				return aws.StringValue(domainValidation.ValidationMethod)
			}
		}
	}

	return "NONE"
}

func resourceOrganizationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMConn

	if d.HasChanges("private_key", "certificate_body", "certificate_chain") {
		// Prior to version 3.0.0 of the Terraform AWS Provider, these attributes were stored in state as hashes.
		// If the changes to these attributes are only changes only match updating the state value, then skip the API call.
		oCBRaw, nCBRaw := d.GetChange("certificate_body")
		oCCRaw, nCCRaw := d.GetChange("certificate_chain")
		oPKRaw, nPKRaw := d.GetChange("private_key")

		if !isChangeNormalizeCertRemoval(oCBRaw, nCBRaw) || !isChangeNormalizeCertRemoval(oCCRaw, nCCRaw) || !isChangeNormalizeCertRemoval(oPKRaw, nPKRaw) {
			input := &workmail.ImportOrganizationInput{
				Organization:    []byte(d.Get("certificate_body").(string)),
				OrganizationArn: aws.String(d.Get("arn").(string)),
				PrivateKey:      []byte(d.Get("private_key").(string)),
			}

			if chain, ok := d.GetOk("certificate_chain"); ok {
				input.OrganizationChain = []byte(chain.(string))
			}

			_, err := conn.ImportOrganization(input)

			if err != nil {
				return fmt.Errorf("error re-importing ACM Organization (%s): %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}
	return resourceOrganizationRead(d, meta)
}

func cleanUpSubjectAlternativeNames(cert *workmail.OrganizationDetail) []string {
	sans := cert.SubjectAlternativeNames
	vs := make([]string, 0)
	for _, v := range sans {
		if aws.StringValue(v) != aws.StringValue(cert.DomainName) {
			vs = append(vs, aws.StringValue(v))
		}
	}
	return vs
}

func convertValidationOptions(certificate *workmail.OrganizationDetail) ([]map[string]interface{}, []string, error) {
	var domainValidationResult []map[string]interface{}
	var emailValidationResult []string

	switch aws.StringValue(certificate.Type) {
	case workmail.OrganizationTypeAmazonIssued:
		if len(certificate.DomainValidationOptions) == 0 && aws.StringValue(certificate.Status) == workmail.DomainStatusPendingValidation {
			log.Printf("[DEBUG] No validation options need to retry.")
			return nil, nil, fmt.Errorf("No validation options need to retry.")
		}
		for _, o := range certificate.DomainValidationOptions {
			if o.ResourceRecord != nil {
				validationOption := map[string]interface{}{
					"domain_name":           aws.StringValue(o.DomainName),
					"resource_record_name":  aws.StringValue(o.ResourceRecord.Name),
					"resource_record_type":  aws.StringValue(o.ResourceRecord.Type),
					"resource_record_value": aws.StringValue(o.ResourceRecord.Value),
				}
				domainValidationResult = append(domainValidationResult, validationOption)
			} else if o.ValidationEmails != nil && len(o.ValidationEmails) > 0 {
				for _, validationEmail := range o.ValidationEmails {
					emailValidationResult = append(emailValidationResult, *validationEmail)
				}
			} else if o.ValidationStatus == nil || aws.StringValue(o.ValidationStatus) == workmail.DomainStatusPendingValidation {
				log.Printf("[DEBUG] Asynchronous ACM service domain validation assignment not complete, need to retry: %#v", o)
				return nil, nil, fmt.Errorf("asynchronous ACM service domain validation assignment not complete, need to retry: %#v", o)
			}
		}
	case workmail.OrganizationTypePrivate:
		// While ACM PRIVATE certificates do not need to be validated, there is a slight delay for
		// the API to fill in all certificate details, which is during the PENDING_VALIDATION status.
		if aws.StringValue(certificate.Status) == workmail.DomainStatusPendingValidation {
			return nil, nil, fmt.Errorf("certificate still pending issuance")
		}
	}

	return domainValidationResult, emailValidationResult, nil
}

func resourceOrganizationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMConn

	log.Printf("[INFO] Deleting ACM Organization: %s", d.Id())

	params := &workmail.DeleteOrganizationInput{
		OrganizationArn: aws.String(d.Id()),
	}

	err := resource.Retry(AcmOrganizationCrossServicePropagationTimeout, func() *resource.RetryError {
		_, err := conn.DeleteOrganization(params)

		if tfawserr.ErrCodeEquals(err, workmail.ErrCodeResourceInUseException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteOrganization(params)
	}

	if tfawserr.ErrCodeEquals(err, workmail.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting ACM Organization (%s): %w", d.Id(), err)
	}

	return nil
}

func workmailDomainValidationOptionsHash(v interface{}) int {
	m, ok := v.(map[string]interface{})

	if !ok {
		return 0
	}

	if v, ok := m["domain_name"].(string); ok {
		return create.StringHashcode(v)
	}

	return 0
}

func expandAcmOrganizationOptions(l []interface{}) *workmail.OrganizationOptions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	options := &workmail.OrganizationOptions{}

	if v, ok := m["certificate_transparency_logging_preference"]; ok {
		options.OrganizationTransparencyLoggingPreference = aws.String(v.(string))
	}

	return options
}

func flattenAcmOrganizationOptions(co *workmail.OrganizationOptions) []interface{} {
	m := map[string]interface{}{
		"certificate_transparency_logging_preference": aws.StringValue(co.OrganizationTransparencyLoggingPreference),
	}

	return []interface{}{m}
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

	newCleanVal := sha1.Sum(stripCR([]byte(strings.TrimSpace(new))))
	return hex.EncodeToString(newCleanVal[:]) == old
}

// strip CRs from raw literals. Lifted from go/scanner/scanner.go
// See https://github.com/golang/go/blob/release-branch.go1.6/src/go/scanner/scanner.go#L479
func stripCR(b []byte) []byte {
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
