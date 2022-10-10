package codepipeline

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/codepipeline"
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

func ResourceCustomActionType() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCustomActionTypeCreate,
		ReadWithoutTimeout:   resourceCustomActionTypeRead,
		UpdateWithoutTimeout: resourceCustomActionTypeUpdate,
		DeleteWithoutTimeout: resourceCustomActionTypeDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"category": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringInSlice(codepipeline.ActionCategory_Values(), false),
			},
			"configuration_property": {
				Type:     schema.TypeList,
				MaxItems: 10,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"key": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"queryable": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"required": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"secret": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(codepipeline.ActionConfigurationPropertyType_Values(), false),
						},
					},
				},
			},
			"input_artifact_details": {
				Type:     schema.TypeList,
				ForceNew: true,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"maximum_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 5),
						},
						"minimum_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 5),
						},
					},
				},
			},
			"output_artifact_details": {
				Type:     schema.TypeList,
				ForceNew: true,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"maximum_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 5),
						},
						"minimum_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 5),
						},
					},
				},
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provider": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 25),
			},
			"settings": {
				Type:     schema.TypeList,
				ForceNew: true,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"entity_url_template": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"execution_url_template": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"revision_url_template": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"third_party_configuration_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"version": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 9),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCustomActionTypeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CodePipelineConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	category := d.Get("category").(string)
	provider := d.Get("provider").(string)
	version := d.Get("version").(string)
	id := CustomActionTypeCreateResourceID(category, provider, version)
	input := &codepipeline.CreateCustomActionTypeInput{
		Category: aws.String(category),
		Provider: aws.String(provider),
		Version:  aws.String(version),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	_, err := conn.CreateCustomActionTypeWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating CodePipeline Custom Action Type (%s): %s", id, err)
	}

	d.SetId(id)

	return resourceCustomActionTypeRead(ctx, d, meta)
}

func resourceCustomActionTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CodePipelineConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	category, provider, version, err := CustomActionTypeParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	actionType, err := FindCustomActionTypeByThreePartKey(ctx, conn, category, provider, version)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodePipeline Custom Action Type %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading CodePipeline Custom Action Type (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   codepipeline.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("actiontype:%s:%s:%s:%s", codepipeline.ActionOwnerCustom, category, provider, version),
	}.String()
	d.Set("arn", arn)
	d.Set("category", actionType.Id.Category)
	d.Set("owner", actionType.Id.Owner)
	d.Set("provider_name", actionType.Id.Provider)
	d.Set("version", actionType.Id.Version)

	tags, err := ListTagsWithContext(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("listing tags for CodePipeline Custom Action Type (%s): %s", arn, err)
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

func resourceCustomActionTypeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CodePipelineConn

	if d.HasChangesExcept("tags", "tags_all") {
		category, provider, version, err := CustomActionTypeParseResourceID(d.Id())

		if err != nil {
			return diag.FromErr(err)
		}

		input := &codepipeline.UpdateActionTypeInput{
			ActionType: &codepipeline.ActionTypeDeclaration{
				Id: &codepipeline.ActionTypeIdentifier{
					Category: aws.String(category),
					Owner:    aws.String(codepipeline.ActionOwnerCustom),
					Provider: aws.String(provider),
					Version:  aws.String(version),
				},
			},
		}

		_, err = conn.UpdateActionTypeWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating CodePipeline Custom Action Type (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		arn := d.Get("arn").(string)

		if err := UpdateTagsWithContext(ctx, conn, arn, o, n); err != nil {
			return diag.Errorf("updating CodePipeline Custom Action Type (%s) tags: %s", arn, err)
		}
	}

	return resourceCustomActionTypeRead(ctx, d, meta)
}

func resourceCustomActionTypeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CodePipelineConn

	category, provider, version, err := CustomActionTypeParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Deleting CodePipeline Custom Action Type: %s", d.Id())
	_, err = conn.DeleteCustomActionTypeWithContext(ctx, &codepipeline.DeleteCustomActionTypeInput{
		Category: aws.String(category),
		Provider: aws.String(provider),
		Version:  aws.String(version),
	})

	if tfawserr.ErrCodeEquals(err, codepipeline.ErrCodeActionTypeNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting CodePipeline (%s): %s", d.Id(), err)
	}

	return nil
}

const customActionTypeResourceIDSeparator = ":"

func CustomActionTypeCreateResourceID(category, provider, version string) string {
	parts := []string{category, provider, version}
	id := strings.Join(parts, customActionTypeResourceIDSeparator)

	return id
}

func CustomActionTypeParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, customActionTypeResourceIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected category%[2]sprovider%[2]sversion", id, customActionTypeResourceIDSeparator)
}

func FindCustomActionTypeByThreePartKey(ctx context.Context, conn *codepipeline.CodePipeline, category, provider, version string) (*codepipeline.ActionTypeDeclaration, error) {
	input := &codepipeline.GetActionTypeInput{
		Category: aws.String(category),
		Owner:    aws.String(codepipeline.ActionOwnerCustom),
		Provider: aws.String(provider),
		Version:  aws.String(version),
	}

	output, err := conn.GetActionTypeWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, codepipeline.ErrCodeActionTypeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ActionType == nil || output.ActionType.Id == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ActionType, nil
}
