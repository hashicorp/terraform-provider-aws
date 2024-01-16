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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.PrincipalType](),
			},
			"target_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"target_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.TargetType](),
			},
		},
	}
}

func resourceAccountAssignmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

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
		PrincipalType:    awstypes.PrincipalType(principalType),
		TargetId:         aws.String(targetID),
		TargetType:       awstypes.TargetType(targetType),
	}

	output, err := conn.CreateAccountAssignment(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSO Account Assignment for %s (%s): %s", principalType, principalID, err)
	}

	if _, err := waitAccountAssignmentCreated(ctx, conn, instanceARN, aws.ToString(output.AccountAssignmentCreationStatus.RequestId), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SSO Account Assignment for %s (%s) create: %s", principalType, principalID, err)
	}

	d.SetId(fmt.Sprintf("%s,%s,%s,%s,%s,%s", principalID, principalType, targetID, targetType, permissionSetARN, instanceARN))

	return append(diags, resourceAccountAssignmentRead(ctx, d, meta)...)
}

func resourceAccountAssignmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

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
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

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
		PrincipalType:    awstypes.PrincipalType(principalType),
		TargetId:         aws.String(targetID),
		TargetType:       awstypes.TargetType(targetType),
	}

	output, err := conn.DeleteAccountAssignment(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSO Account Assignment for Principal (%s): %s", principalID, err)
	}

	if _, err := waitAccountAssignmentDeleted(ctx, conn, instanceARN, aws.ToString(output.AccountAssignmentDeletionStatus.RequestId), d.Timeout(schema.TimeoutDelete)); err != nil {
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

func FindAccountAssignment(ctx context.Context, conn *ssoadmin.Client, principalID, principalType, accountID, permissionSetARN, instanceARN string) (*awstypes.AccountAssignment, error) {
	input := &ssoadmin.ListAccountAssignmentsInput{
		AccountId:        aws.String(accountID),
		InstanceArn:      aws.String(instanceARN),
		PermissionSetArn: aws.String(permissionSetARN),
	}
	filter := func(a awstypes.AccountAssignment) bool {
		return aws.ToString(a.PrincipalId) == principalID && string(a.PrincipalType) == principalType
	}

	return findAccountAssignment(ctx, conn, input, filter)
}

func findAccountAssignment(
	ctx context.Context,
	conn *ssoadmin.Client,
	input *ssoadmin.ListAccountAssignmentsInput,
	filter tfslices.Predicate[awstypes.AccountAssignment],
) (*awstypes.AccountAssignment, error) {
	output, err := findAccountAssignments(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAccountAssignments(
	ctx context.Context,
	conn *ssoadmin.Client,
	input *ssoadmin.ListAccountAssignmentsInput,
	filter tfslices.Predicate[awstypes.AccountAssignment],
) ([]awstypes.AccountAssignment, error) {
	var output []awstypes.AccountAssignment

	paginator := ssoadmin.NewListAccountAssignmentsPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.AccountAssignments {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findAccountAssignmentCreationStatus(
	ctx context.Context,
	conn *ssoadmin.Client,
	instanceARN,
	requestID string,
) (*awstypes.AccountAssignmentOperationStatus, error) {
	input := &ssoadmin.DescribeAccountAssignmentCreationStatusInput{
		AccountAssignmentCreationRequestId: aws.String(requestID),
		InstanceArn:                        aws.String(instanceARN),
	}

	output, err := conn.DescribeAccountAssignmentCreationStatus(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func statusAccountAssignmentCreation(ctx context.Context, conn *ssoadmin.Client, instanceARN, requestID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findAccountAssignmentCreationStatus(ctx, conn, instanceARN, requestID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func findAccountAssignmentDeletionStatus(ctx context.Context, conn *ssoadmin.Client, instanceARN, requestID string) (*awstypes.AccountAssignmentOperationStatus, error) {
	input := &ssoadmin.DescribeAccountAssignmentDeletionStatusInput{
		AccountAssignmentDeletionRequestId: aws.String(requestID),
		InstanceArn:                        aws.String(instanceARN),
	}

	output, err := conn.DescribeAccountAssignmentDeletionStatus(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func statusAccountAssignmentDeletion(ctx context.Context, conn *ssoadmin.Client, instanceARN, requestID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findAccountAssignmentDeletionStatus(ctx, conn, instanceARN, requestID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitAccountAssignmentCreated(ctx context.Context, conn *ssoadmin.Client, instanceARN, requestID string, timeout time.Duration) (*awstypes.AccountAssignmentOperationStatus, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.StatusValuesInProgress),
		Target:     enum.Slice(awstypes.StatusValuesSucceeded),
		Refresh:    statusAccountAssignmentCreation(ctx, conn, instanceARN, requestID),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AccountAssignmentOperationStatus); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.FailureReason)))

		return output, err
	}

	return nil, err
}

func waitAccountAssignmentDeleted(ctx context.Context, conn *ssoadmin.Client, instanceArn, requestID string, timeout time.Duration) (*awstypes.AccountAssignmentOperationStatus, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.StatusValuesInProgress),
		Target:     enum.Slice(awstypes.StatusValuesSucceeded),
		Refresh:    statusAccountAssignmentDeletion(ctx, conn, instanceArn, requestID),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AccountAssignmentOperationStatus); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.FailureReason)))

		return output, err
	}

	return nil, err
}
