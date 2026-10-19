package usecases

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	"github.com/RandySteven/jlpt-exercise-be/logic/exercise_workflow"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
)

func TestGenerateExerciseBlocksOpenExercise(t *testing.T) {
	for _, status := range []string{enums.InGeneratedStatus, enums.GeneratedStatus} {
		t.Run(status, func(t *testing.T) {
			repo := &fakeExerciseRepository{
				latest: &models.Exercise{ID: 9, ReferenceID: "open-key", Status: status, UserID: 4, NLevel: 3},
			}
			workflow := &fakeExerciseWorkflow{}
			usecase := newGenerateExerciseUsecase(repo, &fakeExerciseCache{}, workflow)

			_, customErr := usecase.GenerateExercise(userContext(4), &requests.GenerateExerciseRequest{
				IdempotencyKey: "new-key",
				JLPTNLevel:     3,
			})
			if customErr == nil || customErr.ErrType != apperror.ErrConflict {
				t.Fatalf("status %s: got %#v, want conflict", status, customErr)
			}
			if workflow.called || repo.saved != nil {
				t.Fatal("open exercise still started a new generation")
			}
		})
	}
}

func TestGenerateExerciseAllowsFinishedExercise(t *testing.T) {
	for _, status := range []string{enums.SubmittedStatus, enums.CompletedStatus, enums.FailedStatus} {
		t.Run(status, func(t *testing.T) {
			repo := &fakeExerciseRepository{
				latest: &models.Exercise{ID: 9, ReferenceID: "done-key", Status: status, UserID: 4},
			}
			workflow := &fakeExerciseWorkflow{
				response: &responses.GenerateExerciseResponse{Status: enums.InGeneratedStatus},
			}
			usecase := newGenerateExerciseUsecase(repo, &fakeExerciseCache{}, workflow)

			response, customErr := usecase.GenerateExercise(userContext(4), &requests.GenerateExerciseRequest{
				IdempotencyKey: "new-key",
				JLPTNLevel:     2,
			})
			if customErr != nil {
				t.Fatal(customErr)
			}
			if !workflow.called || repo.saved == nil || repo.saved.Status != enums.InGeneratedStatus {
				t.Fatal("finished exercise did not claim a new generation")
			}
			if response.ID != repo.saved.ID {
				t.Fatalf("response id = %d, want claimed id %d", response.ID, repo.saved.ID)
			}
		})
	}
}

func TestGenerateExerciseReturnsSameOpenExercise(t *testing.T) {
	repo := &fakeExerciseRepository{
		latest: &models.Exercise{ID: 9, ReferenceID: "same-key", Status: enums.GeneratedStatus, UserID: 4, NLevel: 5},
	}
	workflow := &fakeExerciseWorkflow{}
	usecase := newGenerateExerciseUsecase(repo, &fakeExerciseCache{}, workflow)

	response, customErr := usecase.GenerateExercise(userContext(4), &requests.GenerateExerciseRequest{
		IdempotencyKey: "same-key",
		JLPTNLevel:     5,
	})
	if customErr != nil {
		t.Fatal(customErr)
	}
	if response.ID != 9 || workflow.called || repo.saved != nil {
		t.Fatalf("same key restarted generation: response=%#v called=%v saved=%v", response, workflow.called, repo.saved)
	}
}

func TestGenerateExerciseReleasesClaimWhenWorkflowDoesNotStart(t *testing.T) {
	repo := &fakeExerciseRepository{latestErr: sql.ErrNoRows}
	workflow := &fakeExerciseWorkflow{
		err: apperror.NewCustomError(apperror.ErrInternalServer, "start failed", exercise_workflow.ErrWorkflowNotStarted),
	}
	usecase := newGenerateExerciseUsecase(repo, &fakeExerciseCache{}, workflow)

	_, customErr := usecase.GenerateExercise(userContext(4), &requests.GenerateExerciseRequest{IdempotencyKey: "key", JLPTNLevel: 1})
	if customErr == nil {
		t.Fatal("expected workflow error")
	}
	if repo.releasedFrom != enums.InGeneratedStatus || repo.releasedTo != enums.FailedStatus {
		t.Fatalf("claim was not released: from=%s to=%s", repo.releasedFrom, repo.releasedTo)
	}
}

