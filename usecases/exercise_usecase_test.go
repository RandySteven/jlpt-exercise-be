package usecases

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	cachemocks "github.com/RandySteven/jlpt-exercise-be/caches/mocks"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	"github.com/RandySteven/jlpt-exercise-be/logic/exercise_workflow"
	workflowmocks "github.com/RandySteven/jlpt-exercise-be/logic/exercise_workflow/mocks"
	repomocks "github.com/RandySteven/jlpt-exercise-be/repositories/mocks"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
)

func TestGenerateExerciseBlocksOpenExercise(t *testing.T) {
	for _, status := range []string{enums.InGeneratedStatus, enums.GeneratedStatus} {
		t.Run(status, func(t *testing.T) {
			repo := repomocks.NewMockExerciseRepository(t)
			repo.EXPECT().FindLatestByUserID(mock.Anything, uint64(4)).Return(&models.Exercise{
				ID: 9, ReferenceID: "open-key", Status: status, UserID: 4, NLevel: 3,
			}, nil)
			workflow := workflowmocks.NewMockExerciseWorkflow(t)
			usecase := newGenerateExerciseUsecase(t, repo, cachemocks.NewMockExerciseCache(t), workflow)

			_, customErr := usecase.GenerateExercise(userContext(4), &requests.GenerateExerciseRequest{
				IdempotencyKey: "new-key",
				JLPTNLevel:     3,
			})
			if customErr == nil || customErr.ErrType != apperror.ErrConflict {
				t.Fatalf("status %s: got %#v, want conflict", status, customErr)
			}
			workflow.AssertNotCalled(t, "GenerateExercise", mock.Anything, mock.Anything)
			repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
		})
	}
}

func TestGenerateExerciseAllowsFinishedExercise(t *testing.T) {
	for _, status := range []string{enums.SubmittedStatus, enums.CompletedStatus, enums.FailedStatus} {
		t.Run(status, func(t *testing.T) {
			repo := repomocks.NewMockExerciseRepository(t)
			repo.EXPECT().FindLatestByUserID(mock.Anything, uint64(4)).Return(&models.Exercise{
				ID: 9, ReferenceID: "done-key", Status: status, UserID: 4,
			}, nil)
			var saved *models.Exercise
			repo.EXPECT().Save(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, entity *models.Exercise) (*models.Exercise, error) {
				copied := *entity
				copied.ID = 21
				saved = &copied
				return saved, nil
			})
			workflow := workflowmocks.NewMockExerciseWorkflow(t)
			workflow.EXPECT().GenerateExercise(mock.Anything, mock.Anything).Return(&responses.GenerateExerciseResponse{Status: enums.InGeneratedStatus}, nil)
			usecase := newGenerateExerciseUsecase(t, repo, cachemocks.NewMockExerciseCache(t), workflow)

			response, customErr := usecase.GenerateExercise(userContext(4), &requests.GenerateExerciseRequest{
				IdempotencyKey: "new-key",
				JLPTNLevel:     2,
			})
			if customErr != nil {
				t.Fatal(customErr)
			}
			if saved == nil || saved.Status != enums.InGeneratedStatus {
				t.Fatal("finished exercise did not claim a new generation")
			}
			if response.ID != saved.ID {
				t.Fatalf("response id = %d, want claimed id %d", response.ID, saved.ID)
			}
		})
	}
}

