// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/envvar"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// gatewayTargetRuntimeMCPPublicHostPattern matches the public AgentCore Runtime MCP URL
// (https://<runtimeId>.runtime.bedrock-agentcore.<region>.amazonaws.com/mcp). Gateway MCP targets with IAM
// outbound auth require the regional invocations URL instead; see normalizeGatewayTargetMCPIAMEndpointIfRuntimePublicHost.
var gatewayTargetRuntimeMCPPublicHostPattern = regexache.MustCompile(
	`^https://([A-Za-z0-9_-]+)\.runtime\.bedrock-agentcore\.([a-z0-9-]+)\.amazonaws\.com/mcp$`,
)

// normalizeGatewayTargetMCPIAMEndpointIfRuntimePublicHost converts a public-runtime MCP host URL to the
// bedrock-agentcore.<region>.amazonaws.com/runtimes/<url-encoded-arn>/invocations?qualifier=DEFAULT form that
// CreateGatewayTarget uses for implicit tool sync. Other URLs are returned unchanged.
func normalizeGatewayTargetMCPIAMEndpointIfRuntimePublicHost(ctx context.Context, t *testing.T, endpoint string) string {
	t.Helper()

	m := gatewayTargetRuntimeMCPPublicHostPattern.FindStringSubmatch(endpoint)
	if m == nil {
		return endpoint
	}

	runtimeID, region := m[1], m[2]
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		t.Fatalf("loading AWS config to normalize MCP endpoint: %s", err)
	}

	out, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, nil)
	if err != nil {
		t.Fatalf("calling sts:GetCallerIdentity to normalize MCP endpoint: %s", err)
	}

	accountID := aws.ToString(out.Account)
	if accountID == "" {
		t.Fatal("sts:GetCallerIdentity returned an empty account id")
	}

	partition := "aws"
	if a := aws.ToString(out.Arn); strings.HasPrefix(a, "arn:") {
		segs := strings.SplitN(a, ":", 3)
		if len(segs) > 1 && segs[1] != "" {
			partition = segs[1]
		}
	}

	rawARN := fmt.Sprintf("arn:%s:bedrock-agentcore:%s:%s:runtime/%s", partition, region, accountID, runtimeID)
	encoded := strings.ReplaceAll(strings.ReplaceAll(rawARN, ":", "%3A"), "/", "%2F")
	normalized := fmt.Sprintf("https://bedrock-agentcore.%s.amazonaws.com/runtimes/%s/invocations?qualifier=DEFAULT", region, encoded)

	t.Logf("normalized %s from public *.runtime.bedrock-agentcore host to invocations URL (required for gateway MCP + IAM tool sync)", envvar.BedrockAgentCoreGatewayTargetMCPIAMEndpoint)

	return normalized
}

// parseBedrockAgentCoreInvocationsRuntimeARN extracts the AgentCore Runtime ARN from a regional invocations URL
// (encoded full ARN in the path, or runtime id + accountId query parameter per AWS docs).
func parseBedrockAgentCoreInvocationsRuntimeARN(endpoint string) (region string, runtimeARN string, err error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", "", fmt.Errorf("parsing URL: %w", err)
	}

	host := u.Hostname()
	const pfx = "bedrock-agentcore."
	const sfx = ".amazonaws.com"
	if !strings.HasPrefix(host, pfx) || !strings.HasSuffix(host, sfx) {
		return "", "", fmt.Errorf("expected host bedrock-agentcore.<region>.amazonaws.com, got %q", host)
	}

	region = strings.TrimSuffix(strings.TrimPrefix(host, pfx), sfx)
	path := u.Path
	const marker = "/runtimes/"
	i := strings.Index(path, marker)
	if i < 0 {
		return "", "", fmt.Errorf("path %q missing %s", path, marker)
	}

	rest := path[i+len(marker):]
	j := strings.Index(rest, "/invocations")
	if j < 0 {
		return "", "", fmt.Errorf("path %q missing /invocations", path)
	}

	seg := rest[:j]
	decoded, err := url.PathUnescape(seg)
	if err != nil {
		return "", "", fmt.Errorf("path segment decode: %w", err)
	}

	if strings.HasPrefix(decoded, "arn:") {
		return region, decoded, nil
	}

	acct := u.Query().Get("accountId")
	if acct == "" {
		return "", "", fmt.Errorf("runtime path %q is not a full ARN; add accountId= to the invocations URL", decoded)
	}

	// Commercial default; short form is rare in this test.
	runtimeARN = fmt.Sprintf("arn:aws:bedrock-agentcore:%s:%s:runtime/%s", region, acct, decoded)
	return region, runtimeARN, nil
}

func iamRootPrincipalForBedrockRuntimeARN(runtimeARN string) (string, error) {
	parts := strings.SplitN(runtimeARN, ":", 6)
	if len(parts) < 6 || parts[0] != "arn" {
		return "", fmt.Errorf("invalid runtime ARN %q", runtimeARN)
	}

	partition, acct := parts[1], parts[4]
	return fmt.Sprintf("arn:%s:iam::%s:root", partition, acct), nil
}

// iamGatewayExecutionRoleARN returns the IAM ARN for a role name in the given partition and account (default path).
func iamGatewayExecutionRoleARN(partition, accountID, roleName string) string {
	return fmt.Sprintf("arn:%s:iam::%s:role/%s", partition, accountID, roleName)
}

func awsPartitionFromSTSIdentityARN(identityARN string) (partition string) {
	partition = "aws"
	if identityARN == "" || !strings.HasPrefix(identityARN, "arn:") {
		return partition
	}
	segs := strings.SplitN(identityARN, ":", 3)
	if len(segs) > 1 && segs[1] != "" {
		partition = segs[1]
	}
	return partition
}

