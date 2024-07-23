// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package keyspaces

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/keyspaces"
	"github.com/aws/aws-sdk-go-v2/service/keyspaces/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_keyspaces_table", name="Table")
// @Tags(identifierAttribute="arn")
func resourceTable() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTableCreate,
		ReadWithoutTimeout:   resourceTableRead,
		UpdateWithoutTimeout: resourceTableUpdate,
		DeleteWithoutTimeout: resourceTableDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange("client_side_timestamps", func(_ context.Context, old, new, meta interface{}) bool {
				// Client-side timestamps cannot be disabled.
				return len(old.([]interface{})) == 1 && len(new.([]interface{})) == 0
			}),
			customdiff.ForceNewIfChange("schema_definition.0.column", func(_ context.Context, old, new, meta interface{}) bool {
				// Columns can only be added.
				if os, ok := old.(*schema.Set); ok {
					if ns, ok := new.(*schema.Set); ok {
						if del := os.Difference(ns); del.Len() > 0 {
							return true
						}
					}
				}

				return false
			}),
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capacity_specification": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"read_capacity_units": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"throughput_mode": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[types.ThroughputMode](),
						},
						"write_capacity_units": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
					},
				},
			},
			"client_side_timestamps": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrStatus: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.ClientSideTimestampsStatus](),
						},
					},
				},
			},
			names.AttrComment: {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrMessage: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					},
				},
			},
			"default_time_to_live": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 630720000),
			},
			"encryption_specification": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kms_key_identifier": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[types.EncryptionType](),
						},
					},
				},
			},
			"keyspace_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexache.MustCompile(`^[0-9A-Za-z][0-9A-Za-z_]{0,47}$`),
					"The keyspace name can have up to 48 characters. It must begin with an alpha-numeric character and can only contain alpha-numeric characters and underscores.",
				),
			},
			"point_in_time_recovery": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrStatus: {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[types.PointInTimeRecoveryStatus](),
						},
					},
				},
			},
			"schema_definition": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"clustering_key": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
										ValidateFunc: validation.StringMatch(
											regexache.MustCompile(`^[0-9a-z_]{1,48}$`),
											"The column name can have up to 48 characters. It can only contain lowercase alpha-numeric characters and underscores.",
										),
									},
									"order_by": {
										Type:             schema.TypeString,
										Required:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[types.SortOrder](),
									},
								},
							},
						},
						"column": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringMatch(
											regexache.MustCompile(`^[0-9a-z_]{1,48}$`),
											"The column name can have up to 48 characters. It can only contain lowercase alpha-numeric characters and underscores.",
										),
									},
									names.AttrType: {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringMatch(
											regexache.MustCompile(`^[0-9a-z]+(\<[0-9a-z]+(, *[0-9a-z]+){0,1}\>)?$`),
											"The type must consist of lower case alphanumerics and an optional list of upto two lower case alphanumerics enclosed in angle brackets '<>'.",
										),
									},
								},
							},
						},
						"partition_key": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
										ValidateFunc: validation.StringMatch(
											regexache.MustCompile(`^[0-9a-z_]{1,48}$`),
											"The column name can have up to 48 characters. It can only contain lowercase alpha-numeric characters and underscores.",
										),
									},
								},
							},
						},
						"static_column": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
										ValidateFunc: validation.StringMatch(
											regexache.MustCompile(`^[0-9a-z_]{1,48}$`),
											"The column name can have up to 48 characters. It can only contain lowercase alpha-numeric characters and underscores.",
										),
									},
								},
							},
						},
					},
				},
			},
			names.AttrTableName: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexache.MustCompile(`^[0-9A-Za-z_]{1,48}$`),
					"The table name can have up to 48 characters. It can only contain alpha-numeric characters and underscores.",
				),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"ttl": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrStatus: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.TimeToLiveStatus](),
						},
					},
				},
			},
		},
	}
}

func resourceTableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KeyspacesClient(ctx)

	keyspaceName := d.Get("keyspace_name").(string)
	tableName := d.Get(names.AttrTableName).(string)
	id := tableCreateResourceID(keyspaceName, tableName)
	input := &keyspaces.CreateTableInput{
		KeyspaceName: aws.String(keyspaceName),
		TableName:    aws.String(tableName),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("capacity_specification"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.CapacitySpecification = expandCapacitySpecification(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("client_side_timestamps"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ClientSideTimestamps = expandClientSideTimestamps(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk(names.AttrComment); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Comment = expandComment(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("default_time_to_live"); ok {
		input.DefaultTimeToLive = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("encryption_specification"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.EncryptionSpecification = expandEncryptionSpecification(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("point_in_time_recovery"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.PointInTimeRecovery = expandPointInTimeRecovery(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("schema_definition"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.SchemaDefinition = expandSchemaDefinition(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("ttl"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Ttl = expandTimeToLive(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.CreateTable(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Keyspaces Table (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitTableCreated(ctx, conn, keyspaceName, tableName, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Keyspaces Table (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceTableRead(ctx, d, meta)...)
}

func resourceTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KeyspacesClient(ctx)

	keyspaceName, tableName, err := tableParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	table, err := findTableByTwoPartKey(ctx, conn, keyspaceName, tableName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Keyspaces Table (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Keyspaces Table (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, table.ResourceArn)
	if table.CapacitySpecification != nil {
		if err := d.Set("capacity_specification", []interface{}{flattenCapacitySpecificationSummary(table.CapacitySpecification)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting capacity_specification: %s", err)
		}
	} else {
		d.Set("capacity_specification", nil)
	}
	if table.ClientSideTimestamps != nil {
		if err := d.Set("client_side_timestamps", []interface{}{flattenClientSideTimestamps(table.ClientSideTimestamps)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting client_side_timestamps: %s", err)
		}
	} else {
		d.Set("client_side_timestamps", nil)
	}
	if table.Comment != nil {
		if err := d.Set(names.AttrComment, []interface{}{flattenComment(table.Comment)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting comment: %s", err)
		}
	} else {
		d.Set(names.AttrComment, nil)
	}
	d.Set("default_time_to_live", table.DefaultTimeToLive)
	if table.EncryptionSpecification != nil {
		if err := d.Set("encryption_specification", []interface{}{flattenEncryptionSpecification(table.EncryptionSpecification)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting encryption_specification: %s", err)
		}
	} else {
		d.Set("encryption_specification", nil)
	}
	d.Set("keyspace_name", table.KeyspaceName)
	if table.PointInTimeRecovery != nil {
		if err := d.Set("point_in_time_recovery", []interface{}{flattenPointInTimeRecoverySummary(table.PointInTimeRecovery)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting point_in_time_recovery: %s", err)
		}
	} else {
		d.Set("point_in_time_recovery", nil)
	}
	if table.SchemaDefinition != nil {
		if err := d.Set("schema_definition", []interface{}{flattenSchemaDefinition(table.SchemaDefinition)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting schema_definition: %s", err)
		}
	} else {
		d.Set("schema_definition", nil)
	}
	d.Set(names.AttrTableName, table.TableName)
	if table.Ttl != nil {
		if err := d.Set("ttl", []interface{}{flattenTimeToLive(table.Ttl)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ttl: %s", err)
		}
	} else {
		d.Set("ttl", nil)
	}

	return diags
}

func resourceTableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KeyspacesClient(ctx)

	keyspaceName, tableName, err := tableParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		// https://docs.aws.amazon.com/keyspaces/latest/APIReference/API_UpdateTable.html
		// Note that you can only update one specific table setting per update operation.
		if d.HasChange("capacity_specification") {
			if v, ok := d.GetOk("capacity_specification"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input := &keyspaces.UpdateTableInput{
					CapacitySpecification: expandCapacitySpecification(v.([]interface{})[0].(map[string]interface{})),
					KeyspaceName:          aws.String(keyspaceName),
					TableName:             aws.String(tableName),
				}

				_, err := conn.UpdateTable(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "updating Keyspaces Table (%s) CapacitySpecification: %s", d.Id(), err)
				}

				if _, err := waitTableUpdated(ctx, conn, keyspaceName, tableName, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Keyspaces Table (%s) CapacitySpecification update: %s", d.Id(), err)
				}
			}
		}

		if d.HasChange("client_side_timestamps") {
			if v, ok := d.GetOk("client_side_timestamps"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input := &keyspaces.UpdateTableInput{
					ClientSideTimestamps: expandClientSideTimestamps(v.([]interface{})[0].(map[string]interface{})),
					KeyspaceName:         aws.String(keyspaceName),
					TableName:            aws.String(tableName),
				}

				_, err := conn.UpdateTable(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "updating Keyspaces Table (%s) ClientSideTimestamps: %s", d.Id(), err)
				}

				if _, err := waitTableUpdated(ctx, conn, keyspaceName, tableName, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Keyspaces Table (%s) ClientSideTimestamps update: %s", d.Id(), err)
				}
			}
		}

		if d.HasChange("default_time_to_live") {
			input := &keyspaces.UpdateTableInput{
				DefaultTimeToLive: aws.Int32(int32(d.Get("default_time_to_live").(int))),
				KeyspaceName:      aws.String(keyspaceName),
				TableName:         aws.String(tableName),
			}

			_, err := conn.UpdateTable(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Keyspaces Table (%s) DefaultTimeToLive: %s", d.Id(), err)
			}

			if _, err := waitTableUpdated(ctx, conn, keyspaceName, tableName, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Keyspaces Table (%s) DefaultTimeToLive update: %s", d.Id(), err)
			}
		}

		if d.HasChange("encryption_specification") {
			if v, ok := d.GetOk("encryption_specification"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input := &keyspaces.UpdateTableInput{
					EncryptionSpecification: expandEncryptionSpecification(v.([]interface{})[0].(map[string]interface{})),
					KeyspaceName:            aws.String(keyspaceName),
					TableName:               aws.String(tableName),
				}

				_, err := conn.UpdateTable(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "updating Keyspaces Table (%s) EncryptionSpecification: %s", d.Id(), err)
				}

				if _, err := waitTableUpdated(ctx, conn, keyspaceName, tableName, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Keyspaces Table (%s) EncryptionSpecification update: %s", d.Id(), err)
				}
			}
		}

		if d.HasChange("point_in_time_recovery") {
			if v, ok := d.GetOk("point_in_time_recovery"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input := &keyspaces.UpdateTableInput{
					KeyspaceName:        aws.String(keyspaceName),
					PointInTimeRecovery: expandPointInTimeRecovery(v.([]interface{})[0].(map[string]interface{})),
					TableName:           aws.String(tableName),
				}

				_, err := conn.UpdateTable(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "updating Keyspaces Table (%s) PointInTimeRecovery: %s", d.Id(), err)
				}

				if _, err := waitTableUpdated(ctx, conn, keyspaceName, tableName, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Keyspaces Table (%s) PointInTimeRecovery update: %s", d.Id(), err)
				}
			}
		}

		if d.HasChange("ttl") {
			if v, ok := d.GetOk("ttl"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input := &keyspaces.UpdateTableInput{
					KeyspaceName: aws.String(keyspaceName),
					TableName:    aws.String(tableName),
					Ttl:          expandTimeToLive(v.([]interface{})[0].(map[string]interface{})),
				}

				_, err := conn.UpdateTable(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "updating Keyspaces Table (%s) Ttl: %s", d.Id(), err)
				}

				if _, err := waitTableUpdated(ctx, conn, keyspaceName, tableName, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Keyspaces Table (%s) Ttl update: %s", d.Id(), err)
				}
			}
		}

		if d.HasChange("schema_definition") {
			o, n := d.GetChange("schema_definition")
			var os, ns *schema.Set

			if v, ok := o.([]interface{}); ok && len(v) > 0 && v[0] != nil {
				if v, ok := v[0].(map[string]interface{})["column"].(*schema.Set); ok {
					os = v
				}
			}
			if v, ok := n.([]interface{}); ok && len(v) > 0 && v[0] != nil {
				if v, ok := v[0].(map[string]interface{})["column"].(*schema.Set); ok {
					ns = v
				}
			}

			if os != nil && ns != nil {
				if add := ns.Difference(os); add.Len() > 0 {
					input := &keyspaces.UpdateTableInput{
						AddColumns:   expandColumnDefinitions(add.List()),
						KeyspaceName: aws.String(keyspaceName),
						TableName:    aws.String(tableName),
					}

					_, err := conn.UpdateTable(ctx, input)

					if err != nil {
						return sdkdiag.AppendErrorf(diags, "updating Keyspaces Table (%s) AddColumns: %s", d.Id(), err)
					}

					if _, err := waitTableUpdated(ctx, conn, keyspaceName, tableName, d.Timeout(schema.TimeoutUpdate)); err != nil {
						return sdkdiag.AppendErrorf(diags, "waiting for Keyspaces Table (%s) AddColumns update: %s", d.Id(), err)
					}
				}
			}
		}
	}

	return append(diags, resourceTableRead(ctx, d, meta)...)
}

func resourceTableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KeyspacesClient(ctx)

	keyspaceName, tableName, err := tableParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Keyspaces Table: (%s)", d.Id())
	_, err = conn.DeleteTable(ctx, &keyspaces.DeleteTableInput{
		KeyspaceName: aws.String(keyspaceName),
		TableName:    aws.String(tableName),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Keyspaces Table (%s): %s", d.Id(), err)
	}

	if _, err := waitTableDeleted(ctx, conn, keyspaceName, tableName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Keyspaces Table (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const tableIDSeparator = "/"

func tableCreateResourceID(keyspaceName, tableName string) string {
	parts := []string{keyspaceName, tableName}
	id := strings.Join(parts, tableIDSeparator)

	return id
}

func tableParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, tableIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected KEYSPACE-NAME%[2]sTABLE-NAME", id, tableIDSeparator)
}

func findTableByTwoPartKey(ctx context.Context, conn *keyspaces.Client, keyspaceName, tableName string) (*keyspaces.GetTableOutput, error) {
	input := keyspaces.GetTableInput{
		KeyspaceName: aws.String(keyspaceName),
		TableName:    aws.String(tableName),
	}

	output, err := conn.GetTable(ctx, &input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
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

	if status := output.Status; status == types.TableStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func statusTable(ctx context.Context, conn *keyspaces.Client, keyspaceName, tableName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findTableByTwoPartKey(ctx, conn, keyspaceName, tableName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitTableCreated(ctx context.Context, conn *keyspaces.Client, keyspaceName, tableName string, timeout time.Duration) (*keyspaces.GetTableOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.TableStatusCreating),
		Target:  enum.Slice(types.TableStatusActive),
		Refresh: statusTable(ctx, conn, keyspaceName, tableName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*keyspaces.GetTableOutput); ok {
		return output, err
	}

	return nil, err
}

func waitTableDeleted(ctx context.Context, conn *keyspaces.Client, keyspaceName, tableName string, timeout time.Duration) (*keyspaces.GetTableOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.TableStatusActive, types.TableStatusDeleting),
		Target:  []string{},
		Refresh: statusTable(ctx, conn, keyspaceName, tableName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*keyspaces.GetTableOutput); ok {
		return output, err
	}

	return nil, err
}

func waitTableUpdated(ctx context.Context, conn *keyspaces.Client, keyspaceName, tableName string, timeout time.Duration) (*keyspaces.GetTableOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.TableStatusUpdating),
		Target:  enum.Slice(types.TableStatusActive),
		Refresh: statusTable(ctx, conn, keyspaceName, tableName),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*keyspaces.GetTableOutput); ok {
		return output, err
	}

	return nil, err
}

func expandCapacitySpecification(tfMap map[string]interface{}) *types.CapacitySpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.CapacitySpecification{}

	if v, ok := tfMap["read_capacity_units"].(int); ok && v != 0 {
		apiObject.ReadCapacityUnits = aws.Int64(int64(v))
	}

	if v, ok := tfMap["throughput_mode"].(string); ok && v != "" {
		apiObject.ThroughputMode = types.ThroughputMode(v)
	}

	if v, ok := tfMap["write_capacity_units"].(int); ok && v != 0 {
		apiObject.WriteCapacityUnits = aws.Int64(int64(v))
	}

	return apiObject
}

func expandClientSideTimestamps(tfMap map[string]interface{}) *types.ClientSideTimestamps {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ClientSideTimestamps{}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		apiObject.Status = types.ClientSideTimestampsStatus(v)
	}

	return apiObject
}

func expandComment(tfMap map[string]interface{}) *types.Comment {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.Comment{}

	if v, ok := tfMap[names.AttrMessage].(string); ok && v != "" {
		apiObject.Message = aws.String(v)
	}

	return apiObject
}

func expandEncryptionSpecification(tfMap map[string]interface{}) *types.EncryptionSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.EncryptionSpecification{}

	if v, ok := tfMap["kms_key_identifier"].(string); ok && v != "" {
		apiObject.KmsKeyIdentifier = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = types.EncryptionType(v)
	}

	return apiObject
}

func expandPointInTimeRecovery(tfMap map[string]interface{}) *types.PointInTimeRecovery {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PointInTimeRecovery{}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		apiObject.Status = types.PointInTimeRecoveryStatus(v)
	}

	return apiObject
}

func expandSchemaDefinition(tfMap map[string]interface{}) *types.SchemaDefinition {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.SchemaDefinition{}

	if v, ok := tfMap["clustering_key"].([]interface{}); ok && len(v) > 0 {
		apiObject.ClusteringKeys = expandClusteringKeys(v)
	}

	if v, ok := tfMap["column"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AllColumns = expandColumnDefinitions(v.List())
	}

	if v, ok := tfMap["partition_key"].([]interface{}); ok && len(v) > 0 {
		apiObject.PartitionKeys = expandPartitionKeys(v)
	}

	if v, ok := tfMap["static_column"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.StaticColumns = expandStaticColumns(v.List())
	}

	return apiObject
}

func expandTimeToLive(tfMap map[string]interface{}) *types.TimeToLive {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.TimeToLive{}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		apiObject.Status = types.TimeToLiveStatus(v)
	}

	return apiObject
}

func expandColumnDefinition(tfMap map[string]interface{}) *types.ColumnDefinition {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ColumnDefinition{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandColumnDefinitions(tfList []interface{}) []types.ColumnDefinition {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.ColumnDefinition

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandColumnDefinition(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandClusteringKey(tfMap map[string]interface{}) *types.ClusteringKey {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ClusteringKey{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["order_by"].(string); ok && v != "" {
		apiObject.OrderBy = types.SortOrder(v)
	}

	return apiObject
}

func expandClusteringKeys(tfList []interface{}) []types.ClusteringKey {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.ClusteringKey

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandClusteringKey(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandPartitionKey(tfMap map[string]interface{}) *types.PartitionKey {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.PartitionKey{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func expandPartitionKeys(tfList []interface{}) []types.PartitionKey {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.PartitionKey

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandPartitionKey(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandStaticColumn(tfMap map[string]interface{}) *types.StaticColumn {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.StaticColumn{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func expandStaticColumns(tfList []interface{}) []types.StaticColumn {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.StaticColumn

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandStaticColumn(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenCapacitySpecificationSummary(apiObject *types.CapacitySpecificationSummary) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"throughput_mode": apiObject.ThroughputMode,
	}

	if v := apiObject.ReadCapacityUnits; v != nil {
		tfMap["read_capacity_units"] = aws.ToInt64(v)
	}

	if v := apiObject.WriteCapacityUnits; v != nil {
		tfMap["write_capacity_units"] = aws.ToInt64(v)
	}

	return tfMap
}

func flattenClientSideTimestamps(apiObject *types.ClientSideTimestamps) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrStatus: apiObject.Status,
	}

	return tfMap
}

func flattenComment(apiObject *types.Comment) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Message; v != nil {
		tfMap[names.AttrMessage] = aws.ToString(v)
	}

	return tfMap
}

func flattenEncryptionSpecification(apiObject *types.EncryptionSpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrType: apiObject.Type,
	}

	if v := apiObject.KmsKeyIdentifier; v != nil {
		tfMap["kms_key_identifier"] = aws.ToString(v)
	}

	return tfMap
}

func flattenPointInTimeRecoverySummary(apiObject *types.PointInTimeRecoverySummary) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrStatus: apiObject.Status,
	}

	return tfMap
}

func flattenSchemaDefinition(apiObject *types.SchemaDefinition) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AllColumns; v != nil {
		tfMap["column"] = flattenColumnDefinitions(v)
	}

	if v := apiObject.ClusteringKeys; v != nil {
		tfMap["clustering_key"] = flattenClusteringKeys(v)
	}

	if v := apiObject.PartitionKeys; v != nil {
		tfMap["partition_key"] = flattenPartitionKeys(v)
	}

	if v := apiObject.StaticColumns; v != nil {
		tfMap["static_column"] = flattenStaticColumns(v)
	}

	return tfMap
}

func flattenTimeToLive(apiObject *types.TimeToLive) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrStatus: apiObject.Status,
	}

	return tfMap
}

func flattenColumnDefinition(apiObject *types.ColumnDefinition) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.Type; v != nil {
		tfMap[names.AttrType] = aws.ToString(v)
	}

	return tfMap
}

func flattenColumnDefinitions(apiObjects []types.ColumnDefinition) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenColumnDefinition(&apiObject))
	}

	return tfList
}

func flattenClusteringKey(apiObject *types.ClusteringKey) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"order_by": apiObject.OrderBy,
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	return tfMap
}

func flattenClusteringKeys(apiObjects []types.ClusteringKey) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenClusteringKey(&apiObject))
	}

	return tfList
}

func flattenPartitionKey(apiObject *types.PartitionKey) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	return tfMap
}

func flattenPartitionKeys(apiObjects []types.PartitionKey) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenPartitionKey(&apiObject))
	}

	return tfList
}

func flattenStaticColumn(apiObject *types.StaticColumn) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	return tfMap
}

func flattenStaticColumns(apiObjects []types.StaticColumn) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenStaticColumn(&apiObject))
	}

	return tfList
}
