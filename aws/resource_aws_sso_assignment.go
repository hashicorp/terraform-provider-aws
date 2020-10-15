package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	AWSSSOAssignmentCreateRetryTimeout = 5 * time.Minute
	AWSSSOAssignmentDeleteRetryTimeout = 5 * time.Minute
	AWSSSOAssignmentRetryDelay         = 5 * time.Second
	AWSSSOAssignmentRetryMinTimeout    = 3 * time.Second
)

func resourceAwsSsoAssignment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsoAssignmentCreate,
		Read:   resourceAwsSsoAssignmentRead,
		Delete: resourceAwsSsoAssignmentDelete,

		Importer: &schema.ResourceImporter{
			State: resourceAwsSsoAssignmentImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(AWSSSOAssignmentCreateRetryTimeout),
			Delete: schema.DefaultTimeout(AWSSSOAssignmentDeleteRetryTimeout),
		},

		Schema: map[string]*schema.Schema{
			"instance_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(10, 1224),
					validation.StringMatch(regexp.MustCompile(`^arn:aws:sso:::instance/(sso)?ins-[a-zA-Z0-9-.]{16}$`), "must match arn:aws:sso:::instance/(sso)?ins-[a-zA-Z0-9-.]{16}"),
				),
			},

			"permission_set_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(10, 1224),
					validation.StringMatch(regexp.MustCompile(`^arn:aws:sso:::permissionSet/(sso)?ins-[a-zA-Z0-9-.]{16}/ps-[a-zA-Z0-9-./]{16}$`), "must match arn:aws:sso:::permissionSet/(sso)?ins-[a-zA-Z0-9-.]{16}/ps-[a-zA-Z0-9-./]{16}"),
				),
			},

			"target_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ssoadmin.TargetTypeAwsAccount,
				ValidateFunc: validation.StringInSlice([]string{ssoadmin.TargetTypeAwsAccount}, false),
			},

			"target_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},

			"principal_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{ssoadmin.PrincipalTypeUser, ssoadmin.PrincipalTypeGroup}, false),
			},

			"principal_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 47),
					validation.StringMatch(regexp.MustCompile(`^([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}$`), "must match ([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}"),
				),
			},

			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsSsoAssignmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn

	instanceArn := d.Get("instance_arn").(string)
	permissionSetArn := d.Get("permission_set_arn").(string)
	targetType := d.Get("target_type").(string)
	targetID := d.Get("target_id").(string)
	principalType := d.Get("principal_type").(string)
	principalID := d.Get("principal_id").(string)

	id, idErr := resourceAwsSsoAssignmentID(instanceArn, permissionSetArn, targetType, targetID, principalType, principalID)
	if idErr != nil {
		return idErr
	}

	// We need to check if the assignment exists before creating it
	// since the AWS SSO API doesn't prevent us from creating duplicates
	accountAssignment, getAccountAssignmentErr := resourceAwsSsoAssignmentGet(
		conn,
		instanceArn,
		permissionSetArn,
		targetType,
		targetID,
		principalType,
		principalID,
	)
	if getAccountAssignmentErr != nil {
		return getAccountAssignmentErr
	}
	if accountAssignment != nil {
		return fmt.Errorf("AWS SSO Assignment already exists. Import the resource by calling: terraform import <resource_address> %s", id)
	}

	req := &ssoadmin.CreateAccountAssignmentInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
		TargetType:       aws.String(targetType),
		TargetId:         aws.String(targetID),
		PrincipalType:    aws.String(principalType),
		PrincipalId:      aws.String(principalID),
	}

	log.Printf("[INFO] Creating AWS SSO Assignment")
	resp, err := conn.CreateAccountAssignment(req)
	if err != nil {
		return fmt.Errorf("Error creating AWS SSO Assignment: %s", err)
	}

	status := resp.AccountAssignmentCreationStatus

	if status.CreatedDate != nil {
		d.Set("created_date", status.CreatedDate.Format(time.RFC3339))
	}

	waitResp, waitErr := waitForAssignmentCreation(conn, instanceArn, aws.StringValue(status.RequestId), d.Timeout(schema.TimeoutCreate))
	if waitErr != nil {
		return waitErr
	}

	d.SetId(id)

	if waitResp.CreatedDate != nil {
		d.Set("created_date", waitResp.CreatedDate.Format(time.RFC3339))
	}

	return resourceAwsSsoAssignmentRead(d, meta)
}

func resourceAwsSsoAssignmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn

	instanceArn := d.Get("instance_arn").(string)
	permissionSetArn := d.Get("permission_set_arn").(string)
	targetType := d.Get("target_type").(string)
	targetID := d.Get("target_id").(string)
	principalType := d.Get("principal_type").(string)
	principalID := d.Get("principal_id").(string)

	accountAssignment, err := resourceAwsSsoAssignmentGet(
		conn,
		instanceArn,
		permissionSetArn,
		targetType,
		targetID,
		principalType,
		principalID,
	)
	if err != nil {
		return err
	}
	if accountAssignment == nil {
		log.Printf("[DEBUG] Account assignment not found for %s", map[string]string{
			"PrincipalType": principalType,
			"PrincipalId":   principalID,
		})
		d.SetId("")
		return nil
	}

	id, idErr := resourceAwsSsoAssignmentID(instanceArn, permissionSetArn, targetType, targetID, principalType, principalID)
	if idErr != nil {
		return idErr
	}
	d.SetId(id)
	return nil
}

func resourceAwsSsoAssignmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn

	instanceArn := d.Get("instance_arn").(string)
	permissionSetArn := d.Get("permission_set_arn").(string)
	targetType := d.Get("target_type").(string)
	targetID := d.Get("target_id").(string)
	principalType := d.Get("principal_type").(string)
	principalID := d.Get("principal_id").(string)

	req := &ssoadmin.DeleteAccountAssignmentInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
		TargetType:       aws.String(targetType),
		TargetId:         aws.String(targetID),
		PrincipalType:    aws.String(principalType),
		PrincipalId:      aws.String(principalID),
	}

	log.Printf("[INFO] Deleting AWS SSO Assignment")
	resp, err := conn.DeleteAccountAssignment(req)
	if err != nil {
		aerr, ok := err.(awserr.Error)
		if ok && aerr.Code() == ssoadmin.ErrCodeResourceNotFoundException {
			log.Printf("[DEBUG] AWS SSO Assignment not found")
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error deleting AWS SSO Assignment: %s", err)
	}

	status := resp.AccountAssignmentDeletionStatus

	_, waitErr := waitForAssignmentDeletion(conn, instanceArn, aws.StringValue(status.RequestId), d.Timeout(schema.TimeoutDelete))
	if waitErr != nil {
		return waitErr
	}

	d.SetId("")
	return nil
}

func resourceAwsSsoAssignmentImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// id = ${InstanceID}/${PermissionSetID}/${TargetType}/${TargetID}/${PrincipalType}/${PrincipalID}
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 6 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" || idParts[4] == "" || idParts[5] == "" {
		return nil, fmt.Errorf("Unexpected format of id (%s), expected ${InstanceID}/${PermissionSetID}/${TargetType}/${TargetID}/${PrincipalType}/${PrincipalID}", d.Id())
	}

	instanceID := idParts[0]
	permissionSetID := idParts[1]
	targetType := idParts[2]
	targetID := idParts[3]
	principalType := idParts[4]
	principalID := idParts[5]

	var err error

	// arn:${Partition}:sso:::instance/${InstanceId}
	instanceArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "sso",
		Resource:  fmt.Sprintf("instance/%s", instanceID),
	}.String()

	// arn:${Partition}:sso:::permissionSet/${InstanceId}/${PermissionSetId}
	permissionSetArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "sso",
		Resource:  fmt.Sprintf("permissionSet/%s/%s", instanceID, permissionSetID),
	}.String()

	err = d.Set("instance_arn", instanceArn)
	if err != nil {
		return nil, err
	}
	err = d.Set("permission_set_arn", permissionSetArn)
	if err != nil {
		return nil, err
	}
	err = d.Set("target_type", targetType)
	if err != nil {
		return nil, err
	}
	err = d.Set("target_id", targetID)
	if err != nil {
		return nil, err
	}
	err = d.Set("principal_type", principalType)
	if err != nil {
		return nil, err
	}
	err = d.Set("principal_id", principalID)
	if err != nil {
		return nil, err
	}

	id, idErr := resourceAwsSsoAssignmentID(instanceArn, permissionSetArn, targetType, targetID, principalType, principalID)
	if idErr != nil {
		return nil, idErr
	}
	d.SetId(id)

	return []*schema.ResourceData{d}, nil
}

