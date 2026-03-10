// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package glue

import (
	"cmp"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/aws/aws-sdk-go-v2/service/glue/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tfsmithy "github.com/hashicorp/terraform-provider-aws/internal/smithy"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glue_catalog_table", name="Catalog Table")
func resourceCatalogTable() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCatalogTableCreate,
		ReadWithoutTimeout:   resourceCatalogTableRead,
		UpdateWithoutTimeout: resourceCatalogTableUpdate,
		DeleteWithoutTimeout: resourceCatalogTableDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCatalogID: {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			names.AttrDatabaseName: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringDoesNotMatch(regexache.MustCompile(`[A-Z]`), "uppercase characters cannot be used"),
				),
			},
			"open_table_format_input": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"iceberg_input": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"iceberg_table_input": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrLocation: {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 2056),
												},
												"partition_spec": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"fields": {
																Type:     schema.TypeList,
																Required: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"field_id": {
																			Type:     schema.TypeInt,
																			Optional: true,
																		},
																		names.AttrName: {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringLenBetween(1, 1024),
																		},
																		"source_id": {
																			Type:     schema.TypeInt,
																			Required: true,
																		},
																		"transform": {
																			Type:     schema.TypeString,
																			Required: true,
																		},
																	},
																},
															},
															"spec_id": {
																Type:     schema.TypeInt,
																Optional: true,
															},
														},
													},
												},
												names.AttrProperties: {
													Type:     schema.TypeMap,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"schema": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"fields": {
																Type:     schema.TypeList,
																Required: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"doc": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(0, 255),
																		},
																		names.AttrID: {
																			Type:     schema.TypeInt,
																			Required: true,
																		},
																		"initial_default": sdkv2.JSONDocumentSchemaOptional(),
																		names.AttrName: {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringLenBetween(1, 1024),
																		},
																		"required": {
																			Type:     schema.TypeBool,
																			Required: true,
																		},
																		names.AttrType:  sdkv2.JSONDocumentSchemaRequired(),
																		"write_default": sdkv2.JSONDocumentSchemaOptional(),
																	},
																},
															},
															"identifier_field_ids": {
																Type:     schema.TypeList,
																Optional: true,
																Elem:     &schema.Schema{Type: schema.TypeInt},
															},
															"schema_id": {
																Type:     schema.TypeInt,
																Optional: true,
															},
															names.AttrType: {
																Type:             schema.TypeString,
																Optional:         true,
																ValidateDiagFunc: enum.Validate[awstypes.IcebergStructTypeEnum](),
															},
														},
													},
												},
												"write_order": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"fields": {
																Type:     schema.TypeList,
																Required: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"direction": {
																			Type:             schema.TypeString,
																			Required:         true,
																			ValidateDiagFunc: enum.Validate[awstypes.IcebergSortDirection](),
																		},
																		"null_order": {
																			Type:             schema.TypeString,
																			Required:         true,
																			ValidateDiagFunc: enum.Validate[awstypes.IcebergNullOrder](),
																		},
																		"source_id": {
																			Type:     schema.TypeInt,
																			Required: true,
																		},
																		"transform": {
																			Type:     schema.TypeString,
																			Required: true,
																		},
																	},
																},
															},
															"order_id": {
																Type:     schema.TypeInt,
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"metadata_operation": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.MetadataOperation](),
									},
									names.AttrVersion: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
								},
							},
						},
					},
				},
			},
			names.AttrOwner: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrParameters: {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"partition_index": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 3,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"index_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"index_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"keys": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"partition_keys": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrComment: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						names.AttrName: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						names.AttrParameters: {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrType: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 131072),
						},
					},
				},
			},
			"retention": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"storage_descriptor": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"additional_locations": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"bucket_columns": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 255),
							},
						},
						"columns": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrComment: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 255),
									},
									names.AttrName: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									names.AttrParameters: {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									names.AttrType: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 131072),
									},
								},
							},
						},
						"compressed": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"input_format": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrLocation: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"number_of_buckets": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"output_format": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrParameters: {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ser_de_info": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									names.AttrParameters: {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"serialization_library": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
								},
							},
						},
						"schema_reference": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"schema_id": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"registry_name": {
													Type:          schema.TypeString,
													Optional:      true,
													ConflictsWith: []string{"storage_descriptor.0.schema_reference.0.schema_id.0.schema_arn"},
												},
												"schema_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
													ExactlyOneOf: []string{"storage_descriptor.0.schema_reference.0.schema_id.0.schema_arn", "storage_descriptor.0.schema_reference.0.schema_id.0.schema_name"},
												},
												"schema_name": {
													Type:         schema.TypeString,
													Optional:     true,
													ExactlyOneOf: []string{"storage_descriptor.0.schema_reference.0.schema_id.0.schema_arn", "storage_descriptor.0.schema_reference.0.schema_id.0.schema_name"},
												},
											},
										},
									},
									"schema_version_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ExactlyOneOf: []string{"storage_descriptor.0.schema_reference.0.schema_version_id", "storage_descriptor.0.schema_reference.0.schema_id"},
									},
									"schema_version_number": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(1, 100000),
									},
								},
							},
						},
						"skewed_info": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"skewed_column_names": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 255),
										},
									},
									"skewed_column_value_location_maps": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"skewed_column_values": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"sort_columns": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"column": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									"sort_order": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntInSlice([]int{0, 1}),
									},
								},
							},
						},
						"stored_as_sub_directories": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"table_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"target_table": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCatalogID: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrDatabaseName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrRegion: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"view_definition": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"definer": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"is_protected": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"last_refresh_type": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.LastRefreshType](),
						},
						"refresh_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"representations": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dialect": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ViewDialect](),
									},
									"dialect_version": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									"validation_connection": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									"view_expanded_text": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 409600),
									},
									"view_original_text": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 409600),
									},
								},
							},
						},
						"sub_object_version_ids": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeInt},
						},
						"sub_objects": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"view_version_id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"view_version_token": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"view_expanded_text": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 409600),
			},
			"view_original_text": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 409600),
			},
		},
	}
}

func resourceCatalogTableCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.GlueClient(ctx)

	catalogID, dbName, name := cmp.Or(d.Get(names.AttrCatalogID).(string), c.AccountID(ctx)), d.Get(names.AttrDatabaseName).(string), d.Get(names.AttrName).(string)
	id := catalogTableCreateResourceID(catalogID, dbName, name)
	input := glue.CreateTableInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
	}
	if v, ok := d.GetOk("open_table_format_input"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.OpenTableFormatInput = expandOpenTableFormatInput(v.([]any)[0].(map[string]any))
	}
	if v, ok := d.GetOk("partition_index"); ok && len(v.([]any)) > 0 {
		input.PartitionIndexes = expandPartitionIndexes(v.([]any))
	}

	if input.OpenTableFormatInput != nil && input.OpenTableFormatInput.IcebergInput != nil && input.OpenTableFormatInput.IcebergInput.CreateIcebergTableInput != nil {
		// "InvalidInputException: Location information cannot be null while creating an iceberg table".
		input.Name = aws.String(name)
	} else {
		// "InvalidInputException: CreateIcebergTableInput cannot be equal to null".
		input.TableInput = expandTableInput(d)
	}

	_, err := conn.CreateTable(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Catalog Table (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceCatalogTableRead(ctx, d, meta)...)
}

func resourceCatalogTableRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.GlueClient(ctx)

	catalogID, dbName, name, err := catalogTableParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	table, err := findTableByThreePartKey(ctx, conn, catalogID, dbName, name)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Glue Catalog Table (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Catalog Table (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, tableARN(ctx, c, dbName, name))
	d.Set(names.AttrCatalogID, catalogID)
	d.Set(names.AttrDatabaseName, dbName)
	d.Set(names.AttrDescription, table.Description)
	d.Set(names.AttrName, table.Name)
	d.Set(names.AttrOwner, table.Owner)
	if err := d.Set(names.AttrParameters, flattenNonManagedParameters(table.Parameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameters: %s", err)
	}
	if err := d.Set("partition_keys", flattenColumns(table.PartitionKeys)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting partition_keys: %s", err)
	}
	d.Set("retention", table.Retention)
	if err := d.Set("storage_descriptor", flattenStorageDescriptor(table.StorageDescriptor)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting storage_descriptor: %s", err)
	}
	d.Set("table_type", table.TableType)
	if table.TargetTable != nil {
		if err := d.Set("target_table", []any{flattenTableIdentifier(table.TargetTable)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting target_table: %s", err)
		}
	} else {
		d.Set("target_table", nil)
	}
	if table.ViewDefinition != nil {
		if err := d.Set("view_definition", []any{flattenViewDefinition(table.ViewDefinition)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting view_definition: %s", err)
		}
	} else {
		d.Set("view_definition", nil)
	}
	d.Set("view_expanded_text", table.ViewExpandedText)
	d.Set("view_original_text", table.ViewOriginalText)

	input := glue.GetPartitionIndexesInput{
		CatalogId:    table.CatalogId,
		DatabaseName: aws.String(dbName),
		TableName:    aws.String(name),
	}
	partitionIndexes, err := findPartitionIndexes(ctx, conn, &input)
	switch {
	// e.g. "InvalidInputException: Operation not supported on Multi Dialect Views".
	case errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "Operation not supported"):
		d.Set("partition_index", nil)
	case err != nil:
		return sdkdiag.AppendErrorf(diags, "reading Glue Catalog Table (%s) partition indexes: %s", d.Id(), err)
	default:
		if err := d.Set("partition_index", flattenPartitionIndexDescriptors(partitionIndexes)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting partition_index: %s", err)
		}
	}

	return diags
}

func resourceCatalogTableUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID, dbName, name, err := catalogTableParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// Add back any managed parameters. See flattenNonManagedParameters.
	table, err := findTableByThreePartKey(ctx, conn, catalogID, dbName, name)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Catalog Table (%s): %s", d.Id(), err)
	}

	input := glue.UpdateTableInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		TableInput:   expandTableInput(d),
	}
	if allParameters := table.Parameters; allParameters["table_type"] == "ICEBERG" {
		for _, k := range []string{"table_type", "metadata_location"} {
			if v := allParameters[k]; v != "" {
				if input.TableInput.Parameters == nil {
					input.TableInput.Parameters = make(map[string]string)
				}
				input.TableInput.Parameters[k] = v
			}
		}
	}

	_, err = conn.UpdateTable(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Glue Catalog Table (%s): %s", d.Id(), err)
	}

	return append(diags, resourceCatalogTableRead(ctx, d, meta)...)
}

func resourceCatalogTableDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID, dbName, name, err := catalogTableParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Glue Catalog Table: %s", d.Id())
	input := glue.DeleteTableInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		Name:         aws.String(name),
	}
	_, err = conn.DeleteTable(ctx, &input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Catalog Table (%s): %s", d.Id(), err)
	}

	return diags
}

const catalogTableResourceIDSeparator = ":"

func catalogTableCreateResourceID(catalogID, dbName, name string) string {
	parts := []string{catalogID, dbName, name}
	id := strings.Join(parts, catalogTableResourceIDSeparator)

	return id
}

func catalogTableParseResourceID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, catalogTableResourceIDSeparator, 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected catalog-id%[2]sdatabase-name%[2]stable-name", id, catalogTableResourceIDSeparator)
	}

	return parts[0], parts[1], parts[2], nil
}

func findTableByThreePartKey(ctx context.Context, conn *glue.Client, catalogID, dbName, name string) (*awstypes.Table, error) {
	input := glue.GetTableInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		Name:         aws.String(name),
	}

	return findTable(ctx, conn, &input)
}

