package charge_workflow

import (
	temporal_client "github.com/RandySteven/go-cook/temporal"
)

type (
	chargeWorkflow struct {
		workflow       temporal_client.WorkflowExecution
		temporalClient temporal_client.Temporal
	}

	ChargeWorkflow interface {}
)
