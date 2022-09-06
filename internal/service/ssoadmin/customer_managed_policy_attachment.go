package ssoadmin

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	customerPolicyAttachmentTimeout = 5 * time.Minute
)

func ResourceCustomerManagedPolicyAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceCustomerManagedPolicyAttachmentCreate,
		Read:   resourceCustomerManagedPolicyAttachmentRead,
		Delete: resourceCustomerManagedPolicyAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{

			"instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"permission_set_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"customer_managed_policy_reference": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(0, 128),
						},
						"path": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "/",
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(0, 512),
						},
					},
				},
			},
		},
	}
}

func resourceCustomerManagedPolicyAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	instanceArn := d.Get("instance_arn").(string)
	permissionSetArn := d.Get("permission_set_arn").(string)
	policyName, policyPath := expandPolicyReference(d.Get("customer_managed_policy_reference").([]interface{}))

	input := &ssoadmin.AttachCustomerManagedPolicyReferenceToPermissionSetInput{
		InstanceArn:                    aws.String(instanceArn),
		CustomerManagedPolicyReference: formatPolicyReference(d.Get("customer_managed_policy_reference").([]interface{})),
		PermissionSetArn:               aws.String(permissionSetArn),
	}

	err := resource.Retry(customerPolicyAttachmentTimeout, func() *resource.RetryError {
		var err error
		_, err = conn.AttachCustomerManagedPolicyReferenceToPermissionSet(input)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeConflictException) {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeThrottlingException) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil

	})

	if tfresource.TimedOut(err) {
		_, err = conn.AttachCustomerManagedPolicyReferenceToPermissionSet(input)
	}

	if err != nil {
		return fmt.Errorf("error attaching Customer Managed Policy to SSO Permission Set (%s): %w", permissionSetArn, err)
	}

	d.SetId(fmt.Sprintf("%s,%s,%s,%s", policyName, policyPath, permissionSetArn, instanceArn))

	// Provision ALL accounts after attaching the managed policy
	if err := provisionPermissionSet(conn, permissionSetArn, instanceArn); err != nil {
		return err
	}

	return resourceCustomerManagedPolicyAttachmentRead(d, meta)
}

func resourceCustomerManagedPolicyAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	policyName, policyPath, permissionSetArn, instanceArn, err := ParseCustomerManagedPolicyAttachmentID(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing SSO Customer Managed Policy Attachment ID: %w", err)
	}

	policy, err := FindCustomerManagedPolicy(conn, policyName, policyPath, permissionSetArn, instanceArn)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Customer Managed Policy (%s) for SSO Permission Set (%s) not found, removing from state", policyName, permissionSetArn)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Customer Managed Policy (%s) for SSO Permission Set (%s): %w", policyName, permissionSetArn, err)
	}

	if policy == nil {
		log.Printf("[WARN] Customer Managed Policy (%s) for SSO Permission Set (%s) not found, removing from state", policyName, permissionSetArn)
		d.SetId("")
		return nil
	}

	d.Set("instance_arn", instanceArn)
	d.Set("permission_set_arn", permissionSetArn)
	d.Set("customer_managed_policy_reference", flattenPolicyReference(policy))

	return nil
}

func resourceCustomerManagedPolicyAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	policyName, policyPath, permissionSetArn, instanceArn, err := ParseCustomerManagedPolicyAttachmentID(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing SSO Customer Managed Policy Attachment ID: %w", err)
	}

	input := &ssoadmin.DetachCustomerManagedPolicyReferenceFromPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
		CustomerManagedPolicyReference: &ssoadmin.CustomerManagedPolicyReference{
			Name: aws.String(policyName),
			Path: aws.String(policyPath),
		},
	}
	err = resource.Retry(customerPolicyAttachmentTimeout, func() *resource.RetryError {
		var err error
		_, err = conn.DetachCustomerManagedPolicyReferenceFromPermissionSet(input)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeConflictException) {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeThrottlingException) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil

	})

	if tfresource.TimedOut(err) {
		_, err = conn.DetachCustomerManagedPolicyReferenceFromPermissionSet(input)
	}

	if err != nil {
		if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("error detaching Customer Managed Policy (%s) from SSO Permission Set (%s): %w", policyName, permissionSetArn, err)
	}

	// Provision ALL accounts after detaching the managed policy
	if err := provisionPermissionSet(conn, permissionSetArn, instanceArn); err != nil {
		return err
	}

	return nil
}

func ParseCustomerManagedPolicyAttachmentID(id string) (string, string, string, string, error) {
	idParts := strings.Split(id, ",")
	if len(idParts) != 4 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" {
		return "", "", "", "", fmt.Errorf("error parsing ID: expected CUSTOMER_MANAGED_POLICY_NAME, CUSTOMER_MANAGED_POLICY_PATH, PERMISSION_SET_ARN, INSTANCE_ARN")
	}
	return idParts[0], idParts[1], idParts[2], idParts[3], nil
}

func formatPolicyReference(l []interface{}) *ssoadmin.CustomerManagedPolicyReference {
	m := l[0].(map[string]interface{})

	policyRef := &ssoadmin.CustomerManagedPolicyReference{
		Name: aws.String(m["name"].(string)),
		Path: aws.String(m["path"].(string)),
	}

	return policyRef
}

func expandPolicyReference(l []interface{}) (string, string) {
	m := l[0].(map[string]interface{})

	policyName := m["name"].(string)
	policyPath := m["path"].(string)

	return policyName, policyPath
}

func flattenPolicyReference(l *ssoadmin.CustomerManagedPolicyReference) []interface{} {
	if l == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if v := l.Name; v != nil {
		m["name"] = aws.StringValue(v)
	}

	if v := l.Path; v != nil {
		m["path"] = aws.StringValue(v)
	}

	return []interface{}{m}
}
