package iam

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	rolePolicyNameMaxLen       = 128
	rolePolicyNamePrefixMaxLen = rolePolicyNameMaxLen - resource.UniqueIDSuffixLength
)

func ResourceRolePolicy() *schema.Resource {
	return &schema.Resource{
		// PutRolePolicy API is idempotent, so these can be the same.
		Create: resourceRolePolicyPut,
		Update: resourceRolePolicyPut,

		Read:   resourceRolePolicyRead,
		Delete: resourceRolePolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     verify.ValidIAMPolicyJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validRolePolicyName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validResourceName(rolePolicyNamePrefixMaxLen),
			},
			"role": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceRolePolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	request := &iam.PutRolePolicyInput{
		RoleName:       aws.String(d.Get("role").(string)),
		PolicyDocument: aws.String(d.Get("policy").(string)),
	}

	var policyName string
	if v, ok := d.GetOk("name"); ok {
		policyName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		policyName = resource.PrefixedUniqueId(v.(string))
	} else {
		policyName = resource.UniqueId()
	}
	request.PolicyName = aws.String(policyName)

	if _, err := conn.PutRolePolicy(request); err != nil {
		return fmt.Errorf("Error putting IAM role policy %s: %s", *request.PolicyName, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", *request.RoleName, *request.PolicyName))
	return resourceRolePolicyRead(d, meta)
}

func resourceRolePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	role, name, err := RolePolicyParseID(d.Id())
	if err != nil {
		return err
	}

	request := &iam.GetRolePolicyInput{
		PolicyName: aws.String(name),
		RoleName:   aws.String(role),
	}

	var getResp *iam.GetRolePolicyOutput

	err = resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error

		getResp, err = conn.GetRolePolicy(request)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		getResp, err = conn.GetRolePolicy(request)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM Role Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading IAM Role Policy (%s): %w", d.Id(), err)
	}

	if getResp == nil || getResp.PolicyDocument == nil {
		return fmt.Errorf("error reading IAM Role Policy (%s): empty response", d.Id())
	}

	policy, err := url.QueryUnescape(*getResp.PolicyDocument)
	if err != nil {
		return err
	}

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), policy)

	if err != nil {
		return fmt.Errorf("while setting policy (%s), encountered: %w", policyToSet, err)
	}

	d.Set("policy", policyToSet)

	if err := d.Set("name", name); err != nil {
		return err
	}
	return d.Set("role", role)
}

func resourceRolePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	role, name, err := RolePolicyParseID(d.Id())
	if err != nil {
		return err
	}

	request := &iam.DeleteRolePolicyInput{
		PolicyName: aws.String(name),
		RoleName:   aws.String(role),
	}

	if _, err := conn.DeleteRolePolicy(request); err != nil {
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return nil
		}
		return fmt.Errorf("Error deleting IAM role policy %s: %s", d.Id(), err)
	}
	return nil
}

func RolePolicyParseID(id string) (roleName, policyName string, err error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("role_policy id must be of the form <role name>:<policy name>")
		return
	}

	roleName = parts[0]
	policyName = parts[1]
	return
}
