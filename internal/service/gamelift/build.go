package gamelift

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceBuild() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBuildCreate,
		ReadWithoutTimeout:   resourceBuildRead,
		UpdateWithoutTimeout: resourceBuildUpdate,
		DeleteWithoutTimeout: resourceBuildDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"operating_system": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(gamelift.OperatingSystem_Values(), false),
			},
			"storage_location": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"object_version": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"version": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceBuildCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := gamelift.CreateBuildInput{
		Name:            aws.String(d.Get("name").(string)),
		OperatingSystem: aws.String(d.Get("operating_system").(string)),
		StorageLocation: expandStorageLocation(d.Get("storage_location").([]interface{})),
		Tags:            Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("version"); ok {
		input.Version = aws.String(v.(string))
	}

	log.Printf("[INFO] Creating GameLift Build: %s", input)
	var out *gamelift.CreateBuildOutput
	err := resource.RetryContext(ctx, propagationTimeout, func() *resource.RetryError {
		var err error
		out, err = conn.CreateBuildWithContext(ctx, &input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, gamelift.ErrCodeInvalidRequestException, "Provided build is not accessible.") ||
				tfawserr.ErrMessageContains(err, gamelift.ErrCodeInvalidRequestException, "GameLift cannot assume the role") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		out, err = conn.CreateBuildWithContext(ctx, &input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GameLift build client: %s", err)
	}

	d.SetId(aws.StringValue(out.Build.BuildId))

	if _, err := waitBuildReady(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GameLift Build (%s) to ready: %s", d.Id(), err)
	}

	return append(diags, resourceBuildRead(ctx, d, meta)...)
}

func resourceBuildRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	log.Printf("[INFO] Reading GameLift Build: %s", d.Id())
	build, err := FindBuildByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GameLift Build (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GameLift Build (%s): %s", d.Id(), err)
	}

	d.Set("name", build.Name)
	d.Set("operating_system", build.OperatingSystem)
	d.Set("version", build.Version)

	arn := aws.StringValue(build.BuildArn)
	d.Set("arn", arn)
	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Game Lift Build (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceBuildUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftConn()

	if d.HasChangesExcept("tags", "tags_all") {
		log.Printf("[INFO] Updating GameLift Build: %s", d.Id())
		input := gamelift.UpdateBuildInput{
			BuildId: aws.String(d.Id()),
			Name:    aws.String(d.Get("name").(string)),
		}
		if v, ok := d.GetOk("version"); ok {
			input.Version = aws.String(v.(string))
		}

		_, err := conn.UpdateBuildWithContext(ctx, &input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GameLift build client: %s", err)
		}
	}

	if d.HasChange("tags_all") {
		arn := d.Get("arn").(string)
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, arn, o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Game Lift Build (%s) tags: %s", arn, err)
		}
	}

	return append(diags, resourceBuildRead(ctx, d, meta)...)
}

func resourceBuildDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftConn()

	log.Printf("[INFO] Deleting GameLift Build: %s", d.Id())
	_, err := conn.DeleteBuildWithContext(ctx, &gamelift.DeleteBuildInput{
		BuildId: aws.String(d.Id()),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GameLift Build Client (%s): %s", d.Id(), err)
	}
	return diags
}

func expandStorageLocation(cfg []interface{}) *gamelift.S3Location {
	loc := cfg[0].(map[string]interface{})

	location := &gamelift.S3Location{
		Bucket:  aws.String(loc["bucket"].(string)),
		Key:     aws.String(loc["key"].(string)),
		RoleArn: aws.String(loc["role_arn"].(string)),
	}

	if v, ok := loc["object_version"].(string); ok && v != "" {
		location.ObjectVersion = aws.String(v)
	}

	return location
}
