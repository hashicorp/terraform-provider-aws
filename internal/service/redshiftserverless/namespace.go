package redshiftserverless

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceNamespace() *schema.Resource {
	return &schema.Resource{
		Create: resourceNamespaceCreate,
		Read:   resourceNamespaceRead,
		Update: resourceNamespaceUpdate,
		Delete: resourceNamespaceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"namespace_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceNamespaceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := redshiftserverless.CreateNamespaceInput{
		NamespaceName: aws.String(d.Get("namespace_name").(string)),
	}

	// if v, ok := d.GetOk("breach_action"); ok {
	// 	input.BreachAction = aws.String(v.(string))
	// }

	// if v, ok := d.GetOk("period"); ok {
	// 	input.Period = aws.String(v.(string))
	// }

	input.Tags = Tags(tags.IgnoreAWS())

	out, err := conn.CreateNamespace(&input)

	if err != nil {
		return fmt.Errorf("error creating Redshift Serverless Namespace : %w", err)
	}

	d.SetId(aws.StringValue(out.Namespace.NamespaceName))

	return resourceNamespaceRead(d, meta)
}

func resourceNamespaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	out, err := FindNamespaceByName(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Serverless Namespace (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Redshift Serverless Namespace (%s): %w", d.Id(), err)
	}

	arn := aws.StringValue(out.NamespaceArn)
	d.Set("arn", arn)
	d.Set("namespace_name", out.NamespaceName)

	// d.Set("period", out.Period)
	// d.Set("limit_type", out.LimitType)
	// d.Set("feature_type", out.FeatureType)
	// d.Set("breach_action", out.BreachAction)
	// d.Set("cluster_identifier", out.ClusterIdentifier)

	tags, err := ListTags(conn, arn)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, "UnknownOperationException") {
			return nil
		}

		return fmt.Errorf("error listing tags for edshift Serverless Namespace (%s): %w", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceNamespaceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn

	// if d.HasChangesExcept("tags", "tags_all") {
	// 	input := &redshiftserverless.ModifyNamespaceInput{
	// 		NamespaceId: aws.String(d.Id()),
	// 	}

	// 	if d.HasChange("amount") {
	// 		input.Amount = aws.Int64(int64(d.Get("amount").(int)))
	// 	}

	// 	if d.HasChange("breach_action") {
	// 		input.BreachAction = aws.String(d.Get("breach_action").(string))
	// 	}

	// 	_, err := conn.ModifyNamespace(input)
	// 	if err != nil {
	// 		return fmt.Errorf("error updating Redshift Serverless Namespace (%s): %w", d.Id(), err)
	// 	}
	// }

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Redshift Serverless Namespace (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceNamespaceRead(d, meta)
}

func resourceNamespaceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RedshiftServerlessConn

	deleteInput := redshiftserverless.DeleteNamespaceInput{
		NamespaceName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting snapshot copy grant: %s", d.Id())
	_, err := conn.DeleteNamespace(&deleteInput)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, redshiftserverless.ErrCodeResourceNotFoundException) {
			return nil
		}
		return err
	}

	return nil
}
