package exercise_workflow

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/RandySteven/jlpt-exercise-be/apperror"
	"github.com/RandySteven/jlpt-exercise-be/entities/payloads/responses"
	"github.com/RandySteven/jlpt-exercise-be/enums"
)

func (e *exerciseWorkflow) generate(ctx context.Context, execData *ExecutionData) (*ExecutionData, error) {
	result, err := e.deepSeekClient.ChatGenerate(ctx, execData.Prompt)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to generate text", err)
	}
	execData.PromptResult = result
	sanitized, err := sanitizeChatGenerateJSON(execData.PromptResult)
	if err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to generate text", err)
	}
	if execData.GenerateExercise == nil {
		execData.GenerateExercise = &responses.GenerateExerciseResponse{}
	}
	execData.Exercise.Status = enums.GeneratedStatus
	execData.Exercise.Exercise = json.RawMessage(sanitized)
	execData.Exercise.UpdatedAt = time.Now()
	execData.SetActivity(persistGeneratedResponseActivity)
	if err = e.signalExerciseEvent(ctx, execData, sgCompleted); err != nil {
		return nil, apperror.NewCustomError(apperror.ErrInternalServer, "failed to signal completed generation", err)
	}
	return execData, nil
}

func sanitizeChatGenerateJSON(raw string) (string, error) {
	cleaned := stripThinkBlocks(strings.TrimSpace(raw))
	best := ""
	for _, candidate := range jsonCandidates(cleaned) {
		repaired := repairTrailingCommas(strings.TrimSpace(candidate))
		if !jsonFormatChecker(repaired) {
			continue
		}
		if len(repaired) > len(best) {
			best = repaired
		}
	}
	if best == "" {
		return "", errors.New("invalid json format")
	}
	return best, nil
}

func jsonCandidates(value string) []string {
	var candidates []string
	if jsonFormatChecker(repairTrailingCommas(strings.TrimSpace(value))) {
		candidates = append(candidates, value)
	}
	rest := value
	for {
		start := strings.Index(rest, "```")
		if start < 0 {
			break
		}
		rest = rest[start+3:]
		if len(rest) >= 4 && strings.EqualFold(rest[:4], "json") {
			rest = strings.TrimSpace(rest[4:])
		}
		end := strings.Index(rest, "```")
		if end < 0 {
			candidates = append(candidates, rest)
			break
		}
		candidates = append(candidates, rest[:end])
		rest = rest[end+3:]
	}
	for i := 0; i < len(value); i++ {
		if value[i] != '{' {
			continue
		}
		end := matchingJSONBrace(value, i)
		if end < 0 {
			continue
		}
		candidates = append(candidates, value[i:end+1])
		i = end
	}
	return candidates
}

func matchingJSONBrace(value string, start int) int {
	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(value); i++ {
		char := value[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if char == '\\' {
				escaped = true
				continue
			}
			if char == '"' {
				inString = false
			}
			continue
		}
		switch char {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func repairTrailingCommas(value string) string {
	out := make([]byte, 0, len(value))
	inString := false
	escaped := false
	for i := 0; i < len(value); i++ {
		char := value[i]
		if inString {
			out = append(out, char)
			if escaped {
				escaped = false
				continue
			}
			if char == '\\' {
				escaped = true
				continue
			}
			if char == '"' {
				inString = false
			}
			continue
		}
		if char == '"' {
			inString = true
			out = append(out, char)
			continue
		}
		if char == ',' {
			next := i + 1
			for next < len(value) && (value[next] == ' ' || value[next] == '\n' || value[next] == '\r' || value[next] == '\t') {
				next++
			}
			if next < len(value) && (value[next] == '}' || value[next] == ']') {
				continue
			}
		}
		out = append(out, char)
	}
	return string(out)
}

func stripThinkBlocks(value string) string {
	for {
		start := strings.Index(value, "<think>")
		if start < 0 {
			return value
		}
		end := strings.Index(value[start:], "</think>")
		if end < 0 {
			return value[:start]
		}
		value = value[:start] + value[start+end+len("</think>"):]
	}
}

func jsonFormatChecker(str string) bool {
	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(str), &jsonData)
	if err != nil {
		return false
	}
	return true
}
