package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsGlueUserDefinedFunction() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlueUserDefinedFunctionCreate,
		Read:   resourceAwsGlueUserDefinedFunctionRead,
		Update: resourceAwsGlueUserDefinedFunctionUpdate,
		Delete: resourceAwsGlueUserDefinedFunctionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"catalog_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"database_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"class_name": {
				Type:     schema.TypeList,
				Required: true,
			},
			"owner_name": {
				Type:     schema.TypeList,
				Required: true,
			},
			"owner_type": {
				Type:     schema.TypeList,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					glue.PrincipalTypeGroup,
					glue.PrincipalTypeRole,
					glue.PrincipalTypeUser,
				}, false),
			},
			"resource_uris": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								glue.ResourceTypeArchive,
								glue.ResourceTypeFile,
								glue.ResourceTypeJar,
							}, false),
						},
						"uri": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"create_time": {
				Type:     schema.TypeList,
				Computed: true,
			},
		},
	}
}

func resourceAwsGlueUserDefinedFunctionCreate(d *schema.ResourceData, meta interface{}) error {
	glueconn := meta.(*AWSClient).glueconn
	catalogID := createAwsGlueCatalogID(d, meta.(*AWSClient).accountid)
	dbName := d.Get("database_name").(string)
	funcName := d.Get("name").(string)

	input := &glue.CreateUserDefinedFunctionInput{
		CatalogId:     aws.String(catalogID),
		DatabaseName:  aws.String(dbName),
		FunctionInput: expandAwsGlueUserDefinedFunctionInput(d),
	}

	_, err := glueconn.CreateUserDefinedFunction(input)
	if err != nil {
		return fmt.Errorf("error creating Glue User Defined Function: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", catalogID, dbName, funcName))

	return resourceAwsGlueUserDefinedFunctionUpdate(d, meta)
}

func resourceAwsGlueUserDefinedFunctionUpdate(d *schema.ResourceData, meta interface{}) error {
	glueconn := meta.(*AWSClient).glueconn

	catalogID, dbName, funcName, err := readAwsGlueUDFID(d.Id())
	if err != nil {
		return err
	}

	input := &glue.UpdateUserDefinedFunctionInput{
		CatalogId:     aws.String(catalogID),
		DatabaseName:  aws.String(dbName),
		FunctionName:  aws.String(funcName),
		FunctionInput: expandAwsGlueUserDefinedFunctionInput(d),
	}

	if d.HasChange("owner_name") || d.HasChange("owner_type") ||
		d.HasChange("class_name") || d.HasChange("resource_uris") {
		if _, err := glueconn.UpdateUserDefinedFunction(input); err != nil {
			return err
		}
	}

	return resourceAwsGlueUserDefinedFunctionRead(d, meta)
}

func resourceAwsGlueUserDefinedFunctionRead(d *schema.ResourceData, meta interface{}) error {
	glueconn := meta.(*AWSClient).glueconn

	catalogID, dbName, funcName, err := readAwsGlueUDFID(d.Id())
	if err != nil {
		return err
	}

	input := &glue.GetUserDefinedFunctionInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		FunctionName: aws.String(funcName),
	}

	out, err := glueconn.GetUserDefinedFunction(input)
	if err != nil {

		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			log.Printf("[WARN] Glue User Defined Function (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error reading Glue User Defined Function: %s", err.Error())
	}

	udf := out.UserDefinedFunction
	d.Set("name", udf.FunctionName)
	d.Set("catalog_id", catalogID)
	d.Set("owner_type", udf.OwnerType)
	d.Set("owner_name", udf.OwnerName)
	d.Set("class_name", udf.ClassName)
	if udf.CreateTime != nil {
		d.Set("create_time", udf.CreateTime.Format(time.RFC3339))
	}
	if err := d.Set("resource_uris", flattenAwsGlueUserDefinedFunctionResourceUri(udf.ResourceUris)); err != nil {
		return err
	}

	return nil
}

func resourceAwsGlueUserDefinedFunctionDelete(d *schema.ResourceData, meta interface{}) error {
	glueconn := meta.(*AWSClient).glueconn
	catalogID, dbName, funcName, err := readAwsGlueUDFID(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Glue User Defined Function: %s", d.Id())
	_, err = glueconn.DeleteUserDefinedFunction(&glue.DeleteUserDefinedFunctionInput{
		CatalogId:    aws.String(catalogID),
		DatabaseName: aws.String(dbName),
		FunctionName: aws.String(funcName),
	})
	if err != nil {
		return fmt.Errorf("error deleting Glue User Defined Function: %s", err.Error())
	}
	return nil
}

func readAwsGlueUDFID(id string) (catalogID string, dbName string, funcName string, err error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 3 {
		return "", "", "", fmt.Errorf("unexpected format of ID (%q), expected CATALOG-ID:DATABASE-NAME:FUNCTION-NAME", id)
	}
	return idParts[0], idParts[1], idParts[2], nil
}

func expandAwsGlueUserDefinedFunctionInput(d *schema.ResourceData) *glue.UserDefinedFunctionInput {

	udf := &glue.UserDefinedFunctionInput{
		ClassName:    aws.String(d.Get("class_name").(string)),
		FunctionName: aws.String(d.Get("name").(string)),
		OwnerName:    aws.String(d.Get("owner_name").(string)),
		OwnerType:    aws.String(d.Get("owner_type").(string)),
		ResourceUris: expandAwsGlueUserDefinedFunctionResourceUri(d.Get("resource_uris").(*schema.Set)),
	}

	return udf
}

func expandAwsGlueUserDefinedFunctionResourceUri(conf *schema.Set) []*glue.ResourceUri {
	result := make([]*glue.ResourceUri, 0, conf.Len())

	for _, r := range conf.List() {
		uriRaw := r.(map[string]interface{})
		uri := &glue.ResourceUri{
			ResourceType: aws.String(uriRaw["resource_type"].(string)),
			Uri:          aws.String(uriRaw["uri"].(string)),
		}

		result = append(result, uri)
	}

	return result
}

func flattenAwsGlueUserDefinedFunctionResourceUri(uris []*glue.ResourceUri) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(uris))

	for _, i := range uris {
		l := map[string]interface{}{
			"resource_type": aws.StringValue(i.ResourceType),
			"uri":           aws.StringValue(i.Uri),
		}

		result = append(result, l)
	}
	return result
}
