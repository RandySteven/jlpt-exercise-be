package usecases

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/caches"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/requests"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/entities/schemas"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	"github.com/RandySteven/jlpt-exercise-be/logic/exercise_workflow"
	deepseek_client "github.com/RandySteven/jlpt-exercise-be/pkg/deepseek"
	pdf_client "github.com/RandySteven/jlpt-exercise-be/pkg/pdf"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
	"github.com/RandySteven/jlpt-exercise-be/topics"
	"github.com/RandySteven/jlpt-exercise-be/utils"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
)

type (
	ExerciseUsecase interface {
		GenerateExercise(ctx context.Context, request *requests.GenerateExerciseRequest) (response *responses.GenerateExerciseResponse, customErr *apperror.CustomError)
		GetExercise(ctx context.Context, id uint64) (response *responses.GenerateExerciseResponse, customErr *apperror.CustomError)
		SubmitExercise(ctx context.Context, request *requests.SubmitExerciseRequest) (response *responses.SubmitExerciseResponse, customErr *apperror.CustomError)
		ListExercise(ctx context.Context, request *requests.ListExerciseRequest) (response *responses.ListExerciseResponse, customErr *apperror.CustomError)
		DownloadExercise(ctx context.Context, request *requests.DownloadExerciseRequest) (response *responses.ExerciseDownloadResponse, customErr *apperror.CustomError)
	}
)

type (
	exerciseUsecase struct {
		deepSeek                   deepseek_client.DeepSeek
		pdfClient                  pdf_client.PDFClient
		exerciseWorkflow           exercise_workflow.ExerciseWorkflow
		exerciseRepository         repositories.ExerciseRepository
		subscriptionRepository     repositories.SubscriptionRepository
		subscriptionPlanRepository repositories.SubscriptionPlanRepository
		userRepository             repositories.UserRepository
		exerciseTopic              topics.ExerciseTopic
		exerciseCache              caches.ExerciseCache
	}
)

func newExerciseUsecase(
	deepSeek deepseek_client.DeepSeek,
	temporal temporal_client.Temporal,
	exerciseRepository repositories.ExerciseRepository,
	userRepository repositories.UserRepository,
	subscriptionRepository repositories.SubscriptionRepository,
	subscriptionPlanRepository repositories.SubscriptionPlanRepository,
	exerciseCache caches.ExerciseCache,
	exerciseTopic topics.ExerciseTopic,
	pdfClient pdf_client.PDFClient,
) ExerciseUsecase {
	exerciseWorkflow := exercise_workflow.NewExerciseWorkflow(
		temporal_client.NewWorkflowExecution(temporal),
		temporal,
		exerciseRepository,
		userRepository,
		exerciseCache,
		deepSeek,
	)
	return &exerciseUsecase{
		deepSeek:                   deepSeek,
		pdfClient:                  pdfClient,
		exerciseWorkflow:           exerciseWorkflow,
		exerciseRepository:         exerciseRepository,
		userRepository:             userRepository,
		subscriptionRepository:     subscriptionRepository,
		subscriptionPlanRepository: subscriptionPlanRepository,
		exerciseCache:              exerciseCache,
		exerciseTopic:              exerciseTopic,
	}
}

// ListExercise implements [ExerciseUsecase].
func (e *exerciseUsecase) ListExercise(ctx context.Context, request *requests.ListExerciseRequest) (response *responses.ListExerciseResponse, customErr *apperror.CustomError) {
	if request == nil {
		return nil, apperror.NewCustomError(apperror.ErrBadRequest, "list exercise request is nil", fmt.Errorf("list exercise request is nil"))
	}

	userID, ok := ctx.Value(enums.UserID).(uint64)
	if !ok || userID == 0 {
		return nil, apperror.NewCustomError(apperror.ErrUnauthorized, "user is missing", fmt.Errorf("user id is missing"))
	}

	exercises, err := e.exerciseRepository.FindAllByUserID(ctx, userID, request.Page-1, request.PageSize)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find exercises", err)
	}

	totalItems, err := e.exerciseRepository.CountByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to count exercises", err)
	}

	return &responses.ListExerciseResponse{
		Exercises:  exerciseResponsesFromModels(exercises),
		TotalItems: totalItems,
		Page:       request.Page,
		PageSize:   request.PageSize,
	}, nil
}

