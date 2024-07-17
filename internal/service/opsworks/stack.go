// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	securityGroupsCreatedSleepTime = 30 * time.Second
	securityGroupsDeletedSleepTime = 30 * time.Second
)

// @SDKResource("aws_opsworks_stack", name="Stack")
// @Tags
func ResourceStack() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStackCreate,
		ReadWithoutTimeout:   resourceStackRead,
		UpdateWithoutTimeout: resourceStackUpdate,
		DeleteWithoutTimeout: resourceStackDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"berkshelf_version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  defaultBerkshelfVersion,
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
						names.AttrPassword: {
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
						names.AttrType: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(opsworks.SourceType_Values(), false),
						},
						names.AttrURL: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrUsername: {
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
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{names.AttrVPCID},
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
				RequiredWith: []string{names.AttrVPCID},
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrRegion: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			names.AttrServiceRoleARN: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"stack_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
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
			names.AttrVPCID: {
				Type:          schema.TypeString,
				ForceNew:      true,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"default_availability_zone"},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceStackCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpsWorksConn(ctx)

	name := d.Get(names.AttrName).(string)
	region := d.Get(names.AttrRegion).(string)
	input := &opsworks.CreateStackInput{
		ChefConfiguration: &opsworks.ChefConfiguration{
			ManageBerkshelf: aws.Bool(d.Get("manage_berkshelf").(bool)),
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
		ServiceRoleArn:            aws.String(d.Get(names.AttrServiceRoleARN).(string)),
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

	if d.Get("manage_berkshelf").(bool) {
		input.ChefConfiguration.BerkshelfVersion = aws.String(d.Get("berkshelf_version").(string))
	}

	if v, ok := d.GetOk(names.AttrVPCID); ok {
		input.VpcId = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhen(ctx, d.Timeout(schema.TimeoutCreate),
		func() (interface{}, error) {
			return conn.CreateStackWithContext(ctx, input)
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
		return sdkdiag.AppendErrorf(diags, "creating OpsWorks Stack (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*opsworks.CreateStackOutput).StackId))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   opsworks.ServiceName,
		Region:    region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("stack/%s/", d.Id()),
	}.String()

	if err := createTags(ctx, conn, arn, getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting OpsWorks Stack (%s) tags: %s", arn, err)
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

	return append(diags, resourceStackRead(ctx, d, meta)...)
}

func resourceStackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error
	conn := meta.(*conns.AWSClient).OpsWorksConn(ctx)

	if v, ok := d.GetOk("stack_endpoint"); ok {
		conn = meta.(*conns.AWSClient).OpsWorksConnForRegion(ctx, v.(string))
	}

	stack, err := FindStackByID(ctx, conn, d.Id())

	if tfresource.NotFound(err) { // nosemgrep:ci.semgrep.errors.notfound-without-err-checks
		// If it's not found in the default region we're in, we check us-east-1
		// in the event this stack was created with Terraform before version 0.9.
		// See https://github.com/hashicorp/terraform/issues/12842.
		conn = meta.(*conns.AWSClient).OpsWorksConnForRegion(ctx, names.USEast1RegionID)

		stack, err = FindStackByID(ctx, conn, d.Id())
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] OpsWorks Stack %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpsWorks Stack (%s): %s", d.Id(), err)
	}

	// If the stack was found, set the stack_endpoint.
	if v := aws.StringValue(conn.Config.Region); v != "" {
		d.Set("stack_endpoint", v)
	}

	d.Set("agent_version", stack.AgentVersion)
	arn := aws.StringValue(stack.Arn)
	d.Set(names.AttrARN, arn)
	if stack.ChefConfiguration != nil {
		if v := aws.StringValue(stack.ChefConfiguration.BerkshelfVersion); v != "" {
			d.Set("berkshelf_version", v)
		} else {
			d.Set("berkshelf_version", defaultBerkshelfVersion)
		}
		d.Set("manage_berkshelf", stack.ChefConfiguration.ManageBerkshelf)
	}
	if color, ok := stack.Attributes[opsworks.StackAttributesKeysColor]; ok {
		d.Set("color", color)
	}
	if stack.ConfigurationManager != nil {
		d.Set("configuration_manager_name", stack.ConfigurationManager.Name)
		d.Set("configuration_manager_version", stack.ConfigurationManager.Version)
	}
	if stack.CustomCookbooksSource != nil {
		tfMap := flattenSource(stack.CustomCookbooksSource)

		// CustomCookbooksSource.Password and CustomCookbooksSource.SshKey will, on read, contain the placeholder string "*****FILTERED*****",
		// so we ignore it on read and let persist the value already in the state.
		if v, ok := d.GetOk("custom_cookbooks_source"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			v := v.([]interface{})[0].(map[string]interface{})

			tfMap[names.AttrPassword] = v[names.AttrPassword]
			tfMap["ssh_key"] = v["ssh_key"]
		}

		if err := d.Set("custom_cookbooks_source", []interface{}{tfMap}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting custom_cookbooks_source: %s", err)
		}
	} else {
		d.Set("custom_cookbooks_source", nil)
	}
	d.Set("custom_json", stack.CustomJson)
	d.Set("default_availability_zone", stack.DefaultAvailabilityZone)
	d.Set("default_instance_profile_arn", stack.DefaultInstanceProfileArn)
	d.Set("default_os", stack.DefaultOs)
	d.Set("default_root_device_type", stack.DefaultRootDeviceType)
	d.Set("default_ssh_key_name", stack.DefaultSshKeyName)
	d.Set("default_subnet_id", stack.DefaultSubnetId)
	d.Set("hostname_theme", stack.HostnameTheme)
	d.Set(names.AttrName, stack.Name)
	d.Set(names.AttrRegion, stack.Region)
	d.Set(names.AttrServiceRoleARN, stack.ServiceRoleArn)
	d.Set("use_custom_cookbooks", stack.UseCustomCookbooks)
	d.Set("use_opsworks_security_groups", stack.UseOpsworksSecurityGroups)
	d.Set(names.AttrVPCID, stack.VpcId)

	tags, err := listTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for OpsWorks Stack (%s): %s", arn, err)
	}

	setTagsOut(ctx, Tags(tags))

	return diags
}

func resourceStackUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error
	conn := meta.(*conns.AWSClient).OpsWorksConn(ctx)

	if v, ok := d.GetOk("stack_endpoint"); ok {
		conn = meta.(*conns.AWSClient).OpsWorksConnForRegion(ctx, v.(string))
	}

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &opsworks.UpdateStackInput{
			StackId: aws.String(d.Id()),
		}

		if d.HasChange("agent_version") {
			input.AgentVersion = aws.String(d.Get("agent_version").(string))
		}

		if d.HasChanges("berkshelf_version", "manage_berkshelf") {
			input.ChefConfiguration = &opsworks.ChefConfiguration{
				ManageBerkshelf: aws.Bool(d.Get("manage_berkshelf").(bool)),
			}

			if d.Get("manage_berkshelf").(bool) {
				input.ChefConfiguration.BerkshelfVersion = aws.String(d.Get("berkshelf_version").(string))
			}
		}

		if d.HasChange("color") {
			input.Attributes = aws.StringMap(map[string]string{
				opsworks.StackAttributesKeysColor: d.Get("color").(string),
			})
		}

		if d.HasChanges("configuration_manager_name", "configuration_manager_version") {
			input.ConfigurationManager = &opsworks.StackConfigurationManager{
				Name:    aws.String(d.Get("configuration_manager_name").(string)),
				Version: aws.String(d.Get("configuration_manager_version").(string)),
			}
		}

		if d.HasChange("custom_json") {
			input.CustomJson = aws.String(d.Get("custom_json").(string))
		}

		if d.HasChange("custom_cookbooks_source") {
			if v, ok := d.GetOk("custom_cookbooks_source"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.CustomCookbooksSource = expandSource(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChange("default_availability_zone") {
			input.DefaultAvailabilityZone = aws.String(d.Get("default_availability_zone").(string))
		}

		if d.HasChange("default_instance_profile_arn") {
			input.DefaultInstanceProfileArn = aws.String(d.Get("default_instance_profile_arn").(string))
		}

		if d.HasChange("default_os") {
			input.DefaultOs = aws.String(d.Get("default_os").(string))
		}

		if d.HasChange("default_root_device_type") {
			input.DefaultRootDeviceType = aws.String(d.Get("default_root_device_type").(string))
		}

		if d.HasChange("default_ssh_key_name") {
			input.DefaultSshKeyName = aws.String(d.Get("default_ssh_key_name").(string))
		}

		if d.HasChange("default_subnet_id") {
			input.DefaultSubnetId = aws.String(d.Get("default_subnet_id").(string))
		}

		if d.HasChange("hostname_theme") {
			input.HostnameTheme = aws.String(d.Get("hostname_theme").(string))
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		if d.HasChange(names.AttrServiceRoleARN) {
			input.ServiceRoleArn = aws.String(d.Get(names.AttrServiceRoleARN).(string))
		}

		if d.HasChange("use_custom_cookbooks") {
			input.UseCustomCookbooks = aws.Bool(d.Get("use_custom_cookbooks").(bool))
		}

		if d.HasChange("use_opsworks_security_groups") {
			input.UseOpsworksSecurityGroups = aws.Bool(d.Get("use_opsworks_security_groups").(bool))
		}

		_, err = conn.UpdateStackWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating OpsWorks Stack (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrTagsAll) {
		o, n := d.GetChange(names.AttrTagsAll)

		if err := updateTags(ctx, conn, d.Get(names.AttrARN).(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating OpsWorks Stack (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceStackRead(ctx, d, meta)...)
}

func resourceStackDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error
	conn := meta.(*conns.AWSClient).OpsWorksConn(ctx)

	if v, ok := d.GetOk("stack_endpoint"); ok {
		conn = meta.(*conns.AWSClient).OpsWorksConnForRegion(ctx, v.(string))
	}

	log.Printf("[DEBUG] Deleting OpsWorks Stack: %s", d.Id())
	_, err = conn.DeleteStackWithContext(ctx, &opsworks.DeleteStackInput{
		StackId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting OpsWork Stack (%s): %s", d.Id(), err)
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
	if _, ok := d.GetOk(names.AttrVPCID); ok {
		if _, ok := d.GetOk("use_opsworks_security_groups"); ok {
			log.Print("[INFO] Waiting for Opsworks built-in security groups to be deleted")
			time.Sleep(securityGroupsDeletedSleepTime)
		}
	}

	return diags
}

func FindStackByID(ctx context.Context, conn *opsworks.OpsWorks, id string) (*opsworks.Stack, error) {
	input := &opsworks.DescribeStacksInput{
		StackIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeStacksWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, opsworks.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
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

	if v, ok := tfMap[names.AttrPassword].(string); ok && v != "" {
		apiObject.Password = aws.String(v)
	}

	if v, ok := tfMap["revision"].(string); ok && v != "" {
		apiObject.Revision = aws.String(v)
	}

	if v, ok := tfMap["ssh_key"].(string); ok && v != "" {
		apiObject.SshKey = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	if v, ok := tfMap[names.AttrURL].(string); ok && v != "" {
		apiObject.Url = aws.String(v)
	}

	if v, ok := tfMap[names.AttrUsername].(string); ok && v != "" {
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
		tfMap[names.AttrPassword] = aws.StringValue(v)
	}

	if v := apiObject.Revision; v != nil {
		tfMap["revision"] = aws.StringValue(v)
	}

	if v := apiObject.SshKey; v != nil {
		tfMap["ssh_key"] = aws.StringValue(v)
	}

	if v := apiObject.Type; v != nil {
		tfMap[names.AttrType] = aws.StringValue(v)
	}

	if v := apiObject.Url; v != nil {
		tfMap[names.AttrURL] = aws.StringValue(v)
	}

	if v := apiObject.Username; v != nil {
		tfMap[names.AttrUsername] = aws.StringValue(v)
	}

	return tfMap
}
