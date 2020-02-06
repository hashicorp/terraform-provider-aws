package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsAmplifyDomainAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAmplifyDomainAssociationCreate,
		Read:   resourceAwsAmplifyDomainAssociationRead,
		Update: resourceAwsAmplifyDomainAssociationUpdate,
		Delete: resourceAwsAmplifyDomainAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"app_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"enable_auto_sub_domain": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"sub_domain_settings": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"branch_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"prefix": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			// non-API
			"wait_for_verification": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return true
				},
			},
		},
	}
}

func resourceAwsAmplifyDomainAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn
	log.Print("[DEBUG] Creating Amplify DomainAssociation")

	params := &amplify.CreateDomainAssociationInput{
		AppId:      aws.String(d.Get("app_id").(string)),
		DomainName: aws.String(d.Get("domain_name").(string)),
	}

	if v, ok := d.GetOk("sub_domain_settings"); ok {
		params.SubDomainSettings = expandAmplifySubDomainSettings(v.([]interface{}))
	}

	if v, ok := d.GetOk("enable_auto_sub_domain"); ok {
		params.EnableAutoSubDomain = aws.Bool(v.(bool))
	}

	resp, err := conn.CreateDomainAssociation(params)
	if err != nil {
		return fmt.Errorf("Error creating Amplify DomainAssociation: %s", err)
	}

	arn := *resp.DomainAssociation.DomainAssociationArn
	d.SetId(arn[strings.Index(arn, ":apps/")+len(":apps/"):])

	if d.Get("wait_for_verification").(bool) {
		log.Printf("[DEBUG] Waiting until Amplify DomainAssociation (%s) is deployed", d.Id())
		if err := resourceAwsAmplifyDomainAssociationWaitUntilVerified(d.Id(), meta); err != nil {
			return fmt.Errorf("error waiting until Amplify DomainAssociation (%s) is deployed: %s", d.Id(), err)
		}
	}

	return resourceAwsAmplifyDomainAssociationRead(d, meta)
}

func resourceAwsAmplifyDomainAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn
	log.Printf("[DEBUG] Reading Amplify DomainAssociation: %s", d.Id())

	s := strings.Split(d.Id(), "/")
	app_id := s[0]
	domain_name := s[2]

	resp, err := conn.GetDomainAssociation(&amplify.GetDomainAssociationInput{
		AppId:      aws.String(app_id),
		DomainName: aws.String(domain_name),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == amplify.ErrCodeNotFoundException {
			log.Printf("[WARN] Amplify DomainAssociation (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("app_id", app_id)
	d.Set("arn", resp.DomainAssociation.DomainAssociationArn)
	d.Set("domain_name", resp.DomainAssociation.DomainName)
	if err := d.Set("sub_domain_settings", flattenAmplifySubDomainSettings(resp.DomainAssociation.SubDomains)); err != nil {
		return fmt.Errorf("error setting sub_domain_settings: %s", err)
	}
	d.Set("enable_auto_sub_domain", resp.DomainAssociation.EnableAutoSubDomain)

	return nil
}

func resourceAwsAmplifyDomainAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn
	log.Printf("[DEBUG] Updating Amplify DomainAssociation: %s", d.Id())

	s := strings.Split(d.Id(), "/")
	app_id := s[0]
	domain_name := s[2]

	params := &amplify.UpdateDomainAssociationInput{
		AppId:      aws.String(app_id),
		DomainName: aws.String(domain_name),
	}

	if d.HasChange("sub_domain_settings") {
		params.SubDomainSettings = expandAmplifySubDomainSettings(d.Get("sub_domain_settings").([]interface{}))
	}

	if d.HasChange("enable_auto_sub_domain") {
		params.EnableAutoSubDomain = aws.Bool(d.Get("enable_auto_sub_domain").(bool))
	}

	_, err := conn.UpdateDomainAssociation(params)
	if err != nil {
		return fmt.Errorf("Error updating Amplify DomainAssociation: %s", err)
	}

	if d.Get("wait_for_verification").(bool) {
		log.Printf("[DEBUG] Waiting until Amplify DomainAssociation (%s) is deployed", d.Id())
		if err := resourceAwsAmplifyDomainAssociationWaitUntilVerified(d.Id(), meta); err != nil {
			return fmt.Errorf("error waiting until Amplify DomainAssociation (%s) is deployed: %s", d.Id(), err)
		}
	}

	return resourceAwsAmplifyDomainAssociationRead(d, meta)
}

func resourceAwsAmplifyDomainAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).amplifyconn
	log.Printf("[DEBUG] Deleting Amplify DomainAssociation: %s", d.Id())

	s := strings.Split(d.Id(), "/")
	app_id := s[0]
	domain_name := s[2]

	params := &amplify.DeleteDomainAssociationInput{
		AppId:      aws.String(app_id),
		DomainName: aws.String(domain_name),
	}

	_, err := conn.DeleteDomainAssociation(params)
	if err != nil {
		return fmt.Errorf("Error deleting Amplify DomainAssociation: %s", err)
	}

	return nil
}

