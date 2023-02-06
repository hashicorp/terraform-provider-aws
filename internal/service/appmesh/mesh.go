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
)

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
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
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
										Type:     schema.TypeString,
										Optional: true,
										Default:  appmesh.EgressFilterTypeDropAll,
										ValidateFunc: validation.StringInSlice([]string{
											appmesh.EgressFilterTypeAllowAll,
											appmesh.EgressFilterTypeDropAll,
										}, false),
									},
								},
							},
						},
					},
				},
			},

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

			"resource_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tftags.TagsSchema(),

			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceMeshCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	meshName := d.Get("name").(string)
	req := &appmesh.CreateMeshInput{
		MeshName: aws.String(meshName),
		Spec:     expandMeshSpec(d.Get("spec").([]interface{})),
		Tags:     Tags(tags.IgnoreAWS()),
	}

	log.Printf("[DEBUG] Creating App Mesh service mesh: %#v", req)
	_, err := conn.CreateMeshWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating App Mesh service mesh: %s", err)
	}

	d.SetId(meshName)

	return append(diags, resourceMeshRead(ctx, d, meta)...)
}

func resourceMeshRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	req := &appmesh.DescribeMeshInput{
		MeshName: aws.String(d.Id()),
	}
	if v, ok := d.GetOk("mesh_owner"); ok {
		req.MeshOwner = aws.String(v.(string))
	}

	var resp *appmesh.DescribeMeshOutput

	err := resource.RetryContext(ctx, propagationTimeout, func() *resource.RetryError {
		var err error

		resp, err = conn.DescribeMeshWithContext(ctx, req)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		resp, err = conn.DescribeMeshWithContext(ctx, req)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
		log.Printf("[WARN] App Mesh Service Mesh (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Service Mesh: %s", err)
	}

	if resp == nil || resp.Mesh == nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Service Mesh: empty response")
	}

	if aws.StringValue(resp.Mesh.Status.Status) == appmesh.MeshStatusCodeDeleted {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading App Mesh Service Mesh: %s after creation", aws.StringValue(resp.Mesh.Status.Status))
		}

		log.Printf("[WARN] App Mesh Service Mesh (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	arn := aws.StringValue(resp.Mesh.Metadata.Arn)
	d.Set("name", resp.Mesh.MeshName)
	d.Set("arn", arn)
	d.Set("created_date", resp.Mesh.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", resp.Mesh.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("mesh_owner", resp.Mesh.Metadata.MeshOwner)
	d.Set("resource_owner", resp.Mesh.Metadata.ResourceOwner)
	err = d.Set("spec", flattenMeshSpec(resp.Mesh.Spec))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting spec: %s", err)
	}

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for App Mesh service mesh (%s): %s", arn, err)
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
		_, v := d.GetChange("spec")
		req := &appmesh.UpdateMeshInput{
			MeshName: aws.String(d.Id()),
			Spec:     expandMeshSpec(v.([]interface{})),
		}

		log.Printf("[DEBUG] Updating App Mesh service mesh: %#v", req)
		_, err := conn.UpdateMeshWithContext(ctx, req)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating App Mesh service mesh: %s", err)
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, arn, o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating App Mesh service mesh (%s) tags: %s", arn, err)
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
		return sdkdiag.AppendErrorf(diags, "deleting App Mesh service mesh: %s", err)
	}

	return diags
}
