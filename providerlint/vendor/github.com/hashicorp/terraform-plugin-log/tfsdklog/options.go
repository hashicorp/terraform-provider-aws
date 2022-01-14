package tfsdklog

import (
	"os"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-plugin-log/internal/logging"
)

// Options is a collection of logging options, useful for collecting arguments
// to NewSubsystem, NewRootSDKLogger, and NewRootProviderLogger before calling
// them.
type Options []logging.Option

// WithLogName returns an option that will set the logger name explicitly to
// `name`. This has no effect when used with NewSubsystem.
func WithLogName(name string) logging.Option {
	return func(l logging.LoggerOpts) logging.LoggerOpts {
		l.Name = name
		return l
	}
}

// WithLevelFromEnv returns an option that will set the level of the logger
// based on the string in an environment variable. The environment variable
// checked will be `name` and `subsystems`, joined by _ and in all caps.
func WithLevelFromEnv(name string, subsystems ...string) logging.Option {
	return func(l logging.LoggerOpts) logging.LoggerOpts {
		envVar := strings.Join(subsystems, "_")
		if envVar != "" {
			envVar = "_" + envVar
		}
		envVar = strings.ToUpper(name + envVar)
		l.Level = hclog.LevelFromString(os.Getenv(envVar))
		return l
	}
}

// WithLevel returns an option that will set the level of the logger.
func WithLevel(level hclog.Level) logging.Option {
	return func(l logging.LoggerOpts) logging.LoggerOpts {
		l.Level = level
		return l
	}
}

// WithoutLocation returns an option that disables including the location of
// the log line in the log output, which is on by default. This has no effect
// when used with NewSubsystem.
func WithoutLocation() logging.Option {
	return func(l logging.LoggerOpts) logging.LoggerOpts {
		l.IncludeLocation = false
		return l
	}
}

// WithStderrFromInit returns an option that tells the logger to write to the
// os.Stderr that was present when the program started, not the one that is
// available at runtime. Some versions of Terraform overwrite os.Stderr with an
// io.Writer that is never read, so any log lines written to it will be lost.
func WithStderrFromInit() logging.Option {
	return logging.WithOutput(logging.Stderr)
}
