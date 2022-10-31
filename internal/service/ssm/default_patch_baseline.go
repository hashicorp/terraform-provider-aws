package ssm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type lazyClient[T any] struct {
	initOnce sync.Once
	config   *aws.Config
	initf    func(aws.Config) T

	clientOnce sync.Once
	client     T
}

func (l *lazyClient[T]) Init(config *aws.Config, f func(aws.Config) T) {
	l.initOnce.Do(func() {
		l.config = config
		l.initf = f
	})
}

func (l *lazyClient[T]) Client() T {
	l.clientOnce.Do(func() {
		l.client = l.initf(*l.config)
	})
	return l.client
}

var SSMClientV2 lazyClient[*ssm.Client]

func ResourceDefaultPatchBaseline() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDefaultPatchBaselineRegister,
		ReadWithoutTimeout:   resourceDefaultPatchBaselineRead,
		UpdateWithoutTimeout: resourceDefaultPatchBaselineRegister,
		DeleteWithoutTimeout: resourceDefaultPatchBaselineDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"baseline_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

const (
	ResNameDefaultPatchBaseline = "Default Patch Baseline"
)

func resourceDefaultPatchBaselineRegister(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	SSMClientV2.Init(meta.(*conns.AWSClient).Config, func(c aws.Config) *ssm.Client {
		return ssm.NewFromConfig(c)
	})
	conn := SSMClientV2.Client()

	baselineID := d.Get("baseline_id").(string)
	in := &ssm.RegisterDefaultPatchBaselineInput{
		BaselineId: aws.String(baselineID),
	}
	_, err := conn.RegisterDefaultPatchBaseline(ctx, in)
	if err != nil {
		return create.DiagError(names.SSM, "registering", ResNameDefaultPatchBaseline, baselineID, err)
	}

	// We need to retrieve the Operating System from the Patch Baseline to store for the ID
	patchBaseline, err := findPatchBaselineByID(ctx, conn, baselineID)
	if err != nil {
		return create.DiagError(names.SSM, "registering", ResNameDefaultPatchBaseline, baselineID,
			create.Error(names.SSM, create.ErrActionReading, resNamePatchBaseline, baselineID, err),
		)
	}

	d.SetId(string(patchBaseline.OperatingSystem))

	return resourceDefaultPatchBaselineRead(ctx, d, meta)
}

func resourceDefaultPatchBaselineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	SSMClientV2.Init(meta.(*conns.AWSClient).Config, func(c aws.Config) *ssm.Client {
		return ssm.NewFromConfig(c)
	})
	conn := SSMClientV2.Client()

	out, err := FindDefaultPatchBaseline(ctx, conn, types.OperatingSystem(d.Id()))
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSM Default Patch Baseline (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return create.DiagError(names.SSM, create.ErrActionReading, ResNameDefaultPatchBaseline, d.Id(), err)
	}

	d.Set("baseline_id", out.BaselineId)

	return nil
}

func operatingSystemFilter(os ...string) types.PatchOrchestratorFilter {
	return types.PatchOrchestratorFilter{
		Key:    aws.String("OPERATING_SYSTEM"),
		Values: os,
	}
}

func ownerIsAWSFilter() types.PatchOrchestratorFilter { // nosemgrep:ci.aws-in-func-name
	return types.PatchOrchestratorFilter{
		Key:    aws.String("OWNER"),
		Values: []string{"AWS"},
	}
}

func resourceDefaultPatchBaselineDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	SSMClientV2.Init(meta.(*conns.AWSClient).Config, func(c aws.Config) *ssm.Client {
		return ssm.NewFromConfig(c)
	})
	conn := SSMClientV2.Client()

	baselineID, err := FindDefaultDefaultPatchBaselineIDForOS(ctx, conn, types.OperatingSystem(d.Id()))
	if errors.Is(err, tfresource.ErrEmptyResult) {
		diags = errs.AppendWarningf(diags, "no AWS-owned default Patch Baseline found for operating system %q", d.Id())
		return
	}
	var tmr *tfresource.TooManyResultsError
	if errors.As(err, &tmr) {
		diags = errs.AppendWarningf(diags, "found %d AWS-owned default Patch Baselines found for operating system %q", tmr.Count, d.Id())
	}

	in := &ssm.RegisterDefaultPatchBaselineInput{
		BaselineId: aws.String(baselineID),
	}
	_, err = conn.RegisterDefaultPatchBaseline(ctx, in)
	if err != nil {
		diags = errs.AppendErrorf(diags, "restoring SSM Default Patch Baseline for operating system %q to %q: %s", d.Id(), baselineID, err)
	}

	return
}

func FindDefaultPatchBaseline(ctx context.Context, conn *ssm.Client, os types.OperatingSystem) (*ssm.GetDefaultPatchBaselineOutput, error) {
	in := &ssm.GetDefaultPatchBaselineInput{
		OperatingSystem: os,
	}
	out, err := conn.GetDefaultPatchBaseline(ctx, in)
	if err != nil {
		var nfe *types.DoesNotExistException
		if errors.As(err, &nfe) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func findPatchBaselineByID(ctx context.Context, conn *ssm.Client, id string) (*ssm.GetPatchBaselineOutput, error) {
	in := &ssm.GetPatchBaselineInput{
		BaselineId: aws.String(id),
	}
	out, err := conn.GetPatchBaseline(ctx, in)
	if err != nil {
		var nfe *types.DoesNotExistException
		if errors.As(err, &nfe) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func patchBaselinesPaginator(conn *ssm.Client, filters ...types.PatchOrchestratorFilter) *ssm.DescribePatchBaselinesPaginator {
	return ssm.NewDescribePatchBaselinesPaginator(conn, &ssm.DescribePatchBaselinesInput{
		Filters: filters,
	})
}

func FindDefaultDefaultPatchBaselineIDForOS(ctx context.Context, conn *ssm.Client, os types.OperatingSystem) (string, error) {
	paginator := patchBaselinesPaginator(conn,
		operatingSystemFilter(string(os)),
		ownerIsAWSFilter(),
	)
	re := regexp.MustCompile(`^AWS-[A-Za-z0-9]+PatchBaseline$`)
	var baselineIdentityIDs []string
	for paginator.HasMorePages() {
		out, err := paginator.NextPage(ctx)
		if err != nil {
			return "", fmt.Errorf("listing Patch Baselines for operating system %q: %s", os, err)
		}

		for _, identity := range out.BaselineIdentities {
			if id := aws.ToString(identity.BaselineName); re.MatchString(id) {
				baselineIdentityIDs = append(baselineIdentityIDs, aws.ToString(identity.BaselineId))
			}
		}
	}

	if l := len(baselineIdentityIDs); l == 0 {
		return "", tfresource.NewEmptyResultError(nil)
	} else if l > 1 {
		return "", tfresource.NewTooManyResultsError(l, nil)
	}

	return baselineIdentityIDs[0], nil
}
