# Refund workflow

Registered activity graph for a Midtrans refund. Workflow name is `refund_workflow`.

Nothing in `handlers` or `usecases` starts this workflow. `refund` returns without running activities. `checkPayment`, `initiateRefund`, `midtransCall`, and `updatePaymentStatus` return the execution data unchanged.

Registered order, once a run is started:

1. `check_payment_existing` with no signal.
2. `create_refund` with no signal.
3. `midtrans_refund` waits for signal `midtrans_refund`, then goes to `refund_payment_success` or `refund_payment_failed`.

`refund_payment_success` and `refund_payment_failed` are named as next activities and are not registered. `updatePaymentStatus` is not on the registered chain.

```mermaid
sequenceDiagram
    participant Caller
    participant RefundWorkflow
    participant Temporal
    participant Check as check_payment_existing
    participant Create as create_refund
    participant Midtrans as midtrans_refund

    Note over Caller, Midtrans: No HTTP route starts this workflow today
    Caller->>RefundWorkflow: Refund is declared and not implemented
    RefundWorkflow->>Temporal: RegisterWorkflow refund_workflow
    Note over Temporal: refund returns nil and does not execute activities

    rect rgb(245, 245, 245)
        Note over Temporal, Midtrans: Registered transitions, not executed
        Temporal->>Check: check payment
        Check-->>Temporal: execution data unchanged
        Temporal->>Create: initiate refund
        Create-->>Temporal: execution data unchanged
        Temporal->>Temporal: wait for signal midtrans_refund
        Temporal->>Midtrans: midtrans refund call
        Midtrans-->>Temporal: execution data unchanged
        alt success name
            Temporal->>Temporal: refund_payment_success not registered
        else failure name
            Temporal->>Temporal: refund_payment_failed not registered
        end
    end
```

Activity options match the other workflows: 25 second start-to-close timeout and 5 attempts from 1 minute up to 5 minutes.
