package ec2

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_verifiedaccess_group", name="Verified Access Group")
// @Tags(identifierAttribute="id")
func ResourceVerifiedAccessGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVerifiedAccessGroupCreate,
		ReadWithoutTimeout:   resourceVerifiedAccessGroupRead,
		UpdateWithoutTimeout: resourceVerifiedAccessGroupUpdate,
		DeleteWithoutTimeout: resourceVerifiedAccessGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"verified_access_instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},

		CustomizeDiff: customdiff.Sequence(
			verify.SetTagsDiff,
		),
	}
}

const (
	ResNameVerifiedAccessGroup = "Verified Access Group"
)

func resourceVerifiedAccessGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

	in := &ec2.CreateVerifiedAccessGroupInput{
		VerifiedAccessInstanceId: aws.String(d.Get("verified_access_instance_id").(string)),
		TagSpecifications:        getTagSpecificationsIn(ctx, ec2.ResourceTypeVerifiedAccessGroup),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	out, err := conn.CreateVerifiedAccessGroupWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidParameterValue) {
		return create.DiagError(names.EC2, create.ErrActionCreating, ResNameVerifiedAccessGroup, "", err)
	}

	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionCreating, ResNameVerifiedAccessGroup, "", err)
	}

	if out == nil || out.VerifiedAccessGroup == nil {
		return create.DiagError(names.EC2, create.ErrActionCreating, ResNameVerifiedAccessGroup, "", errors.New("empty output"))
	}

	d.SetId(aws.StringValue(out.VerifiedAccessGroup.VerifiedAccessGroupId))

	return resourceVerifiedAccessGroupRead(ctx, d, meta)
}

func resourceVerifiedAccessGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

	out, err := FindVerifiedAccessGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VerifiedAccessGroup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionReading, ResNameVerifiedAccessGroup, d.Id(), err)
	}

	d.Set("arn", out.VerifiedAccessGroupArn)
	d.Set("description", out.Description)
	d.Set("owner", out.Owner)
	d.Set("verified_access_instance_id", out.VerifiedAccessInstanceId)

	SetTagsOut(ctx, out.Tags)

	return nil
}

func resourceVerifiedAccessGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

	update := false

	in := &ec2.ModifyVerifiedAccessGroupInput{
		VerifiedAccessGroupId: aws.String(d.Id()),
	}

	if d.HasChanges("description") {
		in.Description = aws.String(d.Get("an_argument").(string))
		update = true
	}

	if d.HasChanges("verified_access_instance_id") {
		in.VerifiedAccessInstanceId = aws.String(d.Get("verified_access_instance_id").(string))
		update = true
	}

	if !update {
		return nil
	}

	log.Printf("[DEBUG] Updating EC2 VerifiedAccessGroup (%s): %#v", d.Id(), in)

	_, err := conn.ModifyVerifiedAccessGroupWithContext(ctx, in)

	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionUpdating, ResNameVerifiedAccessGroup, d.Id(), err)
	}

	return resourceVerifiedAccessGroupRead(ctx, d, meta)
}

func resourceVerifiedAccessGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

	log.Printf("[INFO] Deleting EC2 VerifiedAccessGroup %s", d.Id())

	_, err := conn.DeleteVerifiedAccessGroupWithContext(ctx, &ec2.DeleteVerifiedAccessGroupInput{
		VerifiedAccessGroupId: aws.String(d.Id()),
	})

	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionDeleting, ResNameVerifiedAccessGroup, d.Id(), err)
	}

	return nil
}