// bedrockAgentCoreRuntimeIDFromARN returns the runtime identifier segment from
// arn:partition:bedrock-agentcore:region:account:runtime/<id>[/...].
func bedrockAgentCoreRuntimeIDFromARN(runtimeARN string) (string, error) {
	const marker = ":runtime/"
	i := strings.Index(runtimeARN, marker)
	if i < 0 {
		return "", fmt.Errorf("no %q in runtime ARN %q", marker, runtimeARN)
	}
	rest := runtimeARN[i+len(marker):]
	if j := strings.Index(rest, "/"); j >= 0 {
		rest = rest[:j]
	}
	if rest == "" {
		return "", fmt.Errorf("empty runtime id in %q", runtimeARN)
	}
	return rest, nil
}

// listAgentRuntimeEndpointResourceARNs returns AgentRuntimeEndpointArn values from ListAgentRuntimeEndpoints.
// Callers should fall back to a guessed .../endpoint/DEFAULT ARN if this returns an error or empty slice.
func listAgentRuntimeEndpointResourceARNs(ctx context.Context, conn *bedrockagentcorecontrol.Client, runtimeARN string) ([]string, error) {
	rid, err := bedrockAgentCoreRuntimeIDFromARN(runtimeARN)
	if err != nil {
		return nil, err
	}

	var (
		outARNs []string
		token   *string
	)
	for {
		out, err := conn.ListAgentRuntimeEndpoints(ctx, &bedrockagentcorecontrol.ListAgentRuntimeEndpointsInput{
			AgentRuntimeId: aws.String(rid),
			NextToken:      token,
		})
		if err != nil {
			return nil, err
		}
		for _, e := range out.RuntimeEndpoints {
			if e.AgentRuntimeEndpointArn != nil {
				outARNs = append(outARNs, aws.ToString(e.AgentRuntimeEndpointArn))
			}
		}
		if out.NextToken == nil || aws.ToString(out.NextToken) == "" {
			break
		}
		token = out.NextToken
	}

	return outARNs, nil
}

// invokeAllowResourcePolicyJSON builds a SigV4-oriented resource policy. AWS documents that for IAM-based
// outbound auth the Principal should include the gateway service role ARN, not only the account root; account
// root is included as a secondary allow. One statement per principal avoids Bedrock rejecting Principal.AWS as
// a JSON array in some cases.
func invokeAllowResourcePolicyJSON(resARN string, principalAWSARNs []string) ([]byte, error) {
	if len(principalAWSARNs) == 0 {
		return nil, fmt.Errorf("principalAWSARNs required")
	}
	statements := make([]any, 0, len(principalAWSARNs))
	for i, arn := range principalAWSARNs {
		statements = append(statements, map[string]any{
			"Sid":       fmt.Sprintf("TfAccAllowInvokeRuntimePrincipal%d", i),
			"Effect":    "Allow",
			"Principal": map[string]any{"AWS": arn},
			"Action": []string{
				"bedrock-agentcore:InvokeAgentRuntime",
				"bedrock-agentcore:InvokeAgentRuntimeForUser",
			},
			"Resource": resARN,
		})
	}
	return json.Marshal(map[string]any{
		"Version":   "2012-10-17",
		"Statement": statements,
	})
}

