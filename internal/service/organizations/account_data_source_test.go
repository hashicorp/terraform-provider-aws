package organizations_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	intOrg "github.com/hashicorp/terraform-provider-aws/internal/organizations"
	providerOrg "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
)

// mockOrganizationsClient is a minimal mock that satisfies the interface used by DataSourceAccount.
type mockOrganizationsClient struct {
	describeAccountFn      func(ctx context.Context, params *organizations.DescribeAccountInput, optFns ...func(*organizations.Options)) (*organizations.DescribeAccountOutput, error)
	listParentsFn          func(ctx context.Context, params *organizations.ListParentsInput, optFns ...func(*organizations.Options)) (*organizations.ListParentsOutput, error)
	describeOrganizationFn func(ctx context.Context, params *organizations.DescribeOrganizationInput, optFns ...func(*organizations.Options)) (*organizations.DescribeOrganizationOutput, error)
}

func (m *mockOrganizationsClient) DescribeAccount(ctx context.Context, params *organizations.DescribeAccountInput, optFns ...func(*organizations.Options)) (*organizations.DescribeAccountOutput, error) {
	return m.describeAccountFn(ctx, params, optFns...)
}
func (m *mockOrganizationsClient) ListParents(ctx context.Context, params *organizations.ListParentsInput, optFns ...func(*organizations.Options)) (*organizations.ListParentsOutput, error) {
	return m.listParentsFn(ctx, params, optFns...)
}
func (m *mockOrganizationsClient) DescribeOrganization(ctx context.Context, params *organizations.DescribeOrganizationInput, optFns ...func(*organizations.Options)) (*organizations.DescribeOrganizationOutput, error) {
	return m.describeOrganizationFn(ctx, params, optFns...)
}

// mockAWSClient to simulate meta.(*conns.AWSClient)
type mockAWSClient struct {
	client *mockOrganizationsClient
}

func (m *mockAWSClient) OrganizationsClient(ctx context.Context) *mockOrganizationsClient {
	return m.client
}

func TestDataSourceAccount_basic(t *testing.T) {
	t.Parallel()

	// Fake AWS client behavior
	mockClient := &mockOrganizationsClient{
		describeAccountFn: func(ctx context.Context, params *organizations.DescribeAccountInput, optFns ...func(*organizations.Options)) (*organizations.DescribeAccountOutput, error) {
			require.Equal(t, "123456789012", aws.ToString(params.AccountId))
			return &organizations.DescribeAccountOutput{
				Account: &types.Account{
					Id:              aws.String("123456789012"),
					Arn:             aws.String("arn:aws:organizations::123456789012:account/o-abc123/r-root/123456789012"),
					Email:           aws.String("test@example.com"),
					Name:            aws.String("test-account"),
					State:           types.AccountStateActive,
					JoinedMethod:    types.AccountJoinedMethodInvited,
					JoinedTimestamp: aws.Time(time.Date(2020, 12, 1, 12, 0, 0, 0, time.UTC)),
				},
			}, nil
		},
		listParentsFn: func(ctx context.Context, params *organizations.ListParentsInput, optFns ...func(*organizations.Options)) (*organizations.ListParentsOutput, error) {
			if aws.ToString(params.ChildId) == "123456789012" {
				return &organizations.ListParentsOutput{
					Parents: []types.Parent{{Id: aws.String("r-root"), Type: types.ParentTypeRoot}},
				}, nil
			}
			return nil, nil
		},
		describeOrganizationFn: func(ctx context.Context, params *organizations.DescribeOrganizationInput, optFns ...func(*organizations.Options)) (*organizations.DescribeOrganizationOutput, error) {
			return &organizations.DescribeOrganizationOutput{
				Organization: &types.Organization{Id: aws.String("o-abc123")},
			}, nil
		},
	}

	meta := &mockAWSClient{client: mockClient}
	dataSource := providerOrg.DataSourceAccount()

	// Build input
	d := schema.TestResourceDataRaw(t, dataSource.Schema, map[string]interface{}{
		"id": "123456789012",
	})

	// Run Read function
	diags := dataSource.ReadContext(context.Background(), d, meta)
	require.Empty(t, diags, "unexpected diagnostics: %v", diags)

	// Assertions
	assert.Equal(t, "123456789012", d.Id())
	assert.Equal(t, "123456789012", d.Get("id"))
	assert.Equal(t, "arn:aws:organizations::123456789012:account/o-abc123/r-root/123456789012", d.Get("arn"))
	assert.Equal(t, "test@example.com", d.Get("email"))
	assert.Equal(t, "test-account", d.Get("name"))
	assert.Equal(t, "ACTIVE", d.Get("state"))
	assert.Equal(t, "INVITED", d.Get("joined_method"))
	assert.Contains(t, d.Get("principal_org_path"), "o-abc123/r-root/123456789012/")
}

func TestDataSourceAccount_errorDescribeAccount(t *testing.T) {
	t.Parallel()

	mockClient := &mockOrganizationsClient{
		describeAccountFn: func(ctx context.Context, params *organizations.DescribeAccountInput, optFns ...func(*organizations.Options)) (*organizations.DescribeAccountOutput, error) {
			return nil, assert.AnError
		},
	}

	meta := &mockAWSClient{client: mockClient}
	dataSource := providerOrg.DataSourceAccount()

	d := schema.TestResourceDataRaw(t, dataSource.Schema, map[string]interface{}{
		"id": "123456789012",
	})

	diags := dataSource.ReadContext(context.Background(), d, meta)
	require.NotEmpty(t, diags)
	assert.Contains(t, diags[0].Summary, "reading account 123456789012")
}

func TestDataSourceAccount_errorPrincipalOrgPath(t *testing.T) {
	t.Parallel()

	mockClient := &mockOrganizationsClient{
		describeAccountFn: func(ctx context.Context, params *organizations.DescribeAccountInput, optFns ...func(*organizations.Options)) (*organizations.DescribeAccountOutput, error) {
			return &organizations.DescribeAccountOutput{
				Account: &types.Account{
					Id: aws.String("123456789012"),
				},
			}, nil
		},
		listParentsFn: func(ctx context.Context, params *organizations.ListParentsInput, optFns ...func(*organizations.Options)) (*organizations.ListParentsOutput, error) {
			return nil, assert.AnError
		},
	}

	meta := &mockAWSClient{client: mockClient}
	dataSource := providerOrg.DataSourceAccount()

	d := schema.TestResourceDataRaw(t, dataSource.Schema, map[string]interface{}{
		"id": "123456789012",
	})

	diags := dataSource.ReadContext(context.Background(), d, meta)
	require.NotEmpty(t, diags)
	assert.Contains(t, diags[0].Summary, "building principal org path")
}
