// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"slices"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsretry "github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestDBNodeActionTransitions(t *testing.T) {
	t.Parallel()

	for name, testCase := range map[string]struct {
		operation      string
		initial        string
		dispatchStatus string
		polls          []string
	}{
		"start":                          {operation: "start", initial: "STOPPED", dispatchStatus: "STARTING", polls: []string{"STOPPED", "STARTING", "AVAILABLE"}},
		"stop":                           {operation: "stop", initial: "AVAILABLE", dispatchStatus: "STOPPING", polls: []string{"AVAILABLE", "STOPPING", "STOPPED"}},
		"start already available":        {operation: "start", initial: "AVAILABLE"},
		"stop already stopped":           {operation: "stop", initial: "STOPPED"},
		"reboot response transition":     {operation: "reboot", initial: "AVAILABLE", dispatchStatus: "UPDATING", polls: []string{"AVAILABLE", "AVAILABLE"}},
		"reboot stale reads":             {operation: "reboot", initial: "AVAILABLE", dispatchStatus: "AVAILABLE", polls: []string{"AVAILABLE", "AVAILABLE", "UPDATING", "AVAILABLE", "AVAILABLE"}},
		"reboot full lifecycle":          {operation: "reboot", initial: "AVAILABLE", polls: []string{"STOPPING", "STOPPED", "STARTING", "AVAILABLE", "AVAILABLE"}},
		"reboot interrupted final state": {operation: "reboot", initial: "AVAILABLE", dispatchStatus: "STOPPING", polls: []string{"AVAILABLE", "UPDATING", "AVAILABLE", "AVAILABLE"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			steps := []dbNodeActionTestStep{{target: "GetDbNode", body: dbNodeActionTestNode(testCase.initial)}}
			if testCase.polls != nil {
				steps = append(steps, dbNodeActionTestStep{
					target: dbNodeActionTestMutation(testCase.operation),
					body:   fmt.Sprintf(`{"dbNodeId":"dbnode-test","status":%q}`, testCase.dispatchStatus),
				})
				for _, status := range testCase.polls {
					steps = append(steps, dbNodeActionTestStep{target: "GetDbNode", body: dbNodeActionTestNode(status)})
				}
			}
			conn := dbNodeActionTestClient(t, steps)
			var progress []string
			err := runDBNodeAction(t.Context(), conn, testCase.operation, "vmcluster-test", "dbnode-test", time.Second, time.Millisecond, func(_ context.Context, format string, args ...any) {
				progress = append(progress, fmt.Sprintf(format, args...))
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(progress) == 0 {
				t.Error("expected operation progress")
			}
		})
	}
}

func TestDBNodeActionRejectsInitialState(t *testing.T) {
	t.Parallel()

	for _, operation := range []string{"start", "stop", "reboot"} {
		for _, status := range []string{"PROVISIONING", "STARTING", "STOPPING", "UPDATING", "TERMINATING", "TERMINATED", "FAILED", "FUTURE_STATE", ""} {
			t.Run(operation+"/"+status, func(t *testing.T) {
				t.Parallel()
				conn := dbNodeActionTestClient(t, []dbNodeActionTestStep{{target: "GetDbNode", body: dbNodeActionTestNode(status)}})
				if err := runDBNodeAction(t.Context(), conn, operation, "vmcluster-test", "dbnode-test", time.Second, time.Millisecond, dbNodeActionTestProgress); err == nil {
					t.Fatalf("expected %s to reject initial state %q without a mutation", operation, status)
				}
			})
		}
	}

	t.Run("reboot stopped", func(t *testing.T) {
		t.Parallel()
		conn := dbNodeActionTestClient(t, []dbNodeActionTestStep{{target: "GetDbNode", body: dbNodeActionTestNode("STOPPED")}})
		if err := runDBNodeAction(t.Context(), conn, "reboot", "vmcluster-test", "dbnode-test", time.Second, time.Millisecond, dbNodeActionTestProgress); err == nil {
			t.Fatal("expected reboot to reject STOPPED without a mutation")
		}
	})
}

func TestDBNodeActionFailures(t *testing.T) {
	t.Parallel()

	for _, operation := range []string{"start", "stop", "reboot"} {
		initial := "AVAILABLE"
		transition := "UPDATING"
		if operation == "start" {
			initial = "STOPPED"
			transition = "STARTING"
		}
		if operation == "stop" {
			transition = "STOPPING"
		}
		read := dbNodeActionTestStep{target: "GetDbNode", body: dbNodeActionTestNode(initial)}
		mutation := dbNodeActionTestStep{target: dbNodeActionTestMutation(operation), body: fmt.Sprintf(`{"dbNodeId":"dbnode-test","status":%q}`, transition)}
		for name, testCase := range map[string]struct {
			steps []dbNodeActionTestStep
			want  string
		}{
			"missing node initially":                  {steps: []dbNodeActionTestStep{{target: "GetDbNode", body: `{}`}}},
			"mismatched node initially":               {steps: []dbNodeActionTestStep{{target: "GetDbNode", body: `{"dbNode":{"dbNodeId":"dbnode-other","status":"AVAILABLE"}}`}}},
			"read denied":                             {steps: []dbNodeActionTestStep{{target: "GetDbNode", code: http.StatusForbidden, body: `{"__type":"AccessDeniedException","message":"test permission denied"}`}}, want: "AccessDenied"},
			"read not found":                          {steps: []dbNodeActionTestStep{{target: "GetDbNode", code: http.StatusNotFound, body: `{"__type":"ResourceNotFoundException","message":"test node missing"}`}}, want: "ResourceNotFound"},
			"mutation denied":                         {steps: []dbNodeActionTestStep{read, {target: mutation.target, code: http.StatusForbidden, body: `{"__type":"AccessDeniedException","message":"test permission denied"}`}}, want: "AccessDenied"},
			"mutation server error is not retried":    {steps: []dbNodeActionTestStep{read, {target: mutation.target, code: http.StatusInternalServerError, body: `{"__type":"InternalServerException","message":"test server failure"}`}}, want: "InternalServer"},
			"mutation throttled is not retried":       {steps: []dbNodeActionTestStep{read, {target: mutation.target, code: http.StatusTooManyRequests, body: `{"__type":"ThrottlingException","message":"test throttled"}`}}, want: "Throttling"},
			"mutation transport error is not retried": {steps: []dbNodeActionTestStep{read, {target: mutation.target, err: errors.New("test connection lost")}}, want: "test connection lost"},
			"missing mutation node ID":                {steps: []dbNodeActionTestStep{read, {target: mutation.target, body: `{}`}}},
			"mismatched mutation node ID":             {steps: []dbNodeActionTestStep{read, {target: mutation.target, body: `{"dbNodeId":"dbnode-other","status":"UPDATING"}`}}},
			"mutation failed":                         {steps: []dbNodeActionTestStep{read, {target: mutation.target, body: `{"dbNodeId":"dbnode-test","status":"FAILED"}`}}, want: "FAILED"},
			"mutation terminated":                     {steps: []dbNodeActionTestStep{read, {target: mutation.target, body: `{"dbNodeId":"dbnode-test","status":"TERMINATED"}`}}, want: "TERMINATED"},
			"mutation unexpected state":               {steps: []dbNodeActionTestStep{read, {target: mutation.target, body: `{"dbNodeId":"dbnode-test","status":"FUTURE_STATE"}`}}, want: "FUTURE_STATE"},
			"node disappeared while polling":          {steps: []dbNodeActionTestStep{read, mutation, {target: "GetDbNode", body: `{}`}}},
			"different node while polling":            {steps: []dbNodeActionTestStep{read, mutation, {target: "GetDbNode", body: `{"dbNode":{"dbNodeId":"dbnode-other","status":"AVAILABLE"}}`}}},
			"poll denied":                             {steps: []dbNodeActionTestStep{read, mutation, {target: "GetDbNode", code: http.StatusForbidden, body: `{"__type":"AccessDeniedException","message":"test permission denied"}`}}, want: "AccessDenied"},
			"poll not found":                          {steps: []dbNodeActionTestStep{read, mutation, {target: "GetDbNode", code: http.StatusNotFound, body: `{"__type":"ResourceNotFoundException","message":"test node missing"}`}}, want: "ResourceNotFound"},
			"poll failed":                             {steps: []dbNodeActionTestStep{read, mutation, {target: "GetDbNode", body: dbNodeActionTestNode("FAILED")}}, want: "FAILED"},
			"poll terminated":                         {steps: []dbNodeActionTestStep{read, mutation, {target: "GetDbNode", body: dbNodeActionTestNode("TERMINATED")}}, want: "TERMINATED"},
			"poll terminating":                        {steps: []dbNodeActionTestStep{read, mutation, {target: "GetDbNode", body: dbNodeActionTestNode("TERMINATING")}}, want: "TERMINATING"},
			"poll unexpected":                         {steps: []dbNodeActionTestStep{read, mutation, {target: "GetDbNode", body: dbNodeActionTestNode("FUTURE_STATE")}}, want: "FUTURE_STATE"},
			"poll empty status":                       {steps: []dbNodeActionTestStep{read, mutation, {target: "GetDbNode", body: dbNodeActionTestNode("")}}},
		} {
			t.Run(operation+"/"+name, func(t *testing.T) {
				t.Parallel()
				conn := dbNodeActionTestClient(t, testCase.steps)
				err := runDBNodeAction(t.Context(), conn, operation, "vmcluster-test", "dbnode-test", time.Second, time.Millisecond, dbNodeActionTestProgress)
				if err == nil || !strings.Contains(err.Error(), testCase.want) {
					t.Fatalf("error = %v, want failure containing %q", err, testCase.want)
				}
			})
		}
	}
}

func TestDBNodeActionTimeout(t *testing.T) {
	t.Parallel()

	read := dbNodeActionTestStep{target: "GetDbNode", body: dbNodeActionTestNode("STOPPED")}
	mutation := dbNodeActionTestStep{target: "StartDbNode", body: `{"dbNodeId":"dbnode-test","status":"STARTING"}`}
	for name, steps := range map[string][]dbNodeActionTestStep{
		"initial read":  {{target: "GetDbNode", wait: true}},
		"mutation":      {read, {target: "StartDbNode", wait: true}},
		"poll request":  {read, mutation, {target: "GetDbNode", wait: true}},
		"poll interval": {read, mutation, {target: "GetDbNode", body: dbNodeActionTestNode("STARTING")}},
		"earlier requests consume wait budget": {
			{target: "GetDbNode", body: read.body, delay: 400 * time.Millisecond},
			{target: "StartDbNode", body: mutation.body, delay: 400 * time.Millisecond},
			{target: "GetDbNode", wait: true},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				conn := dbNodeActionTestClient(t, steps)
				start := time.Now()
				err := runDBNodeAction(t.Context(), conn, "start", "vmcluster-test", "dbnode-test", time.Second, 2*time.Second, dbNodeActionTestProgress)
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("error = %v, want deadline exceeded", err)
				}
				if elapsed := time.Since(start); elapsed != time.Second {
					t.Errorf("elapsed = %s, want total operation deadline of 1s", elapsed)
				}
			})
		})
	}
}

