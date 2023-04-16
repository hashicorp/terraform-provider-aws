package ec2

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_verifiedaccess_instance", name="Verified Access Instance")
// @Tags(identifierAttribute="id")
func ResourceVerifiedAccessInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVerifiedAccessInstanceCreate,
		ReadWithoutTimeout:   resourceVerifiedAccessInstanceRead,
		UpdateWithoutTimeout: resourceVerifiedAccessInstanceUpdate,
		DeleteWithoutTimeout: resourceVerifiedAccessInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameVerifiedAccessInstance = "Verified Access Instance"
)

func resourceVerifiedAccessInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

	in := &ec2.CreateVerifiedAccessInstanceInput{
		TagSpecifications: getTagSpecificationsIn(ctx, ec2.ResourceTypeVerifiedAccessInstance),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	out, err := conn.CreateVerifiedAccessInstanceWithContext(ctx, in)
	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionCreating, ResNameVerifiedAccessInstance, "", err)
	}

	if out == nil || out.VerifiedAccessInstance == nil {
		return create.DiagError(names.EC2, create.ErrActionCreating, ResNameVerifiedAccessInstance, "", errors.New("empty output"))
	}

	d.SetId(aws.StringValue(out.VerifiedAccessInstance.VerifiedAccessInstanceId))

	return resourceVerifiedAccessInstanceRead(ctx, d, meta)
}

func resourceVerifiedAccessInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	conn := meta.(*conns.AWSClient).EC2Conn()

	out, err := FindVerifiedAccessInstanceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VerifiedAccessInstance (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionReading, ResNameVerifiedAccessInstance, d.Id(), err)
	}

	d.Set("description", out.Description)

	SetTagsOut(ctx, out.Tags)

	return nil
}

func resourceVerifiedAccessInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

	if d.HasChangesExcept("tags", "tags_all") {
		update := false

		in := &ec2.ModifyVerifiedAccessInstanceInput{
			VerifiedAccessInstanceId: aws.String(d.Id()),
		}

		if d.HasChanges("description") {
			in.Description = aws.String(d.Get("description").(string))
			update = true
		}

		if !update {
			return nil
		}

		log.Printf("[DEBUG] Updating EC2 VerifiedAccessInstance (%s): %#v", d.Id(), in)
		_, err := conn.ModifyVerifiedAccessInstanceWithContext(ctx, in)
		if err != nil {
			return create.DiagError(names.EC2, create.ErrActionUpdating, ResNameVerifiedAccessInstance, d.Id(), err)
		}
	}

	return resourceVerifiedAccessInstanceRead(ctx, d, meta)
}

func resourceVerifiedAccessInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

	log.Printf("[INFO] Deleting EC2 VerifiedAccessInstance %s", d.Id())

	_, err := conn.DeleteVerifiedAccessInstanceWithContext(ctx, &ec2.DeleteVerifiedAccessInstanceInput{
		VerifiedAccessInstanceId: aws.String(d.Id()),
	})

	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionDeleting, ResNameVerifiedAccessInstance, d.Id(), err)
	}

	return nil
}
