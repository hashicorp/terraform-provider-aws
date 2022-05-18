package lakeformation

import (
	"fmt"
	"log"

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

func ResourceDataLakeSettings() *schema.Resource {
	return &schema.Resource{
		Create: resourceDataLakeSettingsCreate,
		Update: resourceDataLakeSettingsCreate,
		Read:   resourceDataLakeSettingsRead,
		Delete: resourceDataLakeSettingsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"admins": {
				Type:     schema.TypeSet,
				Computed: true,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"catalog_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"create_database_default_permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				MaxItems: 3,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"permissions": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(lakeformation.Permission_Values(), false),
							},
						},
						"principal": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validPrincipal,
						},
					},
				},
			},
			"create_table_default_permissions": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				MaxItems: 3,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"permissions": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(lakeformation.Permission_Values(), false),
							},
						},
						"principal": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validPrincipal,
						},
					},
				},
			},
			"trusted_resource_owners": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidAccountID,
				},
			},
		},
	}
}

func resourceDataLakeSettingsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LakeFormationConn

	input := &lakeformation.PutDataLakeSettingsInput{}

	if v, ok := d.GetOk("catalog_id"); ok {
		input.CatalogId = aws.String(v.(string))
	}

	settings := &lakeformation.DataLakeSettings{}

	if v, ok := d.GetOk("create_database_default_permissions"); ok {
		settings.CreateDatabaseDefaultPermissions = expandDataLakeSettingsCreateDefaultPermissions(v.([]interface{}))
	}

	if v, ok := d.GetOk("create_table_default_permissions"); ok {
		settings.CreateTableDefaultPermissions = expandDataLakeSettingsCreateDefaultPermissions(v.([]interface{}))
	}

	if v, ok := d.GetOk("admins"); ok {
		settings.DataLakeAdmins = expandDataLakeSettingsAdmins(v.(*schema.Set))
	}

	if v, ok := d.GetOk("trusted_resource_owners"); ok {
		settings.TrustedResourceOwners = flex.ExpandStringList(v.([]interface{}))
	}

	input.DataLakeSettings = settings

	var output *lakeformation.PutDataLakeSettingsOutput
	err := resource.Retry(IAMPropagationTimeout, func() *resource.RetryError {
		var err error
		output, err = conn.PutDataLakeSettings(input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, lakeformation.ErrCodeInvalidInputException, "Invalid principal") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeConcurrentModificationException) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(fmt.Errorf("error creating Lake Formation data lake settings: %w", err))
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.PutDataLakeSettings(input)
	}

	if err != nil {
		return fmt.Errorf("error creating Lake Formation data lake settings: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error creating Lake Formation data lake settings: empty response")
	}

	d.SetId(fmt.Sprintf("%d", create.StringHashcode(input.String())))

	return resourceDataLakeSettingsRead(d, meta)
}

func resourceDataLakeSettingsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LakeFormationConn

	input := &lakeformation.GetDataLakeSettingsInput{}

	if v, ok := d.GetOk("catalog_id"); ok {
		input.CatalogId = aws.String(v.(string))
	}

	output, err := conn.GetDataLakeSettings(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeEntityNotFoundException) {
		log.Printf("[WARN] Lake Formation data lake settings (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Lake Formation data lake settings (%s): %w", d.Id(), err)
	}

	if output == nil || output.DataLakeSettings == nil {
		return fmt.Errorf("reading Lake Formation data lake settings (%s): empty response", d.Id())
	}

	settings := output.DataLakeSettings

	d.Set("create_database_default_permissions", flattenDataLakeSettingsCreateDefaultPermissions(settings.CreateDatabaseDefaultPermissions))
	d.Set("create_table_default_permissions", flattenDataLakeSettingsCreateDefaultPermissions(settings.CreateTableDefaultPermissions))
	d.Set("admins", flattenDataLakeSettingsAdmins(settings.DataLakeAdmins))
	d.Set("trusted_resource_owners", flex.FlattenStringList(settings.TrustedResourceOwners))

	return nil
}

func resourceDataLakeSettingsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LakeFormationConn

	input := &lakeformation.PutDataLakeSettingsInput{
		DataLakeSettings: &lakeformation.DataLakeSettings{
			CreateDatabaseDefaultPermissions: make([]*lakeformation.PrincipalPermissions, 0),
			CreateTableDefaultPermissions:    make([]*lakeformation.PrincipalPermissions, 0),
			DataLakeAdmins:                   make([]*lakeformation.DataLakePrincipal, 0),
			TrustedResourceOwners:            make([]*string, 0),
		},
	}

	if v, ok := d.GetOk("catalog_id"); ok {
		input.CatalogId = aws.String(v.(string))
	}

	_, err := conn.PutDataLakeSettings(input)

	if tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeEntityNotFoundException) {
		log.Printf("[WARN] Lake Formation data lake settings (%s) not found, removing from state", d.Id())
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Lake Formation data lake settings (%s): %w", d.Id(), err)
	}

	return nil
}

func expandDataLakeSettingsCreateDefaultPermissions(tfMaps []interface{}) []*lakeformation.PrincipalPermissions {
	apiObjects := make([]*lakeformation.PrincipalPermissions, 0, len(tfMaps))

	for _, tfMap := range tfMaps {
		apiObjects = append(apiObjects, expandDataLakeSettingsCreateDefaultPermission(tfMap.(map[string]interface{})))
	}

	return apiObjects
}

func expandDataLakeSettingsCreateDefaultPermission(tfMap map[string]interface{}) *lakeformation.PrincipalPermissions {
	apiObject := &lakeformation.PrincipalPermissions{
		Permissions: flex.ExpandStringSet(tfMap["permissions"].(*schema.Set)),
		Principal: &lakeformation.DataLakePrincipal{
			DataLakePrincipalIdentifier: aws.String(tfMap["principal"].(string)),
		},
	}

	return apiObject
}

func flattenDataLakeSettingsCreateDefaultPermissions(apiObjects []*lakeformation.PrincipalPermissions) []map[string]interface{} {
	if apiObjects == nil {
		return nil
	}

	tfMaps := make([]map[string]interface{}, len(apiObjects))
	for i, v := range apiObjects {
		tfMaps[i] = flattenDataLakeSettingsCreateDefaultPermission(v)
	}

	return tfMaps
}

func flattenDataLakeSettingsCreateDefaultPermission(apiObject *lakeformation.PrincipalPermissions) map[string]interface{} {
	tfMap := make(map[string]interface{})

	if apiObject == nil {
		return tfMap
	}

	if apiObject.Permissions != nil {
		tfMap["permissions"] = flex.FlattenStringSet(apiObject.Permissions)
	}

	if v := aws.StringValue(apiObject.Principal.DataLakePrincipalIdentifier); v != "" {
		tfMap["principal"] = v
	}

	return tfMap
}

func expandDataLakeSettingsAdmins(tfSet *schema.Set) []*lakeformation.DataLakePrincipal {
	tfSlice := tfSet.List()
	apiObjects := make([]*lakeformation.DataLakePrincipal, 0, len(tfSlice))

	for _, tfItem := range tfSlice {
		val, ok := tfItem.(string)
		if ok && val != "" {
			apiObjects = append(apiObjects, &lakeformation.DataLakePrincipal{
				DataLakePrincipalIdentifier: aws.String(tfItem.(string)),
			})
		}
	}

	return apiObjects
}

func flattenDataLakeSettingsAdmins(apiObjects []*lakeformation.DataLakePrincipal) []interface{} {
	if apiObjects == nil {
		return nil
	}

	tfSlice := make([]interface{}, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfSlice = append(tfSlice, *apiObject.DataLakePrincipalIdentifier)
	}

	return tfSlice
}
