# Workflows

Temporal workflows live in the packages below. Each package README is a sequence diagram of the registered activity order and the HTTP call that starts it.

| Package | Entry | What it does |
|---------|-------|----------------|
| [onboarding_workflow](onboarding_workflow/README.md) | `POST /auth/register`, `POST /auth/activated` | Saves the onboarding record, creates the user, emails an activation link, then waits for the activation signal. |
| [payment_workflow](payment_workflow/README.md) | `POST /payment/subscription`, `POST /payment/confirmation`, `POST /webhook/callback` | Creates or resumes a subscription, opens a Midtrans payment, then confirms it. |
| [exercise_workflow](exercise_workflow/README.md) | `POST /exercise/generate` | Claims an exercise, fills the prompt, asks DeepSeek for questions, and stores the result. |
| [refund_workflow](refund_workflow/README.md) | Not wired to HTTP | Registered refund steps. Activity bodies do not call Midtrans or the database yet. |
| [charge_workflow](charge_workflow/README.md) | Not wired to HTTP | Workflow name is registered. `Charge` and `Capture` return without running activities. |