func findTable(ctx context.Context, conn *glue.Client, input *glue.GetTableInput) (*awstypes.Table, error) {
	output, err := conn.GetTable(ctx, input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Table == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Table, nil
}

func expandTableInput(d *schema.ResourceData) *awstypes.TableInput {
	apiObject := &awstypes.TableInput{
		Name: aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		apiObject.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrOwner); ok {
		apiObject.Owner = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrParameters); ok {
		apiObject.Parameters = flex.ExpandStringValueMap(v.(map[string]any))
	}

	if v, ok := d.GetOk("partition_keys"); ok {
		apiObject.PartitionKeys = expandColumns(v.([]any))
	} else if _, ok = d.GetOk("open_table_format_input"); !ok {
		apiObject.PartitionKeys = []awstypes.Column{}
	}

	if v, ok := d.GetOk("retention"); ok {
		apiObject.Retention = int32(v.(int))
	}

	if v, ok := d.GetOk("storage_descriptor"); ok {
		apiObject.StorageDescriptor = expandStorageDescriptor(v.([]any))
	}

	if v, ok := d.GetOk("table_type"); ok {
		apiObject.TableType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("target_table"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		apiObject.TargetTable = expandTableIdentifier(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("view_definition"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		apiObject.ViewDefinition = expandViewDefinitionInput(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("view_expanded_text"); ok {
		apiObject.ViewExpandedText = aws.String(v.(string))
	}

	if v, ok := d.GetOk("view_original_text"); ok {
		apiObject.ViewOriginalText = aws.String(v.(string))
	}

	return apiObject
}

func expandOpenTableFormatInput(tfMap map[string]any) *awstypes.OpenTableFormatInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.OpenTableFormatInput{}

	if v, ok := tfMap["iceberg_input"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.IcebergInput = expandIcebergInput(v[0].(map[string]any))
	}

	return apiObject
}

func expandIcebergInput(tfMap map[string]any) *awstypes.IcebergInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.IcebergInput{}

	if v, ok := tfMap["iceberg_table_input"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.CreateIcebergTableInput = expandCreateIcebergTableInput(v[0].(map[string]any))
	}

	if v, ok := tfMap["metadata_operation"].(string); ok && v != "" {
		apiObject.MetadataOperation = awstypes.MetadataOperation(v)
	}

	if v, ok := tfMap[names.AttrVersion].(string); ok && v != "" {
		apiObject.Version = aws.String(v)
	}

	return apiObject
}

func expandCreateIcebergTableInput(tfMap map[string]any) *awstypes.CreateIcebergTableInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CreateIcebergTableInput{}

	if v, ok := tfMap[names.AttrLocation].(string); ok && v != "" {
		apiObject.Location = aws.String(v)
	}

	if v, ok := tfMap["partition_spec"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.PartitionSpec = expandIcebergPartitionSpec(v[0].(map[string]any))
	}

	if v, ok := tfMap[names.AttrProperties].(map[string]any); ok && len(v) > 0 {
		apiObject.Properties = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["schema"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.Schema = expandIcebergSchema(v[0].(map[string]any))
	}

	if v, ok := tfMap["write_order"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.WriteOrder = expandIcebergSortOrder(v[0].(map[string]any))
	}

	return apiObject
}

func expandIcebergPartitionSpec(tfMap map[string]any) *awstypes.IcebergPartitionSpec {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.IcebergPartitionSpec{}

	if v, ok := tfMap["fields"].([]any); ok && len(v) > 0 {
		apiObject.Fields = expandIcebergPartitionFields(v)
	}

	if v, ok := tfMap["spec_id"].(int); ok {
		apiObject.SpecId = int32(v)
	}

	return apiObject
}

func expandIcebergPartitionFields(tfList []any) []awstypes.IcebergPartitionField {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.IcebergPartitionField

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.IcebergPartitionField{}

		if v, ok := tfMap["field_id"].(int); ok {
			apiObject.FieldId = int32(v)
		}

		if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
			apiObject.Name = aws.String(v)
		}

		if v, ok := tfMap["source_id"].(int); ok {
			apiObject.SourceId = int32(v)
		}

		if v, ok := tfMap["transform"].(string); ok && v != "" {
			apiObject.Transform = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandIcebergSchema(tfMap map[string]any) *awstypes.IcebergSchema {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.IcebergSchema{}

	if v, ok := tfMap["fields"].([]any); ok && len(v) > 0 {
		apiObject.Fields = expandIcebergStructFields(v)
	}

	if v, ok := tfMap["identifier_field_ids"].([]any); ok && len(v) > 0 {
		apiObject.IdentifierFieldIds = flex.ExpandInt32ValueList(v)
	}

	if v, ok := tfMap["schema_id"].(int); ok {
		apiObject.SchemaId = int32(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.IcebergStructTypeEnum(v)
	}

	return apiObject
}

func expandIcebergStructFields(tfList []any) []awstypes.IcebergStructField {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.IcebergStructField

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.IcebergStructField{}

		if v, ok := tfMap["doc"].(string); ok && v != "" {
			apiObject.Doc = aws.String(v)
		}

		if v, ok := tfMap[names.AttrID].(int); ok {
			apiObject.Id = int32(v)
		}

		if v, ok := tfMap["initial_default"].(string); ok && v != "" {
			apiObject.InitialDefault, _ = tfsmithy.DocumentFromJSONString(v, document.NewLazyDocument)
		}

		if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
			apiObject.Name = aws.String(v)
		}

		if v, ok := tfMap["required"].(bool); ok {
			apiObject.Required = v
		}

		if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
			apiObject.Type, _ = tfsmithy.DocumentFromJSONString(v, document.NewLazyDocument)
		}

		if v, ok := tfMap["write_default"].(string); ok && v != "" {
			apiObject.WriteDefault, _ = tfsmithy.DocumentFromJSONString(v, document.NewLazyDocument)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandIcebergSortOrder(tfMap map[string]any) *awstypes.IcebergSortOrder {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.IcebergSortOrder{}

	if v, ok := tfMap["fields"].([]any); ok && len(v) > 0 {
		apiObject.Fields = expandIcebergSortFields(v)
	}

	if v, ok := tfMap["order_id"].(int); ok {
		apiObject.OrderId = int32(v)
	}

	return apiObject
}

func expandIcebergSortFields(tfList []any) []awstypes.IcebergSortField {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.IcebergSortField

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.IcebergSortField{}

		if v, ok := tfMap["direction"].(string); ok && v != "" {
			apiObject.Direction = awstypes.IcebergSortDirection(v)
		}

		if v, ok := tfMap["null_order"].(string); ok && v != "" {
			apiObject.NullOrder = awstypes.IcebergNullOrder(v)
		}

		if v, ok := tfMap["source_id"].(int); ok {
			apiObject.SourceId = int32(v)
		}

		if v, ok := tfMap["transform"].(string); ok && v != "" {
			apiObject.Transform = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandPartitionIndexes(tfList []any) []awstypes.PartitionIndex {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.PartitionIndex

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandPartitionIndex(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandStorageDescriptor(tfList []any) *awstypes.StorageDescriptor {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.StorageDescriptor{}

	if v, ok := tfMap["additional_locations"]; ok {
		apiObject.AdditionalLocations = flex.ExpandStringValueList(v.([]any))
	}

	if v, ok := tfMap["columns"]; ok {
		apiObject.Columns = expandColumns(v.([]any))
	}

	if v, ok := tfMap[names.AttrLocation]; ok {
		apiObject.Location = aws.String(v.(string))
	}

	if v, ok := tfMap["input_format"]; ok {
		apiObject.InputFormat = aws.String(v.(string))
	}

	if v, ok := tfMap["output_format"]; ok {
		apiObject.OutputFormat = aws.String(v.(string))
	}

	if v, ok := tfMap["compressed"]; ok {
		apiObject.Compressed = v.(bool)
	}

	if v, ok := tfMap["number_of_buckets"]; ok {
		apiObject.NumberOfBuckets = int32(v.(int))
	}

	if v, ok := tfMap["ser_de_info"]; ok {
		apiObject.SerdeInfo = expandSerDeInfo(v.([]any))
	}

	if v, ok := tfMap["bucket_columns"]; ok {
		apiObject.BucketColumns = flex.ExpandStringValueList(v.([]any))
	}

	if v, ok := tfMap["sort_columns"]; ok {
		apiObject.SortColumns = expandOrders(v.([]any))
	}

	if v, ok := tfMap["skewed_info"]; ok {
		apiObject.SkewedInfo = expandSkewedInfo(v.([]any))
	}

	if v, ok := tfMap[names.AttrParameters]; ok {
		apiObject.Parameters = flex.ExpandStringValueMap(v.(map[string]any))
	}

	if v, ok := tfMap["stored_as_sub_directories"]; ok {
		apiObject.StoredAsSubDirectories = v.(bool)
	}

	if v, ok := tfMap["schema_reference"]; ok && len(v.([]any)) > 0 {
		apiObject.Columns = nil
		apiObject.SchemaReference = expandSchemaReference(v.([]any))
	}

	return apiObject
}

func expandColumns(tfList []any) []awstypes.Column {
	apiObjects := []awstypes.Column{}
	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)

		column := awstypes.Column{
			Name: aws.String(tfMap[names.AttrName].(string)),
		}

		if v, ok := tfMap[names.AttrComment]; ok {
			column.Comment = aws.String(v.(string))
		}

		if v, ok := tfMap[names.AttrParameters]; ok {
			column.Parameters = flex.ExpandStringValueMap(v.(map[string]any))
		}

		if v, ok := tfMap[names.AttrType]; ok {
			column.Type = aws.String(v.(string))
		}

		apiObjects = append(apiObjects, column)
	}

	return apiObjects
}

func expandSerDeInfo(tfList []any) *awstypes.SerDeInfo {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.SerDeInfo{}

	if v := tfMap[names.AttrName]; len(v.(string)) > 0 {
		apiObject.Name = aws.String(v.(string))
	}

	if v := tfMap[names.AttrParameters]; len(v.(map[string]any)) > 0 {
		apiObject.Parameters = flex.ExpandStringValueMap(v.(map[string]any))
	}

	if v := tfMap["serialization_library"]; len(v.(string)) > 0 {
		apiObject.SerializationLibrary = aws.String(v.(string))
	}

	return apiObject
}

func expandOrders(tfList []any) []awstypes.Order {
	apiObjects := make([]awstypes.Order, len(tfList))

	for i, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)

		apiObject := awstypes.Order{
			Column: aws.String(tfMap["column"].(string)),
		}

		if v, ok := tfMap["sort_order"]; ok {
			apiObject.SortOrder = int32(v.(int))
		}

		apiObjects[i] = apiObject
	}

	return apiObjects
}

func expandSkewedInfo(tfList []any) *awstypes.SkewedInfo {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.SkewedInfo{}

	if v, ok := tfMap["skewed_column_names"]; ok {
		apiObject.SkewedColumnNames = flex.ExpandStringValueList(v.([]any))
	}

	if v, ok := tfMap["skewed_column_value_location_maps"]; ok {
		apiObject.SkewedColumnValueLocationMaps = flex.ExpandStringValueMap(v.(map[string]any))
	}

	if v, ok := tfMap["skewed_column_values"]; ok {
		apiObject.SkewedColumnValues = flex.ExpandStringValueList(v.([]any))
	}

	return apiObject
}

func expandSchemaReference(tfList []any) *awstypes.SchemaReference {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.SchemaReference{}

	if v, ok := tfMap["schema_version_id"].(string); ok && v != "" {
		apiObject.SchemaVersionId = aws.String(v)
	}

	if v, ok := tfMap["schema_id"]; ok {
		apiObject.SchemaId = expandSchemaId(v.([]any))
	}

	if v, ok := tfMap["schema_version_number"].(int); ok {
		apiObject.SchemaVersionNumber = aws.Int64(int64(v))
	}

	return apiObject
}

func expandSchemaId(tfList []any) *awstypes.SchemaId {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.SchemaId{}

	if v, ok := tfMap["registry_name"].(string); ok && v != "" {
		apiObject.RegistryName = aws.String(v)
	}

	if v, ok := tfMap["schema_name"].(string); ok && v != "" {
		apiObject.SchemaName = aws.String(v)
	}

	if v, ok := tfMap["schema_arn"].(string); ok && v != "" {
		apiObject.SchemaArn = aws.String(v)
	}

	return apiObject
}

func flattenStorageDescriptor(apiObject *awstypes.StorageDescriptor) []any {
	if apiObject == nil {
		return make([]any, 0)
	}

	tfList := make([]any, 1)
	tfMap := make(map[string]any)

	tfMap["additional_locations"] = apiObject.AdditionalLocations
	tfMap["columns"] = flattenColumns(apiObject.Columns)
	tfMap[names.AttrLocation] = aws.ToString(apiObject.Location)
	tfMap["input_format"] = aws.ToString(apiObject.InputFormat)
	tfMap["output_format"] = aws.ToString(apiObject.OutputFormat)
	tfMap["compressed"] = apiObject.Compressed
	tfMap["number_of_buckets"] = apiObject.NumberOfBuckets
	tfMap["ser_de_info"] = flattenSerDeInfo(apiObject.SerdeInfo)
	tfMap["bucket_columns"] = apiObject.BucketColumns
	tfMap["sort_columns"] = flattenOrders(apiObject.SortColumns)
	tfMap[names.AttrParameters] = apiObject.Parameters
	tfMap["skewed_info"] = flattenSkewedInfo(apiObject.SkewedInfo)
	tfMap["stored_as_sub_directories"] = apiObject.StoredAsSubDirectories

	if apiObject.SchemaReference != nil {
		tfMap["schema_reference"] = flattenSchemaReference(apiObject.SchemaReference)
	}

	tfList[0] = tfMap

	return tfList
}

func flattenColumns(apiObjects []awstypes.Column) []any {
	tfList := make([]any, len(apiObjects))
	if len(apiObjects) > 0 {
		for i, apiObject := range apiObjects {
			tfList[i] = flattenColumn(apiObject)
		}
	}

	return tfList
}

func flattenColumn(apiObject awstypes.Column) map[string]any {
	tfMap := make(map[string]any)

	if v := aws.ToString(apiObject.Comment); v != "" {
		tfMap[names.AttrComment] = v
	}

	if v := aws.ToString(apiObject.Name); v != "" {
		tfMap[names.AttrName] = v
	}

	if v := apiObject.Parameters; v != nil {
		tfMap[names.AttrParameters] = v
	}

	if v := aws.ToString(apiObject.Type); v != "" {
		tfMap[names.AttrType] = v
	}

	return tfMap
}

func flattenPartitionIndexDescriptors(apiObjects []awstypes.PartitionIndexDescriptor) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenPartitionIndexDescriptor(&apiObject))
	}

	return tfList
}

func flattenSerDeInfo(apiObject *awstypes.SerDeInfo) []any {
	if apiObject == nil {
		return make([]any, 0)
	}

	tfList := make([]any, 1)
	tfMap := make(map[string]any)

	if v := aws.ToString(apiObject.Name); v != "" {
		tfMap[names.AttrName] = v
	}
	tfMap[names.AttrParameters] = apiObject.Parameters
	if v := aws.ToString(apiObject.SerializationLibrary); v != "" {
		tfMap["serialization_library"] = v
	}

	tfList[0] = tfMap
	return tfList
}

func flattenOrders(apiObjects []awstypes.Order) []any {
	tfList := make([]any, len(apiObjects))
	for i, apiObject := range apiObjects {
		tfMap := make(map[string]any)
		tfMap["column"] = aws.ToString(apiObject.Column)
		tfMap["sort_order"] = apiObject.SortOrder
		tfList[i] = tfMap
	}

	return tfList
}

func flattenSkewedInfo(apiObject *awstypes.SkewedInfo) []any {
	if apiObject == nil {
		return make([]any, 0)
	}

	tfList := make([]any, 1)
	tfMap := make(map[string]any)

	tfMap["skewed_column_names"] = apiObject.SkewedColumnNames
	tfMap["skewed_column_value_location_maps"] = apiObject.SkewedColumnValueLocationMaps
	tfMap["skewed_column_values"] = apiObject.SkewedColumnValues
	tfList[0] = tfMap

	return tfList
}

func flattenSchemaReference(apiObject *awstypes.SchemaReference) []any {
	if apiObject == nil {
		return make([]any, 0)
	}

	tfList := make([]any, 1)
	tfMap := make(map[string]any)

	if apiObject.SchemaVersionId != nil {
		tfMap["schema_version_id"] = aws.ToString(apiObject.SchemaVersionId)
	}

	if apiObject.SchemaVersionNumber != nil {
		tfMap["schema_version_number"] = aws.ToInt64(apiObject.SchemaVersionNumber)
	}

	if apiObject.SchemaId != nil {
		tfMap["schema_id"] = flattenSchemaId(apiObject.SchemaId)
	}

	tfList[0] = tfMap

	return tfList
}

func flattenSchemaId(apiObject *awstypes.SchemaId) []any {
	if apiObject == nil {
		return make([]any, 0)
	}

	tfList := make([]any, 1)
	tfMap := make(map[string]any)

	if apiObject.RegistryName != nil {
		tfMap["registry_name"] = aws.ToString(apiObject.RegistryName)
	}

	if apiObject.SchemaArn != nil {
		tfMap["schema_arn"] = aws.ToString(apiObject.SchemaArn)
	}

	if apiObject.SchemaName != nil {
		tfMap["schema_name"] = aws.ToString(apiObject.SchemaName)
	}

	tfList[0] = tfMap

	return tfList
}

func expandTableIdentifier(tfMap map[string]any) *awstypes.TableIdentifier {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TableIdentifier{}

	if v, ok := tfMap[names.AttrCatalogID].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrDatabaseName].(string); ok && v != "" {
		apiObject.DatabaseName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap[names.AttrRegion].(string); ok && v != "" {
		apiObject.Region = aws.String(v)
	}

	return apiObject
}

func flattenTableIdentifier(apiObject *awstypes.TableIdentifier) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.CatalogId; v != nil {
		tfMap[names.AttrCatalogID] = aws.ToString(v)
	}

	if v := apiObject.DatabaseName; v != nil {
		tfMap[names.AttrDatabaseName] = aws.ToString(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.Region; v != nil {
		tfMap[names.AttrRegion] = aws.ToString(v)
	}

	return tfMap
}

func flattenNonManagedParameters(allParameters map[string]string) map[string]string {
	if allParameters["table_type"] == "ICEBERG" {
		delete(allParameters, "table_type")
		delete(allParameters, "metadata_location")
	}
	return allParameters
}

func expandViewDefinitionInput(tfMap map[string]any) *awstypes.ViewDefinitionInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ViewDefinitionInput{}

	if v, ok := tfMap["definer"].(string); ok && v != "" {
		apiObject.Definer = aws.String(v)
	}

	if v, ok := tfMap["is_protected"].(bool); ok && v {
		apiObject.IsProtected = aws.Bool(v)
	}

	if v, ok := tfMap["last_refresh_type"].(string); ok && v != "" {
		apiObject.LastRefreshType = awstypes.LastRefreshType(v)
	}

	if v, ok := tfMap["refresh_seconds"].(int); ok && v != 0 {
		apiObject.RefreshSeconds = aws.Int64(int64(v))
	}

	if v, ok := tfMap["representations"].([]any); ok && len(v) > 0 {
		apiObject.Representations = expandViewRepresentationInputs(v)
	}

	if v, ok := tfMap["sub_object_version_ids"].([]any); ok && len(v) > 0 {
		apiObject.SubObjectVersionIds = flex.ExpandInt64ValueList(v)
	}

	if v, ok := tfMap["sub_objects"].([]any); ok && len(v) > 0 {
		apiObject.SubObjects = flex.ExpandStringValueList(v)
	}

	if v, ok := tfMap["view_version_id"].(int); ok && v != 0 {
		apiObject.ViewVersionId = int64(v)
	}

	if v, ok := tfMap["view_version_token"].(string); ok && v != "" {
		apiObject.ViewVersionToken = aws.String(v)
	}

	return apiObject
}

func expandViewRepresentationInputs(tfList []any) []awstypes.ViewRepresentationInput {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ViewRepresentationInput

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.ViewRepresentationInput{}

		if v, ok := tfMap["dialect"].(string); ok && v != "" {
			apiObject.Dialect = awstypes.ViewDialect(v)
		}

		if v, ok := tfMap["dialect_version"].(string); ok && v != "" {
			apiObject.DialectVersion = aws.String(v)
		}

		if v, ok := tfMap["validation_connection"].(string); ok && v != "" {
			apiObject.ValidationConnection = aws.String(v)
		}

		if v, ok := tfMap["view_expanded_text"].(string); ok && v != "" {
			apiObject.ViewExpandedText = aws.String(v)
		}

		if v, ok := tfMap["view_original_text"].(string); ok && v != "" {
			apiObject.ViewOriginalText = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenViewDefinition(apiObject *awstypes.ViewDefinition) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := aws.ToString(apiObject.Definer); v != "" {
		tfMap["definer"] = v
	}

	if v := apiObject.IsProtected; v != nil {
		tfMap["is_protected"] = aws.ToBool(v)
	}

	if v := apiObject.LastRefreshType; v != "" {
		tfMap["last_refresh_type"] = v
	}

	if v := apiObject.RefreshSeconds; v != nil {
		tfMap["refresh_seconds"] = aws.ToInt64(v)
	}

	if v := apiObject.Representations; len(v) > 0 {
		tfMap["representations"] = flattenViewRepresentations(v)
	}

	tfMap["view_version_id"] = apiObject.ViewVersionId

	if v := aws.ToString(apiObject.ViewVersionToken); v != "" {
		tfMap["view_version_token"] = v
	}

	return tfMap
}

func flattenViewRepresentations(apiObjects []awstypes.ViewRepresentation) []any {
	tfList := make([]any, len(apiObjects))
	for i, v := range apiObjects {
		tfList[i] = flattenViewRepresentation(v)
	}
	return tfList
}

func flattenViewRepresentation(apiObject awstypes.ViewRepresentation) map[string]any {
	tfMap := make(map[string]any)

	if v := apiObject.Dialect; v != "" {
		tfMap["dialect"] = v
	}

	if v := aws.ToString(apiObject.DialectVersion); v != "" {
		tfMap["dialect_version"] = v
	}

	if v := aws.ToString(apiObject.ValidationConnection); v != "" {
		tfMap["validation_connection"] = v
	}

	if v := aws.ToString(apiObject.ViewExpandedText); v != "" {
		tfMap["view_expanded_text"] = v
	}

	if v := aws.ToString(apiObject.ViewOriginalText); v != "" {
		tfMap["view_original_text"] = v
	}

	return tfMap
}

func tableARN(ctx context.Context, c *conns.AWSClient, dbName, name string) string {
	return c.RegionalARN(ctx, "glue", "table/"+dbName+"/"+name)
}
