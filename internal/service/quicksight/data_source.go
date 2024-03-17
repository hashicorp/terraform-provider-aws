// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
				"arn": {
					Type:     schema.TypeString,
					Computed: true,
				},

				"aws_account_id": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					ForceNew:         true,
					ValidateDiagFunc: validation.ToDiagFunc(verify.ValidAccountID),
				},

				"credentials": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"copy_source_arn": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
								ConflictsWith:    []string{"credentials.0.credential_pair"},
							},
							"credential_pair": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"password": {
											Type:     schema.TypeString,
											Required: true,
											ValidateDiagFunc: validation.AllDiag(
												validation.ToDiagFunc(validation.NoZeroValues),
												validation.ToDiagFunc(validation.StringLenBetween(1, 1024)),
											),
											Sensitive: true,
										},
										"username": {
											Type:     schema.TypeString,
											Required: true,
											ValidateDiagFunc: validation.AllDiag(
												validation.ToDiagFunc(validation.NoZeroValues),
												validation.ToDiagFunc(validation.StringLenBetween(1, 64)),
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

				"name": {
					Type:     schema.TypeString,
					Required: true,
					ValidateDiagFunc: validation.AllDiag(
						validation.ToDiagFunc(validation.NoZeroValues),
						validation.ToDiagFunc(validation.StringLenBetween(1, 128)),
					),
				},

				"parameters": {
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
										"domain": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
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
											Type:             schema.TypeString,
											Optional:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
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
										"database": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"host": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"port": {
											Type:             schema.TypeInt,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
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
										"database": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"host": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"port": {
											Type:             schema.TypeInt,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
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
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
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
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
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
										"database": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"host": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"port": {
											Type:             schema.TypeInt,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
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
										"database": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"host": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"port": {
											Type:             schema.TypeInt,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
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
										"database": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"host": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"port": {
											Type:             schema.TypeInt,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
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
										"database": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"host": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"port": {
											Type:             schema.TypeInt,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
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
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"host": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"port": {
											Type:             schema.TypeInt,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
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
										"database": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"instance_id": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
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
											Type:             schema.TypeString,
											Optional:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"database": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"host": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"port": {
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
													"bucket": {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
													},
													"key": {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
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
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
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
										"database": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"host": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"warehouse": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
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
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"port": {
											Type:             schema.TypeInt,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
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
										"database": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"host": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"port": {
											Type:             schema.TypeInt,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
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
										"database": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"host": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
										},
										"port": {
											Type:             schema.TypeInt,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
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
											Type:             schema.TypeInt,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(1)),
										},
										"query": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
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
							"actions": {
								Type:     schema.TypeSet,
								Required: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
								MinItems: 1,
								MaxItems: 16,
							},
							"principal": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
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

				"type": {
					Type:             schema.TypeString,
					Required:         true,
					ForceNew:         true,
					ValidateDiagFunc: enum.Validate[types.DataSourceType](),
				},

				"vpc_connection_properties": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"vpc_connection_arn": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
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
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountId := meta.(*conns.AWSClient).AccountID
	id := d.Get("data_source_id").(string)

	if v, ok := d.GetOk("aws_account_id"); ok {
		awsAccountId = v.(string)
	}

	params := &quicksight.CreateDataSourceInput{
		AwsAccountId:         aws.String(awsAccountId),
		DataSourceId:         aws.String(id),
		DataSourceParameters: expandDataSourceParameters(d.Get("parameters").([]interface{})),
		Name:                 aws.String(d.Get("name").(string)),
		Tags:                 getTagsIn(ctx),
		Type:                 types.DataSourceType(d.Get("type").(string)),
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

	_, err := conn.CreateDataSource(ctx, params)
	if err != nil {
		return diag.Errorf("creating QuickSight Data Source: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s", awsAccountId, id))

	if _, err := waitCreated(ctx, conn, awsAccountId, id); err != nil {
		return diag.Errorf("waiting from QuickSight Data Source (%s) creation: %s", d.Id(), err)
	}

	return resourceDataSourceRead(ctx, d, meta)
}

func resourceDataSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountId, dataSourceId, err := ParseDataSourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	descOpts := &quicksight.DescribeDataSourceInput{
		AwsAccountId: aws.String(awsAccountId),
		DataSourceId: aws.String(dataSourceId),
	}

	output, err := conn.DescribeDataSource(ctx, descOpts)

	if !d.IsNewResource() && errs.IsA[*types.ResourceNotFoundException](err) {
		log.Printf("[WARN] QuickSight Data Source (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("describing QuickSight Data Source (%s): %s", d.Id(), err)
	}

	if output == nil || output.DataSource == nil {
		return diag.Errorf("describing QuickSight Data Source (%s): empty output", d.Id())
	}

	dataSource := output.DataSource

	d.Set("arn", dataSource.Arn)
	d.Set("aws_account_id", awsAccountId)
	d.Set("data_source_id", dataSource.DataSourceId)
	d.Set("name", dataSource.Name)

	if err := d.Set("parameters", flattenParameters(dataSource.DataSourceParameters)); err != nil {
		return diag.Errorf("setting parameters: %s", err)
	}

	if err := d.Set("ssl_properties", flattenSSLProperties(dataSource.SslProperties)); err != nil {
		return diag.Errorf("setting ssl_properties: %s", err)
	}

	d.Set("type", dataSource.Type)

	if err := d.Set("vpc_connection_properties", flattenVPCConnectionProperties(dataSource.VpcConnectionProperties)); err != nil {
		return diag.Errorf("setting vpc_connection_properties: %s", err)
	}

	permsResp, err := conn.DescribeDataSourcePermissions(ctx, &quicksight.DescribeDataSourcePermissionsInput{
		AwsAccountId: aws.String(awsAccountId),
		DataSourceId: aws.String(dataSourceId),
	})

	if err != nil {
		return diag.Errorf("describing QuickSight Data Source (%s) Permissions: %s", d.Id(), err)
	}

	if err := d.Set("permission", flattenPermissions(permsResp.Permissions)); err != nil {
		return diag.Errorf("setting permission: %s", err)
	}

	return nil
}

func resourceDataSourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	if d.HasChangesExcept("permission", "tags", "tags_all") {
		awsAccountId, dataSourceId, err := ParseDataSourceID(d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		params := &quicksight.UpdateDataSourceInput{
			AwsAccountId: aws.String(awsAccountId),
			DataSourceId: aws.String(dataSourceId),
			Name:         aws.String(d.Get("name").(string)),
		}

		if v, ok := d.GetOk("parameters"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
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

		_, err = conn.UpdateDataSource(ctx, params)

		if err != nil {
			return diag.Errorf("updating QuickSight Data Source (%s): %s", d.Id(), err)
		}

		if _, err := waitUpdated(ctx, conn, awsAccountId, dataSourceId); err != nil {
			return diag.Errorf("waiting for QuickSight Data Source (%s) to update: %s", d.Id(), err)
		}
	}

	if d.HasChange("permission") {
		awsAccountId, dataSourceId, err := ParseDataSourceID(d.Id())
		if err != nil {
			return diag.FromErr(err)
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

		_, err = conn.UpdateDataSourcePermissions(ctx, params)

		if err != nil {
			return diag.Errorf("updating QuickSight Data Source (%s) permissions: %s", dataSourceId, err)
		}
	}

	return resourceDataSourceRead(ctx, d, meta)
}

func resourceDataSourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountId, dataSourceId, err := ParseDataSourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	deleteOpts := &quicksight.DeleteDataSourceInput{
		AwsAccountId: aws.String(awsAccountId),
		DataSourceId: aws.String(dataSourceId),
	}

	_, err = conn.DeleteDataSource(ctx, deleteOpts)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting QuickSight Data Source (%s): %s", d.Id(), err)
	}

	return nil
}

func expandDataSourceCredentials(tfList []interface{}) *types.DataSourceCredentials {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	credentials := &types.DataSourceCredentials{}

	if v, ok := tfMap["copy_source_arn"].(string); ok && v != "" {
		credentials.CopySourceArn = aws.String(v)
	}

	if v, ok := tfMap["credential_pair"].([]interface{}); ok && len(v) > 0 {
		credentials.CredentialPair = expandDataSourceCredentialPair(v)
	}

	return credentials
}

func expandDataSourceCredentialPair(tfList []interface{}) *types.CredentialPair {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	credentialPair := &types.CredentialPair{}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	if v, ok := tfMap["username"].(string); ok && v != "" {
		credentialPair.Username = aws.String(v)
	}

	if v, ok := tfMap["password"].(string); ok && v != "" {
		credentialPair.Password = aws.String(v)
	}

	return credentialPair
}

func expandDataSourceParameters(tfList []interface{}) types.DataSourceParameters {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	var dataSourceParams types.DataSourceParameters

	if v, ok := tfMap["amazon_elasticsearch"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &types.AmazonElasticsearchParameters{}

			if v, ok := m["domain"].(string); ok && v != "" {
				ps.Domain = aws.String(v)
			}

			dataSourceParams = &types.DataSourceParametersMemberAmazonElasticsearchParameters{Value: *ps}
		}
	}

	if v := tfMap["athena"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &types.AthenaParameters{}
			if v, ok := m["work_group"].(string); ok && v != "" {
				ps.WorkGroup = aws.String(v)
			}

			dataSourceParams = &types.DataSourceParametersMemberAthenaParameters{Value: *ps}
		}
	}

	if v := tfMap["aurora"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &types.AuroraParameters{}
			if v, ok := m["database"].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m["port"].(int); ok {
				ps.Port = aws.Int32(int32(v))
			}

			dataSourceParams = &types.DataSourceParametersMemberAuroraParameters{Value: *ps}
		}
	}

	if v, ok := tfMap["aurora_postgresql"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &types.AuroraPostgreSqlParameters{}
			if v, ok := m["database"].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m["port"].(int); ok {
				ps.Port = aws.Int32(int32(v))
			}

			dataSourceParams = &types.DataSourceParametersMemberAuroraPostgreSqlParameters{Value: *ps}
		}
	}

	if v := tfMap["aws_iot_analytics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &types.AwsIotAnalyticsParameters{}
			if v, ok := m["data_set_name"].(string); ok && v != "" {
				ps.DataSetName = aws.String(v)
			}

			dataSourceParams = &types.DataSourceParametersMemberAwsIotAnalyticsParameters{Value: *ps}
		}
	}

	if v := tfMap["jira"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &types.JiraParameters{}
			if v, ok := m["site_base_url"].(string); ok && v != "" {
				ps.SiteBaseUrl = aws.String(v)
			}

			dataSourceParams = &types.DataSourceParametersMemberJiraParameters{Value: *ps}
		}
	}

	if v := tfMap["maria_db"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &types.MariaDbParameters{}
			if v, ok := m["database"].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m["port"].(int); ok {
				ps.Port = aws.Int32(int32(v))
			}

			dataSourceParams = &types.DataSourceParametersMemberMariaDbParameters{Value: *ps}
		}
	}

	if v := tfMap["mysql"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &types.MySqlParameters{}
			if v, ok := m["database"].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m["port"].(int); ok {
				ps.Port = aws.Int32(int32(v))
			}

			dataSourceParams = &types.DataSourceParametersMemberMySqlParameters{Value: *ps}
		}
	}

	if v := tfMap["oracle"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &types.OracleParameters{}
			if v, ok := m["database"].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m["port"].(int); ok {
				ps.Port = aws.Int32(int32(v))
			}

			dataSourceParams = &types.DataSourceParametersMemberOracleParameters{Value: *ps}
		}
	}

	if v := tfMap["postgresql"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &types.PostgreSqlParameters{}
			if v, ok := m["database"].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m["port"].(int); ok {
				ps.Port = aws.Int32(int32(v))
			}

			dataSourceParams = &types.DataSourceParametersMemberPostgreSqlParameters{Value: *ps}
		}
	}
	if v := tfMap["presto"].([]interface{}); ok && len(v) > 0 && v != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &types.PrestoParameters{}
			if v, ok := m["catalog"].(string); ok && v != "" {
				ps.Catalog = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m["port"].(int); ok {
				ps.Port = aws.Int32(int32(v))
			}

			dataSourceParams = &types.DataSourceParametersMemberPrestoParameters{Value: *ps}
		}
	}

	if v := tfMap["rds"].([]interface{}); ok && len(v) > 0 && v != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &types.RdsParameters{}

			if v, ok := m["database"].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m["instance_id"].(string); ok && v != "" {
				ps.InstanceId = aws.String(v)
			}

			dataSourceParams = &types.DataSourceParametersMemberRdsParameters{Value: *ps}
		}
	}

	if v := tfMap["redshift"].([]interface{}); ok && len(v) > 0 && v != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &types.RedshiftParameters{}
			if v, ok := m["cluster_id"].(string); ok && v != "" {
				ps.ClusterId = aws.String(v)
			}
			if v, ok := m["database"].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m["port"].(int); ok {
				ps.Port = int32(v)
			}

			dataSourceParams = &types.DataSourceParametersMemberRedshiftParameters{Value: *ps}
		}
	}

	if v := tfMap["s3"].([]interface{}); ok && len(v) > 0 && v != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &types.S3Parameters{}
			if v, ok := m["manifest_file_location"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
				lm, ok := v[0].(map[string]interface{})
				if ok {
					loc := &types.ManifestFileLocation{}

					if v, ok := lm["bucket"].(string); ok && v != "" {
						loc.Bucket = aws.String(v)
					}

					if v, ok := lm["key"].(string); ok && v != "" {
						loc.Key = aws.String(v)
					}

					ps.ManifestFileLocation = loc
				}
			}

			dataSourceParams = &types.DataSourceParametersMemberS3Parameters{Value: *ps}
		}
	}

	if v := tfMap["service_now"].([]interface{}); ok && len(v) > 0 && v != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &types.ServiceNowParameters{}
			if v, ok := m["site_base_url"].(string); ok && v != "" {
				ps.SiteBaseUrl = aws.String(v)
			}

			dataSourceParams = &types.DataSourceParametersMemberServiceNowParameters{Value: *ps}
		}
	}

	if v := tfMap["snowflake"].([]interface{}); ok && len(v) > 0 && v != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &types.SnowflakeParameters{}

			if v, ok := m["database"].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m["warehouse"].(string); ok && v != "" {
				ps.Warehouse = aws.String(v)
			}

			dataSourceParams = &types.DataSourceParametersMemberSnowflakeParameters{Value: *ps}
		}
	}

	if v := tfMap["spark"].([]interface{}); ok && len(v) > 0 && v != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &types.SparkParameters{}

			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m["port"].(int); ok {
				ps.Port = aws.Int32(int32(v))
			}

			dataSourceParams = &types.DataSourceParametersMemberSparkParameters{Value: *ps}
		}
	}

	if v := tfMap["sql_server"].([]interface{}); ok && len(v) > 0 && v != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &types.SqlServerParameters{}

			if v, ok := m["database"].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m["port"].(int); ok {
				ps.Port = aws.Int32(int32(v))
			}

			dataSourceParams = &types.DataSourceParametersMemberSqlServerParameters{Value: *ps}
		}
	}

	if v := tfMap["teradata"].([]interface{}); ok && len(v) > 0 && v != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &types.TeradataParameters{}

			if v, ok := m["database"].(string); ok && v != "" {
				ps.Database = aws.String(v)
			}
			if v, ok := m["host"].(string); ok && v != "" {
				ps.Host = aws.String(v)
			}
			if v, ok := m["port"].(int); ok {
				ps.Port = aws.Int32(int32(v))
			}

			dataSourceParams = &types.DataSourceParametersMemberTeradataParameters{Value: *ps}
		}
	}

	if v := tfMap["twitter"].([]interface{}); ok && len(v) > 0 && v != nil {
		m, ok := v[0].(map[string]interface{})

		if ok {
			ps := &types.TwitterParameters{}

			if v, ok := m["max_rows"].(int); ok {
				ps.MaxRows = aws.Int32(int32(v))
			}
			if v, ok := m["query"].(string); ok && v != "" {
				ps.Query = aws.String(v)
			}

			dataSourceParams = &types.DataSourceParametersMemberTwitterParameters{Value: *ps}
		}
	}

	return dataSourceParams
}

