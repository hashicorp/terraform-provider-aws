package ec2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_verifiedaccess_trust_provider_attachment", name="Verified Access Trust Provider Attachment")
func ResourceVerifiedAccessTrustProviderAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVerifiedAccessTrustProviderAttachmentCreate,
		ReadWithoutTimeout:   resourceVerifiedAccessTrustProviderAttachmentRead,
		DeleteWithoutTimeout: resourceVerifiedAccessTrustProviderAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceVerifiedAccessTrustProviderAttachmentImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"verified_access_instance_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"verified_access_trust_provider_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

const (
	ResNameVerifiedAccessTrustProviderAttachment = "Verified Access Trust Provider Attachment"
)

func resourceVerifiedAccessTrustProviderAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

	in := &ec2.AttachVerifiedAccessTrustProviderInput{
		VerifiedAccessInstanceId:      aws.String(d.Get("verified_access_instance_id").(string)),
		VerifiedAccessTrustProviderId: aws.String(d.Get("verified_access_trust_provider_id").(string)),
	}

	out, err := conn.AttachVerifiedAccessTrustProviderWithContext(ctx, in)
	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionCreating, ResNameVerifiedAccessTrustProviderAttachment, d.Get("verified_access_trust_provider_id").(string), err)
	}

	if out == nil {
		return create.DiagError(names.EC2, create.ErrActionCreating, ResNameVerifiedAccessTrustProviderAttachment, d.Get("verified_access_trust_provider_id").(string), errors.New("empty output"))
	}

	d.SetId(fmt.Sprintf("%s/%s", d.Get("verified_access_trust_provider_id"), d.Get("verified_access_instance_id")))

	return resourceVerifiedAccessTrustProviderAttachmentRead(ctx, d, meta)
}

func resourceVerifiedAccessTrustProviderAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

	verifiedAccessTrustProviderId := d.Get("verified_access_trust_provider_id").(string)
	verifiedAccessInstanceId := d.Get("verified_access_instance_id").(string)

	_, err := FindVerifiedAccessTrustProviderAttachment(ctx, conn, verifiedAccessTrustProviderId, verifiedAccessInstanceId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VerifiedAccessTrustProviderAttachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionReading, ResNameVerifiedAccessTrustProviderAttachment, d.Id(), err)
	}

	return nil
}

func resourceVerifiedAccessTrustProviderAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

	log.Printf("[INFO] Deleting EC2 VerifiedAccessTrustProviderAttachment %s", d.Id())

	_, err := conn.DetachVerifiedAccessTrustProviderWithContext(ctx, &ec2.DetachVerifiedAccessTrustProviderInput{
		VerifiedAccessInstanceId:      aws.String(d.Get("verified_access_instance_id").(string)),
		VerifiedAccessTrustProviderId: aws.String(d.Get("verified_access_trust_provider_id").(string)),
	})

	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionDeleting, ResNameVerifiedAccessTrustProviderAttachment, d.Id(), err)
	}

	return nil
}

func resourceVerifiedAccessTrustProviderAttachmentImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected <verified-access-trust-provider-id>/<verified-access-instance-id>", d.Id())
	}

	trustProviderId := idParts[0]
	instanceId := idParts[1]

	d.Set("verified_access_trust_provider_id", trustProviderId)
	d.Set("verified_access_instance_id", instanceId)
	d.SetId(fmt.Sprintf("%s/%s", trustProviderId, instanceId))

	return []*schema.ResourceData{d}, nil
}
