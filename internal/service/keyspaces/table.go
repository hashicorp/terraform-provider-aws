package keyspaces

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/keyspaces"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTable() *schema.Resource {
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
			customdiff.ForceNewIfChange("schema_definition.0.column", func(_ context.Context, o, n, meta interface{}) bool {
				// Columns can only be added.
				if os, ok := o.(*schema.Set); ok {
					if ns, ok := n.(*schema.Set); ok {
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
			"arn": {
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
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(keyspaces.ThroughputMode_Values(), false),
						},
						"write_capacity_units": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
					},
				},
			},
			"comment": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"message": {
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
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(keyspaces.EncryptionType_Values(), false),
						},
					},
				},
			},
			"keyspace_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 48),
					validation.StringMatch(
						regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_]{1,47}$`),
						"The name must consist of alphanumerics and underscores.",
					),
				),
			},
			"point_in_time_recovery": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(keyspaces.PointInTimeRecoveryStatus_Values(), false),
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
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 48),
											validation.StringMatch(
												regexp.MustCompile(`^[a-z0-9][a-z0-9_]{1,47}$`),
												"The name must consist of lower case alphanumerics and underscores.",
											),
										),
									},
									"order_by": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(keyspaces.SortOrder_Values(), false),
									},
								},
							},
						},
						"column": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 48),
											validation.StringMatch(
												regexp.MustCompile(`^[a-z0-9][a-z0-9_]{1,47}$`),
												"The name must consist of lower case alphanumerics and underscores.",
											),
										),
									},
									"type": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringMatch(
											regexp.MustCompile(`^[a-z0-9]+(\<[a-z0-9]+(, *[a-z0-9]+){0,1}\>)?$`),
											"The type must consist of lower case alphanumerics and an optional list of upto two lower case alphanumerics enclosed in angle brackets '<>'.",
										),
									},
								},
							},
						},
						"partition_key": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 48),
											validation.StringMatch(
												regexp.MustCompile(`^[a-z0-9][a-z0-9_]{1,47}$`),
												"The name must consist of lower case alphanumerics and underscores.",
											),
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
									"name": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 48),
											validation.StringMatch(
												regexp.MustCompile(`^[a-z0-9][a-z0-9_]{1,47}$`),
												"The name must consist of lower case alphanumerics and underscores.",
											),
										),
									},
								},
							},
						},
					},
				},
			},
			"table_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 48),
					validation.StringMatch(
						regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_]{1,47}$`),
						"The name must consist of alphanumerics and underscores.",
					),
				),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"ttl": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(keyspaces.TimeToLiveStatus_Values(), false),
						},
					},
				},
			},
		},
	}
}

func resourceTableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KeyspacesConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	keyspaceName := d.Get("keyspace_name").(string)
	tableName := d.Get("table_name").(string)
	id := TableCreateResourceID(keyspaceName, tableName)
	input := &keyspaces.CreateTableInput{
		KeyspaceName: aws.String(keyspaceName),
		TableName:    aws.String(tableName),
	}

	if v, ok := d.GetOk("capacity_specification"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.CapacitySpecification = expandCapacitySpecification(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("comment"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Comment = expandComment(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("default_time_to_live"); ok {
		input.DefaultTimeToLive = aws.Int64(int64(v.(int)))
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

	if tags := Tags(tags.IgnoreAWS()); len(tags) > 0 {
		// The Keyspaces API requires that when Tags is set, it's non-empty.
		input.Tags = tags
	}

	log.Printf("[DEBUG] Creating Keyspaces Table: %s", input)
	_, err := conn.CreateTableWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Keyspaces Table (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitTableCreated(ctx, conn, keyspaceName, tableName, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Keyspaces Table (%s) create: %s", d.Id(), err)
	}

	return resourceTableRead(ctx, d, meta)
}

func resourceTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KeyspacesConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	keyspaceName, tableName, err := TableParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	table, err := FindTableByTwoPartKey(ctx, conn, keyspaceName, tableName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Keyspaces Table (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Keyspaces Table (%s): %s", d.Id(), err)
	}

	d.Set("arn", table.ResourceArn)
	if table.CapacitySpecification != nil {
		if err := d.Set("capacity_specification", []interface{}{flattenCapacitySpecificationSummary(table.CapacitySpecification)}); err != nil {
			return diag.Errorf("setting capacity_specification: %s", err)
		}
	} else {
		d.Set("capacity_specification", nil)
	}
	if table.Comment != nil {
		if err := d.Set("comment", []interface{}{flattenComment(table.Comment)}); err != nil {
			return diag.Errorf("setting comment: %s", err)
		}
	} else {
		d.Set("comment", nil)
	}
	d.Set("default_time_to_live", table.DefaultTimeToLive)
	if table.EncryptionSpecification != nil {
		if err := d.Set("encryption_specification", []interface{}{flattenEncryptionSpecification(table.EncryptionSpecification)}); err != nil {
			return diag.Errorf("setting encryption_specification: %s", err)
		}
	} else {
		d.Set("encryption_specification", nil)
	}
	d.Set("keyspace_name", table.KeyspaceName)
	if table.PointInTimeRecovery != nil {
		if err := d.Set("point_in_time_recovery", []interface{}{flattenPointInTimeRecoverySummary(table.PointInTimeRecovery)}); err != nil {
			return diag.Errorf("setting point_in_time_recovery: %s", err)
		}
	} else {
		d.Set("point_in_time_recovery", nil)
	}
	if table.SchemaDefinition != nil {
		if err := d.Set("schema_definition", []interface{}{flattenSchemaDefinition(table.SchemaDefinition)}); err != nil {
			return diag.Errorf("setting schema_definition: %s", err)
		}
	} else {
		d.Set("schema_definition", nil)
	}
	d.Set("table_name", table.TableName)
	if table.Ttl != nil {
		if err := d.Set("ttl", []interface{}{flattenTimeToLive(table.Ttl)}); err != nil {
			return diag.Errorf("setting ttl: %s", err)
		}
	} else {
		d.Set("ttl", nil)
	}

	tags, err := ListTags(conn, d.Get("arn").(string))

	if err != nil {
		return diag.Errorf("listing tags for Keyspaces Table (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceTableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KeyspacesConn

	keyspaceName, tableName, err := TableParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChangesExcept("tags", "tags_all") {
		// https://docs.aws.amazon.com/keyspaces/latest/APIReference/API_UpdateTable.html
		// Note that you can only update one specific table setting per update operation.
		if d.HasChange("capacity_specification") {
			if v, ok := d.GetOk("capacity_specification"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input := &keyspaces.UpdateTableInput{
					CapacitySpecification: expandCapacitySpecification(v.([]interface{})[0].(map[string]interface{})),
					KeyspaceName:          aws.String(keyspaceName),
					TableName:             aws.String(tableName),
				}

				log.Printf("[DEBUG] Updating Keyspaces Table: %s", input)
				_, err := conn.UpdateTableWithContext(ctx, input)

				if err != nil {
					return diag.Errorf("updating Keyspaces Table (%s) CapacitySpecification: %s", d.Id(), err)
				}

				if _, err := waitTableUpdated(ctx, conn, keyspaceName, tableName, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return diag.Errorf("waiting for Keyspaces Table (%s) CapacitySpecification update: %s", d.Id(), err)
				}
			}
		}

		if d.HasChange("default_time_to_live") {
			input := &keyspaces.UpdateTableInput{
				DefaultTimeToLive: aws.Int64(int64(d.Get("default_time_to_live").(int))),
				KeyspaceName:      aws.String(keyspaceName),
				TableName:         aws.String(tableName),
			}

			log.Printf("[DEBUG] Updating Keyspaces Table: %s", input)
			_, err := conn.UpdateTableWithContext(ctx, input)

			if err != nil {
				return diag.Errorf("updating Keyspaces Table (%s) DefaultTimeToLive: %s", d.Id(), err)
			}

			if _, err := waitTableUpdated(ctx, conn, keyspaceName, tableName, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return diag.Errorf("waiting for Keyspaces Table (%s) DefaultTimeToLive update: %s", d.Id(), err)
			}
		}

		if d.HasChange("encryption_specification") {
			if v, ok := d.GetOk("encryption_specification"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input := &keyspaces.UpdateTableInput{
					EncryptionSpecification: expandEncryptionSpecification(v.([]interface{})[0].(map[string]interface{})),
					KeyspaceName:            aws.String(keyspaceName),
					TableName:               aws.String(tableName),
				}

				log.Printf("[DEBUG] Updating Keyspaces Table: %s", input)
				_, err := conn.UpdateTableWithContext(ctx, input)

				if err != nil {
					return diag.Errorf("updating Keyspaces Table (%s) EncryptionSpecification: %s", d.Id(), err)
				}

				if _, err := waitTableUpdated(ctx, conn, keyspaceName, tableName, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return diag.Errorf("waiting for Keyspaces Table (%s) EncryptionSpecification update: %s", d.Id(), err)
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

				log.Printf("[DEBUG] Updating Keyspaces Table: %s", input)
				_, err := conn.UpdateTableWithContext(ctx, input)

				if err != nil {
					return diag.Errorf("updating Keyspaces Table (%s) PointInTimeRecovery: %s", d.Id(), err)
				}

				if _, err := waitTableUpdated(ctx, conn, keyspaceName, tableName, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return diag.Errorf("waiting for Keyspaces Table (%s) PointInTimeRecovery update: %s", d.Id(), err)
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

				log.Printf("[DEBUG] Updating Keyspaces Table: %s", input)
				_, err := conn.UpdateTableWithContext(ctx, input)

				if err != nil {
					return diag.Errorf("updating Keyspaces Table (%s) Ttl: %s", d.Id(), err)
				}

				if _, err := waitTableUpdated(ctx, conn, keyspaceName, tableName, d.Timeout(schema.TimeoutUpdate)); err != nil {
					return diag.Errorf("waiting for Keyspaces Table (%s) Ttl update: %s", d.Id(), err)
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

					log.Printf("[DEBUG] Updating Keyspaces Table: %s", input)
					_, err := conn.UpdateTableWithContext(ctx, input)

					if err != nil {
						return diag.Errorf("updating Keyspaces Table (%s) AddColumns: %s", d.Id(), err)
					}

					if _, err := waitTableUpdated(ctx, conn, keyspaceName, tableName, d.Timeout(schema.TimeoutUpdate)); err != nil {
						return diag.Errorf("waiting for Keyspaces Table (%s) AddColumns update: %s", d.Id(), err)
					}
				}
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("updating Keyspaces Table (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceTableRead(ctx, d, meta)
}

func resourceTableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KeyspacesConn

	keyspaceName, tableName, err := TableParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting Keyspaces Table: (%s)", d.Id())
	_, err = conn.DeleteTableWithContext(ctx, &keyspaces.DeleteTableInput{
		KeyspaceName: aws.String(keyspaceName),
		TableName:    aws.String(tableName),
	})

	if tfawserr.ErrCodeEquals(err, keyspaces.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Keyspaces Table (%s): %s", d.Id(), err)
	}

	if _, err := waitTableDeleted(ctx, conn, keyspaceName, tableName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for Keyspaces Table (%s) delete: %s", d.Id(), err)
	}

	return nil
}

const tableIDSeparator = "/"

func TableCreateResourceID(keyspaceName, tableName string) string {
	parts := []string{keyspaceName, tableName}
	id := strings.Join(parts, tableIDSeparator)

	return id
}

func TableParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, tableIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected KEYSPACE-NAME%[2]sTABLE-NAME", id, tableIDSeparator)
}

func statusTable(ctx context.Context, conn *keyspaces.Keyspaces, keyspaceName, tableName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTableByTwoPartKey(ctx, conn, keyspaceName, tableName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitTableCreated(ctx context.Context, conn *keyspaces.Keyspaces, keyspaceName, tableName string, timeout time.Duration) (*keyspaces.GetTableOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{keyspaces.TableStatusCreating},
		Target:  []string{keyspaces.TableStatusActive},
		Refresh: statusTable(ctx, conn, keyspaceName, tableName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*keyspaces.GetTableOutput); ok {
		return output, err
	}

	return nil, err
}

func waitTableDeleted(ctx context.Context, conn *keyspaces.Keyspaces, keyspaceName, tableName string, timeout time.Duration) (*keyspaces.GetTableOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{keyspaces.TableStatusActive, keyspaces.TableStatusDeleting},
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

func waitTableUpdated(ctx context.Context, conn *keyspaces.Keyspaces, keyspaceName, tableName string, timeout time.Duration) (*keyspaces.GetTableOutput, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{keyspaces.TableStatusUpdating},
		Target:  []string{keyspaces.TableStatusActive},
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

func expandCapacitySpecification(tfMap map[string]interface{}) *keyspaces.CapacitySpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &keyspaces.CapacitySpecification{}

	if v, ok := tfMap["read_capacity_units"].(int); ok && v != 0 {
		apiObject.ReadCapacityUnits = aws.Int64(int64(v))
	}

	if v, ok := tfMap["throughput_mode"].(string); ok && v != "" {
		apiObject.ThroughputMode = aws.String(v)
	}

	if v, ok := tfMap["write_capacity_units"].(int); ok && v != 0 {
		apiObject.WriteCapacityUnits = aws.Int64(int64(v))
	}

	return apiObject
}

func expandComment(tfMap map[string]interface{}) *keyspaces.Comment {
	if tfMap == nil {
		return nil
	}

	apiObject := &keyspaces.Comment{}

	if v, ok := tfMap["message"].(string); ok && v != "" {
		apiObject.Message = aws.String(v)
	}

	return apiObject
}

func expandEncryptionSpecification(tfMap map[string]interface{}) *keyspaces.EncryptionSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &keyspaces.EncryptionSpecification{}

	if v, ok := tfMap["kms_key_identifier"].(string); ok && v != "" {
		apiObject.KmsKeyIdentifier = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandPointInTimeRecovery(tfMap map[string]interface{}) *keyspaces.PointInTimeRecovery {
	if tfMap == nil {
		return nil
	}

	apiObject := &keyspaces.PointInTimeRecovery{}

	if v, ok := tfMap["status"].(string); ok && v != "" {
		apiObject.Status = aws.String(v)
	}

	return apiObject
}

func expandSchemaDefinition(tfMap map[string]interface{}) *keyspaces.SchemaDefinition {
	if tfMap == nil {
		return nil
	}

	apiObject := &keyspaces.SchemaDefinition{}

	if v, ok := tfMap["column"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AllColumns = expandColumnDefinitions(v.List())
	}

	if v, ok := tfMap["clustering_key"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ClusteringKeys = expandClusteringKeys(v.List())
	}

	if v, ok := tfMap["partition_key"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.PartitionKeys = expandPartitionKeys(v.List())
	}

	if v, ok := tfMap["static_column"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.StaticColumns = expandStaticColumns(v.List())
	}

	return apiObject
}

func expandTimeToLive(tfMap map[string]interface{}) *keyspaces.TimeToLive {
	if tfMap == nil {
		return nil
	}

	apiObject := &keyspaces.TimeToLive{}

	if v, ok := tfMap["status"].(string); ok && v != "" {
		apiObject.Status = aws.String(v)
	}

	return apiObject
}

func expandColumnDefinition(tfMap map[string]interface{}) *keyspaces.ColumnDefinition {
	if tfMap == nil {
		return nil
	}

	apiObject := &keyspaces.ColumnDefinition{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandColumnDefinitions(tfList []interface{}) []*keyspaces.ColumnDefinition {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*keyspaces.ColumnDefinition

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandColumnDefinition(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandClusteringKey(tfMap map[string]interface{}) *keyspaces.ClusteringKey {
	if tfMap == nil {
		return nil
	}

	apiObject := &keyspaces.ClusteringKey{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["order_by"].(string); ok && v != "" {
		apiObject.OrderBy = aws.String(v)
	}

	return apiObject
}

func expandClusteringKeys(tfList []interface{}) []*keyspaces.ClusteringKey {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*keyspaces.ClusteringKey

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandClusteringKey(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandPartitionKey(tfMap map[string]interface{}) *keyspaces.PartitionKey {
	if tfMap == nil {
		return nil
	}

	apiObject := &keyspaces.PartitionKey{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func expandPartitionKeys(tfList []interface{}) []*keyspaces.PartitionKey {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*keyspaces.PartitionKey

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandPartitionKey(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandStaticColumn(tfMap map[string]interface{}) *keyspaces.StaticColumn {
	if tfMap == nil {
		return nil
	}

	apiObject := &keyspaces.StaticColumn{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func expandStaticColumns(tfList []interface{}) []*keyspaces.StaticColumn {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*keyspaces.StaticColumn

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandStaticColumn(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenCapacitySpecificationSummary(apiObject *keyspaces.CapacitySpecificationSummary) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ReadCapacityUnits; v != nil {
		tfMap["read_capacity_units"] = aws.Int64Value(v)
	}

	if v := apiObject.ThroughputMode; v != nil {
		tfMap["throughput_mode"] = aws.StringValue(v)
	}

	if v := apiObject.WriteCapacityUnits; v != nil {
		tfMap["write_capacity_units"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenComment(apiObject *keyspaces.Comment) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Message; v != nil {
		tfMap["message"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenEncryptionSpecification(apiObject *keyspaces.EncryptionSpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.KmsKeyIdentifier; v != nil {
		tfMap["kms_key_identifier"] = aws.StringValue(v)
	}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenPointInTimeRecoverySummary(apiObject *keyspaces.PointInTimeRecoverySummary) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Status; v != nil {
		tfMap["status"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenSchemaDefinition(apiObject *keyspaces.SchemaDefinition) map[string]interface{} {
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

func flattenTimeToLive(apiObject *keyspaces.TimeToLive) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Status; v != nil {
		tfMap["status"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenColumnDefinition(apiObject *keyspaces.ColumnDefinition) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenColumnDefinitions(apiObjects []*keyspaces.ColumnDefinition) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenColumnDefinition(apiObject))
	}

	return tfList
}

func flattenClusteringKey(apiObject *keyspaces.ClusteringKey) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	if v := apiObject.OrderBy; v != nil {
		tfMap["order_by"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenClusteringKeys(apiObjects []*keyspaces.ClusteringKey) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenClusteringKey(apiObject))
	}

	return tfList
}

func flattenPartitionKey(apiObject *keyspaces.PartitionKey) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenPartitionKeys(apiObjects []*keyspaces.PartitionKey) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenPartitionKey(apiObject))
	}

	return tfList
}

func flattenStaticColumn(apiObject *keyspaces.StaticColumn) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenStaticColumns(apiObjects []*keyspaces.StaticColumn) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenStaticColumn(apiObject))
	}

	return tfList
}