func TestGenerateExerciseConflictOnConcurrentClaim(t *testing.T) {
	repo := &fakeExerciseRepository{
		latestErr: sql.ErrNoRows,
		saveErr:   &pgconn.PgError{Code: "23505"},
	}
	workflow := &fakeExerciseWorkflow{}
	usecase := newGenerateExerciseUsecase(repo, &fakeExerciseCache{}, workflow)

	_, customErr := usecase.GenerateExercise(userContext(4), &requests.GenerateExerciseRequest{IdempotencyKey: "key", JLPTNLevel: 1})
	if customErr == nil || customErr.ErrType != apperror.ErrConflict {
		t.Fatalf("got %#v, want conflict", customErr)
	}
	if workflow.called {
		t.Fatal("concurrent claim started a workflow")
	}
}

func TestListExerciseRequiresRequest(t *testing.T) {
	usecase := newGenerateExerciseUsecase(&fakeExerciseRepository{}, &fakeExerciseCache{}, &fakeExerciseWorkflow{})

	_, customErr := usecase.ListExercise(userContext(4), nil)
	if customErr == nil || customErr.ErrType != apperror.ErrBadRequest {
		t.Fatalf("got %#v, want bad request", customErr)
	}
}

func TestListExerciseRequiresUser(t *testing.T) {
	usecase := newGenerateExerciseUsecase(&fakeExerciseRepository{}, &fakeExerciseCache{}, &fakeExerciseWorkflow{})
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
	repo := &fakeExerciseRepository{
		listed: []*models.Exercise{
			{
				ID:        11,
				NLevel:    3,
				Status:    enums.GeneratedStatus,
				Exercise:  json.RawMessage(`{"q":1}`),
				CreatedAt: created,
				UpdatedAt: updated,
			},
		},
		total: 4,
	}
	usecase := newGenerateExerciseUsecase(repo, &fakeExerciseCache{}, &fakeExerciseWorkflow{})

	response, customErr := usecase.ListExercise(userContext(4), &requests.ListExerciseRequest{Page: 2, PageSize: 10})
	if customErr != nil {
		t.Fatal(customErr)
	}
	if repo.listUserID != 4 || repo.listPage != 2 || repo.listPageSize != 10 || repo.countUserID != 4 {
		t.Fatalf("repository called with user=%d page=%d pageSize=%d countUser=%d", repo.listUserID, repo.listPage, repo.listPageSize, repo.countUserID)
	}
	if response.TotalItems != 4 || response.Page != 2 || response.PageSize != 10 || len(response.Exercises) != 1 {
		t.Fatalf("page response = %#v", response)
	}
	got := response.Exercises[0]
	if got.ID != 11 || got.Status != enums.GeneratedStatus || got.JLPTLevel != "3" || string(got.Exercise) != `{"q":1}` || !got.CreatedAt.Equal(created) || !got.UpdatedAt.Equal(updated) {
		t.Fatalf("mapped exercise = %#v", got)
	}
}

func TestListExerciseEmptyPage(t *testing.T) {
	repo := &fakeExerciseRepository{listed: []*models.Exercise{}, total: 0}
	usecase := newGenerateExerciseUsecase(repo, &fakeExerciseCache{}, &fakeExerciseWorkflow{})

	response, customErr := usecase.ListExercise(userContext(4), &requests.ListExerciseRequest{Page: 1, PageSize: 20})
	if customErr != nil {
		t.Fatal(customErr)
	}
	if response.TotalItems != 0 || len(response.Exercises) != 0 || response.Page != 1 || response.PageSize != 20 {
		t.Fatalf("empty page = %#v", response)
	}
}

func TestListExerciseFindError(t *testing.T) {
	repo := &fakeExerciseRepository{listErr: errors.New("db down")}
	usecase := newGenerateExerciseUsecase(repo, &fakeExerciseCache{}, &fakeExerciseWorkflow{})

	_, customErr := usecase.ListExercise(userContext(4), &requests.ListExerciseRequest{Page: 1, PageSize: 10})
	if customErr == nil || customErr.ErrType != apperror.ErrInternalServer {
		t.Fatalf("got %#v, want internal server", customErr)
	}
	if repo.countCalled {
		t.Fatal("count ran after find failed")
	}
}

func TestListExerciseCountError(t *testing.T) {
	repo := &fakeExerciseRepository{countErr: errors.New("count failed")}
	usecase := newGenerateExerciseUsecase(repo, &fakeExerciseCache{}, &fakeExerciseWorkflow{})

	_, customErr := usecase.ListExercise(userContext(4), &requests.ListExerciseRequest{Page: 1, PageSize: 10})
	if customErr == nil || customErr.ErrType != apperror.ErrInternalServer {
		t.Fatalf("got %#v, want internal server", customErr)
	}
	if !repo.listCalled {
		t.Fatal("find did not run before count failed")
	}
}

func userContext(userID uint64) context.Context {
	return context.WithValue(context.Background(), enums.UserID, userID)
}

