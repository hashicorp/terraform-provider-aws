// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lightsail_lb_certificate")
func ResourceLoadBalancerCertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLoadBalancerCertificateCreate,
		ReadWithoutTimeout:   resourceLoadBalancerCertificateRead,
		DeleteWithoutTimeout: resourceLoadBalancerCertificateDelete,

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
			"domain_validation_records": {
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
			"lb_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 255),
					validation.StringMatch(regexache.MustCompile(`^[A-Za-z]`), "must begin with an alphabetic character"),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+[^_.-]$`), "must contain only alphanumeric characters, underscores, hyphens, and dots"),
				),
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
			"support_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// Tags are documented in the API, but not supported. API returns:
			// An error occurred (InvalidInputException) when calling the TagResource operation: The resource type, LoadBalancerTlsCertificate, is not taggable.
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
		),
	}
}

func resourceLoadBalancerCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)
	certName := d.Get(names.AttrName).(string)
	in := lightsail.CreateLoadBalancerTlsCertificateInput{
		CertificateDomainName: aws.String(d.Get(names.AttrDomainName).(string)),
		CertificateName:       aws.String(certName),
		LoadBalancerName:      aws.String(d.Get("lb_name").(string)),
	}

	if v, ok := d.GetOk("subject_alternative_names"); ok {
		in.CertificateAlternativeNames = expandSubjectAlternativeNames(v)
	}

	out, err := conn.CreateLoadBalancerTlsCertificate(ctx, &in)

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, string(types.OperationTypeCreateLoadBalancerTlsCertificate), ResLoadBalancerCertificate, certName, err)
	}

	diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeCreateLoadBalancerTlsCertificate, ResLoadBalancerCertificate, certName)

	if diag != nil {
		return diag
	}

	// Generate an ID
	vars := []string{
		d.Get("lb_name").(string),
		certName,
	}

	d.SetId(strings.Join(vars, ","))

	return append(diags, resourceLoadBalancerCertificateRead(ctx, d, meta)...)
}

func resourceLoadBalancerCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	out, err := FindLoadBalancerCertificateById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResLoadBalancerCertificate, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, create.ErrActionReading, ResLoadBalancerCertificate, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set(names.AttrCreatedAt, out.CreatedAt.Format(time.RFC3339))
	d.Set(names.AttrDomainName, out.DomainName)
	d.Set("domain_validation_records", flattenLoadBalancerDomainValidationRecords(out.DomainValidationRecords))
	d.Set("lb_name", out.LoadBalancerName)
	d.Set(names.AttrName, out.Name)
	d.Set("subject_alternative_names", out.SubjectAlternativeNames)
	d.Set("support_code", out.SupportCode)

	return diags
}

func resourceLoadBalancerCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	id_parts := strings.SplitN(d.Id(), ",", -1)
	lbName := id_parts[0]
	certName := id_parts[1]

	out, err := conn.DeleteLoadBalancerTlsCertificate(ctx, &lightsail.DeleteLoadBalancerTlsCertificateInput{
		CertificateName:  aws.String(certName),
		LoadBalancerName: aws.String(lbName),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.Lightsail, string(types.OperationTypeDeleteLoadBalancerTlsCertificate), ResLoadBalancerCertificate, certName, err)
	}

	diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeDeleteLoadBalancerTlsCertificate, ResLoadBalancerCertificate, certName)

	if diag != nil {
		return diag
	}

	return diags
}

func flattenLoadBalancerDomainValidationRecords(domainValidationRecords []types.LoadBalancerTlsCertificateDomainValidationRecord) []map[string]interface{} {
	var domainValidationResult []map[string]interface{}

	for _, o := range domainValidationRecords {
		validationOption := map[string]interface{}{
			names.AttrDomainName:    aws.ToString(o.DomainName),
			"resource_record_name":  aws.ToString(o.Name),
			"resource_record_type":  aws.ToString(o.Type),
			"resource_record_value": aws.ToString(o.Value),
		}
		domainValidationResult = append(domainValidationResult, validationOption)
	}

	return domainValidationResult
}

func FindLoadBalancerCertificateById(ctx context.Context, conn *lightsail.Client, id string) (*types.LoadBalancerTlsCertificate, error) {
	id_parts := strings.SplitN(id, ",", -1)
	if len(id_parts) != 2 {
		return nil, errors.New("invalid load balancer certificate id")
	}

	lbName := id_parts[0]
	cName := id_parts[1]

	in := &lightsail.GetLoadBalancerTlsCertificatesInput{LoadBalancerName: aws.String(lbName)}
	out, err := conn.GetLoadBalancerTlsCertificates(ctx, in)

	if IsANotFoundError(err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	var entry types.LoadBalancerTlsCertificate
	entryExists := false

	for _, n := range out.TlsCertificates {
		if cName == aws.ToString(n.Name) {
			entry = n
			entryExists = true
			break
		}
	}

	if !entryExists {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return &entry, nil
}
