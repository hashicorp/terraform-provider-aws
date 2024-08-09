// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk

import ( // nosemgrep:ci.semgrep.aws.multiple-service-imports
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	sdktypes "github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func settingSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrNamespace: {
				Type:     schema.TypeString,
				Required: true,
			},
			"resource": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrValue: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

const (
	environmentTierWebServer = "WebServer"
	environmentTierWorker    = "Worker"
)

func environmentTier_Values() []string {
	return []string{
		environmentTierWebServer,
		environmentTierWorker,
	}
}

const (
	environmentTierTypeSQSHTTP  = "SQS/HTTP"
	environmentTierTypeStandard = "Standard"
)

var (
	environmentCNAMERegex = regexache.MustCompile(`(^[^.]+)(.\w{2}-\w{4,9}-\d)?\.(elasticbeanstalk\.com|eb\.amazonaws\.com\.cn)$`)
)

// @SDKResource("aws_elastic_beanstalk_environment", name="Environment")
// @Tags(identifierAttribute="arn")
func ResourceEnvironment() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceEnvironmentCreate,
		ReadWithoutTimeout:   resourceEnvironmentRead,
		UpdateWithoutTimeout: resourceEnvironmentUpdate,
		DeleteWithoutTimeout: resourceEnvironmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		SchemaVersion: 1,
		MigrateState:  EnvironmentMigrateState,

		Schema: map[string]*schema.Schema{
			"all_settings": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     settingSchema(),
				Set:      optionSettingValueHash,
			},
			"application": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"autoscaling_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"cname": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cname_prefix": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"endpoint_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instances": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"launch_configurations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"load_balancers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"platform_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"solution_stack_name", "template_name"},
			},
			"poll_interval": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: sdktypes.ValidateDurationBetween(10*time.Second, 3*time.Minute), //nolint:mnd // these are the limits set by AWS
			},
			"queues": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"setting": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     settingSchema(),
				Set:      optionSettingValueHash,
			},
			"solution_stack_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"platform_arn", "template_name"},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"template_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"solution_stack_name", "platform_arn"},
			},
			"tier": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      environmentTierWebServer,
				ValidateFunc: validation.StringInSlice(environmentTier_Values(), false),
			},
			names.AttrTriggers: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"version_label": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"wait_for_ready_timeout": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "20m",
				ValidateDiagFunc: sdktypes.ValidateDuration,
			},
		},
	}
}

func resourceEnvironmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &elasticbeanstalk.CreateEnvironmentInput{
		ApplicationName: aws.String(d.Get("application").(string)),
		EnvironmentName: aws.String(name),
		OptionSettings:  extractOptionSettings(d.Get("setting").(*schema.Set)),
		Tags:            getTagsIn(ctx),
	}

	if v := d.Get(names.AttrDescription); v.(string) != "" {
		input.Description = aws.String(v.(string))
	}

	if v := d.Get("platform_arn"); v.(string) != "" {
		input.PlatformArn = aws.String(v.(string))
	}

	if v := d.Get("solution_stack_name"); v.(string) != "" {
		input.SolutionStackName = aws.String(v.(string))
	}

	if v := d.Get("template_name"); v.(string) != "" {
		input.TemplateName = aws.String(v.(string))
	}

	if v := d.Get("version_label"); v.(string) != "" {
		input.VersionLabel = aws.String(v.(string))
	}

	tier := d.Get("tier").(string)
	var tierType string

	if v := d.Get("cname_prefix"); v.(string) != "" {
		if tier != environmentTierWebServer {
			return sdkdiag.AppendErrorf(diags, "cname_prefix conflicts with tier: %s", tier)
		}

		input.CNAMEPrefix = aws.String(v.(string))
	}

	switch tier {
	case environmentTierWebServer:
		tierType = environmentTierTypeStandard
	case environmentTierWorker:
		tierType = environmentTierTypeSQSHTTP
	}
	input.Tier = &awstypes.EnvironmentTier{
		Name: aws.String(tier),
		Type: aws.String(tierType),
	}

	opTime := time.Now()
	output, err := conn.CreateEnvironment(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Elastic Beanstalk Environment (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.EnvironmentId))

	waitForReadyTimeOut, _, err := sdktypes.Duration(d.Get("wait_for_ready_timeout").(string)).Value()

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing wait_for_ready_timeout: %s", err)
	}

	pollInterval, _, err := sdktypes.Duration(d.Get("poll_interval").(string)).Value()

	if err != nil {
		pollInterval = 0
	}

	if _, err := waitEnvironmentReady(ctx, conn, d.Id(), pollInterval, waitForReadyTimeOut); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Elastic Beanstalk Environment (%s) create: %s", d.Id(), err)
	}

	err = findEnvironmentErrorsByID(ctx, conn, d.Id(), opTime)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Elastic Beanstalk Environment (%s): %s", d.Id(), err)
	}

	return append(diags, resourceEnvironmentRead(ctx, d, meta)...)
}

func resourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	env, err := FindEnvironmentByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Elastic Beanstalk Environment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Environment (%s): %s", d.Id(), err)
	}

	resources, err := conn.DescribeEnvironmentResources(ctx, &elasticbeanstalk.DescribeEnvironmentResourcesInput{
		EnvironmentId: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Environment (%s) resources: %s", d.Id(), err)
	}

	applicationName := aws.ToString(env.ApplicationName)
	environmentName := aws.ToString(env.EnvironmentName)
	configurationSettings, err := findConfigurationSettingsByTwoPartKey(ctx, conn, applicationName, environmentName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Environment (%s) configuration settings: %s", d.Id(), err)
	}

	d.Set("application", applicationName)
	arn := aws.ToString(env.EnvironmentArn)
	d.Set(names.AttrARN, arn)
	if err := d.Set("autoscaling_groups", flattenASG(resources.EnvironmentResources.AutoScalingGroups)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting autoscaling_groups: %s", err)
	}
	cname := aws.ToString(env.CNAME)
	d.Set("cname", cname)
	if cname != "" {
		var cnamePrefix string

		if cnamePrefixMatch := environmentCNAMERegex.FindStringSubmatch(cname); len(cnamePrefixMatch) > 1 {
			cnamePrefix = cnamePrefixMatch[1]
		}

		d.Set("cname_prefix", cnamePrefix)
	} else {
		d.Set("cname_prefix", "")
	}
	d.Set(names.AttrDescription, env.Description)
	d.Set("endpoint_url", env.EndpointURL)
	if err := d.Set("instances", flattenInstances(resources.EnvironmentResources.Instances)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting instances: %s", err)
	}
	if err := d.Set("launch_configurations", flattenLaunchConfigurations(resources.EnvironmentResources.LaunchConfigurations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting launch_configurations: %s", err)
	}
	if err := d.Set("load_balancers", flattenLoadBalancers(resources.EnvironmentResources.LoadBalancers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting load_balancers: %s", err)
	}
	d.Set(names.AttrName, environmentName)
	d.Set("platform_arn", env.PlatformArn)
	if err := d.Set("queues", flattenQueues(resources.EnvironmentResources.Queues)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting queues: %s", err)
	}
	d.Set("solution_stack_name", env.SolutionStackName)
	d.Set("tier", env.Tier.Name)
	if err := d.Set(names.AttrTriggers, flattenTriggers(resources.EnvironmentResources.Triggers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting triggers: %s", err)
	}
	d.Set("version_label", env.VersionLabel)

	allSettings := &schema.Set{F: optionSettingValueHash}
	for _, optionSetting := range configurationSettings.OptionSettings {
		m := map[string]interface{}{}

		if optionSetting.Namespace != nil {
			m[names.AttrNamespace] = aws.ToString(optionSetting.Namespace)
		}

		if optionSetting.OptionName != nil {
			m[names.AttrName] = aws.ToString(optionSetting.OptionName)
		}

		if aws.ToString(optionSetting.Namespace) == "aws:autoscaling:scheduledaction" && optionSetting.ResourceName != nil {
			m["resource"] = aws.ToString(optionSetting.ResourceName)
		}

		if value := aws.ToString(optionSetting.Value); value != "" {
			switch aws.ToString(optionSetting.OptionName) {
			case "SecurityGroups":
				m[names.AttrValue] = dropGeneratedSecurityGroup(ctx, meta.(*conns.AWSClient).EC2Client(ctx), value)
			case "Subnets", "ELBSubnets":
				m[names.AttrValue] = sortValues(value)
			default:
				m[names.AttrValue] = value
			}
		}

		allSettings.Add(m)
	}
	settings := d.Get("setting").(*schema.Set)

	// perform the set operation with only name/namespace as keys, excluding value
	// this is so we override things in the settings resource data key with updated values
	// from the api.  we skip values we didn't know about before because there are so many
	// defaults set by the eb api that we would delete many useful defaults.
	//
	// there is likely a better way to do this
	allSettingsKeySet := schema.NewSet(optionSettingKeyHash, allSettings.List())
	settingsKeySet := schema.NewSet(optionSettingKeyHash, settings.List())
	updatedSettingsKeySet := allSettingsKeySet.Intersection(settingsKeySet)

	updatedSettings := schema.NewSet(optionSettingValueHash, updatedSettingsKeySet.List())

	if err := d.Set("all_settings", allSettings.List()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting all_settings: %s", err)
	}

	if err := d.Set("setting", updatedSettings.List()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting setting: %s", err)
	}

	return diags
}

func resourceEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	waitForReadyTimeOut, _, err := sdktypes.Duration(d.Get("wait_for_ready_timeout").(string)).Value()

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing wait_for_ready_timeout: %s", err)
	}

	pollInterval, _, err := sdktypes.Duration(d.Get("poll_interval").(string)).Value()

	if err != nil {
		pollInterval = 0
	}

	opTime := time.Now()

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll, "poll_interval", "wait_for_ready_timeout") {
		if d.HasChange(names.AttrTagsAll) {
			if _, err := waitEnvironmentReady(ctx, conn, d.Id(), pollInterval, waitForReadyTimeOut); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Elastic Beanstalk Environment (%s) tags update: %s", d.Id(), err)
			}
		}

		input := elasticbeanstalk.UpdateEnvironmentInput{
			EnvironmentId: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("platform_arn") {
			if v, ok := d.GetOk("platform_arn"); ok {
				input.PlatformArn = aws.String(v.(string))
			}
		}

		if d.HasChange("setting") {
			o, n := d.GetChange("setting")
			if o == nil {
				o = &schema.Set{F: optionSettingValueHash}
			}
			if n == nil {
				n = &schema.Set{F: optionSettingValueHash}
			}

			os := o.(*schema.Set)
			ns := n.(*schema.Set)

			rm := extractOptionSettings(os.Difference(ns))
			add := extractOptionSettings(ns.Difference(os))

			// Additions and removals of options are done in a single API call, so we
			// can't do our normal "remove these" and then later "add these", re-adding
			// any updated settings.
			// Because of this, we need to exclude any settings in the "removable"
			// settings that are also found in the "add" settings, otherwise they
			// conflict. Here we loop through all the initial removables from the set
			// difference, and create a new slice `remove` that contains those settings
			// found in `rm` but not in `add`
			var remove []awstypes.ConfigurationOptionSetting
			if len(add) > 0 {
				for _, r := range rm {
					var update = false
					for _, a := range add {
						// ResourceNames are optional. Some defaults come with it, some do
						// not. We need to guard against nil/empty in state as well as
						// nil/empty from the API
						if a.ResourceName != nil {
							if r.ResourceName == nil {
								continue
							}
							if aws.ToString(r.ResourceName) != aws.ToString(a.ResourceName) {
								continue
							}
						}
						if aws.ToString(r.Namespace) == aws.ToString(a.Namespace) &&
							aws.ToString(r.OptionName) == aws.ToString(a.OptionName) {
							log.Printf("[DEBUG] Updating Beanstalk setting (%s::%s) \"%s\" => \"%s\"", *a.Namespace, *a.OptionName, *r.Value, *a.Value)
							update = true
							break
						}
					}
					// Only remove options that are not updates
					if !update {
						remove = append(remove, r)
					}
				}
			} else {
				remove = rm
			}

			for _, elem := range remove {
				input.OptionsToRemove = append(input.OptionsToRemove, awstypes.OptionSpecification{
					Namespace:  elem.Namespace,
					OptionName: elem.OptionName,
				})
			}

			input.OptionSettings = add
		}

		if d.HasChange("solution_stack_name") {
			if v, ok := d.GetOk("solution_stack_name"); ok {
				input.SolutionStackName = aws.String(v.(string))
			}
		}

		if d.HasChange("template_name") {
			if v, ok := d.GetOk("template_name"); ok {
				input.TemplateName = aws.String(v.(string))
			}
		}

		if d.HasChange("version_label") {
			input.VersionLabel = aws.String(d.Get("version_label").(string))
		}

		_, err := conn.UpdateEnvironment(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Environment (%s): %s", d.Id(), err)
		}
	}

	if _, err := waitEnvironmentReady(ctx, conn, d.Id(), pollInterval, waitForReadyTimeOut); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Elastic Beanstalk Environment (%s) update: %s", d.Id(), err)
	}

	err = findEnvironmentErrorsByID(ctx, conn, d.Id(), opTime)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Environment (%s): %s", d.Id(), err)
	}

	return append(diags, resourceEnvironmentRead(ctx, d, meta)...)
}

func resourceEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	waitForReadyTimeOut, _, err := sdktypes.Duration(d.Get("wait_for_ready_timeout").(string)).Value()

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing wait_for_ready_timeout: %s", err)
	}

	pollInterval, _, err := sdktypes.Duration(d.Get("poll_interval").(string)).Value()

	if err != nil {
		pollInterval = 0
	}

	// Environment must be Ready before it can be deleted.
	if _, err := waitEnvironmentReady(ctx, conn, d.Id(), pollInterval, waitForReadyTimeOut); err != nil {
		if tfresource.NotFound(err) {
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "waiting for Elastic Beanstalk Environment (%s) update: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting Elastic Beanstalk Environment: %s", d.Id())
	_, err = conn.TerminateEnvironment(ctx, &elasticbeanstalk.TerminateEnvironmentInput{
		EnvironmentId:      aws.String(d.Id()),
		TerminateResources: aws.Bool(true),
	})

	if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "No Environment found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Elastic Beanstalk Environment (%s): %s", d.Id(), err)
	}

	if _, err := waitEnvironmentDeleted(ctx, conn, d.Id(), pollInterval, waitForReadyTimeOut); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Elastic Beanstalk Environment (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindEnvironmentByID(ctx context.Context, conn *elasticbeanstalk.Client, id string) (*awstypes.EnvironmentDescription, error) {
	input := &elasticbeanstalk.DescribeEnvironmentsInput{
		EnvironmentIds: []string{id},
	}

	output, err := conn.DescribeEnvironments(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Environments) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Environments); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	environment := output.Environments[0]

	if status := environment.Status; status == awstypes.EnvironmentStatusTerminated {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(environment.EnvironmentId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return &environment, nil
}

func findEnvironmentErrorsByID(ctx context.Context, conn *elasticbeanstalk.Client, id string, since time.Time) error {
	input := &elasticbeanstalk.DescribeEventsInput{
		EnvironmentId: aws.String(id),
		Severity:      awstypes.EventSeverityError,
		StartTime:     aws.Time(since),
	}
	var output []awstypes.EventDescription

	pages := elasticbeanstalk.NewDescribeEventsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return err
		}

		output = append(output, page.Events...)
	}

	if len(output) == 0 {
		return nil
	}

	slices.SortFunc(output, func(a, b awstypes.EventDescription) int {
		if a.EventDate.Before(aws.ToTime(b.EventDate)) {
			return -1
		}
		if a.EventDate.After(aws.ToTime(b.EventDate)) {
			return 1
		}
		return 0
	})

	var errs []error

	for _, v := range output {
		errs = append(errs, fmt.Errorf("%s: %s", v.EventDate, aws.ToString(v.Message)))
	}

	return errors.Join(errs...)
}

