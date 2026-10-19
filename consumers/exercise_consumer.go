package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/RandySteven/jlpt-exercise-be/caches"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/entities/schemas"
	"github.com/RandySteven/jlpt-exercise-be/enums"
	deepseek_client "github.com/RandySteven/jlpt-exercise-be/pkg/deepseek"
	"github.com/RandySteven/jlpt-exercise-be/repositories"
)

type (
	ExerciseConsumer interface {
		Consumer
	}

	exerciseConsumer struct {
		exerciseRepository repositories.ExerciseRepository
		exerciseCache      caches.ExerciseCache
		deepSeek           deepseek_client.DeepSeek
	}

	generatedExercise struct {
		Content struct {
			Grammar   questionGroup `json:"grammar"`
			Language  questionGroup `json:"language"`
			Listening questionGroup `json:"listening"`
		} `json:"content"`
	}

	questionGroup struct {
		Questions []generatedQuestion `json:"questions"`
	}

	generatedQuestion struct {
		Answers       []string `json:"answers"`
		CorrectAnswer int      `json:"correct_answer"`
	}

	reviewNote struct {
		Note  string           `json:"note"`
		Items []reviewNoteItem `json:"items"`
	}

	reviewNoteItem struct {
		Section        string `json:"section"`
		QuestionNumber int    `json:"question_number"`
		Correct        bool   `json:"correct"`
		FeedbackEN     string `json:"feedback_en"`
		FeedbackID     string `json:"feedback_id"`
	}
)

var reviewNotePromptPath = "files/txt/review_note.txt"

func NewExerciseConsumer(exerciseRepository repositories.ExerciseRepository, exerciseCache caches.ExerciseCache, deepSeek deepseek_client.DeepSeek) ExerciseConsumer {
	return &exerciseConsumer{
		exerciseRepository: exerciseRepository,
		exerciseCache:      exerciseCache,
		deepSeek:           deepSeek,
	}
}

func (e *exerciseConsumer) Consumer(ctx context.Context, body string) (err error) {
	var schema schemas.ExerciseSchema
	err = json.Unmarshal([]byte(body), &schema)
	if err != nil {
		log.Println("failed to unmarshal exercise schema", err)
		return err
	}

	if e.exerciseRepository == nil || e.exerciseCache == nil || e.deepSeek == nil {
		err = errors.New("exercise consumer is not configured")
		log.Println(err)
		return err
	}

	exercise, err := e.exerciseRepository.FindByID(ctx, schema.ID)
	if err != nil {
		log.Println("failed to find exercise", err)
		return err
	}
	if exercise == nil {
		err = fmt.Errorf("exercise %d not found", schema.ID)
		log.Println(err)
		return err
	}
	if exercise.Status == enums.CompletedStatus && strings.TrimSpace(exercise.Note) != "" {
		return e.flushSubmitExercise(ctx, exercise.ID)
	}

	if exercise.Status == enums.SubmittedStatus {
		exercise.Status = enums.InReviewStatus
		_, err = e.exerciseRepository.Update(ctx, exercise)
		if err != nil {
			log.Println("failed to update exercise", err)
			return err
		}

		if cacheErr := e.exerciseCache.SetSubmitExercise(ctx, exercise.ID, submitExerciseResponseFromExercise(exercise)); cacheErr != nil {
			log.Println("failed to cache review exercise", cacheErr)
		}
	}

	score, total, err := scoreCalculation(exercise)
	if err != nil {
		log.Println("failed to calculate score", err)
		return err
	}

	prompt, err := buildReviewNotePrompt(exercise, score, total)
	if err != nil {
		log.Println("failed to build review note prompt", err)
		return err
	}

	rawNote, err := e.deepSeek.ChatGenerate(ctx, prompt)
	if err != nil {
		log.Println("failed to generate note", err)
		return err
	}
	note, err := reviewNoteFromModel(rawNote)
	if err != nil {
		log.Println("failed to read review note", err)
		return err
	}

	exercise.Note = note
	exercise.Score = score
	exercise.Status = enums.CompletedStatus
	_, err = e.exerciseRepository.Update(ctx, exercise)
	if err != nil {
		log.Println("failed to update exercise", err)
		return err
	}

	return e.flushSubmitExercise(ctx, exercise.ID)
}

