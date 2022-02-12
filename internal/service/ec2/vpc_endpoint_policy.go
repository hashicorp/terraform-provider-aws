package ec2

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceVPCEndpointPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCEndpointPolicyPut,
		Update: resourceVPCEndpointPolicyPut,
		Read:   resourceVPCEndpointPolicyRead,
		Delete: resourceVPCEndpointPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"policy": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
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

func resourceVPCEndpointPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	endpointID := d.Get("vpc_endpoint_id").(string)
	req := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId: aws.String(endpointID),
	}

	policy, err := structure.NormalizeJsonString(d.Get("policy"))
	if err != nil {
		return fmt.Errorf("policy contains an invalid JSON: %w", err)
	}

	if policy == "" {
		req.ResetPolicy = aws.Bool(true)
	} else {
		req.PolicyDocument = aws.String(policy)
	}

	log.Printf("[DEBUG] Updating VPC Endpoint Policy: %#v", req)
	if _, err := conn.ModifyVpcEndpoint(req); err != nil {
		return fmt.Errorf("Error updating VPC Endpoint Policy: %w", err)
	}
	d.SetId(endpointID)

	_, err = WaitVPCEndpointAvailable(conn, endpointID, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("error waiting for VPC Endpoint (%s) to policy to set: %w", endpointID, err)
	}

	return resourceVPCEndpointPolicyRead(d, meta)
}

func resourceVPCEndpointPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	vpce, err := FindVPCEndpointByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPC Endpoint Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading VPC Endpoint Policy (%s): %w", d.Id(), err)
	}

	d.Set("vpc_endpoint_id", d.Id())

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.StringValue(vpce.PolicyDocument))

	if err != nil {
		return fmt.Errorf("while setting policy (%s), encountered: %w", policyToSet, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", policyToSet, err)
	}

	d.Set("policy", policyToSet)
	return nil
}

func resourceVPCEndpointPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	req := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId: aws.String(d.Id()),
		ResetPolicy:   aws.Bool(true),
	}

	log.Printf("[DEBUG] Resetting VPC Endpoint Policy: %#v", req)
	if _, err := conn.ModifyVpcEndpoint(req); err != nil {
		return fmt.Errorf("Error Resetting VPC Endpoint Policy: %w", err)
	}

	_, err := WaitVPCEndpointAvailable(conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return fmt.Errorf("error waiting for VPC Endpoint (%s) to be reset: %w", d.Id(), err)
	}

	return nil
}