func findConfigurationSettingsByTwoPartKey(ctx context.Context, conn *elasticbeanstalk.Client, applicationName, environmentName string) (*awstypes.ConfigurationSettingsDescription, error) {
	input := &elasticbeanstalk.DescribeConfigurationSettingsInput{
		ApplicationName: aws.String(applicationName),
		EnvironmentName: aws.String(environmentName),
	}

	output, err := conn.DescribeConfigurationSettings(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ConfigurationSettings) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.ConfigurationSettings); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return &output.ConfigurationSettings[0], nil
}

func statusEnvironment(ctx context.Context, conn *elasticbeanstalk.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindEnvironmentByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitEnvironmentReady(ctx context.Context, conn *elasticbeanstalk.Client, id string, pollInterval, timeout time.Duration) (*awstypes.EnvironmentDescription, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:      enum.Slice(awstypes.EnvironmentStatusLaunching, awstypes.EnvironmentStatusUpdating),
		Target:       enum.Slice(awstypes.EnvironmentStatusReady),
		Refresh:      statusEnvironment(ctx, conn, id),
		Timeout:      timeout,
		Delay:        10 * time.Second,
		PollInterval: pollInterval,
		MinTimeout:   3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.EnvironmentDescription); ok {
		return output, err
	}

	return nil, err
}