func newGenerateExerciseUsecase(repo *fakeExerciseRepository, cache *fakeExerciseCache, workflow *fakeExerciseWorkflow) *exerciseUsecase {
	return &exerciseUsecase{
		exerciseRepository: repo,
		exerciseCache:      cache,
		exerciseWorkflow:   workflow,
	}
}

type fakeExerciseWorkflow struct {
	called   bool
	response *responses.GenerateExerciseResponse
	err      *apperror.CustomError
}

func (f *fakeExerciseWorkflow) GenerateExercise(ctx context.Context, request *requests.GenerateExerciseRequest) (*responses.GenerateExerciseResponse, *apperror.CustomError) {
	f.called = true
	return f.response, f.err
}

type fakeExerciseCache struct {
	cached *responses.GenerateExerciseResponse
}

func (f *fakeExerciseCache) SetExercise(ctx context.Context, id uint64, response *responses.GenerateExerciseResponse) error {
	return nil
}
func (f *fakeExerciseCache) GetExercise(ctx context.Context, id uint64) (*responses.GenerateExerciseResponse, error) {
	return nil, redis.Nil
}
func (f *fakeExerciseCache) SetExerciseByIdempotencyKey(ctx context.Context, key string, response *responses.GenerateExerciseResponse) error {
	return nil
}
func (f *fakeExerciseCache) DeleteExercise(ctx context.Context, id uint64) error {
	return nil
}
func (f *fakeExerciseCache) DeleteExerciseByIdempotencyKey(ctx context.Context, key string) error {
	return nil
}
func (f *fakeExerciseCache) GetExerciseByIdempotencyKey(ctx context.Context, key string) (*responses.GenerateExerciseResponse, error) {
	if f.cached == nil {
		return nil, redis.Nil
	}
	return f.cached, nil
}
func (f *fakeExerciseCache) SetSubmitExercise(ctx context.Context, id uint64, response *responses.SubmitExerciseResponse) error {
	return nil
}
func (f *fakeExerciseCache) GetSubmitExercise(ctx context.Context, id uint64) (*responses.SubmitExerciseResponse, error) {
	return nil, redis.Nil
}

type fakeExerciseRepository struct {
	latest       *models.Exercise
	latestErr    error
	saved        *models.Exercise
	saveErr      error
	releasedFrom string
	releasedTo   string
	nextID       uint64
	listed       []*models.Exercise
	listErr      error
	listCalled   bool
	listUserID   uint64
	listPage     int64
	listPageSize int64
	total        int64
	countErr     error
	countCalled  bool
	countUserID  uint64
}

func (f *fakeExerciseRepository) Save(ctx context.Context, entity *models.Exercise) (*models.Exercise, error) {
	if f.saveErr != nil {
		return nil, f.saveErr
	}
	copied := *entity
	if f.nextID == 0 {
		f.nextID = 21
	}
	copied.ID = f.nextID
	f.saved = &copied
	return f.saved, nil
}
func (f *fakeExerciseRepository) FindByID(ctx context.Context, id uint64) (*models.Exercise, error) {
	return nil, sql.ErrNoRows
}
func (f *fakeExerciseRepository) FindAll(ctx context.Context, skip uint64, take uint64) ([]*models.Exercise, error) {
	return nil, nil
}
func (f *fakeExerciseRepository) Update(ctx context.Context, entity *models.Exercise) (*models.Exercise, error) {
	return entity, nil
}
func (f *fakeExerciseRepository) DeleteByID(ctx context.Context, id uint64) error {
	return nil
}
func (f *fakeExerciseRepository) FindLatestByUserID(ctx context.Context, userID uint64) (*models.Exercise, error) {
	if f.latestErr != nil {
		return nil, f.latestErr
	}
	if f.latest == nil {
		return nil, sql.ErrNoRows
	}
	return f.latest, nil
}
func (f *fakeExerciseRepository) UpdateStatusIfCurrent(ctx context.Context, id uint64, currentStatus, nextStatus, note string) (bool, error) {
	f.releasedFrom = currentStatus
	f.releasedTo = nextStatus
	return true, nil
}
func (f *fakeExerciseRepository) FindAllByUserID(ctx context.Context, userID uint64, page int64, pageSize int64) ([]*models.Exercise, error) {
	f.listCalled = true
	f.listUserID = userID
	f.listPage = page
	f.listPageSize = pageSize
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listed, nil
}
func (f *fakeExerciseRepository) CountByUserID(ctx context.Context, userID uint64) (int64, error) {
	f.countCalled = true
	f.countUserID = userID
	if f.countErr != nil {
		return 0, f.countErr
	}
	return f.total, nil
}
