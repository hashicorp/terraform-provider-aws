// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeartifact

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codeartifact_repository", name="Repository")
// @Tags(identifierAttribute="arn")
func ResourceRepository() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRepositoryCreate,
		ReadWithoutTimeout:   resourceRepositoryRead,
		UpdateWithoutTimeout: resourceRepositoryUpdate,
		DeleteWithoutTimeout: resourceRepositoryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"upstream": {
				Type:     schema.TypeList,
				MinItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"repository_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"external_connections": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"external_connection_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"package_format": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"administrator_account": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRepositoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactConn(ctx)

	input := &codeartifact.CreateRepositoryInput{
		Repository: aws.String(d.Get("repository").(string)),
		Domain:     aws.String(d.Get("domain").(string)),
		Tags:       getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("domain_owner"); ok {
		input.DomainOwner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("upstream"); ok {
		input.Upstreams = expandUpstreams(v.([]interface{}))
	}

	res, err := conn.CreateRepositoryWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeArtifact Repository: %s", err)
	}

	repo := res.Repository
	d.SetId(aws.StringValue(repo.Arn))

	if v, ok := d.GetOk("external_connections"); ok {
		externalConnection := v.([]interface{})[0].(map[string]interface{})
		input := &codeartifact.AssociateExternalConnectionInput{
			Domain:             repo.DomainName,
			Repository:         repo.Name,
			DomainOwner:        repo.DomainOwner,
			ExternalConnection: aws.String(externalConnection["external_connection_name"].(string)),
		}

		_, err := conn.AssociateExternalConnectionWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "associating external connection to CodeArtifact repository: %s", err)
		}
	}

	return append(diags, resourceRepositoryRead(ctx, d, meta)...)
}

func resourceRepositoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactConn(ctx)

	owner, domain, repo, err := DecodeRepositoryID(d.Id())
	if err != nil {
		return create.DiagError(names.CodeArtifact, create.ErrActionReading, ResNameRepository, d.Id(), err)
	}
	sm, err := conn.DescribeRepositoryWithContext(ctx, &codeartifact.DescribeRepositoryInput{
		Repository:  aws.String(repo),
		Domain:      aws.String(domain),
		DomainOwner: aws.String(owner),
	})
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, codeartifact.ErrCodeResourceNotFoundException) {
		create.LogNotFoundRemoveState(names.CodeArtifact, create.ErrActionReading, ResNameRepository, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.CodeArtifact, create.ErrActionReading, ResNameRepository, d.Id(), err)
	}

	arn := aws.StringValue(sm.Repository.Arn)
	d.Set("repository", sm.Repository.Name)
	d.Set("arn", arn)
	d.Set("domain_owner", sm.Repository.DomainOwner)
	d.Set("domain", sm.Repository.DomainName)
	d.Set("administrator_account", sm.Repository.AdministratorAccount)
	d.Set("description", sm.Repository.Description)

	if sm.Repository.Upstreams != nil {
		if err := d.Set("upstream", flattenUpstreams(sm.Repository.Upstreams)); err != nil {
			return sdkdiag.AppendErrorf(diags, "[WARN] Error setting upstream: %s", err)
		}
	}

	if sm.Repository.ExternalConnections != nil {
		if err := d.Set("external_connections", flattenExternalConnections(sm.Repository.ExternalConnections)); err != nil {
			return sdkdiag.AppendErrorf(diags, "[WARN] Error setting external_connections: %s", err)
		}
	}

	return diags
}

func resourceRepositoryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactConn(ctx)
	log.Print("[DEBUG] Updating CodeArtifact Repository")

	needsUpdate := false
	params := &codeartifact.UpdateRepositoryInput{
		Repository:  aws.String(d.Get("repository").(string)),
		Domain:      aws.String(d.Get("domain").(string)),
		DomainOwner: aws.String(d.Get("domain_owner").(string)),
	}

	if d.HasChange("description") {
		if v, ok := d.GetOk("description"); ok {
			params.Description = aws.String(v.(string))
			needsUpdate = true
		}
	}

	if d.HasChange("upstream") {
		if v, ok := d.GetOk("upstream"); ok {
			params.Upstreams = expandUpstreams(v.([]interface{}))
			needsUpdate = true
		}
	}

	if needsUpdate {
		_, err := conn.UpdateRepositoryWithContext(ctx, params)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CodeArtifact Repository: %s", err)
		}
	}

	if d.HasChange("external_connections") {
		if v, ok := d.GetOk("external_connections"); ok {
			externalConnection := v.([]interface{})[0].(map[string]interface{})
			input := &codeartifact.AssociateExternalConnectionInput{
				Repository:         aws.String(d.Get("repository").(string)),
				Domain:             aws.String(d.Get("domain").(string)),
				DomainOwner:        aws.String(d.Get("domain_owner").(string)),
				ExternalConnection: aws.String(externalConnection["external_connection_name"].(string)),
			}

			_, err := conn.AssociateExternalConnectionWithContext(ctx, input)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "associating external connection to CodeArtifact repository: %s", err)
			}
		} else {
			oldConn, _ := d.GetChange("external_connections")
			externalConnection := oldConn.([]interface{})[0].(map[string]interface{})
			input := &codeartifact.DisassociateExternalConnectionInput{
				Repository:         aws.String(d.Get("repository").(string)),
				Domain:             aws.String(d.Get("domain").(string)),
				DomainOwner:        aws.String(d.Get("domain_owner").(string)),
				ExternalConnection: aws.String(externalConnection["external_connection_name"].(string)),
			}

			_, err := conn.DisassociateExternalConnectionWithContext(ctx, input)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disassociating external connection to CodeArtifact repository: %s", err)
			}
		}
	}

	return append(diags, resourceRepositoryRead(ctx, d, meta)...)
}

func resourceRepositoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeArtifactConn(ctx)
	log.Printf("[DEBUG] Deleting CodeArtifact Repository: %s", d.Id())

	owner, domain, repo, err := DecodeRepositoryID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeArtifact Repository (%s): %s", d.Id(), err)
	}
	input := &codeartifact.DeleteRepositoryInput{
		Repository:  aws.String(repo),
		Domain:      aws.String(domain),
		DomainOwner: aws.String(owner),
	}

	_, err = conn.DeleteRepositoryWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, codeartifact.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeArtifact Repository (%s): %s", d.Id(), err)
	}

	return diags
}

func expandUpstreams(l []interface{}) []*codeartifact.UpstreamRepository {
	upstreams := []*codeartifact.UpstreamRepository{}

	for _, mRaw := range l {
		m := mRaw.(map[string]interface{})
		upstream := &codeartifact.UpstreamRepository{
			RepositoryName: aws.String(m["repository_name"].(string)),
		}

		upstreams = append(upstreams, upstream)
	}

	return upstreams
}

func flattenUpstreams(upstreams []*codeartifact.UpstreamRepositoryInfo) []interface{} {
	if len(upstreams) == 0 {
		return nil
	}

	var ls []interface{}

	for _, upstream := range upstreams {
		m := map[string]interface{}{
			"repository_name": aws.StringValue(upstream.RepositoryName),
		}

		ls = append(ls, m)
	}

	return ls
}

func flattenExternalConnections(connections []*codeartifact.RepositoryExternalConnectionInfo) []interface{} {
	if len(connections) == 0 {
		return nil
	}

	var ls []interface{}

	for _, connection := range connections {
		m := map[string]interface{}{
			"external_connection_name": aws.StringValue(connection.ExternalConnectionName),
			"package_format":           aws.StringValue(connection.PackageFormat),
			"status":                   aws.StringValue(connection.Status),
		}

		ls = append(ls, m)
	}

	return ls
}

func DecodeRepositoryID(id string) (string, string, string, error) {
	repoArn, err := arn.Parse(id)
	if err != nil {
		return "", "", "", err
	}

	idParts := strings.Split(strings.TrimPrefix(repoArn.Resource, "repository/"), "/")
	if len(idParts) != 2 {
		return "", "", "", fmt.Errorf("expected resource part of arn in format DomainName/RepositoryName, received: %s", repoArn.Resource)
	}
	return repoArn.AccountID, idParts[0], idParts[1], nil
}