func (e *exerciseUsecase) GenerateExercise(ctx context.Context, request *requests.GenerateExerciseRequest) (response *responses.GenerateExerciseResponse, customErr *apperror.CustomError) {
	if request == nil {
		return nil, apperror.NewCustomError(apperror.ErrBadRequest, "generate exercise request is nil", fmt.Errorf("generate exercise request is nil"))
	}

	var isSubscribed bool = false
	if isSubscribed, customErr = e.checkSubscription(ctx); customErr != nil {
		return nil, customErr
	}
	if !isSubscribed {
		return nil, apperror.NewCustomError(apperror.ErrForbidden, "user is not subscribed", fmt.Errorf("user is not subscribed"))
	}

	cached, err := e.exerciseCache.GetExerciseByIdempotencyKey(ctx, request.IdempotencyKey)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to read generated exercise", err)
	}
	if cached != nil {
		return cached, nil
	}

	userID, ok := ctx.Value(enums.UserID).(uint64)
	if !ok || userID == 0 {
		return nil, apperror.NewCustomError(apperror.ErrUnauthorized, "user is missing", fmt.Errorf("user id is missing"))
	}

	latest, err := e.exerciseRepository.FindLatestByUserID(ctx, userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find latest exercise", err)
	}
	if latest != nil && enums.ExerciseInProgress(latest.Status) {
		if latest.ReferenceID == request.IdempotencyKey {
			return exerciseResponseFromModel(latest), nil
		}
		return nil, apperror.NewCustomError(apperror.ErrConflict, "exercise is still in progress", fmt.Errorf("user %d exercise %d status %s", userID, latest.ID, latest.Status))
	}

	claimed, err := e.exerciseRepository.Save(ctx, &models.Exercise{
		NLevel:      request.JLPTNLevel,
		UserID:      userID,
		ReferenceID: request.IdempotencyKey,
		Status:      enums.InGeneratedStatus,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, apperror.NewCustomError(apperror.ErrConflict, "exercise is still in progress", err)
		}
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to claim exercise", err)
	}

	response, customErr = e.exerciseWorkflow.GenerateExercise(ctx, request)
	if customErr != nil {
		if errors.Is(customErr, exercise_workflow.ErrWorkflowNotStarted) {
			_, _ = e.exerciseRepository.UpdateStatusIfCurrent(ctx, claimed.ID, enums.InGeneratedStatus, enums.FailedStatus, customErr.Error())
		}
		return nil, customErr
	}
	if response != nil && response.ID == 0 {
		response.ID = claimed.ID
	}
	return response, nil
}

func exerciseResponseFromModel(exercise *models.Exercise) *responses.GenerateExerciseResponse {
	exerciseQuestions := mapExerciseQuestionToJSON(exercise.Exercise, exercise.Answer)

	return &responses.GenerateExerciseResponse{
		ID:        exercise.ID,
		Status:    exercise.Status,
		JLPTLevel: strconv.Itoa(exercise.NLevel),
		Exercise:  exerciseQuestions,
		CreatedAt: exercise.CreatedAt,
		UpdatedAt: exercise.UpdatedAt,
	}
}

