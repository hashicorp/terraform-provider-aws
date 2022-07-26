package gamelift

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	fleetCreatedDefaultTimeout = 70 * time.Minute
	FleetDeletedDefaultTimeout = 20 * time.Minute
)

func ResourceFleet() *schema.Resource {
	return &schema.Resource{
		Create: resourceFleetCreate,
		Read:   resourceFleetRead,
		Update: resourceFleetUpdate,
		Delete: resourceFleetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(fleetCreatedDefaultTimeout),
			Delete: schema.DefaultTimeout(FleetDeletedDefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"build_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"build_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"build_id", "script_id"},
			},
			"certificate_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Computed: true,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      gamelift.CertificateTypeDisabled,
							ValidateFunc: validation.StringInSlice(gamelift.CertificateType_Values(), false),
						},
					},
				},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"ec2_inbound_permission": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				MaxItems: 50,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumber,
						},
						"ip_range": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidCIDRNetworkAddress,
						},
						"protocol": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(gamelift.IpProtocol_Values(), false),
						},
						"to_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumber,
						},
					},
				},
			},
			"ec2_instance_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(gamelift.EC2InstanceType_Values(), false),
			},
			"fleet_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      gamelift.FleetTypeOnDemand,
				ValidateFunc: validation.StringInSlice(gamelift.FleetType_Values(), false),
			},
			"instance_role_arn": {
				Type:         schema.TypeString,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
				Optional:     true,
			},
			"log_paths": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"metric_groups": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 255),
				},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"new_game_session_protection_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      gamelift.ProtectionPolicyNoProtection,
				ValidateFunc: validation.StringInSlice(gamelift.ProtectionPolicy_Values(), false),
			},
			"operating_system": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_creation_limit_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"new_game_sessions_per_creator": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"policy_period_in_minutes": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
					},
				},
			},
			"runtime_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"game_session_activation_timeout_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 600),
						},
						"max_concurrent_game_session_activations": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 2147483647),
						},
						"server_process": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 50,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"concurrent_executions": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"launch_path": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
									"parameters": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
								},
							},
						},
					},
				},
			},
			"script_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"script_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"build_id", "script_id"},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceFleetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GameLiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &gamelift.CreateFleetInput{
		EC2InstanceType: aws.String(d.Get("ec2_instance_type").(string)),
		Name:            aws.String(d.Get("name").(string)),
		Tags:            Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("build_id"); ok {
		input.BuildId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("script_id"); ok {
		input.ScriptId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("fleet_type"); ok {
		input.FleetType = aws.String(v.(string))
	}
	if v, ok := d.GetOk("ec2_inbound_permission"); ok {
		input.EC2InboundPermissions = expandIPPermissions(v.(*schema.Set))
	}

	if v, ok := d.GetOk("instance_role_arn"); ok {
		input.InstanceRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("metric_groups"); ok {
		input.MetricGroups = flex.ExpandStringList(v.([]interface{}))
	}
	if v, ok := d.GetOk("new_game_session_protection_policy"); ok {
		input.NewGameSessionProtectionPolicy = aws.String(v.(string))
	}
	if v, ok := d.GetOk("resource_creation_limit_policy"); ok {
		input.ResourceCreationLimitPolicy = expandResourceCreationLimitPolicy(v.([]interface{}))
	}
	if v, ok := d.GetOk("runtime_configuration"); ok {
		input.RuntimeConfiguration = expandRuntimeConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("certificate_configuration"); ok {
		input.CertificateConfiguration = expandCertificateConfiguration(v.([]interface{}))
	}

	log.Printf("[INFO] Creating GameLift Fleet: %s", input)
	var out *gamelift.CreateFleetOutput
	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error
		out, err = conn.CreateFleet(input)

		if tfawserr.ErrMessageContains(err, gamelift.ErrCodeInvalidRequestException, "GameLift is not authorized to perform") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		out, err = conn.CreateFleet(input)
	}

	if err != nil {
		return fmt.Errorf("error creating GameLift Fleet (%s): %w", d.Get("name").(string), err)
	}

	d.SetId(aws.StringValue(out.FleetAttributes.FleetId))

	if _, err := waitFleetActive(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for GameLift Fleet (%s) to active: %w", d.Id(), err)
	}

	return resourceFleetRead(d, meta)
}

func resourceFleetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GameLiftConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	log.Printf("[INFO] Describing GameLift Fleet: %s", d.Id())
	fleet, err := FindFleetByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GameLift Fleet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading GameLift Fleet (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(fleet.FleetArn)
	d.Set("build_arn", fleet.BuildArn)
	d.Set("build_id", fleet.BuildId)
	d.Set("description", fleet.Description)
	d.Set("arn", arn)
	d.Set("log_paths", aws.StringValueSlice(fleet.LogPaths))
	d.Set("metric_groups", flex.FlattenStringList(fleet.MetricGroups))
	d.Set("name", fleet.Name)
	d.Set("fleet_type", fleet.FleetType)
	d.Set("instance_role_arn", fleet.InstanceRoleArn)
	d.Set("ec2_instance_type", fleet.InstanceType)
	d.Set("new_game_session_protection_policy", fleet.NewGameSessionProtectionPolicy)
	d.Set("operating_system", fleet.OperatingSystem)
	d.Set("script_arn", fleet.ScriptArn)
	d.Set("script_id", fleet.ScriptId)

	if err := d.Set("certificate_configuration", flattenCertificateConfiguration(fleet.CertificateConfiguration)); err != nil {
		return fmt.Errorf("error setting certificate_configuration: %w", err)
	}

	if err := d.Set("resource_creation_limit_policy", flattenResourceCreationLimitPolicy(fleet.ResourceCreationLimitPolicy)); err != nil {
		return fmt.Errorf("error setting resource_creation_limit_policy: %w", err)
	}

	portInput := &gamelift.DescribeFleetPortSettingsInput{
		FleetId: aws.String(d.Id()),
	}

	portConfig, err := conn.DescribeFleetPortSettings(portInput)
	if err != nil {
		return fmt.Errorf("error reading for GameLift Fleet ec2 inbound permission (%s): %w", d.Id(), err)
	}

	if err := d.Set("ec2_inbound_permission", flattenIPPermissions(portConfig.InboundPermissions)); err != nil {
		return fmt.Errorf("error setting ec2_inbound_permission: %w", err)
	}

	tags, err := ListTags(conn, arn)

	if tfawserr.ErrMessageContains(err, gamelift.ErrCodeInvalidRequestException, fmt.Sprintf("Resource %s is not in a taggable state", d.Id())) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing tags for Game Lift Fleet (%s): %w", arn, err)
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

func resourceFleetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GameLiftConn

	log.Printf("[INFO] Updating GameLift Fleet: %s", d.Id())

	if d.HasChanges("description", "metric_groups", "name", "new_game_session_protection_policy", "resource_creation_limit_policy") {
		_, err := conn.UpdateFleetAttributes(&gamelift.UpdateFleetAttributesInput{
			Description:                    aws.String(d.Get("description").(string)),
			FleetId:                        aws.String(d.Id()),
			MetricGroups:                   flex.ExpandStringList(d.Get("metric_groups").([]interface{})),
			Name:                           aws.String(d.Get("name").(string)),
			NewGameSessionProtectionPolicy: aws.String(d.Get("new_game_session_protection_policy").(string)),
			ResourceCreationLimitPolicy:    expandResourceCreationLimitPolicy(d.Get("resource_creation_limit_policy").([]interface{})),
		})
		if err != nil {
			return fmt.Errorf("error updating for GameLift Fleet attributes (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("ec2_inbound_permission") {
		oldPerms, newPerms := d.GetChange("ec2_inbound_permission")
		authorizations, revocations := DiffPortSettings(oldPerms.(*schema.Set).List(), newPerms.(*schema.Set).List())

		_, err := conn.UpdateFleetPortSettings(&gamelift.UpdateFleetPortSettingsInput{
			FleetId:                         aws.String(d.Id()),
			InboundPermissionAuthorizations: authorizations,
			InboundPermissionRevocations:    revocations,
		})
		if err != nil {
			return fmt.Errorf("error updating for GameLift Fleet port settings (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("runtime_configuration") {
		_, err := conn.UpdateRuntimeConfiguration(&gamelift.UpdateRuntimeConfigurationInput{
			FleetId:              aws.String(d.Id()),
			RuntimeConfiguration: expandRuntimeConfiguration(d.Get("runtime_configuration").([]interface{})),
		})
		if err != nil {
			return fmt.Errorf("error updating for GameLift Fleet runtime configuration (%s): %w", d.Id(), err)
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Game Lift Fleet (%s) tags: %w", arn, err)
		}
	}

	return resourceFleetRead(d, meta)
}

func resourceFleetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GameLiftConn

	log.Printf("[INFO] Deleting GameLift Fleet: %s", d.Id())
	// It can take ~ 1 hr as GameLift will keep retrying on errors like
	// invalid launch path and remain in state when it can't be deleted :/
	input := &gamelift.DeleteFleetInput{
		FleetId: aws.String(d.Id()),
	}
	err := resource.Retry(60*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteFleet(input)
		if err != nil {
			msg := fmt.Sprintf("Cannot delete fleet %s that is in status of ", d.Id())
			if tfawserr.ErrMessageContains(err, gamelift.ErrCodeInvalidRequestException, msg) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteFleet(input)
	}
	if err != nil {
		if tfawserr.ErrCodeEquals(err, gamelift.ErrCodeNotFoundException) {
			return nil
		}
		return fmt.Errorf("Error deleting GameLift fleet: %w", err)
	}

	if _, err := waitFleetTerminated(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for GameLift Fleet (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}

func expandIPPermissions(cfgs *schema.Set) []*gamelift.IpPermission {
	if cfgs.Len() < 1 {
		return []*gamelift.IpPermission{}
	}

	perms := make([]*gamelift.IpPermission, cfgs.Len())
	for i, rawCfg := range cfgs.List() {
		cfg := rawCfg.(map[string]interface{})
		perms[i] = expandIPPermission(cfg)
	}
	return perms
}

func expandIPPermission(cfg map[string]interface{}) *gamelift.IpPermission {
	return &gamelift.IpPermission{
		FromPort: aws.Int64(int64(cfg["from_port"].(int))),
		IpRange:  aws.String(cfg["ip_range"].(string)),
		Protocol: aws.String(cfg["protocol"].(string)),
		ToPort:   aws.Int64(int64(cfg["to_port"].(int))),
	}
}

func flattenIPPermissions(apiObjects []*gamelift.IpPermission) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		if v := flattenIPPermission(apiObject); len(v) > 0 {
			tfList = append(tfList, v)
		}
	}

	return tfList
}

func flattenIPPermission(apiObject *gamelift.IpPermission) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["from_port"] = aws.Int64Value(apiObject.FromPort)
	tfMap["to_port"] = aws.Int64Value(apiObject.ToPort)
	tfMap["protocol"] = aws.StringValue(apiObject.Protocol)
	tfMap["ip_range"] = aws.StringValue(apiObject.IpRange)

	return tfMap
}

func expandResourceCreationLimitPolicy(cfg []interface{}) *gamelift.ResourceCreationLimitPolicy {
	if len(cfg) < 1 {
		return nil
	}
	out := gamelift.ResourceCreationLimitPolicy{}
	m := cfg[0].(map[string]interface{})

	if v, ok := m["new_game_sessions_per_creator"]; ok {
		out.NewGameSessionsPerCreator = aws.Int64(int64(v.(int)))
	}
	if v, ok := m["policy_period_in_minutes"]; ok {
		out.PolicyPeriodInMinutes = aws.Int64(int64(v.(int)))
	}

	return &out
}

func flattenResourceCreationLimitPolicy(policy *gamelift.ResourceCreationLimitPolicy) []interface{} {
	if policy == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	m["new_game_sessions_per_creator"] = aws.Int64Value(policy.NewGameSessionsPerCreator)
	m["policy_period_in_minutes"] = aws.Int64Value(policy.PolicyPeriodInMinutes)

	return []interface{}{m}
}

func expandRuntimeConfiguration(cfg []interface{}) *gamelift.RuntimeConfiguration {
	if len(cfg) < 1 {
		return nil
	}
	out := gamelift.RuntimeConfiguration{}
	m := cfg[0].(map[string]interface{})

	if v, ok := m["game_session_activation_timeout_seconds"].(int); ok && v > 0 {
		out.GameSessionActivationTimeoutSeconds = aws.Int64(int64(v))
	}
	if v, ok := m["max_concurrent_game_session_activations"].(int); ok && v > 0 {
		out.MaxConcurrentGameSessionActivations = aws.Int64(int64(v))
	}
	if v, ok := m["server_process"]; ok {
		out.ServerProcesses = expandServerProcesses(v.([]interface{}))
	}

	return &out
}

func expandServerProcesses(cfgs []interface{}) []*gamelift.ServerProcess {
	if len(cfgs) < 1 {
		return []*gamelift.ServerProcess{}
	}

	processes := make([]*gamelift.ServerProcess, len(cfgs))
	for i, rawCfg := range cfgs {
		cfg := rawCfg.(map[string]interface{})
		process := &gamelift.ServerProcess{
			ConcurrentExecutions: aws.Int64(int64(cfg["concurrent_executions"].(int))),
			LaunchPath:           aws.String(cfg["launch_path"].(string)),
		}
		if v, ok := cfg["parameters"].(string); ok && len(v) > 0 {
			process.Parameters = aws.String(v)
		}
		processes[i] = process
	}
	return processes
}

func expandCertificateConfiguration(cfg []interface{}) *gamelift.CertificateConfiguration {
	if len(cfg) < 1 {
		return nil
	}
	out := gamelift.CertificateConfiguration{}
	m := cfg[0].(map[string]interface{})

	if v, ok := m["certificate_type"].(string); ok {
		out.CertificateType = aws.String(v)
	}

	return &out
}

func flattenCertificateConfiguration(config *gamelift.CertificateConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	m["certificate_type"] = aws.StringValue(config.CertificateType)

	return []interface{}{m}
}

func DiffPortSettings(oldPerms, newPerms []interface{}) (a []*gamelift.IpPermission, r []*gamelift.IpPermission) {
OUTER:
	for i, op := range oldPerms {
		oldPerm := op.(map[string]interface{})
		for j, np := range newPerms {
			newPerm := np.(map[string]interface{})

			// Remove permissions which have not changed so we're not wasting
			// API calls for removal & subseq. addition of the same ones
			if reflect.DeepEqual(oldPerm, newPerm) {
				oldPerms = append(oldPerms[:i], oldPerms[i+1:]...)
				newPerms = append(newPerms[:j], newPerms[j+1:]...)
				continue OUTER
			}
		}

		// Add what's left for revocation
		r = append(r, expandIPPermission(oldPerm))
	}

	for _, np := range newPerms {
		newPerm := np.(map[string]interface{})
		// Add what's left for authorization
		a = append(a, expandIPPermission(newPerm))
	}
	return
}
