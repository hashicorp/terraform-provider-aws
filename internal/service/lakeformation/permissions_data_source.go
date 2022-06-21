package lakeformation

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourcePermissions() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePermissionsRead,

		Schema: map[string]*schema.Schema{
			"catalog_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"catalog_resource": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"data_location": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"catalog_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: verify.ValidAccountID,
						},
					},
				},
			},
			"database": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"catalog_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: verify.ValidAccountID,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"permissions_with_grant_option": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"lf_tag": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						"values": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							MaxItems: 15,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateLFTagValues(),
							},
							Set: schema.HashString,
						},
						"catalog_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"lf_tag_policy": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"catalog_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: verify.ValidAccountID,
						},
						"expression": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 5,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
									"values": {
										Type:     schema.TypeSet,
										Required: true,
										MinItems: 1,
										MaxItems: 15,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validateLFTagValues(),
										},
										Set: schema.HashString,
									},
								},
							},
						},
						"resource_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(lakeformation.ResourceType_Values(), false),
						},
					},
				},
			},
			"principal": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validPrincipal,
			},
			"table": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"catalog_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: verify.ValidAccountID,
						},
						"database_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"wildcard": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"table_with_columns": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"catalog_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: verify.ValidAccountID,
						},
						"column_names": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.NoZeroValues,
							},
						},
						"database_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"excluded_column_names": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.NoZeroValues,
							},
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"wildcard": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func dataSourcePermissionsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LakeFormationConn

	input := &lakeformation.ListPermissionsInput{
		Principal: &lakeformation.DataLakePrincipal{
			DataLakePrincipalIdentifier: aws.String(d.Get("principal").(string)),
		},
		Resource: &lakeformation.Resource{},
	}

	if v, ok := d.GetOk("catalog_id"); ok {
		input.CatalogId = aws.String(v.(string))
	}

	if _, ok := d.GetOk("catalog_resource"); ok {
		input.Resource.Catalog = ExpandCatalogResource()
	}

	if v, ok := d.GetOk("data_location"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Resource.DataLocation = ExpandDataLocationResource(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("database"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Resource.Database = ExpandDatabaseResource(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("lf_tag"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Resource.LFTag = ExpandLFTagKeyResource(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("lf_tag_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Resource.LFTagPolicy = ExpandLFTagPolicyResource(v.([]interface{})[0].(map[string]interface{}))
	}

	tableType := ""

	if v, ok := d.GetOk("table"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Resource.Table = ExpandTableResource(v.([]interface{})[0].(map[string]interface{}))
		tableType = TableTypeTable
	}

	if v, ok := d.GetOk("table_with_columns"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		// can't ListPermissions for TableWithColumns, so use Table instead
		input.Resource.Table = ExpandTableWithColumnsResourceAsTable(v.([]interface{})[0].(map[string]interface{}))
		tableType = TableTypeTableWithColumns
	}

	columnNames := make([]*string, 0)
	excludedColumnNames := make([]*string, 0)
	columnWildcard := false

	if tableType == TableTypeTableWithColumns {
		if v, ok := d.GetOk("table_with_columns.0.wildcard"); ok {
			columnWildcard = v.(bool)
		}

		if v, ok := d.GetOk("table_with_columns.0.column_names"); ok {
			if v, ok := v.(*schema.Set); ok && v.Len() > 0 {
				columnNames = flex.ExpandStringSet(v)
			}
		}

		if v, ok := d.GetOk("table_with_columns.0.excluded_column_names"); ok {
			if v, ok := v.(*schema.Set); ok && v.Len() > 0 {
				excludedColumnNames = flex.ExpandStringSet(v)
			}
		}
	}

	log.Printf("[DEBUG] Reading Lake Formation permissions: %v", input)

	allPermissions, err := waitPermissionsReady(conn, input, tableType, columnNames, excludedColumnNames, columnWildcard)

	d.SetId(fmt.Sprintf("%d", create.StringHashcode(input.String())))

	if err != nil {
		return fmt.Errorf("error reading Lake Formation permissions: %w", err)
	}

	// clean permissions = filter out permissions that do not pertain to this specific resource
	cleanPermissions := FilterPermissions(input, tableType, columnNames, excludedColumnNames, columnWildcard, allPermissions)

	if len(cleanPermissions) != len(allPermissions) {
		log.Printf("[INFO] Resource Lake Formation clean permissions (%d) and all permissions (%d) have different lengths (this is not necessarily a problem): %s", len(cleanPermissions), len(allPermissions), d.Id())
	}

	d.Set("principal", cleanPermissions[0].Principal.DataLakePrincipalIdentifier)
	d.Set("permissions", flattenPermissions(cleanPermissions))
	d.Set("permissions_with_grant_option", flattenGrantPermissions(cleanPermissions))

	if cleanPermissions[0].Resource.Catalog != nil {
		d.Set("catalog_resource", true)
	} else {
		d.Set("catalog_resource", false)
	}

	if cleanPermissions[0].Resource.DataLocation != nil {
		if err := d.Set("data_location", []interface{}{flattenDataLocationResource(cleanPermissions[0].Resource.DataLocation)}); err != nil {
			return fmt.Errorf("error setting data_location: %w", err)
		}
	} else {
		d.Set("data_location", nil)
	}

	if cleanPermissions[0].Resource.Database != nil {
		if err := d.Set("database", []interface{}{flattenDatabaseResource(cleanPermissions[0].Resource.Database)}); err != nil {
			return fmt.Errorf("error setting database: %w", err)
		}
	} else {
		d.Set("database", nil)
	}

	if cleanPermissions[0].Resource.LFTag != nil {
		if err := d.Set("lf_tag", []interface{}{flattenLFTagKeyResource(cleanPermissions[0].Resource.LFTag)}); err != nil {
			return fmt.Errorf("error setting LF-tag: %w", err)
		}
	} else {
		d.Set("lf_tag", nil)
	}

	if cleanPermissions[0].Resource.LFTagPolicy != nil {
		if err := d.Set("lf_tag_policy", []interface{}{flattenLFTagPolicyResource(cleanPermissions[0].Resource.LFTagPolicy)}); err != nil {
			return fmt.Errorf("error setting LF-tag policy: %w", err)
		}
	} else {
		d.Set("lf_tag_policy", nil)
	}

	tableSet := false

	if v, ok := d.GetOk("table"); ok && len(v.([]interface{})) > 0 {
		// since perm list could include TableWithColumns, get the right one
		for _, perm := range cleanPermissions {
			if perm.Resource == nil {
				continue
			}

			if perm.Resource.TableWithColumns != nil && perm.Resource.TableWithColumns.ColumnWildcard != nil {
				if err := d.Set("table", []interface{}{flattenTableColumnsResourceAsTable(perm.Resource.TableWithColumns)}); err != nil {
					return fmt.Errorf("error setting table: %w", err)
				}
				tableSet = true
				break
			}

			if perm.Resource.Table != nil {
				if err := d.Set("table", []interface{}{flattenTableResource(perm.Resource.Table)}); err != nil {
					return fmt.Errorf("error setting table: %w", err)
				}
				tableSet = true
				break
			}
		}
	}

	if !tableSet {
		d.Set("table", nil)
	}

	twcSet := false

	if v, ok := d.GetOk("table_with_columns"); ok && len(v.([]interface{})) > 0 {
		// since perm list could include Table, get the right one
		for _, perm := range cleanPermissions {
			if perm.Resource.TableWithColumns != nil {
				if err := d.Set("table_with_columns", []interface{}{flattenTableColumnsResource(perm.Resource.TableWithColumns)}); err != nil {
					return fmt.Errorf("error setting table_with_columns: %w", err)
				}
				twcSet = true
				break
			}
		}
	}

	if !twcSet {
		d.Set("table_with_columns", nil)
	}

	return nil
}
