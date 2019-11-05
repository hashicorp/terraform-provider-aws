package aws

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsGreengrassCoreDefinition() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGreengrassCoreDefinitionCreate,
		Read:   resourceAwsGreengrassCoreDefinitionRead,
		Update: resourceAwsGreengrassCoreDefinitionUpdate,
		Delete: resourceAwsGreengrassCoreDefinitionDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"core_definition_version": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"core": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"certificate_arn": {
										Type:     schema.TypeString,
										Required: true,
									},
									"id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"sync_shadow": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"thing_arn": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func convertInterfaceMapToStringMap(interfaceMap map[string]interface{}) map[string]*string {
	stringMap := make(map[string]*string)
	for k, v := range interfaceMap {
		strVal := v.(string)
		stringMap[k] = &strVal
	}
	return stringMap
}

func createCoreDefinitionVersion(d *schema.ResourceData, conn *greengrass.Greengrass) error {
	var rawData map[string]interface{}
	if v := d.Get("core_definition_version").(*schema.Set).List(); len(v) == 0 {
		return nil
	} else {
		rawData = v[0].(map[string]interface{})
	}

	params := &greengrass.CreateCoreDefinitionVersionInput{
		CoreDefinitionId: aws.String(d.Id()),
	}

	if v := os.Getenv("AMZN_CLIENT_TOKEN"); v != "" {
		params.AmznClientToken = aws.String(v)
	}

	cores := make([]*greengrass.Core, 0)
	for _, coreToCast := range rawData["core"].(*schema.Set).List() {
		rawCore := coreToCast.(map[string]interface{})
		core := &greengrass.Core{
			CertificateArn: aws.String(rawCore["certificate_arn"].(string)),
			Id:             aws.String(rawCore["id"].(string)),
			SyncShadow:     aws.Bool(rawCore["sync_shadow"].(bool)),
			ThingArn:       aws.String(rawCore["thing_arn"].(string)),
		}
		cores = append(cores, core)
	}
	params.Cores = cores

	log.Printf("[DEBUG] Creating Greengrass Core Definition Version: %s", params)
	_, err := conn.CreateCoreDefinitionVersion(params)

	if err != nil {
		return err
	}

	return nil
}

func resourceAwsGreengrassCoreDefinitionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.CreateCoreDefinitionInput{
		Name: aws.String(d.Get("name").(string)),
	}

	log.Printf("[DEBUG] Creating Greengrass Core Definition: %s", params)
	out, err := conn.CreateCoreDefinition(params)
	if err != nil {
		return err
	}

	d.SetId(*out.Id)

	err = createCoreDefinitionVersion(d, conn)

	if err != nil {
		return err
	}

	return resourceAwsGreengrassCoreDefinitionRead(d, meta)
}

func setCoreDefinitionVersion(latestVersion string, d *schema.ResourceData, conn *greengrass.Greengrass) error {
	params := &greengrass.GetCoreDefinitionVersionInput{
		CoreDefinitionId:        aws.String(d.Id()),
		CoreDefinitionVersionId: aws.String(latestVersion),
	}

	out, err := conn.GetCoreDefinitionVersion(params)

	if err != nil {
		return err
	}

	rawVersion := make(map[string]interface{})
	rawVersion["arn"] = *out.Arn

	rawCoreList := make([]map[string]interface{}, 0)
	for _, core := range out.Definition.Cores {
		rawCore := make(map[string]interface{})
		rawCore["certificate_arn"] = *core.CertificateArn
		rawCore["sync_shadow"] = *core.SyncShadow
		rawCore["thing_arn"] = *core.ThingArn
		rawCore["id"] = *core.Id
		rawCoreList = append(rawCoreList, rawCore)
	}

	rawVersion["core"] = rawCoreList

	d.Set("core_definition_version", []map[string]interface{}{rawVersion})

	return nil
}

func resourceAwsGreengrassCoreDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.GetCoreDefinitionInput{
		CoreDefinitionId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading Greengrass Core Definition: %s", params)
	out, err := conn.GetCoreDefinition(params)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Received Greengrass Core Definition: %s", out)

	d.Set("arn", out.Arn)
	d.Set("name", out.Name)

	if out.LatestVersion != nil {
		err = setCoreDefinitionVersion(*out.LatestVersion, d, conn)

		if err != nil {
			return err
		}
	}

	return nil
}

func resourceAwsGreengrassCoreDefinitionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.UpdateCoreDefinitionInput{
		Name:             aws.String(d.Get("name").(string)),
		CoreDefinitionId: aws.String(d.Id()),
	}

	_, err := conn.UpdateCoreDefinition(params)
	if err != nil {
		return err
	}

	if d.HasChange("core_definition_version") {
		err = createCoreDefinitionVersion(d, conn)
		if err != nil {
			return err
		}
	}
	return resourceAwsGreengrassCoreDefinitionRead(d, meta)
}

func resourceAwsGreengrassCoreDefinitionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.DeleteCoreDefinitionInput{
		CoreDefinitionId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deleting Greengrass Core Definition: %s", params)

	_, err := conn.DeleteCoreDefinition(params)

	if err != nil {
		return err
	}

	return nil
}
