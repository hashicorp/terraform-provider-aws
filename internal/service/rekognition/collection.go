package rekognition

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/rekognition/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
	"context"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rekognition_collection")
// @Tags(identifierAttribute="arn")
func ResourceCollection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCollectionCreate,
		ReadWithoutTimeout:   resourceCollectionRead,
		UpdateWithoutTimeout: resourceCollectionUpdate,
		DeleteWithoutTimeout: resourceCollectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"collection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"face_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"face_model_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameCollection = "Collection"
)

func resourceCollectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RekognitionClient()

	in := &rekognition.CreateCollectionInput{
		CollectionId: aws.String(d.Get("collection_id").(string)),
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateCollection(ctx, in)
	if err != nil {
		return create.DiagError(names.Rekognition, create.ErrActionCreating, ResNameCollection, d.Get("collection_id").(string), err)
	}

	if out == nil || out.CollectionArn == nil {
		return create.DiagError(names.Rekognition, create.ErrActionCreating, ResNameCollection, d.Get("collection_id").(string), errors.New("empty output"))
	}

	arn := aws.ToString(out.CollectionArn)
	d.SetId(arn[strings.LastIndex(arn, "/")+1:])

	return resourceCollectionRead(ctx, d, meta)
}

func resourceCollectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RekognitionClient()

	out, err := findCollectionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Rekognition Collection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Rekognition, create.ErrActionReading, ResNameCollection, d.Id(), err)
	}

	arn := aws.ToString(out.CollectionARN)
	d.Set("collection_id", arn[strings.LastIndex(arn, "/")+1:])
	d.Set("arn", out.CollectionARN)
	d.Set("face_count", out.FaceCount)
	d.Set("face_model_version", out.FaceModelVersion)

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))
	if err != nil {
		return create.DiagError(names.Rekognition, create.ErrActionReading, ResNameCollection, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.Rekognition, create.ErrActionSetting, ResNameCollection, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.Rekognition, create.ErrActionSetting, ResNameCollection, d.Id(), err)
	}

	return nil
}

func resourceCollectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.HasChangesExcept("tags", "tags_all") {
		return resourceStreamProcessorRead(ctx, d, meta)
	} else {
		return nil
	}
}

func resourceCollectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RekognitionClient()

	log.Printf("[INFO] Deleting Rekognition Collection %s", d.Id())

	_, err := conn.DeleteCollection(ctx, &rekognition.DeleteCollectionInput{
		CollectionId: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.Rekognition, create.ErrActionDeleting, ResNameCollection, d.Id(), err)
	}

	return nil
}

func findCollectionByID(ctx context.Context, conn *rekognition.Client, id string) (*rekognition.DescribeCollectionOutput, error) {
	in := &rekognition.DescribeCollectionInput{
		CollectionId: aws.String(id),
	}
	out, err := conn.DescribeCollection(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}
		return nil, err
	}
	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
