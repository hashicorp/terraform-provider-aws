package iot

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourcePolicyCreate,
		Read:   resourcePolicyRead,
		Update: resourcePolicyUpdate,
		Delete: resourcePolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_version_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourcePolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", policy, err)
	}

	out, err := conn.CreatePolicy(&iot.CreatePolicyInput{
		PolicyName:     aws.String(d.Get("name").(string)),
		PolicyDocument: aws.String(policy),
	})

	if err != nil {
		return fmt.Errorf("error creating IoT Policy: %s", err)
	}

	d.SetId(aws.StringValue(out.PolicyName))

	return resourcePolicyRead(d, meta)
}

func resourcePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	out, err := conn.GetPolicy(&iot.GetPolicyInput{
		PolicyName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] IoT Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading IoT Policy (%s): %s", d.Id(), err)
	}

	d.Set("arn", out.PolicyArn)
	d.Set("default_version_id", out.DefaultVersionId)
	d.Set("name", out.PolicyName)

	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), aws.StringValue(out.PolicyDocument))

	if err != nil {
		return err
	}

	d.Set("policy", policyToSet)

	return nil
}

func resourcePolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	if d.HasChange("policy") {
		policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

		if err != nil {
			return fmt.Errorf("policy (%s) is invalid JSON: %w", policy, err)
		}

		_, err = conn.CreatePolicyVersion(&iot.CreatePolicyVersionInput{
			PolicyName:     aws.String(d.Id()),
			PolicyDocument: aws.String(policy),
			SetAsDefault:   aws.Bool(true),
		})

		if err != nil {
			return fmt.Errorf("error updating IoT Policy (%s): %s", d.Id(), err)
		}
	}

	return resourcePolicyRead(d, meta)
}

func resourcePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IoTConn

	out, err := conn.ListPolicyVersions(&iot.ListPolicyVersionsInput{
		PolicyName: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error listing IoT Policy (%s) versions: %s", d.Id(), err)
	}

	// Delete all non-default versions of the policy
	for _, ver := range out.PolicyVersions {
		if !aws.BoolValue(ver.IsDefaultVersion) {
			_, err = conn.DeletePolicyVersion(&iot.DeletePolicyVersionInput{
				PolicyName:      aws.String(d.Id()),
				PolicyVersionId: ver.VersionId,
			})

			if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error deleting IoT Policy (%s) version (%s): %s", d.Id(), aws.StringValue(ver.VersionId), err)
			}
		}
	}

	//Delete default policy version
	_, err = conn.DeletePolicy(&iot.DeletePolicyInput{
		PolicyName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting IoT Policy (%s): %s", d.Id(), err)
	}

	return nil
}