func TestGenerateExerciseReturnsSameOpenExercise(t *testing.T) {
	repo := repomocks.NewMockExerciseRepository(t)
	repo.EXPECT().FindLatestByUserID(mock.Anything, uint64(4)).Return(&models.Exercise{
		ID: 9, ReferenceID: "same-key", Status: enums.GeneratedStatus, UserID: 4, NLevel: 5,
	}, nil)
	workflow := workflowmocks.NewMockExerciseWorkflow(t)
	usecase := newGenerateExerciseUsecase(t, repo, cachemocks.NewMockExerciseCache(t), workflow)

	response, customErr := usecase.GenerateExercise(userContext(4), &requests.GenerateExerciseRequest{
		IdempotencyKey: "same-key",
		JLPTNLevel:     5,
	})
	if customErr != nil {
		t.Fatal(customErr)
	}
	if response.ID != 9 {
		t.Fatalf("same key restarted generation: response=%#v", response)
	}
	workflow.AssertNotCalled(t, "GenerateExercise", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestGenerateExerciseReleasesClaimWhenWorkflowDoesNotStart(t *testing.T) {
	repo := repomocks.NewMockExerciseRepository(t)
	repo.EXPECT().FindLatestByUserID(mock.Anything, uint64(4)).Return(nil, sql.ErrNoRows)
	repo.EXPECT().Save(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, entity *models.Exercise) (*models.Exercise, error) {
		copied := *entity
		copied.ID = 21
		return &copied, nil
	})
	repo.EXPECT().UpdateStatusIfCurrent(mock.Anything, uint64(21), enums.InGeneratedStatus, enums.FailedStatus, mock.Anything).Return(true, nil)
	workflow := workflowmocks.NewMockExerciseWorkflow(t)
	workflow.EXPECT().GenerateExercise(mock.Anything, mock.Anything).Return(nil, apperror.NewCustomError(apperror.ErrInternalServer, "start failed", exercise_workflow.ErrWorkflowNotStarted))
	usecase := newGenerateExerciseUsecase(t, repo, cachemocks.NewMockExerciseCache(t), workflow)

	_, customErr := usecase.GenerateExercise(userContext(4), &requests.GenerateExerciseRequest{IdempotencyKey: "key", JLPTNLevel: 1})
	if customErr == nil {
		t.Fatal("expected workflow error")
	}
}

func TestGenerateExerciseConflictOnConcurrentClaim(t *testing.T) {
	repo := repomocks.NewMockExerciseRepository(t)
	repo.EXPECT().FindLatestByUserID(mock.Anything, uint64(4)).Return(nil, sql.ErrNoRows)
	repo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil, &pgconn.PgError{Code: "23505"})
	workflow := workflowmocks.NewMockExerciseWorkflow(t)
	usecase := newGenerateExerciseUsecase(t, repo, cachemocks.NewMockExerciseCache(t), workflow)

	_, customErr := usecase.GenerateExercise(userContext(4), &requests.GenerateExerciseRequest{IdempotencyKey: "key", JLPTNLevel: 1})
	if customErr == nil || customErr.ErrType != apperror.ErrConflict {
		t.Fatalf("got %#v, want conflict", customErr)
	}
	workflow.AssertNotCalled(t, "GenerateExercise", mock.Anything, mock.Anything)
}

func TestListExerciseRequiresRequest(t *testing.T) {
	usecase := newGenerateExerciseUsecase(t, repomocks.NewMockExerciseRepository(t), cachemocks.NewMockExerciseCache(t), workflowmocks.NewMockExerciseWorkflow(t))

	_, customErr := usecase.ListExercise(userContext(4), nil)
	if customErr == nil || customErr.ErrType != apperror.ErrBadRequest {
		t.Fatalf("got %#v, want bad request", customErr)
	}
}

func TestListExerciseRequiresUser(t *testing.T) {
	usecase := newGenerateExerciseUsecase(t, repomocks.NewMockExerciseRepository(t), cachemocks.NewMockExerciseCache(t), workflowmocks.NewMockExerciseWorkflow(t))
	request := &requests.ListExerciseRequest{Page: 1, PageSize: 10}

	for _, tc := range []struct {
		name string
		ctx  context.Context
	}{
		{name: "missing", ctx: context.Background()},
		{name: "zero", ctx: userContext(0)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, customErr := usecase.ListExercise(tc.ctx, request)
			if customErr == nil || customErr.ErrType != apperror.ErrUnauthorized {
				t.Fatalf("got %#v, want unauthorized", customErr)
			}
		})
	}
}

func TestListExerciseReturnsMappedPage(t *testing.T) {
	created := time.Date(2026, 10, 19, 1, 2, 3, 0, time.UTC)
	updated := created.Add(time.Hour)
	repo := repomocks.NewMockExerciseRepository(t)
	repo.EXPECT().FindAllByUserID(mock.Anything, uint64(4), int64(1), int64(10)).Return([]*models.Exercise{{
		ID:        11,
		NLevel:    3,
		Status:    enums.GeneratedStatus,
		Exercise:  json.RawMessage(`{"q":1}`),
		CreatedAt: created,
		UpdatedAt: updated,
	}}, nil)
	repo.EXPECT().CountByUserID(mock.Anything, uint64(4)).Return(int64(4), nil)
	usecase := newGenerateExerciseUsecase(t, repo, cachemocks.NewMockExerciseCache(t), workflowmocks.NewMockExerciseWorkflow(t))

	response, customErr := usecase.ListExercise(userContext(4), &requests.ListExerciseRequest{Page: 2, PageSize: 10})
	if customErr != nil {
		t.Fatal(customErr)
	}
	if response.TotalItems != 4 || response.Page != 2 || response.PageSize != 10 || len(response.Exercises) != 1 {
		t.Fatalf("page response = %#v", response)
	}
	got := response.Exercises[0]
	if got.ID != 11 || got.Status != enums.GeneratedStatus || got.JLPTLevel != "3" || !got.CreatedAt.Equal(created) || !got.UpdatedAt.Equal(updated) {
		t.Fatalf("mapped exercise = %#v", got)
	}
}

