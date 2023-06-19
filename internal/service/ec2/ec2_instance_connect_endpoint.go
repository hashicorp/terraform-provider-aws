package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_instance_state")
func ResourceInstanceConnectEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceConnectEndpointCreate,
		ReadWithoutTimeout:   resourceInstanceConnectEndpointRead,
		UpdateWithoutTimeout: resourceInstanceConnectEndpointUpdate,
		DeleteWithoutTimeout: resourceInstanceConnectEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_group_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"preserve_client_ip": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				Default:  true,
			},
		},
	}
}

func resourceInstanceConnectEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)
	instanceId := d.Get("instance_id").(string)

	d.SetId(d.Get("instance_id").(string))

	return resourceInstanceConnectEndpointRead(ctx, d, meta)
}

func resourceInstanceConnectEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	state, err := FindInstanceConnectEndpointByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.EC2, create.ErrActionReading, ResInstanceConnectEndpoint, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.EC2, create.ErrActionReading, ResInstanceConnectEndpoint, d.Id(), err)
	}

	d.Set("instance_id", d.Id())
	d.Set("state", state.Name)
	d.Set("force", d.Get("force").(bool))

	return nil
}

func resourceInstanceConnectEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	instance, instanceErr := WaitInstanceReadyWithContext(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate))

	if instanceErr != nil {
		return create.DiagError(names.EC2, create.ErrActionReading, ResInstance, aws.StringValue(instance.InstanceId), instanceErr)
	}

	if d.HasChange("state") {
		o, n := d.GetChange("state")
		err := UpdateInstanceConnectEndpoint(ctx, conn, d.Id(), o.(string), n.(string), d.Get("force").(bool))

		if err != nil {
			return err
		}
	}

	return resourceInstanceConnectEndpointRead(ctx, d, meta)
}

func resourceInstanceConnectEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] %s %s deleting an aws_ec2_instance_state resource only stops managing instance state, The Instance is left in its current state.: %s", names.EC2, ResInstanceConnectEndpoint, d.Id())

	return nil
}
