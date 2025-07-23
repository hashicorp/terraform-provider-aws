// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk

import ( // nosemgrep:ci.semgrep.aws.multiple-service-imports
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"slices"
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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
	environmentCNAMERegex = regexache.MustCompile(`(^[^.]+)(.\w{2}-\w{4,9}-\d{1,2})?\.(elasticbeanstalk\.com|eb\.amazonaws\.com\.cn)$`)
)

// @SDKResource("aws_elastic_beanstalk_environment", name="Environment")
// @Tags(identifierAttribute="arn")
func resourceEnvironment() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceEnvironmentCreate,
		ReadWithoutTimeout:   resourceEnvironmentRead,
		UpdateWithoutTimeout: resourceEnvironmentUpdate,
		DeleteWithoutTimeout: resourceEnvironmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 1,
		MigrateState:  EnvironmentMigrateState,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"all_settings": {
					Type:     schema.TypeSet,
					Computed: true,
					Elem:     settingSchema(),
					Set:      hashSettingsValue,
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
					Set:      hashSettingsValue,
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
			}
		},
	}
}

func resourceEnvironmentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &elasticbeanstalk.CreateEnvironmentInput{
		ApplicationName: aws.String(d.Get("application").(string)),
		EnvironmentName: aws.String(name),
		Tags:            getTagsIn(ctx),
	}

	if v := d.Get(names.AttrDescription); v.(string) != "" {
		input.Description = aws.String(v.(string))
	}

	if v := d.Get("platform_arn"); v.(string) != "" {
		input.PlatformArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("setting"); ok && v.(*schema.Set).Len() > 0 {
		input.OptionSettings = expandConfigurationOptionSettings(v.(*schema.Set).List())
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

	if v := d.Get("cname_prefix"); v.(string) != "" {
		if tier != environmentTierWebServer {
			return sdkdiag.AppendErrorf(diags, "cname_prefix conflicts with tier: %s", tier)
		}

		input.CNAMEPrefix = aws.String(v.(string))
	}

	var tierType string
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
		return sdkdiag.AppendFromErr(diags, err)
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

func resourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	env, err := findEnvironmentByID(ctx, conn, d.Id())

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
	input := &elasticbeanstalk.DescribeConfigurationSettingsInput{
		ApplicationName: aws.String(applicationName),
		EnvironmentName: aws.String(environmentName),
	}
	configurationSettings, err := findConfigurationSettings(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Environment (%s) configuration settings: %s", d.Id(), err)
	}

	d.Set("application", applicationName)
	d.Set(names.AttrARN, env.EnvironmentArn)
	if err := d.Set("autoscaling_groups", flattenAutoScalingGroups(resources.EnvironmentResources.AutoScalingGroups)); err != nil {
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

	var configuredSettings []any
	if v, ok := d.GetOk("setting"); ok && v.(*schema.Set).Len() > 0 {
		configuredSettings = v.(*schema.Set).List()
	}
	apiSettings := flattenConfigurationOptionSettings(ctx, meta, configurationSettings.OptionSettings)
	var settings []any

	for _, apiSetting := range apiSettings {
		tfMap := apiSetting.(map[string]any)
		isMatch := func(v any) bool {
			m := v.(map[string]any)

			return m[names.AttrNamespace].(string) == tfMap[names.AttrNamespace].(string) &&
				m[names.AttrName].(string) == tfMap[names.AttrName].(string) &&
				m["resource"].(string) == tfMap["resource"].(string)
		}
		if slices.ContainsFunc(configuredSettings, isMatch) {
			settings = append(settings, apiSetting)
		}
	}

	if err := d.Set("all_settings", apiSettings); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting all_settings: %s", err)
	}

	if err := d.Set("setting", settings); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting setting: %s", err)
	}

	return diags
}

func resourceEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	waitForReadyTimeOut, _, err := sdktypes.Duration(d.Get("wait_for_ready_timeout").(string)).Value()

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
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
			os, ns := o.(*schema.Set), n.(*schema.Set)
			add, del := expandConfigurationOptionSettings(ns.Difference(os).List()), expandConfigurationOptionSettings(os.Difference(ns).List())

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
				for _, r := range del {
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
						if aws.ToString(r.Namespace) == aws.ToString(a.Namespace) && aws.ToString(r.OptionName) == aws.ToString(a.OptionName) {
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
				remove = del
			}

			for _, v := range remove {
				input.OptionsToRemove = append(input.OptionsToRemove, awstypes.OptionSpecification{
					Namespace:  v.Namespace,
					OptionName: v.OptionName,
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

func resourceEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	waitForReadyTimeOut, _, err := sdktypes.Duration(d.Get("wait_for_ready_timeout").(string)).Value()

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
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

	if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "No Environment found") {
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

func findEnvironmentByID(ctx context.Context, conn *elasticbeanstalk.Client, id string) (*awstypes.EnvironmentDescription, error) {
	input := &elasticbeanstalk.DescribeEnvironmentsInput{
		EnvironmentIds: []string{id},
	}
	output, err := findEnvironment(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.Status; status == awstypes.EnvironmentStatusTerminated {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.EnvironmentId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findEnvironment(ctx context.Context, conn *elasticbeanstalk.Client, input *elasticbeanstalk.DescribeEnvironmentsInput) (*awstypes.EnvironmentDescription, error) {
	output, err := findEnvironments(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findEnvironments(ctx context.Context, conn *elasticbeanstalk.Client, input *elasticbeanstalk.DescribeEnvironmentsInput) ([]awstypes.EnvironmentDescription, error) {
	var output []awstypes.EnvironmentDescription

	err := describeEnvironmentsPages(ctx, conn, input, func(page *elasticbeanstalk.DescribeEnvironmentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.Environments...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findEnvironmentErrorsByID(ctx context.Context, conn *elasticbeanstalk.Client, id string, since time.Time) error {
	input := &elasticbeanstalk.DescribeEventsInput{
		EnvironmentId: aws.String(id),
		Severity:      awstypes.EventSeverityError,
		StartTime:     aws.Time(since),
	}
	output, err := findEvents(ctx, conn, input)

	if err != nil {
		return err
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
		errs = append(errs, fmt.Errorf("%s: %s", aws.ToTime(v.EventDate), aws.ToString(v.Message)))
	}

	return errors.Join(errs...)
}

func findEvents(ctx context.Context, conn *elasticbeanstalk.Client, input *elasticbeanstalk.DescribeEventsInput) ([]awstypes.EventDescription, error) {
	var output []awstypes.EventDescription

	pages := elasticbeanstalk.NewDescribeEventsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Events...)
	}

	return output, nil
}

func statusEnvironment(ctx context.Context, conn *elasticbeanstalk.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findEnvironmentByID(ctx, conn, id)

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

func hashSettingsValue(v any) int {
	tfMap := v.(map[string]any)
	var str strings.Builder

	str.WriteString(tfMap[names.AttrNamespace].(string))
	str.WriteRune(':')
	str.WriteString(tfMap[names.AttrName].(string))
	var resourceName string
	if v, ok := tfMap["resource"].(string); ok {
		resourceName = v
	}
	str.WriteString(resourceName)
	str.WriteRune('=')
	if value := tfMap[names.AttrValue].(string); json.Valid([]byte(value)) {
		value, _ = structure.NormalizeJsonString(value)
		str.WriteString(value)
	} else {
		values := strings.Split(value, ",")
		slices.Sort(values)
		str.WriteString(strings.Join(values, ","))
	}

	return create.StringHashcode(str.String())
}

func expandConfigurationOptionSettings(tfList []any) []awstypes.ConfigurationOptionSetting {
	apiObjects := []awstypes.ConfigurationOptionSetting{}

	if tfList == nil {
		return apiObjects
	}

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObject := awstypes.ConfigurationOptionSetting{
			Namespace:  aws.String(tfMap[names.AttrNamespace].(string)),
			OptionName: aws.String(tfMap[names.AttrName].(string)),
			Value:      aws.String(tfMap[names.AttrValue].(string)),
		}

		if aws.ToString(apiObject.Namespace) == "aws:autoscaling:scheduledaction" {
			if v, ok := tfMap["resource"].(string); ok && v != "" {
				apiObject.ResourceName = aws.String(v)
			}
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenConfigurationOptionSettings(ctx context.Context, meta any, apiObjects []awstypes.ConfigurationOptionSetting) []any {
	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.Namespace != nil {
			tfMap[names.AttrNamespace] = aws.ToString(apiObject.Namespace)
		}

		if apiObject.OptionName != nil {
			tfMap[names.AttrName] = aws.ToString(apiObject.OptionName)
		}

		if aws.ToString(apiObject.Namespace) == "aws:autoscaling:scheduledaction" && apiObject.ResourceName != nil {
			tfMap["resource"] = aws.ToString(apiObject.ResourceName)
		} else {
			tfMap["resource"] = ""
		}

		if value := aws.ToString(apiObject.Value); value != "" {
			switch aws.ToString(apiObject.OptionName) {
			case "SecurityGroups":
				tfMap[names.AttrValue] = dropGeneratedSecurityGroup(ctx, meta.(*conns.AWSClient).EC2Client(ctx), value)
			case "Subnets", "ELBSubnets":
				values := strings.Split(value, ",")
				slices.Sort(values)
				tfMap[names.AttrValue] = strings.Join(values, ",")
			default:
				tfMap[names.AttrValue] = value
			}
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
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

	slices.Sort(legitGroups)

	return strings.Join(legitGroups, ",")
}

func flattenAutoScalingGroups(apiObjects []awstypes.AutoScalingGroup) []string {
	return tfslices.ApplyToAll(apiObjects, func(v awstypes.AutoScalingGroup) string {
		return aws.ToString(v.Name)
	})
}

func flattenLoadBalancers(apiObjects []awstypes.LoadBalancer) []string {
	return tfslices.ApplyToAll(apiObjects, func(v awstypes.LoadBalancer) string {
		return aws.ToString(v.Name)
	})
}

func flattenInstances(apiObjects []awstypes.Instance) []string {
	return tfslices.ApplyToAll(apiObjects, func(v awstypes.Instance) string {
		return aws.ToString(v.Id)
	})
}

func flattenLaunchConfigurations(apiObjects []awstypes.LaunchConfiguration) []string {
	return tfslices.ApplyToAll(apiObjects, func(v awstypes.LaunchConfiguration) string {
		return aws.ToString(v.Name)
	})
}

func flattenQueues(apiObjects []awstypes.Queue) []string {
	return tfslices.ApplyToAll(apiObjects, func(v awstypes.Queue) string {
		return aws.ToString(v.URL)
	})
}

func flattenTriggers(apiObjects []awstypes.Trigger) []string {
	return tfslices.ApplyToAll(apiObjects, func(v awstypes.Trigger) string {
		return aws.ToString(v.Name)
	})
}
