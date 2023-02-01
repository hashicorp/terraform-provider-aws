package networkfirewall

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceResourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourcePolicyPut,
		ReadWithoutTimeout:   resourceResourcePolicyRead,
		UpdateWithoutTimeout: resourceResourcePolicyPut,
		DeleteWithoutTimeout: resourceResourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"resource_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceResourcePolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkFirewallConn()
	resourceArn := d.Get("resource_arn").(string)

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return diag.Errorf("policy (%s) is invalid JSON: %s", policy, err)
	}

	input := &networkfirewall.PutResourcePolicyInput{
		ResourceArn: aws.String(resourceArn),
		Policy:      aws.String(policy),
	}

	log.Printf("[DEBUG] Putting NetworkFirewall Resource Policy for resource: %s", resourceArn)

	_, err = conn.PutResourcePolicyWithContext(ctx, input)
	if err != nil {
		return diag.Errorf("error putting NetworkFirewall Resource Policy (for resource: %s): %s", resourceArn, err)
	}

	d.SetId(resourceArn)

	return resourceResourcePolicyRead(ctx, d, meta)
}

func resourceResourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkFirewallConn()
	resourceArn := d.Id()

	log.Printf("[DEBUG] Reading NetworkFirewall Resource Policy for resource: %s", resourceArn)

	policy, err := FindResourcePolicy(ctx, conn, resourceArn)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] NetworkFirewall Resource Policy (for resource: %s) not found, removing from state", resourceArn)
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Resource Policy (for resource: %s): %w", resourceArn, err))
	}

	if policy == nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Resource Policy (for resource: %s): empty output", resourceArn))
	}

	d.Set("resource_arn", resourceArn)

	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), aws.StringValue(policy))

	if err != nil {
		return diag.Errorf("setting policy %s: %s", aws.StringValue(policy), err)
	}

	d.Set("policy", policyToSet)

	return nil
}

func resourceResourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkFirewallConn()

	log.Printf("[DEBUG] Deleting NetworkFirewall Resource Policy for resource: %s", d.Id())

	input := &networkfirewall.DeleteResourcePolicyInput{
		ResourceArn: aws.String(d.Id()),
	}

	_, err := tfresource.RetryWhen(ctx, resourcePolicyDeleteTimeout,
		func() (interface{}, error) {
			return conn.DeleteResourcePolicyWithContext(ctx, input)
		},
		func(err error) (bool, error) {
			// RAM managed permissions eventual consistency
			if tfawserr.ErrMessageContains(err, networkfirewall.ErrCodeInvalidResourcePolicyException, "The supplied policy does not match RAM managed permissions") {
				return true, err
			}
			return false, err
		},
	)

	if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting NetworkFirewall Resource Policy (for resource: %s): %w", d.Id(), err))
	}

	return nil
}