// maybeEnsureBedrockAgentCoreRuntimeInvokesForGatewayIAM updates (or creates) resource-based policies on the AgentCore
// Runtime and, when permitted, each agent runtime endpoint (from ListAgentRuntimeEndpoints) so IAM principals in the
// account can InvokeAgentRuntime. If listing endpoints fails, we fall back to the guessed .../endpoint/DEFAULT ARN.
// Some IAM policies allow PutResourcePolicy only on runtime ARNs, not endpoints; endpoint puts then log and skip.
//
// gatewayIAMRoleName must match aws_iam_role.test's name in testAccGatewayTargetConfig_infra (the gateway execution role).
func maybeEnsureBedrockAgentCoreRuntimeInvokesForGatewayIAM(ctx context.Context, t *testing.T, invocationsEndpoint, gatewayIAMRoleName string) {
	t.Helper()

	region, runtimeARN, err := parseBedrockAgentCoreInvocationsRuntimeARN(invocationsEndpoint)
	if err != nil {
		t.Logf("skipping AgentCore runtime resource policy setup: %s", err)
		return
	}

	force := os.Getenv(envvar.BedrockAgentCoreGatewayTargetMCPIAMEnsureRuntimeResourcePolicy) == "1"

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		t.Fatalf("loading AWS config for runtime resource policy: %s", err)
	}

	conn := bedrockagentcorecontrol.NewFromConfig(cfg)
	rootARN, err := iamRootPrincipalForBedrockRuntimeARN(runtimeARN)
	if err != nil {
		t.Fatalf("deriving IAM account root ARN: %s", err)
	}

	stsOut, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, nil)
	if err != nil {
		t.Fatalf("sts:GetCallerIdentity for gateway execution role ARN: %s", err)
	}
	acct := aws.ToString(stsOut.Account)
	if strings.TrimSpace(acct) == "" {
		t.Fatal("sts:GetCallerIdentity returned an empty account id")
	}
	partition := awsPartitionFromSTSIdentityARN(aws.ToString(stsOut.Arn))
	execRoleARN := iamGatewayExecutionRoleARN(partition, acct, gatewayIAMRoleName)
	principalARNs := []string{execRoleARN, rootARN}

	endpointARNs, listErr := listAgentRuntimeEndpointResourceARNs(ctx, conn, runtimeARN)
	if listErr != nil {
		t.Logf("warning: ListAgentRuntimeEndpoints: %s (falling back to guessed .../endpoint/DEFAULT for resource policy)", listErr)
	}
	if len(endpointARNs) == 0 {
		fallback := runtimeARN + "/endpoint/DEFAULT"
		t.Logf("using fallback endpoint resource ARN %s (ListAgentRuntimeEndpoints returned none or was skipped)", fallback)
		endpointARNs = []string{fallback}
	} else {
		t.Logf("discovered %d agent runtime endpoint ARN(s) for resource policy", len(endpointARNs))
	}

	resourceARNs := append([]string{runtimeARN}, endpointARNs...)

	var runtimeSkippedDueToExistingPolicy bool
	endpointPutDenied := false
	guessedDefaultARN := runtimeARN + "/endpoint/DEFAULT"
	var lastPolicyStr string

	for _, resARN := range resourceARNs {
		needPut := force
		if !needPut {
			gpo, gerr := conn.GetResourcePolicy(ctx, &bedrockagentcorecontrol.GetResourcePolicyInput{
				ResourceArn: aws.String(resARN),
			})
			switch {
			case errs.IsA[*awstypes.ResourceNotFoundException](gerr):
				needPut = true
			case gerr == nil && (gpo.Policy == nil || strings.TrimSpace(aws.ToString(gpo.Policy)) == ""):
				needPut = true
			case gerr != nil:
				t.Logf("warning: GetResourcePolicy(%s): %s", resARN, gerr)
				continue
			default:
				if resARN == runtimeARN {
					runtimeSkippedDueToExistingPolicy = true
				}
				continue
			}
		}

		if !needPut {
			continue
		}

		policyDoc, err := invokeAllowResourcePolicyJSON(resARN, principalARNs)
		if err != nil {
			t.Fatalf("marshaling resource policy: %s", err)
		}
		policyStr := string(policyDoc)
		lastPolicyStr = policyStr

		_, err = conn.PutResourcePolicy(ctx, &bedrockagentcorecontrol.PutResourcePolicyInput{
			ResourceArn: aws.String(resARN),
			Policy:      aws.String(policyStr),
		})
		if err != nil {
			isEndpoint := strings.Contains(resARN, "/endpoint/")
			switch {
			case isEndpoint &&
				(errs.IsA[*awstypes.ResourceNotFoundException](err) ||
					errs.IsA[*awstypes.ValidationException](err) ||
					errs.IsA[*awstypes.AccessDeniedException](err)):
				if errs.IsA[*awstypes.AccessDeniedException](err) {
					endpointPutDenied = true
				}
				t.Logf("skipping resource policy on %s: %s", resARN, err)
				continue
			case errs.IsA[*awstypes.AccessDeniedException](err) && resARN == runtimeARN:
				t.Logf("skipping PutResourcePolicy on runtime %s (access denied); grant bedrock-agentcore:PutResourcePolicy on this runtime or attach an InvokeAgentRuntime allow manually: %s", resARN, err)
				return
			default:
				t.Fatalf("PutResourcePolicy(%s): %s", resARN, err)
			}
		}

		t.Logf("PutResourcePolicy on %s to allow principals %v for InvokeAgentRuntime (IAM MCP acceptance)", resARN, principalARNs)
	}

	if runtimeSkippedDueToExistingPolicy && !force {
		t.Logf("runtime %s already has a resource policy; if CreateGatewayTarget fails with execution-role errors, allow InvokeAgentRuntime for your gateway execution role %s (and account root if needed) on the runtime and each endpoint, or set %s=1 to replace (destructive)",
			runtimeARN, execRoleARN, envvar.BedrockAgentCoreGatewayTargetMCPIAMEnsureRuntimeResourcePolicy)
	}

	if endpointPutDenied {
		primary := guessedDefaultARN
		if len(endpointARNs) > 0 {
			primary = endpointARNs[0]
		}
		remediationJSON := policyStrForEndpointRemediation(primary, principalARNs)
		if remediationJSON == "" && lastPolicyStr != "" {
			remediationJSON = lastPolicyStr
		}
		logAgentCoreRuntimeEndpointResourcePolicyRemediation(ctx, t, conn, region, endpointARNs, primary, execRoleARN, remediationJSON)
	}
}

// policyStrForEndpointRemediation returns the same JSON shape used for PutResourcePolicy on the endpoint ARN.
func policyStrForEndpointRemediation(endpointARN string, principalAWSARNs []string) string {
	b, err := invokeAllowResourcePolicyJSON(endpointARN, principalAWSARNs)
	if err != nil {
		return ""
	}
	return string(b)
}

func logAgentCoreRuntimeEndpointResourcePolicyRemediation(ctx context.Context, t *testing.T, conn *bedrockagentcorecontrol.Client, region string, endpointARNs []string, exampleEndpointARN, gatewayExecutionRoleARN, policyJSON string) {
	t.Helper()

	t.Logf("REMEDIATION: Implicit MCP tool sync requires resource policies on agent runtime endpoint(s), not only the runtime. Include the gateway execution role %s as Principal.AWS alongside the account root. Discovered endpoint ARN(s): %v. Your identity could not PutResourcePolicy on at least one (org IAM often scopes to ...:runtime/* and omits endpoint ARNs).", gatewayExecutionRoleARN, endpointARNs)
	t.Logf("Widen the tester SSO/role identity policy Resource list; see envvar comment on %s.", envvar.BedrockAgentCoreGatewayTargetMCPIAMEnsureRuntimeResourcePolicy)
	t.Logf("Grant bedrock-agentcore:ListAgentRuntimeEndpoints, GetResourcePolicy, and PutResourcePolicy on runtime and endpoint ARNs (or use an admin role). Apply the same InvokeAgentRuntime allow Resource-by-resource for each endpoint ARN.")

	if exampleEndpointARN != "" {
		if gpo, err := conn.GetResourcePolicy(ctx, &bedrockagentcorecontrol.GetResourcePolicyInput{
			ResourceArn: aws.String(exampleEndpointARN),
		}); err == nil && gpo.Policy != nil {
			t.Logf("GetResourcePolicy(%s): policy present (length %d chars)", exampleEndpointARN, len(strings.TrimSpace(aws.ToString(gpo.Policy))))
		} else if err != nil {
			t.Logf("GetResourcePolicy(%s): %s", exampleEndpointARN, err)
		}

		t.Logf("Example (save JSON to a file, then use file://): aws bedrock-agentcore-control put-resource-policy --region %s --resource-arn %s --policy file://endpoint-policy.json", region, exampleEndpointARN)
	}
	t.Logf("Policy document for one endpoint (Resource must match that endpoint's ARN exactly): %s", policyJSON)
}