func (e *exerciseConsumer) flushSubmitExercise(ctx context.Context, id uint64) error {
	if err := e.exerciseCache.DeleteSubmitExercise(ctx, id); err != nil {
		log.Println("failed to flush submitted exercise cache", err)
		return err
	}
	return nil
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

func buildReviewNotePrompt(exercise *models.Exercise, score int, total int) (string, error) {
	template, err := os.ReadFile(reviewNotePromptPath)
	if err != nil {
		return "", err
	}

	exerciseJSON, err := prettyJSON(exercise.Exercise)
	if err != nil {
		return "", err
	}
	answerJSON, err := prettyJSON(exercise.Answer)
	if err != nil {
		return "", err
	}

	prompt := string(template)
	prompt = strings.ReplaceAll(prompt, "{{jlpt_level}}", fmt.Sprintf("%d", exercise.NLevel))
	prompt = strings.ReplaceAll(prompt, "{{score}}", fmt.Sprintf("%d", score))
	prompt = strings.ReplaceAll(prompt, "{{total}}", fmt.Sprintf("%d", total))
	prompt = strings.ReplaceAll(prompt, "{{exercise}}", exerciseJSON)
	prompt = strings.ReplaceAll(prompt, "{{answer}}", answerJSON)
	return prompt, nil
}

func scoreCalculation(exercise *models.Exercise) (score int, total int, err error) {
	if exercise == nil || len(exercise.Exercise) == 0 {
		return 0, 0, errors.New("exercise content is empty")
	}

	var doc generatedExercise
	err = json.Unmarshal(exercise.Exercise, &doc)
	if err != nil {
		return 0, 0, err
	}

	answers := map[string]map[int]string{}
	if len(exercise.Answer) > 0 && string(exercise.Answer) != "null" {
		err = json.Unmarshal(exercise.Answer, &answers)
		if err != nil {
			return 0, 0, err
		}
	}

	sections := []struct {
		name  string
		group questionGroup
	}{
		{name: "grammar", group: doc.Content.Grammar},
		{name: "language", group: doc.Content.Language},
		{name: "listening", group: doc.Content.Listening},
	}
	for _, section := range sections {
		picked := answers[section.name]
		for index, question := range section.group.Questions {
			total++
			if answerIsCorrect(picked[index+1], question) {
				score++
			}
		}
	}
	if total == 0 {
		return 0, 0, errors.New("exercise has no questions")
	}
	return score, total, nil
}

func answerIsCorrect(selected string, question generatedQuestion) bool {
	selected = strings.TrimSpace(selected)
	if selected == "" || question.CorrectAnswer < 1 || question.CorrectAnswer > len(question.Answers) {
		return false
	}
	return selected == strings.TrimSpace(question.Answers[question.CorrectAnswer-1])
}

func reviewNoteFromModel(raw string) (string, error) {
	cleaned := strings.TrimSpace(raw)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	var parsed reviewNote
	err := json.Unmarshal([]byte(cleaned), &parsed)
	if err != nil {
		start := strings.Index(cleaned, "{")
		end := strings.LastIndex(cleaned, "}")
		if start < 0 || end <= start {
			return "", err
		}
		err = json.Unmarshal([]byte(cleaned[start:end+1]), &parsed)
		if err != nil {
			return "", err
		}
	}
	if strings.TrimSpace(parsed.Note) == "" {
		return "", errors.New("review note is empty")
	}
	if parsed.Items == nil {
		parsed.Items = []reviewNoteItem{}
	}
	encoded, err := json.Marshal(parsed)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func prettyJSON(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "{}", nil
	}
	var value any
	err := json.Unmarshal(raw, &value)
	if err != nil {
		return "", err
	}
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

var _ ExerciseConsumer = &exerciseConsumer{}
