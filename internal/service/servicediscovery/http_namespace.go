package servicediscovery

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceHTTPNamespace() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHTTPNamespaceCreate,
		ReadWithoutTimeout:   resourceHTTPNamespaceRead,
		UpdateWithoutTimeout: resourceHTTPNamespaceUpdate,
		DeleteWithoutTimeout: resourceHTTPNamespaceDelete,

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
			"http_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validNamespaceName,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceHTTPNamespaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &servicediscovery.CreateHttpNamespaceInput{
		CreatorRequestId: aws.String(resource.UniqueId()),
		Name:             aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Service Discovery HTTP Namespace: %s", input)
	output, err := conn.CreateHttpNamespaceWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Service Discovery HTTP Namespace (%s): %s", name, err)
	}

	operation, err := WaitOperationSuccess(ctx, conn, aws.StringValue(output.OperationId))

	if err != nil {
		return diag.Errorf("waiting for Service Discovery HTTP Namespace (%s) create: %s", name, err)
	}

	namespaceID, ok := operation.Targets[servicediscovery.OperationTargetTypeNamespace]

	if !ok {
		return diag.Errorf("creating Service Discovery HTTP Namespace (%s): operation response missing Namespace ID", name)
	}

	d.SetId(aws.StringValue(namespaceID))

	return resourceHTTPNamespaceRead(ctx, d, meta)
}

func resourceHTTPNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	ns, err := FindNamespaceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Discovery HTTP Namespace %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Service Discovery HTTP Namespace (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(ns.Arn)
	d.Set("arn", arn)
	d.Set("description", ns.Description)
	if ns.Properties != nil && ns.Properties.HttpProperties != nil {
		d.Set("http_name", ns.Properties.HttpProperties.HttpName)
	} else {
		d.Set("http_name", nil)
	}
	d.Set("name", ns.Name)

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("listing tags for Service Discovery HTTP Namespace (%s): %s", arn, err)
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

func resourceHTTPNamespaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("updating Service Discovery HTTP Namespace (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceHTTPNamespaceRead(ctx, d, meta)
}

func resourceHTTPNamespaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn()

	log.Printf("[INFO] Deleting Service Discovery HTTP Namespace: %s", d.Id())
	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 2*time.Minute, func() (interface{}, error) {
		return conn.DeleteNamespaceWithContext(ctx, &servicediscovery.DeleteNamespaceInput{
			Id: aws.String(d.Id()),
		})
	}, servicediscovery.ErrCodeResourceInUse)

	if tfawserr.ErrCodeEquals(err, servicediscovery.ErrCodeNamespaceNotFound) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Service Discovery HTTP Namespace (%s): %s", d.Id(), err)
	}

	if output := outputRaw.(*servicediscovery.DeleteNamespaceOutput); output != nil && output.OperationId != nil {
		if _, err := WaitOperationSuccess(ctx, conn, aws.StringValue(output.OperationId)); err != nil {
			return diag.Errorf("waiting for Service Discovery HTTP Namespace (%s) delete: %s", d.Id(), err)
		}
	}

	return nil
}
