package ec2

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @SDKResource("aws_verifiedaccess_access_group", name="Verified Access Access Group")
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
			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"last_updated_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_document": {
				Type:      schema.TypeString,
				Optional:  true,
			},
			"verified_access_group_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"verified_access_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"verified_access_instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameVerifiedAccessGroup = "Verified Access Group"
)

func resourceVerifiedAccessGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	in := &ec2.CreateVerifiedAccessGroupInput{
		VerifiedAccessInstanceId: aws.String(d.Get("verified_access_instance_id").(string)),
		TagSpecifications:        getTagSpecificationsInV2(ctx, types.ResourceTypeVerifiedAccessGroup),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}


	if v, ok := d.GetOk("policy_document"); ok {
		in.PolicyDocument = aws.String(v.(string))
	}

	out, err := conn.CreateVerifiedAccessGroup(ctx, in)
	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionCreating, ResNameVerifiedAccessGroup, "", err)
	}

	if out == nil || out.VerifiedAccessGroup == nil {
		return create.DiagError(names.EC2, create.ErrActionCreating, ResNameVerifiedAccessGroup, "", errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.VerifiedAccessGroup.VerifiedAccessGroupId))

	return resourceVerifiedAccessGroupRead(ctx, d, meta)
}

func resourceVerifiedAccessGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	out, err := findVerifiedAccessGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VerifiedAccessGroup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionReading, ResNameVerifiedAccessGroup, d.Id(), err)...)
	}

	// Set simple properties
	d.Set("creation_time", out.VerifiedAccessGroups[0].CreationTime)
	d.Set("deletion_time", out.VerifiedAccessGroups[0].DeletionTime)
	d.Set("description", out.VerifiedAccessGroups[0].Description)
	d.Set("last_updated_time", out.VerifiedAccessGroups[0].LastUpdatedTime)
	d.Set("owner", out.VerifiedAccessGroups[0].Owner)
	d.Set("verified_access_group_arn", out.VerifiedAccessGroups[0].VerifiedAccessGroupArn)
	d.Set("verified_access_group_id", out.VerifiedAccessGroups[0].VerifiedAccessGroupId)
	d.Set("verified_access_instance_id", out.VerifiedAccessGroups[0].VerifiedAccessInstanceId)

	// Set tags
	setTagsOutV2(ctx, out.VerifiedAccessGroups[0].Tags)

	// Retrieve policy
	output, err := findVerifiedAccessGroupPolicyByGroupID(ctx, conn, d.Id())

	if err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionReading, ResNameVerifiedAccessGroup, d.Id(), err)...)
	}
	d.Set("policy_document", output.PolicyDocument)

	return diags
}

func resourceVerifiedAccessGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	update := false

	in := &ec2.ModifyVerifiedAccessGroupInput{
		VerifiedAccessGroupId: aws.String(d.Id()),
	}

	if d.HasChange("description") {
		if v, ok := d.GetOk("description"); ok {
			in.Description = aws.String(v.(string))
			update = true
		}
	}

	if d.HasChange("verified_access_instance_id") {
		if v, ok := d.GetOk("verified_access_instance_id"); ok {
			in.VerifiedAccessInstanceId = aws.String(v.(string))
			update = true
		}
	}

	if !update {
		return diags
	}

	log.Printf("[DEBUG] Updating EC2 VerifiedAccessGroup (%s): %#v", d.Id(), in)
	_, err := conn.ModifyVerifiedAccessGroup(ctx, in)

	if err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionUpdating, ResNameVerifiedAccessGroup, d.Id(), err)...)
	}

	return append(diags, resourceVerifiedAccessGroupRead(ctx, d, meta)...)
}

func resourceVerifiedAccessGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting EC2 VerifiedAccessGroup %s", d.Id())
	_, err := conn.DeleteVerifiedAccessGroup(ctx, &ec2.DeleteVerifiedAccessGroupInput{
		VerifiedAccessGroupId: aws.String(d.Id()),
	})

	if err != nil {
		return append(diags, create.DiagError(names.EC2, create.ErrActionDeleting, ResNameVerifiedAccessGroup, d.Id(), err)...)
	}

	return diags
}

// TIP: ==== STATUS CONSTANTS ====
// Create constants for states and statuses if the service does not
// already have suitable constants. We prefer that you use the constants
// provided in the service if available (e.g., amp.WorkspaceStatusCodeActive).
const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)


func findVerifiedAccessGroupPolicyByGroupID(ctx context.Context, conn *ec2.Client, id string) (*ec2.GetVerifiedAccessGroupPolicyOutput, error) {
	in := &ec2.GetVerifiedAccessGroupPolicyInput{
		VerifiedAccessGroupId: &id,
	}
	out, err := conn.GetVerifiedAccessGroupPolicy(ctx, in)

	if err != nil {
		return nil, err
	}

	return out, nil
}


func findVerifiedAccessGroupByID(ctx context.Context, conn *ec2.Client, id string) (*ec2.DescribeVerifiedAccessGroupsOutput, error) {
	in := &ec2.DescribeVerifiedAccessGroupsInput{
		VerifiedAccessGroupIds: []string{id},
	}
	out, err := conn.DescribeVerifiedAccessGroups(ctx, in)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVerifiedAccessGroupIdNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.VerifiedAccessGroups == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
