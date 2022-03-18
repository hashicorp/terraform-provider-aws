package logging

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
	helperlogging "github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	testing "github.com/mitchellh/go-testing-interface"
)

// InitContext creates SDK logger contexts.
func InitContext(ctx context.Context) context.Context {
	ctx = tfsdklog.NewRootSDKLogger(ctx)
	ctx = tfsdklog.NewSubsystem(ctx, SubsystemHelperResource, tfsdklog.WithLevelFromEnv(EnvTfLogSdkHelperResource))
	ctx = tfsdklog.NewSubsystem(ctx, SubsystemHelperSchema, tfsdklog.WithLevelFromEnv(EnvTfLogSdkHelperSchema))

	return ctx
}

// InitTestContext registers the terraform-plugin-log/tfsdklog test sink,
// configures the standard library log package, and creates SDK logger
// contexts.
//
// It may be possible to eliminate the helper/logging handling if all
// log package calls are replaced with tfsdklog and any go-plugin or
// terraform-exec logger configurations are updated to the tfsdklog logger.
func InitTestContext(ctx context.Context, t testing.T) context.Context {
	helperlogging.SetOutput(t)

	ctx = tfsdklog.RegisterTestSink(ctx, t)
	ctx = InitContext(ctx)
	ctx = TestNameContext(ctx, t.Name())

	return ctx
}

// TestNameContext adds the current test name to loggers.
func TestNameContext(ctx context.Context, testName string) context.Context {
	ctx = tfsdklog.SubsystemWith(ctx, SubsystemHelperResource, KeyTestName, testName)
	ctx = tfsdklog.SubsystemWith(ctx, SubsystemHelperSchema, KeyTestName, testName)

	return ctx
}

// TestStepNumberContext adds the current test step number to loggers.
func TestStepNumberContext(ctx context.Context, stepNumber int) context.Context {
	ctx = tfsdklog.SubsystemWith(ctx, SubsystemHelperResource, KeyTestStepNumber, stepNumber)
	ctx = tfsdklog.SubsystemWith(ctx, SubsystemHelperSchema, KeyTestStepNumber, stepNumber)

	return ctx
}

// TestTerraformPathContext adds the current test Terraform CLI path to loggers.
func TestTerraformPathContext(ctx context.Context, terraformPath string) context.Context {
	ctx = tfsdklog.SubsystemWith(ctx, SubsystemHelperResource, KeyTestTerraformPath, terraformPath)
	ctx = tfsdklog.SubsystemWith(ctx, SubsystemHelperSchema, KeyTestTerraformPath, terraformPath)

	return ctx
}

// TestWorkingDirectoryContext adds the current test working directory to loggers.
func TestWorkingDirectoryContext(ctx context.Context, workingDirectory string) context.Context {
	ctx = tfsdklog.SubsystemWith(ctx, SubsystemHelperResource, KeyTestWorkingDirectory, workingDirectory)
	ctx = tfsdklog.SubsystemWith(ctx, SubsystemHelperSchema, KeyTestWorkingDirectory, workingDirectory)

	return ctx
}
