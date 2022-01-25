package mediastore

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceContainerPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceContainerPolicyPut,
		Read:   resourceContainerPolicyRead,
		Update: resourceContainerPolicyPut,
		Delete: resourceContainerPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"container_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     verify.ValidIAMPolicyJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceContainerPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).MediaStoreConn

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", policy, err)
	}

	input := &mediastore.PutContainerPolicyInput{
		ContainerName: aws.String(d.Get("container_name").(string)),
		Policy:        aws.String(policy),
	}

	_, err = conn.PutContainerPolicy(input)
	if err != nil {
		return err
	}

	d.SetId(d.Get("container_name").(string))
	return resourceContainerPolicyRead(d, meta)
}

func resourceContainerPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).MediaStoreConn

	input := &mediastore.GetContainerPolicyInput{
		ContainerName: aws.String(d.Id()),
	}

	resp, err := conn.GetContainerPolicy(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, mediastore.ErrCodeContainerNotFoundException, "") {
			log.Printf("[WARN] MediaContainer Policy %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		if tfawserr.ErrMessageContains(err, mediastore.ErrCodePolicyNotFoundException, "") {
			log.Printf("[WARN] MediaContainer Policy %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("container_name", d.Id())

	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), aws.StringValue(resp.Policy))

	if err != nil {
		return err
	}

	d.Set("policy", policyToSet)

	return nil
}

func resourceContainerPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).MediaStoreConn

	input := &mediastore.DeleteContainerPolicyInput{
		ContainerName: aws.String(d.Id()),
	}

	_, err := conn.DeleteContainerPolicy(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, mediastore.ErrCodeContainerNotFoundException, "") {
			return nil
		}
		if tfawserr.ErrMessageContains(err, mediastore.ErrCodePolicyNotFoundException, "") {
			return nil
		}
		// if isAWSErr(err, mediastore.ErrCodeContainerInUseException, "Container must be ACTIVE in order to perform this operation") {
		// 	return nil
		// }
		return err
	}

	return nil
}
