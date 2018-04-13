package aws

import (
	"strings"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsGlueCatalogTable() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlueCatalogTableCreate,
		Read:   resourceAwsGlueCatalogTableRead,
		Update: resourceAwsGlueCatalogTableUpdate,
		Delete: resourceAwsGlueCatalogTableDelete,
		Exists: resourceAwsGlueCatalogTableExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			...
		},
	}
}

func readAwsGlueTableID(id string) (catalogID string, dbName string, name string) {
	idParts := strings.Split(id, ":")
	return idParts[0], idParts[1], idParts[2]
}
