// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_ssoadmin_account_assignment")
func ResourceAccountAssignment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccountAssignmentCreate,
		ReadWithoutTimeout:   resourceAccountAssignmentRead,
		DeleteWithoutTimeout: resourceAccountAssignmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
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
			"principal_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 47),
					validation.StringMatch(regexache.MustCompile(`^([0-9a-f]{10}-|)[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}$`), "must match ([0-9a-f]{10}-|)[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}"),
				),
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

func resourceAccountAssignmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn(ctx)

	instanceARN := d.Get("instance_arn").(string)
	permissionSetARN := d.Get("permission_set_arn").(string)
	principalID := d.Get("principal_id").(string)
	principalType := d.Get("principal_type").(string)
	targetID := d.Get("target_id").(string)
	targetType := d.Get("target_type").(string)

	// We need to check if the assignment exists before creating it since the AWS SSO API doesn't prevent us from creating duplicates.
	_, err := FindAccountAssignment(ctx, conn, principalID, principalType, targetID, permissionSetARN, instanceARN)

	if err == nil {
		return sdkdiag.AppendErrorf(diags, "creating SSO Account Assignment for %s (%s): already exists", principalType, principalID)
	} else if !tfresource.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "listing SSO Account Assignments for Account ID (%s) Permission Set (%s): %s", targetID, permissionSetARN, err)
	}

	input := &ssoadmin.CreateAccountAssignmentInput{
		InstanceArn:      aws.String(instanceARN),
		PermissionSetArn: aws.String(permissionSetARN),
		PrincipalId:      aws.String(principalID),
		PrincipalType:    aws.String(principalType),
		TargetId:         aws.String(targetID),
		TargetType:       aws.String(targetType),
	}

	output, err := conn.CreateAccountAssignmentWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSO Account Assignment for %s (%s): %s", principalType, principalID, err)
	}

	if _, err := waitAccountAssignmentCreated(ctx, conn, instanceARN, aws.StringValue(output.AccountAssignmentCreationStatus.RequestId), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SSO Account Assignment for %s (%s) create: %s", principalType, principalID, err)
	}

	d.SetId(fmt.Sprintf("%s,%s,%s,%s,%s,%s", principalID, principalType, targetID, targetType, permissionSetARN, instanceARN))

	return append(diags, resourceAccountAssignmentRead(ctx, d, meta)...)
}

func resourceAccountAssignmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn(ctx)

	idParts, err := ParseAccountAssignmentID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	principalID := idParts[0]
	principalType := idParts[1]
	targetID := idParts[2]
	targetType := idParts[3]
	permissionSetARN := idParts[4]
	instanceARN := idParts[5]

	accountAssignment, err := FindAccountAssignment(ctx, conn, principalID, principalType, targetID, permissionSetARN, instanceARN)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSO Account Assignment for Principal (%s) not found, removing from state", principalID)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSO Account Assignment for Principal (%s): %s", principalID, err)
	}

	d.Set("instance_arn", instanceARN)
	d.Set("permission_set_arn", accountAssignment.PermissionSetArn)
	d.Set("principal_id", accountAssignment.PrincipalId)
	d.Set("principal_type", accountAssignment.PrincipalType)
	d.Set("target_id", accountAssignment.AccountId)
	d.Set("target_type", targetType)

	return diags
}

func resourceAccountAssignmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminConn(ctx)

	idParts, err := ParseAccountAssignmentID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	principalID := idParts[0]
	principalType := idParts[1]
	targetID := idParts[2]
	targetType := idParts[3]
	permissionSetARN := idParts[4]
	instanceARN := idParts[5]

	input := &ssoadmin.DeleteAccountAssignmentInput{
		InstanceArn:      aws.String(instanceARN),
		PermissionSetArn: aws.String(permissionSetARN),
		PrincipalId:      aws.String(principalID),
		PrincipalType:    aws.String(principalType),
		TargetType:       aws.String(targetType),
		TargetId:         aws.String(targetID),
	}

	output, err := conn.DeleteAccountAssignmentWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSO Account Assignment for Principal (%s): %s", principalID, err)
	}

	if _, err := waitAccountAssignmentDeleted(ctx, conn, instanceARN, aws.StringValue(output.AccountAssignmentDeletionStatus.RequestId), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SSO Account Assignment for Principal (%s) delete: %s", principalID, err)
	}

	return diags
}

