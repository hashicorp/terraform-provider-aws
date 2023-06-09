package pipes

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pipes"
	"github.com/aws/aws-sdk-go-v2/service/pipes/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_pipes_pipe", name="Pipe")
// @Tags(identifierAttribute="arn")
func ResourcePipe() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePipeCreate,
		ReadWithoutTimeout:   resourcePipeRead,
		UpdateWithoutTimeout: resourcePipeUpdate,
		DeleteWithoutTimeout: resourcePipeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},
			"desired_state": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          string(types.RequestedPipeStateRunning),
				ValidateDiagFunc: enum.Validate[types.RequestedPipeState](),
			},
			"enrichment": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"enrichment_parameters": enrichmentParametersSchema(),
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[\.\-_A-Za-z0-9]+`), ""),
				),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64-id.UniqueIDSuffixLength),
					validation.StringMatch(regexp.MustCompile(`^[\.\-_A-Za-z0-9]+`), ""),
				),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"source": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					verify.ValidARN,
					validation.StringMatch(regexp.MustCompile(`^smk://(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9]):[0-9]{1,5}|arn:(aws[a-zA-Z0-9-]*):([a-zA-Z0-9\-]+):([a-z]{2}((-gov)|(-iso(b?)))?-[a-z]+-\d{1})?:(\d{12})?:(.+)$`), ""),
				),
			},
			"source_parameters": sourceParametersSchema(),
			"target": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"target_parameters": target_parameters_schema,
			names.AttrTags:      tftags.TagsSchema(),
			names.AttrTagsAll:   tftags.TagsSchemaComputed(),
		},
	}
}

const (
	ResNamePipe = "Pipe"
)

func resourcePipeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).PipesClient()

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &pipes.CreatePipeInput{
		DesiredState: types.RequestedPipeState(d.Get("desired_state").(string)),
		Name:         aws.String(name),
		RoleArn:      aws.String(d.Get("role_arn").(string)),
		Source:       aws.String(d.Get("source").(string)),
		Tags:         GetTagsIn(ctx),
		Target:       aws.String(d.Get("target").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("enrichment"); ok && v != "" {
		input.Enrichment = aws.String(v.(string))
	}

	if v, ok := d.GetOk("enrichment_parameters"); ok {
		input.EnrichmentParameters = expandEnrichmentParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("source_parameters"); ok {
		input.SourceParameters = expandSourceParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("target_parameters"); ok {
		input.TargetParameters = expandTargetParameters(v.([]interface{}))
	}

	output, err := conn.CreatePipe(ctx, input)

	if err != nil {
		return create.DiagError(names.Pipes, create.ErrActionCreating, ResNamePipe, name, err)
	}

	d.SetId(aws.ToString(output.Name))

	if _, err := waitPipeCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.Pipes, create.ErrActionWaitingForCreation, ResNamePipe, d.Id(), err)
	}

	return resourcePipeRead(ctx, d, meta)
}

func resourcePipeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).PipesClient()

	output, err := FindPipeByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Pipes Pipe (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Pipes, create.ErrActionReading, ResNamePipe, d.Id(), err)
	}

	d.Set("arn", output.Arn)
	d.Set("description", output.Description)
	d.Set("desired_state", output.DesiredState)
	d.Set("enrichment", output.Enrichment)
	if err := d.Set("enrichment_parameters", flattenEnrichmentParameters(output.EnrichmentParameters)); err != nil {
		return diag.Errorf("setting enrichment_parameters: %s", err)
	}
	d.Set("name", output.Name)
	d.Set("name_prefix", create.NamePrefixFromName(aws.ToString(output.Name)))
	d.Set("role_arn", output.RoleArn)
	d.Set("source", output.Source)
	if err := d.Set("source_parameters", flattenSourceParameters(output.SourceParameters)); err != nil {
		return diag.Errorf("setting source_parameters: %s", err)
	}
	d.Set("target", output.Target)
	if err := d.Set("target_parameters", flattenTargetParameters(output.TargetParameters)); err != nil {
		return diag.Errorf("setting target_parameters: %s", err)
	}

	return nil
}

func resourcePipeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).PipesClient()

	if d.HasChangesExcept("tags", "tags_all") {
		input := &pipes.UpdatePipeInput{
			Description:  aws.String(d.Get("description").(string)),
			DesiredState: types.RequestedPipeState(d.Get("desired_state").(string)),
			Name:         aws.String(d.Id()),
			RoleArn:      aws.String(d.Get("role_arn").(string)),
			// Reset state in case it's a deletion.
			SourceParameters: &types.UpdatePipeSourceParameters{
				FilterCriteria: &types.FilterCriteria{},
			},
			Target: aws.String(d.Get("target").(string)),
			// Reset state in case it's a deletion, have to set the input to an empty string otherwise it doesn't get overwritten.
			TargetParameters: &types.PipeTargetParameters{
				InputTemplate: aws.String(""),
			},
		}

		if d.HasChange("enrichment") {
			if v, ok := d.GetOk("enrichment"); ok && v.(string) != "" {
				input.Enrichment = aws.String(v.(string))
			}
		}

		if v, ok := d.GetOk("enrichment_parameters"); ok {
			input.EnrichmentParameters = expandEnrichmentParameters(v.([]interface{}))
		}

		if v, ok := d.GetOk("source_parameters"); ok {
			input.SourceParameters = expandSourceUpdateParameters(v.([]interface{}))
		}

		if v, ok := d.GetOk("target_parameters"); ok {
			input.TargetParameters = expandTargetParameters(v.([]interface{}))
		}

		output, err := conn.UpdatePipe(ctx, input)

		if err != nil {
			return create.DiagError(names.Pipes, create.ErrActionUpdating, ResNamePipe, d.Id(), err)
		}

		if _, err := waitPipeUpdated(ctx, conn, aws.ToString(output.Name), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.DiagError(names.Pipes, create.ErrActionWaitingForUpdate, ResNamePipe, d.Id(), err)
		}
	}

	return resourcePipeRead(ctx, d, meta)
}

func resourcePipeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).PipesClient()

	log.Printf("[INFO] Deleting EventBridge Pipes Pipe: %s", d.Id())
	_, err := conn.DeletePipe(ctx, &pipes.DeletePipeInput{
		Name: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.Pipes, create.ErrActionDeleting, ResNamePipe, d.Id(), err)
	}

	if _, err := waitPipeDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.Pipes, create.ErrActionWaitingForDeletion, ResNamePipe, d.Id(), err)
	}

	return nil
}
