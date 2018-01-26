package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsEcrLoginToken() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEcrLoginTokenRead,

		Schema: map[string]*schema.Schema{
			"registry_ids": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"token": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"proxy_endpoint": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"expires_at": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsEcrLoginTokenRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecrconn

	input := &ecr.GetAuthorizationTokenInput{}

	if ids, ok := d.GetOk("registryIds"); ok {
		input.RegistryIds = expandStringList(ids.([]interface{}))
	}

	out, err := conn.GetAuthorizationToken(input)
	if err != nil {
		if ecrerr, ok := err.(awserr.Error); ok {
			log.Printf("[ERROR] ECR Login Token error: %s, Code: %s", ecrerr.Message(), ecrerr.Code())
		}

		d.SetId("")
		return fmt.Errorf("[ERROR] ECR Login Token error: %s", err)
	}

	log.Printf("[DEBUG] Received ECR AuthorizationData %v", out.AuthorizationData)

	var tokens, endpoints, expires []string

	for _, authObject := range out.AuthorizationData {
		tokens = append(tokens, *authObject.AuthorizationToken)
		endpoints = append(endpoints, *authObject.ProxyEndpoint)
		expires = append(expires, authObject.ExpiresAt.String())
	}

	d.SetId(time.Now().UTC().String())
	d.Set("token", tokens)
	d.Set("proxy_endpoint", endpoints)
	d.Set("expires_at", expires)

	return nil
}
