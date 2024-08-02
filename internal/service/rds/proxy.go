// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_db_proxy", name="DB Proxy")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func resourceProxy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProxyCreate,
		ReadWithoutTimeout:   resourceProxyRead,
		UpdateWithoutTimeout: resourceProxyUpdate,
		DeleteWithoutTimeout: resourceProxyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth_scheme": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.AuthScheme](),
						},
						"client_password_auth_type": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[types.ClientPasswordAuthType](),
						},
						names.AttrDescription: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"iam_auth": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.IAMAuthMode](),
						},
						"secret_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrUsername: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Set: sdkv2.SimpleSchemaSetFunc("auth_scheme", names.AttrDescription, "iam_auth", "secret_arn", names.AttrUsername),
			},
			"debug_logging": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrEndpoint: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_family": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.EngineFamily](),
			},
			"idle_client_timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validIdentifier,
			},
			"require_tls": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceProxyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &rds.CreateDBProxyInput{
		Auth:         expandUserAuthConfigs(d.Get("auth").(*schema.Set).List()),
		DBProxyName:  aws.String(name),
		EngineFamily: types.EngineFamily(d.Get("engine_family").(string)),
		RoleArn:      aws.String(d.Get(names.AttrRoleARN).(string)),
		Tags:         getTagsInV2(ctx),
		VpcSubnetIds: flex.ExpandStringValueSet(d.Get("vpc_subnet_ids").(*schema.Set)),
	}

	if v, ok := d.GetOk("debug_logging"); ok {
		input.DebugLogging = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("idle_client_timeout"); ok {
		input.IdleClientTimeout = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("require_tls"); ok {
		input.RequireTLS = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk(names.AttrVPCSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
		input.VpcSecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	output, err := conn.CreateDBProxy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS DB Proxy (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.DBProxy.DBProxyName))

	if _, err := waitDBProxyCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Proxy (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceProxyRead(ctx, d, meta)...)
}

func resourceProxyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbProxy, err := findDBProxyByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Proxy %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Proxy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, dbProxy.DBProxyArn)
	d.Set("auth", flattenUserAuthConfigInfos(dbProxy.Auth))
	d.Set(names.AttrName, dbProxy.DBProxyName)
	d.Set("debug_logging", dbProxy.DebugLogging)
	d.Set("engine_family", dbProxy.EngineFamily)
	d.Set("idle_client_timeout", dbProxy.IdleClientTimeout)
	d.Set("require_tls", dbProxy.RequireTLS)
	d.Set(names.AttrRoleARN, dbProxy.RoleArn)
	d.Set("vpc_subnet_ids", dbProxy.VpcSubnetIds)
	d.Set(names.AttrVPCSecurityGroupIDs, dbProxy.VpcSecurityGroupIds)
	d.Set(names.AttrEndpoint, dbProxy.Endpoint)

	return diags
}

func resourceProxyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		oName, nName := d.GetChange(names.AttrName)
		input := &rds.ModifyDBProxyInput{
			Auth:           expandUserAuthConfigs(d.Get("auth").(*schema.Set).List()),
			DBProxyName:    aws.String(oName.(string)),
			DebugLogging:   aws.Bool(d.Get("debug_logging").(bool)),
			NewDBProxyName: aws.String(nName.(string)),
			RequireTLS:     aws.Bool(d.Get("require_tls").(bool)),
			RoleArn:        aws.String(d.Get(names.AttrRoleARN).(string)),
		}

		if v, ok := d.GetOk("idle_client_timeout"); ok {
			input.IdleClientTimeout = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk(names.AttrVPCSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
			input.SecurityGroups = flex.ExpandStringValueSet(v.(*schema.Set))
		}

		_, err := conn.ModifyDBProxy(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RDS DB Proxy (%s): %s", d.Id(), err)
		}

		// DB Proxy Name is used as an ID as the API doesn't provide a way to read/
		// update/delete DB proxies using the ARN
		d.SetId(nName.(string))

		if _, err := waitDBProxyUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Proxy (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceProxyRead(ctx, d, meta)...)
}

func resourceProxyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	log.Printf("[DEBUG] Deleting RDS DB Proxy: %s", d.Id())
	_, err := conn.DeleteDBProxy(ctx, &rds.DeleteDBProxyInput{
		DBProxyName: aws.String(d.Id()),
	})

	if errs.IsA[*types.DBProxyNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Proxy (%s): %s", d.Id(), err)
	}

	if _, err := waitDBProxyDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Proxy (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findDBProxyByName(ctx context.Context, conn *rds.Client, name string) (*types.DBProxy, error) {
	input := &rds.DescribeDBProxiesInput{
		DBProxyName: aws.String(name),
	}
	output, err := findDBProxy(ctx, conn, input, tfslices.PredicateTrue[*types.DBProxy]())

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.DBProxyName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findDBProxy(ctx context.Context, conn *rds.Client, input *rds.DescribeDBProxiesInput, filter tfslices.Predicate[*types.DBProxy]) (*types.DBProxy, error) {
	output, err := findDBProxies(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBProxies(ctx context.Context, conn *rds.Client, input *rds.DescribeDBProxiesInput, filter tfslices.Predicate[*types.DBProxy]) ([]types.DBProxy, error) {
	var output []types.DBProxy

	pages := rds.NewDescribeDBProxiesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.DBProxyNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBProxies {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusDBProxy(ctx context.Context, conn *rds.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDBProxyByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitDBProxyCreated(ctx context.Context, conn *rds.Client, name string, timeout time.Duration) (*types.DBProxy, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.DBProxyStatusCreating),
		Target:  enum.Slice(types.DBProxyStatusAvailable),
		Refresh: statusDBProxy(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBProxy); ok {
		return output, err
	}

	return nil, err
}

func waitDBProxyDeleted(ctx context.Context, conn *rds.Client, name string, timeout time.Duration) (*types.DBProxy, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.DBProxyStatusDeleting),
		Target:  []string{},
		Refresh: statusDBProxy(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBProxy); ok {
		return output, err
	}

	return nil, err
}

func waitDBProxyUpdated(ctx context.Context, conn *rds.Client, name string, timeout time.Duration) (*types.DBProxy, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.DBProxyStatusModifying),
		Target:  enum.Slice(types.DBProxyStatusAvailable),
		Refresh: statusDBProxy(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DBProxy); ok {
		return output, err
	}

	return nil, err
}

func expandUserAuthConfigs(tfList []interface{}) []types.UserAuthConfig {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make([]types.UserAuthConfig, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := types.UserAuthConfig{}

		if v, ok := tfMap["auth_scheme"].(string); ok && v != "" {
			apiObject.AuthScheme = types.AuthScheme(v)
		}

		if v, ok := tfMap["client_password_auth_type"].(string); ok && v != "" {
			apiObject.ClientPasswordAuthType = types.ClientPasswordAuthType(v)
		}

		if v, ok := tfMap[names.AttrDescription].(string); ok && v != "" {
			apiObject.Description = aws.String(v)
		}

		if v, ok := tfMap["iam_auth"].(string); ok && v != "" {
			apiObject.IAMAuth = types.IAMAuthMode(v)
		}

		if v, ok := tfMap["secret_arn"].(string); ok && v != "" {
			apiObject.SecretArn = aws.String(v)
		}

		if v, ok := tfMap[names.AttrUsername].(string); ok && v != "" {
			apiObject.UserName = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenUserAuthConfigInfo(apiObject types.UserAuthConfigInfo) map[string]interface{} {
	tfMap := map[string]interface{}{
		"auth_scheme":               apiObject.AuthScheme,
		"client_password_auth_type": apiObject.ClientPasswordAuthType,
		"iam_auth":                  apiObject.IAMAuth,
	}

	if v := apiObject.Description; v != nil {
		tfMap[names.AttrDescription] = aws.ToString(v)
	}

	if v := apiObject.SecretArn; v != nil {
		tfMap["secret_arn"] = aws.ToString(v)
	}

	if v := apiObject.UserName; v != nil {
		tfMap[names.AttrUsername] = aws.ToString(v)
	}

	return tfMap
}

func flattenUserAuthConfigInfos(apiObjects []types.UserAuthConfigInfo) []interface{} {
	tfList := []interface{}{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenUserAuthConfigInfo(apiObject))
	}

	return tfList
}
