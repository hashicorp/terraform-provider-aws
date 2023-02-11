package swf

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/swf"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDomain() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainCreate,
		ReadWithoutTimeout:   resourceDomainRead,
		UpdateWithoutTimeout: resourceDomainUpdate,
		DeleteWithoutTimeout: resourceDomainDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"workflow_execution_retention_period_in_days": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
					value, err := strconv.Atoi(v.(string))
					if err != nil || value > 90 || value < 0 {
						es = append(es, fmt.Errorf(
							"%q must be between 0 and 90 days inclusive", k))
					}
					return
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SWFConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &swf.RegisterDomainInput{
		Name:                                   aws.String(name),
		Tags:                                   Tags(tags.IgnoreAWS()),
		WorkflowExecutionRetentionPeriodInDays: aws.String(d.Get("workflow_execution_retention_period_in_days").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := conn.RegisterDomainWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating SWF Domain (%s): %s", name, err)
	}

	d.SetId(name)

	return resourceDomainRead(ctx, d, meta)
}

func resourceDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SWFConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindDomainByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SWF Domain (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading SWF Domain (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(output.DomainInfo.Arn)
	d.Set("arn", arn)
	d.Set("description", output.DomainInfo.Description)
	d.Set("name", output.DomainInfo.Name)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(output.DomainInfo.Name)))
	d.Set("workflow_execution_retention_period_in_days", output.Configuration.WorkflowExecutionRetentionPeriodInDays)

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("listing tags for SWF Domain (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SWFConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("updating SWF Domain (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceDomainRead(ctx, d, meta)
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SWFConn()

	_, err := conn.DeprecateDomainWithContext(ctx, &swf.DeprecateDomainInput{
		Name: aws.String(d.Get("name").(string)),
	})

	if tfawserr.ErrCodeEquals(err, swf.ErrCodeDomainDeprecatedFault, swf.ErrCodeUnknownResourceFault) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting SWF Domain (%s): %s", d.Id(), err)
	}

	return nil
}

func FindDomainByName(ctx context.Context, conn *swf.SWF, name string) (*swf.DescribeDomainOutput, error) {
	input := &swf.DescribeDomainInput{
		Name: aws.String(name),
	}

	output, err := conn.DescribeDomainWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, swf.ErrCodeUnknownResourceFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Configuration == nil || output.DomainInfo == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if status := aws.StringValue(output.DomainInfo.Status); status == swf.RegistrationStatusDeprecated {
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}
