// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	quicksightschema "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// Allow IAM role to become visible to the index
	propagationTimeout = 2 * time.Minute

	// accessDeniedExceptionMessage describes the error returned when the IAM role has not yet propagated
	accessDeniedExceptionAssumeRoleMessage              = "Failed to assume your role. Verify the trust relationships of the role in the IAM console"
	accessDeniedExceptionInsufficientPermissionsMessage = "Insufficient permission to access the manifest file"
)

// @SDKResource("aws_quicksight_data_source", name="Data Source")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/quicksight/types;awstypes;awstypes.DataSource")
// @Testing(skipEmptyTags=true, skipNullTags=true)
func resourceDataSource() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDataSourceCreate,
		ReadWithoutTimeout:   resourceDataSourceRead,
		UpdateWithoutTimeout: resourceDataSourceUpdate,
		DeleteWithoutTimeout: resourceDataSourceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrAWSAccountID: {
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     true,
					ForceNew:     true,
					ValidateFunc: verify.ValidAccountID,
				},
				"credentials": quicksightschema.DataSourceCredentialsSchema(),
				"data_source_id": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.All(
						validation.NoZeroValues,
						validation.StringLenBetween(1, 128),
					),
				},
				names.AttrParameters: quicksightschema.DataSourceParametersSchema(),
				"permission":         quicksightschema.PermissionsSchema(),
				"ssl_properties":     quicksightschema.SSLPropertiesSchema(),
				names.AttrTags:       tftags.TagsSchema(),
				names.AttrTagsAll:    tftags.TagsSchemaComputed(),
				names.AttrType: {
					Type:             schema.TypeString,
					Required:         true,
					ForceNew:         true,
					ValidateDiagFunc: enum.Validate[awstypes.DataSourceType](),
				},
				"vpc_connection_properties": quicksightschema.VPCConnectionPropertiesSchema(),
			}
		},
	}
}

func resourceDataSourceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID(ctx)
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountID = v.(string)
	}
	dataSourceID := d.Get("data_source_id").(string)
	id := dataSourceCreateResourceID(awsAccountID, dataSourceID)
	input := &quicksight.CreateDataSourceInput{
		AwsAccountId:         aws.String(awsAccountID),
		DataSourceId:         aws.String(dataSourceID),
		DataSourceParameters: quicksightschema.ExpandDataSourceParameters(d.Get(names.AttrParameters).([]any)),
		Name:                 aws.String(d.Get(names.AttrName).(string)),
		Tags:                 getTagsIn(ctx),
		Type:                 awstypes.DataSourceType(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk("credentials"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Credentials = quicksightschema.ExpandDataSourceCredentials(v.([]any))
	}

	if v, ok := d.GetOk("permission"); ok && v.(*schema.Set).Len() != 0 {
		input.Permissions = quicksightschema.ExpandResourcePermissions(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("ssl_properties"); ok && len(v.([]any)) != 0 && v.([]any)[0] != nil {
		input.SslProperties = quicksightschema.ExpandSSLProperties(v.([]any))
	}

	if v, ok := d.GetOk("vpc_connection_properties"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.VpcConnectionProperties = quicksightschema.ExpandVPCConnectionProperties(v.([]any))
	}

	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (any, error) {
			return conn.CreateDataSource(ctx, input)
		},
		func(err error) (bool, error) {
			var accessDeniedException *awstypes.AccessDeniedException

			if errors.As(err, &accessDeniedException) && (strings.Contains(accessDeniedException.ErrorMessage(), accessDeniedExceptionAssumeRoleMessage) || strings.Contains(accessDeniedException.ErrorMessage(), accessDeniedExceptionInsufficientPermissionsMessage)) {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating QuickSight Data Source (%s): %s", id, err)
	}

	if outputRaw == nil {
		return sdkdiag.AppendErrorf(diags, "creating QuickSight Data Source (%s): empty output", id)
	}

	d.SetId(id)

	if _, err := waitDataSourceCreated(ctx, conn, awsAccountID, dataSourceID); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting from QuickSight Data Source (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceDataSourceRead(ctx, d, meta)...)
}

func resourceDataSourceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, dataSourceID, err := dataSourceParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	dataSource, err := findDataSourceByTwoPartKey(ctx, conn, awsAccountID, dataSourceID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] QuickSight Data Source (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Data Source (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, dataSource.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountID)
	d.Set("data_source_id", dataSource.DataSourceId)
	d.Set(names.AttrName, dataSource.Name)
	if err := d.Set(names.AttrParameters, quicksightschema.FlattenDataSourceParameters(dataSource.DataSourceParameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameters: %s", err)
	}
	if err := d.Set("ssl_properties", quicksightschema.FlattenSSLProperties(dataSource.SslProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ssl_properties: %s", err)
	}
	d.Set(names.AttrType, dataSource.Type)
	if err := d.Set("vpc_connection_properties", quicksightschema.FlattenVPCConnectionProperties(dataSource.VpcConnectionProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc_connection_properties: %s", err)
	}

	permissions, err := findDataSourcePermissionsByTwoPartKey(ctx, conn, awsAccountID, dataSourceID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Data Source (%s) permissions: %s", d.Id(), err)
	}

	if err := d.Set("permission", quicksightschema.FlattenPermissions(permissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting permission: %s", err)
	}

	return diags
}

func resourceDataSourceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, dataSourceID, err := dataSourceParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept("permission", names.AttrTags, names.AttrTagsAll) {
		input := &quicksight.UpdateDataSourceInput{
			AwsAccountId: aws.String(awsAccountID),
			DataSourceId: aws.String(dataSourceID),
			Name:         aws.String(d.Get(names.AttrName).(string)),
		}

		if v, ok := d.GetOk("credentials"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.Credentials = quicksightschema.ExpandDataSourceCredentials(v.([]any))
		}

		if v, ok := d.GetOk(names.AttrParameters); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.DataSourceParameters = quicksightschema.ExpandDataSourceParameters(v.([]any))
		}

		if v, ok := d.GetOk("ssl_properties"); ok && len(v.([]any)) != 0 && v.([]any)[0] != nil {
			input.SslProperties = quicksightschema.ExpandSSLProperties(v.([]any))
		}

		if v, ok := d.GetOk("vpc_connection_properties"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.VpcConnectionProperties = quicksightschema.ExpandVPCConnectionProperties(v.([]any))
		}

		outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
			func() (any, error) {
				return conn.UpdateDataSource(ctx, input)
			},
			func(err error) (bool, error) {
				var accessDeniedException *awstypes.AccessDeniedException

				if errors.As(err, &accessDeniedException) && (strings.Contains(accessDeniedException.ErrorMessage(), accessDeniedExceptionAssumeRoleMessage) || strings.Contains(accessDeniedException.ErrorMessage(), accessDeniedExceptionInsufficientPermissionsMessage)) {
					return true, err
				}

				return false, err
			},
		)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Data Source (%s): %s", d.Id(), err)
		}

		if outputRaw == nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Data Source (%s): empty output", d.Id())
		}

		if _, err := waitDataSourceUpdated(ctx, conn, awsAccountID, dataSourceID); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for QuickSight Data Source (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("permission") {
		o, n := d.GetChange("permission")
		os, ns := o.(*schema.Set), n.(*schema.Set)
		toGrant, toRevoke := quicksightschema.DiffPermissions(os.List(), ns.List())

		input := &quicksight.UpdateDataSourcePermissionsInput{
			AwsAccountId: aws.String(awsAccountID),
			DataSourceId: aws.String(dataSourceID),
		}

		if len(toGrant) > 0 {
			input.GrantPermissions = toGrant
		}

		if len(toRevoke) > 0 {
			input.RevokePermissions = toRevoke
		}

		_, err = conn.UpdateDataSourcePermissions(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Data Source (%s) permissions: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDataSourceRead(ctx, d, meta)...)
}

func resourceDataSourceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID, dataSourceID, err := dataSourceParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting QuickSight Data Source: %s", d.Id())
	_, err = conn.DeleteDataSource(ctx, &quicksight.DeleteDataSourceInput{
		AwsAccountId: aws.String(awsAccountID),
		DataSourceId: aws.String(dataSourceID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting QuickSight Data Source (%s): %s", d.Id(), err)
	}

	return diags
}

const dataSourceResourceIDSeparator = "/"

func dataSourceCreateResourceID(awsAccountID, dataSourceID string) string {
	parts := []string{awsAccountID, dataSourceID}
	id := strings.Join(parts, dataSourceResourceIDSeparator)

	return id
}

func dataSourceParseResourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, dataSourceResourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected AWS_ACCOUNT_ID%[2]sDATA_SOURCE_ID", id, dataSourceResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

func findDataSourceByTwoPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, dataSourceID string) (*awstypes.DataSource, error) {
	input := &quicksight.DescribeDataSourceInput{
		AwsAccountId: aws.String(awsAccountID),
		DataSourceId: aws.String(dataSourceID),
	}

	return findDataSource(ctx, conn, input)
}

func findDataSource(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeDataSourceInput) (*awstypes.DataSource, error) {
	output, err := conn.DescribeDataSource(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DataSource == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DataSource, nil
}

func findDataSourcePermissionsByTwoPartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, dataSourceID string) ([]awstypes.ResourcePermission, error) {
	input := &quicksight.DescribeDataSourcePermissionsInput{
		AwsAccountId: aws.String(awsAccountID),
		DataSourceId: aws.String(dataSourceID),
	}

	return findDataSourcePermissions(ctx, conn, input)
}

func findDataSourcePermissions(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeDataSourcePermissionsInput) ([]awstypes.ResourcePermission, error) {
	output, err := conn.DescribeDataSourcePermissions(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Permissions, nil
}

func statusDataSource(ctx context.Context, conn *quicksight.Client, awsAccountID, dataSourceID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findDataSourceByTwoPartKey(ctx, conn, awsAccountID, dataSourceID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitDataSourceCreated(ctx context.Context, conn *quicksight.Client, awsAccountID, dataSourceID string) (*awstypes.DataSource, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceStatusCreationInProgress),
		Target:  enum.Slice(awstypes.ResourceStatusCreationSuccessful),
		Refresh: statusDataSource(ctx, conn, awsAccountID, dataSourceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DataSource); ok {
		if status, errorInfo := output.Status, output.ErrorInfo; status == awstypes.ResourceStatusCreationFailed {
			tfresource.SetLastError(err, dataSourceError(errorInfo))
		}

		return output, err
	}

	return nil, err
}

func waitDataSourceUpdated(ctx context.Context, conn *quicksight.Client, awsAccountID, dataSourceID string) (*awstypes.DataSource, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceStatusUpdateInProgress),
		Target:  enum.Slice(awstypes.ResourceStatusUpdateSuccessful),
		Refresh: statusDataSource(ctx, conn, awsAccountID, dataSourceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DataSource); ok {
		if status, errorInfo := output.Status, output.ErrorInfo; status == awstypes.ResourceStatusUpdateFailed {
			tfresource.SetLastError(err, dataSourceError(errorInfo))
		}

		return output, err
	}

	return nil, err
}

func dataSourceError(apiObject *awstypes.DataSourceErrorInfo) error {
	if apiObject == nil {
		return nil
	}

	return fmt.Errorf("%s: %s", apiObject.Type, aws.ToString(apiObject.Message))
}
