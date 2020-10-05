package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsSsoAssignment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsoAssignmentCreate,
		Read:   resourceAwsSsoAssignmentRead,
		// Update: resourceAwsSsoAssignmentUpdate,
		Delete: resourceAwsSsoAssignmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"failure_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"request_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
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

	log.Printf("[INFO] Creating AWS SSO Assignment")

	instanceArn := d.Get("instance_arn").(string)
	permissionSetArn := d.Get("permission_set_arn").(string)
	principalID := d.Get("principal_id").(string)
	principalType := d.Get("principal_type").(string)
	targetID := d.Get("target_id").(string)
	targetType := d.Get("target_type").(string)

	vars := []string{
		permissionSetArn,
		targetType,
		targetID,
		principalType,
		principalID,
	}
	d.SetId(strings.Join(vars, "_"))

	req := &ssoadmin.CreateAccountAssignmentInput{
		InstanceArn:      aws.String(instanceArn),
		PermissionSetArn: aws.String(permissionSetArn),
		PrincipalId:      aws.String(principalID),
		PrincipalType:    aws.String(principalType),
		TargetId:         aws.String(targetID),
		TargetType:       aws.String(targetType),
	}

	resp, err := conn.CreateAccountAssignment(req)
	if err != nil {
		return fmt.Errorf("Error creating AWS SSO Assignment: %s", err)
	}

	status := resp.AccountAssignmentCreationStatus

	if status.CreatedDate != nil {
		d.Set("created_date", status.CreatedDate.Format(time.RFC3339))
	}
	if status.FailureReason != nil {
		d.Set("failure_reason", status.FailureReason)
	}
	if status.RequestId != nil {
		d.Set("request_id", status.RequestId)
	}
	if status.Status != nil {
		d.Set("status", status.Status)
	}

	waitResp, waitErr := waitForAssignmentCreation(d, conn, instanceArn, aws.StringValue(status.RequestId))
	if waitErr != nil {
		return fmt.Errorf("Error waiting for AWS SSO Assignment: %s", waitErr)
	}

	// IN_PROGRESS | FAILED | SUCCEEDED
	if aws.StringValue(waitResp.Status) == "FAILED" {
		return fmt.Errorf("Failed to create AWS SSO Assignment: %s", aws.StringValue(waitResp.FailureReason))
	}

	return resourceAwsSsoAssignmentRead(d, meta)
}

func resourceAwsSsoAssignmentRead(d *schema.ResourceData, meta interface{}) error {
	// conn := meta.(*AWSClient).ssoadminconn
	// TODO
	return nil
}

// func resourceAwsSsoAssignmentUpdate(d *schema.ResourceData, meta interface{}) error {
// 	// conn := meta.(*AWSClient).ssoadminconn
// 	// TODO
// 	return resourceAwsSsoAssignmentRead(d, meta)
// }

func resourceAwsSsoAssignmentDelete(d *schema.ResourceData, meta interface{}) error {
	// conn := meta.(*AWSClient).ssoadminconn
	// TODO
	return nil
}

func waitForAssignmentCreation(d *schema.ResourceData, conn *ssoadmin.SSOAdmin, instanceArn string, requestID string) (*ssoadmin.AccountAssignmentOperationStatus, error) {
	var status *ssoadmin.AccountAssignmentOperationStatus

	// TODO: timeout
	for {
		resp, err := conn.DescribeAccountAssignmentCreationStatus(&ssoadmin.DescribeAccountAssignmentCreationStatusInput{
			InstanceArn:                        aws.String(instanceArn),
			AccountAssignmentCreationRequestId: aws.String(requestID),
		})

		if err != nil {
			return nil, err
		}

		status = resp.AccountAssignmentCreationStatus

		if status.CreatedDate != nil {
			d.Set("created_date", status.CreatedDate.Format(time.RFC3339))
		}
		if status.FailureReason != nil {
			d.Set("failure_reason", status.FailureReason)
		}
		if status.RequestId != nil {
			d.Set("request_id", status.RequestId)
		}
		if status.Status != nil {
			d.Set("status", status.Status)
		}

		if aws.StringValue(status.Status) != "IN_PROGRESS" {
			break
		}

		time.Sleep(time.Second)
	}

	return status, nil
}

// func waitForAssignmentDeletion(conn *ssoadmin.SSOAdmin, instanceArn string, requestId string) error {
// }