func resourceAwsAmplifyDomainAssociationWaitUntilVerified(id string, meta interface{}) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			amplify.DomainStatusPendingVerification,
			amplify.DomainStatusInProgress,
			amplify.DomainStatusCreating,
			amplify.DomainStatusRequestingCertificate,
			amplify.DomainStatusUpdating,
		},
		Target: []string{
			// It takes up to 30 minutes, so skip waiting for deployment.
			amplify.DomainStatusPendingDeployment,
			amplify.DomainStatusAvailable,
		},
		Refresh:    resourceAwsAmplifyDomainAssociationStateRefreshFunc(id, meta),
		Timeout:    15 * time.Minute,
		MinTimeout: 15 * time.Second,
		Delay:      10 * time.Second,
	}

	_, err := stateConf.WaitForState()
	return err
}

func resourceAwsAmplifyDomainAssociationStateRefreshFunc(id string, meta interface{}) resource.StateRefreshFunc {
	s := strings.Split(id, "/")
	app_id := s[0]
	domain_name := s[2]

	return func() (interface{}, string, error) {
		conn := meta.(*AWSClient).amplifyconn

		resp, err := conn.GetDomainAssociation(&amplify.GetDomainAssociationInput{
			AppId:      aws.String(app_id),
			DomainName: aws.String(domain_name),
		})
		if err != nil {
			log.Printf("[WARN] Error retrieving Amplify DomainAssociation %q details: %s", id, err)
			return nil, "", err
		}

		if *resp.DomainAssociation.DomainStatus == amplify.DomainStatusFailed {
			return nil, "", fmt.Errorf("%s", *resp.DomainAssociation.StatusReason)
		}

		return resp.DomainAssociation, *resp.DomainAssociation.DomainStatus, nil
	}
}

func expandAmplifySubDomainSettings(values []interface{}) []*amplify.SubDomainSetting {
	settings := make([]*amplify.SubDomainSetting, 0)

	for _, v := range values {
		e := v.(map[string]interface{})

		setting := &amplify.SubDomainSetting{}

		if ev, ok := e["branch_name"].(string); ok && ev != "" {
			setting.BranchName = aws.String(ev)
		}

		if ev, ok := e["prefix"].(string); ok {
			setting.Prefix = aws.String(ev)
		}

		settings = append(settings, setting)
	}

	return settings
}

func flattenAmplifySubDomainSettings(sub_domains []*amplify.SubDomain) []map[string]interface{} {
	values := make([]map[string]interface{}, 0)

	for _, v := range sub_domains {
		kv := make(map[string]interface{})

		if v.SubDomainSetting.BranchName != nil {
			kv["branch_name"] = *v.SubDomainSetting.BranchName
		}

		if v.SubDomainSetting.Prefix != nil {
			kv["prefix"] = *v.SubDomainSetting.Prefix
		}

		values = append(values, kv)
	}

	return values
}
