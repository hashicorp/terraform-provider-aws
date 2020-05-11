package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsLakeFormationDataLakeSettings() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLakeFormationDataLakeSettingsPut,
		Update: resourceAwsLakeFormationDataLakeSettingsPut,
		Read:   resourceAwsLakeFormationDataLakeSettingsRead,
		Delete: resourceAwsLakeFormationDataLakeSettingsReset,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"catalog_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"admins": {
				Type:     schema.TypeSet,
				Set:      schema.HashString,
				Required: true,
				MinItems: 0,
				MaxItems: 10,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.NoZeroValues,
				},
			},
		},
	}
}

func resourceAwsLakeFormationDataLakeSettingsPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn
	catalogId := createAwsDataCatalogId(d, meta.(*AWSClient).accountid)

	input := &lakeformation.PutDataLakeSettingsInput{
		CatalogId: aws.String(catalogId),
		DataLakeSettings: &lakeformation.DataLakeSettings{
			DataLakeAdmins: expandLakeFormationDataLakePrincipal(d),
		},
	}

	_, err := conn.PutDataLakeSettings(input)
	if err != nil {
		return fmt.Errorf("Error updating DataLakeSettings: %s", err)
	}

	awsRegion := meta.(*AWSClient).region
	d.SetId(fmt.Sprintf("lakeformation:%s:%s", awsRegion, catalogId))
	d.Set("catalog_id", catalogId)

	return resourceAwsLakeFormationDataLakeSettingsRead(d, meta)
}

func resourceAwsLakeFormationDataLakeSettingsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn
	catalogId := d.Get("catalog_id").(string)

	input := &lakeformation.GetDataLakeSettingsInput{
		CatalogId: aws.String(catalogId),
	}

	out, err := conn.GetDataLakeSettings(input)
	if err != nil {
		return fmt.Errorf("Error reading DataLakeSettings: %s", err)
	}

	d.Set("catalog_id", catalogId)
	if err := d.Set("admins", flattenLakeFormationDataLakePrincipal(out.DataLakeSettings.DataLakeAdmins)); err != nil {
		return fmt.Errorf("Error setting admins from DataLakeSettings: %s", err)
	}
	// TODO: Add CreateDatabaseDefaultPermissions and CreateTableDefaultPermissions

	return nil
}

func resourceAwsLakeFormationDataLakeSettingsReset(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lakeformationconn
	catalogId := d.Get("catalog_id").(string)

	input := &lakeformation.PutDataLakeSettingsInput{
		CatalogId: aws.String(catalogId),
		DataLakeSettings: &lakeformation.DataLakeSettings{
			DataLakeAdmins: make([]*lakeformation.DataLakePrincipal, 0),
		},
	}

	_, err := conn.PutDataLakeSettings(input)
	if err != nil {
		return fmt.Errorf("Error reseting DataLakeSettings: %s", err)
	}

	return nil
}

func createAwsDataCatalogId(d *schema.ResourceData, accountId string) (catalogId string) {
	if inputCatalogId, ok := d.GetOkExists("catalog_id"); ok {
		catalogId = inputCatalogId.(string)
	} else {
		catalogId = accountId
	}
	return
}

func expandLakeFormationDataLakePrincipal(d *schema.ResourceData) []*lakeformation.DataLakePrincipal {
	xs := d.Get("admins").(*schema.Set).List()
	ys := make([]*lakeformation.DataLakePrincipal, len(xs))

	for i, x := range xs {
		ys[i] = &lakeformation.DataLakePrincipal{
			DataLakePrincipalIdentifier: aws.String(x.(string)),
		}
	}

	return ys
}

func flattenLakeFormationDataLakePrincipal(xs []*lakeformation.DataLakePrincipal) []string {
	ys := make([]string, len(xs))
	for i, x := range xs {
		ys[i] = aws.StringValue(x.DataLakePrincipalIdentifier)
	}

	return ys
}
