// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
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

// @SDKResource("aws_quicksight_data_source", name="Data Source")
// @Tags(identifierAttribute="arn")
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
				"permission": quicksightschema.PermissionsSchema(),
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
					Type:             schema.TypeString,
					Required:         true,
					ForceNew:         true,
					ValidateDiagFunc: enum.Validate[awstypes.DataSourceType](),
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
	conn := meta.(*conns.AWSClient).QuickSightClient(ctx)

	awsAccountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk(names.AttrAWSAccountID); ok {
		awsAccountID = v.(string)
	}
	dataSourceID := d.Get("data_source_id").(string)
	id := dataSourceCreateResourceID(awsAccountID, dataSourceID)
	input := &quicksight.CreateDataSourceInput{
		AwsAccountId:         aws.String(awsAccountID),
		DataSourceId:         aws.String(dataSourceID),
		DataSourceParameters: expandDataSourceParameters(d.Get(names.AttrParameters).([]interface{})),
		Name:                 aws.String(d.Get(names.AttrName).(string)),
		Tags:                 getTagsIn(ctx),
		Type:                 awstypes.DataSourceType(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk("credentials"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Credentials = expandDataSourceCredentials(v.([]interface{}))
	}

	if v, ok := d.Get(names.AttrPermissions).(*schema.Set); ok && v.Len() > 0 {
		input.Permissions = quicksightschema.ExpandResourcePermissions(v.List())
	}

	if v, ok := d.GetOk("ssl_properties"); ok && len(v.([]interface{})) != 0 && v.([]interface{})[0] != nil {
		input.SslProperties = expandDataSourceSSLProperties(v.([]interface{}))
	}

	if v, ok := d.GetOk("vpc_connection_properties"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.VpcConnectionProperties = expandVPCConnectionProperties(v.([]interface{}))
	}

	_, err := conn.CreateDataSource(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating QuickSight Data Source (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitDataSourceCreated(ctx, conn, awsAccountID, dataSourceID); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting from QuickSight Data Source (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceDataSourceRead(ctx, d, meta)...)
}

func resourceDataSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	if err := d.Set(names.AttrParameters, flattenDataSourceParameters(dataSource.DataSourceParameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameters: %s", err)
	}
	if err := d.Set("ssl_properties", flattenSSLProperties(dataSource.SslProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ssl_properties: %s", err)
	}
	d.Set(names.AttrType, dataSource.Type)
	if err := d.Set("vpc_connection_properties", flattenVPCConnectionProperties(dataSource.VpcConnectionProperties)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc_connection_properties: %s", err)
	}

	permissions, err := findDataSourcePermissionsByTwoPartKey(ctx, conn, awsAccountID, dataSourceID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading QuickSight Data Source (%s) permissions: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrPermissions, quicksightschema.FlattenPermissions(permissions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting permissions: %s", err)
	}

	return diags
}

func resourceDataSourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

		if v, ok := d.GetOk("credentials"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.Credentials = expandDataSourceCredentials(v.([]interface{}))
		}

		if v, ok := d.GetOk(names.AttrParameters); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.DataSourceParameters = expandDataSourceParameters(v.([]interface{}))
		}

		if v, ok := d.GetOk("ssl_properties"); ok && len(v.([]interface{})) != 0 && v.([]interface{})[0] != nil {
			input.SslProperties = expandDataSourceSSLProperties(v.([]interface{}))
		}

		if v, ok := d.GetOk("vpc_connection_properties"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.VpcConnectionProperties = expandVPCConnectionProperties(v.([]interface{}))
		}

		_, err = conn.UpdateDataSource(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating QuickSight Data Source (%s): %s", d.Id(), err)
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

func resourceDataSourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	return func() (interface{}, string, error) {
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

func expandDataSourceCredentials(tfList []interface{}) *awstypes.DataSourceCredentials {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DataSourceCredentials{}

	if v, ok := tfMap["copy_source_arn"].(string); ok && v != "" {
		apiObject.CopySourceArn = aws.String(v)
	}

	if v, ok := tfMap["credential_pair"].([]interface{}); ok && len(v) > 0 {
		apiObject.CredentialPair = expandDataSourceCredentialPair(v)
	}

	return apiObject
}

func expandDataSourceCredentialPair(tfList []interface{}) *awstypes.CredentialPair {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.CredentialPair{}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	if v, ok := tfMap[names.AttrUsername].(string); ok && v != "" {
		apiObject.Username = aws.String(v)
	}

	if v, ok := tfMap[names.AttrPassword].(string); ok && v != "" {
		apiObject.Password = aws.String(v)
	}

	return apiObject
}

func expandDataSourceParameters(tfList []interface{}) awstypes.DataSourceParameters {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	var apiObject awstypes.DataSourceParameters

	if v, ok := tfMap["amazon_elasticsearch"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberAmazonElasticsearchParameters{}

			if v, ok := tfMap[names.AttrDomain].(string); ok && v != "" {
				ps.Value.Domain = aws.String(v)
			}

			apiObject = ps
		}
	}

	if v := tfMap["athena"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberAthenaParameters{}

			if v, ok := tfMap["work_group"].(string); ok && v != "" {
				ps.Value.WorkGroup = aws.String(v)
			}

			apiObject = ps
		}
	}

	if v := tfMap["aurora"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberAuroraParameters{}

			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = aws.Int32(int32(v))
			}

			apiObject = ps
		}
	}

	if v, ok := tfMap["aurora_postgresql"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberAuroraPostgreSqlParameters{}

			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = aws.Int32(int32(v))
			}

			apiObject = ps
		}
	}

	if v := tfMap["aws_iot_analytics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberAwsIotAnalyticsParameters{}

			if v, ok := tfMap["data_set_name"].(string); ok && v != "" {
				ps.Value.DataSetName = aws.String(v)
			}

			apiObject = ps
		}
	}

	if v := tfMap["jira"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberJiraParameters{}

			if v, ok := tfMap["site_base_url"].(string); ok && v != "" {
				ps.Value.SiteBaseUrl = aws.String(v)
			}

			apiObject = ps
		}
	}

	if v := tfMap["maria_db"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberMariaDbParameters{}

			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = aws.Int32(int32(v))
			}

			apiObject = ps
		}
	}

	if v := tfMap["mysql"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberMySqlParameters{}

			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = aws.Int32(int32(v))
			}

			apiObject = ps
		}
	}

	if v := tfMap["oracle"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberOracleParameters{}

			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = aws.Int32(int32(v))
			}

			apiObject = ps
		}
	}

	if v := tfMap["postgresql"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberPostgreSqlParameters{}

			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = aws.Int32(int32(v))
			}

			apiObject = ps
		}
	}

	if v := tfMap["presto"].([]interface{}); ok && len(v) > 0 && v != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberPrestoParameters{}

			if v, ok := tfMap["catalog"].(string); ok && v != "" {
				ps.Value.Catalog = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = aws.Int32(int32(v))
			}

			apiObject = ps
		}
	}

	if v := tfMap["rds"].([]interface{}); ok && len(v) > 0 && v != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberRdsParameters{}

			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap[names.AttrInstanceID].(string); ok && v != "" {
				ps.Value.InstanceId = aws.String(v)
			}

			apiObject = ps
		}
	}

	if v := tfMap["redshift"].([]interface{}); ok && len(v) > 0 && v != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberRedshiftParameters{}

			if v, ok := tfMap["cluster_id"].(string); ok && v != "" {
				ps.Value.ClusterId = aws.String(v)
			}
			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = int32(v)
			}

			apiObject = ps
		}
	}

	if v := tfMap["s3"].([]interface{}); ok && len(v) > 0 && v != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberS3Parameters{}

			if v, ok := tfMap["manifest_file_location"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
				if tfMap, ok := v[0].(map[string]interface{}); ok {
					apiObject := &awstypes.ManifestFileLocation{}

					if v, ok := tfMap[names.AttrBucket].(string); ok && v != "" {
						apiObject.Bucket = aws.String(v)
					}
					if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
						apiObject.Key = aws.String(v)
					}

					ps.Value.ManifestFileLocation = apiObject
				}
			}

			apiObject = ps
		}
	}

	if v := tfMap["service_now"].([]interface{}); ok && len(v) > 0 && v != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberServiceNowParameters{}

			if v, ok := tfMap["site_base_url"].(string); ok && v != "" {
				ps.Value.SiteBaseUrl = aws.String(v)
			}

			apiObject = ps
		}
	}

	if v := tfMap["snowflake"].([]interface{}); ok && len(v) > 0 && v != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberSnowflakeParameters{}

			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap["warehouse"].(string); ok && v != "" {
				ps.Value.Warehouse = aws.String(v)
			}

			apiObject = ps
		}
	}

	if v := tfMap["spark"].([]interface{}); ok && len(v) > 0 && v != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberSparkParameters{}

			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = aws.Int32(int32(v))
			}

			apiObject = ps
		}
	}

	if v := tfMap["sql_server"].([]interface{}); ok && len(v) > 0 && v != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberSqlServerParameters{}

			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = aws.Int32(int32(v))
			}

			apiObject = ps
		}
	}

	if v := tfMap["teradata"].([]interface{}); ok && len(v) > 0 && v != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberTeradataParameters{}

			if v, ok := tfMap[names.AttrDatabase].(string); ok && v != "" {
				ps.Value.Database = aws.String(v)
			}
			if v, ok := tfMap["host"].(string); ok && v != "" {
				ps.Value.Host = aws.String(v)
			}
			if v, ok := tfMap[names.AttrPort].(int); ok {
				ps.Value.Port = aws.Int32(int32(v))
			}

			apiObject = ps
		}
	}

	if v := tfMap["twitter"].([]interface{}); ok && len(v) > 0 && v != nil {
		if tfMap, ok := v[0].(map[string]interface{}); ok {
			ps := &awstypes.DataSourceParametersMemberTwitterParameters{}

			if v, ok := tfMap["max_rows"].(int); ok {
				ps.Value.MaxRows = aws.Int32(int32(v))
			}
			if v, ok := tfMap["query"].(string); ok && v != "" {
				ps.Value.Query = aws.String(v)
			}

			apiObject = ps
		}
	}

	return apiObject
}

