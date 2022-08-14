package inspector2

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/inspector2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceFilter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFilterCreate,
		ReadWithoutTimeout:   resourceFilterRead,
		UpdateWithoutTimeout: resourceFilterUpdate,
		DeleteWithoutTimeout: resourceFilterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"action": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"NONE", "SUPPRESS"}, false),
			},
			"filter_criteria": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"aws_account_id":            stringFilter(),
						"component_id":              stringFilter(),
						"component_type":            stringFilter(),
						"ec2_instance_image_id":     stringFilter(),
						"ec2_instance_subnet_id":    stringFilter(),
						"ec2_instance_vpc_id":       stringFilter(),
						"ecr_image_architecture":    stringFilter(),
						"ecr_image_hash":            stringFilter(),
						"ecr_image_pushed_at":       dateFilter(),
						"ecr_image_registry":        stringFilter(),
						"ecr_image_repository_name": stringFilter(),
						"ecr_image_tags":            stringFilter(),
						"finding_arn":               stringFilter(),
						"finding_status":            stringFilterWithValues([]string{"ACTIVE", "SUPPRESSED", "CLOSED"}),
						"finding_type":              stringFilter(),
						"first_observed_at":         dateFilter(),
						"inspector_score":           numberFilter(),
						"last_observed_at":          dateFilter(),
						"network_protocol":          stringFilter(),
						"port_range":                portRangeFilter(),
						"related_vulnerabilities":   stringFilter(),
						"resource_id":               stringFilter(),
						"resource_tags":             mapFilter(),
						"resource_type":             stringFilterWithValues([]string{"AWS_EC2_INSTANCE", "AWS_ECR_CONTAINER_IMAGE"}),
						"severity":                  stringFilterWithValues([]string{"INFORMATIONAL", "LOW", "MEDIUM", "HIGH", "CRITICAL", "UNTRIAGED"}),
						"title":                     stringFilter(),
						"updated_at":                dateFilter(),
						"vendor_severity":           stringFilter(),
						"vulnerability_id":          stringFilter(),
						"vulnerability_source":      stringFilter(),
						"vulnerable_packages":       packageFilter(),
					},
				},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name_prefix"},
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ValidateFunc:  validation.NoZeroValues,
				ConflictsWith: []string{"name"},
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"reason": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameFilter = "Filter"
)

func resourceFilterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Inspector2Conn

	in := &inspector2.CreateFilterInput{
		Name:           aws.String(create.Name(d.Get("name").(string), d.Get("name_prefix").(string))),
		Action:         aws.String(d.Get("action").(string)),
		FilterCriteria: expandFilterCriteria(d.Get("filter_criteria").([]interface{})),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("reason"); ok {
		in.Reason = aws.String(v.(string))
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateFilterWithContext(ctx, in)
	if err != nil {
		return create.DiagError(names.Inspector2, create.ErrActionCreating, ResNameFilter, d.Get("name").(string), err)
	}

	if out == nil || out.Arn == nil {
		return create.DiagError(names.Inspector2, create.ErrActionCreating, ResNameFilter, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.StringValue(out.Arn))

	return resourceFilterRead(ctx, d, meta)
}

func resourceFilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Inspector2Conn

	out, err := findFilterByArn(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Inspector2 Filter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Inspector2, create.ErrActionReading, ResNameFilter, d.Id(), err)
	}

	d.Set("action", out.Action)
	d.Set("arn", out.Arn)
	d.Set("description", out.Description)
	d.Set("name", out.Name)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(out.Name)))

	if err := d.Set("filter_criteria", flattenFilterCriteria(out.Criteria)); err != nil {
		return create.DiagError(names.Inspector2, create.ErrActionSetting, ResNameFilter, d.Id(), err)
	}

	d.Set("owner_id", out.OwnerId)
	d.Set("reason", out.Reason)

	tags, err := ListTagsWithContext(ctx, conn, d.Id())
	if err != nil {
		return create.DiagError(names.Inspector2, create.ErrActionReading, ResNameFilter, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.Inspector2, create.ErrActionSetting, ResNameFilter, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.Inspector2, create.ErrActionSetting, ResNameFilter, d.Id(), err)
	}

	return nil
}

func resourceFilterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Inspector2Conn

	update := false

	in := &inspector2.UpdateFilterInput{
		FilterArn: aws.String(d.Id()),
	}

	if d.HasChanges("name", "name_prefix") {
		in.Name = aws.String(create.Name(d.Get("name").(string), d.Get("name_prefix").(string)))
		update = true
	}

	if d.HasChange("action") {
		in.Action = aws.String(d.Get("action").(string))
		update = true
	}

	if d.HasChange("filter_criteria") {
		in.FilterCriteria = expandFilterCriteria(d.Get("filter_criteria").([]interface{}))
		update = true
	}

	if d.HasChange("description") {
		in.Description = aws.String(d.Get("description").(string))
		update = true
	}

	if d.HasChange("reason") {
		in.Reason = aws.String(d.Get("reason").(string))
		update = true
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTagsWithContext(ctx, conn, d.Id(), o, n); err != nil {
			return create.DiagError(names.Inspector2, create.ErrActionUpdating, ResNameFilter, d.Id(), err)
		}
	}

	if update {
		log.Printf("[DEBUG] Updating Inspector2 Filter (%s): %#v", d.Id(), in)
		_, err := conn.UpdateFilterWithContext(ctx, in)
		if err != nil {
			return create.DiagError(names.Inspector2, create.ErrActionUpdating, ResNameFilter, d.Id(), err)
		}
	}

	return resourceFilterRead(ctx, d, meta)
}

func resourceFilterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Inspector2Conn

	log.Printf("[INFO] Deleting Inspector2 Filter %s", d.Id())

	_, err := conn.DeleteFilterWithContext(ctx, &inspector2.DeleteFilterInput{
		Arn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, inspector2.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.Inspector2, create.ErrActionDeleting, ResNameFilter, d.Id(), err)
	}

	return nil
}

func findFilterByArn(ctx context.Context, conn *inspector2.Inspector2, id string) (*inspector2.Filter, error) {
	in := &inspector2.ListFiltersInput{
		Arns: []*string{aws.String(id)},
	}
	out, err := conn.ListFiltersWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, inspector2.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.Filters) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Filters[0], nil
}
