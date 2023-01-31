package location

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceRouteCalculator() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRouteCalculatorCreate,
		ReadWithoutTimeout:   resourceRouteCalculatorRead,
		UpdateWithoutTimeout: resourceRouteCalculatorUpdate,
		DeleteWithoutTimeout: resourceRouteCalculatorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"calculator_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"calculator_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_source": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1000),
			},
			"update_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRouteCalculatorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LocationConn()

	in := &locationservice.CreateRouteCalculatorInput{
		CalculatorName: aws.String(d.Get("calculator_name").(string)),
		DataSource:     aws.String(d.Get("data_source").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateRouteCalculatorWithContext(ctx, in)
	if err != nil {
		return diag.Errorf("creating Location Service Route Calculator (%s): %s", d.Get("calculator_name").(string), err)
	}

	if out == nil {
		return diag.Errorf("creating Location Service Route Calculator (%s): empty output", d.Get("calculator_name").(string))
	}

	d.SetId(aws.StringValue(out.CalculatorName))

	return resourceRouteCalculatorRead(ctx, d, meta)
}

func resourceRouteCalculatorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LocationConn()

	out, err := findRouteCalculatorByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Location Service Route Calculator (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Location Service Route Calculator (%s): %s", d.Id(), err)
	}

	d.Set("calculator_arn", out.CalculatorArn)
	d.Set("calculator_name", out.CalculatorName)
	d.Set("create_time", aws.TimeValue(out.CreateTime).Format(time.RFC3339))
	d.Set("data_source", out.DataSource)
	d.Set("description", out.Description)
	d.Set("update_time", aws.TimeValue(out.UpdateTime).Format(time.RFC3339))

	tags, err := ListTags(ctx, conn, d.Get("calculator_arn").(string))
	if err != nil {
		return diag.Errorf("listing tags for Location Service Route Calculator (%s): %s", d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceRouteCalculatorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LocationConn()

	update := false

	in := &locationservice.UpdateRouteCalculatorInput{
		CalculatorName: aws.String(d.Get("calculator_name").(string)),
	}

	if d.HasChange("description") {
		in.Description = aws.String(d.Get("description").(string))
		update = true
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("calculator_arn").(string), o, n); err != nil {
			return diag.Errorf("updating tags for Location Service Route Calculator (%s): %s", d.Id(), err)
		}
	}

	if !update {
		return nil
	}

	log.Printf("[DEBUG] Updating Location Service Route Calculator (%s): %#v", d.Id(), in)
	_, err := conn.UpdateRouteCalculatorWithContext(ctx, in)
	if err != nil {
		return diag.Errorf("updating Location Service Route Calculator (%s): %s", d.Id(), err)
	}

	return resourceRouteCalculatorRead(ctx, d, meta)
}

func resourceRouteCalculatorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LocationConn()

	log.Printf("[INFO] Deleting Location Service Route Calculator %s", d.Id())

	_, err := conn.DeleteRouteCalculatorWithContext(ctx, &locationservice.DeleteRouteCalculatorInput{
		CalculatorName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Location Service Route Calculator (%s): %s", d.Id(), err)
	}

	return nil
}

func findRouteCalculatorByName(ctx context.Context, conn *locationservice.LocationService, name string) (*locationservice.DescribeRouteCalculatorOutput, error) {
	in := &locationservice.DescribeRouteCalculatorInput{
		CalculatorName: aws.String(name),
	}

	out, err := conn.DescribeRouteCalculatorWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
