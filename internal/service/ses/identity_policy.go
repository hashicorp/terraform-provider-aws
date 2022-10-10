package ses

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceIdentityPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceIdentityPolicyCreate,
		Read:   resourceIdentityPolicyRead,
		Update: resourceIdentityPolicyUpdate,
		Delete: resourceIdentityPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"identity": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\-\_]+$`), "must contain only alphanumeric characters, dashes, and underscores"),
				),
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
		},
	}
}

func resourceIdentityPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn

	identity := d.Get("identity").(string)
	policyName := d.Get("name").(string)

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", d.Get("policy").(string), err)
	}

	input := &ses.PutIdentityPolicyInput{
		Identity:   aws.String(identity),
		PolicyName: aws.String(policyName),
		Policy:     aws.String(policy),
	}

	_, err = conn.PutIdentityPolicy(input)
	if err != nil {
		return fmt.Errorf("creating SES Identity (%s) Policy: %s", identity, err)
	}

	d.SetId(fmt.Sprintf("%s|%s", identity, policyName))

	return resourceIdentityPolicyRead(d, meta)
}

func resourceIdentityPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn

	identity, policyName, err := IdentityPolicyParseID(d.Id())
	if err != nil {
		return err
	}

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", d.Get("policy").(string), err)
	}

	req := ses.PutIdentityPolicyInput{
		Identity:   aws.String(identity),
		PolicyName: aws.String(policyName),
		Policy:     aws.String(policy),
	}

	_, err = conn.PutIdentityPolicy(&req)
	if err != nil {
		return fmt.Errorf("updating SES Identity (%s) Policy (%s): %s", identity, policyName, err)
	}

	return resourceIdentityPolicyRead(d, meta)
}

func resourceIdentityPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn

	identity, policyName, err := IdentityPolicyParseID(d.Id())
	if err != nil {
		return err
	}

	input := &ses.GetIdentityPoliciesInput{
		Identity:    aws.String(identity),
		PolicyNames: aws.StringSlice([]string{policyName}),
	}

	output, err := conn.GetIdentityPolicies(input)

	if err != nil {
		return fmt.Errorf("getting SES Identity (%s) Policy (%s): %s", identity, policyName, err)
	}

	if output == nil {
		return fmt.Errorf("getting SES Identity (%s) Policy (%s): empty result", identity, policyName)
	}

	if len(output.Policies) == 0 {
		log.Printf("[WARN] SES Identity (%s) Policy (%s) not found, removing from state", identity, policyName)
		d.SetId("")
		return nil
	}

	policy, ok := output.Policies[policyName]
	if !ok {
		log.Printf("[WARN] SES Identity (%s) Policy (%s) not found, removing from state", identity, policyName)
		d.SetId("")
		return nil
	}

	d.Set("identity", identity)
	d.Set("name", policyName)

	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), aws.StringValue(policy))

	if err != nil {
		return err
	}

	d.Set("policy", policyToSet)

	return nil
}

func resourceIdentityPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn

	identity, policyName, err := IdentityPolicyParseID(d.Id())
	if err != nil {
		return err
	}

	input := &ses.DeleteIdentityPolicyInput{
		Identity:   aws.String(identity),
		PolicyName: aws.String(policyName),
	}

	log.Printf("[DEBUG] Deleting SES Identity Policy: %s", input)
	_, err = conn.DeleteIdentityPolicy(input)

	if err != nil {
		return fmt.Errorf("deleting SES Identity (%s) Policy (%s): %s", identity, policyName, err)
	}

	return nil
}

func IdentityPolicyParseID(id string) (string, string, error) {
	idParts := strings.SplitN(id, "|", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected IDENTITY|NAME", id)
	}
	return idParts[0], idParts[1], nil
}