// testAccEnsureBedrockAgentCoreRuntimePoliciesFunc returns a check that runs after Terraform has created
// aws_iam_role.test so PutResourcePolicy can include a valid gateway execution role ARN (Principal must refer to
// an existing IAM principal).
func testAccEnsureBedrockAgentCoreRuntimePoliciesFunc(ctx context.Context, t *testing.T, invocationsEndpoint, gatewayIAMRoleName string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		t.Helper()
		maybeEnsureBedrockAgentCoreRuntimeInvokesForGatewayIAM(ctx, t, invocationsEndpoint, gatewayIAMRoleName)
		return nil
	}
}

func TestBedrockAgentCoreRuntimeIDFromARN(t *testing.T) {
	t.Parallel()
	const arn = "arn:aws:bedrock-agentcore:us-east-1:123456789012:runtime/example-aBcDeF1234"
	id, err := bedrockAgentCoreRuntimeIDFromARN(arn)
	if err != nil {
		t.Fatal(err)
	}
	if id != "example-aBcDeF1234" {
		t.Fatalf("id: got %s", id)
	}
	withSuffix, err := bedrockAgentCoreRuntimeIDFromARN(arn + "/endpoint/DEFAULT")
	if err != nil {
		t.Fatal(err)
	}
	if withSuffix != "example-aBcDeF1234" {
		t.Fatalf("id with suffix: got %s", withSuffix)
	}
}

func TestPolicyStrForEndpointRemediation(t *testing.T) {
	t.Parallel()
	ep := "arn:aws:bedrock-agentcore:us-east-1:123456789012:runtime/example-aBcDeF1234/endpoint/DEFAULT"
	execRole := "arn:aws:iam::123456789012:role/tf-acc-gateway"
	root := "arn:aws:iam::123456789012:root"
	s := policyStrForEndpointRemediation(ep, []string{execRole, root})
	if !strings.Contains(s, ep) || !strings.Contains(s, execRole) || !strings.Contains(s, root) {
		t.Fatalf("policy JSON: %s", s)
	}
}

func TestParseBedrockAgentCoreInvocationsRuntimeARN(t *testing.T) {
	t.Parallel()

	const u = "https://bedrock-agentcore.us-east-1.amazonaws.com/runtimes/arn%3Aaws%3Abedrock-agentcore%3Aus-east-1%3A123456789012%3Aruntime%2Fexample-aBcDeF1234/invocations?qualifier=DEFAULT"
	region, runtimeARN, err := parseBedrockAgentCoreInvocationsRuntimeARN(u)
	if err != nil {
		t.Fatal(err)
	}
	if region != "us-east-1" {
		t.Fatalf("region: got %s", region)
	}
	const want = "arn:aws:bedrock-agentcore:us-east-1:123456789012:runtime/example-aBcDeF1234"
	if runtimeARN != want {
		t.Fatalf("runtime ARN: got %s want %s", runtimeARN, want)
	}
	root, err := iamRootPrincipalForBedrockRuntimeARN(runtimeARN)
	if err != nil {
		t.Fatal(err)
	}
	if root != "arn:aws:iam::123456789012:root" {
		t.Fatalf("iam root: got %s", root)
	}
}

func TestAccBedrockAgentCoreGatewayTarget_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayTarget bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayTargetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "gateway_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "target_id"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.gateway_iam_role.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.lambda.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "target_configuration.0.mcp.0.lambda.0.lambda_arn"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.lambda.0.tool_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.lambda.0.tool_schema.0.inline_payload.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "gateway_identifier", "target_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "target_id",
			},
		},
	})
}

func TestAccBedrockAgentCoreGatewayTarget_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayTarget bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayTargetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceGatewayTarget, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreGatewayTarget_targetConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayTarget, gatewayTargetPrev bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_primitive()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.lambda.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.lambda.0.tool_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.lambda.0.tool_schema.0.inline_payload.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			// Example 2: Object with properties + required
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_objectWithProperties()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTargetPrev),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Example 3: Array of primitives
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_arrayOfPrimitives()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Example 4: Array of objects
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_arrayOfObjects()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Example 5: Array of arrays
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_arrayOfArrays()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			//Example 6: Mixed nested object/array
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_mixedNested()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTargetPrev),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Example 7: Array with ignored keywords
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_arrayWithIgnoredKeywords()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Invalid Example 8: Both items and properties at the same node
			{
				Config:      testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_invalidBothItemsAndProperties()),
				ExpectError: regexache.MustCompile("Invalid Attribute Combination"),
			},
			// Invalid Example 9: Missing type
			{
				Config:      testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_invalidMissingType()),
				ExpectError: regexache.MustCompile("Missing required argument"),
			},
			// Invalid Example 10: Unsupported type
			{
				Config:      testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_invalidUnsupportedType()),
				ExpectError: regexache.MustCompile("Invalid String Enum Value"),
			},
			// Return to valid configuration to proceed with post-test destroy
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_objectWithProperties()),
			},
		},
	})
}