func TestDBNodeActionRebootRequiresObservedTransition(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		conn := dbNodeActionTestClient(t, []dbNodeActionTestStep{
			{target: "GetDbNode", body: dbNodeActionTestNode("AVAILABLE")},
			{target: "RebootDbNode", body: `{"dbNodeId":"dbnode-test","status":"AVAILABLE"}`},
			{target: "GetDbNode", body: dbNodeActionTestNode("AVAILABLE"), repeat: true},
		})
		err := runDBNodeAction(t.Context(), conn, "reboot", "vmcluster-test", "dbnode-test", time.Second, time.Second/10, dbNodeActionTestProgress)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("error = %v, want timeout because stale AVAILABLE reads cannot confirm reboot", err)
		}
	})
}

func TestDBNodeActionCancellation(t *testing.T) {
	t.Parallel()
	for _, stage := range []string{"before request", "initial read", "mutation", "poll"} {
		t.Run(stage, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				var steps []dbNodeActionTestStep
				if stage == "before request" {
					cancel()
				} else {
					if stage != "initial read" {
						steps = append(steps, dbNodeActionTestStep{target: "GetDbNode", body: dbNodeActionTestNode("STOPPED")})
					}
					if stage == "poll" {
						steps = append(steps, dbNodeActionTestStep{target: "StartDbNode", body: `{"dbNodeId":"dbnode-test","status":"STARTING"}`})
					}
					target := "GetDbNode"
					if stage == "mutation" {
						target = "StartDbNode"
					}
					steps = append(steps, dbNodeActionTestStep{target: target, wait: true})
					time.AfterFunc(time.Second, cancel)
				}
				conn := dbNodeActionTestClient(t, steps)
				err := runDBNodeAction(ctx, conn, "start", "vmcluster-test", "dbnode-test", time.Hour, time.Millisecond, dbNodeActionTestProgress)
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("error = %v, want cancellation", err)
				}
			})
		})
	}
}

