package ssoadmin

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceAccountAssignments() *schema.Resource {
	return &schema.Resource{
		Create: resourceAccountAssignmentsCreate,
		Read:   resourceAccountAssignmentsRead,
		Delete: resourceAccountAssignmentsDelete,
		Update: resourceAccountAssignmentsUpdate,

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

			"principal_ids": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 47),
						validation.StringMatch(regexp.MustCompile(`^([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}$`), "must match ([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}"),
					),
				},
			},

			"principal_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ssoadmin.PrincipalType_Values(), false),
			},

			"target_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},

			"target_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ssoadmin.TargetType_Values(), false),
			},
		},
	}
}

func resourceAccountAssignmentsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	principalIDs := []*string{}
	if v, ok := d.GetOk("principal_ids"); ok {
		principalIDs = flex.ExpandStringSet(v.(*schema.Set))
	}

	instanceArn := d.Get("instance_arn").(string)
	permissionSetArn := d.Get("permission_set_arn").(string)
	principalType := d.Get("principal_type").(string)
	targetID := d.Get("target_id").(string)
	targetType := d.Get("target_type").(string)

	// We need to check if any of the assignments exists before creating them
	// since the AWS SSO API doesn't prevent us from creating duplicates
	assignedIDs, err := FindAccountAssignmentPrincipals(conn, principalType, targetID, permissionSetArn, instanceArn)
	if err != nil {
		return fmt.Errorf("error listing SSO Account Assignments for AccountId (%s) PermissionSet (%s): %w", targetID, permissionSetArn, err)
	}

	if len(assignedIDs) > 0 {
		return fmt.Errorf("error creating SSO Account Assignments for %s: already exists", principalType)
	}

	err = createAccountAssignments(conn, instanceArn, permissionSetArn, targetType, targetID, principalType, principalIDs)

	if err != nil {
		return fmt.Errorf("error creating SSO Account Assignments for %s: %w", principalType, err)
	}

	d.SetId(fmt.Sprintf("%s,%s,%s,%s,%s", principalType, targetID, targetType, permissionSetArn, instanceArn))

	return resourceAccountAssignmentsRead(d, meta)
}

func resourceAccountAssignmentsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	idParts, err := ParseAccountAssignmentsID(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing SSO Account Assignment ID: %w", err)
	}

	principalType := idParts[0]
	targetID := idParts[1]
	targetType := idParts[2]
	permissionSetArn := idParts[3]
	instanceArn := idParts[4]

	assignedIDs, err := FindAccountAssignmentPrincipals(conn, principalType, targetID, permissionSetArn, instanceArn)

	if err != nil {
		return fmt.Errorf("error listing SSO Account Assignments for AccountId (%s) PermissionSet (%s): %w", targetID, permissionSetArn, err)
	}

	d.Set("instance_arn", instanceArn)
	d.Set("permission_set_arn", permissionSetArn)
	d.Set("principal_ids", assignedIDs)
	d.Set("principal_type", principalType)
	d.Set("target_id", targetID)
	d.Set("target_type", targetType)

	return nil
}

func resourceAccountAssignmentsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	idParts, err := ParseAccountAssignmentsID(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing SSO Account Assignment ID: %w", err)
	}

	principalType := idParts[0]
	targetID := idParts[1]
	targetType := idParts[2]
	permissionSetArn := idParts[3]
	instanceArn := idParts[4]

	principalIDs := []*string{}
	if v, ok := d.GetOk("principal_ids"); ok {
		principalIDs = flex.ExpandStringSet(v.(*schema.Set))
	}

	err = deleteAccountAssignments(conn, instanceArn, permissionSetArn, targetType, targetID, principalType, principalIDs)

	if err != nil {
		return fmt.Errorf("error deleting SSO Account Assignments for %s: %w", principalType, err)
	}

	return nil
}

func resourceAccountAssignmentsUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	idParts, err := ParseAccountAssignmentsID(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing SSO Account Assignment ID: %w", err)
	}

	principalType := idParts[0]
	targetID := idParts[1]
	targetType := idParts[2]
	permissionSetArn := idParts[3]
	instanceArn := idParts[4]

	principalIDs := []*string{}
	if v, ok := d.GetOk("principal_ids"); ok {
		principalIDs = flex.ExpandStringSet(v.(*schema.Set))
	}

	assignedIDs, err := FindAccountAssignmentPrincipals(conn, principalType, targetID, permissionSetArn, instanceArn)

	if err != nil {
		return fmt.Errorf("error listing SSO Account Assignments for AccountId (%s) PermissionSet (%s): %w", targetID, permissionSetArn, err)
	}

	var createPrincipalIDs []*string
	for _, principalID := range principalIDs {
		found := false
		for _, assignedID := range assignedIDs {
			if principalID == assignedID {
				found = true
				break
			}
		}
		if !found {
			createPrincipalIDs = append(createPrincipalIDs, principalID)
		}
	}

	err = createAccountAssignments(conn, instanceArn, permissionSetArn, targetType, targetID, principalType, createPrincipalIDs)

	if err != nil {
		return fmt.Errorf("error creating SSO Account Assignments for %s: %w", principalType, err)
	}

	var deletePrincipalIDs []*string
	for _, assignedID := range assignedIDs {
		found := false
		for _, principalID := range principalIDs {
			if principalID == assignedID {
				found = true
				break
			}
		}
		if !found {
			deletePrincipalIDs = append(deletePrincipalIDs, assignedID)
		}
	}

	err = deleteAccountAssignments(conn, instanceArn, permissionSetArn, targetType, targetID, principalType, deletePrincipalIDs)

	if err != nil {
		return fmt.Errorf("error deleting SSO Account Assignments for %s: %w", principalType, err)
	}

	return resourceAccountAssignmentsRead(d, meta)
}

func createAccountAssignments(conn *ssoadmin.SSOAdmin, instanceArn string, permissionSetArn string, targetType string, targetID string, principalType string, principalIDs []*string) error {

	for _, principalID := range principalIDs {
		input := &ssoadmin.CreateAccountAssignmentInput{
			InstanceArn:      aws.String(instanceArn),
			PermissionSetArn: aws.String(permissionSetArn),
			PrincipalId:      aws.String(*principalID),
			PrincipalType:    aws.String(principalType),
			TargetId:         aws.String(targetID),
			TargetType:       aws.String(targetType),
		}

		output, err := conn.CreateAccountAssignment(input)
		if err != nil {
			return fmt.Errorf("error creating SSO Account Assignment for %s (%s): %w", principalType, *principalID, err)
		}

		if output == nil || output.AccountAssignmentCreationStatus == nil {
			return fmt.Errorf("error creating SSO Account Assignment for %s (%s): empty output", principalType, *principalID)

		}

		status := output.AccountAssignmentCreationStatus

		_, err = waitAccountAssignmentCreated(conn, instanceArn, aws.StringValue(status.RequestId))
		if err != nil {
			return fmt.Errorf("error waiting for SSO Account Assignment for %s (%s) to be created: %w", principalType, *principalID, err)
		}

	}
	return nil
}

func deleteAccountAssignments(conn *ssoadmin.SSOAdmin, instanceArn string, permissionSetArn string, targetType string, targetID string, principalType string, principalIDs []*string) error {

	for _, principalID := range principalIDs {

		input := &ssoadmin.DeleteAccountAssignmentInput{
			PrincipalId:      aws.String(*principalID),
			InstanceArn:      aws.String(instanceArn),
			PermissionSetArn: aws.String(permissionSetArn),
			TargetType:       aws.String(targetType),
			TargetId:         aws.String(targetID),
			PrincipalType:    aws.String(principalType),
		}

		output, err := conn.DeleteAccountAssignment(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
				return nil
			}
			return fmt.Errorf("error deleting SSO Account Assignment for Principal (%s): %w", *principalID, err)
		}

		if output == nil || output.AccountAssignmentDeletionStatus == nil {
			return fmt.Errorf("error deleting SSO Account Assignment for Principal (%s): empty output", *principalID)
		}

		status := output.AccountAssignmentDeletionStatus

		_, err = waitAccountAssignmentDeleted(conn, instanceArn, aws.StringValue(status.RequestId))
		if err != nil {
			return fmt.Errorf("error waiting for SSO Account Assignment for Principal (%s) to be deleted: %w", *principalID, err)
		}
	}

	return nil
}

func ParseAccountAssignmentsID(id string) ([]string, error) {
	idParts := strings.Split(id, ",")
	if len(idParts) != 5 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" ||
		idParts[3] == "" || idParts[4] == "" {
		return nil, fmt.Errorf("unexpected format for ID (%q), expected PRINCIPAL_TYPE,TARGET_ID,TARGET_TYPE,PERMISSION_SET_ARN,INSTANCE_ARN", id)
	}
	return idParts, nil
}
