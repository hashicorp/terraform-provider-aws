package ssoadmin

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
			"customer_managed_policy_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"customer_managed_policy_path": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/",
				ForceNew: true,
			},
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
		},
	}
}

func resourceCustomerManagedPolicyAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	instanceArn := d.Get("instance_arn").(string)
	permissionSetArn := d.Get("permission_set_arn").(string)
	policyName := d.Get("customer_managed_policy_name").(string)
	policyPath := d.Get("customer_managed_policy_path").(string)

	input := &ssoadmin.AttachCustomerManagedPolicyReferenceToPermissionSetInput{
		InstanceArn:                    aws.String(instanceArn),
		CustomerManagedPolicyReference: createCustomerManagedPolicyReference(policyName, policyPath),
		PermissionSetArn:               aws.String(permissionSetArn),
	}

	_, err := conn.AttachCustomerManagedPolicyReferenceToPermissionSet(input)

	if err != nil {
		return fmt.Errorf("error attaching Customer Managed Policy to SSO Permission Set (%s): %w", permissionSetArn, err)
	}

	d.SetId(fmt.Sprintf("%s,%s,%s", policyName, permissionSetArn, instanceArn))

	// Provision ALL accounts after attaching the managed policy
	if err := provisionPermissionSet(conn, permissionSetArn, instanceArn); err != nil {
		return err
	}

	return resourceCustomerManagedPolicyAttachmentRead(d, meta)
}

func resourceCustomerManagedPolicyAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	customerManagedPolicyReference, permissionSetArn, instanceArn, err := ParseCustomerManagedPolicyAttachmentID(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing SSO Customer Managed Policy Attachment ID: %w", err)
	}

	policy, err := FindManagedPolicy(conn, customerManagedPolicyReference, permissionSetArn, instanceArn)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Customer Managed Policy (%s) for SSO Permission Set (%s) not found, removing from state", customerManagedPolicyReference, permissionSetArn)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Customer Managed Policy (%s) for SSO Permission Set (%s): %w", customerManagedPolicyReference, permissionSetArn, err)
	}

	if policy == nil {
		log.Printf("[WARN] Customer Managed Policy (%s) for SSO Permission Set (%s) not found, removing from state", customerManagedPolicyReference, permissionSetArn)
		d.SetId("")
		return nil
	}

	d.Set("instance_arn", instanceArn)
	d.Set("customer_managed_policy_name", policy.Arn)   // check
	d.Set("customer_managed_policy_path", policy.Name) //check
	d.Set("permission_set_arn", permissionSetArn)

	return nil
}

func resourceCustomerManagedPolicyAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	customerManagedPolicyReference, permissionSetArn, instanceArn, err := ParseCustomerManagedPolicyAttachmentID(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing SSO Customer Managed Policy Attachment ID: %w", err)
	}

	input := &ssoadmin.DetachCustomerManagedPolicyFromPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
		CustomerManagedPolicyReference: aws.String(customerManagedPolicyReference),
	}

	_, err = conn.DetachCustomerManagedPolicyFromPermissionSet(input)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("error detaching Customer Managed Policy (%s) from SSO Permission Set (%s): %w", customerManagedPolicyReference, permissionSetArn, err)
	}

	// Provision ALL accounts after detaching the managed policy
	if err := provisionPermissionSet(conn, permissionSetArn, instanceArn); err != nil {
		return err
	}

	return nil
}

func ParseCustomerManagedPolicyAttachmentID(id string) (string, string, string, error) { 
	idParts := strings.Split(id, ",")
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		return "", "", "", fmt.Errorf("error parsing ID: expected CUSTOMER_MANAGED_POLICY_REFERENCE,PERMISSION_SET_ARN,INSTANCE_ARN")
	}
	return idParts[0], idParts[1], idParts[2], nil
}

func createCustomerManagedPolicyReference(name string, path string) {
	customerManagedPolicyReference := map[string]string{"Name": path, "Path": path}
	return customerManagedPolicyReference
}