func TestDBNodeActionRejectsInvalidInput(t *testing.T) {
	t.Parallel()
	for name, testCase := range map[string]struct {
		operation string
		clusterID string
		nodeID    string
		timeout   time.Duration
	}{
		"invalid operation": {operation: "delete", clusterID: "vmcluster-test", nodeID: "dbnode-test", timeout: time.Hour},
		"empty cluster":     {operation: "start", nodeID: "dbnode-test", timeout: time.Hour},
		"short cluster":     {operation: "start", clusterID: "short", nodeID: "dbnode-test", timeout: time.Hour},
		"long cluster":      {operation: "start", clusterID: strings.Repeat("a", 65), nodeID: "dbnode-test", timeout: time.Hour},
		"cluster path":      {operation: "start", clusterID: "cluster/other", nodeID: "dbnode-test", timeout: time.Hour},
		"empty node":        {operation: "start", clusterID: "vmcluster-test", timeout: time.Hour},
		"short node":        {operation: "start", clusterID: "vmcluster-test", nodeID: "short", timeout: time.Hour},
		"long node":         {operation: "start", clusterID: "vmcluster-test", nodeID: strings.Repeat("a", 65), timeout: time.Hour},
		"node path":         {operation: "start", clusterID: "vmcluster-test", nodeID: "node/other", timeout: time.Hour},
		"zero timeout":      {operation: "start", clusterID: "vmcluster-test", nodeID: "dbnode-test"},
		"negative timeout":  {operation: "start", clusterID: "vmcluster-test", nodeID: "dbnode-test", timeout: -time.Second},
		"excessive timeout": {operation: "start", clusterID: "vmcluster-test", nodeID: "dbnode-test", timeout: 24*time.Hour + time.Second},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			conn := dbNodeActionTestClient(t, nil)
			if err := runDBNodeAction(t.Context(), conn, testCase.operation, testCase.clusterID, testCase.nodeID, testCase.timeout, time.Millisecond, dbNodeActionTestProgress); err == nil {
				t.Fatal("expected invalid configuration to fail before contacting AWS")
			}
		})
	}
}

