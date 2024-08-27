// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_quicksight_data_source", name="Data Source")
// @Tags(identifierAttribute="arn")
func ResourceDataSource() *schema.Resource {
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

				"credentials": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"copy_source_arn": {
								Type:          schema.TypeString,
								Optional:      true,
								ValidateFunc:  verify.ValidARN,
								ConflictsWith: []string{"credentials.0.credential_pair"},
							},
							"credential_pair": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrPassword: {
											Type:     schema.TypeString,
											Required: true,
											ValidateFunc: validation.All(
												validation.NoZeroValues,
												validation.StringLenBetween(1, 1024),
											),
											Sensitive: true,
										},
										names.AttrUsername: {
											Type:     schema.TypeString,
											Required: true,
											ValidateFunc: validation.All(
												validation.NoZeroValues,
												validation.StringLenBetween(1, 64),
											),
											Sensitive: true,
										},
									},
								},
								ConflictsWith: []string{"credentials.0.copy_source_arn"},
							},
						},
					},
				},

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

				names.AttrParameters: {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					MinItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"amazon_elasticsearch": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrDomain: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
									},
								},
							},
							"athena": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"work_group": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.NoZeroValues,
										},
									},
								},
							},
							"aurora": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrDatabase: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										"host": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										names.AttrPort: {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IntAtLeast(1),
										},
									},
								},
							},
							"aurora_postgresql": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrDatabase: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										"host": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										names.AttrPort: {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IntAtLeast(1),
										},
									},
								},
							},
							"aws_iot_analytics": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"data_set_name": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
									},
								},
							},
							"jira": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"site_base_url": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
									},
								},
							},
							"maria_db": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrDatabase: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										"host": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										names.AttrPort: {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IntAtLeast(1),
										},
									},
								},
							},
							"mysql": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrDatabase: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										"host": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										names.AttrPort: {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IntAtLeast(1),
										},
									},
								},
							},
							"oracle": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrDatabase: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										"host": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										names.AttrPort: {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IntAtLeast(1),
										},
									},
								},
							},
							"postgresql": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrDatabase: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										"host": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										names.AttrPort: {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IntAtLeast(1),
										},
									},
								},
							},
							"presto": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"catalog": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										"host": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										names.AttrPort: {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IntAtLeast(1),
										},
									},
								},
							},
							"rds": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrDatabase: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										names.AttrInstanceID: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
									},
								},
							},
							"redshift": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"cluster_id": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										names.AttrDatabase: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										"host": {
											Type:     schema.TypeString,
											Optional: true,
										},
										names.AttrPort: {
											Type:     schema.TypeInt,
											Optional: true,
										},
									},
								},
							},
							"s3": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"manifest_file_location": {
											Type:     schema.TypeList,
											Required: true,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrBucket: {
														Type:         schema.TypeString,
														Required:     true,
														ValidateFunc: validation.NoZeroValues,
													},
													names.AttrKey: {
														Type:         schema.TypeString,
														Required:     true,
														ValidateFunc: validation.NoZeroValues,
													},
												},
											},
										},
									},
								},
							},
							"service_now": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"site_base_url": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
									},
								},
							},
							"snowflake": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrDatabase: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										"host": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										"warehouse": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
									},
								},
							},
							"spark": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"host": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										names.AttrPort: {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IntAtLeast(1),
										},
									},
								},
							},
							"sql_server": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrDatabase: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										"host": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										names.AttrPort: {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IntAtLeast(1),
										},
									},
								},
							},
							"teradata": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrDatabase: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										"host": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										names.AttrPort: {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IntAtLeast(1),
										},
									},
								},
							},
							"twitter": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"max_rows": {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IntAtLeast(1),
										},
										"query": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
									},
								},
							},
						},
					},
				},

				"permission": {
					Type:     schema.TypeSet,
					Optional: true,
					MinItems: 1,
					MaxItems: 64,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrActions: {
								Type:     schema.TypeSet,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
								MinItems: 1,
								MaxItems: 16,
							},
							names.AttrPrincipal: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				},

				"ssl_properties": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"disable_ssl": {
								Type:     schema.TypeBool,
								Required: true,
							},
						},
					},
				},

				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),

				names.AttrType: {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringInSlice(quicksight.DataSourceType_Values(), false),
				},

				"vpc_connection_properties": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"vpc_connection_arn": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				},
			}
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDataSourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId := meta.(*conns.AWSClient).AccountID
	id := d.Get("data_source_id").(string)

	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountId = v.(string)
	}

	params := &quicksight.CreateDataSourceInput{
		AwsAccountId:         aws.String(awsAccountId),
		DataSourceId:         aws.String(id),
		DataSourceParameters: expandDataSourceParameters(d.Get(names.AttrParameters).([]interface{})),
		Name:                 aws.String(d.Get(names.AttrName).(string)),
		Tags:                 getTagsIn(ctx),
		Type:                 aws.String(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk("credentials"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.Credentials = expandDataSourceCredentials(v.([]interface{}))
	}

	if v, ok := d.GetOk("permission"); ok && v.(*schema.Set).Len() > 0 {
		params.Permissions = expandResourcePermissions(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("ssl_properties"); ok && len(v.([]interface{})) != 0 && v.([]interface{})[0] != nil {
		params.SslProperties = expandDataSourceSSLProperties(v.([]interface{}))
	}

	if v, ok := d.GetOk("vpc_connection_properties"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		params.VpcConnectionProperties = expandDataSourceVPCConnectionProperties(v.([]interface{}))
	}

	_, err := conn.CreateDataSourceWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating QuickSight Data Source: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s", awsAccountId, id))

	if _, err := waitCreated(ctx, conn, awsAccountId, id); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting from QuickSight Data Source (%s) creation: %s", d.Id(), err)
	}

	return append(diags, resourceDataSourceRead(ctx, d, meta)...)
}

func resourceDataSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId, dataSourceId, err := ParseDataSourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	descOpts := &quicksight.DescribeDataSourceInput{
		AwsAccountId: aws.String(awsAccountId),
		DataSourceId: aws.String(dataSourceId),
	}

	output, err := conn.DescribeDataSourceWithContext(ctx, descOpts)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] QuickSight Data Source (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing QuickSight Data Source (%s): %s", d.Id(), err)
	}

	if output == nil || output.DataSource == nil {
		return sdkdiag.AppendErrorf(diags, "describing QuickSight Data Source (%s): empty output", d.Id())
	}

	dataSource := output.DataSource

	d.Set(names.AttrARN, dataSource.Arn)
	d.Set(names.AttrAWSAccountID, awsAccountId)
	d.Set("data_source_id", dataSource.DataSourceId)
	d.Set(names.AttrName, dataSource.Name)

	if err := d.Set(names.AttrParameters, flattenParameters(dataSource.DataSourceParameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameters: %s", err)
	}

	if err := d.Set("ssl_properties", flattenSSLProperties(dataSource.SslProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ssl_properties: %s", err)
	}

	d.Set(names.AttrType, dataSource.Type)

	if err := d.Set("vpc_connection_properties", flattenVPCConnectionProperties(dataSource.VpcConnectionProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc_connection_properties: %s", err)
	}

	permsResp, err := conn.DescribeDataSourcePermissionsWithContext(ctx, &quicksight.DescribeDataSourcePermissionsInput{
		AwsAccountId: aws.String(awsAccountId),
		DataSourceId: aws.String(dataSourceId),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing QuickSight Data Source (%s) Permissions: %s", d.Id(), err)
	}

	if err := d.Set("permission", flattenPermissions(permsResp.Permissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting permission: %s", err)
	}

	return diags
}

func resourceDataSourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	if d.HasChangesExcept("permission", names.AttrTags, names.AttrTagsAll) {
		awsAccountId, dataSourceId, err := ParseDataSourceID(d.Id())
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		params := &quicksight.UpdateDataSourceInput{
			AwsAccountId: aws.String(awsAccountId),
			DataSourceId: aws.String(dataSourceId),
			Name:         aws.String(d.Get(names.AttrName).(string)),
		}

		if v, ok := d.GetOk(names.AttrParameters); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			params.DataSourceParameters = expandDataSourceParameters(v.([]interface{}))
		}

		if v, ok := d.GetOk("credentials"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			params.Credentials = expandDataSourceCredentials(v.([]interface{}))
		}

		if v, ok := d.GetOk("ssl_properties"); ok && len(v.([]interface{})) != 0 && v.([]interface{})[0] != nil {
			params.SslProperties = expandDataSourceSSLProperties(v.([]interface{}))
		}

		if v, ok := d.GetOk("vpc_connection_properties"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			params.VpcConnectionProperties = expandDataSourceVPCConnectionProperties(v.([]interface{}))
		}

		_, err = conn.UpdateDataSourceWithContext(ctx, params)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Data Source (%s): %s", d.Id(), err)
		}

		if _, err := waitUpdated(ctx, conn, awsAccountId, dataSourceId); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for QuickSight Data Source (%s) to update: %s", d.Id(), err)
		}
	}

	if d.HasChange("permission") {
		awsAccountId, dataSourceId, err := ParseDataSourceID(d.Id())
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		oraw, nraw := d.GetChange("permission")
		o := oraw.(*schema.Set).List()
		n := nraw.(*schema.Set).List()

		toGrant, toRevoke := DiffPermissions(o, n)

		params := &quicksight.UpdateDataSourcePermissionsInput{
			AwsAccountId: aws.String(awsAccountId),
			DataSourceId: aws.String(dataSourceId),
		}

		if len(toGrant) > 0 {
			params.GrantPermissions = toGrant
		}

		if len(toRevoke) > 0 {
			params.RevokePermissions = toRevoke
		}

		_, err = conn.UpdateDataSourcePermissionsWithContext(ctx, params)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Data Source (%s) permissions: %s", dataSourceId, err)
		}
	}

	return append(diags, resourceDataSourceRead(ctx, d, meta)...)
}

func resourceDataSourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).QuickSightConn(ctx)

	awsAccountId, dataSourceId, err := ParseDataSourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	deleteOpts := &quicksight.DeleteDataSourceInput{
		AwsAccountId: aws.String(awsAccountId),
		DataSourceId: aws.String(dataSourceId),
	}

	_, err = conn.DeleteDataSourceWithContext(ctx, deleteOpts)

	if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting QuickSight Data Source (%s): %s", d.Id(), err)
	}

	return diags
}

func expandDataSourceCredentials(tfList []interface{}) *quicksight.DataSourceCredentials {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	credentials := &quicksight.DataSourceCredentials{}

	if v, ok := tfMap["copy_source_arn"].(string); ok && v != "" {
		credentials.CopySourceArn = aws.String(v)
	}

	if v, ok := tfMap["credential_pair"].([]interface{}); ok && len(v) > 0 {
		credentials.CredentialPair = expandDataSourceCredentialPair(v)
	}

	return credentials
}

func expandDataSourceCredentialPair(tfList []interface{}) *quicksight.CredentialPair {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	credentialPair := &quicksight.CredentialPair{}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	if v, ok := tfMap[names.AttrUsername].(string); ok && v != "" {
		credentialPair.Username = aws.String(v)
	}

	if v, ok := tfMap[names.AttrPassword].(string); ok && v != "" {
		credentialPair.Password = aws.String(v)
	}

	return credentialPair
}

