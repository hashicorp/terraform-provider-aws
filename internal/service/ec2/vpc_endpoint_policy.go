package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVPCEndpointPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCEndpointPolicyPut,
		UpdateWithoutTimeout: resourceVPCEndpointPolicyPut,
		ReadWithoutTimeout:   resourceVPCEndpointPolicyRead,
		DeleteWithoutTimeout: resourceVPCEndpointPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"policy": {
				Type:                  schema.TypeString,
				Optional:              true,
				Computed:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"vpc_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceVPCEndpointPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	endpointID := d.Get("vpc_endpoint_id").(string)
	req := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId: aws.String(endpointID),
	}

	policy, err := structure.NormalizeJsonString(d.Get("policy"))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy contains an invalid JSON: %s", err)
	}

	if policy == "" {
		req.ResetPolicy = aws.Bool(true)
	} else {
		req.PolicyDocument = aws.String(policy)
	}

	log.Printf("[DEBUG] Updating VPC Endpoint Policy: %#v", req)
	if _, err := conn.ModifyVpcEndpointWithContext(ctx, req); err != nil {
		return sdkdiag.AppendErrorf(diags, "Error updating VPC Endpoint Policy: %s", err)
	}
	d.SetId(endpointID)

	_, err = WaitVPCEndpointAvailable(ctx, conn, endpointID, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for VPC Endpoint (%s) to policy to set: %s", endpointID, err)
	}

	return append(diags, resourceVPCEndpointPolicyRead(ctx, d, meta)...)
}

func resourceVPCEndpointPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	vpce, err := FindVPCEndpointByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPC Endpoint Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPC Endpoint Policy (%s): %s", d.Id(), err)
	}

	d.Set("vpc_endpoint_id", d.Id())

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.StringValue(vpce.PolicyDocument))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "while setting policy (%s), encountered: %s", policyToSet, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policyToSet, err)
	}

	d.Set("policy", policyToSet)
	return diags
}

func resourceVPCEndpointPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	req := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId: aws.String(d.Id()),
		ResetPolicy:   aws.Bool(true),
	}

	log.Printf("[DEBUG] Resetting VPC Endpoint Policy: %#v", req)
	if _, err := conn.ModifyVpcEndpointWithContext(ctx, req); err != nil {
		return sdkdiag.AppendErrorf(diags, "Error Resetting VPC Endpoint Policy: %s", err)
	}

	_, err := WaitVPCEndpointAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for VPC Endpoint (%s) to be reset: %s", d.Id(), err)
	}

	return diags
}
