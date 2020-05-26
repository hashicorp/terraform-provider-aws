package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func AwsLakeFormationPermissions() []string {
	return []string{
		lakeformation.PermissionAll,
		lakeformation.PermissionSelect,
		lakeformation.PermissionAlter,
		lakeformation.PermissionDrop,
		lakeformation.PermissionDelete,
		lakeformation.PermissionInsert,
		lakeformation.PermissionCreateDatabase,
		lakeformation.PermissionCreateTable,
		lakeformation.PermissionDataLocationAccess,
	}
}

func resourceAwsLakeFormationPermissions() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLakeFormationPermissionsGrant,
		Read:   resourceAwsLakeFormationPermissionsList,
		Delete: resourceAwsLakeFormationPermissionsRevoke,

		Schema: map[string]*schema.Schema{
			"catalog_id": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateAwsAccountId,
			},
			"permissions": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(AwsLakeFormationPermissions(), false),
				},
			},
			"permissions_with_grant_option": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(AwsLakeFormationPermissions(), false),
				},
			},
			"principal": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"database": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"location", "table"},
			},
			"location": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ValidateFunc:  validateArn,
				ConflictsWith: []string{"database", "table"},
			},
			"table": {
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"database", "location"},
				MinItems:      0,
				MaxItems:      1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"column_names": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.NoZeroValues,
							},
						},
						"excluded_column_names": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.NoZeroValues,
							},
						},
					},
				},
			},
		},
	}
}

func resourceAwsLakeFormationPermissionsGrant(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn
	catalogId := createAwsDataCatalogId(d, meta.(*AWSClient).accountid)
	resource := expandAwsLakeFormationResource(d)

	input := &lakeformation.GrantPermissionsInput{
		CatalogId:   aws.String(catalogId),
		Permissions: expandStringList(d.Get("permissions").([]interface{})),
		Principal:   expandAwsLakeFormationPrincipal(d),
		Resource:    resource,
	}
	if vs, ok := d.GetOk("permissions_with_grant_option"); ok {
		input.PermissionsWithGrantOption = expandStringList(vs.([]interface{}))
	}

	_, err := conn.GrantPermissions(input)
	if err != nil {
		return fmt.Errorf("Error granting LakeFormation Permissions: %s", err)
	}

	d.SetId(fmt.Sprintf("lakeformation:resource:%s:%s", catalogId, time.Now().UTC().String()))
	d.Set("catalog_id", catalogId)

	return resourceAwsLakeFormationPermissionsList(d, meta)
}

func resourceAwsLakeFormationPermissionsList(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn
	catalogId := d.Get("catalog_id").(string)

	// This operation does not support getting privileges on a table with columns.
	// Instead, call this operation on the table, and the operation returns the
	// table and the table w columns.
	resource := expandAwsLakeFormationResource(d)
	isTableWithColumnsResource := false
	if table := resource.TableWithColumns; table != nil {
		resource.Table = &lakeformation.TableResource{
			DatabaseName: resource.TableWithColumns.DatabaseName,
			Name:         resource.TableWithColumns.Name,
		}
		resource.TableWithColumns = nil
		isTableWithColumnsResource = true
	}

	var resourceType string
	if table := resource.Catalog; table != nil {
		resourceType = lakeformation.DataLakeResourceTypeCatalog
	} else if location := resource.DataLocation; location != nil {
		resourceType = lakeformation.DataLakeResourceTypeDataLocation
	} else if DB := resource.Database; DB != nil {
		resourceType = lakeformation.DataLakeResourceTypeDatabase
	} else {
		resourceType = lakeformation.DataLakeResourceTypeTable
	}

	input := &lakeformation.ListPermissionsInput{
		CatalogId:    aws.String(catalogId),
		Principal:    expandAwsLakeFormationPrincipal(d),
		Resource:     resource,
		ResourceType: &resourceType,
	}

	out, err := conn.ListPermissions(input)
	if err != nil {
		return fmt.Errorf("Error listing LakeFormation Permissions: %s", err)
	}

	permissions := out.PrincipalResourcePermissions
	if len(permissions) == 0 {
		return fmt.Errorf("Error no LakeFormation Permissions found: %s", input)
	}

	// This operation does not support getting privileges on a table with columns.
	// Instead, call this operation on the table, and the operation returns the
	// table and the table w columns.
	if isTableWithColumnsResource {
		filtered := make([]*lakeformation.PrincipalResourcePermissions, 0)
		for _, p := range permissions {
			if table := p.Resource.TableWithColumns; table != nil {
				filtered = append(filtered, p)
			}
		}
		permissions = filtered
	}

	permissionsHead := permissions[0]
	d.Set("catalog_id", catalogId)
	d.Set("principal", permissionsHead.Principal.DataLakePrincipalIdentifier)
	if dataLocation := permissionsHead.Resource.DataLocation; dataLocation != nil {
		d.Set("location", dataLocation.ResourceArn)
	}
	if database := permissionsHead.Resource.Database; database != nil {
		d.Set("database", database.Name)
	}
	if table := permissionsHead.Resource.Table; table != nil {
		d.Set("table", flattenAWSLakeFormationTable(table))
	}
	if table := permissionsHead.Resource.TableWithColumns; table != nil {
		d.Set("table", flattenAWSLakeFormationTableWithColumns(table))
	}
	var allPermissions, allPermissionsWithGrant []*string
	for _, p := range permissions {
		allPermissions = append(allPermissions, p.Permissions...)
		allPermissionsWithGrant = append(allPermissionsWithGrant, p.PermissionsWithGrantOption...)
	}
	d.Set("permissions", allPermissions)
	d.Set("permissions_with_grant_option", allPermissionsWithGrant)

	return nil
}

