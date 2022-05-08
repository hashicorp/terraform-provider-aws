package opensearch

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceDomainSAMLOptions() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainSAMLOptionsPut,
		Read:   resourceDomainSAMLOptionsRead,
		Update: resourceDomainSAMLOptionsPut,
		Delete: resourceDomainSAMLOptionsDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("domain_name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(180 * time.Minute),
			Delete: schema.DefaultTimeout(90 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"saml_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"idp": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"entity_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"metadata_content": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringIsNotEmpty,
									},
								},
							},
						},
						"master_backend_role": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"master_user_name": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"roles_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"session_timeout_minutes": {
							Type:             schema.TypeInt,
							Optional:         true,
							Default:          60,
							ValidateFunc:     validation.IntBetween(1, 1440),
							DiffSuppressFunc: domainSamlOptionsDiffSupress,
						},
						"subject_key": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "",
							DiffSuppressFunc: domainSamlOptionsDiffSupress,
						},
					},
				},
			},
		},
	}
}
func domainSamlOptionsDiffSupress(k, old, new string, d *schema.ResourceData) bool {
	if v, ok := d.Get("saml_options").([]interface{}); ok && len(v) > 0 {
		if enabled, ok := v[0].(map[string]interface{})["enabled"].(bool); ok && !enabled {
			return true
		}
	}
	return false
}

func resourceDomainSAMLOptionsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OpenSearchConn

	ds, err := FindDomainByName(conn, d.Get("domain_name").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] OpenSearch Domain SAML Options (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading OpenSearch Domain SAML Options (%s): %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Received OpenSearch domain: %s", ds)

	options := ds.AdvancedSecurityOptions.SAMLOptions

	if err := d.Set("saml_options", flattenESSAMLOptions(d, options)); err != nil {
		return fmt.Errorf("error setting saml_options for OpenSearch Configuration: %w", err)
	}

	return nil
}

func resourceDomainSAMLOptionsPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OpenSearchConn

	domainName := d.Get("domain_name").(string)
	config := opensearchservice.AdvancedSecurityOptionsInput_{}
	config.SetSAMLOptions(expandESSAMLOptions(d.Get("saml_options").([]interface{})))

	log.Printf("[DEBUG] Updating OpenSearch domain SAML Options %s", config)

	_, err := conn.UpdateDomainConfig(&opensearchservice.UpdateDomainConfigInput{
		DomainName:              aws.String(domainName),
		AdvancedSecurityOptions: &config,
	})

	if err != nil {
		return err
	}

	d.SetId(domainName)

	if err := waitForDomainUpdate(conn, d.Get("domain_name").(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return fmt.Errorf("error waiting for OpenSearch Domain SAML Options update (%s) to succeed: %w", d.Id(), err)
	}

	return resourceDomainSAMLOptionsRead(d, meta)
}

func resourceDomainSAMLOptionsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OpenSearchConn

	domainName := d.Get("domain_name").(string)
	config := opensearchservice.AdvancedSecurityOptionsInput_{}
	config.SetSAMLOptions(nil)

	_, err := conn.UpdateDomainConfig(&opensearchservice.UpdateDomainConfigInput{
		DomainName:              aws.String(domainName),
		AdvancedSecurityOptions: &config,
	})
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Waiting for OpenSearch domain SAML Options %q to be deleted", d.Get("domain_name").(string))

	if err := waitForDomainUpdate(conn, d.Get("domain_name").(string), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for OpenSearch Domain SAML Options (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