func TestDBNodeActionSchema(t *testing.T) {
	t.Parallel()
	for name, factory := range map[string]func(context.Context) (action.ActionWithConfigure, error){
		"start":  newStartDBNodeAction,
		"stop":   newStopDBNodeAction,
		"reboot": newRebootDBNodeAction,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			a, err := factory(ctx)
			if err != nil {
				t.Fatal(err)
			}
			var response action.SchemaResponse
			a.Schema(ctx, action.SchemaRequest{}, &response)
			if response.Diagnostics.HasError() {
				t.Fatal(response.Diagnostics)
			}
			for _, attribute := range []string{"cloud_vm_cluster_id", "db_node_id"} {
				field, ok := response.Schema.Attributes[attribute].(schema.StringAttribute)
				if !ok || !field.Required || field.Optional {
					t.Fatalf("%s must be a required string", attribute)
				}
				for _, testCase := range []struct {
					value   types.String
					invalid bool
				}{
					{value: types.StringValue("a12345")},
					{value: types.StringValue("a_~.-Z")},
					{value: types.StringValue(strings.Repeat("a", 64))},
					{value: types.StringNull()},
					{value: types.StringUnknown()},
					{value: types.StringValue(""), invalid: true},
					{value: types.StringValue("short"), invalid: true},
					{value: types.StringValue(strings.Repeat("a", 65)), invalid: true},
					{value: types.StringValue("node/other"), invalid: true},
					{value: types.StringValue("node name"), invalid: true},
					{value: types.StringValue("dbnode-test\n"), invalid: true},
					{value: types.StringValue("dbnode-秘密"), invalid: true},
				} {
					request := validator.StringRequest{Path: path.Root(attribute), ConfigValue: testCase.value}
					var result validator.StringResponse
					for _, v := range field.Validators {
						v.ValidateString(ctx, request, &result)
					}
					if got := result.Diagnostics.HasError(); got != testCase.invalid {
						t.Errorf("%s value %s: validation error = %t, want %t; %v", attribute, testCase.value, got, testCase.invalid, result.Diagnostics)
					}
				}
			}
			field, ok := response.Schema.Attributes[names.AttrTimeout].(schema.Int64Attribute)
			if !ok || !field.Optional || field.Required {
				t.Fatal("timeout must be an optional integer")
			}
			for _, testCase := range []struct {
				value   types.Int64
				invalid bool
			}{
				{value: types.Int64Null()},
				{value: types.Int64Unknown()},
				{value: types.Int64Value(1)},
				{value: types.Int64Value(3600)},
				{value: types.Int64Value(86400)},
				{value: types.Int64Value(-1), invalid: true},
				{value: types.Int64Value(0), invalid: true},
				{value: types.Int64Value(86401), invalid: true},
			} {
				request := validator.Int64Request{Path: path.Root(names.AttrTimeout), ConfigValue: testCase.value}
				var result validator.Int64Response
				for _, v := range field.Validators {
					v.ValidateInt64(ctx, request, &result)
				}
				if got := result.Diagnostics.HasError(); got != testCase.invalid {
					t.Errorf("timeout value %s: validation error = %t, want %t; %v", testCase.value, got, testCase.invalid, result.Diagnostics)
				}
			}
		})
	}
}