func TestListExerciseEmptyPage(t *testing.T) {
	repo := repomocks.NewMockExerciseRepository(t)
	repo.EXPECT().FindAllByUserID(mock.Anything, uint64(4), int64(0), int64(20)).Return([]*models.Exercise{}, nil)
	repo.EXPECT().CountByUserID(mock.Anything, uint64(4)).Return(int64(0), nil)
	usecase := newGenerateExerciseUsecase(t, repo, cachemocks.NewMockExerciseCache(t), workflowmocks.NewMockExerciseWorkflow(t))

	response, customErr := usecase.ListExercise(userContext(4), &requests.ListExerciseRequest{Page: 1, PageSize: 20})
	if customErr != nil {
		t.Fatal(customErr)
	}
	if response.TotalItems != 0 || len(response.Exercises) != 0 || response.Page != 1 || response.PageSize != 20 {
		t.Fatalf("empty page = %#v", response)
	}
}

func TestListExerciseFindError(t *testing.T) {
	repo := repomocks.NewMockExerciseRepository(t)
	repo.EXPECT().FindAllByUserID(mock.Anything, uint64(4), int64(0), int64(10)).Return(nil, errors.New("db down"))
	usecase := newGenerateExerciseUsecase(t, repo, cachemocks.NewMockExerciseCache(t), workflowmocks.NewMockExerciseWorkflow(t))

	_, customErr := usecase.ListExercise(userContext(4), &requests.ListExerciseRequest{Page: 1, PageSize: 10})
	if customErr == nil || customErr.ErrType != apperror.ErrInternalServer {
		t.Fatalf("got %#v, want internal server", customErr)
	}
	repo.AssertNotCalled(t, "CountByUserID", mock.Anything, mock.Anything)
}

func TestListExerciseCountError(t *testing.T) {
	repo := repomocks.NewMockExerciseRepository(t)
	repo.EXPECT().FindAllByUserID(mock.Anything, uint64(4), int64(0), int64(10)).Return([]*models.Exercise{}, nil)
	repo.EXPECT().CountByUserID(mock.Anything, uint64(4)).Return(int64(0), errors.New("count failed"))
	usecase := newGenerateExerciseUsecase(t, repo, cachemocks.NewMockExerciseCache(t), workflowmocks.NewMockExerciseWorkflow(t))

	_, customErr := usecase.ListExercise(userContext(4), &requests.ListExerciseRequest{Page: 1, PageSize: 10})
	if customErr == nil || customErr.ErrType != apperror.ErrInternalServer {
		t.Fatalf("got %#v, want internal server", customErr)
	}
}

func TestMapExerciseQuestionToJSON(t *testing.T) {
	exercise := json.RawMessage(`{
		"content": {
			"grammar": {"questions": [{"question": "q1", "answers": ["a", "b"], "correct_answer": 1, "explanation": {"en": "because", "id": "karena"}}]},
			"language": {"questions": [{"question": "q2", "answers": ["c", "d"], "correct_answer": 2, "explanation": "plain"}]},
			"listening": {"questions": [{"question": "q3", "answers": ["e", "f"], "correct_answer": 1}]}
		}
	}`)
	answer := json.RawMessage(`{"grammar":{"1":"a"},"language":{"1":"c"},"listening":{}}`)

	got := mapExerciseQuestionToJSON(exercise, answer)
	grammar := got["grammar"]
	language := got["language"]
	listening := got["listening"]
	if len(grammar) != 1 || grammar[0].Question != "q1" || grammar[0].CorrectAnswer != 1 || grammar[0].UserAnswer != "a" || !grammar[0].IsCorrect || grammar[0].Explanation != "because\nkarena" {
		t.Fatalf("grammar = %#v", grammar)
	}
	if len(language) != 1 || language[0].UserAnswer != "c" || language[0].IsCorrect || language[0].Explanation != "plain" {
		t.Fatalf("language = %#v", language)
	}
	if len(listening) != 1 || listening[0].UserAnswer != "" || listening[0].IsCorrect || len(listening[0].Answers) != 2 {
		t.Fatalf("listening = %#v", listening)
	}

	if len(mapExerciseQuestionToJSON(nil, nil)) != 0 {
		t.Fatal("empty exercise was mapped")
	}
	if len(mapExerciseQuestionToJSON(json.RawMessage(`{"title":"N4"}`), nil)) != 0 {
		t.Fatal("exercise without questions was mapped")
	}
}

func userContext(userID uint64) context.Context {
	return context.WithValue(context.Background(), enums.UserID, userID)
}

func newGenerateExerciseUsecase(t *testing.T, repo *repomocks.MockExerciseRepository, cache *cachemocks.MockExerciseCache, workflow *workflowmocks.MockExerciseWorkflow) *exerciseUsecase {
	t.Helper()
	cache.EXPECT().GetExerciseByIdempotencyKey(mock.Anything, mock.Anything).Return(nil, redis.Nil).Maybe()
	subscriptions := repomocks.NewMockSubscriptionRepository(t)
	subscriptions.EXPECT().FindCurrentByUserID(mock.Anything, mock.Anything).Return(&models.Subscription{ID: 1}, nil).Maybe()
	return &exerciseUsecase{
		exerciseRepository:     repo,
		exerciseCache:          cache,
		exerciseWorkflow:       workflow,
		subscriptionRepository: subscriptions,
	}
}
