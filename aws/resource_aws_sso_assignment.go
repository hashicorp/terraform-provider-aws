package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
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

		// TODO
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(AWSSSOAssignmentCreateRetryTimeout),
			Delete: schema.DefaultTimeout(AWSSSOAssignmentDeleteRetryTimeout),
		},

		Schema: map[string]*schema.Schema{
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

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

			"principal_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 47),
					validation.StringMatch(regexp.MustCompile(`^([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}$`), "must match ([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}"),
				),
			},

			"principal_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"USER", "GROUP"}, false),
			},

			"target_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},

			"target_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "AWS_ACCOUNT",
				ValidateFunc: validation.StringInSlice([]string{"AWS_ACCOUNT"}, false),
			},
		},
	}
}

func resourceAwsSsoAssignmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn

	instanceArn := d.Get("instance_arn").(string)
	permissionSetArn := d.Get("permission_set_arn").(string)
	principalID := d.Get("principal_id").(string)
	principalType := d.Get("principal_type").(string)
	targetID := d.Get("target_id").(string)
	targetType := d.Get("target_type").(string)

	req := &ssoadmin.CreateAccountAssignmentInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
		PrincipalId:      aws.String(principalID),
		PrincipalType:    aws.String(principalType),
		TargetId:         aws.String(targetID),
		TargetType:       aws.String(targetType),
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

	vars := []string{
		permissionSetArn,
		targetType,
		targetID,
		principalType,
		principalID,
	}
	d.SetId(strings.Join(vars, "_"))

	if waitResp.CreatedDate != nil {
		d.Set("created_date", waitResp.CreatedDate.Format(time.RFC3339))
	}

	return resourceAwsSsoAssignmentRead(d, meta)
}

func resourceAwsSsoAssignmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn

	instanceArn := d.Get("instance_arn").(string)
	permissionSetArn := d.Get("permission_set_arn").(string)
	principalID := d.Get("principal_id").(string)
	principalType := d.Get("principal_type").(string)
	targetID := d.Get("target_id").(string)
	targetType := d.Get("target_type").(string)

	req := &ssoadmin.ListAccountAssignmentsInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
		AccountId:        aws.String(targetID),
	}

	log.Printf("[DEBUG] Reading AWS SSO Assignments for %s", req)
	resp, err := conn.ListAccountAssignments(req)
	if err != nil {
		return fmt.Errorf("Error getting AWS SSO Assignments: %s", err)
	}

	if resp == nil || len(resp.AccountAssignments) == 0 {
		log.Printf("[DEBUG] No account assignments found")
		d.SetId("")
		return nil
	}

	for _, accountAssignment := range resp.AccountAssignments {
		if aws.StringValue(accountAssignment.PrincipalType) == principalType {
			if aws.StringValue(accountAssignment.PrincipalId) == principalID {
				vars := []string{
					permissionSetArn,
					targetType,
					targetID,
					principalType,
					principalID,
				}
				d.SetId(strings.Join(vars, "_"))
				return nil
			}
		}
	}

	log.Printf("[DEBUG] Account assignment not found for %s", map[string]string{
		"PrincipalType": principalType,
		"PrincipalId":   principalID,
	})
	d.SetId("")
	return nil
}

func resourceAwsSsoAssignmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn

	instanceArn := d.Get("instance_arn").(string)
	permissionSetArn := d.Get("permission_set_arn").(string)
	principalID := d.Get("principal_id").(string)
	principalType := d.Get("principal_type").(string)
	targetID := d.Get("target_id").(string)
	targetType := d.Get("target_type").(string)

	req := &ssoadmin.DeleteAccountAssignmentInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
		PrincipalId:      aws.String(principalID),
		PrincipalType:    aws.String(principalType),
		TargetId:         aws.String(targetID),
		TargetType:       aws.String(targetType),
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
