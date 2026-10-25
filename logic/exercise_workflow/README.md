# Exercise workflow

Generates a JLPT exercise for a subscribed user. Workflow id is `exercise_workflow-{idempotencyKey}`.

The HTTP handler returns as soon as the workflow query `GenerateExerciseResponse` is readable. That query is ready while status is `processing`, so the client can receive the in-progress exercise before DeepSeek finishes. Activities keep running on Temporal after that response.

Registered order:

1. `persist_user_request_activity`
2. `read_prompt_txt_activity` signals `sg_processing_generation`
3. `generate_exercise_activity` waits for `sg_processing_generation`, then signals `sg_completed_generation`
4. `persist_generated_response_activity` waits for `sg_completed_generation`
5. `publish_activity` marks the run complete

A failed activity is named `failed_activity`. This package does not register a handler for it.

```mermaid
sequenceDiagram
    actor Client
    participant Handler as ExerciseHandler
    participant Usecase as ExerciseUsecase
    participant Cache as ExerciseCache
    participant Repo as ExerciseRepository
    participant UserRepo as UserRepository
    participant WF as ExerciseWorkflow
    participant Temporal
    participant Prompt as generate_exercise.txt
    participant DeepSeek

    Client->>Handler: POST /exercise/generate
    Handler->>Usecase: GenerateExercise
    Usecase->>Usecase: require an active subscription
    Usecase->>Cache: GetExerciseByIdempotencyKey
    alt cache hit
        Cache-->>Usecase: stored response
        Usecase-->>Client: cached exercise
    else cache miss
        Usecase->>Repo: FindLatestByUserID
        alt same key already in progress
            Usecase-->>Client: existing exercise
        else another exercise is in progress
            Usecase-->>Client: 409 conflict
        else no exercise in progress
            Usecase->>Repo: Save claimed row status in_generated
            Usecase->>WF: GenerateExercise
            WF->>Temporal: StartWorkflow exercise_workflow-{idempotencyKey}
            Temporal->>Temporal: init query GenerateExerciseResponse status processing
            WF->>Temporal: Query GenerateExerciseResponse
            Temporal-->>WF: processing response
            WF-->>Client: processing exercise

            Temporal->>UserRepo: persist_user_request FindByID
            Temporal->>Repo: persist_user_request reuse or Save exercise
            Temporal->>Prompt: read_prompt_txt
            Temporal->>Cache: cache processing exercise
            Temporal->>Temporal: signal sg_processing_generation
            Temporal->>DeepSeek: generate ChatGenerate
            DeepSeek-->>Temporal: raw exercise JSON
            Temporal->>Temporal: sanitize JSON and signal sg_completed_generation
            Temporal->>Repo: persist_generated_response Update exercise
            Temporal->>Cache: flush and store generated exercise
            Temporal->>Temporal: publish sets complete_activity
        end
    end
```

`persist_user_request` refuses a second in-progress exercise for the same user unless the idempotency key matches the row already claimed by the usecase. `read_prompt_txt` fills `files/txt/generate_exercise.txt` with the JLPT level and the question counts for that level. `generate` stores the sanitized JSON on the exercise and sets status to generated. `publish` only advances the activity name to `complete_activity`.
