package mq

import (
	"context"
	"encoding/base64"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfigurationCreate,
		ReadWithoutTimeout:   resourceConfigurationRead,
		UpdateWithoutTimeout: resourceConfigurationUpdate,
		DeleteWithoutTimeout: schema.NoopContext, // Delete is not available in the API

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.Sequence(
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				if diff.HasChange("description") {
					return diff.SetNewComputed("latest_revision")
				}
				if diff.HasChange("data") {
					o, n := diff.GetChange("data")
					os := o.(string)
					ns := n.(string)
					if !suppressXMLEquivalentConfig("data", os, ns, nil) {
						return diff.SetNewComputed("latest_revision")
					}
				}
				return nil
			},
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_strategy": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(mq.AuthenticationStrategy_Values(), true),
			},
			"data": {
				Type:                  schema.TypeString,
				Required:              true,
				DiffSuppressFunc:      suppressXMLEquivalentConfig,
				DiffSuppressOnRefresh: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"engine_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(mq.EngineType_Values(), true),
			},
			"engine_version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"latest_revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MQConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &mq.CreateConfigurationRequest{
		EngineType:    aws.String(d.Get("engine_type").(string)),
		EngineVersion: aws.String(d.Get("engine_version").(string)),
		Name:          aws.String(name),
	}

	if v, ok := d.GetOk("authentication_strategy"); ok {
		input.AuthenticationStrategy = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreateConfigurationWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating MQ Configuration (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Id))

	if v, ok := d.GetOk("data"); ok {
		input := &mq.UpdateConfigurationRequest{
			ConfigurationId: aws.String(d.Id()),
			Data:            aws.String(base64.StdEncoding.EncodeToString([]byte(v.(string)))),
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		_, err := conn.UpdateConfigurationWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating MQ Configuration (%s): %s", d.Id(), err)
		}
	}

	return resourceConfigurationRead(ctx, d, meta)
}

func resourceConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MQConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	configuration, err := FindConfigurationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MQ Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading MQ Configuration (%s): %s", d.Id(), err)
	}

	d.Set("arn", configuration.Arn)
	d.Set("authentication_strategy", configuration.AuthenticationStrategy)
	d.Set("description", configuration.LatestRevision.Description)
	d.Set("engine_type", configuration.EngineType)
	d.Set("engine_version", configuration.EngineVersion)
	d.Set("latest_revision", configuration.LatestRevision.Revision)
	d.Set("name", configuration.Name)

	revision := strconv.FormatInt(aws.Int64Value(configuration.LatestRevision.Revision), 10)
	configurationRevision, err := conn.DescribeConfigurationRevisionWithContext(ctx, &mq.DescribeConfigurationRevisionInput{
		ConfigurationId:       aws.String(d.Id()),
		ConfigurationRevision: aws.String(revision),
	})

	if err != nil {
		return diag.Errorf("reading MQ Configuration (%s) revision (%s): %s", d.Id(), revision, err)
	}

	data, err := base64.StdEncoding.DecodeString(aws.StringValue(configurationRevision.Data))

	if err != nil {
		return diag.Errorf("base64 decoding: %s", err)
	}

	d.Set("data", string(data))

	tags := KeyValueTags(configuration.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MQConn()

	if d.HasChanges("data", "description") {
		input := &mq.UpdateConfigurationRequest{
			ConfigurationId: aws.String(d.Id()),
			Data:            aws.String(base64.StdEncoding.EncodeToString([]byte(d.Get("data").(string)))),
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		_, err := conn.UpdateConfigurationWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating MQ Configuration (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("updating MQ Configuration (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceConfigurationRead(ctx, d, meta)
}

func FindConfigurationByID(ctx context.Context, conn *mq.MQ, id string) (*mq.DescribeConfigurationOutput, error) {
	input := &mq.DescribeConfigurationInput{
		ConfigurationId: aws.String(id),
	}

	output, err := conn.DescribeConfigurationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, mq.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func suppressXMLEquivalentConfig(k, old, new string, d *schema.ResourceData) bool {
	os, err := CanonicalXML(old)
	if err != nil {
		log.Printf("[ERR] Error getting cannonicalXML from state (%s): %s", k, err)
		return false
	}
	ns, err := CanonicalXML(new)
	if err != nil {
		log.Printf("[ERR] Error getting cannonicalXML from config (%s): %s", k, err)
		return false
	}

	return os == ns
}