func resourceAwsSsoAssignmentID(
	instanceArn string,
	permissionSetArn string,
	targetType string,
	targetID string,
	principalType string,
	principalID string,
) (string, error) {
	// arn:${Partition}:sso:::instance/${InstanceId}
	iArn, err := arn.Parse(instanceArn)
	if err != nil {
		return "", err
	}
	iArnResourceParts := strings.Split(iArn.Resource, "/")
	if len(iArnResourceParts) != 2 || iArnResourceParts[0] != "instance" || iArnResourceParts[1] == "" {
		return "", fmt.Errorf("Unexpected format of ARN (%s), expected arn:${Partition}:sso:::instance/${InstanceId}", instanceArn)
	}
	instanceID := iArnResourceParts[1]

	// arn:${Partition}:sso:::permissionSet/${InstanceId}/${PermissionSetId}
	pArn, err := arn.Parse(permissionSetArn)
	if err != nil {
		return "", err
	}
	pArnResourceParts := strings.Split(pArn.Resource, "/")
	if len(pArnResourceParts) != 3 || pArnResourceParts[0] != "permissionSet" || pArnResourceParts[1] == "" || pArnResourceParts[2] == "" {
		return "", fmt.Errorf("Unexpected format of ARN (%s), expected arn:${Partition}:sso:::permissionSet/${InstanceId}/${PermissionSetId}", permissionSetArn)
	}
	permissionSetID := pArnResourceParts[2]

	vars := []string{
		instanceID,
		permissionSetID,
		targetType,
		targetID,
		principalType,
		principalID,
	}
	return strings.Join(vars, "/"), nil
}

func resourceAwsSsoAssignmentGet(
	conn *ssoadmin.SSOAdmin,
	instanceArn string,
	permissionSetArn string,
	targetType string,
	targetID string,
	principalType string,
	principalID string,
) (*ssoadmin.AccountAssignment, error) {
	if targetType != ssoadmin.TargetTypeAwsAccount {
		return nil, fmt.Errorf("Invalid AWS SSO Assignments Target type %s. Only %s is supported", targetType, ssoadmin.TargetTypeAwsAccount)
	}

	req := &ssoadmin.ListAccountAssignmentsInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
		AccountId:        aws.String(targetID),
	}

	log.Printf("[DEBUG] Reading AWS SSO Assignments for %s", req)
	resp, err := conn.ListAccountAssignments(req)
	if err != nil {
		return nil, fmt.Errorf("Error getting AWS SSO Assignments: %s", err)
	}

	if resp == nil || len(resp.AccountAssignments) == 0 {
		log.Printf("[DEBUG] No account assignments found")
		return nil, nil
	}

	for _, accountAssignment := range resp.AccountAssignments {
		if aws.StringValue(accountAssignment.PrincipalType) == principalType {
			if aws.StringValue(accountAssignment.PrincipalId) == principalID {
				return accountAssignment, nil
			}
		}
	}

	// not found
	return nil, nil
}

func waitForAssignmentCreation(conn *ssoadmin.SSOAdmin, instanceArn string, requestID string, timeout time.Duration) (*ssoadmin.AccountAssignmentOperationStatus, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ssoadmin.StatusValuesInProgress},
		Target:  []string{ssoadmin.StatusValuesSucceeded},
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeAccountAssignmentCreationStatus(&ssoadmin.DescribeAccountAssignmentCreationStatusInput{
				InstanceArn:                        aws.String(instanceArn),
				AccountAssignmentCreationRequestId: aws.String(requestID),
			})
			if err != nil {
				return resp, "", fmt.Errorf("Error describing account assignment creation status: %s", err)
			}
			status := resp.AccountAssignmentCreationStatus
			return status, aws.StringValue(status.Status), nil
		},
		Timeout:    timeout,
		Delay:      AWSSSOAssignmentRetryDelay,
		MinTimeout: AWSSSOAssignmentRetryMinTimeout,
	}
	status, err := stateConf.WaitForState()
	if err != nil {
		return nil, fmt.Errorf("Error waiting for account assignment to be created: %s", err)
	}
	return status.(*ssoadmin.AccountAssignmentOperationStatus), nil
}

func waitForAssignmentDeletion(conn *ssoadmin.SSOAdmin, instanceArn string, requestID string, timeout time.Duration) (*ssoadmin.AccountAssignmentOperationStatus, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ssoadmin.StatusValuesInProgress},
		Target:  []string{ssoadmin.StatusValuesSucceeded},
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeAccountAssignmentDeletionStatus(&ssoadmin.DescribeAccountAssignmentDeletionStatusInput{
				InstanceArn:                        aws.String(instanceArn),
				AccountAssignmentDeletionRequestId: aws.String(requestID),
			})
			if err != nil {
				return resp, "", fmt.Errorf("Error describing account assignment deletion status: %s", err)
			}
			status := resp.AccountAssignmentDeletionStatus
			return status, aws.StringValue(status.Status), nil
		},
		Timeout:    timeout,
		Delay:      AWSSSOAssignmentRetryDelay,
		MinTimeout: AWSSSOAssignmentRetryMinTimeout,
	}
	status, err := stateConf.WaitForState()
	if err != nil {
		return nil, fmt.Errorf("Error waiting for account assignment to be deleted: %s", err)
	}
	return status.(*ssoadmin.AccountAssignmentOperationStatus), nil
}
