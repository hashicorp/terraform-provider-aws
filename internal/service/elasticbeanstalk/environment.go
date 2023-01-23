package elasticbeanstalk

import ( // nosemgrep:ci.aws-sdk-go-multiple-service-imports
	"context"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func resourceOptionSetting() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"namespace": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"value": {
				Type:     schema.TypeString,
				Required: true,
			},
			"resource": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"application": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"version_label": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cname": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cname_prefix": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"endpoint_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tier": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "WebServer",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					switch value {
					case
						"Worker",
						"WebServer":
						return
					}
					errors = append(errors, fmt.Errorf("%s is not a valid tier. Valid options are WebServer or Worker", value))
					return
				},
				ForceNew: true,
			},
			"setting": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     resourceOptionSetting(),
				Set:      optionSettingValueHash,
			},
			"all_settings": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     resourceOptionSetting(),
				Set:      optionSettingValueHash,
			},
			"solution_stack_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"platform_arn", "template_name"},
			},
			"platform_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"solution_stack_name", "template_name"},
			},
			"template_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"solution_stack_name", "platform_arn"},
			},
			"wait_for_ready_timeout": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "20m",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					duration, err := time.ParseDuration(value)
					if err != nil {
						errors = append(errors, fmt.Errorf(
							"%q cannot be parsed as a duration: %s", k, err))
					}
					if duration < 0 {
						errors = append(errors, fmt.Errorf(
							"%q must be greater than zero", k))
					}
					return
				},
			},
			"poll_interval": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					duration, err := time.ParseDuration(value)
					if err != nil {
						errors = append(errors, fmt.Errorf(
							"%q cannot be parsed as a duration: %s", k, err))
					}
					if duration < 10*time.Second || duration > 60*time.Second {
						errors = append(errors, fmt.Errorf(
							"%q must be between 10s and 180s", k))
					}
					return
				},
			},
			"autoscaling_groups": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
			"queues": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"triggers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceEnvironmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	// Get values from config
	name := d.Get("name").(string)
	cnamePrefix := d.Get("cname_prefix").(string)
	tier := d.Get("tier").(string)
	app := d.Get("application").(string)
	desc := d.Get("description").(string)
	version := d.Get("version_label").(string)
	settings := d.Get("setting").(*schema.Set)
	solutionStack := d.Get("solution_stack_name").(string)
	platformArn := d.Get("platform_arn").(string)
	templateName := d.Get("template_name").(string)

	createOpts := elasticbeanstalk.CreateEnvironmentInput{
		EnvironmentName: aws.String(name),
		ApplicationName: aws.String(app),
		OptionSettings:  extractOptionSettings(settings),
		Tags:            Tags(tags.IgnoreElasticbeanstalk()),
	}

	if desc != "" {
		createOpts.Description = aws.String(desc)
	}

	if cnamePrefix != "" {
		if tier != "WebServer" {
			return sdkdiag.AppendErrorf(diags, "cannot set cname_prefix for tier: %s", tier)
		}
		createOpts.CNAMEPrefix = aws.String(cnamePrefix)
	}

	if tier != "" {
		var tierType string

		switch tier {
		case "WebServer":
			tierType = "Standard"
		case "Worker":
			tierType = "SQS/HTTP"
		}
		environmentTier := elasticbeanstalk.EnvironmentTier{
			Name: aws.String(tier),
			Type: aws.String(tierType),
		}
		createOpts.Tier = &environmentTier
	}

	if solutionStack != "" {
		createOpts.SolutionStackName = aws.String(solutionStack)
	}

	if platformArn != "" {
		createOpts.PlatformArn = aws.String(platformArn)
	}

	if templateName != "" {
		createOpts.TemplateName = aws.String(templateName)
	}

	if version != "" {
		createOpts.VersionLabel = aws.String(version)
	}

	// Get the current time to filter getEnvironmentErrors messages
	t := time.Now()
	resp, err := conn.CreateEnvironmentWithContext(ctx, &createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Elastic Beanstalk Environment (%s): %s", name, err)
	}

	// Assign the application name as the resource ID
	d.SetId(aws.StringValue(resp.EnvironmentId))

	waitForReadyTimeOut, err := time.ParseDuration(d.Get("wait_for_ready_timeout").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Elastic Beanstalk Environment (%s): %s", name, err)
	}

	pollInterval, err := time.ParseDuration(d.Get("poll_interval").(string))
	if err != nil {
		pollInterval = 0
		log.Printf("[WARN] Error parsing poll_interval, using default backoff")
	}

	err = waitForEnvironmentReady(ctx, conn, d.Id(), waitForReadyTimeOut, pollInterval, t)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Elastic Beanstalk Environment (%s): waiting for completion: %s", name, err)
	}

	envErrors, err := getEnvironmentErrors(ctx, conn, d.Id(), t)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Elastic Beanstalk Environment (%s): %s", name, err)
	}
	if envErrors != nil {
		return sdkdiag.AppendErrorf(diags, "creating Elastic Beanstalk Environment (%s): %s", name, envErrors)
	}

	return append(diags, resourceEnvironmentRead(ctx, d, meta)...)
}

func resourceEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn()

	envID := d.Id()

	var hasChange bool

	updateOpts := elasticbeanstalk.UpdateEnvironmentInput{
		EnvironmentId: aws.String(envID),
	}

	if d.HasChange("description") {
		hasChange = true
		updateOpts.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("solution_stack_name") {
		hasChange = true
		if v, ok := d.GetOk("solution_stack_name"); ok {
			updateOpts.SolutionStackName = aws.String(v.(string))
		}
	}

	if d.HasChange("setting") {
		hasChange = true
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
		var remove []*elasticbeanstalk.ConfigurationOptionSetting
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
						if aws.StringValue(r.ResourceName) != aws.StringValue(a.ResourceName) {
							continue
						}
					}
					if aws.StringValue(r.Namespace) == aws.StringValue(a.Namespace) &&
						aws.StringValue(r.OptionName) == aws.StringValue(a.OptionName) {
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
			updateOpts.OptionsToRemove = append(updateOpts.OptionsToRemove, &elasticbeanstalk.OptionSpecification{
				Namespace:  elem.Namespace,
				OptionName: elem.OptionName,
			})
		}

		updateOpts.OptionSettings = add
	}

	if d.HasChange("platform_arn") {
		hasChange = true
		if v, ok := d.GetOk("platform_arn"); ok {
			updateOpts.PlatformArn = aws.String(v.(string))
		}
	}

	if d.HasChange("template_name") {
		hasChange = true
		if v, ok := d.GetOk("template_name"); ok {
			updateOpts.TemplateName = aws.String(v.(string))
		}
	}

	if d.HasChange("version_label") {
		hasChange = true
		updateOpts.VersionLabel = aws.String(d.Get("version_label").(string))
	}

	if hasChange {
		// Get the current time to filter getEnvironmentErrors messages
		t := time.Now()
		_, err := conn.UpdateEnvironmentWithContext(ctx, &updateOpts)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Environment (%s): %s", d.Id(), err)
		}

		waitForReadyTimeOut, err := time.ParseDuration(d.Get("wait_for_ready_timeout").(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Environment (%s): %s", d.Id(), err)
		}
		pollInterval, err := time.ParseDuration(d.Get("poll_interval").(string))
		if err != nil {
			pollInterval = 0
			log.Printf("[WARN] Error parsing poll_interval, using default backoff")
		}

		err = waitForEnvironmentReady(ctx, conn, d.Id(), waitForReadyTimeOut, pollInterval, t)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Environment (%s): waiting for completion: %s", d.Id(), err)
		}

		envErrors, err := getEnvironmentErrors(ctx, conn, d.Id(), t)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Environment (%s): %s", d.Id(), err)
		}
		if envErrors != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Environment (%s): %s", d.Id(), envErrors)
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		// Get the current time to filter getEnvironmentErrors messages
		t := time.Now()
		if err := UpdateTags(ctx, conn, arn, o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk environment (%s): updating tags: %s", arn, err)
		}

		waitForReadyTimeOut, err := time.ParseDuration(d.Get("wait_for_ready_timeout").(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Environment (%s): %s", d.Id(), err)
		}
		pollInterval, err := time.ParseDuration(d.Get("poll_interval").(string))
		if err != nil {
			pollInterval = 0
			log.Printf("[WARN] Error parsing poll_interval, using default backoff")
		}

		err = waitForEnvironmentReady(ctx, conn, d.Id(), waitForReadyTimeOut, pollInterval, t)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Environment (%s): waiting for completion: %s", d.Id(), err)
		}

		envErrors, err := getEnvironmentErrors(ctx, conn, d.Id(), t)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Environment (%s): %s", d.Id(), err)
		}
		if envErrors != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Environment (%s): %s", d.Id(), envErrors)
		}
	}

	return append(diags, resourceEnvironmentRead(ctx, d, meta)...)
}

func resourceEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	envID := d.Id()

	log.Printf("[DEBUG] Elastic Beanstalk environment read %s: id %s", d.Get("name").(string), d.Id())

	resp, err := conn.DescribeEnvironmentsWithContext(ctx, &elasticbeanstalk.DescribeEnvironmentsInput{
		EnvironmentIds: []*string{aws.String(envID)},
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Environment (%s): %s", d.Id(), err)
	}

	if len(resp.Environments) == 0 {
		log.Printf("[DEBUG] Elastic Beanstalk environment properties: could not find environment %s", d.Id())

		d.SetId("")
		return diags
	} else if len(resp.Environments) != 1 {
		return sdkdiag.AppendErrorf(diags, "reading application properties: found %d environments, expected 1", len(resp.Environments))
	}

	env := resp.Environments[0]

	if aws.StringValue(env.Status) == elasticbeanstalk.EnvironmentStatusTerminated {
		log.Printf("[DEBUG] Elastic Beanstalk environment %s was terminated", d.Id())

		d.SetId("")
		return diags
	}

	resources, err := conn.DescribeEnvironmentResourcesWithContext(ctx, &elasticbeanstalk.DescribeEnvironmentResourcesInput{
		EnvironmentId: aws.String(envID),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Environment (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(env.EnvironmentArn)
	d.Set("arn", arn)

	d.Set("name", env.EnvironmentName)

	d.Set("application", env.ApplicationName)

	d.Set("description", env.Description)

	d.Set("cname", env.CNAME)

	d.Set("version_label", env.VersionLabel)

	d.Set("tier", env.Tier.Name)

	if env.CNAME != nil {
		beanstalkCnamePrefixRegexp := regexp.MustCompile(`(^[^.]+)(.\w{2}-\w{4,9}-\d)?\.(elasticbeanstalk\.com|eb\.amazonaws\.com\.cn)$`)
		var cnamePrefix string
		cnamePrefixMatch := beanstalkCnamePrefixRegexp.FindStringSubmatch(*env.CNAME)

		if cnamePrefixMatch == nil {
			cnamePrefix = ""
		} else {
			cnamePrefix = cnamePrefixMatch[1]
		}

		d.Set("cname_prefix", cnamePrefix)
	} else {
		d.Set("cname_prefix", "")
	}

	d.Set("solution_stack_name", env.SolutionStackName)

	d.Set("platform_arn", env.PlatformArn)

	if err := d.Set("autoscaling_groups", flattenASG(resources.EnvironmentResources.AutoScalingGroups)); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Environment (%s): %s", d.Id(), err)
	}

	if err := d.Set("instances", flattenInstances(resources.EnvironmentResources.Instances)); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Environment (%s): %s", d.Id(), err)
	}
	if err := d.Set("launch_configurations", flattenLc(resources.EnvironmentResources.LaunchConfigurations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Environment (%s): %s", d.Id(), err)
	}
	if err := d.Set("load_balancers", flattenELB(resources.EnvironmentResources.LoadBalancers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Environment (%s): %s", d.Id(), err)
	}
	if err := d.Set("queues", flattenSQS(resources.EnvironmentResources.Queues)); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Environment (%s): %s", d.Id(), err)
	}
	if err := d.Set("triggers", flattenTrigger(resources.EnvironmentResources.Triggers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Environment (%s): %s", d.Id(), err)
	}
	if err := d.Set("endpoint_url", env.EndpointURL); err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Environment (%s): %s", d.Id(), err)
	}

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Elastic Beanstalk environment (%s): %s", arn, err)
	}

	tags = tags.IgnoreElasticbeanstalk().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return append(diags, resourceEnvironmentSettingsRead(ctx, d, meta)...)
}

func fetchEnvironmentSettings(ctx context.Context, d *schema.ResourceData, meta interface{}) (*schema.Set, error) {
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn()

	app := d.Get("application").(string)
	name := d.Get("name").(string)

	resp, err := conn.DescribeConfigurationSettingsWithContext(ctx, &elasticbeanstalk.DescribeConfigurationSettingsInput{
		ApplicationName: aws.String(app),
		EnvironmentName: aws.String(name),
	})

	if err != nil {
		return nil, err
	}

	if len(resp.ConfigurationSettings) != 1 {
		return nil, fmt.Errorf("Error reading environment settings: received %d settings groups, expected 1", len(resp.ConfigurationSettings))
	}

	settings := &schema.Set{F: optionSettingValueHash}
	for _, optionSetting := range resp.ConfigurationSettings[0].OptionSettings {
		m := map[string]interface{}{}

		if optionSetting.Namespace != nil {
			m["namespace"] = aws.StringValue(optionSetting.Namespace)
		} else {
			return nil, fmt.Errorf("Error reading environment settings: option setting with no namespace: %v", optionSetting)
		}

		if optionSetting.OptionName != nil {
			m["name"] = aws.StringValue(optionSetting.OptionName)
		} else {
			return nil, fmt.Errorf("Error reading environment settings: option setting with no name: %v", optionSetting)
		}

		if aws.StringValue(optionSetting.Namespace) == "aws:autoscaling:scheduledaction" && optionSetting.ResourceName != nil {
			m["resource"] = aws.StringValue(optionSetting.ResourceName)
		}

		if optionSetting.Value != nil {
			switch *optionSetting.OptionName {
			case "SecurityGroups":
				m["value"] = dropGeneratedSecurityGroup(ctx, aws.StringValue(optionSetting.Value), meta)
			case "Subnets", "ELBSubnets":
				m["value"] = sortValues(aws.StringValue(optionSetting.Value))
			default:
				m["value"] = aws.StringValue(optionSetting.Value)
			}
		}

		settings.Add(m)
	}

	return settings, nil
}

func resourceEnvironmentSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("[DEBUG] Elastic Beanstalk environment settings read %s: id %s", d.Get("name").(string), d.Id())

	allSettings, err := fetchEnvironmentSettings(ctx, d, meta)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
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
		return sdkdiag.AppendFromErr(diags, err)
	}

	if err := d.Set("setting", updatedSettings.List()); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func resourceEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn()

	waitForReadyTimeOut, err := time.ParseDuration(d.Get("wait_for_ready_timeout").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Elastic Beanstalk Environment (%s): %s", d.Id(), err)
	}
	pollInterval, err := time.ParseDuration(d.Get("poll_interval").(string))
	if err != nil {
		pollInterval = 0
		log.Printf("[WARN] Error parsing poll_interval, using default backoff")
	}

	// The Environment needs to be in a Ready state before it can be terminated
	err = waitForEnvironmentReadyIgnoreErrorEvents(ctx, conn, d.Id(), waitForReadyTimeOut, pollInterval)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Elastic Beanstalk Environment (%s): waiting for ready state: %s", d.Id(), err)
	}

	if err := DeleteEnvironment(ctx, conn, d.Id(), waitForReadyTimeOut, pollInterval); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Elastic Beanstalk Environment (%s): %s", d.Id(), err)
	}
	return diags
}

