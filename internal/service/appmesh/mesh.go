package appmesh

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appmesh_mesh")
func ResourceMesh() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMeshCreate,
		ReadWithoutTimeout:   resourceMeshRead,
		UpdateWithoutTimeout: resourceMeshUpdate,
		DeleteWithoutTimeout: resourceMeshDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mesh_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"resource_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"spec": {
				Type:             schema.TypeList,
				Optional:         true,
				MinItems:         0,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"egress_filter": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      appmesh.EgressFilterTypeDropAll,
										ValidateFunc: validation.StringInSlice(appmesh.EgressFilterType_Values(), false),
									},
								},
							},
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceMeshCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &appmesh.CreateMeshInput{
		MeshName: aws.String(name),
		Spec:     expandMeshSpec(d.Get("spec").([]interface{})),
		Tags:     Tags(tags.IgnoreAWS()),
	}

	_, err := conn.CreateMeshWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating App Mesh Service Mesh (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceMeshRead(ctx, d, meta)...)
}

func resourceMeshRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindMeshByTwoPartKey(ctx, conn, d.Id(), d.Get("mesh_owner").(string))
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] App Mesh Service Mesh (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Service Mesh (%s): %s", d.Id(), err)
	}

	mesh := outputRaw.(*appmesh.MeshData)
	arn := aws.StringValue(mesh.Metadata.Arn)
	d.Set("arn", arn)
	d.Set("created_date", mesh.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", mesh.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("mesh_owner", mesh.Metadata.MeshOwner)
	d.Set("name", mesh.MeshName)
	d.Set("resource_owner", mesh.Metadata.ResourceOwner)
	if err := d.Set("spec", flattenMeshSpec(mesh.Spec)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting spec: %s", err)
	}

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for App Mesh Service Mesh (%s): %s", arn, err)
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

func resourceMeshUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()

	if d.HasChange("spec") {
		input := &appmesh.UpdateMeshInput{
			MeshName: aws.String(d.Id()),
			Spec:     expandMeshSpec(d.Get("spec").([]interface{})),
		}

		_, err := conn.UpdateMeshWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating App Mesh Service Mesh (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		arn := d.Get("arn").(string)
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, arn, o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating App Mesh Service Mesh (%s) tags: %s", arn, err)
		}
	}

	return append(diags, resourceMeshRead(ctx, d, meta)...)
}

func resourceMeshDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()

	log.Printf("[DEBUG] Deleting App Mesh Service Mesh: %s", d.Id())
	_, err := conn.DeleteMeshWithContext(ctx, &appmesh.DeleteMeshInput{
		MeshName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting App Mesh Service Mesh (%s): %s", d.Id(), err)
	}

	return diags
}

func FindMeshByTwoPartKey(ctx context.Context, conn *appmesh.AppMesh, name, owner string) (*appmesh.MeshData, error) {
	input := &appmesh.DescribeMeshInput{
		MeshName: aws.String(name),
	}
	if owner != "" {
		input.MeshOwner = aws.String(owner)
	}

	output, err := findMesh(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := aws.StringValue(output.Status.Status); status == appmesh.MeshStatusCodeDeleted {
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}

func findMesh(ctx context.Context, conn *appmesh.AppMesh, input *appmesh.DescribeMeshInput) (*appmesh.MeshData, error) {
	output, err := conn.DescribeMeshWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Mesh == nil || output.Mesh.Metadata == nil || output.Mesh.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Mesh, nil
}
