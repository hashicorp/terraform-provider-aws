package ds

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceRadiusSettings() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRadiusSettingsCreate,
		ReadWithoutTimeout:   resourceRadiusSettingsRead,
		UpdateWithoutTimeout: resourceRadiusSettingsUpdate,
		DeleteWithoutTimeout: resourceRadiusSettingsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"authentication_protocol": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(directoryservice.RadiusAuthenticationProtocol_Values(), false),
			},
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"display_label": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"radius_port": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IsPortNumber,
			},
			"radius_retries": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 10),
			},
			"radius_servers": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 256),
				},
			},
			"radius_timeout": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 50),
			},
			"shared_secret": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(8, 512),
			},
			"use_same_username": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceRadiusSettingsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DSConn()

	directoryID := d.Get("directory_id").(string)
	input := &directoryservice.EnableRadiusInput{
		DirectoryId: aws.String(directoryID),
		RadiusSettings: &directoryservice.RadiusSettings{
			AuthenticationProtocol: aws.String(d.Get("authentication_protocol").(string)),
			DisplayLabel:           aws.String(d.Get("display_label").(string)),
			RadiusPort:             aws.Int64(int64(d.Get("radius_port").(int))),
			RadiusRetries:          aws.Int64(int64(d.Get("radius_retries").(int))),
			RadiusServers:          flex.ExpandStringSet(d.Get("radius_servers").(*schema.Set)),
			RadiusTimeout:          aws.Int64(int64(d.Get("radius_timeout").(int))),
			SharedSecret:           aws.String(d.Get("shared_secret").(string)),
			UseSameUsername:        aws.Bool(d.Get("use_same_username").(bool)),
		},
	}

	log.Printf("[DEBUG] Enabling Directory Service Directory RADIUS: %s", input)
	_, err := conn.EnableRadiusWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("enabling Directory Service Directory (%s) RADIUS: %s", directoryID, err)
	}

	d.SetId(directoryID)

	if _, err := waitRadiusCompleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Directory Service Directory (%s) RADIUS create: %s", d.Id(), err)
	}

	return resourceRadiusSettingsRead(ctx, d, meta)
}

func resourceRadiusSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DSConn()

	output, err := FindRadiusSettings(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Directory Service Directory (%s) RADIUS settings not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Directory Service Directory (%s) RADIUS settings: %s", d.Id(), err)
	}

	d.Set("authentication_protocol", output.AuthenticationProtocol)
	d.Set("display_label", output.DisplayLabel)
	d.Set("directory_id", d.Id())
	d.Set("radius_port", output.RadiusPort)
	d.Set("radius_retries", output.RadiusRetries)
	d.Set("radius_servers", aws.StringValueSlice(output.RadiusServers))
	d.Set("radius_timeout", output.RadiusTimeout)
	d.Set("shared_secret", output.SharedSecret)
	d.Set("use_same_username", output.UseSameUsername)

	return nil
}

func resourceRadiusSettingsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DSConn()

	input := &directoryservice.UpdateRadiusInput{
		DirectoryId: aws.String(d.Id()),
		RadiusSettings: &directoryservice.RadiusSettings{
			AuthenticationProtocol: aws.String(d.Get("authentication_protocol").(string)),
			DisplayLabel:           aws.String(d.Get("display_label").(string)),
			RadiusPort:             aws.Int64(int64(d.Get("radius_port").(int))),
			RadiusRetries:          aws.Int64(int64(d.Get("radius_retries").(int))),
			RadiusServers:          flex.ExpandStringSet(d.Get("radius_servers").(*schema.Set)),
			RadiusTimeout:          aws.Int64(int64(d.Get("radius_timeout").(int))),
			SharedSecret:           aws.String(d.Get("shared_secret").(string)),
			UseSameUsername:        aws.Bool(d.Get("use_same_username").(bool)),
		},
	}

	log.Printf("[DEBUG] Updating Directory Service Directory RADIUS: %s", input)
	_, err := conn.UpdateRadiusWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("updating Directory Service Directory (%s) RADIUS: %s", d.Id(), err)
	}

	if _, err := waitRadiusCompleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return diag.Errorf("waiting for Directory Service Directory (%s) RADIUS update: %s", d.Id(), err)
	}

	return resourceRadiusSettingsRead(ctx, d, meta)
}

func resourceRadiusSettingsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DSConn()

	_, err := conn.DisableRadiusWithContext(ctx, &directoryservice.DisableRadiusInput{
		DirectoryId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeDirectoryDoesNotExistException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("disabling Directory Service Directory (%s) RADIUS: %s", d.Id(), err)
	}

	return nil
}