func DeleteEnvironment(ctx context.Context, conn *elasticbeanstalk.ElasticBeanstalk, id string, timeout, pollInterval time.Duration) error {
	opts := elasticbeanstalk.TerminateEnvironmentInput{
		EnvironmentId:      aws.String(id),
		TerminateResources: aws.Bool(true),
	}
	log.Printf("[DEBUG] Elastic Beanstalk Environment terminate opts: %s", opts)

	_, err := conn.TerminateEnvironmentWithContext(ctx, &opts)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, "InvalidConfiguration.NotFound") || tfawserr.ErrCodeEquals(err, "ValidationError") {
			log.Printf("[DEBUG] Elastic Beanstalk Environment %q not found", id)
			return nil
		}
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:      []string{"Terminating"},
		Target:       []string{"Terminated"},
		Refresh:      environmentIgnoreErrorEventsStateRefreshFunc(ctx, conn, id),
		Timeout:      timeout,
		Delay:        10 * time.Second,
		PollInterval: pollInterval,
		MinTimeout:   3 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("waiting for completion: %w", err)
	}

	return nil
}

func waitForEnvironmentReady(ctx context.Context, conn *elasticbeanstalk.ElasticBeanstalk, id string, timeout, pollInterval time.Duration, startTime time.Time) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			elasticbeanstalk.EnvironmentStatusLaunching,
			elasticbeanstalk.EnvironmentStatusUpdating,
		},
		Target:       []string{elasticbeanstalk.EnvironmentStatusReady},
		Refresh:      environmentStateRefreshFunc(ctx, conn, id, startTime),
		Timeout:      timeout,
		Delay:        10 * time.Second,
		PollInterval: pollInterval,
		MinTimeout:   3 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