func resourceAwsLakeFormationPermissionsRevoke(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn
	catalogId := d.Get("catalog_id").(string)

	input := &lakeformation.RevokePermissionsInput{
		CatalogId:   aws.String(catalogId),
		Permissions: expandStringList(d.Get("permissions").([]interface{})),
		Principal:   expandAwsLakeFormationPrincipal(d),
		Resource:    expandAwsLakeFormationResource(d),
	}
	if vs, ok := d.GetOk("permissions_with_grant_option"); ok {
		input.PermissionsWithGrantOption = expandStringList(vs.([]interface{}))
	}

	_, err := conn.RevokePermissions(input)
	if err != nil {
		return fmt.Errorf("Error revoking LakeFormation Permissions: %s", err)
	}

	return nil
}

func expandAwsLakeFormationPrincipal(d *schema.ResourceData) *lakeformation.DataLakePrincipal {
	return &lakeformation.DataLakePrincipal{
		DataLakePrincipalIdentifier: aws.String(d.Get("principal").(string)),
	}
}

func expandAwsLakeFormationResource(d *schema.ResourceData) *lakeformation.Resource {
	if v, ok := d.GetOk("database"); ok {
		databaseName := v.(string)
		if len(databaseName) > 0 {
			return &lakeformation.Resource{
				Database: &lakeformation.DatabaseResource{
					Name: aws.String(databaseName),
				},
			}
		}
	}
	if v, ok := d.GetOk("location"); ok {
		location := v.(string)
		if len(location) > 0 {
			return &lakeformation.Resource{
				DataLocation: &lakeformation.DataLocationResource{
					ResourceArn: aws.String(v.(string)),
				},
			}
		}
	}
	if vs, ok := d.GetOk("table"); ok {
		tables := vs.([]interface{})
		if len(tables) > 0 {
			table := tables[0].(map[string]interface{})

			resource := &lakeformation.Resource{}
			var databaseName, tableName string
			var columnNames, excludedColumnNames []interface{}
			if x, ok := table["database"]; ok {
				databaseName = x.(string)
			}
			if x, ok := table["name"]; ok {
				tableName = x.(string)
			}
			if xs, ok := table["column_names"]; ok {
				columnNames = xs.([]interface{})
			}
			if xs, ok := table["excluded_column_names"]; ok {
				excludedColumnNames = xs.([]interface{})
			}

			if len(columnNames) > 0 || len(excludedColumnNames) > 0 {
				tableWithColumns := &lakeformation.TableWithColumnsResource{
					DatabaseName: aws.String(databaseName),
					Name:         aws.String(tableName),
				}
				if len(columnNames) > 0 {
					tableWithColumns.ColumnNames = expandStringList(columnNames)
				}
				if len(excludedColumnNames) > 0 {
					tableWithColumns.ColumnWildcard = &lakeformation.ColumnWildcard{
						ExcludedColumnNames: expandStringList(excludedColumnNames),
					}
				}
				resource.TableWithColumns = tableWithColumns
			} else {
				resource.Table = &lakeformation.TableResource{
					DatabaseName: aws.String(databaseName),
					Name:         aws.String(tableName),
				}
			}
			return resource
		}
	}
	return &lakeformation.Resource{
		Catalog: &lakeformation.CatalogResource{},
	}
}

func flattenAWSLakeFormationTable(tb *lakeformation.TableResource) map[string]interface{} {
	m := make(map[string]interface{})

	m["database"] = tb.DatabaseName
	m["name"] = tb.Name

	return m
}

func flattenAWSLakeFormationTableWithColumns(tb *lakeformation.TableWithColumnsResource) map[string]interface{} {
	m := make(map[string]interface{})

	m["database"] = tb.DatabaseName
	m["name"] = tb.Name
	if columnNames := tb.ColumnNames; columnNames != nil {
		m["column_names"] = columnNames
	}
	if columnWildcard := tb.ColumnWildcard; columnWildcard != nil {
		m["excluded_column_names"] = columnWildcard.ExcludedColumnNames
	}

	return m
}
