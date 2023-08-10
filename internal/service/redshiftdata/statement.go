// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftdata

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftdataapiservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_redshiftdata_statement")
func ResourceStatement() *schema.Resource {
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
			"cluster_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"database": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"db_user": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"parameters": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"value": {
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
	conn := meta.(*conns.AWSClient).RedshiftDataConn(ctx)

	input := &redshiftdataapiservice.ExecuteStatementInput{
		Database:  aws.String(d.Get("database").(string)),
		Sql:       aws.String(d.Get("sql").(string)),
		WithEvent: aws.Bool(d.Get("with_event").(bool)),
	}

	if v, ok := d.GetOk("cluster_identifier"); ok {
		input.ClusterIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("db_user"); ok {
		input.DbUser = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameters"); ok && len(v.([]interface{})) > 0 {
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

	output, err := conn.ExecuteStatementWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "executing Redshift Data Statement: %s", err)
	}

	d.SetId(aws.StringValue(output.Id))

	if _, err := waitStatementFinished(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Redshift Data Statement (%s) to finish: %s", d.Id(), err)
	}

	return append(diags, resourceStatementRead(ctx, d, meta)...)
}

func resourceStatementRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftDataConn(ctx)

	sub, err := FindStatementByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Data Statement (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Data Statement (%s): %s", d.Id(), err)
	}

	d.Set("cluster_identifier", sub.ClusterIdentifier)
	d.Set("secret_arn", sub.SecretArn)
	d.Set("database", d.Get("database").(string))
	d.Set("db_user", d.Get("db_user").(string))
	if err := d.Set("parameters", flattenParameters(sub.QueryParameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameters: %s", err)
	}
	d.Set("sql", sub.QueryString)
	d.Set("workgroup_name", sub.WorkgroupName)

	return diags
}

// FindStatementByID will only find full statement info for statements created recently.
// For statements that AWS thinks are expired, FindStatementByID will just return a bare bones DescribeStatementOutput w/ only the Id present.
func FindStatementByID(ctx context.Context, conn *redshiftdataapiservice.RedshiftDataAPIService, id string) (*redshiftdataapiservice.DescribeStatementOutput, error) {
	input := &redshiftdataapiservice.DescribeStatementInput{
		Id: aws.String(id),
	}

	output, err := conn.DescribeStatementWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, redshiftdataapiservice.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if tfawserr.ErrCodeEquals(err, redshiftdataapiservice.ErrCodeValidationException) && strings.Contains(err.Error(), "expired") {
		return &redshiftdataapiservice.DescribeStatementOutput{
			Id:     aws.String(id),
			Status: aws.String("EXPIRED"),
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

func statusStatement(ctx context.Context, conn *redshiftdataapiservice.RedshiftDataAPIService, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindStatementByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitStatementFinished(ctx context.Context, conn *redshiftdataapiservice.RedshiftDataAPIService, id string, timeout time.Duration) (*redshiftdataapiservice.DescribeStatementOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			redshiftdataapiservice.StatusStringPicked,
			redshiftdataapiservice.StatusStringStarted,
			redshiftdataapiservice.StatusStringSubmitted,
		},
		Target:     []string{redshiftdataapiservice.StatusStringFinished},
		Refresh:    statusStatement(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshiftdataapiservice.DescribeStatementOutput); ok {
		if status := aws.StringValue(output.Status); status == redshiftdataapiservice.StatusStringFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.Error)))
		}

		return output, err
	}

	return nil, err
}

func expandParameter(tfMap map[string]interface{}) *redshiftdataapiservice.SqlParameter {
	if tfMap == nil {
		return nil
	}

	apiObject := &redshiftdataapiservice.SqlParameter{}

	if v, ok := tfMap["name"].(string); ok {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["value"].(string); ok {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func expandParameters(tfList []interface{}) []*redshiftdataapiservice.SqlParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*redshiftdataapiservice.SqlParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandParameter(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenParameter(apiObject *redshiftdataapiservice.SqlParameter) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	if v := apiObject.Value; v != nil {
		tfMap["value"] = aws.StringValue(v)
	}
	return tfMap
}

func flattenParameters(apiObjects []*redshiftdataapiservice.SqlParameter) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenParameter(apiObject))
	}

	return tfList
}
