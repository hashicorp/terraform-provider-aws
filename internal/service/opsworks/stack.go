package opsworks

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	securityGroupsCreatedSleepTime = 30 * time.Second
	securityGroupsDeletedSleepTime = 30 * time.Second
)

func ResourceStack() *schema.Resource {
	return &schema.Resource{
		Create: resourceStackCreate,
		Read:   resourceStackRead,
		Update: resourceStackUpdate,
		Delete: resourceStackDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"agent_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"region": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},

			"service_role_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"default_instance_profile_arn": {
				Type:     schema.TypeString,
				Required: true,
			},

			"color": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"configuration_manager_name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Chef",
			},

			"configuration_manager_version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "11.10",
			},

			"manage_berkshelf": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"berkshelf_version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "3.2.0",
			},

			"custom_cookbooks_source": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},

						"url": {
							Type:     schema.TypeString,
							Required: true,
						},

						"username": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"password": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},

						"revision": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"ssh_key": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
					},
				},
			},

			"custom_json": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
			},

			"default_availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"default_os": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Ubuntu 12.04 LTS",
			},

			"default_root_device_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "instance-store",
			},

			"default_ssh_key_name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"default_subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"hostname_theme": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Layer_Dependent",
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),

			"use_custom_cookbooks": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"use_opsworks_security_groups": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Computed: true,
				Optional: true,
			},

			"stack_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceStackValidate(d *schema.ResourceData) error {
	cookbooksSourceCount := d.Get("custom_cookbooks_source.#").(int)
	if cookbooksSourceCount > 1 {
		return fmt.Errorf("Only one custom_cookbooks_source is permitted")
	}

	vpcId := d.Get("vpc_id").(string)
	if vpcId != "" {
		if d.Get("default_subnet_id").(string) == "" {
			return fmt.Errorf("default_subnet_id must be set if vpc_id is set")
		}
	} else {
		if d.Get("default_availability_zone").(string) == "" {
			return fmt.Errorf("either vpc_id or default_availability_zone must be set")
		}
	}

	return nil
}

func resourceStackCustomCookbooksSource(d *schema.ResourceData) *opsworks.Source {
	count := d.Get("custom_cookbooks_source.#").(int)
	if count == 0 {
		return nil
	}

	return &opsworks.Source{
		Type:     aws.String(d.Get("custom_cookbooks_source.0.type").(string)),
		Url:      aws.String(d.Get("custom_cookbooks_source.0.url").(string)),
		Username: aws.String(d.Get("custom_cookbooks_source.0.username").(string)),
		Password: aws.String(d.Get("custom_cookbooks_source.0.password").(string)),
		Revision: aws.String(d.Get("custom_cookbooks_source.0.revision").(string)),
		SshKey:   aws.String(d.Get("custom_cookbooks_source.0.ssh_key").(string)),
	}
}

func resourceSetStackCustomCookbooksSource(d *schema.ResourceData, v *opsworks.Source) error {
	nv := make([]interface{}, 0, 1)
	if v != nil && aws.StringValue(v.Type) != "" {
		m := make(map[string]interface{})
		if v.Type != nil {
			m["type"] = aws.StringValue(v.Type)
		}
		if v.Url != nil {
			m["url"] = aws.StringValue(v.Url)
		}
		if v.Username != nil {
			m["username"] = aws.StringValue(v.Username)
		}
		if v.Revision != nil {
			m["revision"] = aws.StringValue(v.Revision)
		}

		// v.Password and v.SshKey will, on read, contain the placeholder string
		// "*****FILTERED*****", so we ignore it on read and let persist
		// the value already in the state.
		m["password"] = d.Get("custom_cookbooks_source.0.password").(string)
		m["ssh_key"] = d.Get("custom_cookbooks_source.0.ssh_key").(string)

		nv = append(nv, m)
	}

	err := d.Set("custom_cookbooks_source", nv)
	if err != nil {
		// should never happen
		return err
	}
	return nil
}

func resourceStackRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OpsWorksConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	var conErr error
	if v := d.Get("stack_endpoint").(string); v != "" {
		conn, conErr = connForRegion(v, meta)
		if conErr != nil {
			return conErr
		}
	}

	req := &opsworks.DescribeStacksInput{
		StackIds: []*string{
			aws.String(d.Id()),
		},
	}

	log.Printf("[DEBUG] Reading OpsWorks stack: %s", d.Id())

	// notFound represents the number of times we've called DescribeStacks looking
	// for this Stack. If it's not found in the the default region we're in, we
	// check us-east-1 in the event this stack was created with Terraform before
	// version 0.9
	// See https://github.com/hashicorp/terraform/issues/12842
	var notFound int
	var resp *opsworks.DescribeStacksOutput
	var dErr error

	for {
		resp, dErr = conn.DescribeStacks(req)
		if dErr != nil {
			if awserr, ok := dErr.(awserr.Error); ok {
				if awserr.Code() == "ResourceNotFoundException" {
					if notFound < 1 {
						// If we haven't already, try us-east-1, legacy connection
						notFound++
						var connErr error
						conn, connErr = connForRegion("us-east-1", meta) //lintignore:AWSAT003
						if connErr != nil {
							return connErr
						}
						// start again from the top of the FOR loop, but with a client
						// configured to talk to us-east-1
						continue
					}

					// We've tried both the original and us-east-1 endpoint, and the stack
					// is still not found
					log.Printf("[DEBUG] OpsWorks stack (%s) not found", d.Id())
					d.SetId("")
					return nil
				}
				// not ResoureNotFoundException, fall through to returning error
			}
			return dErr
		}
		// If the stack was found, set the stack_endpoint
		if region := aws.StringValue(conn.Config.Region); region != "" {
			log.Printf("[DEBUG] Setting stack_endpoint for (%s) to (%s)", d.Id(), region)
			if err := d.Set("stack_endpoint", region); err != nil {
				log.Printf("[WARN] Error setting stack_endpoint: %s", err)
			}
		}
		log.Printf("[DEBUG] Breaking stack endpoint search, found stack for (%s)", d.Id())
		// Break the FOR loop
		break
	}

	stack := resp.Stacks[0]
	arn := aws.StringValue(stack.Arn)
	d.Set("arn", arn)
	d.Set("agent_version", stack.AgentVersion)
	d.Set("name", stack.Name)
	d.Set("region", stack.Region)
	d.Set("default_instance_profile_arn", stack.DefaultInstanceProfileArn)
	d.Set("service_role_arn", stack.ServiceRoleArn)
	d.Set("default_availability_zone", stack.DefaultAvailabilityZone)
	d.Set("default_os", stack.DefaultOs)
	d.Set("default_root_device_type", stack.DefaultRootDeviceType)
	d.Set("default_ssh_key_name", stack.DefaultSshKeyName)
	d.Set("default_subnet_id", stack.DefaultSubnetId)
	d.Set("hostname_theme", stack.HostnameTheme)
	d.Set("use_custom_cookbooks", stack.UseCustomCookbooks)
	if stack.CustomJson != nil {
		d.Set("custom_json", stack.CustomJson)
	}
	d.Set("use_opsworks_security_groups", stack.UseOpsworksSecurityGroups)
	d.Set("vpc_id", stack.VpcId)
	if color, ok := stack.Attributes["Color"]; ok {
		d.Set("color", color)
	}
	if stack.ConfigurationManager != nil {
		d.Set("configuration_manager_name", stack.ConfigurationManager.Name)
		d.Set("configuration_manager_version", stack.ConfigurationManager.Version)
	}
	if stack.ChefConfiguration != nil {
		d.Set("berkshelf_version", stack.ChefConfiguration.BerkshelfVersion)
		d.Set("manage_berkshelf", stack.ChefConfiguration.ManageBerkshelf)
	}
	err := resourceSetStackCustomCookbooksSource(d, stack.CustomCookbooksSource)
	if err != nil {
		return err
	}

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Opsworks stack (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

// opsworksConn will return a connection for the stack_endpoint in the
// configuration. Stacks can only be accessed or managed within the endpoint
// in which they are created, so we allow users to specify an original endpoint
// for Stacks created before multiple endpoints were offered (Terraform v0.9.0).
// See:
//  - https://github.com/hashicorp/terraform/pull/12688
//  - https://github.com/hashicorp/terraform/issues/12842
func connForRegion(region string, meta interface{}) (*opsworks.OpsWorks, error) {
	originalConn := meta.(*conns.AWSClient).OpsWorksConn

	// Regions are the same, no need to reconfigure
	if aws.StringValue(originalConn.Config.Region) == region {
		return originalConn, nil
	}

	sess, err := conns.NewSessionForRegion(&originalConn.Config, region, meta.(*conns.AWSClient).TerraformVersion)

	if err != nil {
		return nil, fmt.Errorf("error creating AWS session: %w", err)
	}

	return opsworks.New(sess), nil
}

func resourceStackCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OpsWorksConn

	err := resourceStackValidate(d)
	if err != nil {
		return err
	}

	req := &opsworks.CreateStackInput{
		DefaultInstanceProfileArn: aws.String(d.Get("default_instance_profile_arn").(string)),
		Name:                      aws.String(d.Get("name").(string)),
		Region:                    aws.String(d.Get("region").(string)),
		ServiceRoleArn:            aws.String(d.Get("service_role_arn").(string)),
		DefaultOs:                 aws.String(d.Get("default_os").(string)),
		UseOpsworksSecurityGroups: aws.Bool(d.Get("use_opsworks_security_groups").(bool)),
	}
	req.ConfigurationManager = &opsworks.StackConfigurationManager{
		Name:    aws.String(d.Get("configuration_manager_name").(string)),
		Version: aws.String(d.Get("configuration_manager_version").(string)),
	}
	inVpc := false
	if vpcId, ok := d.GetOk("vpc_id"); ok {
		req.VpcId = aws.String(vpcId.(string))
		inVpc = true
	}
	if defaultSubnetId, ok := d.GetOk("default_subnet_id"); ok {
		req.DefaultSubnetId = aws.String(defaultSubnetId.(string))
	}
	if defaultAvailabilityZone, ok := d.GetOk("default_availability_zone"); ok {
		req.DefaultAvailabilityZone = aws.String(defaultAvailabilityZone.(string))
	}
	if defaultRootDeviceType, ok := d.GetOk("default_root_device_type"); ok {
		req.DefaultRootDeviceType = aws.String(defaultRootDeviceType.(string))
	}

	log.Printf("[DEBUG] Creating OpsWorks stack: %s", req)

	var resp *opsworks.CreateStackOutput
	err = resource.Retry(20*time.Minute, func() *resource.RetryError {
		resp, err = conn.CreateStack(req)
		if err != nil {
			// If Terraform is also managing the service IAM role, it may have just been created and not yet be
			// propagated. AWS doesn't provide a machine-readable code for this specific error, so we're forced
			// to do fragile message matching.
			// The full error we're looking for looks something like the following:
			// Service Role Arn: [...] is not yet propagated, please try again in a couple of minutes
			propErr := "not yet propagated"
			trustErr := "not the necessary trust relationship"
			validateErr := "validate IAM role permission"

			if tfawserr.ErrMessageContains(err, "ValidationException", propErr) || tfawserr.ErrMessageContains(err, "ValidationException", trustErr) || tfawserr.ErrMessageContains(err, "ValidationException", validateErr) {
				log.Printf("[INFO] Waiting for service IAM role to propagate")
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		resp, err = conn.CreateStack(req)
	}
	if err != nil {
		return fmt.Errorf("Error creating Opsworks stack: %s", err)
	}

	d.SetId(aws.StringValue(resp.StackId))

	if inVpc && *req.UseOpsworksSecurityGroups {
		// For VPC-based stacks, OpsWorks asynchronously creates some default
		// security groups which must exist before layers can be created.
		// Unfortunately it doesn't tell us what the ids of these are, so
		// we can't actually check for them. Instead, we just wait a nominal
		// amount of time for their creation to complete.
		log.Print("[INFO] Waiting for OpsWorks built-in security groups to be created")
		time.Sleep(securityGroupsCreatedSleepTime)
	}

	return resourceStackUpdate(d, meta)
}

func resourceStackUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OpsWorksConn
	var conErr error
	if v := d.Get("stack_endpoint").(string); v != "" {
		conn, conErr = connForRegion(v, meta)
		if conErr != nil {
			return conErr
		}
	}

	err := resourceStackValidate(d)
	if err != nil {
		return err
	}

	req := &opsworks.UpdateStackInput{
		CustomJson:                aws.String(d.Get("custom_json").(string)),
		DefaultInstanceProfileArn: aws.String(d.Get("default_instance_profile_arn").(string)),
		DefaultRootDeviceType:     aws.String(d.Get("default_root_device_type").(string)),
		DefaultSshKeyName:         aws.String(d.Get("default_ssh_key_name").(string)),
		Name:                      aws.String(d.Get("name").(string)),
		ServiceRoleArn:            aws.String(d.Get("service_role_arn").(string)),
		StackId:                   aws.String(d.Id()),
		UseCustomCookbooks:        aws.Bool(d.Get("use_custom_cookbooks").(bool)),
		UseOpsworksSecurityGroups: aws.Bool(d.Get("use_opsworks_security_groups").(bool)),
		Attributes:                make(map[string]*string),
		CustomCookbooksSource:     resourceStackCustomCookbooksSource(d),
	}
	if v, ok := d.GetOk("agent_version"); ok {
		req.AgentVersion = aws.String(v.(string))
	}
	if v, ok := d.GetOk("default_os"); ok {
		req.DefaultOs = aws.String(v.(string))
	}
	if v, ok := d.GetOk("default_subnet_id"); ok {
		req.DefaultSubnetId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("default_availability_zone"); ok {
		req.DefaultAvailabilityZone = aws.String(v.(string))
	}
	if v, ok := d.GetOk("hostname_theme"); ok {
		req.HostnameTheme = aws.String(v.(string))
	}
	if v, ok := d.GetOk("color"); ok {
		req.Attributes["Color"] = aws.String(v.(string))
	}

	req.ChefConfiguration = &opsworks.ChefConfiguration{
		BerkshelfVersion: aws.String(d.Get("berkshelf_version").(string)),
		ManageBerkshelf:  aws.Bool(d.Get("manage_berkshelf").(bool)),
	}

	req.ConfigurationManager = &opsworks.StackConfigurationManager{
		Name:    aws.String(d.Get("configuration_manager_name").(string)),
		Version: aws.String(d.Get("configuration_manager_version").(string)),
	}

	log.Printf("[DEBUG] Updating OpsWorks stack: %s", req)

	_, err = conn.UpdateStack(req)
	if err != nil {
		return err
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "opsworks",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("stack/%s/", d.Id()),
	}.String()
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Opsworks stack (%s) tags: %s", arn, err)
		}
	}

	return resourceStackRead(d, meta)
}

func resourceStackDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OpsWorksConn
	var conErr error
	if v := d.Get("stack_endpoint").(string); v != "" {
		conn, conErr = connForRegion(v, meta)
		if conErr != nil {
			return conErr
		}
	}

	req := &opsworks.DeleteStackInput{
		StackId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting OpsWorks stack: %s", d.Id())

	_, err := conn.DeleteStack(req)

	if tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
		log.Printf("[DEBUG] OpsWorks Stack (%s) not found to delete; removed from state", d.Id())
		return nil
	}

	if err != nil {
		return fmt.Errorf("while deleting OpsWork Stack (%s, %s): %w", d.Id(), d.Get("name").(string), err)
	}

	// For a stack in a VPC, OpsWorks has created some default security groups
	// in the VPC, which it will now delete.
	// Unfortunately, the security groups are deleted asynchronously and there
	// is no robust way for us to determine when it is done. The VPC itself
	// isn't deletable until the security groups are cleaned up, so this could
	// make 'terraform destroy' fail if the VPC is also managed and we don't
	// wait for the security groups to be deleted.
	// There is no robust way to check for this, so we'll just wait a
	// nominal amount of time.
	_, inVpc := d.GetOk("vpc_id")
	_, useOpsworksDefaultSg := d.GetOk("use_opsworks_security_groups")

	if inVpc && useOpsworksDefaultSg {
		log.Print("[INFO] Waiting for Opsworks built-in security groups to be deleted")
		time.Sleep(securityGroupsDeletedSleepTime)
	}

	return nil
}
