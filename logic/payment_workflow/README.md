# Payment workflow

Opens a subscription payment and confirms it after Midtrans reports a result. Workflow id is `payment_workflow-{idempotencyKey}`. Query name is `query_payment_workflow`.

`POST /payment/subscription` chooses the start:

- Active subscription: the usecase returns the current row and does not start this workflow.
- Pending subscription: `ContinueSubscription` signals `create_payment` on the existing run.
- No subscription: `CreateSubscription` starts the run and signals `create_subscription`.

Registered order:

1. `check_plan_activity` waits for `create_subscription`, then picks the next activity.
2. `plan_exist_activity` when a non-pending subscription already exists.
3. `create_subscription_activity` saves a pending subscription and signals `create_payment`.
4. `persist_payment_activity` waits for `create_payment`, saves the payment, and signals `midtrans_call`.
5. `midtrans_call_activity` waits for `midtrans_call` and creates the Midtrans transaction.
6. `update_payment_status_activity` waits for `confirm_payment`.
7. `publish_payment_activity` writes the payment topic.
8. `send_subscription_email_activity` sends the active-subscription email.

`payment_failed_activity` marks the payment and subscription failed, then returns a non-retryable rejection. `cancel_payment_activity` is named on the Midtrans branch and is not registered.

```mermaid
sequenceDiagram
    actor Client
    participant Handler as PaymentHandler
    participant Usecase as PaymentUsecase
    participant SubRepo as SubscriptionRepository
    participant WF as PaymentWorkflow
    participant Temporal
    participant UserRepo as UserRepository
    participant OnboardingRepo as OnboardingRepository
    participant PlanRepo as SubscriptionPlanRepository
    participant PayRepo as PaymentRepository
    participant Midtrans
    participant NSQ as PaymentTopic
    participant SMTP

    Client->>Handler: POST /payment/subscription
    Handler->>Usecase: CreateSubscription
    Usecase->>SubRepo: FindCurrentByUserID
    alt status ACTIVE
        Usecase-->>Client: current subscription
    else status PENDING
        Usecase->>WF: ContinueSubscription
        WF->>Temporal: signal create_payment
        Note over Temporal: persist_payment then midtrans_call
        WF->>Temporal: poll query until RedirectURL is set
        Temporal-->>Client: subscription response with redirect URL
    else no current subscription
        Usecase->>WF: CreateSubscription
        WF->>Temporal: StartWorkflow payment_workflow-{idempotencyKey}
        WF->>Temporal: signal create_subscription
        Temporal->>UserRepo: check_plan FindByID
        Temporal->>OnboardingRepo: check_plan load customer email
        Temporal->>PlanRepo: check_plan FindByID
        Temporal->>SubRepo: check_plan FindCurrentByUserID
        alt subscription already pending
            Temporal->>PayRepo: persist_payment
        else subscription exists and is not pending
            Temporal->>Temporal: plan_exist returns current subscription
        else no subscription
            Temporal->>SubRepo: create_subscription Save status PENDING
            Note over WF, Temporal: query is ready once SubscriptionID is set
            WF->>Temporal: Query query_payment_workflow
            Temporal-->>Client: subscription response
            Temporal->>Temporal: signal create_payment
            Temporal->>PayRepo: persist_payment Save or reuse by reference id
            Temporal->>Temporal: signal midtrans_call
            Temporal->>Midtrans: CreateTransaction
            Midtrans-->>Temporal: token and redirect URL
            Temporal->>PayRepo: store token and redirect URL
        end
    end

    Client->>Handler: POST /webhook/callback or POST /payment/confirmation
    Handler->>Usecase: ConfirmPayment
    Usecase->>WF: PaymentConfirmation
    WF->>Temporal: Query query_payment_workflow
    WF->>Temporal: signal confirm_payment
    Temporal->>PayRepo: confirm_payment Update payment status
    alt settlement or capture
        Temporal->>SubRepo: set subscription ACTIVE and dates
        Temporal->>NSQ: publish payment message
        Temporal->>SMTP: send subscription active email
    else cancel deny expire or failure
        Temporal->>PayRepo: payment_failed
        Temporal->>SubRepo: payment_failed
    end
    WF->>Temporal: GetWorkflowResult
    Temporal-->>Client: PaymentConfirmationResponse redirect URL
```

`CreateSubscription` returns when the query has a subscription id, which `create_subscription` sets before Midtrans returns a redirect URL. `ContinueSubscription` keeps polling until `RedirectURL` is present. A failed or rejected run cancels that wait.

Input and rejection errors use non-retryable types and are not retried. Each activity allows 5 attempts, starting at 1 minute and backing off to 5 minutes, with a 25 second start-to-close timeout.