func TestAccBedrockAgentCoreGatewayTarget_targetConfigurationMCPServer(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayTarget bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayTargetConfig_targetConfigurationMCPServer(rName, "https://knowledge-mcp.global.api.aws"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.mcp_server.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.mcp_server.0.endpoint", "https://knowledge-mcp.global.api.aws"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccGatewayTargetConfig_targetConfigurationMCPServer(rName, "https://docs.mcp.cloudflare.com/mcp"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.mcp_server.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.mcp_server.0.endpoint", "https://docs.mcp.cloudflare.com/mcp"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreGatewayTarget_targetConfigurationAPIGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayTarget bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayTargetConfig_targetConfigurationAPIGateway(rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "target_configuration.0.mcp.0.api_gateway.0.rest_api_id", "aws_api_gateway_rest_api.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "target_configuration.0.mcp.0.api_gateway.0.stage", "aws_api_gateway_stage.test", "stage_name"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_filter.0.filter_path", "/pets"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_filter.0.methods.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_filter.0.methods.*", "GET"),
					resource.TestCheckTypeSetElemAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_filter.0.methods.*", "POST"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_override.*", map[string]string{
						names.AttrName:        "ListPets",
						names.AttrPath:        "/pets",
						"method":              "GET",
						names.AttrDescription: "Retrieves all available pets",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_override.*", map[string]string{
						names.AttrName:        "RegisterPets",
						names.AttrPath:        "/pets",
						"method":              "POST",
						names.AttrDescription: "Register pets",
					}),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "gateway_identifier", "target_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "target_id",
			},
			{
				Config: testAccGatewayTargetConfig_targetConfigurationAPIGateway(rName, "2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "target_configuration.0.mcp.0.api_gateway.0.rest_api_id", "aws_api_gateway_rest_api.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "target_configuration.0.mcp.0.api_gateway.0.stage", "aws_api_gateway_stage.test", "stage_name"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_filter.0.filter_path", "/pets"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_filter.0.methods.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_filter.0.methods.*", "GET"),
					resource.TestCheckTypeSetElemAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_filter.0.methods.*", "POST"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_override.*", map[string]string{
						names.AttrName:        "ListPets2",
						names.AttrPath:        "/pets",
						"method":              "GET",
						names.AttrDescription: "Retrieves all available pets2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_override.*", map[string]string{
						names.AttrName:        "RegisterPets2",
						names.AttrPath:        "/pets",
						"method":              "POST",
						names.AttrDescription: "Register pets2",
					}),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreGatewayTarget_credentialProvider(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayTarget, gatewayTargetPrev bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Gateway IAM Role provider with Lambda target
			{
				Config: testAccGatewayTargetConfig_credentialProvider(rName, testAccCredentialProvider_gatewayIAMRole()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.gateway_iam_role.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.api_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.oauth.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			// Step 2: API Key provider with OpenAPI Schema target (creates new resource)
			{
				Config: testAccGatewayTargetConfig_credentialProviderNonLambda(rName, testAccCredentialProvider_apiKey()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTargetPrev),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.api_key.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.gateway_iam_role.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.oauth.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "credential_provider_configuration.0.api_key.0.provider_arn"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			// Step 3: OAuth provider with OpenAPI Schema target (updates credential provider only)
			{
				Config: testAccGatewayTargetConfig_credentialProviderNonLambda(rName, testAccCredentialProvider_oauth()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.oauth.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.api_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.gateway_iam_role.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "credential_provider_configuration.0.oauth.0.provider_arn"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Step 4: Gateway IAM Role provider with Smithy Model target (creates new resource due to both changes)
			{
				Config: testAccGatewayTargetConfig_credentialProviderSmithy(rName, testAccCredentialProvider_gatewayIAMRole()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTargetPrev),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.gateway_iam_role.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.api_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.oauth.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			// Step 5: Back to Gateway IAM Role with Lambda target (creates new resource again)
			{
				Config: testAccGatewayTargetConfig_credentialProvider(rName, testAccCredentialProvider_gatewayIAMRole()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTargetPrev),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.gateway_iam_role.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.api_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.oauth.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "gateway_identifier", "target_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "target_id",
			},
		},
	})
}

// TestAccBedrockAgentCoreGatewayTarget_gatewayIAMRoleServiceRegionMCPServer creates an MCP server target with
// gateway_iam_role { service = "bedrock-agentcore", region = data.aws_region.current } when
// TF_ACC_BEDROCK_AGENTCORE_GATEWAY_TARGET_MCP_IAM_ENDPOINT is set to the HTTPS **invocations** URL for the Agent Core
// Runtime MCP endpoint (same account/region as the test), URL-encoded runtime ARN in the path, for example:
//
//	https://bedrock-agentcore.us-east-1.amazonaws.com/runtimes/arn%3Aaws%3Abedrock-agentcore%3Aus-east-1%3A123456789012%3Aruntime%2Fexample-aBcDeF1234/invocations?qualifier=DEFAULT
//
// The public host https://<runtime-id>.runtime.bedrock-agentcore.<region>.amazonaws.com/mcp is accepted for
// convenience: the test normalizes it to the invocations URL using sts:GetCallerIdentity (same credentials/region).
// Prefer setting the invocations URL directly. The public host alone typically fails implicit tool sync with:
// "Gateway service is not authorized to assume the execution role" even when the gateway role trust policy is valid.
// If unset, the test is skipped (including in default CI). SigV4 MCP to AgentCore Runtime also requires resource-based
// policies on the **runtime** and on each **agent runtime endpoint** (ARNs from ListAgentRuntimeEndpoints; not always
// .../endpoint/DEFAULT) so InvokeAgentRuntime is allowed; otherwise CreateGatewayTarget can fail with a misleading
// execution-role assume error. The test uses two apply steps: first testAccGatewayTargetConfig_infra creates
// aws_iam_role.test so the execution role ARN exists, then a check calls PutResourcePolicy (gateway role ARN + account root).
// A second apply adds the gateway target. PutResourcePolicy before the role exists returns ValidationException: Invalid principal.
// If a policy already exists on the runtime, set
// TF_ACC_BEDROCK_AGENTCORE_GATEWAY_TARGET_MCP_IAM_ENSURE_RUNTIME_POLICY=1 to replace it (destructive) or edit the policy per
// https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/security.html#resource-based-policies
func TestAccBedrockAgentCoreGatewayTarget_gatewayIAMRoleServiceRegionMCPServer(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayTarget bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"
	endpoint := envvar.SkipIfEmpty(t, envvar.BedrockAgentCoreGatewayTargetMCPIAMEndpoint,
		"set to the bedrock-agentcore.<region>.amazonaws.com/runtimes/.../invocations?qualifier=... MCP URL (see test godoc); or the *.runtime.bedrock-agentcore.../mcp host (auto-normalized)")
	endpoint = normalizeGatewayTargetMCPIAMEndpointIfRuntimePublicHost(ctx, t, endpoint)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayTargetConfig_infra(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccEnsureBedrockAgentCoreRuntimePoliciesFunc(ctx, t, endpoint, rName),
				),
			},
			{
				Config: testAccGatewayTargetConfig_gatewayIAMRoleMCPServer(rName, endpoint),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.gateway_iam_role.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.gateway_iam_role.0.service", "bedrock-agentcore"),
					resource.TestCheckResourceAttrPair(resourceName, "credential_provider_configuration.0.gateway_iam_role.0.region", "data.aws_region.current", "name"),
				),
			},
		},
	})
}