func expandDataSourceParameters(tfList []interface{}) *quicksight.DataSourceParameters {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	dataSourceParams := &quicksight.DataSourceParameters{}

	if v, ok := tfMap["amazon_elasticsearch"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &quicksight.AmazonElasticsearchParameters{}

			if v, ok := m[names.AttrDomain].(string); ok && v != "" {
				ps.Domain = aws.String(v)
			}

			dataSourceParams.AmazonElasticsearchParameters = ps
		}
	}

	if v := tfMap["athena"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &quicksight.AthenaParameters{}
			if v, ok := m["work_group"].(string); ok && v != "" {
				ps.WorkGroup = aws.String(v)
			}

			dataSourceParams.AthenaParameters = ps
		}
	}

	if v := tfMap["aurora"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &quicksight.AuroraParameters{}
			if v, ok := m[names.AttrDatabase].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m[names.AttrPort].(int); ok {
				ps.Port = aws.Int64(int64(v))
			}

			dataSourceParams.AuroraParameters = ps
		}
	}

	if v, ok := tfMap["aurora_postgresql"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &quicksight.AuroraPostgreSqlParameters{}
			if v, ok := m[names.AttrDatabase].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m[names.AttrPort].(int); ok {
				ps.Port = aws.Int64(int64(v))
			}

			dataSourceParams.AuroraPostgreSqlParameters = ps
		}
	}

	if v := tfMap["aws_iot_analytics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &quicksight.AwsIotAnalyticsParameters{}
			if v, ok := m["data_set_name"].(string); ok && v != "" {
				ps.DataSetName = aws.String(v)
			}

			dataSourceParams.AwsIotAnalyticsParameters = ps
		}
	}

	if v := tfMap["jira"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &quicksight.JiraParameters{}
			if v, ok := m["site_base_url"].(string); ok && v != "" {
				ps.SiteBaseUrl = aws.String(v)
			}

			dataSourceParams.JiraParameters = ps
		}
	}

	if v := tfMap["maria_db"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &quicksight.MariaDbParameters{}
			if v, ok := m[names.AttrDatabase].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m[names.AttrPort].(int); ok {
				ps.Port = aws.Int64(int64(v))
			}

			dataSourceParams.MariaDbParameters = ps
		}
	}

	if v := tfMap["mysql"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &quicksight.MySqlParameters{}
			if v, ok := m[names.AttrDatabase].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m[names.AttrPort].(int); ok {
				ps.Port = aws.Int64(int64(v))
			}

			dataSourceParams.MySqlParameters = ps
		}
	}

	if v := tfMap["oracle"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &quicksight.OracleParameters{}
			if v, ok := m[names.AttrDatabase].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m[names.AttrPort].(int); ok {
				ps.Port = aws.Int64(int64(v))
			}

			dataSourceParams.OracleParameters = ps
		}
	}

	if v := tfMap["postgresql"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &quicksight.PostgreSqlParameters{}
			if v, ok := m[names.AttrDatabase].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m[names.AttrPort].(int); ok {
				ps.Port = aws.Int64(int64(v))
			}

			dataSourceParams.PostgreSqlParameters = ps
		}
	}
	if v := tfMap["presto"].([]interface{}); ok && len(v) > 0 && v != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &quicksight.PrestoParameters{}
			if v, ok := m["catalog"].(string); ok && v != "" {
				ps.Catalog = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m[names.AttrPort].(int); ok {
				ps.Port = aws.Int64(int64(v))
			}

			dataSourceParams.PrestoParameters = ps
		}
	}

	if v := tfMap["rds"].([]interface{}); ok && len(v) > 0 && v != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &quicksight.RdsParameters{}

			if v, ok := m[names.AttrDatabase].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m[names.AttrInstanceID].(string); ok && v != "" {
				ps.InstanceId = aws.String(v)
			}

			dataSourceParams.RdsParameters = ps
		}
	}

	if v := tfMap["redshift"].([]interface{}); ok && len(v) > 0 && v != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &quicksight.RedshiftParameters{}
			if v, ok := m["cluster_id"].(string); ok && v != "" {
				ps.ClusterId = aws.String(v)
			}
			if v, ok := m[names.AttrDatabase].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m[names.AttrPort].(int); ok {
				ps.Port = aws.Int64(int64(v))
			}

			dataSourceParams.RedshiftParameters = ps
		}
	}

	if v := tfMap["s3"].([]interface{}); ok && len(v) > 0 && v != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &quicksight.S3Parameters{}
			if v, ok := m["manifest_file_location"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
				lm, ok := v[0].(map[string]interface{})
				if ok {
					loc := &quicksight.ManifestFileLocation{}

					if v, ok := lm[names.AttrBucket].(string); ok && v != "" {
						loc.Bucket = aws.String(v)
					}

					if v, ok := lm[names.AttrKey].(string); ok && v != "" {
						loc.Key = aws.String(v)
					}

					ps.ManifestFileLocation = loc
				}
			}

			dataSourceParams.S3Parameters = ps
		}
	}

	if v := tfMap["service_now"].([]interface{}); ok && len(v) > 0 && v != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &quicksight.ServiceNowParameters{}
			if v, ok := m["site_base_url"].(string); ok && v != "" {
				ps.SiteBaseUrl = aws.String(v)
			}

			dataSourceParams.ServiceNowParameters = ps
		}
	}

	if v := tfMap["snowflake"].([]interface{}); ok && len(v) > 0 && v != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &quicksight.SnowflakeParameters{}

			if v, ok := m[names.AttrDatabase].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m["warehouse"].(string); ok && v != "" {
				ps.Warehouse = aws.String(v)
			}

			dataSourceParams.SnowflakeParameters = ps
		}
	}

	if v := tfMap["spark"].([]interface{}); ok && len(v) > 0 && v != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &quicksight.SparkParameters{}

			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m[names.AttrPort].(int); ok {
				ps.Port = aws.Int64(int64(v))
			}

			dataSourceParams.SparkParameters = ps
		}
	}

	if v := tfMap["sql_server"].([]interface{}); ok && len(v) > 0 && v != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &quicksight.SqlServerParameters{}

			if v, ok := m[names.AttrDatabase].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m[names.AttrPort].(int); ok {
				ps.Port = aws.Int64(int64(v))
			}

			dataSourceParams.SqlServerParameters = ps
		}
	}

	if v := tfMap["teradata"].([]interface{}); ok && len(v) > 0 && v != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &quicksight.TeradataParameters{}

			if v, ok := m[names.AttrDatabase].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m[names.AttrPort].(int); ok {
				ps.Port = aws.Int64(int64(v))
			}

			dataSourceParams.TeradataParameters = ps
		}
	}

	if v := tfMap["twitter"].([]interface{}); ok && len(v) > 0 && v != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &quicksight.TwitterParameters{}

			if v, ok := m["max_rows"].(int); ok {
				ps.MaxRows = aws.Int64(int64(v))
			}
			if v, ok := m["query"].(string); ok && v != "" {
				ps.Query = aws.String(v)
			}

			dataSourceParams.TwitterParameters = ps
		}
	}

	return dataSourceParams
}

