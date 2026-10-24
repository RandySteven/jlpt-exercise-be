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
	cachemocks "github.com/RandySteven/jlpt-exercise-be/caches/mocks"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	deepseekmocks "github.com/RandySteven/jlpt-exercise-be/pkg/deepseek/mocks"
	repomocks "github.com/RandySteven/jlpt-exercise-be/repositories/mocks"
	"github.com/stretchr/testify/mock"
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
		var keys []string
		cache := cachemocks.NewMockExerciseCache(t)
		cache.EXPECT().SetExerciseByIdempotencyKey(mock.Anything, mock.Anything, mock.Anything).
			Run(func(_ context.Context, key string, _ *responses.GenerateExerciseResponse) {
				keys = append(keys, key)
			}).Return(errors.New("redis down"))
		execution := newWorkflowDouble(t)
		execution.allowSignal(nil)
		workflow := &exerciseWorkflow{exerciseCache: cache, workflow: execution.exec}
		data := processingExecution("key", "wf-1")
		got, err := workflow.readPromptTxt(context.Background(), data)
		if err != nil {
			t.Fatal(err)
		}
		if got.CurrActivity != generateExerciseActivity || !strings.Contains(got.Prompt, "JLPT N1") || strings.Contains(got.Prompt, "{{") {
			t.Fatalf("prompt activity = %s prompt has placeholders = %v", got.CurrActivity, strings.Contains(got.Prompt, "{{"))
		}
		if len(keys) != 1 || keys[0] != "key" {
			t.Fatalf("cached keys = %#v", keys)
		}
		if execution.signalName != sgProcessing || execution.signaledID != "wf-1" || execution.signalArg != true {
			t.Fatalf("signal name=%s id=%s arg=%#v", execution.signalName, execution.signaledID, execution.signalArg)
		}
		if strings.Count(got.Prompt, "Exactly 10 questions") != 3 {
			t.Fatalf("question counts were not substituted")
		}
	})

	t.Run("signal failure", func(t *testing.T) {
		chdirModuleRoot(t)
		signalErr := errors.New("signal failed")
		cache := cachemocks.NewMockExerciseCache(t)
		cache.EXPECT().SetExerciseByIdempotencyKey(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
		execution := newWorkflowDouble(t)
		execution.allowSignal(signalErr)
		workflow := &exerciseWorkflow{exerciseCache: cache, workflow: execution.exec}
		_, err := workflow.readPromptTxt(context.Background(), processingExecution("key", "wf-1"))
		custom := requireCustom(t, err, apperror.ErrInternalServer)
		if !errors.Is(custom, signalErr) {
			t.Fatalf("cause = %v", custom)
		}
	})
}

func TestCacheProcessingExercise(t *testing.T) {
	var setErr error
	var keys []string
	cache := cachemocks.NewMockExerciseCache(t)
	cache.EXPECT().SetExerciseByIdempotencyKey(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, key string, _ *responses.GenerateExerciseResponse) error {
			keys = append(keys, key)
			return setErr
		}).Maybe()
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
	if len(keys) != 0 {
		t.Fatal("cache was written for a response that is not processing")
	}

	data := &ExecutionData{GenerateExercise: &responses.GenerateExerciseResponse{Status: exerciseStatusProcessing}}
	if err := workflow.cacheProcessingExercise(context.Background(), data); err != nil || len(keys) != 0 {
		t.Fatal("empty key was cached")
	}

	data.UserRequest = &requests.GenerateExerciseRequest{IdempotencyKey: "from-request"}
	if err := workflow.cacheProcessingExercise(context.Background(), data); err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 || keys[0] != "from-request" {
		t.Fatalf("cached keys = %#v", keys)
	}

	setErr = errors.New("redis down")
	data.IdempotencyKey = "explicit"
	if err := workflow.cacheProcessingExercise(context.Background(), data); !errors.Is(err, setErr) {
		t.Fatalf("got %v", err)
	}
}