func waitForEnvironmentReadyIgnoreErrorEvents(ctx context.Context, conn *elasticbeanstalk.ElasticBeanstalk, id string, timeout, pollInterval time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			elasticbeanstalk.EnvironmentStatusLaunching,
			elasticbeanstalk.EnvironmentStatusTerminating,
			elasticbeanstalk.EnvironmentStatusUpdating,
		},
		Target: []string{
			elasticbeanstalk.EnvironmentStatusReady,
			elasticbeanstalk.EnvironmentStatusTerminated,
		},
		Refresh:      environmentIgnoreErrorEventsStateRefreshFunc(ctx, conn, id),
		Timeout:      timeout,
		Delay:        10 * time.Second,
		PollInterval: pollInterval,
		MinTimeout:   3 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	return err
}

// environmentStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch
// the creation of the Beanstalk Environment
func environmentStateRefreshFunc(ctx context.Context, conn *elasticbeanstalk.ElasticBeanstalk, environmentID string, t time.Time) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeEnvironmentsWithContext(ctx, &elasticbeanstalk.DescribeEnvironmentsInput{
			EnvironmentIds: []*string{aws.String(environmentID)},
		})
		if err != nil {
			log.Printf("[Err] Error waiting for Elastic Beanstalk Environment state: %s", err)
			return -1, "failed", fmt.Errorf("error waiting for Elastic Beanstalk Environment state: %w", err)
		}

		if resp == nil || len(resp.Environments) == 0 {
			// Sometimes AWS just has consistency issues and doesn't see
			// our instance yet. Return an empty state.
			return nil, "", nil
		}

		var env *elasticbeanstalk.EnvironmentDescription
		for _, e := range resp.Environments {
			if environmentID == aws.StringValue(e.EnvironmentId) {
				env = e
			}
		}

		if env == nil {
			return -1, "failed", fmt.Errorf("Error finding Elastic Beanstalk Environment, environment not found")
		}

		envErrors, err := getEnvironmentErrors(ctx, conn, environmentID, t)
		if err != nil {
			return -1, "failed", err
		}
		if envErrors != nil {
			return -1, "failed", envErrors
		}

		return env, aws.StringValue(env.Status), nil
	}
}

func environmentIgnoreErrorEventsStateRefreshFunc(ctx context.Context, conn *elasticbeanstalk.ElasticBeanstalk, environmentID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeEnvironmentsWithContext(ctx, &elasticbeanstalk.DescribeEnvironmentsInput{
			EnvironmentIds: []*string{aws.String(environmentID)},
		})
		if err != nil {
			log.Printf("[Err] Error waiting for Elastic Beanstalk Environment state: %s", err)
			return -1, "failed", fmt.Errorf("error waiting for Elastic Beanstalk Environment state: %w", err)
		}

		if resp == nil || len(resp.Environments) == 0 {
			// Sometimes AWS just has consistency issues and doesn't see
			// our instance yet. Return an empty state.
			return nil, "", nil
		}

		var env *elasticbeanstalk.EnvironmentDescription
		for _, e := range resp.Environments {
			if environmentID == aws.StringValue(e.EnvironmentId) {
				env = e
			}
		}

		if env == nil {
			return -1, "failed", fmt.Errorf("Error finding Elastic Beanstalk Environment, environment not found")
		}

		return env, aws.StringValue(env.Status), nil
	}
}

// we use the following two functions to allow us to split out defaults
// as they become overridden from within the template
func optionSettingValueHash(v interface{}) int {
	rd := v.(map[string]interface{})
	namespace := rd["namespace"].(string)
	optionName := rd["name"].(string)
	var resourceName string
	if v, ok := rd["resource"].(string); ok {
		resourceName = v
	}
	value, _ := rd["value"].(string)
	value, _ = structure.NormalizeJsonString(value)
	hk := fmt.Sprintf("%s:%s%s=%s", namespace, optionName, resourceName, sortValues(value))
	log.Printf("[DEBUG] Elastic Beanstalk optionSettingValueHash(%#v): %s: hk=%s,hc=%d", v, optionName, hk, create.StringHashcode(hk))
	return create.StringHashcode(hk)
}

