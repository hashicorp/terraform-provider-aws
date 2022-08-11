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

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
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
			"berkshelf_version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "3.2.0",
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
			"custom_cookbooks_source": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
					},
				},
			},
			"custom_json": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
			},
			"default_availability_zone": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"default_availability_zone", "vpc_id"},
			},
			"default_instance_profile_arn": {
				Type:     schema.TypeString,
				Required: true,
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
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				RequiredWith: []string{"vpc_id"},
			},
			"hostname_theme": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Layer_Dependent",
			},
			"manage_berkshelf": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
			"stack_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
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
				Type:         schema.TypeString,
				ForceNew:     true,
				Computed:     true,
				Optional:     true,
				ExactlyOneOf: []string{"default_availability_zone", "vpc_id"},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceStackCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OpsWorksConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	region := d.Get("region").(string)
	input := &opsworks.CreateStackInput{
		ChefConfiguration: &opsworks.ChefConfiguration{
			BerkshelfVersion: aws.String(d.Get("berkshelf_version").(string)),
			ManageBerkshelf:  aws.Bool(d.Get("manage_berkshelf").(bool)),
		},
		ConfigurationManager: &opsworks.StackConfigurationManager{
			Name:    aws.String(d.Get("configuration_manager_name").(string)),
			Version: aws.String(d.Get("configuration_manager_version").(string)),
		},
		DefaultInstanceProfileArn: aws.String(d.Get("default_instance_profile_arn").(string)),
		DefaultOs:                 aws.String(d.Get("default_os").(string)),
		HostnameTheme:             aws.String(d.Get("hostname_theme").(string)),
		Name:                      aws.String(name),
		Region:                    aws.String(region),
		ServiceRoleArn:            aws.String(d.Get("service_role_arn").(string)),
		UseCustomCookbooks:        aws.Bool(d.Get("use_custom_cookbooks").(bool)),
		UseOpsworksSecurityGroups: aws.Bool(d.Get("use_opsworks_security_groups").(bool)),
	}

	if v, ok := d.GetOk("agent_version"); ok {
		input.AgentVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("color"); ok {
		input.Attributes = aws.StringMap(map[string]string{
			opsworks.StackAttributesKeysColor: v.(string),
		})
	}

	if v, ok := d.GetOk("custom_cookbooks_source"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.CustomCookbooksSource = expandSource(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("custom_json"); ok {
		input.CustomJson = aws.String(v.(string))
	}

	if v, ok := d.GetOk("default_availability_zone"); ok {
		input.DefaultAvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("default_root_device_type"); ok {
		input.DefaultRootDeviceType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("default_ssh_key_name"); ok {
		input.DefaultSshKeyName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("default_subnet_id"); ok {
		input.DefaultSubnetId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc_id"); ok {
		input.VpcId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating OpsWorks Stack: %s", input)
	outputRaw, err := tfresource.RetryWhen(d.Timeout(schema.TimeoutCreate),
		func() (interface{}, error) {
			return conn.CreateStack(input)
		},
		func(err error) (bool, error) {
			// If Terraform is also managing the service IAM role, it may have just been created and not yet be
			// propagated. AWS doesn't provide a machine-readable code for this specific error, so we're forced
			// to do fragile message matching.
			// The full error we're looking for looks something like the following:
			// Service Role Arn: [...] is not yet propagated, please try again in a couple of minutes
			if tfawserr.ErrMessageContains(err, opsworks.ErrCodeValidationException, "not yet propagated") ||
				tfawserr.ErrMessageContains(err, opsworks.ErrCodeValidationException, "not the necessary trust relationship") ||
				tfawserr.ErrMessageContains(err, opsworks.ErrCodeValidationException, "validate IAM role permission") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return fmt.Errorf("creating OpsWorks Stack (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*opsworks.CreateStackOutput).StackId))

	if len(tags) > 0 {
		arn := arn.ARN{
			Partition: meta.(*conns.AWSClient).Partition,
			Service:   opsworks.ServiceName,
			Region:    region,
			AccountID: meta.(*conns.AWSClient).AccountID,
			Resource:  fmt.Sprintf("stack/%s/", d.Id()),
		}.String()

		if err := UpdateTags(conn, arn, nil, tags); err != nil {
			return fmt.Errorf("adding OpsWorks Stack (%s) tags: %w", arn, err)
		}
	}

	if aws.StringValue(input.VpcId) != "" && aws.BoolValue(input.UseOpsworksSecurityGroups) {
		// For VPC-based stacks, OpsWorks asynchronously creates some default
		// security groups which must exist before layers can be created.
		// Unfortunately it doesn't tell us what the ids of these are, so
		// we can't actually check for them. Instead, we just wait a nominal
		// amount of time for their creation to complete.
		log.Print("[INFO] Waiting for OpsWorks built-in security groups to be created")
		time.Sleep(securityGroupsCreatedSleepTime)
	}

	return resourceStackRead(d, meta)
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

	if stack.CustomCookbooksSource != nil {
		tfMap := flattenSource(stack.CustomCookbooksSource)

		// CustomCookbooksSource.Password and CustomCookbooksSource.SshKey will, on read, contain the placeholder string "*****FILTERED*****",
		// so we ignore it on read and let persist the value already in the state.
		if v, ok := d.GetOk("custom_cookbooks_source"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			v := v.([]interface{})[0].(map[string]interface{})

			tfMap["password"] = v["password"]
			tfMap["ssh_key"] = v["ssh_key"]
		}

		if err := d.Set("custom_cookbooks_source", []interface{}{tfMap}); err != nil {
			return fmt.Errorf("setting custom_cookbooks_source: %w", err)
		}
	} else {
		d.Set("custom_cookbooks_source", nil)
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

func resourceStackUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OpsWorksConn
	var conErr error
	if v := d.Get("stack_endpoint").(string); v != "" {
		conn, conErr = connForRegion(v, meta)
		if conErr != nil {
			return conErr
		}
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
	}

	if v, ok := d.GetOk("agent_version"); ok {
		req.AgentVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("custom_cookbooks_source"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		req.CustomCookbooksSource = expandSource(v.([]interface{})[0].(map[string]interface{}))
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

	_, err := conn.UpdateStack(req)
	if err != nil {
		return err
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    d.Get("region").(string),
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

func FindStackByID(conn *opsworks.OpsWorks, id string) (*opsworks.Stack, error) {
	input := &opsworks.DescribeStacksInput{
		StackIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeStacks(input)

	if tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Stacks) == 0 || output.Stacks[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Stacks); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.Stacks[0], nil
}

func expandSource(tfMap map[string]interface{}) *opsworks.Source {
	if tfMap == nil {
		return nil
	}

	apiObject := &opsworks.Source{}

	if v, ok := tfMap["password"].(string); ok && v != "" {
		apiObject.Password = aws.String(v)
	}

	if v, ok := tfMap["revision"].(string); ok && v != "" {
		apiObject.Revision = aws.String(v)
	}

	if v, ok := tfMap["ssh_key"].(string); ok && v != "" {
		apiObject.SshKey = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	if v, ok := tfMap["url"].(string); ok && v != "" {
		apiObject.Url = aws.String(v)
	}

	if v, ok := tfMap["username"].(string); ok && v != "" {
		apiObject.Username = aws.String(v)
	}

	return apiObject
}

func flattenSource(apiObject *opsworks.Source) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Password; v != nil {
		tfMap["password"] = aws.StringValue(v)
	}

	if v := apiObject.Revision; v != nil {
		tfMap["revision"] = aws.StringValue(v)
	}

	if v := apiObject.SshKey; v != nil {
		tfMap["ssh_key"] = aws.StringValue(v)
	}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	if v := apiObject.Url; v != nil {
		tfMap["url"] = aws.StringValue(v)
	}

	if v := apiObject.Username; v != nil {
		tfMap["username"] = aws.StringValue(v)
	}

	return tfMap
}

// opsworksConn will return a connection for the stack_endpoint in the
// configuration. Stacks can only be accessed or managed within the endpoint
// in which they are created, so we allow users to specify an original endpoint
// for Stacks created before multiple endpoints were offered (Terraform v0.9.0).
// See:
//   - https://github.com/hashicorp/terraform/pull/12688
//   - https://github.com/hashicorp/terraform/issues/12842
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
