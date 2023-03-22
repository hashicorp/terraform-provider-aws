package lightsail

import (
	"context"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_lightsail_lb_certificate_attachment")
func ResourceLoadBalancerCertificateAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLoadBalancerCertificateAttachmentCreate,
		ReadWithoutTimeout:   resourceLoadBalancerCertificateAttachmentRead,
		DeleteWithoutTimeout: resourceLoadBalancerCertificateAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
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
			"certificate_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceLoadBalancerCertificateAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()
	certName := d.Get("certificate_name").(string)
	req := lightsail.AttachLoadBalancerTlsCertificateInput{
		LoadBalancerName: aws.String(d.Get("lb_name").(string)),
		CertificateName:  aws.String(certName),
	}

	out, err := conn.AttachLoadBalancerTlsCertificateWithContext(ctx, &req)

	if err != nil {
		return create.DiagError(names.Lightsail, lightsail.OperationTypeAttachLoadBalancerTlsCertificate, ResLoadBalancerCertificateAttachment, certName, err)
	}

	diag := expandOperations(ctx, conn, out.Operations, lightsail.OperationTypeAttachLoadBalancerTlsCertificate, ResLoadBalancerCertificateAttachment, certName)

	if diag != nil {
		return diag
	}

	// Generate an ID
	vars := []string{
		d.Get("lb_name").(string),
		certName,
	}

	d.SetId(strings.Join(vars, ","))

	return resourceLoadBalancerCertificateAttachmentRead(ctx, d, meta)
}

func resourceLoadBalancerCertificateAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	out, err := FindLoadBalancerCertificateAttachmentById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResLoadBalancerCertificateAttachment, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResLoadBalancerCertificateAttachment, d.Id(), err)
	}

	d.Set("certificate_name", out)
	d.Set("lb_name", expandLoadBalancerNameFromId(d.Id()))

	return nil
}

func resourceLoadBalancerCertificateAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[WARN] Cannot destroy Lightsail Load Balancer Certificate Attachment. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}
