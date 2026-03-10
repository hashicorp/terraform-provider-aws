// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func ResourcePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourcePolicyCreate,
		Read:   resourcePolicyRead,
		Update: resourcePolicyUpdate,
		Delete: resourcePolicyDelete,
		Schema: map[string]*schema.Schema{
			"resource_arn": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ARN of the Bedrock Agent Core resource to attach the policy to.",
			},
			"policy": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The JSON policy document.",
			},
			"policy_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Identifier for the resource policy returned by the service.",
			},
		},
	}
}

func resourcePolicyCreate(d *schema.ResourceData, meta any) error {
	conn := meta.(*conns.AWSClient).BedrockAgentCoreClient(context.Background())

	arn := d.Get("resource_arn").(string)
	policyRaw := d.Get("policy").(string)

	policy, err := structure.NormalizeJsonString(policyRaw)
	if err != nil {
		return fmt.Errorf("invalid policy JSON: %s", err)
	}

	input := bedrockagentcorecontrol.PutResourcePolicyInput{
		ResourceArn: aws.String(arn),
		Policy:      aws.String(policy),
	}

	_, err = conn.PutResourcePolicy(context.Background(), &input)
	if err != nil {
		return fmt.Errorf("creating Bedrock Agent Core Resource Policy (%s): %s", arn, err)
	}

	// Use the resource ARN as the Terraform ID.
	d.SetId(arn)

	return resourcePolicyRead(d, meta)
}

func resourcePolicyRead(d *schema.ResourceData, meta any) error {
	conn := meta.(*conns.AWSClient).BedrockAgentCoreClient(context.Background())

	id := d.Id()
	if id == "" {
		// fallback to resource_arn attribute
		if v, ok := d.GetOk("resource_arn"); ok {
			id = v.(string)
		}
	}

	input := bedrockagentcorecontrol.GetResourcePolicyInput{
		ResourceArn: aws.String(id),
	}

	out, err := conn.GetResourcePolicy(context.Background(), &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		// resource missing -> remove from state
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading Bedrock Agent Core Resource Policy (%s): %s", id, err)
	}

	if out != nil {
		if out.Policy != nil {
			// normalize returned policy JSON
			policy, err := structure.NormalizeJsonString(aws.ToString(out.Policy))
			if err == nil {
				d.Set("policy", policy)
			} else {
				// fall back to raw value if normalization fails
				d.Set("policy", aws.ToString(out.Policy))
			}
		}
		// Resource ARN is known from state (ID or attribute); ensure attribute is set.
		if id != "" {
			d.Set("resource_arn", id)
			d.SetId(id)
		}
	}

	return nil
}

func resourcePolicyUpdate(d *schema.ResourceData, meta any) error {
	conn := meta.(*conns.AWSClient).BedrockAgentCoreClient(context.Background())

	arn := d.Get("resource_arn").(string)
	policyRaw := d.Get("policy").(string)

	policy, err := structure.NormalizeJsonString(policyRaw)
	if err != nil {
		return fmt.Errorf("invalid policy JSON: %s", err)
	}

	input := bedrockagentcorecontrol.PutResourcePolicyInput{
		ResourceArn: aws.String(arn),
		Policy:      aws.String(policy),
	}

	_, err = conn.PutResourcePolicy(context.Background(), &input)
	if err != nil {
		return fmt.Errorf("updating Bedrock Agent Core Resource Policy (%s): %s", arn, err)
	}

	return resourcePolicyRead(d, meta)
}

func resourcePolicyDelete(d *schema.ResourceData, meta any) error {
	conn := meta.(*conns.AWSClient).BedrockAgentCoreClient(context.Background())

	arn := d.Id()
	if arn == "" {
		arn = d.Get("resource_arn").(string)
	}

	input := bedrockagentcorecontrol.DeleteResourcePolicyInput{
		ResourceArn: aws.String(arn),
	}

	_, err := conn.DeleteResourcePolicy(context.Background(), &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Bedrock Agent Core Resource Policy (%s): %s", arn, err)
	}

	d.SetId("")
	return nil
}
