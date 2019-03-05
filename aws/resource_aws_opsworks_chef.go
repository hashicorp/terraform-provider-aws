package aws

import (
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/opsworkscm"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsOpsworksChef() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsOpsworksChefCreate,
		Read:   resourceAwsOpsworksChefRead,
		Update: resourceAwsOpsworksChefUpdate,
		Delete: resourceAwsOpsworksChefDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"associate_public_ip_address": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"backup_retention_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},

			"disable_automated_backup": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			// I kinda want to make this required because how else are we going to get at that info?
			"chef_pivotal_key": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"chef_delivery_admin_password": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			// TODO:
			// default and only valid. I'll leave it configurable for future proofing but
			// validation failure, maybe?
			"engine_model": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Single",
			},

			// TODO:
			// default and only valid. I'll leave it configurable for future proofing but
			// validation failure, maybe?
			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "12",
			},

			// TODO: document the instance role required
			// https://s3.amazonaws.com/opsworks-cm-us-east-1-prod-default-assets/misc/opsworks-cm-roles.yaml
			"instance_profile_arn": {
				Type:     schema.TypeString,
				Required: true,
			},

			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
			},

			"ssh_key_pair": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"preferred_backup_window": {
				Type:     schema.TypeString,
				Required: true, // TODO: ?
			},

			"preferred_maintenance_window": {
				Type:     schema.TypeString,
				Required: true,
			},

			// TODO:
			// if you don't specify this, it'll create one. I'm of the opinion that users should
			// create it themselves. Or maybe some sort of wrapper module? I dunno.
			"security_group_ids": {
				Type:     schema.TypeList,
				Required: true,
			},

			// TODO: document this role creation
			// https://s3.amazonaws.com/opsworks-cm-us-east-1-prod-default-assets/misc/opsworks-cm-roles.yaml
			"service_role_arn": {
				Type:     schema.TypeString,
				Required: true,
			},

			// TODO: huh?
			// The IDs of subnets in which to launch the server EC2 instance.
			//
			// Amazon EC2-Classic customers: This field is required. All servers must run
			// within a VPC. The VPC must have "Auto Assign Public IP" enabled.
			//
			// EC2-VPC customers: This field is optional. If you do not specify subnet IDs,
			// your EC2 instances are created in a default subnet that is selected by Amazon
			// EC2. If you specify subnet IDs, the VPC must have "Auto Assign Public IP"
			// enabled.

			"subnet_ids": {
				Type:     schema.TypeList,
				Required: true,
			},
		},
	}
}

func resourceAwsOpsworksChefValidate(d *schema.ResourceData) error {
	// TODO: parameter validation function
	return nil
}

func resourceAwsOpsworksChefRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworkscmconn

	req := &opsworkscm.DescribeServersInput{
		ServerName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading OpsWorks Chef server: %s", d.Id())

	var resp *opsworkscm.DescribeServersOutput
	var dErr error

	resp, dErr = client.DescribeServers(req)
	if dErr != nil {
		log.Printf("[DEBUG] OpsWorks Chef (%s) not found", d.Id())
		d.SetId("")
		return dErr
	}

	server := resp.Servers[0]
	// TODO: figure / set attributes
	d.Set("arn", server.ServerArn)

	return nil
}

func resourceAwsOpsworksChefCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworkscmconn

	err := resourceAwsOpsworksChefValidate(d)
	if err != nil {
		return err
	}

	req := &opsworkscm.CreateServerInput{
		// TODO: parameters
		// "name":         d.Id(),
		// "engine":       "Chef",
		// "engine_model": "single", // required?
	}

	log.Printf("[DEBUG] Creating OpsWorks Chef server: %s", req)

	var resp *opsworkscm.CreateServerOutput
	err = resource.Retry(20*time.Minute, func() *resource.RetryError {
		var cerr error
		resp, cerr = client.CreateServer(req)
		if cerr != nil {
			if opserr, ok := cerr.(awserr.Error); ok {
				// If Terraform is also managing the service IAM role,
				// it may have just been created and not yet be
				// propagated.
				// AWS doesn't provide a machine-readable code for this
				// specific error, so we're forced to do fragile message
				// matching.
				// The full error we're looking for looks something like
				// the following:
				// Service Role Arn: [...] is not yet propagated, please try again in a couple of minutes
				propErr := "not yet propagated"
				trustErr := "not the necessary trust relationship"
				validateErr := "validate IAM role permission"
				if opserr.Code() == "ValidationException" && (strings.Contains(opserr.Message(), trustErr) || strings.Contains(opserr.Message(), propErr) || strings.Contains(opserr.Message(), validateErr)) {
					log.Printf("[INFO] Waiting for service IAM role to propagate")
					return resource.RetryableError(cerr)
				}
			}
			return resource.NonRetryableError(cerr)
		}
		return nil
	})
	if err != nil {
		return err
	}

	serverName := *resp.Server.ServerName
	d.SetId(serverName)

	return resourceAwsOpsworksChefUpdate(d, meta)
}

func resourceAwsOpsworksChefUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworkscmconn

	err := resourceAwsOpsworksChefValidate(d)
	if err != nil {
		return err
	}

	req := &opsworkscm.UpdateServerInput{
		// TODO: params
	}

	// TODO: params? d.GetOk() stuff wat

	log.Printf("[DEBUG] Updating OpsWorks Chef server: %s", req)

	_, err = client.UpdateServer(req)
	if err != nil {
		return err
	}

	return resourceAwsOpsworksChefRead(d, meta)
}

func resourceAwsOpsworksChefDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworkscmconn

	req := &opsworkscm.DeleteServerInput{
		ServerName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting OpsWorks Chef server: %s", d.Id())

	_, err := client.DeleteServer(req)
	if err != nil {
		return err
	}

	return nil
}
