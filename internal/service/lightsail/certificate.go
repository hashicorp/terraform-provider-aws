package lightsail

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

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
				Set: domainValidationOptionsHash,
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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
			verify.SetTagsDiff,
		),
	}
}

func resourceCertificateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	req := lightsail.CreateCertificateInput{
		CertificateName: aws.String(d.Get("name").(string)),
		DomainName:      aws.String(d.Get("domain_name").(string)),
	}

	if v, ok := d.GetOk("subject_alternative_names"); ok {
		req.SubjectAlternativeNames = aws.StringSlice(expandSubjectAlternativeNames(v))
	}

	if len(tags) > 0 {
		req.Tags = Tags(tags.IgnoreAWS())
	}

	resp, err := conn.CreateCertificateWithContext(ctx, &req)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateCertificate, ResCertificate, d.Get("name").(string), err)
	}

	if len(resp.Operations) == 0 {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateCertificate, ResCertificate, d.Get("name").(string), errors.New("No operations found for CreateCertificate request"))
	}

	op := resp.Operations[0]

	err = waitOperation(ctx, conn, op.Id)
	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeCreateCertificate, ResCertificate, d.Get("name").(string), errors.New("Error waiting for Create Certificate request operation"))
	}

	d.SetId(d.Get("name").(string))

	return resourceCertificateRead(ctx, d, meta)
}

func resourceCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	certificate, err := FindCertificateByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.CE, create.ErrActionReading, ResCertificate, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.CE, create.ErrActionReading, ResCertificate, d.Id(), err)
	}

	d.Set("arn", certificate.Arn)
	d.Set("created_at", certificate.CreatedAt.Format(time.RFC3339))
	d.Set("domain_name", certificate.DomainName)
	d.Set("domain_validation_options", flattenDomainValidationRecords(certificate.DomainValidationRecords))
	d.Set("name", certificate.Name)
	d.Set("subject_alternative_names", aws.StringValueSlice(certificate.SubjectAlternativeNames))

	tags := KeyValueTags(certificate.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResCertificate, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResCertificate, d.Id(), err)
	}

	return nil
}

func resourceCertificateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return create.DiagError(names.Lightsail, create.ErrActionUpdating, ResCertificate, d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return create.DiagError(names.Lightsail, create.ErrActionUpdating, ResCertificate, d.Id(), err)
		}
	}

	return resourceCertificateRead(ctx, d, meta)
}

func resourceCertificateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	resp, err := conn.DeleteCertificateWithContext(ctx, &lightsail.DeleteCertificateInput{
		CertificateName: aws.String(d.Id()),
	})

	if err != nil && tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.CE, create.ErrActionDeleting, ResCertificate, d.Id(), err)
	}

	op := resp.Operations[0]

	err = waitOperation(ctx, conn, op.Id)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeDeleteCertificate, ResCertificate, d.Id(), err)
	}

	return nil
}

func domainValidationOptionsHash(v interface{}) int {
	m, ok := v.(map[string]interface{})

	if !ok {
		return 0
	}

	if v, ok := m["domain_name"].(string); ok {
		return create.StringHashcode(v)
	}

	return 0
}

func flattenDomainValidationRecords(domainValidationRecords []*lightsail.DomainValidationRecord) []map[string]interface{} {
	var domainValidationResult []map[string]interface{}

	for _, o := range domainValidationRecords {
		if o.ResourceRecord != nil {
			validationOption := map[string]interface{}{
				"domain_name":           aws.StringValue(o.DomainName),
				"resource_record_name":  aws.StringValue(o.ResourceRecord.Name),
				"resource_record_type":  aws.StringValue(o.ResourceRecord.Type),
				"resource_record_value": aws.StringValue(o.ResourceRecord.Value),
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
