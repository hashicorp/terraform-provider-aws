package lightsail

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceLoadBalancerCertificate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLoadBalancerCertificateCreate,
		ReadWithoutTimeout:   resourceLoadBalancerCertificateRead,
		DeleteWithoutTimeout: resourceLoadBalancerCertificateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_name": {
				// AWS Provider 3.0.0 aws_route53_zone references no longer contain a
				// trailing period, no longer requiring a custom StateFunc
				// to prevent ACM API error
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringDoesNotMatch(regexp.MustCompile(`\.$`), "cannot end with a period"),
			},
			"domain_validation_records": {
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
				Set: domainValidationOptionsHash,
			},
			"lb_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 255),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z]`), "must begin with an alphabetic character"),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_\-.]+[^._\-]$`), "must contain only alphanumeric characters, underscores, hyphens, and dots"),
				),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 255),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z]`), "must begin with an alphabetic character"),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_\-.]+[^._\-]$`), "must contain only alphanumeric characters, underscores, hyphens, and dots"),
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
						validation.StringDoesNotMatch(regexp.MustCompile(`\.$`), "cannot end with a period"),
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
		),
	}
}

func resourceLoadBalancerCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	in := lightsail.CreateLoadBalancerTlsCertificateInput{
		CertificateDomainName: aws.String(d.Get("domain_name").(string)),
		CertificateName:       aws.String(d.Get("name").(string)),
		LoadBalancerName:      aws.String(d.Get("lb_name").(string)),
	}

	if v, ok := d.GetOk("subject_alternative_names"); ok {
		in.CertificateAlternativeNames = aws.StringSlice(expandSubjectAlternativeNames(v))
	}

	out, err := conn.CreateLoadBalancerTlsCertificateWithContext(ctx, &in)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateLoadBalancerTlsCertificate, ResLoadBalancerCertificate, d.Get("name").(string), err)
	}

	if len(out.Operations) == 0 {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateLoadBalancerTlsCertificate, ResLoadBalancerCertificate, d.Get("name").(string), errors.New("No operations found for Create Load Balancer Certificate request"))
	}

	op := out.Operations[0]

	err = waitOperation(ctx, conn, op.Id)
	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateLoadBalancerTlsCertificate, ResLoadBalancerCertificate, d.Get("name").(string), errors.New("Error waiting for Create Load Balancer request operation"))
	}

	// Generate an ID
	vars := []string{
		d.Get("lb_name").(string),
		d.Get("name").(string),
	}

	d.SetId(strings.Join(vars, ","))

	return resourceLoadBalancerCertificateRead(ctx, d, meta)
}

func resourceLoadBalancerCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	out, err := FindLoadBalancerCertificateById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResLoadBalancerCertificate, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResLoadBalancerCertificate, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("created_at", out.CreatedAt.Format(time.RFC3339))
	d.Set("domain_name", out.DomainName)
	d.Set("domain_validation_records", flattenLoadBalancerDomainValidationRecords(out.DomainValidationRecords))
	d.Set("lb_name", out.LoadBalancerName)
	d.Set("name", out.Name)
	d.Set("subject_alternative_names", aws.StringValueSlice(out.SubjectAlternativeNames))
	d.Set("support_code", out.SupportCode)

	return nil
}

func resourceLoadBalancerCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	id_parts := strings.SplitN(d.Id(), ",", -1)
	lbName := id_parts[0]
	cName := id_parts[1]

	out, err := conn.DeleteLoadBalancerTlsCertificateWithContext(ctx, &lightsail.DeleteLoadBalancerTlsCertificateInput{
		CertificateName:  aws.String(cName),
		LoadBalancerName: aws.String(lbName),
	})

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeDeleteLoadBalancerTlsCertificate, ResLoadBalancerCertificate, d.Get("name").(string), err)
	}

	if len(out.Operations) == 0 {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeDeleteLoadBalancerTlsCertificate, ResLoadBalancerCertificate, d.Get("name").(string), errors.New("No operations found for Delete Load Balancer Certificate request"))
	}

	op := out.Operations[0]

	err = waitOperation(ctx, conn, op.Id)
	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateLoadBalancerTlsCertificate, ResLoadBalancerCertificate, d.Get("name").(string), errors.New("Error waiting for Delete Load Balancer Certificate request operation"))
	}

	return nil
}

func flattenLoadBalancerDomainValidationRecords(domainValidationRecords []*lightsail.LoadBalancerTlsCertificateDomainValidationRecord) []map[string]interface{} {
	var domainValidationResult []map[string]interface{}

	for _, o := range domainValidationRecords {
		validationOption := map[string]interface{}{
			"domain_name":           aws.StringValue(o.DomainName),
			"resource_record_name":  aws.StringValue(o.Name),
			"resource_record_type":  aws.StringValue(o.Type),
			"resource_record_value": aws.StringValue(o.Value),
		}
		domainValidationResult = append(domainValidationResult, validationOption)
	}

	return domainValidationResult
}
