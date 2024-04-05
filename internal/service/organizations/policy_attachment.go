// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_organizations_policy_attachment")
func ResourcePolicyAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyAttachmentCreate,
		ReadWithoutTimeout:   resourcePolicyAttachmentRead,
		UpdateWithoutTimeout: resourcePolicyAttachmentUpdate,
		DeleteWithoutTimeout: resourcePolicyAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"policy_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"skip_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"target_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourcePolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	policyID := d.Get("policy_id").(string)
	targetID := d.Get("target_id").(string)
	id := fmt.Sprintf("%s:%s", targetID, policyID)
	input := &organizations.AttachPolicyInput{
		PolicyId: aws.String(policyID),
		TargetId: aws.String(targetID),
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 4*time.Minute, func() (interface{}, error) {
		return conn.AttachPolicyWithContext(ctx, input)
	}, organizations.ErrCodeFinalizingOrganizationException)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Organizations Policy Attachment (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourcePolicyAttachmentRead(ctx, d, meta)...)
}

func resourcePolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	targetID, policyID, err := DecodePolicyAttachmentID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organizations Policy Attachment (%s): %s", d.Id(), err)
	}

	_, err = FindPolicyAttachmentByTwoPartKey(ctx, conn, targetID, policyID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Organizations Policy Attachment %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Organizations Policy Attachment (%s): %s", d.Id(), err)
	}

	d.Set("policy_id", policyID)
	d.Set("target_id", targetID)

	return diags
}

func resourcePolicyAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Update is just a pass-through to allow skip_destroy to be updated in-place
	var diags diag.Diagnostics
	return append(diags, resourcePolicyAttachmentRead(ctx, d, meta)...)
}

func resourcePolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if v, ok := d.GetOk("skip_destroy"); ok && v.(bool) {
		log.Printf("[DEBUG] Retaining Organizations Policy Attachment: %s", d.Id())
		return nil
	}

	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OrganizationsConn(ctx)

	targetID, policyID, err := DecodePolicyAttachmentID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Organizations Policy Attachment (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting Organizations Policy Attachment: %s", d.Id())
	_, err = conn.DetachPolicyWithContext(ctx, &organizations.DetachPolicyInput{
		PolicyId: aws.String(policyID),
		TargetId: aws.String(targetID),
	})

	if tfawserr.ErrCodeEquals(err, organizations.ErrCodeTargetNotFoundException, organizations.ErrCodePolicyNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Organizations Policy Attachment (%s): %s", d.Id(), err)
	}

	return diags
}

func DecodePolicyAttachmentID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("expected ID in format of TARGETID:POLICYID, received: %s", id)
	}
	return idParts[0], idParts[1], nil
}
