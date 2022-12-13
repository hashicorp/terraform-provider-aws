package ssoadmin

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	permissionsBoundaryAttachmentTimeout = 5 * time.Minute
)

func ResourcePermissionsBoundaryAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourcePermissionsBoundaryAttachmentCreate,
		Read:   resourcePermissionsBoundaryAttachmentRead,
		Delete: resourcePermissionsBoundaryAttachmentDelete,

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
			"permissions_boundary": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"customer_managed_policy_reference": {
							Type:     schema.TypeList,
							Optional: true,
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
						"managed_policy_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "",
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(0, 2048),
						},
					},
				},
			},
		},
	}
}

func resourcePermissionsBoundaryAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	tfMap := d.Get("permissions_boundary").([]interface{})[0].(map[string]interface{})
	instanceARN := d.Get("instance_arn").(string)
	permissionSetARN := d.Get("permission_set_arn").(string)
	id := PermissionsBoundaryAttachmentCreateResourceID(permissionSetARN, instanceARN)
	input := &ssoadmin.PutPermissionsBoundaryToPermissionSetInput{
		PermissionsBoundary: expandPermissionsBoundary(tfMap),
		InstanceArn:         aws.String(instanceARN),
		PermissionSetArn:    aws.String(permissionSetARN),
	}

	log.Printf("[INFO] Attaching permissions boundary to permission set: %s", input)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(permissionsBoundaryAttachmentTimeout, func() (interface{}, error) {
		return conn.PutPermissionsBoundaryToPermissionSet(input)
	}, ssoadmin.ErrCodeConflictException, ssoadmin.ErrCodeThrottlingException)

	if err != nil {
		return fmt.Errorf("creating SSO Permissions Boundary Attachment (%s): %w", id, err)
	}

	d.SetId(id)

	// After the policy has been attached to the permission set, provision in all accounts that use this permission set.
	if err := provisionPermissionSet(conn, permissionSetARN, instanceARN); err != nil {
		return err
	}

	return resourcePermissionsBoundaryAttachmentRead(d, meta)
}

func resourcePermissionsBoundaryAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	permissionSetARN, instanceARN, err := PermissionsBoundaryAttachmentParseResourceID(d.Id())

	if err != nil {
		return err
	}

	policy, err := FindPermissionsBoundary(conn, permissionSetARN, instanceARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSO Permissions Boundary Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading SSO Permissions Boundary Attachment (%s): %w", d.Id(), err)
	}

	if err := d.Set("permissions_boundary", []interface{}{flattenPermissionsBoundary(policy)}); err != nil {
		return fmt.Errorf("setting permissions_boundary: %w", err)
	}
	d.Set("instance_arn", instanceARN)
	d.Set("permission_set_arn", permissionSetARN)

	return nil
}

func resourcePermissionsBoundaryAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	permissionSetARN, instanceARN, err := PermissionsBoundaryAttachmentParseResourceID(d.Id())

	if err != nil {
		return err
	}

	input := &ssoadmin.DeletePermissionsBoundaryFromPermissionSetInput{
		InstanceArn:      aws.String(instanceARN),
		PermissionSetArn: aws.String(permissionSetARN),
	}

	log.Printf("[INFO] Detaching permissions boundary from permission set: %s", input)
	_, err = tfresource.RetryWhenAWSErrCodeEquals(permissionsBoundaryAttachmentTimeout, func() (interface{}, error) {
		return conn.DeletePermissionsBoundaryFromPermissionSet(input)
	}, ssoadmin.ErrCodeConflictException, ssoadmin.ErrCodeThrottlingException)

	if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting SSO Permissions Boundary Attachment (%s): %w", d.Id(), err)
	}

	// After the policy has been detached from the permission set, provision in all accounts that use this permission set.
	if err := provisionPermissionSet(conn, permissionSetARN, instanceARN); err != nil {
		return err
	}

	return nil
}

const permissionsBoundaryAttachmentIDSeparator = ","

func PermissionsBoundaryAttachmentCreateResourceID(permissionSetARN, instanceARN string) string {
	parts := []string{permissionSetARN, instanceARN}
	id := strings.Join(parts, permissionsBoundaryAttachmentIDSeparator)

	return id
}

func PermissionsBoundaryAttachmentParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, permissionsBoundaryAttachmentIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected PERMISSION_SET_ARN%[2]sINSTANCE_ARN", id, permissionsBoundaryAttachmentIDSeparator)
}

func expandPermissionsBoundary(tfMap map[string]interface{}) *ssoadmin.PermissionsBoundary {
	if tfMap == nil {
		return nil
	}

	apiObject := &ssoadmin.PermissionsBoundary{}

	if v, ok := tfMap["customer_managed_policy_reference"].([]interface{}); ok && len(v) > 0 {
		if cmpr, ok := v[0].(map[string]interface{}); ok {
			apiObject.CustomerManagedPolicyReference = expandCustomerManagedPolicyReference(cmpr)
		}
	}
	if v, ok := tfMap["managed_policy_arn"].(string); ok && v != "" {
		apiObject.ManagedPolicyArn = aws.String(v)
	}

	return apiObject
}

func flattenPermissionsBoundary(apiObject *ssoadmin.PermissionsBoundary) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ManagedPolicyArn; v != nil {
		tfMap["managed_policy_arn"] = aws.StringValue(v)
	} else if v := apiObject.CustomerManagedPolicyReference; v != nil {
		tfMap["customer_managed_policy_reference"] = []map[string]interface{}{flattenCustomerManagedPolicyReference(v)}
	}

	return tfMap
}