func waitEnvironmentDeleted(ctx context.Context, conn *elasticbeanstalk.Client, id string, pollInterval, timeout time.Duration) (*awstypes.EnvironmentDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:      enum.Slice(awstypes.EnvironmentStatusTerminating),
		Target:       []string{},
		Refresh:      statusEnvironment(ctx, conn, id),
		Timeout:      timeout,
		Delay:        10 * time.Second,
		PollInterval: pollInterval,
		MinTimeout:   3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.EnvironmentDescription); ok {
		return output, err
	}

	return nil, err
}

// we use the following two functions to allow us to split out defaults
// as they become overridden from within the template
func optionSettingValueHash(v interface{}) int {
	rd := v.(map[string]interface{})
	namespace := rd[names.AttrNamespace].(string)
	optionName := rd[names.AttrName].(string)
	var resourceName string
	if v, ok := rd["resource"].(string); ok {
		resourceName = v
	}
	value, _ := rd[names.AttrValue].(string)
	value, _ = structure.NormalizeJsonString(value)
	hk := fmt.Sprintf("%s:%s%s=%s", namespace, optionName, resourceName, sortValues(value))
	log.Printf("[DEBUG] Elastic Beanstalk optionSettingValueHash(%#v): %s: hk=%s,hc=%d", v, optionName, hk, create.StringHashcode(hk))
	return create.StringHashcode(hk)
}

func optionSettingKeyHash(v interface{}) int {
	rd := v.(map[string]interface{})
	namespace := rd[names.AttrNamespace].(string)
	optionName := rd[names.AttrName].(string)
	var resourceName string
	if v, ok := rd["resource"].(string); ok {
		resourceName = v
	}
	hk := fmt.Sprintf("%s:%s%s", namespace, optionName, resourceName)
	log.Printf("[DEBUG] Elastic Beanstalk optionSettingKeyHash(%#v): %s: hk=%s,hc=%d", v, optionName, hk, create.StringHashcode(hk))
	return create.StringHashcode(hk)
}

func sortValues(v string) string {
	values := strings.Split(v, ",")
	sort.Strings(values)
	return strings.Join(values, ",")
}

func extractOptionSettings(s *schema.Set) []awstypes.ConfigurationOptionSetting {
	settings := []awstypes.ConfigurationOptionSetting{}

	if s != nil {
		for _, setting := range s.List() {
			optionSetting := awstypes.ConfigurationOptionSetting{
				Namespace:  aws.String(setting.(map[string]interface{})[names.AttrNamespace].(string)),
				OptionName: aws.String(setting.(map[string]interface{})[names.AttrName].(string)),
				Value:      aws.String(setting.(map[string]interface{})[names.AttrValue].(string)),
			}
			if aws.ToString(optionSetting.Namespace) == "aws:autoscaling:scheduledaction" {
				if v, ok := setting.(map[string]interface{})["resource"].(string); ok && v != "" {
					optionSetting.ResourceName = aws.String(v)
				}
			}
			settings = append(settings, optionSetting)
		}
	}

	return settings
}

func dropGeneratedSecurityGroup(ctx context.Context, conn *ec2.Client, settingValue string) string {
	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: strings.Split(settingValue, ","),
	}

	securityGroup, err := tfec2.FindSecurityGroups(ctx, conn, input)

	if err != nil {
		return settingValue
	}

	var legitGroups []string
	for _, group := range securityGroup {
		if !strings.HasPrefix(aws.ToString(group.GroupName), "awseb") {
			legitGroups = append(legitGroups, aws.ToString(group.GroupId))
		}
	}

	sort.Strings(legitGroups)

	return strings.Join(legitGroups, ",")
}