func TestSignalExerciseEvent(t *testing.T) {
	workflow := &exerciseWorkflow{workflow: newWorkflowDouble(t).exec}
	if err := workflow.signalExerciseEvent(context.Background(), nil, sgProcessing); err == nil {
		t.Fatal("nil execution data was signaled")
	}
	if err := workflow.signalExerciseEvent(context.Background(), &ExecutionData{}, sgProcessing); err == nil {
		t.Fatal("missing workflow id was signaled")
	}

	signalErr := errors.New("signal failed")
	execution := newWorkflowDouble(t)
	execution.allowSignal(signalErr)
	workflow.workflow = execution.exec
	err := workflow.signalExerciseEvent(context.Background(), &ExecutionData{WorkflowID: "wf", RunID: "run"}, sgCompleted)
	if !errors.Is(err, signalErr) || execution.signalName != sgCompleted || execution.signalArg != true {
		t.Fatalf("signal result = %v name=%s arg=%#v", err, execution.signalName, execution.signalArg)
	}
}

func TestGenerateActivity(t *testing.T) {
	t.Run("client error", func(t *testing.T) {
		clientErr := errors.New("deepseek down")
		client := deepseekmocks.NewMockDeepSeek(t)
		client.EXPECT().ChatGenerate(mock.Anything, mock.Anything).Return("", clientErr)
		workflow := &exerciseWorkflow{deepSeekClient: client}
		_, err := workflow.generate(context.Background(), &ExecutionData{Prompt: "prompt"})
		custom := requireCustom(t, err, apperror.ErrInternalServer)
		if !errors.Is(custom, clientErr) {
			t.Fatalf("cause = %v", custom)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		client := deepseekmocks.NewMockDeepSeek(t)
		client.EXPECT().ChatGenerate(mock.Anything, mock.Anything).Return("not json", nil)
		workflow := &exerciseWorkflow{deepSeekClient: client}
		_, err := workflow.generate(context.Background(), &ExecutionData{Exercise: &models.Exercise{}})
		requireCustom(t, err, apperror.ErrInternalServer)
	})

	t.Run("signal error", func(t *testing.T) {
		signalErr := errors.New("signal failed")
		client := deepseekmocks.NewMockDeepSeek(t)
		client.EXPECT().ChatGenerate(mock.Anything, mock.Anything).Return(`{"title":"N3",}`, nil)
		execution := newWorkflowDouble(t)
		execution.allowSignal(signalErr)
		workflow := &exerciseWorkflow{deepSeekClient: client, workflow: execution.exec}
		_, err := workflow.generate(context.Background(), &ExecutionData{WorkflowID: "wf", Exercise: &models.Exercise{}})
		custom := requireCustom(t, err, apperror.ErrInternalServer)
		if !errors.Is(custom, signalErr) {
			t.Fatalf("cause = %v", custom)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := deepseekmocks.NewMockDeepSeek(t)
		client.EXPECT().ChatGenerate(mock.Anything, "prompt").Return("```json\n{\"title\":\"N3\"}\n```", nil)
		execution := newWorkflowDouble(t)
		execution.allowSignal(nil)
		workflow := &exerciseWorkflow{deepSeekClient: client, workflow: execution.exec}
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
	users := repomocks.NewMockUserRepository(t)
	exercises := repomocks.NewMockExerciseRepository(t)
	workflow := &exerciseWorkflow{userRepository: users, exerciseRepository: exercises}
	_, err := workflow.persistUserRequest(context.Background(), nil)
	requireInputError(t, err, "user request is missing")
	_, err = workflow.persistUserRequest(context.Background(), &ExecutionData{})
	requireInputError(t, err, "user request is missing")
	_, err = workflow.persistUserRequest(context.Background(), &ExecutionData{UserRequest: &requests.GenerateExerciseRequest{}})
	requireInputError(t, err, "user id is missing")

	findErr := errors.New("user missing")
	users.EXPECT().FindByID(mock.Anything, uint64(4)).Return(nil, findErr).Once()
	_, err = workflow.persistUserRequest(context.Background(), &ExecutionData{UserID: 4, UserRequest: &requests.GenerateExerciseRequest{IdempotencyKey: "key"}})
	custom := requireCustom(t, err, apperror.ErrInternalServer)
	if !errors.Is(custom, findErr) {
		t.Fatalf("cause = %v", custom)
	}

	users.EXPECT().FindByID(mock.Anything, uint64(4)).Return(&models.User{ID: 4}, nil)
	latestErr := errors.New("db down")
	exercises.EXPECT().FindLatestByUserID(mock.Anything, uint64(4)).Return(nil, latestErr).Once()
	_, err = workflow.persistUserRequest(context.Background(), &ExecutionData{UserID: 4, UserRequest: &requests.GenerateExerciseRequest{IdempotencyKey: "key"}})
	if !errors.Is(err, latestErr) {
		t.Fatalf("got %v", err)
	}

	var saved *models.Exercise
	exercises.EXPECT().FindLatestByUserID(mock.Anything, uint64(4)).Return(&models.Exercise{ID: 9, ReferenceID: "same", Status: enums.GeneratedStatus}, nil).Once()
	got, err := workflow.persistUserRequest(context.Background(), &ExecutionData{UserID: 4, UserRequest: &requests.GenerateExerciseRequest{IdempotencyKey: "same", JLPTNLevel: 3}})
	if err != nil || got.Exercise.ID != 9 || saved != nil {
		t.Fatalf("same key result = %#v err = %v saved = %#v", got, err, saved)
	}

	exercises.EXPECT().FindLatestByUserID(mock.Anything, uint64(4)).Return(&models.Exercise{ID: 9, ReferenceID: "other", Status: enums.GeneratedStatus}, nil).Once()
	_, err = workflow.persistUserRequest(context.Background(), &ExecutionData{UserID: 4, UserRequest: &requests.GenerateExerciseRequest{IdempotencyKey: "new"}})
	requireInputError(t, err, "user already has an exercise in progress")

	saveErr := errors.New("insert failed")
	exercises.EXPECT().FindLatestByUserID(mock.Anything, uint64(4)).Return(&models.Exercise{ID: 9, ReferenceID: "other", Status: enums.SubmittedStatus}, nil).Once()
	exercises.EXPECT().Save(mock.Anything, mock.Anything).Return(nil, saveErr).Once()
	_, err = workflow.persistUserRequest(context.Background(), &ExecutionData{UserID: 4, UserRequest: &requests.GenerateExerciseRequest{IdempotencyKey: "new", JLPTNLevel: 2}})
	if !errors.Is(err, saveErr) {
		t.Fatalf("got %v", err)
	}

	exercises.EXPECT().FindLatestByUserID(mock.Anything, uint64(4)).Return(nil, sql.ErrNoRows).Once()
	exercises.EXPECT().Save(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, entity *models.Exercise) (*models.Exercise, error) {
		copied := *entity
		copied.ID = 21
		saved = &copied
		return saved, nil
	}).Once()
	got, err = workflow.persistUserRequest(context.Background(), &ExecutionData{UserID: 4, UserRequest: &requests.GenerateExerciseRequest{IdempotencyKey: "new", JLPTNLevel: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if got.User.ID != 4 || got.Exercise.ID != 21 || got.Exercise.Status != enums.InGeneratedStatus || got.Exercise.ReferenceID != "new" || got.Exercise.NLevel != 2 || got.Exercise.UserID != 4 {
		t.Fatalf("saved execution = %#v exercise = %#v", got, got.Exercise)
	}
}

func TestPersistExerciseGeneratedResponse(t *testing.T) {
	workflow := &exerciseWorkflow{
		exerciseRepository: repomocks.NewMockExerciseRepository(t),
		exerciseCache:      cachemocks.NewMockExerciseCache(t),
	}
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
	exercises := repomocks.NewMockExerciseRepository(t)
	exercises.EXPECT().Update(mock.Anything, mock.Anything).Return(nil, updateErr)
	workflow.exerciseRepository = exercises
	_, err = workflow.persistExerciseGeneratedResponse(context.Background(), readyGeneratedExecution("key"))
	if !errors.Is(err, updateErr) {
		t.Fatalf("got %v", err)
	}

	deleteErr := errors.New("delete failed")
	exercises, cache := updatingExerciseMocks(t)
	cache.EXPECT().DeleteExercise(mock.Anything, mock.Anything).Return(deleteErr)
	workflow.exerciseRepository = exercises
	workflow.exerciseCache = cache
	_, err = workflow.persistExerciseGeneratedResponse(context.Background(), readyGeneratedExecution("key"))
	if !errors.Is(err, deleteErr) {
		t.Fatalf("got %v", err)
	}
	cache.AssertNotCalled(t, "DeleteExerciseByIdempotencyKey", mock.Anything, mock.Anything)

	deleteByKeyErr := errors.New("delete key failed")
	exercises, cache = updatingExerciseMocks(t)
	cache.EXPECT().DeleteExercise(mock.Anything, mock.Anything).Return(nil)
	cache.EXPECT().DeleteExerciseByIdempotencyKey(mock.Anything, mock.Anything).Return(deleteByKeyErr)
	workflow.exerciseRepository = exercises
	workflow.exerciseCache = cache
	_, err = workflow.persistExerciseGeneratedResponse(context.Background(), readyGeneratedExecution("key"))
	if !errors.Is(err, deleteByKeyErr) {
		t.Fatalf("got %v", err)
	}
	cache.AssertNotCalled(t, "SetExercise", mock.Anything, mock.Anything, mock.Anything)

	setErr := errors.New("set failed")
	exercises, cache = updatingExerciseMocks(t)
	cache.EXPECT().DeleteExercise(mock.Anything, mock.Anything).Return(nil)
	cache.EXPECT().DeleteExerciseByIdempotencyKey(mock.Anything, mock.Anything).Return(nil)
	cache.EXPECT().SetExercise(mock.Anything, mock.Anything, mock.Anything).Return(setErr)
	workflow.exerciseRepository = exercises
	workflow.exerciseCache = cache
	_, err = workflow.persistExerciseGeneratedResponse(context.Background(), readyGeneratedExecution("key"))
	if !errors.Is(err, setErr) {
		t.Fatalf("got %v", err)
	}
	cache.AssertNotCalled(t, "SetExerciseByIdempotencyKey", mock.Anything, mock.Anything, mock.Anything)

	setByKeyErr := errors.New("set key failed")
	exercises, cache = updatingExerciseMocks(t)
	cache.EXPECT().DeleteExercise(mock.Anything, mock.Anything).Return(nil)
	cache.EXPECT().DeleteExerciseByIdempotencyKey(mock.Anything, mock.Anything).Return(nil)
	cache.EXPECT().SetExercise(mock.Anything, mock.Anything, mock.Anything).Return(nil)
	cache.EXPECT().SetExerciseByIdempotencyKey(mock.Anything, mock.Anything, mock.Anything).Return(setByKeyErr)
	workflow.exerciseRepository = exercises
	workflow.exerciseCache = cache
	_, err = workflow.persistExerciseGeneratedResponse(context.Background(), readyGeneratedExecution("key"))
	if !errors.Is(err, setByKeyErr) {
		t.Fatalf("got %v", err)
	}

	var deletedIDs []uint64
	var deletedKeys []string
	var setKeys []string
	exercises, cache = updatingExerciseMocks(t)
	cache.EXPECT().DeleteExercise(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, id uint64) error {
		deletedIDs = append(deletedIDs, id)
		return nil
	})
	cache.EXPECT().DeleteExerciseByIdempotencyKey(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, key string) error {
		deletedKeys = append(deletedKeys, key)
		return nil
	})
	cache.EXPECT().SetExercise(mock.Anything, mock.Anything, mock.Anything).Return(nil)
	cache.EXPECT().SetExerciseByIdempotencyKey(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, key string, _ *responses.GenerateExerciseResponse) error {
		setKeys = append(setKeys, key)
		return nil
	})
	workflow.exerciseRepository = exercises
	workflow.exerciseCache = cache
	created := time.Date(2026, 10, 21, 1, 2, 3, 0, time.UTC)
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
	if len(deletedIDs) != 1 || len(deletedKeys) != 1 || deletedKeys[0] != "from-request" || len(setKeys) != 1 {
		t.Fatalf("cache calls ids=%#v keys=%#v set=%#v", deletedIDs, deletedKeys, setKeys)
	}
}

func updatingExerciseMocks(t *testing.T) (*repomocks.MockExerciseRepository, *cachemocks.MockExerciseCache) {
	t.Helper()
	exercises := repomocks.NewMockExerciseRepository(t)
	exercises.EXPECT().Update(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, entity *models.Exercise) (*models.Exercise, error) {
		copied := *entity
		return &copied, nil
	})
	return exercises, cachemocks.NewMockExerciseCache(t)
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
