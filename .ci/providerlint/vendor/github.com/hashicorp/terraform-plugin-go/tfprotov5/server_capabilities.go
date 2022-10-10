package tfprotov5

// ServerCapabilities allows providers to communicate optionally supported
// protocol features, such as forward-compatible Terraform behavior changes.
//
// This information is used in GetProviderSchemaResponse as capabilities are
// static features which must be known upfront in the provider server.
type ServerCapabilities struct {
	// PlanDestroy signals that a provider expects a call to
	// PlanResourceChange when a resource is going to be destroyed. This is
	// opt-in to prevent unexpected errors or panics since the
	// ProposedNewState in PlanResourceChangeRequest will be a null value.
	PlanDestroy bool
}
