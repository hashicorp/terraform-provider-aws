package servicecatalog

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceTagOption() *schema.Resource {
	return &schema.Resource{
		Create: resourceTagOptionCreate,
		Read:   resourceTagOptionRead,
		Update: resourceTagOptionUpdate,
		Delete: resourceTagOptionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"value": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceTagOptionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceCatalogConn

	input := &servicecatalog.CreateTagOptionInput{
		Key:   aws.String(d.Get("key").(string)),
		Value: aws.String(d.Get("value").(string)),
	}

	var output *servicecatalog.CreateTagOptionOutput
	err := resource.Retry(tfiam.PropagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.CreateTagOption(input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateTagOption(input)
	}

	if err != nil {
		return fmt.Errorf("error creating Service Catalog Tag Option: %w", err)
	}

	if output == nil || output.TagOptionDetail == nil {
		return fmt.Errorf("error creating Service Catalog Tag Option: empty response")
	}

	d.SetId(aws.StringValue(output.TagOptionDetail.Id))

	// Active is not a field of CreateTagOption but is a field of UpdateTagOption. In order to create an
	// inactive Tag Option, you must create an active one and then update it (but calling this resource's
	// Update will error with ErrCodeDuplicateResourceException because Value is unchanged).
	if v, ok := d.GetOk("active"); !ok {
		_, err = conn.UpdateTagOption(&servicecatalog.UpdateTagOptionInput{
			Id:     aws.String(d.Id()),
			Active: aws.Bool(v.(bool)),
		})

		if err != nil {
			return fmt.Errorf("error creating Service Catalog Tag Option, updating active (%s): %w", d.Id(), err)
		}
	}

	return resourceTagOptionRead(d, meta)
}

func resourceTagOptionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceCatalogConn

	output, err := WaitTagOptionReady(conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Service Catalog Tag Option (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Service Catalog Tag Option (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting Service Catalog Tag Option (%s): empty response", d.Id())
	}

	d.Set("active", output.Active)
	d.Set("key", output.Key)
	d.Set("owner", output.Owner)
	d.Set("value", output.Value)

	return nil
}

func resourceTagOptionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceCatalogConn

	input := &servicecatalog.UpdateTagOptionInput{
		Id: aws.String(d.Id()),
	}

	// UpdateTagOption() is very particular about what it receives. Only fields that change should
	// be included or it will throw servicecatalog.ErrCodeDuplicateResourceException, "already exists"

	if d.HasChange("active") {
		input.Active = aws.Bool(d.Get("active").(bool))
	}

	if d.HasChange("value") {
		input.Value = aws.String(d.Get("value").(string))
	}

	err := resource.Retry(tfiam.PropagationTimeout, func() *resource.RetryError {
		_, err := conn.UpdateTagOption(input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.UpdateTagOption(input)
	}

	if err != nil {
		return fmt.Errorf("error updating Service Catalog Tag Option (%s): %w", d.Id(), err)
	}

	return resourceTagOptionRead(d, meta)
}

func resourceTagOptionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceCatalogConn

	input := &servicecatalog.DeleteTagOptionInput{
		Id: aws.String(d.Id()),
	}

	_, err := conn.DeleteTagOption(input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Service Catalog Tag Option (%s): %w", d.Id(), err)
	}

	if err := WaitTagOptionDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Service Catalog Tag Option (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
