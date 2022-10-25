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

func ResourceManagedPolicyAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceManagedPolicyAttachmentCreate,
		Read:   resourceManagedPolicyAttachmentRead,
		Delete: resourceManagedPolicyAttachmentDelete,
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

			"managed_policy_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},

			"managed_policy_name": {
				Type:     schema.TypeString,
				Computed: true,
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

func resourceManagedPolicyAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	instanceArn := d.Get("instance_arn").(string)
	managedPolicyArn := d.Get("managed_policy_arn").(string)
	permissionSetArn := d.Get("permission_set_arn").(string)

	input := &ssoadmin.AttachManagedPolicyToPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		ManagedPolicyArn: aws.String(managedPolicyArn),
		PermissionSetArn: aws.String(permissionSetArn),
	}

	_, err := conn.AttachManagedPolicyToPermissionSet(input)

	if err != nil {
		return fmt.Errorf("error attaching Managed Policy to SSO Permission Set (%s): %w", permissionSetArn, err)
	}

	d.SetId(fmt.Sprintf("%s,%s,%s", managedPolicyArn, permissionSetArn, instanceArn))

	// Provision ALL accounts after attaching the managed policy
	if err := provisionPermissionSet(conn, permissionSetArn, instanceArn); err != nil {
		return err
	}

	return resourceManagedPolicyAttachmentRead(d, meta)
}

func resourceManagedPolicyAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	managedPolicyArn, permissionSetArn, instanceArn, err := ParseManagedPolicyAttachmentID(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing SSO Managed Policy Attachment ID: %w", err)
	}

	policy, err := FindManagedPolicy(conn, managedPolicyArn, permissionSetArn, instanceArn)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Managed Policy (%s) for SSO Permission Set (%s) not found, removing from state", managedPolicyArn, permissionSetArn)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Managed Policy (%s) for SSO Permission Set (%s): %w", managedPolicyArn, permissionSetArn, err)
	}

	if policy == nil {
		log.Printf("[WARN] Managed Policy (%s) for SSO Permission Set (%s) not found, removing from state", managedPolicyArn, permissionSetArn)
		d.SetId("")
		return nil
	}

	d.Set("instance_arn", instanceArn)
	d.Set("managed_policy_arn", policy.Arn)
	d.Set("managed_policy_name", policy.Name)
	d.Set("permission_set_arn", permissionSetArn)

	return nil
}

func resourceManagedPolicyAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	managedPolicyArn, permissionSetArn, instanceArn, err := ParseManagedPolicyAttachmentID(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing SSO Managed Policy Attachment ID: %w", err)
	}

	input := &ssoadmin.DetachManagedPolicyFromPermissionSetInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
		ManagedPolicyArn: aws.String(managedPolicyArn),
	}

	_, err = conn.DetachManagedPolicyFromPermissionSet(input)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("error detaching Managed Policy (%s) from SSO Permission Set (%s): %w", managedPolicyArn, permissionSetArn, err)
	}

	// Provision ALL accounts after detaching the managed policy
	if err := provisionPermissionSet(conn, permissionSetArn, instanceArn); err != nil {
		return err
	}

	return nil
}

func ParseManagedPolicyAttachmentID(id string) (string, string, string, error) {
	idParts := strings.Split(id, ",")
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		return "", "", "", fmt.Errorf("error parsing ID: expected MANAGED_POLICY_ARN,PERMISSION_SET_ARN,INSTANCE_ARN")
	}
	return idParts[0], idParts[1], idParts[2], nil
}
