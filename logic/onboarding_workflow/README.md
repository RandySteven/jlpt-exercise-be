# Onboarding workflow

Registers a user and finishes activation in the same Temporal run. Workflow id is `onboarding_workflow_execution-{requestID}`.

Registered order:

1. `persist_onboarding_request_activity`
2. `register_user_activity`
3. `send_activation_email_activity`
4. `update_onboarding_status_activity` waits for signal `activated_user_signal`

`POST /auth/register` returns when query `RegisterResponse` has a response. That happens in `register_user`, before the activation email is sent and before the user opens the link. The run stays open until `POST /auth/activated` signals it.

```mermaid
sequenceDiagram
    actor Client
    participant Handler as OnboardingHandler
    participant Usecase as OnboardingUsecase
    participant Cache as OnboardingCache
    participant WF as OnboardingWorkflow
    participant Temporal
    participant OnboardingRepo as OnboardingRepository
    participant UserRepo as UserRepository
    participant SMTP

    Client->>Handler: POST /auth/register
    Handler->>Usecase: Register
    Usecase->>Cache: GetRegisterResponse by email
    alt cached registration
        Cache-->>Client: RegisterResponse
    else first registration
        Usecase->>WF: OnboardingRegisterWorkflow
        WF->>Temporal: StartWorkflow onboarding_workflow_execution-{requestID}
        Temporal->>OnboardingRepo: persist_onboarding_request Save
        Temporal->>UserRepo: register_user Save status PROCESSING
        Temporal->>OnboardingRepo: register_user Update status ACTIVE
        Temporal->>Temporal: build activation token and RegisterResponse
        Temporal->>Cache: SetRegisterResponse inside query handler
        WF->>Temporal: Query RegisterResponse
        Temporal-->>Client: RegisterResponse with activation token

        Temporal->>SMTP: send_activation_email HTML template
        Note over Temporal: update_onboarding_status waits for activated_user_signal
    end

    Client->>Handler: POST /auth/activated
    Handler->>Usecase: ActivateUser
    Usecase->>WF: ActivateUserRegisterWorkflow
    WF->>WF: decode activation token into workflow id and run id
    WF->>Temporal: signal activated_user_signal
    Temporal->>OnboardingRepo: update status ONBOARDED
    Temporal->>UserRepo: update status ACTIVE
    WF->>Temporal: GetWorkflowResult
    Temporal-->>Client: ActivateResponse success
```

The activation token carries the user id, onboarding id, workflow id, and run id. A failed or rejected run cancels the register wait. Input errors use the non-retryable type `OnboardingInputError`.
