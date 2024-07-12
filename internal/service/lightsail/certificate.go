// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lightsail_certificate", name="Certificate")
// @Tags(identifierAttribute="id", resourceType="Certificate")
func ResourceCertificate() *schema.Resource {
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
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomainName: {
				// AWS Provider 3.0.0 aws_route53_zone references no longer contain a
				// trailing period, no longer requiring a custom StateFunc
				// to prevent ACM API error
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringDoesNotMatch(regexache.MustCompile(`\.$`), "cannot end with a period"),
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 255),
					validation.StringMatch(regexache.MustCompile(`^[A-Za-z]`), "must begin with an alphabetic character"),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+[^_.-]$`), "must contain only alphanumeric characters, underscores, hyphens, and dots"),
				),
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
				Set: schema.HashString,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: customdiff.Sequence(
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				// Lightsail automatically adds the domain_name value to the list of SANs. Mimic Lightsail's behavior
				// so that the user doesn't need to explicitly set it themselves.
				if diff.HasChange(names.AttrDomainName) || diff.HasChange("subject_alternative_names") {
					domain_name := diff.Get(names.AttrDomainName).(string)

					if sanSet, ok := diff.Get("subject_alternative_names").(*schema.Set); ok {
						sanSet.Add(domain_name)
						if err := diff.SetNew("subject_alternative_names", sanSet); err != nil {
							return fmt.Errorf("setting new subject_alternative_names diff: %w", err)
						}
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

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	req := lightsail.CreateCertificateInput{
		CertificateName: aws.String(d.Get(names.AttrName).(string)),
		DomainName:      aws.String(d.Get(names.AttrDomainName).(string)),
		Tags:            getTagsIn(ctx),
	}

	if v, ok := d.GetOk("subject_alternative_names"); ok {
		req.SubjectAlternativeNames = expandSubjectAlternativeNames(v)
	}

	resp, err := conn.CreateCertificate(ctx, &req)

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, string(types.OperationTypeCreateCertificate), ResCertificate, d.Get(names.AttrName).(string), err)
	}

	id := d.Get(names.AttrName).(string)
	diag := expandOperations(ctx, conn, resp.Operations, types.OperationTypeCreateCertificate, ResCertificate, id)

	if diag != nil {
		return diag
	}

	d.SetId(id)

	return append(diags, resourceCertificateRead(ctx, d, meta)...)
}

func resourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	certificate, err := FindCertificateById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResCertificate, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, create.ErrActionReading, ResCertificate, d.Id(), err)
	}

	d.Set(names.AttrARN, certificate.Arn)
	d.Set(names.AttrCreatedAt, certificate.CreatedAt.Format(time.RFC3339))
	d.Set(names.AttrDomainName, certificate.DomainName)
	d.Set("domain_validation_options", flattenDomainValidationRecords(certificate.DomainValidationRecords))
	d.Set(names.AttrName, certificate.Name)
	d.Set("subject_alternative_names", certificate.SubjectAlternativeNames)

	setTagsOut(ctx, certificate.Tags)

	return diags
}

func resourceCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceCertificateRead(ctx, d, meta)
}

func resourceCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	resp, err := conn.DeleteCertificate(ctx, &lightsail.DeleteCertificateInput{
		CertificateName: aws.String(d.Id()),
	})

	if err != nil && errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, create.ErrActionDeleting, ResCertificate, d.Id(), err)
	}

	diag := expandOperations(ctx, conn, resp.Operations, types.OperationTypeDeleteCertificate, ResCertificate, d.Id())

	if diag != nil {
		return diag
	}

	return diags
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

func flattenDomainValidationRecords(domainValidationRecords []types.DomainValidationRecord) []map[string]interface{} {
	var domainValidationResult []map[string]interface{}

	for _, o := range domainValidationRecords {
		if o.ResourceRecord != nil {
			validationOption := map[string]interface{}{
				names.AttrDomainName:    aws.ToString(o.DomainName),
				"resource_record_name":  aws.ToString(o.ResourceRecord.Name),
				"resource_record_type":  aws.ToString(o.ResourceRecord.Type),
				"resource_record_value": aws.ToString(o.ResourceRecord.Value),
			}
			domainValidationResult = append(domainValidationResult, validationOption)
		}
	}

	return domainValidationResult
}

func expandSubjectAlternativeNames(sans interface{}) []string {
	subjectAlternativeNames := make([]string, len(sans.(*schema.Set).List()))
	for i, sanRaw := range sans.(*schema.Set).List() {
		subjectAlternativeNames[i] = sanRaw.(string)
	}

	return subjectAlternativeNames
}

func FindCertificateById(ctx context.Context, conn *lightsail.Client, name string) (*types.Certificate, error) {
	in := &lightsail.GetCertificatesInput{
		CertificateName: aws.String(name),
	}

	out, err := conn.GetCertificates(ctx, in)

	if IsANotFoundError(err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.Certificates) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Certificates[0].CertificateDetail, nil
}