func TestAccBedrockAgentCoreGatewayTarget_credentialProvider_invalid(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Invalid: Multiple credential providers
			{
				Config:      testAccGatewayTargetConfig_credentialProvider(rName, testAccCredentialProvider_multipleProviders()),
				ExpectError: regexache.MustCompile(`Invalid Attribute Combination|cannot be specified`),
			},
			{
				Config:      testAccGatewayTargetConfig_credentialProvider(rName, testAccCredentialProvider_empty()),
				ExpectError: regexache.MustCompile("Invalid Credential Provider Configuration|At least one credential provider must be configured"),
			},
		},
	})
}

func TestAccBedrockAgentCoreGatewayTarget_metadataConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayTarget bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayTargetConfig_metadataConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "metadata_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metadata_configuration.0.allowed_request_headers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "metadata_configuration.0.allowed_response_headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metadata_configuration.0.allowed_query_parameters.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "gateway_identifier", "target_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "target_id",
			},
			// Update metadata configuration
			{
				Config: testAccGatewayTargetConfig_metadataConfigurationUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, "metadata_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metadata_configuration.0.allowed_request_headers.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "metadata_configuration.0.allowed_response_headers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "metadata_configuration.0.allowed_query_parameters.#", "2"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Remove metadata configuration
			{
				Config: testAccGatewayTargetConfig_metadataConfigurationRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, "metadata_configuration.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreGatewayTarget_metadataConfiguration_invalidHeaders(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Invalid: restricted header Authorization
			{
				Config:      testAccGatewayTargetConfig_metadataConfigurationInvalidHeader(rName, "Authorization"),
				ExpectError: regexache.MustCompile(`none of \(case-insensitive\)`),
			},
			// Invalid: restricted header Content-Type
			{
				Config:      testAccGatewayTargetConfig_metadataConfigurationInvalidHeader(rName, "Content-Type"),
				ExpectError: regexache.MustCompile(`none of \(case-insensitive\)`),
			},
			// Invalid: restricted header Host
			{
				Config:      testAccGatewayTargetConfig_metadataConfigurationInvalidHeader(rName, "Host"),
				ExpectError: regexache.MustCompile(`none of \(case-insensitive\)`),
			},
			// Invalid: X-Amzn- prefix
			{
				Config:      testAccGatewayTargetConfig_metadataConfigurationInvalidHeader(rName, "X-Amzn-Custom"),
				ExpectError: regexache.MustCompile(`must not begin with \(case-insensitive\)`),
			},
			// Invalid: header with special characters
			{
				Config:      testAccGatewayTargetConfig_metadataConfigurationInvalidHeader(rName, "Invalid Header!"),
				ExpectError: regexache.MustCompile(`header names must contain only alphanumeric characters`),
			},
			// Valid: X-Amzn-Bedrock-AgentCore-Runtime-Custom- prefix is allowed
			{
				Config: testAccGatewayTargetConfig_metadataConfigurationInvalidHeader(rName, "X-Amzn-Bedrock-AgentCore-Runtime-Custom-MyHeader"),
			},
		},
	})
}

func testAccCheckGatewayTargetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_gateway_target" {
				continue
			}

			_, err := tfbedrockagentcore.FindGatewayTargetByTwoPartKey(ctx, conn, rs.Primary.Attributes["gateway_identifier"], rs.Primary.Attributes["target_id"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core Gateway Target %s still exists", rs.Primary.Attributes["target_id"])
		}

		return nil
	}
}

func testAccCheckGatewayTargetExists(ctx context.Context, t *testing.T, n string, v *bedrockagentcorecontrol.GetGatewayTargetOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindGatewayTargetByTwoPartKey(ctx, conn, rs.Primary.Attributes["gateway_identifier"], rs.Primary.Attributes["target_id"])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccGatewayTargetConfig_infra(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

data "aws_region" "infra" {}

# Match production gateway execution role trust: both Bedrock and AgentCore service
# principals may assume this role when the gateway connects to targets (including
# MCP server targets on AgentCore Runtime with IAM / SigV4). See
# https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/gateway-prerequisites-permissions.html
#
# Do not add aws:SourceAccount / aws:SourceArn conditions here: some gateway → MCP
# flows omit those STS context keys, and StringEquals then fails the trust policy
# with "Gateway service is not authorized to assume the execution role".
# Use one statement per service principal (matches AWS doc examples).
data "aws_iam_policy_document" "test" {
  statement {
    sid     = "AllowBedrockAgentCoreAssumeGatewayExecutionRole"
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }

  statement {
    sid     = "AllowBedrockAssumeGatewayExecutionRole"
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "lambda:*"
        Resource = "*"
      },
      {
        Sid    = "AgentCoreGatewayRuntimeMCP"
        Effect = "Allow"
        Action = [
          "bedrock-agentcore:InvokeAgentRuntime",
          "bedrock-agentcore:InvokeAgentRuntimeForUser",
          "bedrock-agentcore:InvokeGateway",
        ]
        Resource = "*"
      },
      {
        Sid    = "AgentCoreWorkloadIdentityAndOutboundAuth"
        Effect = "Allow"
        Action = [
          "bedrock-agentcore:GetWorkloadAccessToken",
          "bedrock-agentcore:GetResourceOauth2Token",
          "bedrock-agentcore:GetResourceApiKey",
        ]
        Resource = [
          "arn:${data.aws_partition.current.partition}:bedrock-agentcore:${data.aws_region.infra.id}:${data.aws_caller_identity.current.account_id}:workload-identity-directory/default",
          "arn:${data.aws_partition.current.partition}:bedrock-agentcore:${data.aws_region.infra.id}:${data.aws_caller_identity.current.account_id}:workload-identity-directory/default/workload-identity/*",
          "arn:${data.aws_partition.current.partition}:bedrock-agentcore:${data.aws_region.infra.id}:${data.aws_caller_identity.current.account_id}:token-vault/*",
        ]
      },
    ]
  })
}

