// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
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
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"verifiedaccess_group_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"verifiedaccess_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"verifiedaccess_instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVerifiedAccessGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateVerifiedAccessGroupInput{
		TagSpecifications:        getTagSpecificationsInV2(ctx, types.ResourceTypeVerifiedAccessGroup),
		VerifiedAccessInstanceId: aws.String(d.Get("verifiedaccess_instance_id").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("policy_document"); ok {
		input.PolicyDocument = aws.String(v.(string))
	}

	output, err := conn.CreateVerifiedAccessGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Verified Access Group: %s", err)
	}

	d.SetId(aws.ToString(output.VerifiedAccessGroup.VerifiedAccessGroupId))

	return append(diags, resourceVerifiedAccessGroupRead(ctx, d, meta)...)
}

func resourceVerifiedAccessGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	group, err := FindVerifiedAccessGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Verified Access Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Verified Access Group (%s): %s", d.Id(), err)
	}

	d.Set("creation_time", group.CreationTime)
	d.Set("deletion_time", group.DeletionTime)
	d.Set("description", group.Description)
	d.Set("last_updated_time", group.LastUpdatedTime)
	d.Set("owner", group.Owner)
	d.Set("verifiedaccess_group_arn", group.VerifiedAccessGroupArn)
	d.Set("verifiedaccess_group_id", group.VerifiedAccessGroupId)
	d.Set("verifiedaccess_instance_id", group.VerifiedAccessInstanceId)

	setTagsOutV2(ctx, group.Tags)

	output, err := FindVerifiedAccessGroupPolicyByID(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Verified Access Group (%s) policy: %s", d.Id(), err)
	}

	d.Set("policy_document", output.PolicyDocument)

	return diags
}

func resourceVerifiedAccessGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChangesExcept("policy_document", "tags", "tags_all") {
		input := &ec2.ModifyVerifiedAccessGroupInput{
			VerifiedAccessGroupId: aws.String(d.Id()),
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("verified_access_instance_id") {
			input.VerifiedAccessInstanceId = aws.String(d.Get("description").(string))
		}

		_, err := conn.ModifyVerifiedAccessGroup(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Verified Access Group (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("policy_document") {
		input := &ec2.ModifyVerifiedAccessGroupPolicyInput{
			PolicyDocument:        aws.String(d.Get("policy_document").(string)),
			VerifiedAccessGroupId: aws.String(d.Id()),
			PolicyEnabled:         aws.Bool(true),
		}

		_, err := conn.ModifyVerifiedAccessGroupPolicy(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Verified Access Group (%s) policy: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVerifiedAccessGroupRead(ctx, d, meta)...)
}

func resourceVerifiedAccessGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[INFO] Deleting Verified Access Group: %s", d.Id())
	_, err := conn.DeleteVerifiedAccessGroup(ctx, &ec2.DeleteVerifiedAccessGroupInput{
		VerifiedAccessGroupId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVerifiedAccessGroupIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Verified Access Group (%s): %s", d.Id(), err)
	}

	return diags
}