func expandDataSourceSSLProperties(tfList []interface{}) *quicksight.SslProperties {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	props := &quicksight.SslProperties{}

	if v, ok := tfMap["disable_ssl"].(bool); ok {
		props.DisableSsl = aws.Bool(v)
	}

	return props
}

func expandDataSourceVPCConnectionProperties(tfList []interface{}) *quicksight.VpcConnectionProperties {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	props := &quicksight.VpcConnectionProperties{}

	if v, ok := tfMap["vpc_connection_arn"].(string); ok && v != "" {
		props.VpcConnectionArn = aws.String(v)
	}

	return props
}

func flattenParameters(parameters *quicksight.DataSourceParameters) []interface{} {
	if parameters == nil {
		return []interface{}{}
	}

	var params []interface{}

	if parameters.AmazonElasticsearchParameters != nil {
		params = append(params, map[string]interface{}{
			"amazon_elasticsearch": []interface{}{
				map[string]interface{}{
					names.AttrDomain: parameters.AmazonElasticsearchParameters.Domain,
				},
			},
		})
	}

	if parameters.AthenaParameters != nil {
		params = append(params, map[string]interface{}{
			"athena": []interface{}{
				map[string]interface{}{
					"work_group": parameters.AthenaParameters.WorkGroup,
				},
			},
		})
	}

	if parameters.AuroraParameters != nil {
		params = append(params, map[string]interface{}{
			"aurora": []interface{}{
				map[string]interface{}{
					names.AttrDatabase: parameters.AuroraParameters.Database,
					"host":             parameters.AuroraParameters.Host,
					names.AttrPort:     parameters.AuroraParameters.Port,
				},
			},
		})
	}

	if parameters.AuroraPostgreSqlParameters != nil {
		params = append(params, map[string]interface{}{
			"aurora_postgresql": []interface{}{
				map[string]interface{}{
					names.AttrDatabase: parameters.AuroraPostgreSqlParameters.Database,
					"host":             parameters.AuroraPostgreSqlParameters.Host,
					names.AttrPort:     parameters.AuroraPostgreSqlParameters.Port,
				},
			},
		})
	}

	if parameters.AwsIotAnalyticsParameters != nil {
		params = append(params, map[string]interface{}{
			"aws_iot_analytics": []interface{}{
				map[string]interface{}{
					"data_set_name": parameters.AwsIotAnalyticsParameters.DataSetName,
				},
			},
		})
	}

	if parameters.JiraParameters != nil {
		params = append(params, map[string]interface{}{
			"jira": []interface{}{
				map[string]interface{}{
					"site_base_url": parameters.JiraParameters.SiteBaseUrl,
				},
			},
		})
	}

	if parameters.MariaDbParameters != nil {
		params = append(params, map[string]interface{}{
			"maria_db": []interface{}{
				map[string]interface{}{
					names.AttrDatabase: parameters.MariaDbParameters.Database,
					"host":             parameters.MariaDbParameters.Host,
					names.AttrPort:     parameters.MariaDbParameters.Port,
				},
			},
		})
	}

	if parameters.MySqlParameters != nil {
		params = append(params, map[string]interface{}{
			"mysql": []interface{}{
				map[string]interface{}{
					names.AttrDatabase: parameters.MySqlParameters.Database,
					"host":             parameters.MySqlParameters.Host,
					names.AttrPort:     parameters.MySqlParameters.Port,
				},
			},
		})
	}

	if parameters.OracleParameters != nil {
		params = append(params, map[string]interface{}{
			"oracle": []interface{}{
				map[string]interface{}{
					names.AttrDatabase: parameters.OracleParameters.Database,
					"host":             parameters.OracleParameters.Host,
					names.AttrPort:     parameters.OracleParameters.Port,
				},
			},
		})
	}

	if parameters.PostgreSqlParameters != nil {
		params = append(params, map[string]interface{}{
			"postgresql": []interface{}{
				map[string]interface{}{
					names.AttrDatabase: parameters.PostgreSqlParameters.Database,
					"host":             parameters.PostgreSqlParameters.Host,
					names.AttrPort:     parameters.PostgreSqlParameters.Port,
				},
			},
		})
	}

	if parameters.PrestoParameters != nil {
		params = append(params, map[string]interface{}{
			"presto": []interface{}{
				map[string]interface{}{
					"catalog":      parameters.PrestoParameters.Catalog,
					"host":         parameters.PrestoParameters.Host,
					names.AttrPort: parameters.PrestoParameters.Port,
				},
			},
		})
	}

	if parameters.RdsParameters != nil {
		params = append(params, map[string]interface{}{
			"rds": []interface{}{
				map[string]interface{}{
					names.AttrDatabase:   parameters.RdsParameters.Database,
					names.AttrInstanceID: parameters.RdsParameters.InstanceId,
				},
			},
		})
	}

	if parameters.RedshiftParameters != nil {
		params = append(params, map[string]interface{}{
			"redshift": []interface{}{
				map[string]interface{}{
					"cluster_id":       parameters.RedshiftParameters.ClusterId,
					names.AttrDatabase: parameters.RedshiftParameters.Database,
					"host":             parameters.RedshiftParameters.Host,
					names.AttrPort:     parameters.RedshiftParameters.Port,
				},
			},
		})
	}

	if parameters.S3Parameters != nil {
		params = append(params, map[string]interface{}{
			"s3": []interface{}{
				map[string]interface{}{
					"manifest_file_location": []interface{}{
						map[string]interface{}{
							names.AttrBucket: parameters.S3Parameters.ManifestFileLocation.Bucket,
							names.AttrKey:    parameters.S3Parameters.ManifestFileLocation.Key,
						},
					},
				},
			},
		})
	}

	if parameters.ServiceNowParameters != nil {
		params = append(params, map[string]interface{}{
			"service_now": []interface{}{
				map[string]interface{}{
					"site_base_url": parameters.ServiceNowParameters.SiteBaseUrl,
				},
			},
		})
	}

	if parameters.SnowflakeParameters != nil {
		params = append(params, map[string]interface{}{
			"snowflake": []interface{}{
				map[string]interface{}{
					names.AttrDatabase: parameters.SnowflakeParameters.Database,
					"host":             parameters.SnowflakeParameters.Host,
					"warehouse":        parameters.SnowflakeParameters.Warehouse,
				},
			},
		})
	}

	if parameters.SparkParameters != nil {
		params = append(params, map[string]interface{}{
			"spark": []interface{}{
				map[string]interface{}{
					"host":         parameters.SparkParameters.Host,
					names.AttrPort: parameters.SparkParameters.Port,
				},
			},
		})
	}

	if parameters.SqlServerParameters != nil {
		params = append(params, map[string]interface{}{
			"sql_server": []interface{}{
				map[string]interface{}{
					names.AttrDatabase: parameters.SqlServerParameters.Database,
					"host":             parameters.SqlServerParameters.Host,
					names.AttrPort:     parameters.SqlServerParameters.Port,
				},
			},
		})
	}

	if parameters.TeradataParameters != nil {
		params = append(params, map[string]interface{}{
			"teradata": []interface{}{
				map[string]interface{}{
					names.AttrDatabase: parameters.TeradataParameters.Database,
					"host":             parameters.TeradataParameters.Host,
					names.AttrPort:     parameters.TeradataParameters.Port,
				},
			},
		})
	}

	if parameters.TwitterParameters != nil {
		params = append(params, map[string]interface{}{
			"twitter": []interface{}{
				map[string]interface{}{
					"max_rows": parameters.TwitterParameters.MaxRows,
					"query":    parameters.TwitterParameters.Query,
				},
			},
		})
	}

	return params
}

func flattenSSLProperties(props *quicksight.SslProperties) []interface{} {
	if props == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if props.DisableSsl != nil {
		m["disable_ssl"] = aws.BoolValue(props.DisableSsl)
	}

	return []interface{}{m}
}

func flattenVPCConnectionProperties(props *quicksight.VpcConnectionProperties) []interface{} {
	if props == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if props.VpcConnectionArn != nil {
		m["vpc_connection_arn"] = aws.StringValue(props.VpcConnectionArn)
	}

	return []interface{}{m}
}

func ParseDataSourceID(id string) (string, string, error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID/DATA_SOURCE_ID", id)
	}
	return parts[0], parts[1], nil
}
