package exercise_workflow

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	"go.temporal.io/sdk/temporal"
)

func TestJLPTLevelNumberQuestionMapping(t *testing.T) {
	for _, tc := range []struct {
		level, listening, reading, grammar int
	}{
		{level: 1, listening: 10, reading: 10, grammar: 10},
		{level: 2, listening: 20, reading: 20, grammar: 20},
		{level: 3, listening: 30, reading: 30, grammar: 30},
		{level: 4, listening: 40, reading: 40, grammar: 40},
		{level: 5, listening: 50, reading: 50, grammar: 50},
		{level: 0},
		{level: 9},
	} {
		listening, reading, grammar := jlptLevelNumberQuestionMapping(tc.level)
		if listening != tc.listening || reading != tc.reading || grammar != tc.grammar {
			t.Fatalf("level %d = %d %d %d, want %d %d %d", tc.level, listening, reading, grammar, tc.listening, tc.reading, tc.grammar)
		}
	}
}

func TestReadPromptTxt(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		t.Chdir(t.TempDir())
		workflow := &exerciseWorkflow{}
		_, err := workflow.readPromptTxt(context.Background(), &ExecutionData{})
		requireCustom(t, err, apperror.ErrInternalServer)
	})

	t.Run("missing request", func(t *testing.T) {
		chdirModuleRoot(t)
		workflow := &exerciseWorkflow{}
		_, err := workflow.readPromptTxt(context.Background(), nil)
		requireInputError(t, err, "generate exercise request is missing")
		_, err = workflow.readPromptTxt(context.Background(), &ExecutionData{})
		requireInputError(t, err, "generate exercise request is missing")
	})

	t.Run("cache failure still signals", func(t *testing.T) {
		chdirModuleRoot(t)
		cache := &fakeExerciseCache{setByKeyErr: errors.New("redis down")}
		execution := &fakeWorkflowExecution{}
		workflow := &exerciseWorkflow{exerciseCache: cache, workflow: execution}
		data := processingExecution("key", "wf-1")
		got, err := workflow.readPromptTxt(context.Background(), data)
		if err != nil {
			t.Fatal(err)
		}
		if got.CurrActivity != generateExerciseActivity || !strings.Contains(got.Prompt, "JLPT N1") || strings.Contains(got.Prompt, "{{") {
			t.Fatalf("prompt activity = %s prompt has placeholders = %v", got.CurrActivity, strings.Contains(got.Prompt, "{{"))
		}
		if len(cache.setKeys) != 1 || cache.setKeys[0] != "key" {
			t.Fatalf("cached keys = %#v", cache.setKeys)
		}
		if execution.signalName != sgProcessing || execution.signaledID != "wf-1" || execution.signalArg != true {
			t.Fatalf("signal = %#v", execution)
		}
		if strings.Count(got.Prompt, "Exactly 10 questions") != 3 {
			t.Fatalf("question counts were not substituted")
		}
	})

	t.Run("signal failure", func(t *testing.T) {
		chdirModuleRoot(t)
		execution := &fakeWorkflowExecution{signalErr: errors.New("signal failed")}
		workflow := &exerciseWorkflow{exerciseCache: &fakeExerciseCache{}, workflow: execution}
		_, err := workflow.readPromptTxt(context.Background(), processingExecution("key", "wf-1"))
		custom := requireCustom(t, err, apperror.ErrInternalServer)
		if !errors.Is(custom, execution.signalErr) {
			t.Fatalf("cause = %v", custom)
		}
	})
}

