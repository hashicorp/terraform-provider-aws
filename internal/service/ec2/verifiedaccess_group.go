package ec2

import (
	"context"
	"errors"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_verifiedaccess_group")
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
			"policy_document": {
				Type:                  schema.TypeString,
				Optional:              true,
				DiffSuppressFunc:      SuppressEquivalentGroupPolicyDiffs,
				DiffSuppressOnRefresh: true,
				ValidateFunc:          validation.StringIsNotEmpty,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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
	ResNameVerifiedAccessGroup  = "Verified Access Group"
	ResNameVerifiedAccessPolicy = "Verified Access Policy"
)

func resourceVerifiedAccessGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

	in := &ec2.CreateVerifiedAccessGroupInput{
		VerifiedAccessInstanceId: aws.String(d.Get("verified_access_instance_id").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("policy_document"); ok {
		in.PolicyDocument = aws.String(v.(string))
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.TagSpecifications = tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeVerifiedAccessGroup)
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
	var diags diag.Diagnostics
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

	verifiedAccessGroupId := d.Id()

	in := &ec2.GetVerifiedAccessGroupPolicyInput{
		VerifiedAccessGroupId: &verifiedAccessGroupId,
	}

	policy_output, err := conn.GetVerifiedAccessGroupPolicyWithContext(ctx, in)

	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionReading, ResNameVerifiedAccessPolicy, d.Id(), err)
	}

	d.Set("policy_document", policy_output.PolicyDocument)

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	tags := KeyValueTags(ctx, out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return nil
}

func resourceVerifiedAccessGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	if d.HasChangesExcept("policy_document", "tags", "tags_all") {
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

		if err != nil {
			return create.DiagError(names.EC2, create.ErrActionReading, ResNameVerifiedAccessPolicy, d.Id(), err)
		}
	}

	if d.HasChanges("policy_document") {
		verifiedAccessGroupId := d.Id()

		in := &ec2.ModifyVerifiedAccessGroupPolicyInput{
			VerifiedAccessGroupId: &verifiedAccessGroupId,
		}

		if v, ok := d.GetOk("policy_document"); ok {
			in.PolicyDocument = aws.String(v.(string))
			in.PolicyEnabled = aws.Bool(true)
		} else {
			in.PolicyEnabled = aws.Bool(false)
		}

		_, err := conn.ModifyVerifiedAccessGroupPolicyWithContext(ctx, in)

		if err != nil {
			return create.DiagError(names.EC2, create.ErrActionUpdating, ResNameVerifiedAccessGroup, d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating verified access group (%s) tags: %s", d.Id(), err)
		}
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

func SuppressEquivalentGroupPolicyDiffs(k, old, new string, d *schema.ResourceData) bool {
	// ignore leading and trailing whitespace for policies
	old_policy := strings.TrimSpace(old)
	new_policy := strings.TrimSpace(new)

	return reflect.DeepEqual(old_policy, new_policy)
}