data "aws_iam_policy_document" "lambda_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "lambda" {
  name = "%[1]s-lambda"

  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.lambda.arn
  handler       = "lambdatest.handler"
  runtime       = "nodejs20.x"
}

resource "aws_bedrockagentcore_gateway" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  authorizer_type = "CUSTOM_JWT"
  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test"]
    }
  }

  protocol_configuration {
    mcp {
      instructions       = "Do something"
      search_type        = "SEMANTIC"
      supported_versions = ["2025-11-25"]
    }
  }

  protocol_type = "MCP"
}
`, rName)
}

func testAccGatewayTargetConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  credential_provider_configuration {
    gateway_iam_role {}
  }

  target_configuration {
    mcp {
      lambda {
        lambda_arn = aws_lambda_function.test.arn

        tool_schema {
          inline_payload {
            name        = "test_tool"
            description = "A test tool"

            input_schema {
              type = "object"

              property {
                name        = "input"
                description = "some input"
                type        = "string"
                required    = true
              }
            }
          }
        }
      }
    }
  }
}

`, rName))
}

func testAccGatewayTargetConfig_credentialProvider(rName, credentialProviderContent string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  credential_provider_configuration {
%[2]s
  }

  target_configuration {
    mcp {
      lambda {
        lambda_arn = aws_lambda_function.test.arn

        tool_schema {
          inline_payload {
            name        = "test_tool"
            description = "A test tool"

            input_schema {
              type        = "string"
              description = "Basic schema for credential provider test"
            }
          }
        }
      }
    }
  }
}
`, rName, credentialProviderContent))
}

func testAccGatewayTargetConfig_gatewayIAMRoleMCPServer(rName, endpoint string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  credential_provider_configuration {
    gateway_iam_role {
      service = "bedrock-agentcore"
      region  = data.aws_region.current.name
    }
  }

  target_configuration {
    mcp {
      mcp_server {
        endpoint = %[2]q
      }
    }
  }
}
`, rName, endpoint))
}

func testAccGatewayTargetConfig_credentialProviderNonLambda(rName, credentialProviderContent string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  credential_provider_configuration {
%[2]s
  }

  target_configuration {
    mcp {
      open_api_schema {
        inline_payload {
          payload = jsonencode({
            openapi = "3.0.0"
            info = {
              title   = "Test API"
              version = "1.0.0"
            }
            servers = [
              {
                url = "https://api.example.com"
              }
            ]
            paths = {
              "/test" = {
                get = {
                  operationId = "getTest"
                  summary     = "Test endpoint"
                  responses = {
                    "200" = {
                      description = "Success"
                    }
                  }
                }
              }
            }
          })
        }
      }
    }
  }
}
`, rName, credentialProviderContent))
}

func testAccGatewayTargetConfig_credentialProviderSmithy(rName, credentialProviderContent string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  credential_provider_configuration {
%[2]s
  }

  target_configuration {
    mcp {
      smithy_model {
        inline_payload {
          payload = jsonencode({
            "smithy" = "2.0"
            "shapes" = {
              "com.example#TestService" = {
                "type"    = "service"
                "version" = "1.0"
                "operations" = [
                  {
                    "target" = "com.example#TestOperation"
                  }
                ]
                "traits" = {
                  "aws.auth#sigv4" = {
                    "name" = "testservice"
                  }
                  "aws.protocols#restJson1" = {}
                }
              }
              "com.example#TestOperation" = {
                "type" = "operation"
                "input" = {
                  "target" = "com.example#TestInput"
                }
                "output" = {
                  "target" = "com.example#TestOutput"
                }
                "traits" = {
                  "smithy.api#http" = {
                    "method" = "POST"
                    "uri"    = "/test"
                  }
                }
              }
              "com.example#TestInput" = {
                "type" = "structure"
                "members" = {
                  "message" = {
                    "target" = "smithy.api#String"
                    "traits" = {
                      "smithy.api#required" = {}
                    }
                  }
                }
              }
              "com.example#TestOutput" = {
                "type" = "structure"
                "members" = {
                  "result" = {
                    "target" = "smithy.api#String"
                  }
                }
              }
            }
          })
        }
      }
    }
  }
}
`, rName, credentialProviderContent))
}

func testAccGatewayTargetConfig_targetConfiguration(rName, schemaContent string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  credential_provider_configuration {
    gateway_iam_role {}
  }

  target_configuration {
    mcp {
      lambda {
        lambda_arn = aws_lambda_function.test.arn

        tool_schema {
          inline_payload {
            name        = "test_tool"
            description = "A test tool"

            input_schema {
              %[2]s
            }
          }
        }
      }
    }
  }
}
`, rName, schemaContent))
}

func testAccGatewayTargetConfig_targetConfigurationMCPServer(rName, endpoint string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  target_configuration {
    mcp {
      mcp_server {
        endpoint = %[2]q
      }
    }
  }
}
`, rName, endpoint))
}

func testAccSchema_primitive() string {
	return `
			type        = "string"
			description = "A token"
 		   `
}

func testAccSchema_objectWithProperties() string {
	return `
			type        = "object"
			description = "User"

			property {
				name     = "id"
				type     = "string"
				required = true
			}

			property {
				name = "age"
				type = "integer"
			}

			property {
				name = "paid"
				type = "boolean"
			}
		 `
}

func testAccSchema_arrayOfPrimitives() string {
	return `
			type        = "array"
			description = "Tags"

			items {
				type = "string"
			}
		 `
}