func TestCacheProcessingExercise(t *testing.T) {
	cache := &fakeExerciseCache{}
	workflow := &exerciseWorkflow{exerciseCache: cache}

	if err := workflow.cacheProcessingExercise(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if err := workflow.cacheProcessingExercise(context.Background(), &ExecutionData{}); err != nil {
		t.Fatal(err)
	}
	if err := workflow.cacheProcessingExercise(context.Background(), &ExecutionData{GenerateExercise: &responses.GenerateExerciseResponse{Status: exerciseStatusCompleted}}); err != nil {
		t.Fatal(err)
	}
	if len(cache.setKeys) != 0 {
		t.Fatal("cache was written for a response that is not processing")
	}

	data := &ExecutionData{GenerateExercise: &responses.GenerateExerciseResponse{Status: exerciseStatusProcessing}}
	if err := workflow.cacheProcessingExercise(context.Background(), data); err != nil || len(cache.setKeys) != 0 {
		t.Fatal("empty key was cached")
	}

	data.UserRequest = &requests.GenerateExerciseRequest{IdempotencyKey: "from-request"}
	if err := workflow.cacheProcessingExercise(context.Background(), data); err != nil {
		t.Fatal(err)
	}
	if len(cache.setKeys) != 1 || cache.setKeys[0] != "from-request" {
		t.Fatalf("cached keys = %#v", cache.setKeys)
	}

	cache.setByKeyErr = errors.New("redis down")
	data.IdempotencyKey = "explicit"
	if err := workflow.cacheProcessingExercise(context.Background(), data); !errors.Is(err, cache.setByKeyErr) {
		t.Fatalf("got %v", err)
	}
}

func TestSignalExerciseEvent(t *testing.T) {
	workflow := &exerciseWorkflow{workflow: &fakeWorkflowExecution{}}
	if err := workflow.signalExerciseEvent(context.Background(), nil, sgProcessing); err == nil {
		t.Fatal("nil execution data was signaled")
	}
	if err := workflow.signalExerciseEvent(context.Background(), &ExecutionData{}, sgProcessing); err == nil {
		t.Fatal("missing workflow id was signaled")
	}

	execution := &fakeWorkflowExecution{signalErr: errors.New("signal failed")}
	workflow.workflow = execution
	err := workflow.signalExerciseEvent(context.Background(), &ExecutionData{WorkflowID: "wf", RunID: "run"}, sgCompleted)
	if !errors.Is(err, execution.signalErr) || execution.signalName != sgCompleted || execution.signalArg != true {
		t.Fatalf("signal result = %v recorded = %#v", err, execution)
	}
}

func TestGenerateActivity(t *testing.T) {
	t.Run("client error", func(t *testing.T) {
		clientErr := errors.New("deepseek down")
		workflow := &exerciseWorkflow{deepSeekClient: &fakeDeepSeek{err: clientErr}}
		_, err := workflow.generate(context.Background(), &ExecutionData{Prompt: "prompt"})
		custom := requireCustom(t, err, apperror.ErrInternalServer)
		if !errors.Is(custom, clientErr) {
			t.Fatalf("cause = %v", custom)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		workflow := &exerciseWorkflow{deepSeekClient: &fakeDeepSeek{result: "not json"}}
		_, err := workflow.generate(context.Background(), &ExecutionData{Exercise: &models.Exercise{}})
		requireCustom(t, err, apperror.ErrInternalServer)
	})

	t.Run("signal error", func(t *testing.T) {
		signalErr := errors.New("signal failed")
		workflow := &exerciseWorkflow{
			deepSeekClient: &fakeDeepSeek{result: `{"title":"N3",}`},
			workflow:       &fakeWorkflowExecution{signalErr: signalErr},
		}
		_, err := workflow.generate(context.Background(), &ExecutionData{WorkflowID: "wf", Exercise: &models.Exercise{}})
		custom := requireCustom(t, err, apperror.ErrInternalServer)
		if !errors.Is(custom, signalErr) {
			t.Fatalf("cause = %v", custom)
		}
	})

	t.Run("success", func(t *testing.T) {
		execution := &fakeWorkflowExecution{}
		workflow := &exerciseWorkflow{
			deepSeekClient: &fakeDeepSeek{result: "```json\n{\"title\":\"N3\"}\n```"},
			workflow:       execution,
		}
		data := &ExecutionData{WorkflowID: "wf", RunID: "run", Prompt: "prompt", Exercise: &models.Exercise{ID: 4}}
		got, err := workflow.generate(context.Background(), data)
		if err != nil {
			t.Fatal(err)
		}
		if got.PromptResult == "" || got.CurrActivity != persistGeneratedResponseActivity {
			t.Fatalf("activity = %s result empty = %v", got.CurrActivity, got.PromptResult == "")
		}
		if string(got.Exercise.Exercise) != `{"title":"N3"}` || string(got.Exercise.Answer) != `{}` || got.Exercise.Status != enums.GeneratedStatus || got.Exercise.UpdatedAt.IsZero() {
			t.Fatalf("exercise = %#v", got.Exercise)
		}
		if got.GenerateExercise == nil || execution.signalName != sgCompleted || execution.signaledID != "wf" {
			t.Fatalf("response = %#v signal = %s", got.GenerateExercise, execution.signalName)
		}
	})
}

func TestPersistUserRequest(t *testing.T) {
	workflow := &exerciseWorkflow{userRepository: &fakeUserRepository{}, exerciseRepository: &fakeExerciseRepository{}}
	_, err := workflow.persistUserRequest(context.Background(), nil)
	requireInputError(t, err, "user request is missing")
	_, err = workflow.persistUserRequest(context.Background(), &ExecutionData{})
	requireInputError(t, err, "user request is missing")
	_, err = workflow.persistUserRequest(context.Background(), &ExecutionData{UserRequest: &requests.GenerateExerciseRequest{}})
	requireInputError(t, err, "user id is missing")

	findErr := errors.New("user missing")
	workflow.userRepository = &fakeUserRepository{findErr: findErr}
	_, err = workflow.persistUserRequest(context.Background(), &ExecutionData{UserID: 4, UserRequest: &requests.GenerateExerciseRequest{IdempotencyKey: "key"}})
	custom := requireCustom(t, err, apperror.ErrInternalServer)
	if !errors.Is(custom, findErr) {
		t.Fatalf("cause = %v", custom)
	}

	latestErr := errors.New("db down")
	workflow.userRepository = &fakeUserRepository{user: &models.User{ID: 4}}
	workflow.exerciseRepository = &fakeExerciseRepository{latestErr: latestErr}
	_, err = workflow.persistUserRequest(context.Background(), &ExecutionData{UserID: 4, UserRequest: &requests.GenerateExerciseRequest{IdempotencyKey: "key"}})
	if !errors.Is(err, latestErr) {
		t.Fatalf("got %v", err)
	}

	repo := &fakeExerciseRepository{latest: &models.Exercise{ID: 9, ReferenceID: "same", Status: enums.GeneratedStatus}}
	workflow.exerciseRepository = repo
	got, err := workflow.persistUserRequest(context.Background(), &ExecutionData{UserID: 4, UserRequest: &requests.GenerateExerciseRequest{IdempotencyKey: "same", JLPTNLevel: 3}})
	if err != nil || got.Exercise.ID != 9 || repo.saved != nil {
		t.Fatalf("same key result = %#v err = %v saved = %#v", got, err, repo.saved)
	}

	repo.latest.ReferenceID = "other"
	_, err = workflow.persistUserRequest(context.Background(), &ExecutionData{UserID: 4, UserRequest: &requests.GenerateExerciseRequest{IdempotencyKey: "new"}})
	requireInputError(t, err, "user already has an exercise in progress")

	repo.latest.Status = enums.SubmittedStatus
	repo.saveErr = errors.New("insert failed")
	_, err = workflow.persistUserRequest(context.Background(), &ExecutionData{UserID: 4, UserRequest: &requests.GenerateExerciseRequest{IdempotencyKey: "new", JLPTNLevel: 2}})
	if !errors.Is(err, repo.saveErr) {
		t.Fatalf("got %v", err)
	}

	repo.saveErr = nil
	repo.latestErr = sql.ErrNoRows
	repo.latest = nil
	got, err = workflow.persistUserRequest(context.Background(), &ExecutionData{UserID: 4, UserRequest: &requests.GenerateExerciseRequest{IdempotencyKey: "new", JLPTNLevel: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if got.User.ID != 4 || got.Exercise.ID != 21 || got.Exercise.Status != enums.InGeneratedStatus || got.Exercise.ReferenceID != "new" || got.Exercise.NLevel != 2 || got.Exercise.UserID != 4 {
		t.Fatalf("saved execution = %#v exercise = %#v", got, got.Exercise)
	}
}

func TestPersistExerciseGeneratedResponse(t *testing.T) {
	workflow := &exerciseWorkflow{exerciseRepository: &fakeExerciseRepository{}, exerciseCache: &fakeExerciseCache{}}
	_, err := workflow.persistExerciseGeneratedResponse(context.Background(), nil)
	requireInputError(t, err, "generated exercise is missing")
	_, err = workflow.persistExerciseGeneratedResponse(context.Background(), &ExecutionData{})
	requireInputError(t, err, "generated exercise is missing")
	_, err = workflow.persistExerciseGeneratedResponse(context.Background(), &ExecutionData{GenerateExercise: &responses.GenerateExerciseResponse{}})
	requireInputError(t, err, "idempotency key is missing")
	_, err = workflow.persistExerciseGeneratedResponse(context.Background(), &ExecutionData{
		GenerateExercise: &responses.GenerateExerciseResponse{},
		UserRequest:      &requests.GenerateExerciseRequest{IdempotencyKey: "key"},
	})
	requireInputError(t, err, "exercise is missing")

	updateErr := errors.New("update failed")
	repo := &fakeExerciseRepository{updateErr: updateErr}
	workflow.exerciseRepository = repo
	_, err = workflow.persistExerciseGeneratedResponse(context.Background(), readyGeneratedExecution("key"))
	if !errors.Is(err, updateErr) {
		t.Fatalf("got %v", err)
	}

	cache := &fakeExerciseCache{deleteErr: errors.New("delete failed")}
	workflow.exerciseRepository = &fakeExerciseRepository{}
	workflow.exerciseCache = cache
	_, err = workflow.persistExerciseGeneratedResponse(context.Background(), readyGeneratedExecution("key"))
	if !errors.Is(err, cache.deleteErr) || len(cache.deletedKeys) != 0 {
		t.Fatalf("err = %v deleted keys = %#v", err, cache.deletedKeys)
	}

	cache = &fakeExerciseCache{deleteByKeyErr: errors.New("delete key failed")}
	workflow.exerciseCache = cache
	_, err = workflow.persistExerciseGeneratedResponse(context.Background(), readyGeneratedExecution("key"))
	if !errors.Is(err, cache.deleteByKeyErr) || len(cache.setIDs) != 0 {
		t.Fatalf("err = %v set ids = %#v", err, cache.setIDs)
	}

	cache = &fakeExerciseCache{setErr: errors.New("set failed")}
	workflow.exerciseCache = cache
	_, err = workflow.persistExerciseGeneratedResponse(context.Background(), readyGeneratedExecution("key"))
	if !errors.Is(err, cache.setErr) || len(cache.setKeys) != 0 {
		t.Fatalf("err = %v set keys = %#v", err, cache.setKeys)
	}

	cache = &fakeExerciseCache{setByKeyErr: errors.New("set key failed")}
	workflow.exerciseCache = cache
	_, err = workflow.persistExerciseGeneratedResponse(context.Background(), readyGeneratedExecution("key"))
	if !errors.Is(err, cache.setByKeyErr) {
		t.Fatalf("got %v", err)
	}

	created := time.Date(2026, 10, 21, 1, 2, 3, 0, time.UTC)
	cache = &fakeExerciseCache{}
	workflow.exerciseCache = cache
	data := readyGeneratedExecution("")
	data.UserRequest = &requests.GenerateExerciseRequest{IdempotencyKey: "from-request"}
	data.Exercise.CreatedAt = created
	got, err := workflow.persistExerciseGeneratedResponse(context.Background(), data)
	if err != nil {
		t.Fatal(err)
	}
	if got.CurrActivity != publishActivity || got.IdempotencyKey != "from-request" || got.GenerateExercise.ID != data.Exercise.ID || got.GenerateExercise.JLPTLevel != "3" || !got.GenerateExercise.CreatedAt.Equal(created) {
		t.Fatalf("persisted = %#v response = %#v", got, got.GenerateExercise)
	}
	if len(cache.deletedIDs) != 1 || len(cache.deletedKeys) != 1 || cache.deletedKeys[0] != "from-request" || len(cache.setKeys) != 1 {
		t.Fatalf("cache calls ids=%#v keys=%#v set=%#v", cache.deletedIDs, cache.deletedKeys, cache.setKeys)
	}
}

func TestPublishSetsCompleteActivity(t *testing.T) {
	got, err := (&exerciseWorkflow{}).publish(context.Background(), &ExecutionData{})
	if err != nil || got.CurrActivity != completeActivity {
		t.Fatalf("got %#v err %v", got, err)
	}
}

func TestGenerateExerciseResponseFromExercise(t *testing.T) {
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	updated := created.Add(time.Hour)
	got := generateExerciseResponseFromExercise(&models.Exercise{
		ID:        7,
		NLevel:    4,
		Status:    enums.GeneratedStatus,
		Exercise:  []byte(`{"title":"N4"}`),
		CreatedAt: created,
		UpdatedAt: updated,
	})
	if got.ID != 7 || got.JLPTLevel != "4" || got.Status != enums.GeneratedStatus || string(got.Exercise) != `{"title":"N4"}` || !got.CreatedAt.Equal(created) || !got.UpdatedAt.Equal(updated) {
		t.Fatalf("mapped = %#v", got)
	}
}

func processingExecution(key, workflowID string) *ExecutionData {
	return &ExecutionData{
		WorkflowID:     workflowID,
		IdempotencyKey: key,
		UserRequest:    &requests.GenerateExerciseRequest{IdempotencyKey: key, JLPTNLevel: 1},
		GenerateExercise: &responses.GenerateExerciseResponse{
			Status: exerciseStatusProcessing,
		},
	}
}

func readyGeneratedExecution(key string) *ExecutionData {
	return &ExecutionData{
		IdempotencyKey:   key,
		GenerateExercise: &responses.GenerateExerciseResponse{Status: exerciseStatusCompleted},
		Exercise: &models.Exercise{
			ID:       15,
			NLevel:   3,
			Status:   enums.GeneratedStatus,
			Exercise: []byte(`{"title":"N3"}`),
		},
	}
}

func chdirModuleRoot(t *testing.T) {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			t.Chdir(dir)
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root not found")
		}
		dir = parent
	}
}

func requireCustom(t *testing.T, err error, kind apperror.ErrType) *apperror.CustomError {
	t.Helper()
	custom, ok := err.(*apperror.CustomError)
	if !ok || custom == nil || custom.ErrType != kind {
		t.Fatalf("got %#v, want custom error %v", err, kind)
	}
	return custom
}

func requireInputError(t *testing.T, err error, message string) {
	t.Helper()
	var appErr *temporal.ApplicationError
	if !errors.As(err, &appErr) || appErr.Type() != exerciseInputErrorType || !appErr.NonRetryable() || appErr.Message() != message {
		t.Fatalf("got %v, want non-retryable %s %q", err, exerciseInputErrorType, message)
	}
}
