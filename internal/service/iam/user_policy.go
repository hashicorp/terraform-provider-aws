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

func ResourceUserPolicy() *schema.Resource {
	return &schema.Resource{
		// PutUserPolicy API is idempotent, so these can be the same.
		Create: resourceUserPolicyPut,
		Read:   resourceUserPolicyRead,
		Update: resourceUserPolicyPut,
		Delete: resourceUserPolicyDelete,

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
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
			},
			"user": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceUserPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	request := &iam.PutUserPolicyInput{
		UserName:       aws.String(d.Get("user").(string)),
		PolicyDocument: aws.String(d.Get("policy").(string)),
	}

	var policyName string
	var err error
	if !d.IsNewResource() {
		_, policyName, err = UserPolicyParseID(d.Id())
		if err != nil {
			return err
		}
	} else if v, ok := d.GetOk("name"); ok {
		policyName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		policyName = resource.PrefixedUniqueId(v.(string))
	} else {
		policyName = resource.UniqueId()
	}
	request.PolicyName = aws.String(policyName)

	if _, err := conn.PutUserPolicy(request); err != nil {
		return fmt.Errorf("Error putting IAM user policy %s: %s", *request.PolicyName, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", *request.UserName, *request.PolicyName))
	return nil
}

func resourceUserPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	user, name, err := UserPolicyParseID(d.Id())
	if err != nil {
		return err
	}

	request := &iam.GetUserPolicyInput{
		PolicyName: aws.String(name),
		UserName:   aws.String(user),
	}

	var getResp *iam.GetUserPolicyOutput

	err = resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error

		getResp, err = conn.GetUserPolicy(request)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		getResp, err = conn.GetUserPolicy(request)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM User Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading IAM User Policy (%s): %w", d.Id(), err)
	}

	if getResp == nil || getResp.PolicyDocument == nil {
		return fmt.Errorf("error reading IAM User Policy (%s): empty response", d.Id())
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
	return d.Set("user", user)
}

func resourceUserPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	user, name, err := UserPolicyParseID(d.Id())
	if err != nil {
		return err
	}

	request := &iam.DeleteUserPolicyInput{
		PolicyName: aws.String(name),
		UserName:   aws.String(user),
	}

	if _, err := conn.DeleteUserPolicy(request); err != nil {
		if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return nil
		}
		return fmt.Errorf("Error deleting IAM user policy %s: %s", d.Id(), err)
	}
	return nil
}

func UserPolicyParseID(id string) (userName, policyName string, err error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		err = fmt.Errorf("user_policy id must be of the form <user name>:<policy name>")
		return
	}

	userName = parts[0]
	policyName = parts[1]
	return
}