func TestDBNodeActionInvoke(t *testing.T) {
	t.Parallel()
	for operation, factory := range map[string]func(context.Context) (action.ActionWithConfigure, error){
		"start":  newStartDBNodeAction,
		"stop":   newStopDBNodeAction,
		"reboot": newRebootDBNodeAction,
	} {
		for scenario, testCase := range map[string]struct {
			timeout        types.Int64
			remaining      time.Duration
			invalidTimeout bool
		}{
			"default timeout":           {timeout: types.Int64Null(), remaining: time.Hour},
			"explicit timeout":          {timeout: types.Int64Value(60), remaining: time.Minute},
			"minimum timeout":           {timeout: types.Int64Value(1), remaining: time.Second},
			"maximum timeout":           {timeout: types.Int64Value(86400), remaining: 24 * time.Hour},
			"zero timeout":              {timeout: types.Int64Value(0), invalidTimeout: true},
			"negative timeout":          {timeout: types.Int64Value(-1), invalidTimeout: true},
			"excessive timeout":         {timeout: types.Int64Value(86401), invalidTimeout: true},
			"positive overflow timeout": {timeout: types.Int64Value(18446744100), invalidTimeout: true},
			"negative overflow timeout": {timeout: types.Int64Value(-18446744073), invalidTimeout: true},
			"maximum int64 timeout":     {timeout: types.Int64Value(math.MaxInt64), invalidTimeout: true},
			"minimum int64 timeout":     {timeout: types.Int64Value(math.MinInt64), invalidTimeout: true},
			"read denied":               {timeout: types.Int64Null()},
			"unknown cluster":           {timeout: types.Int64Null()},
			"unknown node":              {timeout: types.Int64Null()},
			"unknown timeout":           {timeout: types.Int64Unknown()},
		} {
			t.Run(operation+"/"+scenario, func(t *testing.T) {
				t.Parallel()
				synctest.Test(t, func(t *testing.T) {
					ctx := t.Context()
					config := dbNodeActionModel{CloudVMClusterID: types.StringValue("vmcluster-test"), DBNodeID: types.StringValue("dbnode-test"), Timeout: testCase.timeout}
					initial, transition, final := "STOPPED", "STARTING", "AVAILABLE"
					if operation == "stop" {
						initial, transition, final = "AVAILABLE", "STOPPING", "STOPPED"
					}
					if operation == "reboot" {
						initial, transition = "AVAILABLE", "UPDATING"
					}
					steps := []dbNodeActionTestStep{
						{target: "GetDbNode", body: dbNodeActionTestNode(initial), remaining: testCase.remaining},
						{target: dbNodeActionTestMutation(operation), body: fmt.Sprintf(`{"dbNodeId":"dbnode-test","status":%q}`, transition)},
						{target: "GetDbNode", body: dbNodeActionTestNode(final)},
					}
					if operation == "reboot" {
						steps = append(steps, dbNodeActionTestStep{target: "GetDbNode", body: dbNodeActionTestNode(final)})
					}
					wantError := testCase.invalidTimeout
					if testCase.invalidTimeout {
						steps = nil
					}
					switch scenario {
					case "read denied", "minimum timeout":
						wantError = true
						steps = []dbNodeActionTestStep{{target: "GetDbNode", code: http.StatusForbidden, body: `{"__type":"AccessDeniedException","message":"test permission denied"}`, remaining: testCase.remaining}}
					case "unknown cluster":
						config.CloudVMClusterID = types.StringUnknown()
						steps, wantError = nil, true
					case "unknown node":
						config.DBNodeID = types.StringUnknown()
						steps, wantError = nil, true
					case "unknown timeout":
						steps, wantError = nil, true
					}
					conn := dbNodeActionTestClient(t, steps)
					client := new(conns.AWSClient)
					client.SetHTTPClient(ctx, &http.Client{Transport: autonomousDatabaseSweepTransport(conn.Options().HTTPClient.Do)})
					client.SetServicePackages(ctx, map[string]conns.ServicePackage{names.ODB: &servicePackage{}})
					providerConfig := conns.Config{AccessKey: "test", SecretKey: "test", Region: endpoints.UsEast1RegionID, SkipCredsValidation: true, SkipRequestingAccountId: true, MaxRetries: 0, SharedConfigFiles: []string{}, SharedCredentialsFiles: []string{}}
					client, diagnostics := providerConfig.ConfigureProvider(ctx, client)
					if diagnostics.HasError() {
						t.Fatal(diagnostics)
					}
					a, err := factory(ctx)
					if err != nil {
						t.Fatal(err)
					}
					var configured action.ConfigureResponse
					a.Configure(ctx, action.ConfigureRequest{ProviderData: client}, &configured)
					if configured.Diagnostics.HasError() {
						t.Fatal(configured.Diagnostics)
					}
					var actionSchema action.SchemaResponse
					a.Schema(ctx, action.SchemaRequest{}, &actionSchema)
					actionSchema.Schema.Attributes[names.AttrRegion] = schema.StringAttribute{Optional: true}
					plan := tfsdk.Plan{Schema: actionSchema.Schema}
					if diags := plan.Set(ctx, config); diags.HasError() {
						t.Fatal(diags)
					}
					request := action.InvokeRequest{Config: tfsdk.Config{Schema: actionSchema.Schema, Raw: plan.Raw}}
					var progress []string
					response := action.InvokeResponse{SendProgress: func(event action.InvokeProgressEvent) { progress = append(progress, event.Message) }}
					a.Invoke(ctx, request, &response)
					if response.Diagnostics.HasError() != wantError {
						t.Fatalf("HasError = %t, want %t; diagnostics: %v", response.Diagnostics.HasError(), wantError, response.Diagnostics)
					}
					if testCase.invalidTimeout && !strings.Contains(fmt.Sprint(response.Diagnostics), "db node action timeout must be between 1 and 86400 seconds") {
						t.Errorf("diagnostics = %v, want invalid timeout", response.Diagnostics)
					}
					if !wantError && len(progress) == 0 {
						t.Error("successful Invoke did not send progress to Terraform")
					}
				})
			})
		}
	}
}

