package exercise_workflow

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"

	temporal_client "github.com/RandySteven/go-cook/temporal"
	"github.com/RandySteven/jlpt-exercise-be/apperror"
)

func (e *exerciseWorkflow) readPromptTxt(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	promptTxt, err := os.ReadFile(fileContentPath)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to read prompt txt", err)
	}

	if execData == nil || execData.UserRequest == nil {
		return nil, temporal_client.NewNonRetryableError(exerciseInputErrorType, "generate exercise request is missing")
	}

	userRequest := execData.UserRequest
	numbOfListeningQuestions, numbOfLanguageQuestions, numbOfGrammarQuestions := jlptLevelNumberQuestionMapping(userRequest.JLPTNLevel)

	promptTxtString := string(promptTxt)
	promptTxtString = strings.ReplaceAll(promptTxtString, "{{jlpt_level}}", strconv.Itoa(userRequest.JLPTNLevel))
	promptTxtString = strings.ReplaceAll(promptTxtString, "{{numb_of_listening_exams}}", strconv.Itoa(numbOfListeningQuestions))
	promptTxtString = strings.ReplaceAll(promptTxtString, "{{numb_of_language_exams}}", strconv.Itoa(numbOfLanguageQuestions))
	promptTxtString = strings.ReplaceAll(promptTxtString, "{{numb_of_grammar_exams}}", strconv.Itoa(numbOfGrammarQuestions))

	execData.Prompt = promptTxtString
	execData.SetActivity(generateExerciseActivity)
	if cacheErr := e.cacheProcessingExercise(ctx, execData); cacheErr != nil {
		log.Println("failed to cache processing exercise", cacheErr)
	}
	if err = e.signalExerciseEvent(ctx, execData, sgProcessing); err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to signal processing", err)
	}

	return execData, nil
}

func jlptLevelNumberQuestionMapping(jlptLevel int) (numbOfListeningQuestions int, numbOfReadingQuestions int, numbOfGrammarQuestions int) {
	switch jlptLevel {
	case 1:
		return 10, 10, 10
	case 2:
		return 20, 20, 20
	case 3:
		return 30, 30, 30
	case 4:
		return 40, 40, 40
	case 5:
		return 50, 50, 50
	}
	return 0, 0, 0
}
