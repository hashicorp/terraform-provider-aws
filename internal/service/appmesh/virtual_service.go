package appmesh

import (
	"context"
	"fmt"
	"log"
	"strings"
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

func ResourceVirtualService() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVirtualServiceCreate,
		ReadWithoutTimeout:   resourceVirtualServiceRead,
		UpdateWithoutTimeout: resourceVirtualServiceUpdate,
		DeleteWithoutTimeout: resourceVirtualServiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVirtualServiceImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"mesh_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"mesh_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},

			"spec": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"provider": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"virtual_node": {
										Type:          schema.TypeList,
										Optional:      true,
										MinItems:      0,
										MaxItems:      1,
										ConflictsWith: []string{"spec.0.provider.0.virtual_router"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"virtual_node_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 255),
												},
											},
										},
									},

									"virtual_router": {
										Type:          schema.TypeList,
										Optional:      true,
										MinItems:      0,
										MaxItems:      1,
										ConflictsWith: []string{"spec.0.provider.0.virtual_node"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"virtual_router_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 255),
												},
											},
										},
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

func resourceVirtualServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	req := &appmesh.CreateVirtualServiceInput{
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		VirtualServiceName: aws.String(d.Get("name").(string)),
		Spec:               expandVirtualServiceSpec(d.Get("spec").([]interface{})),
		Tags:               Tags(tags.IgnoreAWS()),
	}
	if v, ok := d.GetOk("mesh_owner"); ok {
		req.MeshOwner = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating App Mesh virtual service: %#v", req)
	resp, err := conn.CreateVirtualServiceWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating App Mesh virtual service: %s", err)
	}

	d.SetId(aws.StringValue(resp.VirtualService.Metadata.Uid))

	return append(diags, resourceVirtualServiceRead(ctx, d, meta)...)
}

func resourceVirtualServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	req := &appmesh.DescribeVirtualServiceInput{
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		VirtualServiceName: aws.String(d.Get("name").(string)),
	}
	if v, ok := d.GetOk("mesh_owner"); ok {
		req.MeshOwner = aws.String(v.(string))
	}

	var resp *appmesh.DescribeVirtualServiceOutput

	err := resource.RetryContext(ctx, propagationTimeout, func() *resource.RetryError {
		var err error

		resp, err = conn.DescribeVirtualServiceWithContext(ctx, req)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		resp, err = conn.DescribeVirtualServiceWithContext(ctx, req)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
		log.Printf("[WARN] App Mesh Virtual Service (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Virtual Service: %s", err)
	}

	if resp == nil || resp.VirtualService == nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Virtual Service: empty response")
	}

	if aws.StringValue(resp.VirtualService.Status.Status) == appmesh.VirtualServiceStatusCodeDeleted {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading App Mesh Virtual Service: %s after creation", aws.StringValue(resp.VirtualService.Status.Status))
		}

		log.Printf("[WARN] App Mesh Virtual Service (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	arn := aws.StringValue(resp.VirtualService.Metadata.Arn)
	d.Set("name", resp.VirtualService.VirtualServiceName)
	d.Set("mesh_name", resp.VirtualService.MeshName)
	d.Set("mesh_owner", resp.VirtualService.Metadata.MeshOwner)
	d.Set("arn", arn)
	d.Set("created_date", resp.VirtualService.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", resp.VirtualService.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("resource_owner", resp.VirtualService.Metadata.ResourceOwner)
	err = d.Set("spec", flattenVirtualServiceSpec(resp.VirtualService.Spec))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting spec: %s", err)
	}

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for App Mesh virtual service (%s): %s", arn, err)
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

func resourceVirtualServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()

	if d.HasChange("spec") {
		_, v := d.GetChange("spec")
		req := &appmesh.UpdateVirtualServiceInput{
			MeshName:           aws.String(d.Get("mesh_name").(string)),
			VirtualServiceName: aws.String(d.Get("name").(string)),
			Spec:               expandVirtualServiceSpec(v.([]interface{})),
		}
		if v, ok := d.GetOk("mesh_owner"); ok {
			req.MeshOwner = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Updating App Mesh virtual service: %#v", req)
		_, err := conn.UpdateVirtualServiceWithContext(ctx, req)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating App Mesh virtual service: %s", err)
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, arn, o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating App Mesh virtual service (%s) tags: %s", arn, err)
		}
	}

	return append(diags, resourceVirtualServiceRead(ctx, d, meta)...)
}

func resourceVirtualServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()

	log.Printf("[DEBUG] Deleting App Mesh Virtual Service: %s", d.Id())
	_, err := conn.DeleteVirtualServiceWithContext(ctx, &appmesh.DeleteVirtualServiceInput{
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		VirtualServiceName: aws.String(d.Get("name").(string)),
	})
	if tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting App Mesh virtual service: %s", err)
	}

	return diags
}

func resourceVirtualServiceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'mesh-name/virtual-service-name'", d.Id())
	}

	mesh := parts[0]
	name := parts[1]
	log.Printf("[DEBUG] Importing App Mesh virtual service %s from mesh %s", name, mesh)

	conn := meta.(*conns.AWSClient).AppMeshConn()

	resp, err := conn.DescribeVirtualServiceWithContext(ctx, &appmesh.DescribeVirtualServiceInput{
		MeshName:           aws.String(mesh),
		VirtualServiceName: aws.String(name),
	})
	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(resp.VirtualService.Metadata.Uid))
	d.Set("name", resp.VirtualService.VirtualServiceName)
	d.Set("mesh_name", resp.VirtualService.MeshName)

	return []*schema.ResourceData{d}, nil
}
