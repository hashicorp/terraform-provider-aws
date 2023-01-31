package iot

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourcePolicyAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyAttachmentCreate,
		ReadWithoutTimeout:   resourcePolicyAttachmentRead,
		DeleteWithoutTimeout: resourcePolicyAttachmentDelete,
		Schema: map[string]*schema.Schema{
			"policy": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"target": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourcePolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn()

	policyName := d.Get("policy").(string)
	target := d.Get("target").(string)

	_, err := conn.AttachPolicyWithContext(ctx, &iot.AttachPolicyInput{
		PolicyName: aws.String(policyName),
		Target:     aws.String(target),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "attaching policy %s to target %s: %s", policyName, target, err)
	}

	d.SetId(fmt.Sprintf("%s|%s", policyName, target))
	return append(diags, resourcePolicyAttachmentRead(ctx, d, meta)...)
}

func ListPolicyAttachmentPages(ctx context.Context, conn *iot.IoT, input *iot.ListAttachedPoliciesInput,
	fn func(out *iot.ListAttachedPoliciesOutput, lastPage bool) bool) error {
	for {
		page, err := conn.ListAttachedPoliciesWithContext(ctx, input)
		if err != nil {
			return err
		}
		lastPage := page.NextMarker == nil

		shouldContinue := fn(page, lastPage)
		if !shouldContinue || lastPage {
			break
		}
		input.Marker = page.NextMarker
	}
	return nil
}

func GetPolicyAttachment(ctx context.Context, conn *iot.IoT, target, policyName string) (*iot.Policy, error) {
	var policy *iot.Policy

	input := &iot.ListAttachedPoliciesInput{
		PageSize:  aws.Int64(250),
		Recursive: aws.Bool(false),
		Target:    aws.String(target),
	}

	err := ListPolicyAttachmentPages(ctx, conn, input, func(out *iot.ListAttachedPoliciesOutput, lastPage bool) bool {
		for _, att := range out.Policies {
			if policyName == aws.StringValue(att.PolicyName) {
				policy = att
				return false
			}
		}
		return true
	})

	return policy, err
}

func resourcePolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn()

	policyName := d.Get("policy").(string)
	target := d.Get("target").(string)

	var policy *iot.Policy

	policy, err := GetPolicyAttachment(ctx, conn, target, policyName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing policy attachments for target %s: %s", target, err)
	}

	if policy == nil {
		log.Printf("[WARN] IOT Policy Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	return diags
}

func resourcePolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn()

	policyName := d.Get("policy").(string)
	target := d.Get("target").(string)

	_, err := conn.DetachPolicyWithContext(ctx, &iot.DetachPolicyInput{
		PolicyName: aws.String(policyName),
		Target:     aws.String(target),
	})

	// DetachPolicy doesn't return an error if the policy doesn't exist,
	// but it returns an error if the Target is not found.
	if tfawserr.ErrMessageContains(err, iot.ErrCodeInvalidRequestException, "Invalid Target") {
		log.Printf("[WARN] IOT target (%s) not found, removing attachment to policy (%s) from state", target, policyName)
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "detaching policy %s from target %s: %s", policyName, target, err)
	}

	return diags
}
