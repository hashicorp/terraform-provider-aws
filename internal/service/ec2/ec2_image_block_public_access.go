package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_ec2_image_block_public_access", name="Image Block Public Access")
func ResourceEC2ImageBlockPublicAccess() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEC2ImageBlockPublicAccessCreate,
		ReadWithoutTimeout:   resourceEC2ImageBlockPublicAccessRead,
		UpdateWithoutTimeout: resourceEC2ImageBlockPublicAccessUpdate,
		DeleteWithoutTimeout: resourceEC2ImageBlockPublicAccessDelete,

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceEC2ImageBlockPublicAccessCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var desiredState string
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	// It is possible to create this Terraform resource with block public access disabled
	// By default, AWS accounts allow sharing AMIs publicly (effectively `enabled=false`)
	if d.Get("enabled").(bool) {
		desiredState = string(ec2types.ImageBlockPublicAccessEnabledStateBlockNewSharing)
		diags = append(diags, enableEC2ImageBlockPublicAccess(ctx, d, conn)...)
	} else {
		desiredState = string(ec2types.ImageBlockPublicAccessDisabledStateUnblocked)
		diags = append(diags, disableEC2ImageBlockPublicAccess(ctx, d, conn)...)
	}

	err := waitForEC2ImageBlockPublicAccessDesiredState(
		ctx,
		desiredState,
		conn,
	)

	if err != nil {
		return append(diags, sdkdiag.AppendErrorf(diags, "setting EC2 image block public access to '%s' (%s): %s", desiredState, d.Id(), err)...)
	}

	// There's no unique identifier for this resource
	d.SetId("ec2_image_block_public_access")

	return append(diags, resourceEC2ImageBlockPublicAccessRead(ctx, d, meta)...)
}

func resourceEC2ImageBlockPublicAccessRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	blockState, err := getEC2ImageBlockPublicAccessState(ctx, conn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 image block public access state (%s): %s", d.Id(), err)
	}

	if blockState == string(ec2types.ImageBlockPublicAccessEnabledStateBlockNewSharing) {
		d.Set("enabled", true)
	} else {
		d.Set("enabled", false)
	}

	return diags
}

func resourceEC2ImageBlockPublicAccessDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// In order to delete the public access block on the AWS-side, we simply set it to a disabled state
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	diags = append(diags, disableEC2ImageBlockPublicAccess(ctx, d, conn)...)

	err := waitForEC2ImageBlockPublicAccessDesiredState(
		ctx,
		string(ec2types.ImageBlockPublicAccessDisabledStateUnblocked),
		conn,
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disabling ec2 public image access block %s", err)
	}

	d.SetId("")

	return diags
}

func resourceEC2ImageBlockPublicAccessUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChange("enabled") {
		var desiredState string

		_, n := d.GetChange("enabled")
		enabled := n.(bool)

		if enabled {
			diags = append(diags, enableEC2ImageBlockPublicAccess(ctx, d, conn)...)
			desiredState = string(ec2types.ImageBlockPublicAccessEnabledStateBlockNewSharing)
		} else {
			diags = append(diags, disableEC2ImageBlockPublicAccess(ctx, d, conn)...)
			desiredState = string(ec2types.ImageBlockPublicAccessDisabledStateUnblocked)
		}

		// Wait for change to propagate, which can take up to 10 minutes
		err := waitForEC2ImageBlockPublicAccessDesiredState(
			ctx,
			desiredState,
			conn,
		)

		if err != nil {
			return append(diags, sdkdiag.AppendErrorf(diags, "setting EC2 image block public access to '%s' (%s): %s", desiredState, d.Id(), err)...)
		}
	}

	return append(diags, resourceEC2ImageBlockPublicAccessRead(ctx, d, meta)...)
}

func getEC2ImageBlockPublicAccessState(ctx context.Context, client *ec2.Client) (string, error) {
	input := ec2.GetImageBlockPublicAccessStateInput{}

	output, err := client.GetImageBlockPublicAccessState(ctx, &input)

	if err != nil {
		log.Printf("[ERROR] reading image block public access state %s", err.Error())
		return "", err
	}

	return *output.ImageBlockPublicAccessState, err
}

func disableEC2ImageBlockPublicAccess(ctx context.Context, d *schema.ResourceData, client *ec2.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	input := ec2.DisableImageBlockPublicAccessInput{}

	_, err := client.DisableImageBlockPublicAccess(ctx, &input)

	if err != nil {
		log.Printf("[ERROR] disabling image block public access %s", err.Error())
		return sdkdiag.AppendErrorf(diags, "removing EC2 image block public access (%s): %s", d.Id(), err)
	}

	return diags
}

func enableEC2ImageBlockPublicAccess(ctx context.Context, d *schema.ResourceData, client *ec2.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	input := ec2.EnableImageBlockPublicAccessInput{
		ImageBlockPublicAccessState: ec2types.ImageBlockPublicAccessEnabledStateBlockNewSharing,
	}

	_, err := client.EnableImageBlockPublicAccess(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 image block public access (%s): %s", d.Id(), err)
	}

	return diags
}

func waitForEC2ImageBlockPublicAccessDesiredState(ctx context.Context, desired string, client *ec2.Client) error {
	blockState, err := getEC2ImageBlockPublicAccessState(ctx, client)

	if err != nil {
		log.Printf("[ERROR] waiting for EC2 image block public access state to be '%s'", desired)
		return err
	}

	for blockState != desired {
		log.Printf(
			"[DEBUG] waiting for EC2 image block public access state to be '%s', it is '%s'",
			desired,
			blockState,
		)
		time.Sleep(10 * time.Second)

		blockState, err = getEC2ImageBlockPublicAccessState(ctx, client)

		if err != nil {
			log.Printf("[ERROR] waiting for EC2 image block public access state to be '%s' %s", desired, err.Error())
			return err
		}
	}

	return nil
}