func expandDataSourceSSLProperties(tfList []interface{}) *types.SslProperties {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	props := &types.SslProperties{}

	if v, ok := tfMap["disable_ssl"].(bool); ok {
		props.DisableSsl = v
	}

	return props
}

func expandDataSourceVPCConnectionProperties(tfList []interface{}) *types.VpcConnectionProperties {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})

	if !ok {
		return nil
	}

	props := &types.VpcConnectionProperties{}

	if v, ok := tfMap["vpc_connection_arn"].(string); ok && v != "" {
		props.VpcConnectionArn = aws.String(v)
	}

	return props
}

func flattenParameters(parameters types.DataSourceParameters) []interface{} {
	if parameters == nil {
		return []interface{}{}
	}

	var params []interface{}

	switch parameters := parameters.(type) {

	case *types.DataSourceParametersMemberAmazonElasticsearchParameters:
		params = append(params, map[string]interface{}{
			"amazon_elasticsearch": []interface{}{
				map[string]interface{}{
					"domain": parameters.Value.Domain,
				},
			},
		})

	case *types.DataSourceParametersMemberAthenaParameters:
		params = append(params, map[string]interface{}{
			"athena": []interface{}{
				map[string]interface{}{
					"work_group": parameters.Value.WorkGroup,
				},
			},
		})

	case *types.DataSourceParametersMemberAuroraParameters:
		params = append(params, map[string]interface{}{
			"aurora": []interface{}{
				map[string]interface{}{
					"database": parameters.Value.Database,
					"host":     parameters.Value.Host,
					"port":     parameters.Value.Port,
				},
			},
		})

	case *types.DataSourceParametersMemberAuroraPostgreSqlParameters:
		params = append(params, map[string]interface{}{
			"aurora_postgresql": []interface{}{
				map[string]interface{}{
					"database": parameters.Value.Database,
					"host":     parameters.Value.Host,
					"port":     parameters.Value.Port,
				},
			},
		})

	case *types.DataSourceParametersMemberAwsIotAnalyticsParameters:
		params = append(params, map[string]interface{}{
			"aws_iot_analytics": []interface{}{
				map[string]interface{}{
					"data_set_name": parameters.Value.DataSetName,
				},
			},
		})

	case *types.DataSourceParametersMemberJiraParameters:
		params = append(params, map[string]interface{}{
			"jira": []interface{}{
				map[string]interface{}{
					"site_base_url": parameters.Value.SiteBaseUrl,
				},
			},
		})

	case *types.DataSourceParametersMemberMariaDbParameters:
		params = append(params, map[string]interface{}{
			"maria_db": []interface{}{
				map[string]interface{}{
					"database": parameters.Value.Database,
					"host":     parameters.Value.Host,
					"port":     parameters.Value.Port},
			},
		})

	case *types.DataSourceParametersMemberMySqlParameters:
		params = append(params, map[string]interface{}{
			"mysql": []interface{}{
				map[string]interface{}{
					"database": parameters.Value.Database,
					"host":     parameters.Value.Host,
					"port":     parameters.Value.Port,
				},
			},
		})

	case *types.DataSourceParametersMemberOracleParameters:
		params = append(params, map[string]interface{}{
			"oracle": []interface{}{
				map[string]interface{}{
					"database": parameters.Value.Database,
					"host":     parameters.Value.Host,
					"port":     parameters.Value.Port,
				},
			},
		})

	case *types.DataSourceParametersMemberPostgreSqlParameters:
		params = append(params, map[string]interface{}{
			"postgresql": []interface{}{
				map[string]interface{}{
					"database": parameters.Value.Database,
					"host":     parameters.Value.Host,
					"port":     parameters.Value.Port},
			},
		})

	case *types.DataSourceParametersMemberPrestoParameters:
		params = append(params, map[string]interface{}{
			"presto": []interface{}{
				map[string]interface{}{
					"catalog": parameters.Value.Catalog,
					"host":    parameters.Value.Host,
					"port":    parameters.Value.Port,
				},
			},
		})

	case *types.DataSourceParametersMemberRdsParameters:
		params = append(params, map[string]interface{}{
			"rds": []interface{}{
				map[string]interface{}{
					"database":    parameters.Value.Database,
					"instance_id": parameters.Value.InstanceId,
				},
			},
		})

	case *types.DataSourceParametersMemberRedshiftParameters:
		params = append(params, map[string]interface{}{
			"redshift": []interface{}{
				map[string]interface{}{
					"cluster_id": parameters.Value.ClusterId,
					"database":   parameters.Value.Database,
					"host":       parameters.Value.Host,
					"port":       parameters.Value.Port,
				},
			},
		})

	case *types.DataSourceParametersMemberS3Parameters:
		params = append(params, map[string]interface{}{
			"s3": []interface{}{
				map[string]interface{}{
					"manifest_file_location": []interface{}{
						map[string]interface{}{
							"bucket": parameters.Value.ManifestFileLocation.Bucket,
							"key":    parameters.Value.ManifestFileLocation.Key,
						},
					},
				},
			},
		})

	case *types.DataSourceParametersMemberServiceNowParameters:
		params = append(params, map[string]interface{}{
			"service_now": []interface{}{
				map[string]interface{}{
					"site_base_url": parameters.Value.SiteBaseUrl,
				},
			},
		})

	case *types.DataSourceParametersMemberSnowflakeParameters:
		params = append(params, map[string]interface{}{
			"snowflake": []interface{}{
				map[string]interface{}{
					"database":  parameters.Value.Database,
					"host":      parameters.Value.Host,
					"warehouse": parameters.Value.Warehouse,
				},
			},
		})

	case *types.DataSourceParametersMemberSparkParameters:
		params = append(params, map[string]interface{}{
			"spark": []interface{}{
				map[string]interface{}{
					"host": parameters.Value.Host,
					"port": parameters.Value.Port,
				},
			},
		})

	case *types.DataSourceParametersMemberSqlServerParameters:
		params = append(params, map[string]interface{}{
			"sql_server": []interface{}{
				map[string]interface{}{
					"database": parameters.Value.Database,
					"host":     parameters.Value.Host,
					"port":     parameters.Value.Port,
				},
			},
		})

	case *types.DataSourceParametersMemberTeradataParameters:
		params = append(params, map[string]interface{}{
			"teradata": []interface{}{
				map[string]interface{}{
					"database": parameters.Value.Database,
					"host":     parameters.Value.Host,
					"port":     parameters.Value.Port,
				},
			},
		})

	case *types.DataSourceParametersMemberTwitterParameters:
		params = append(params, map[string]interface{}{
			"twitter": []interface{}{
				map[string]interface{}{
					"max_rows": parameters.Value.MaxRows,
					"query":    parameters.Value.Query,
				},
			},
		})
	}

	return params
}

func flattenSSLProperties(props *types.SslProperties) []interface{} {
	if props == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	m["disable_ssl"] = props.DisableSsl

	return []interface{}{m}
}

func flattenVPCConnectionProperties(props *types.VpcConnectionProperties) []interface{} {
	if props == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if props.VpcConnectionArn != nil {
		m["vpc_connection_arn"] = aws.ToString(props.VpcConnectionArn)
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