func optionSettingKeyHash(v interface{}) int {
	rd := v.(map[string]interface{})
	namespace := rd["namespace"].(string)
	optionName := rd["name"].(string)
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

func extractOptionSettings(s *schema.Set) []*elasticbeanstalk.ConfigurationOptionSetting {
	settings := []*elasticbeanstalk.ConfigurationOptionSetting{}

	if s != nil {
		for _, setting := range s.List() {
			optionSetting := elasticbeanstalk.ConfigurationOptionSetting{
				Namespace:  aws.String(setting.(map[string]interface{})["namespace"].(string)),
				OptionName: aws.String(setting.(map[string]interface{})["name"].(string)),
				Value:      aws.String(setting.(map[string]interface{})["value"].(string)),
			}
			if aws.StringValue(optionSetting.Namespace) == "aws:autoscaling:scheduledaction" {
				if v, ok := setting.(map[string]interface{})["resource"].(string); ok && v != "" {
					optionSetting.ResourceName = aws.String(v)
				}
			}
			settings = append(settings, &optionSetting)
		}
	}

	return settings
}

func dropGeneratedSecurityGroup(ctx context.Context, settingValue string, meta interface{}) string {
	conn := meta.(*conns.AWSClient).EC2Conn()

	groups := strings.Split(settingValue, ",")

	// Check to see if groups are ec2-classic or vpc security groups
	ec2Classic := true
	beanstalkSGRegexp := regexp.MustCompile("sg-[0-9a-fA-F]{8}")
	for _, g := range groups {
		if beanstalkSGRegexp.MatchString(g) {
			ec2Classic = false
			break
		}
	}

	var resp *ec2.DescribeSecurityGroupsOutput
	var err error

	if ec2Classic {
		resp, err = conn.DescribeSecurityGroupsWithContext(ctx, &ec2.DescribeSecurityGroupsInput{
			GroupNames: aws.StringSlice(groups),
		})
	} else {
		resp, err = conn.DescribeSecurityGroupsWithContext(ctx, &ec2.DescribeSecurityGroupsInput{
			GroupIds: aws.StringSlice(groups),
		})
	}

	if err != nil {
		log.Printf("[DEBUG] Elastic Beanstalk error describing SecurityGroups: %v", err)
		return settingValue
	}

	log.Printf("[DEBUG] Elastic Beanstalk using ec2-classic security-groups: %t", ec2Classic)
	var legitGroups []string
	for _, group := range resp.SecurityGroups {
		log.Printf("[DEBUG] Elastic Beanstalk SecurityGroup: %v", *group.GroupName)
		if !strings.HasPrefix(*group.GroupName, "awseb") {
			if ec2Classic {
				legitGroups = append(legitGroups, *group.GroupName)
			} else {
				legitGroups = append(legitGroups, *group.GroupId)
			}
		}
	}

	sort.Strings(legitGroups)

	return strings.Join(legitGroups, ",")
}

type beanstalkEnvironmentError struct {
	eventDate     *time.Time
	environmentID string
	message       *string
}

func (e beanstalkEnvironmentError) Error() string {
	return e.eventDate.String() + " (" + e.environmentID + ") : " + *e.message
}

type beanstalkEnvironmentErrors []*beanstalkEnvironmentError

func (e beanstalkEnvironmentErrors) Len() int      { return len(e) }
func (e beanstalkEnvironmentErrors) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
func (e beanstalkEnvironmentErrors) Less(i, j int) bool {
	return e[i].eventDate.Before(*e[j].eventDate)
}

func getEnvironmentErrors(ctx context.Context, conn *elasticbeanstalk.ElasticBeanstalk, environmentId string, t time.Time) (*multierror.Error, error) {
	environmentErrors, err := conn.DescribeEventsWithContext(ctx, &elasticbeanstalk.DescribeEventsInput{
		EnvironmentId: aws.String(environmentId),
		Severity:      aws.String("ERROR"),
		StartTime:     aws.Time(t),
	})

	if err != nil {
		return nil, fmt.Errorf("getting events: %w", err)
	}

	var events beanstalkEnvironmentErrors
	for _, event := range environmentErrors.Events {
		e := &beanstalkEnvironmentError{
			eventDate:     event.EventDate,
			environmentID: environmentId,
			message:       event.Message,
		}
		events = append(events, e)
	}
	sort.Sort(events)

	var result *multierror.Error
	for _, event := range events {
		result = multierror.Append(result, event)
	}

	return result, nil
}