func TestDBNodeActionProgressAndLogs(t *testing.T) {
	t.Parallel()
	const sensitive = "sensitive-database-status-detail"
	var logs bytes.Buffer
	ctx := tflogtest.RootLogger(t.Context(), &logs)
	conn := dbNodeActionTestClient(t, []dbNodeActionTestStep{
		{target: "GetDbNode", body: `{"dbNode":{"dbNodeId":"dbnode-test","status":"STOPPED","statusReason":"` + sensitive + `"}}`},
		{target: "StartDbNode", body: `{"dbNodeId":"dbnode-test","status":"STARTING","statusReason":"` + sensitive + `"}`},
		{target: "GetDbNode", body: `{"dbNode":{"dbNodeId":"dbnode-test","status":"STARTING","statusReason":"` + sensitive + `"}}`},
		{target: "GetDbNode", body: dbNodeActionTestNode("AVAILABLE")},
	})
	var progress []string
	err := runDBNodeAction(ctx, conn, "start", "vmcluster-test", "dbnode-test", time.Second, time.Millisecond, func(_ context.Context, format string, args ...any) {
		progress = append(progress, fmt.Sprintf(format, args...))
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(progress) < 2 {
		t.Errorf("progress messages = %v, want dispatch and completion updates", progress)
	}
	entries, err := tflogtest.MultilineJSONDecode(&logs)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected structured action logs")
	}
	output := fmt.Sprint(entries, progress)
	if strings.Contains(output, sensitive) {
		t.Fatal("logs or progress exposed unfiltered status reason")
	}
	for _, want := range []string{"dbnode-test", "start"} {
		if !strings.Contains(output, want) {
			t.Errorf("logs and progress should identify %q", want)
		}
	}
}

type dbNodeActionTestStep struct {
	target    string
	body      string
	code      int
	err       error
	wait      bool
	repeat    bool
	delay     time.Duration
	remaining time.Duration
}

func dbNodeActionTestClient(t *testing.T, steps []dbNodeActionTestStep) *odb.Client {
	t.Helper()
	calls := 0
	t.Cleanup(func() {
		if calls != len(steps) && (len(steps) == 0 || !steps[len(steps)-1].repeat || calls < len(steps)) {
			t.Errorf("API calls = %d, want %d", calls, len(steps))
		}
	})
	return odb.New(odb.Options{
		Region:      endpoints.UsEast1RegionID,
		Credentials: aws.AnonymousCredentials{},
		Retryer: awsretry.NewStandard(func(options *awsretry.StandardOptions) {
			options.MaxAttempts = 3
			options.Backoff = awsretry.BackoffDelayerFunc(func(int, error) (time.Duration, error) { return 0, nil })
		}),
		HTTPClient: smithyhttp.ClientDoFunc(func(request *http.Request) (*http.Response, error) {
			index := calls
			calls++
			if len(steps) > 0 && steps[len(steps)-1].repeat && index >= len(steps) {
				index = len(steps) - 1
			}
			if index >= len(steps) {
				t.Errorf("unexpected additional API call: %s", request.Header.Get("X-Amz-Target"))
				return nil, errors.New("unexpected additional API call")
			}
			step := steps[index]
			if got, want := request.Header.Get("X-Amz-Target"), "Odb."+step.target; got != want {
				t.Errorf("API call %d = %s, want %s", index, got, want)
			}
			var payload map[string]string
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			if got := payload["cloudVmClusterId"]; got != "vmcluster-test" {
				t.Errorf("cloudVmClusterId = %q, want vmcluster-test", got)
			}
			if got := payload["dbNodeId"]; got != "dbnode-test" {
				t.Errorf("dbNodeId = %q, want dbnode-test", got)
			}
			for key := range payload {
				if !slices.Contains([]string{"cloudVmClusterId", "dbNodeId"}, key) {
					t.Errorf("unexpected request field %q", key)
				}
			}
			if step.remaining > 0 {
				deadline, ok := request.Context().Deadline()
				if !ok || time.Until(deadline) != step.remaining {
					t.Errorf("request deadline remaining = %s, want %s", time.Until(deadline), step.remaining)
				}
			}
			if step.delay > 0 {
				time.Sleep(step.delay)
			}
			if step.wait {
				<-request.Context().Done()
				return nil, request.Context().Err()
			}
			if step.err != nil {
				return nil, step.err
			}
			code := step.code
			if code == 0 {
				code = http.StatusOK
			}
			return &http.Response{
				StatusCode: code,
				Header:     http.Header{"Content-Type": {"application/x-amz-json-1.0"}},
				Body:       io.NopCloser(strings.NewReader(step.body)),
				Request:    request,
			}, nil
		}),
	})
}

func dbNodeActionTestNode(status string) string {
	return fmt.Sprintf(`{"dbNode":{"dbNodeId":"dbnode-test","status":%q}}`, status)
}

func dbNodeActionTestMutation(operation string) string {
	switch operation {
	case "start":
		return "StartDbNode"
	case "stop":
		return "StopDbNode"
	default:
		return "RebootDbNode"
	}
}

func dbNodeActionTestProgress(context.Context, string, ...any) {}
