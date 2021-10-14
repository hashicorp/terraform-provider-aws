package aws

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceGroupPolicy() *schema.Resource {
	return &schema.Resource{
		// PutGroupPolicy API is idempotent, so these can be the same.
		Create: resourceAwsIamGroupPolicyPut,
		Update: resourceAwsIamGroupPolicyPut,

		Read:   resourceGroupPolicyRead,
		Delete: resourceGroupPolicyDelete,

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
			"group": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsIamGroupPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	request := &iam.PutGroupPolicyInput{
		GroupName:      aws.String(d.Get("group").(string)),
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

	if _, err := conn.PutGroupPolicy(request); err != nil {
		return fmt.Errorf("Error putting IAM group policy %s: %s", *request.PolicyName, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", *request.GroupName, *request.PolicyName))
	return nil
}

func resourceGroupPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	group, name, err := resourceAwsIamGroupPolicyParseId(d.Id())
	if err != nil {
		return err
	}

	request := &iam.GetGroupPolicyInput{
		PolicyName: aws.String(name),
		GroupName:  aws.String(group),
	}

	var getResp *iam.GetGroupPolicyOutput

	err = resource.Retry(waiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		getResp, err = conn.GetGroupPolicy(request)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		getResp, err = conn.GetGroupPolicy(request)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
		log.Printf("[WARN] IAM Group Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading IAM Group Policy (%s): %w", d.Id(), err)
	}

	if getResp == nil || getResp.PolicyDocument == nil {
		return fmt.Errorf("error reading IAM Group Policy (%s): empty response", d.Id())
	}

	policy, err := url.QueryUnescape(*getResp.PolicyDocument)
	if err != nil {
		return err
	}

	if err := d.Set("policy", policy); err != nil {
		return fmt.Errorf("error setting policy: %s", err)
	}

	if err := d.Set("name", name); err != nil {
		return fmt.Errorf("error setting name: %s", err)
	}

	if err := d.Set("group", group); err != nil {
		return fmt.Errorf("error setting group: %s", err)
	}

	return nil
}

func resourceGroupPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	group, name, err := resourceAwsIamGroupPolicyParseId(d.Id())
	if err != nil {
		return err
	}

	request := &iam.DeleteGroupPolicyInput{
		PolicyName: aws.String(name),
		GroupName:  aws.String(group),
	}

	if _, err := conn.DeleteGroupPolicy(request); err != nil {
		if tfawserr.ErrMessageContains(err, iam.ErrCodeNoSuchEntityException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting IAM group policy %s: %s", d.Id(), err)
	}
	return nil
}

func resourceAwsIamGroupPolicyParseId(id string) (groupName, policyName string, err error) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		err = fmt.Errorf("group_policy id must be of the form <group name>:<policy name>")
		return
	}

	groupName = parts[0]
	policyName = parts[1]
	return
}
