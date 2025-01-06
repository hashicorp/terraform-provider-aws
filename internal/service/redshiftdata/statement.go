// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftdata

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata"
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshiftdata_statement")
func resourceStatement() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStatementCreate,
		ReadWithoutTimeout:   resourceStatementRead,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrClusterIdentifier: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrDatabase: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"db_user": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrParameters: {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"secret_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"sql": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"statement_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"with_event": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"workgroup_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceStatementCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftDataClient(ctx)

	input := &redshiftdata.ExecuteStatementInput{
		Database:  aws.String(d.Get(names.AttrDatabase).(string)),
		Sql:       aws.String(d.Get("sql").(string)),
		WithEvent: aws.Bool(d.Get("with_event").(bool)),
	}

	if v, ok := d.GetOk(names.AttrClusterIdentifier); ok {
		input.ClusterIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("db_user"); ok {
		input.DbUser = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrParameters); ok && len(v.([]interface{})) > 0 {
		input.Parameters = expandParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("secret_arn"); ok {
		input.SecretArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("statement_name"); ok {
		input.StatementName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("workgroup_name"); ok {
		input.WorkgroupName = aws.String(v.(string))
	}

	output, err := conn.ExecuteStatement(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "executing Redshift Data Statement: %s", err)
	}

	d.SetId(aws.ToString(output.Id))

	if _, err := waitStatementFinished(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Data Statement (%s) finish: %s", d.Id(), err)
	}

	return append(diags, resourceStatementRead(ctx, d, meta)...)
}

func resourceStatementRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftDataClient(ctx)

	sub, err := FindStatementByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Data Statement (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Data Statement (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrClusterIdentifier, sub.ClusterIdentifier)
	d.Set(names.AttrDatabase, d.Get(names.AttrDatabase).(string))
	d.Set("db_user", d.Get("db_user").(string))
	if err := d.Set(names.AttrParameters, flattenParameters(sub.QueryParameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameters: %s", err)
	}
	d.Set("secret_arn", sub.SecretArn)
	d.Set("sql", sub.QueryString)
	d.Set("workgroup_name", sub.WorkgroupName)

	return diags
}

// FindStatementByID will only find full statement info for statements created recently.
// For statements that AWS thinks are expired, FindStatementByID will just return a bare bones DescribeStatementOutput w/ only the Id present.
func FindStatementByID(ctx context.Context, conn *redshiftdata.Client, id string) (*redshiftdata.DescribeStatementOutput, error) {
	input := &redshiftdata.DescribeStatementInput{
		Id: aws.String(id),
	}

	output, err := conn.DescribeStatement(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if errs.IsAErrorMessageContains[*types.ValidationException](err, "expired") {
		return &redshiftdata.DescribeStatementOutput{
			Id:     aws.String(id),
			Status: types.StatusString("EXPIRED"),
		}, nil
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusStatement(ctx context.Context, conn *redshiftdata.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindStatementByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitStatementFinished(ctx context.Context, conn *redshiftdata.Client, id string, timeout time.Duration) (*redshiftdata.DescribeStatementOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			types.StatusStringPicked,
			types.StatusStringStarted,
			types.StatusStringSubmitted,
		),
		Target:     enum.Slice(types.StatusStringFinished),
		Refresh:    statusStatement(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshiftdata.DescribeStatementOutput); ok {
		if status := output.Status; status == types.StatusStringFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.Error)))
		}

		return output, err
	}

	return nil, err
}

func expandParameter(tfMap map[string]interface{}) *types.SqlParameter {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.SqlParameter{}

	if v, ok := tfMap[names.AttrName].(string); ok {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func expandParameters(tfList []interface{}) []types.SqlParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.SqlParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandParameter(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenParameter(apiObject types.SqlParameter) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.Value; v != nil {
		tfMap[names.AttrValue] = aws.ToString(v)
	}
	return tfMap
}

func flattenParameters(apiObjects []types.SqlParameter) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenParameter(apiObject))
	}

	return tfList
}
