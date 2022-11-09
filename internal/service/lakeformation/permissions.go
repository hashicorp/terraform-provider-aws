package lakeformation

import (
	"fmt"
	"log"
	"reflect"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePermissions() *schema.Resource {
	return &schema.Resource{
		Create: resourcePermissionsCreate,
		Read:   resourcePermissionsRead,
		Delete: resourcePermissionsDelete,

		Schema: map[string]*schema.Schema{
			"catalog_id": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"catalog_resource": {
				Type:     schema.TypeBool,
				Default:  false,
				ForceNew: true,
				Optional: true,
				ExactlyOneOf: []string{
					"catalog_resource",
					"data_location",
					"database",
					"lf_tag",
					"lf_tag_policy",
					"table",
					"table_with_columns",
				},
			},
			"data_location": {
				Type:     schema.TypeList,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Optional: true,
				ExactlyOneOf: []string{
					"catalog_resource",
					"data_location",
					"database",
					"lf_tag",
					"lf_tag_policy",
					"table",
					"table_with_columns",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"catalog_id": {
							Type:         schema.TypeString,
							Computed:     true,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: verify.ValidAccountID,
						},
					},
				},
			},
			"database": {
				Type:     schema.TypeList,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Optional: true,
				ExactlyOneOf: []string{
					"catalog_resource",
					"data_location",
					"database",
					"lf_tag",
					"lf_tag_policy",
					"table",
					"table_with_columns",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"catalog_id": {
							Type:         schema.TypeString,
							Computed:     true,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: verify.ValidAccountID,
						},
						"name": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
					},
				},
			},
			"lf_tag": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				ExactlyOneOf: []string{
					"catalog_resource",
					"data_location",
					"database",
					"lf_tag",
					"lf_tag_policy",
					"table",
					"table_with_columns",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
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
							ForceNew: true,
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
				ExactlyOneOf: []string{
					"catalog_resource",
					"data_location",
					"database",
					"lf_tag",
					"lf_tag_policy",
					"table",
					"table_with_columns",
				},
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
			"permissions": {
				Type:     schema.TypeList,
				ForceNew: true,
				MinItems: 1,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(lakeformation.Permission_Values(), false),
				},
			},
			"permissions_with_grant_option": {
				Type:     schema.TypeList,
				Computed: true,
				ForceNew: true,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(lakeformation.Permission_Values(), false),
				},
			},
			"principal": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validPrincipal,
			},
			"table": {
				Type:     schema.TypeList,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Optional: true,
				ExactlyOneOf: []string{
					"catalog_resource",
					"data_location",
					"database",
					"lf_tag",
					"lf_tag_policy",
					"table",
					"table_with_columns",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"catalog_id": {
							Type:         schema.TypeString,
							Computed:     true,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: verify.ValidAccountID,
						},
						"database_name": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
							ForceNew: true,
							Optional: true,
							AtLeastOneOf: []string{
								"table.0.name",
								"table.0.wildcard",
							},
						},
						"wildcard": {
							Type:     schema.TypeBool,
							Default:  false,
							ForceNew: true,
							Optional: true,
							AtLeastOneOf: []string{
								"table.0.name",
								"table.0.wildcard",
							},
						},
					},
				},
			},
			"table_with_columns": {
				Type:     schema.TypeList,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Optional: true,
				ExactlyOneOf: []string{
					"catalog_resource",
					"data_location",
					"database",
					"lf_tag",
					"lf_tag_policy",
					"table",
					"table_with_columns",
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"catalog_id": {
							Type:         schema.TypeString,
							Computed:     true,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: verify.ValidAccountID,
						},
						"column_names": {
							Type:     schema.TypeSet,
							ForceNew: true,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.NoZeroValues,
							},
							AtLeastOneOf: []string{
								"table_with_columns.0.column_names",
								"table_with_columns.0.wildcard",
							},
						},
						"database_name": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
						"excluded_column_names": {
							Type:     schema.TypeSet,
							ForceNew: true,
							Optional: true,
							Set:      schema.HashString,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.NoZeroValues,
							},
						},
						"name": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
						"wildcard": {
							Type:     schema.TypeBool,
							Default:  false,
							ForceNew: true,
							Optional: true,
							AtLeastOneOf: []string{
								"table_with_columns.0.column_names",
								"table_with_columns.0.wildcard",
							},
						},
					},
				},
			},
		},
	}
}

