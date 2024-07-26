// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package athena

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_athena_database")
func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDatabaseCreate,
		ReadWithoutTimeout:   resourceDatabaseRead,
		UpdateWithoutTimeout: schema.NoopContext, // force_destroy isn't ForceNew.
		DeleteWithoutTimeout: resourceDatabaseDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"acl_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_acl_option": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[types.S3AclOption](),
						},
					},
				},
			},
			names.AttrBucket: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrComment: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrEncryptionConfiguration: {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption_option": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[types.EncryptionOption](),
						},
						names.AttrKMSKey: {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			names.AttrExpectedBucketOwner: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrForceDestroy: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile("^[0-9a-z_]+$"), "must be lowercase letters, numbers, or underscore ('_')"),
			},
			names.AttrProperties: {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	name := d.Get(names.AttrName).(string)
	createStmt := fmt.Sprintf("create database `%s`", name)
	var queryString bytes.Buffer
	queryString.WriteString(createStmt)

	if v, ok := d.GetOk(names.AttrComment); ok && v.(string) != "" {
		commentStmt := fmt.Sprintf(" comment '%s'", strings.Replace(v.(string), "'", "\\'", -1))
		queryString.WriteString(commentStmt)
	}

	if v, ok := d.GetOk(names.AttrProperties); ok && len(v.(map[string]interface{})) > 0 {
		var props []string
		for k, v := range v.(map[string]interface{}) {
			prop := fmt.Sprintf(" '%[1]s' = '%[2]s' ", k, v.(string))
			props = append(props, prop)
		}

		propStmt := fmt.Sprintf(" WITH DBPROPERTIES(%s)", strings.Join(props, ","))
		queryString.WriteString(propStmt)
	}

	queryString.WriteString(";")

	input := &athena.StartQueryExecutionInput{
		QueryString:         aws.String(queryString.String()),
		ResultConfiguration: expandResultConfiguration(d),
	}

	output, err := conn.StartQueryExecution(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Athena Database (%s): %s", name, err)
	}

	if err := executeAndExpectNoRows(ctx, conn, aws.ToString(output.QueryExecutionId)); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Athena Database (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceDatabaseRead(ctx, d, meta)...)
}

func resourceDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	db, err := findDatabaseByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Athena Database (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Athena Database (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrComment, db.Description)
	d.Set(names.AttrName, db.Name)
	d.Set(names.AttrProperties, db.Parameters)

	return diags
}

func resourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AthenaClient(ctx)

	queryString := fmt.Sprintf("drop database `%s`", d.Id())
	if d.Get(names.AttrForceDestroy).(bool) {
		queryString += " cascade"
	}
	queryString += ";"

	input := &athena.StartQueryExecutionInput{
		QueryString:         aws.String(queryString),
		ResultConfiguration: expandResultConfiguration(d),
	}

	output, err := conn.StartQueryExecution(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Athena Database (%s): %s", d.Id(), err)
	}

	if err := executeAndExpectNoRows(ctx, conn, aws.ToString(output.QueryExecutionId)); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Athena Database (%s): %s", d.Id(), err)
	}

	return diags
}

func findDatabaseByName(ctx context.Context, conn *athena.Client, name string) (*types.Database, error) {
	input := &athena.GetDatabaseInput{
		CatalogName:  aws.String("AwsDataCatalog"),
		DatabaseName: aws.String(name),
	}

	output, err := conn.GetDatabase(ctx, input)

	if errs.IsAErrorMessageContains[*types.MetadataException](err, "not found") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Database == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Database, nil
}

func expandResultConfiguration(d *schema.ResourceData) *types.ResultConfiguration {
	resultConfig := &types.ResultConfiguration{
		OutputLocation:          aws.String("s3://" + d.Get(names.AttrBucket).(string)),
		EncryptionConfiguration: expandResultConfigurationEncryptionConfig(d.Get(names.AttrEncryptionConfiguration).([]interface{})),
	}

	if v, ok := d.GetOk(names.AttrExpectedBucketOwner); ok {
		resultConfig.ExpectedBucketOwner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("acl_configuration"); ok && len(v.([]interface{})) > 0 {
		resultConfig.AclConfiguration = expandResultConfigurationACLConfig(v.([]interface{}))
	}

	return resultConfig
}

func expandResultConfigurationEncryptionConfig(config []interface{}) *types.EncryptionConfiguration {
	if len(config) == 0 {
		return nil
	}

	data := config[0].(map[string]interface{})

	encryptionConfig := &types.EncryptionConfiguration{
		EncryptionOption: types.EncryptionOption(data["encryption_option"].(string)),
	}

	if v, ok := data[names.AttrKMSKey].(string); ok && v != "" {
		encryptionConfig.KmsKey = aws.String(v)
	}

	return encryptionConfig
}

func expandResultConfigurationACLConfig(config []interface{}) *types.AclConfiguration {
	if len(config) == 0 {
		return nil
	}

	data := config[0].(map[string]interface{})

	encryptionConfig := &types.AclConfiguration{
		S3AclOption: types.S3AclOption(data["s3_acl_option"].(string)),
	}

	return encryptionConfig
}

func executeAndExpectNoRows(ctx context.Context, conn *athena.Client, qeid string) error {
	rs, err := queryExecutionResult(ctx, conn, qeid)
	if err != nil {
		return err
	}
	if len(rs.Rows) != 0 {
		return fmt.Errorf("unexpected query result: %s", flattenResultSet(rs))
	}
	return nil
}

func queryExecutionResult(ctx context.Context, conn *athena.Client, qeid string) (*types.ResultSet, error) {
	executionStateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.QueryExecutionStateQueued, types.QueryExecutionStateRunning),
		Target:     enum.Slice(types.QueryExecutionStateSucceeded),
		Refresh:    queryExecutionStateRefreshFunc(ctx, conn, qeid),
		Timeout:    10 * time.Minute,
		Delay:      3 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := executionStateConf.WaitForStateContext(ctx)

	if err != nil {
		return nil, err
	}

	qrinput := &athena.GetQueryResultsInput{
		QueryExecutionId: aws.String(qeid),
	}
	resp, err := conn.GetQueryResults(ctx, qrinput)
	if err != nil {
		return nil, err
	}
	return resp.ResultSet, nil
}

func queryExecutionStateRefreshFunc(ctx context.Context, conn *athena.Client, qeid string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &athena.GetQueryExecutionInput{
			QueryExecutionId: aws.String(qeid),
		}
		out, err := conn.GetQueryExecution(ctx, input)
		if err != nil {
			return nil, "failed", err
		}

		if out == nil || out.QueryExecution == nil || out.QueryExecution.Status == nil {
			return nil, "", nil
		}

		status := out.QueryExecution.Status

		if status.State == types.QueryExecutionStateFailed && status.StateChangeReason != nil {
			err = fmt.Errorf("reason: %s", aws.ToString(status.StateChangeReason))
		}

		return out, string(out.QueryExecution.Status.State), err
	}
}

func flattenResultSet(rs *types.ResultSet) string {
	ss := make([]string, 0)
	for _, row := range rs.Rows {
		for _, datum := range row.Data {
			ss = append(ss, aws.ToString(datum.VarCharValue))
		}
	}
	return strings.Join(ss, "\n")
}