func ParseAccountAssignmentID(id string) ([]string, error) {
	idParts := strings.Split(id, ",")
	if len(idParts) != 6 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" ||
		idParts[3] == "" || idParts[4] == "" || idParts[5] == "" {
		return nil, fmt.Errorf("unexpected format for ID (%q), expected PRINCIPAL_ID,PRINCIPAL_TYPE,TARGET_ID,TARGET_TYPE,PERMISSION_SET_ARN,INSTANCE_ARN", id)
	}
	return idParts, nil
}

func FindAccountAssignment(ctx context.Context, conn *ssoadmin.SSOAdmin, principalID, principalType, accountID, permissionSetARN, instanceARN string) (*ssoadmin.AccountAssignment, error) {
	input := &ssoadmin.ListAccountAssignmentsInput{
		AccountId:        aws.String(accountID),
		InstanceArn:      aws.String(instanceARN),
		PermissionSetArn: aws.String(permissionSetARN),
	}
	filter := func(a *ssoadmin.AccountAssignment) bool {
		return aws.StringValue(a.PrincipalId) == principalID && aws.StringValue(a.PrincipalType) == principalType
	}

	return findAccountAssignment(ctx, conn, input, filter)
}

func findAccountAssignment(ctx context.Context, conn *ssoadmin.SSOAdmin, input *ssoadmin.ListAccountAssignmentsInput, filter tfslices.Predicate[*ssoadmin.AccountAssignment]) (*ssoadmin.AccountAssignment, error) {
	output, err := findAccountAssignments(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findAccountAssignments(ctx context.Context, conn *ssoadmin.SSOAdmin, input *ssoadmin.ListAccountAssignmentsInput, filter tfslices.Predicate[*ssoadmin.AccountAssignment]) ([]*ssoadmin.AccountAssignment, error) {
	var output []*ssoadmin.AccountAssignment

	err := conn.ListAccountAssignmentsPagesWithContext(ctx, input, func(page *ssoadmin.ListAccountAssignmentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.AccountAssignments {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findAccountAssignmentCreationStatus(ctx context.Context, conn *ssoadmin.SSOAdmin, instanceARN, requestID string) (*ssoadmin.AccountAssignmentOperationStatus, error) {
	input := &ssoadmin.DescribeAccountAssignmentCreationStatusInput{
		AccountAssignmentCreationRequestId: aws.String(requestID),
		InstanceArn:                        aws.String(instanceARN),
	}

	output, err := conn.DescribeAccountAssignmentCreationStatusWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AccountAssignmentCreationStatus == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AccountAssignmentCreationStatus, nil
}

func statusAccountAssignmentCreation(ctx context.Context, conn *ssoadmin.SSOAdmin, instanceARN, requestID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findAccountAssignmentCreationStatus(ctx, conn, instanceARN, requestID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func findAccountAssignmentDeletionStatus(ctx context.Context, conn *ssoadmin.SSOAdmin, instanceARN, requestID string) (*ssoadmin.AccountAssignmentOperationStatus, error) {
	input := &ssoadmin.DescribeAccountAssignmentDeletionStatusInput{
		AccountAssignmentDeletionRequestId: aws.String(requestID),
		InstanceArn:                        aws.String(instanceARN),
	}

	output, err := conn.DescribeAccountAssignmentDeletionStatusWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AccountAssignmentDeletionStatus == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AccountAssignmentDeletionStatus, nil
}

func statusAccountAssignmentDeletion(ctx context.Context, conn *ssoadmin.SSOAdmin, instanceARN, requestID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findAccountAssignmentDeletionStatus(ctx, conn, instanceARN, requestID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitAccountAssignmentCreated(ctx context.Context, conn *ssoadmin.SSOAdmin, instanceARN, requestID string, timeout time.Duration) (*ssoadmin.AccountAssignmentOperationStatus, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ssoadmin.StatusValuesInProgress},
		Target:     []string{ssoadmin.StatusValuesSucceeded},
		Refresh:    statusAccountAssignmentCreation(ctx, conn, instanceARN, requestID),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ssoadmin.AccountAssignmentOperationStatus); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureReason)))

		return output, err
	}

	return nil, err
}

func waitAccountAssignmentDeleted(ctx context.Context, conn *ssoadmin.SSOAdmin, instanceArn, requestID string, timeout time.Duration) (*ssoadmin.AccountAssignmentOperationStatus, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ssoadmin.StatusValuesInProgress},
		Target:     []string{ssoadmin.StatusValuesSucceeded},
		Refresh:    statusAccountAssignmentDeletion(ctx, conn, instanceArn, requestID),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ssoadmin.AccountAssignmentOperationStatus); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureReason)))

		return output, err
	}

	return nil, err
}