// The challenges with Lake Formation permissions are many. These are largely undocumented and
// discovered through trial and error. These are specific problems discovered thus far:
// 1. Implicit permissions granted by Lake Formation to data lake administrators are indistinguishable
//    from explicit permissions. However, implicit permissions cannot be changed, revoked, or narrowed.
// 2. One set of permissions for one LF Resource going in, can come back from AWS as multiple sets of
//    permissions for multiple LF Resources (e.g., SELECT, Table, TableWithColumns).
// 3. Valid permissions for a Table LF resource can come back in TableWithColumns and vice versa.

// For 2 & 3, some peeking at the config (i.e., d.Get()) is necessary to filter the permissions AWS
// returns.

func resourcePermissionsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LakeFormationConn

	input := &lakeformation.GrantPermissionsInput{
		Permissions: flex.ExpandStringList(d.Get("permissions").([]interface{})),
		Principal: &lakeformation.DataLakePrincipal{
			DataLakePrincipalIdentifier: aws.String(d.Get("principal").(string)),
		},
		Resource: &lakeformation.Resource{},
	}

	if v, ok := d.GetOk("catalog_id"); ok {
		input.CatalogId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("permissions_with_grant_option"); ok {
		input.PermissionsWithGrantOption = flex.ExpandStringList(v.([]interface{}))
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

	if v, ok := d.GetOk("table"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Resource.Table = ExpandTableResource(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("table_with_columns"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Resource.TableWithColumns = expandTableColumnsResource(v.([]interface{})[0].(map[string]interface{}))
	}

	var output *lakeformation.GrantPermissionsOutput
	err := resource.Retry(IAMPropagationTimeout, func() *resource.RetryError {
		var err error
		output, err = conn.GrantPermissions(input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, lakeformation.ErrCodeInvalidInputException, "Invalid principal") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, lakeformation.ErrCodeInvalidInputException, "Grantee has no permissions") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, lakeformation.ErrCodeInvalidInputException, "register the S3 path") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeConcurrentModificationException) {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, "AccessDeniedException", "is not authorized to access requested permissions") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(fmt.Errorf("error creating Lake Formation Permissions: %w", err))
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.GrantPermissions(input)
	}

	if err != nil {
		return fmt.Errorf("error creating Lake Formation Permissions (input: %v): %w", input, err)
	}

	if output == nil {
		return fmt.Errorf("error creating Lake Formation Permissions: empty response")
	}

	d.SetId(fmt.Sprintf("%d", create.StringHashcode(input.String())))

	return resourcePermissionsRead(d, meta)
}

func resourcePermissionsRead(d *schema.ResourceData, meta interface{}) error {
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

	if !d.IsNewResource() {
		if tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeEntityNotFoundException) {
			log.Printf("[WARN] Resource Lake Formation permissions (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		if tfawserr.ErrMessageContains(err, "AccessDeniedException", "Resource does not exist") {
			log.Printf("[WARN] Resource Lake Formation permissions (%s) not found, removing from state: %s", d.Id(), err)
			d.SetId("")
			return nil
		}

		if len(allPermissions) == 0 {
			log.Printf("[WARN] Resource Lake Formation permissions (%s) not found, removing from state (0 permissions)", d.Id())
			d.SetId("")
			return nil
		}
	}

	if err != nil {
		return fmt.Errorf("error reading Lake Formation permissions: %w", err)
	}

	// clean permissions = filter out permissions that do not pertain to this specific resource
	cleanPermissions := FilterPermissions(input, tableType, columnNames, excludedColumnNames, columnWildcard, allPermissions)

	if len(cleanPermissions) == 0 {
		log.Printf("[WARN] No Lake Formation permissions (%s) found", d.Id())
		d.Set("catalog_resource", false)
		d.Set("data_location", nil)
		d.Set("database", nil)
		d.Set("lf_tag", nil)
		d.Set("lf_tag_policy", nil)
		d.Set("table_with_columns", nil)
		d.Set("table", nil)
		return nil
	}

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
			return fmt.Errorf("error setting database: %w", err)
		}
	} else {
		d.Set("lf_tag", nil)
	}

	if cleanPermissions[0].Resource.LFTagPolicy != nil {
		if err := d.Set("lf_tag_policy", []interface{}{flattenLFTagPolicyResource(cleanPermissions[0].Resource.LFTagPolicy)}); err != nil {
			return fmt.Errorf("error setting database: %w", err)
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

func resourcePermissionsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LakeFormationConn

	input := &lakeformation.RevokePermissionsInput{
		Permissions:                flex.ExpandStringList(d.Get("permissions").([]interface{})),
		PermissionsWithGrantOption: flex.ExpandStringList(d.Get("permissions_with_grant_option").([]interface{})),
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

	if v, ok := d.GetOk("table"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Resource.Table = ExpandTableResource(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("table_with_columns"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Resource.TableWithColumns = expandTableColumnsResource(v.([]interface{})[0].(map[string]interface{}))
	}

	if input.Resource == nil || reflect.DeepEqual(input.Resource, &lakeformation.Resource{}) {
		// if resource is empty, don't delete = it won't delete anything since this is the predicate
		log.Printf("[WARN] No Lake Formation Resource with permissions to revoke")
		return nil
	}

	err := resource.Retry(permissionsDeleteRetryTimeout, func() *resource.RetryError {
		var err error
		_, err = conn.RevokePermissions(input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, lakeformation.ErrCodeInvalidInputException, "register the S3 path") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeConcurrentModificationException) {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, "AccessDeniedException", "is not authorized to access requested permissions") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(fmt.Errorf("unable to revoke Lake Formation Permissions: %w", err))
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.RevokePermissions(input)
	}

	if tfawserr.ErrMessageContains(err, lakeformation.ErrCodeInvalidInputException, "No permissions revoked. Grantee") {
		return nil
	}

	if tfawserr.ErrMessageContains(err, lakeformation.ErrCodeInvalidInputException, "cannot grant/revoke permission on non-existent column") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("unable to revoke LakeFormation Permissions (input: %v): %w", input, err)
	}

	// Attempted to add a waiter here to wait for delete to complete. However, ListPermissions() returns
	// permissions, at least for catalog/CREATE_DATABASE permission, even if they do not exist. That makes
	// knowing when the delete is complete impossible. Instead, we'll retry until we get the right error.

	// Knowing when the delete is complete is complicated:
	// You can't just wait until permissions = 0 because there could be many other unrelated permissions
	// on the resource and filtering is non-trivial for table with columns.

	err = resource.Retry(permissionsDeleteRetryTimeout, func() *resource.RetryError {
		var err error
		_, err = conn.RevokePermissions(input)

		if !tfawserr.ErrMessageContains(err, lakeformation.ErrCodeInvalidInputException, "No permissions revoked. Grantee has no") {
			return resource.RetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.RevokePermissions(input)
	}

	if tfawserr.ErrMessageContains(err, lakeformation.ErrCodeInvalidInputException, "No permissions revoked. Grantee") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("unable to revoke LakeFormation Permissions (input: %v): %w", input, err)
	}

	return nil
}

func ExpandCatalogResource() *lakeformation.CatalogResource {
	return &lakeformation.CatalogResource{}
}

func ExpandDataLocationResource(tfMap map[string]interface{}) *lakeformation.DataLocationResource {
	if tfMap == nil {
		return nil
	}

	apiObject := &lakeformation.DataLocationResource{}

	if v, ok := tfMap["catalog_id"].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap["arn"].(string); ok && v != "" {
		apiObject.ResourceArn = aws.String(v)
	}

	return apiObject
}

func flattenDataLocationResource(apiObject *lakeformation.DataLocationResource) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CatalogId; v != nil {
		tfMap["catalog_id"] = aws.StringValue(v)
	}

	if v := apiObject.ResourceArn; v != nil {
		tfMap["arn"] = aws.StringValue(v)
	}

	return tfMap
}

func ExpandDatabaseResource(tfMap map[string]interface{}) *lakeformation.DatabaseResource {
	if tfMap == nil {
		return nil
	}

	apiObject := &lakeformation.DatabaseResource{}

	if v, ok := tfMap["catalog_id"].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func flattenDatabaseResource(apiObject *lakeformation.DatabaseResource) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CatalogId; v != nil {
		tfMap["catalog_id"] = aws.StringValue(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	return tfMap
}

func ExpandLFTagPolicyResource(tfMap map[string]interface{}) *lakeformation.LFTagPolicyResource {
	if tfMap == nil {
		return nil
	}

	apiObject := &lakeformation.LFTagPolicyResource{}

	if v, ok := tfMap["catalog_id"].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap["expression"]; ok && v != nil {
		apiObject.Expression = ExpandLFTagExpression(v.([]interface{}))
	}

	if v, ok := tfMap["resource_type"].(string); ok && v != "" {
		apiObject.ResourceType = aws.String(v)
	}

	return apiObject
}

func ExpandLFTagExpression(expression []interface{}) []*lakeformation.LFTag {
	tagSlice := []*lakeformation.LFTag{}
	for _, element := range expression {
		elementMap := element.(map[string]interface{})

		tag := &lakeformation.LFTag{
			TagKey:    aws.String(elementMap["key"].(string)),
			TagValues: flex.ExpandStringSet(elementMap["values"].(*schema.Set)),
		}

		tagSlice = append(tagSlice, tag)
	}

	return tagSlice
}

func flattenLFTagPolicyResource(apiObject *lakeformation.LFTagPolicyResource) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CatalogId; v != nil {
		tfMap["catalog_id"] = aws.StringValue(v)
	}

	if v := apiObject.Expression; v != nil {
		tfMap["expression"] = flattenLFTagExpression(v)
	}

	if v := apiObject.ResourceType; v != nil {
		tfMap["resource_type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenLFTagExpression(ts []*lakeformation.LFTag) []map[string]interface{} {
	tagSlice := make([]map[string]interface{}, len(ts))
	if len(ts) > 0 {
		for i, t := range ts {
			tag := make(map[string]interface{})

			if v := aws.StringValue(t.TagKey); v != "" {
				tag["key"] = v
			}

			if v := flex.FlattenStringList(t.TagValues); v != nil {
				tag["values"] = v
			}

			tagSlice[i] = tag
		}
	}

	return tagSlice
}

func ExpandLFTagKeyResource(tfMap map[string]interface{}) *lakeformation.LFTagKeyResource {
	if tfMap == nil {
		return nil
	}

	apiObject := &lakeformation.LFTagKeyResource{}

	if v, ok := tfMap["catalog_id"].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap["key"].(string); ok && v != "" {
		apiObject.TagKey = aws.String(v)
	}

	if v, ok := tfMap["values"].(*schema.Set); ok && v != nil {
		apiObject.TagValues = flex.ExpandStringSet(v)
	}

	return apiObject
}

func flattenLFTagKeyResource(apiObject *lakeformation.LFTagKeyResource) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CatalogId; v != nil {
		tfMap["catalog_id"] = aws.StringValue(v)
	}

	if v := apiObject.TagKey; v != nil {
		tfMap["key"] = aws.StringValue(v)
	}

	if v := apiObject.TagValues; v != nil {
		tfMap["values"] = flex.FlattenStringSet(v)
	}

	return tfMap
}

func ExpandTableResource(tfMap map[string]interface{}) *lakeformation.TableResource {
	if tfMap == nil {
		return nil
	}

	apiObject := &lakeformation.TableResource{}

	if v, ok := tfMap["catalog_id"].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap["database_name"].(string); ok && v != "" {
		apiObject.DatabaseName = aws.String(v)
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["wildcard"].(bool); ok && v {
		apiObject.TableWildcard = &lakeformation.TableWildcard{}
	}

	return apiObject
}

func ExpandTableWithColumnsResourceAsTable(tfMap map[string]interface{}) *lakeformation.TableResource {
	if tfMap == nil {
		return nil
	}

	apiObject := &lakeformation.TableResource{}

	if v, ok := tfMap["catalog_id"].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap["database_name"].(string); ok && v != "" {
		apiObject.DatabaseName = aws.String(v)
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func flattenTableResource(apiObject *lakeformation.TableResource) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CatalogId; v != nil {
		tfMap["catalog_id"] = aws.StringValue(v)
	}

	if v := apiObject.DatabaseName; v != nil {
		tfMap["database_name"] = aws.StringValue(v)
	}

	if v := apiObject.Name; v != nil {
		if aws.StringValue(v) != TableNameAllTables || apiObject.TableWildcard == nil {
			tfMap["name"] = aws.StringValue(v)
		}
	}

	if v := apiObject.TableWildcard; v != nil {
		tfMap["wildcard"] = true
	}

	return tfMap
}

func expandTableColumnsResource(tfMap map[string]interface{}) *lakeformation.TableWithColumnsResource {
	if tfMap == nil {
		return nil
	}

	apiObject := &lakeformation.TableWithColumnsResource{}

	if v, ok := tfMap["catalog_id"].(string); ok && v != "" {
		apiObject.CatalogId = aws.String(v)
	}

	if v, ok := tfMap["column_names"]; ok {
		if v, ok := v.(*schema.Set); ok && v.Len() > 0 {
			apiObject.ColumnNames = flex.ExpandStringSet(v)
		}
	}

	if v, ok := tfMap["database_name"].(string); ok && v != "" {
		apiObject.DatabaseName = aws.String(v)
	}

	if v, ok := tfMap["excluded_column_names"]; ok {
		if v, ok := v.(*schema.Set); ok && v.Len() > 0 {
			apiObject.ColumnWildcard = &lakeformation.ColumnWildcard{
				ExcludedColumnNames: flex.ExpandStringSet(v),
			}
		}
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["wildcard"].(bool); ok && v && apiObject.ColumnWildcard == nil {
		apiObject.ColumnWildcard = &lakeformation.ColumnWildcard{}
	}

	return apiObject
}

func flattenTableColumnsResource(apiObject *lakeformation.TableWithColumnsResource) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CatalogId; v != nil {
		tfMap["catalog_id"] = aws.StringValue(v)
	}

	tfMap["column_names"] = flex.FlattenStringSet(apiObject.ColumnNames)

	if v := apiObject.DatabaseName; v != nil {
		tfMap["database_name"] = aws.StringValue(v)
	}

	if v := apiObject.ColumnWildcard; v != nil {
		tfMap["wildcard"] = true
		tfMap["excluded_column_names"] = flex.FlattenStringSet(v.ExcludedColumnNames)
	}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	return tfMap
}

// This only happens in very specific situations:
// (Select) TWC + ColumnWildcard              = (Select) Table
// (Select) TWC + ColumnWildcard + ALL_TABLES = (Select) Table + TableWildcard
func flattenTableColumnsResourceAsTable(apiObject *lakeformation.TableWithColumnsResource) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CatalogId; v != nil {
		tfMap["catalog_id"] = aws.StringValue(v)
	}

	if v := apiObject.DatabaseName; v != nil {
		tfMap["database_name"] = aws.StringValue(v)
	}

	if v := apiObject.Name; v != nil && aws.StringValue(v) == TableNameAllTables && apiObject.ColumnWildcard != nil {
		tfMap["wildcard"] = true
	} else if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenPermissions(apiObjects []*lakeformation.PrincipalResourcePermissions) []string {
	if apiObjects == nil {
		return nil
	}

	tfList := make([]string, 0)

	for _, resourcePermission := range apiObjects {
		for _, permission := range resourcePermission.Permissions {
			tfList = append(tfList, aws.StringValue(permission))
		}
	}

	sort.Strings(tfList)

	return tfList
}

func flattenGrantPermissions(apiObjects []*lakeformation.PrincipalResourcePermissions) []string {
	if apiObjects == nil {
		return nil
	}

	tfList := make([]string, 0)

	for _, resourcePermission := range apiObjects {
		for _, grantPermission := range resourcePermission.PermissionsWithGrantOption {
			tfList = append(tfList, aws.StringValue(grantPermission))
		}
	}

	sort.Strings(tfList)

	return tfList
}
