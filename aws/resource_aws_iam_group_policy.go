package aws

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsIamGroupPolicy() *schema.Resource {
	return &schema.Resource{
		// PutGroupPolicy API is idempotent, so these can be the same.
		Create: resourceAwsIamGroupPolicyPut,
		Update: resourceAwsIamGroupPolicyPut,

		Read:   resourceAwsIamGroupPolicyRead,
		Delete: resourceAwsIamGroupPolicyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validateIAMPolicyJson,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
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
	iamconn := meta.(*AWSClient).iamconn

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

	if _, err := iamconn.PutGroupPolicy(request); err != nil {
		return fmt.Errorf("Error putting IAM group policy %s: %s", *request.PolicyName, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", *request.GroupName, *request.PolicyName))
	return nil
}

func resourceAwsIamGroupPolicyRead(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	group, name, err := resourceAwsIamGroupPolicyParseId(d.Id())
	if err != nil {
		return err
	}

	request := &iam.GetGroupPolicyInput{
		PolicyName: aws.String(name),
		GroupName:  aws.String(group),
	}

	getResp, err := iamconn.GetGroupPolicy(request)
	if err != nil {
		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			log.Printf("[WARN] IAM Group Policy (%s) for %s not found, removing from state", name, group)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading IAM policy %s from group %s: %s", name, group, err)
	}

	if getResp.PolicyDocument == nil {
		return fmt.Errorf("GetGroupPolicy returned a nil policy document")
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

func resourceAwsIamGroupPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	group, name, err := resourceAwsIamGroupPolicyParseId(d.Id())
	if err != nil {
		return err
	}

	request := &iam.DeleteGroupPolicyInput{
		PolicyName: aws.String(name),
		GroupName:  aws.String(group),
	}

	if _, err := iamconn.DeleteGroupPolicy(request); err != nil {
		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
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