func testAccSchema_arrayOfObjects() string {
	return `
			type = "array"

			items {
				type = "object"

				property {
					name     = "id"
					type     = "string"
					required = true
				}

				property {
					name = "email"
					type = "string"
				}

				property {
					name = "age"
					type = "integer"
				}
			}
		 `
}

func testAccSchema_arrayOfArrays() string {
	return `
			type = "array"

			items {
				type = "array"

				items {
					type = "number"
				}
			}
		 `
}

func testAccSchema_mixedNested() string {
	return `
			type = "object"

			property {
				name = "profile"
				type = "object"

				property {
					name       = "nested_tags"
					type       = "array"
					items_json = jsonencode({
						type = "string"
					})
				}
			}
		 `
}

func testAccSchema_arrayWithIgnoredKeywords() string {
	return `
			type = "array"

			items {
				type = "string"
			}
		 `
}

func testAccSchema_invalidBothItemsAndProperties() string {
	return `
			type = "object"

			items {
				type = "string"
			}

			property {
				name = "a"
				type = "string"
			}
		 `
}

func testAccSchema_invalidMissingType() string {
	return `
			description = "No type here"
		 `
}

func testAccSchema_invalidUnsupportedType() string {
	return `
			type = "date"
		 `
}

func testAccCredentialProvider_gatewayIAMRole() string {
	return `    gateway_iam_role {}`
}

func testAccCredentialProvider_apiKey() string {
	return `    api_key {
      provider_arn              = "arn:${data.aws_partition.current.partition}:iam::123456789012:oidc-provider/example.com"
      credential_location       = "HEADER"
      credential_parameter_name = "X-API-Key"
      credential_prefix         = "Bearer"
    }`
}

func testAccCredentialProvider_oauth() string {
	return `    oauth {
      provider_arn       = "arn:${data.aws_partition.current.partition}:iam::123456789012:oidc-provider/oauth.example.com"
      scopes             = ["read", "write"]
	  grant_type         = "AUTHORIZATION_CODE"
	  default_return_url = "https://example.com/callback"

      custom_parameters = {
        "client_type" = "confidential"
      }
    }`
}

func testAccCredentialProvider_multipleProviders() string {
	return `    gateway_iam_role {}
    api_key {
      provider_arn = "arn:${data.aws_partition.current.partition}:iam::123456789012:oidc-provider/example.com"
    }`
}

func testAccCredentialProvider_empty() string {
	return `    # No providers configured`
}

func testAccGatewayTargetConfig_metadataConfiguration(rName string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  target_configuration {
    mcp {
      mcp_server {
        endpoint = "https://docs.mcp.cloudflare.com/mcp"
      }
    }
  }

  metadata_configuration {
    allowed_request_headers  = ["x-correlation-id", "x-tenant-id"]
    allowed_response_headers = ["x-rate-limit-remaining"]
    allowed_query_parameters = ["version"]
  }
}
`, rName))
}

func testAccGatewayTargetConfig_metadataConfigurationUpdated(rName string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  target_configuration {
    mcp {
      mcp_server {
        endpoint = "https://docs.mcp.cloudflare.com/mcp"
      }
    }
  }

  metadata_configuration {
    allowed_request_headers  = ["x-correlation-id", "x-tenant-id", "x-request-id"]
    allowed_response_headers = ["x-rate-limit-remaining", "x-request-id"]
    allowed_query_parameters = ["version", "format"]
  }
}
`, rName))
}

func testAccGatewayTargetConfig_metadataConfigurationRemoved(rName string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  target_configuration {
    mcp {
      mcp_server {
        endpoint = "https://docs.mcp.cloudflare.com/mcp"
      }
    }
  }
}
`, rName))
}

func testAccGatewayTargetConfig_metadataConfigurationInvalidHeader(rName, headerName string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  target_configuration {
    mcp {
      mcp_server {
        endpoint = "https://docs.mcp.cloudflare.com/mcp"
      }
    }
  }

  metadata_configuration {
    allowed_request_headers = [%[2]q]
  }
}
`, rName, headerName))
}

func testAccGatewayTargetConfig_targetConfigurationAPIGateway(rName, toolOverrideSuffix string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "pets" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "pets"
}

resource "aws_api_gateway_method" "get_pets" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.pets.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_method" "post_pets" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.pets.id
  http_method   = "POST"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "get_pets_200" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.pets.id
  http_method = aws_api_gateway_method.get_pets.http_method
  status_code = "200"
}

resource "aws_api_gateway_method_response" "post_pets_200" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.pets.id
  http_method = aws_api_gateway_method.post_pets.http_method
  status_code = "200"
}

resource "aws_api_gateway_integration" "get_pets" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.pets.id
  http_method = aws_api_gateway_method.get_pets.http_method
  type        = "MOCK"
}

resource "aws_api_gateway_integration" "post_pets" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.pets.id
  http_method = aws_api_gateway_method.post_pets.http_method
  type        = "MOCK"
}

resource "aws_api_gateway_deployment" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  depends_on = [
    aws_api_gateway_integration.get_pets,
    aws_api_gateway_integration.post_pets,
  ]
}

resource "aws_api_gateway_stage" "test" {
  stage_name    = "prod"
  rest_api_id   = aws_api_gateway_rest_api.test.id
  deployment_id = aws_api_gateway_deployment.test.id
}

resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  target_configuration {
    mcp {
      api_gateway {
        rest_api_id = aws_api_gateway_rest_api.test.id
        stage       = aws_api_gateway_stage.test.stage_name

        api_gateway_tool_configuration {
          tool_filter {
            filter_path = "/pets"
            methods     = ["GET", "POST"]
          }

          tool_override {
            name        = "ListPets%[2]s"
            path        = "/pets"
            method      = "GET"
            description = "Retrieves all available pets%[2]s"
          }

          tool_override {
            name        = "RegisterPets%[2]s"
            path        = "/pets"
            method      = "POST"
            description = "Register pets%[2]s"
          }
        }
      }
    }
  }
}
`, rName, toolOverrideSuffix))
}
