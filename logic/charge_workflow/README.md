# Charge workflow

Placeholder for charging and capturing a subscription. Workflow name is `ChargeWorkflow`.

`chargeWorkflowAndActivities` registers that workflow name and does not add activities. `Charge` and `Capture` return `nil, nil`. `charge` also returns `nil, nil`. No handler or usecase calls this package.

Activity names reserved in this package, and not registered:

- `charge_activity`
- `capture_activity`
- `publish_activity`

```mermaid
sequenceDiagram
    participant Caller
    participant ChargeWorkflow
    participant Temporal

    Note over Caller, Temporal: No HTTP route starts this workflow today
    Caller->>ChargeWorkflow: Charge or Capture
    ChargeWorkflow-->>Caller: nil response

    ChargeWorkflow->>Temporal: RegisterWorkflow ChargeWorkflow
    Note over Temporal: charge returns nil and does not run charge_activity capture_activity or publish_activity
```

The unused activity options in `chargeWorkflowAndActivities` match the other workflows: 25 second start-to-close timeout and 5 attempts from 1 minute up to 5 minutes.