func mapExerciseQuestionToJSON(exerciseQuestion json.RawMessage, exerciseAnswer json.RawMessage) map[string][]*responses.ExerciseQuestion {
	return responses.ExerciseQuestionsFromJSON(exerciseQuestion, exerciseAnswer)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (e *exerciseUsecase) GetExercise(ctx context.Context, id uint64) (response *responses.GenerateExerciseResponse, customErr *apperror.CustomError) {
	exercise, err := e.exerciseRepository.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find exercise", err)
	}
	if exercise == nil {
		return nil, apperror.NewCustomError(apperror.ErrNotFound, "exercise not found", fmt.Errorf("exercise %d not found", id))
	}
	return exerciseResponseFromModel(exercise), nil
}

func (e *exerciseUsecase) DownloadExercise(ctx context.Context, request *requests.DownloadExerciseRequest) (response *responses.ExerciseDownloadResponse, customErr *apperror.CustomError) {
	userID, ok := ctx.Value(enums.UserID).(uint64)
	if !ok || userID == 0 {
		return nil, apperror.NewCustomError(apperror.ErrUnauthorized, "user is missing", fmt.Errorf("user id is missing"))
	}

	if request == nil {
		return nil, apperror.NewCustomError(apperror.ErrBadRequest, "download exercise request is nil", fmt.Errorf("download exercise request is nil"))
	}

	exercise, err := e.exerciseRepository.FindByID(ctx, request.ExerciseID)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find exercise", err)
	}
	if exercise == nil {
		return nil, apperror.NewCustomError(apperror.ErrNotFound, "exercise not found", fmt.Errorf("exercise %d not found", request.ExerciseID))
	}

	if exercise.UserID != userID {
		return nil, apperror.NewCustomError(apperror.ErrForbidden, "exercise not found", fmt.Errorf("exercise %d not found", request.ExerciseID))
	}

	if exercise.Status != enums.CompletedStatus {
		return nil, apperror.NewCustomError(apperror.ErrForbidden, "exercise is not completed", fmt.Errorf("exercise %d is not completed", request.ExerciseID))
	}

	_, err = e.pdfClient.GeneratePDF("files/html/exam_exercise_result.template.html", map[string]interface{}{
		"ExerciseID":    exercise.ID,
		"ExerciseScore": exercise.Score,
	})
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to generate PDF", err)
	}

	response = &responses.ExerciseDownloadResponse{}
	return response, nil
}

func (e *exerciseUsecase) SubmitExercise(ctx context.Context, request *requests.SubmitExerciseRequest) (response *responses.SubmitExerciseResponse, customErr *apperror.CustomError) {
	if request == nil {
		return nil, apperror.NewCustomError(apperror.ErrBadRequest, "submit exercise request is nil", fmt.Errorf("submit exercise request is nil"))
	}

	cached, err := e.exerciseCache.GetSubmitExercise(ctx, request.ID)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to read submitted exercise", err)
	}
	if cached != nil {
		return cached, nil
	}

	exercise, err := e.exerciseRepository.FindByID(ctx, request.ID)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find exercise", err)
	}
	if exercise == nil {
		return nil, apperror.NewCustomError(apperror.ErrNotFound, "exercise not found", fmt.Errorf("exercise %d not found", request.ID))
	}

	if isCachedSubmitStatus(exercise.Status) {
		return e.cacheSubmitExercise(ctx, exercise)
	}

	if exercise.Status == enums.GeneratedStatus {
		exercise.Answer = answerMapToJSON(request.Answer)
		exercise.Status = enums.SubmittedStatus
		_, err = e.exerciseRepository.Update(ctx, exercise)
		if err != nil {
			return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to update exercise", err)
		}

		schema := mapExerciseSchemaFromModel(exercise)
		err = e.exerciseTopic.WriteMessage(ctx, schema)
		if err != nil {
			log.Println("failed to write exercise to topic", err)
		}
		return e.cacheSubmitExercise(ctx, exercise)
	}

	return submitExerciseResponseFromExercise(exercise), nil
}

func isCachedSubmitStatus(status string) bool {
	switch status {
	case enums.SubmittedStatus, enums.InReviewStatus, enums.CompletedStatus:
		return true
	default:
		return false
	}
}

