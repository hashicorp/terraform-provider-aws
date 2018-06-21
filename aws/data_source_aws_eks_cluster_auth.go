package aws

import (
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

const (
	clusterIDHeader = "x-k8s-aws-id"
	v1Prefix        = "k8s-aws-v1."
)

func dataSourceAwsEksClusterAuth() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEksClusterAuthRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"duration": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  60,
			},

			"token": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceAwsEksClusterAuthRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).stsconn
	name := d.Get("name").(string)
	duration := d.Get("duration").(int)

	request, _ := conn.GetCallerIdentityRequest(&sts.GetCallerIdentityInput{})
	request.HTTPRequest.Header.Add(clusterIDHeader, name)

	url, err := request.Presign(time.Duration(duration) * time.Second)
	if err != nil {
		return fmt.Errorf("error presigning request: %v", err)
	}

	log.Printf("[DEBUG] Generated request: %s", url)

	token := v1Prefix + base64.RawURLEncoding.EncodeToString([]byte(url))

	d.SetId(time.Now().UTC().String())
	d.Set("token", token)

	return nil
}
