package aws

import (
	"regexp"

	"github.com/hashicorp/terraform/helper/validation"

	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsAthenaWorkgroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAthenaWorkgroupCreate,
		Read:   resourceAwsAthenaWorkgroupRead,
		Update: resourceAwsAthenaWorkgroupUpdate,
		Delete: resourceAwsAthenaWorkgroupDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bytes_scanned_cutoff_per_query": {
							Type: schema.TypeInt,
							Optional: true,
						},
						"enforce_workgroup_configuration": {
							Type: schema.TypeBool,
							Optional: true,
						},
						"publish_cloudwatch_metrics_enable": {
							Type: schema.TypeBool,
							Optional: true,
						}
						"result_configuration":{
							Type: schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								"output_location": {
									Type: schema.TypeString,
									Optional: true,
								},
								"encryption_configuration": {
									Type: schema.TypeList,
									Optional: true,
									MaxItems: 1,
									Elem: &schema.Resource{
										"encryption_option": {
											Type: schema.TypeString,
											Optional: true,
											ValidateFunc: validation.StringInSlice([]string{"SSE_S3", "SSE_KMS", "CSE_KMS"}, false)
										},
										"kms_key": {
											Type: schema.TypeString,
											Optional: true,
										},
									},
								},
							},
						},
					},
				},
			},
			"tags" : tagsSchema(),
		},
	}
}