func (e *exerciseUsecase) cacheSubmitExercise(ctx context.Context, exercise *models.Exercise) (*responses.SubmitExerciseResponse, *apperror.CustomError) {
	response := submitExerciseResponseFromExercise(exercise)
	if err := e.exerciseCache.SetSubmitExercise(ctx, exercise.ID, response); err != nil {
		log.Println("failed to cache submitted exercise", err)
	}
	return response, nil
}

func submitExerciseResponseFromExercise(exercise *models.Exercise) *responses.SubmitExerciseResponse {
	return &responses.SubmitExerciseResponse{
		ID:        exercise.ID,
		Status:    exercise.Status,
		Answer:    exercise.Answer,
		CreatedAt: exercise.CreatedAt,
		UpdatedAt: exercise.UpdatedAt,
	}
}

func exerciseResponsesFromModels(exercises []*models.Exercise) []*responses.ListExerciseDataResponse {
	responses := make([]*responses.ListExerciseDataResponse, len(exercises))
	for i, exercise := range exercises {
		responses[i] = mapListExerciseDataResponseFromModel(exercise)
	}
	return responses
}

func mapListExerciseDataResponseFromModel(exercise *models.Exercise) *responses.ListExerciseDataResponse {
	return &responses.ListExerciseDataResponse{
		ID:          exercise.ID,
		Status:      exercise.Status,
		JLPTLevel:   strconv.Itoa(exercise.NLevel),
		RedirectURL: fmt.Sprintf("%s/%d", enums.ExercisePrefix, exercise.ID),
		CreatedAt:   exercise.CreatedAt,
		UpdatedAt:   exercise.UpdatedAt,
	}
}

func mapExerciseSchemaFromModel(exercise *models.Exercise) *schemas.ExerciseSchema {
	return &schemas.ExerciseSchema{
		ID:        exercise.ID,
		UserID:    exercise.UserID,
		Exercise:  exercise.Exercise,
		Answer:    exercise.Answer,
		NLevel:    exercise.NLevel,
		Score:     exercise.Score,
		Status:    exercise.Status,
		Note:      exercise.Note,
		CreatedAt: exercise.CreatedAt,
		UpdatedAt: exercise.UpdatedAt,
	}
}

func answerMapToJSON(answer map[string]map[int]string) json.RawMessage {
	jsonAnswer, err := json.Marshal(answer)
	if err != nil {
		return nil
	}
	return jsonAnswer
}

func (e *exerciseUsecase) checkSubscription(ctx context.Context) (isSubscribed bool, customErr *apperror.CustomError) {
	userID, ok := ctx.Value(enums.UserID).(uint64)
	if !ok || userID == 0 {
		return false, apperror.NewCustomError(apperror.ErrUnauthorized, "user is missing", fmt.Errorf("user id is missing"))
	}

	subscription, err := e.subscriptionRepository.FindCurrentByUserID(ctx, userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find subscription", err)
	}

	if subscription != nil {
		isEnabled, customErr := e.checkFeatureFlag(ctx, subscription.PlanID, "exercise_generator")
		if customErr != nil {
			return false, customErr
		}
		return isEnabled, nil
	}

	return false, nil
}

func (e *exerciseUsecase) checkFeatureFlag(ctx context.Context, planID uint64, featureFlag string) (isEnabled bool, customErr *apperror.CustomError) {
	plan, err := e.subscriptionPlanRepository.FindByID(ctx, planID)
	if err != nil {
		return false, apperror.NewCustomError(apperror.ErrInternalServer, "failed to find subscription plan", err)
	}
	if plan == nil {
		return false, apperror.NewCustomError(apperror.ErrNotFound, "subscription plan not found", fmt.Errorf("subscription plan %d not found", planID))
	}
	featureFlags := utils.FeatureFlagsToMap(plan.FeatureFlags)
	return featureFlags[featureFlag], nil
}