func flattenDataSourceParameters(apiObject awstypes.DataSourceParameters) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{}

	switch v := apiObject.(type) {
	case *awstypes.DataSourceParametersMemberAmazonElasticsearchParameters:
		tfMap["amazon_elasticsearch"] = []interface{}{
			map[string]interface{}{
				names.AttrDomain: aws.ToString(v.Value.Domain),
			},
		}
	case *awstypes.DataSourceParametersMemberAthenaParameters:
		tfMap["athena"] = []interface{}{
			map[string]interface{}{
				"work_group": aws.ToString(v.Value.WorkGroup),
			},
		}
	case *awstypes.DataSourceParametersMemberAuroraParameters:
		tfMap["aurora"] = []interface{}{
			map[string]interface{}{
				names.AttrDatabase: aws.ToString(v.Value.Database),
				"host":             aws.ToString(v.Value.Host),
				names.AttrPort:     aws.ToInt32(v.Value.Port),
			},
		}
	case *awstypes.DataSourceParametersMemberAuroraPostgreSqlParameters:
		tfMap["aurora_postgresql"] = []interface{}{
			map[string]interface{}{
				names.AttrDatabase: aws.ToString(v.Value.Database),
				"host":             aws.ToString(v.Value.Host),
				names.AttrPort:     aws.ToInt32(v.Value.Port),
			},
		}
	case *awstypes.DataSourceParametersMemberAwsIotAnalyticsParameters:
		tfMap["aws_iot_analytics"] = []interface{}{
			map[string]interface{}{
				"data_set_name": aws.ToString(v.Value.DataSetName),
			},
		}
	case *awstypes.DataSourceParametersMemberJiraParameters:
		tfMap["jira"] = []interface{}{
			map[string]interface{}{
				"site_base_url": aws.ToString(v.Value.SiteBaseUrl),
			},
		}
	case *awstypes.DataSourceParametersMemberMariaDbParameters:
		tfMap["maria_db"] = []interface{}{
			map[string]interface{}{
				names.AttrDatabase: aws.ToString(v.Value.Database),
				"host":             aws.ToString(v.Value.Host),
				names.AttrPort:     aws.ToInt32(v.Value.Port),
			},
		}
	case *awstypes.DataSourceParametersMemberMySqlParameters:
		tfMap["mysql"] = []interface{}{
			map[string]interface{}{
				names.AttrDatabase: aws.ToString(v.Value.Database),
				"host":             aws.ToString(v.Value.Host),
				names.AttrPort:     aws.ToInt32(v.Value.Port),
			},
		}
	case *awstypes.DataSourceParametersMemberOracleParameters:
		tfMap["oracle"] = []interface{}{
			map[string]interface{}{
				names.AttrDatabase: aws.ToString(v.Value.Database),
				"host":             aws.ToString(v.Value.Host),
				names.AttrPort:     aws.ToInt32(v.Value.Port),
			},
		}
	case *awstypes.DataSourceParametersMemberPostgreSqlParameters:
		tfMap["postgresql"] = []interface{}{
			map[string]interface{}{
				names.AttrDatabase: aws.ToString(v.Value.Database),
				"host":             aws.ToString(v.Value.Host),
				names.AttrPort:     aws.ToInt32(v.Value.Port),
			},
		}
	case *awstypes.DataSourceParametersMemberPrestoParameters:
		tfMap["postgresql"] = []interface{}{
			map[string]interface{}{
				"catalog":      aws.ToString(v.Value.Catalog),
				"host":         aws.ToString(v.Value.Host),
				names.AttrPort: aws.ToInt32(v.Value.Port),
			},
		}
	case *awstypes.DataSourceParametersMemberRdsParameters:
		tfMap["rds"] = []interface{}{
			map[string]interface{}{
				names.AttrDatabase:   aws.ToString(v.Value.Database),
				names.AttrInstanceID: aws.ToString(v.Value.InstanceId),
			},
		}
	case *awstypes.DataSourceParametersMemberRedshiftParameters:
		tfMap["redshift"] = []interface{}{
			map[string]interface{}{
				"cluster_id":       aws.ToString(v.Value.ClusterId),
				names.AttrDatabase: aws.ToString(v.Value.Database),
				"host":             aws.ToString(v.Value.Host),
				names.AttrPort:     v.Value.Port,
			},
		}
	case *awstypes.DataSourceParametersMemberS3Parameters:
		tfMap["s3"] = []interface{}{
			map[string]interface{}{
				names.AttrBucket: aws.ToString(v.Value.ManifestFileLocation.Bucket),
				names.AttrKey:    aws.ToString(v.Value.ManifestFileLocation.Key),
			},
		}
	case *awstypes.DataSourceParametersMemberServiceNowParameters:
		tfMap["service_now"] = []interface{}{
			map[string]interface{}{
				"site_base_url": aws.ToString(v.Value.SiteBaseUrl),
			},
		}
	case *awstypes.DataSourceParametersMemberSnowflakeParameters:
		tfMap["snowflake"] = []interface{}{
			map[string]interface{}{
				names.AttrDatabase: aws.ToString(v.Value.Database),
				"host":             aws.ToString(v.Value.Host),
				"warehouse":        aws.ToString(v.Value.Warehouse),
			},
		}
	case *awstypes.DataSourceParametersMemberSparkParameters:
		tfMap["snowflake"] = []interface{}{
			map[string]interface{}{
				"host":         aws.ToString(v.Value.Host),
				names.AttrPort: aws.ToInt32(v.Value.Port),
			},
		}
	case *awstypes.DataSourceParametersMemberSqlServerParameters:
		tfMap["sql_server"] = []interface{}{
			map[string]interface{}{
				names.AttrDatabase: aws.ToString(v.Value.Database),
				"host":             aws.ToString(v.Value.Host),
				names.AttrPort:     v.Value.Port,
			},
		}
	case *awstypes.DataSourceParametersMemberTeradataParameters:
		tfMap["teradata"] = []interface{}{
			map[string]interface{}{
				names.AttrDatabase: aws.ToString(v.Value.Database),
				"host":             aws.ToString(v.Value.Host),
				names.AttrPort:     v.Value.Port,
			},
		}
	case *awstypes.DataSourceParametersMemberTwitterParameters:
		tfMap["teradata"] = []interface{}{
			map[string]interface{}{
				"max_rows": aws.ToInt32(v.Value.MaxRows),
				"query":    aws.ToString(v.Value.Query),
			},
		}
	default:
		return nil
	}

	return []interface{}{tfMap}
}

func expandDataSourceSSLProperties(tfList []interface{}) *awstypes.SslProperties {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SslProperties{}

	if v, ok := tfMap["disable_ssl"].(bool); ok {
		apiObject.DisableSsl = v
	}

	return apiObject
}

func flattenSSLProperties(apiObject *awstypes.SslProperties) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"disable_ssl": apiObject.DisableSsl,
	}

	return []interface{}{tfMap}
}

func expandVPCConnectionProperties(tfList []interface{}) *awstypes.VpcConnectionProperties {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.VpcConnectionProperties{}

	if v, ok := tfMap["vpc_connection_arn"].(string); ok && v != "" {
		apiObject.VpcConnectionArn = aws.String(v)
	}

	return apiObject
}

func flattenVPCConnectionProperties(apiObject *awstypes.VpcConnectionProperties) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{}

	if apiObject.VpcConnectionArn != nil {
		tfMap["vpc_connection_arn"] = aws.ToString(apiObject.VpcConnectionArn)
	}

	return []interface{}{tfMap}
}
