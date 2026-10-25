package responses

import (
	"encoding/json"
	"strings"
)

type storedExercise struct {
	Content struct {
		Grammar   storedQuestionGroup `json:"grammar"`
		Language  storedQuestionGroup `json:"language"`
		Listening storedQuestionGroup `json:"listening"`
	} `json:"content"`
}

type storedQuestionGroup struct {
	Questions []storedQuestion `json:"questions"`
}

type storedQuestion struct {
	Question      string          `json:"question"`
	Answers       []string        `json:"answers"`
	CorrectAnswer int             `json:"correct_answer"`
	Explanation   json.RawMessage `json:"explanation"`
}

func ExerciseQuestionsFromJSON(exerciseJSON json.RawMessage, answerJSON json.RawMessage) map[string][]*ExerciseQuestion {
	questions := make(map[string][]*ExerciseQuestion)
	if len(strings.TrimSpace(string(exerciseJSON))) == 0 || string(exerciseJSON) == "null" {
		return questions
	}

	var stored storedExercise
	if err := json.Unmarshal(exerciseJSON, &stored); err != nil {
		return questions
	}

	answers := map[string]map[int]string{}
	if len(strings.TrimSpace(string(answerJSON))) > 0 && string(answerJSON) != "null" {
		_ = json.Unmarshal(answerJSON, &answers)
	}

	sections := []struct {
		name  string
		group storedQuestionGroup
	}{
		{name: "grammar", group: stored.Content.Grammar},
		{name: "language", group: stored.Content.Language},
		{name: "listening", group: stored.Content.Listening},
	}
	for _, section := range sections {
		if len(section.group.Questions) == 0 {
			continue
		}
		picked := answers[section.name]
		mapped := make([]*ExerciseQuestion, 0, len(section.group.Questions))
		for index, question := range section.group.Questions {
			userAnswer := strings.TrimSpace(picked[index+1])
			correctAnswer := uint64(0)
			if question.CorrectAnswer > 0 {
				correctAnswer = uint64(question.CorrectAnswer)
			}
			choices := question.Answers
			if choices == nil {
				choices = []string{}
			}
			mapped = append(mapped, &ExerciseQuestion{
				Question:      question.Question,
				Answers:       choices,
				CorrectAnswer: correctAnswer,
				Explanation:   explanationText(question.Explanation),
				UserAnswer:    userAnswer,
				IsCorrect:     selectedIsCorrect(userAnswer, question.Answers, question.CorrectAnswer),
			})
		}
		questions[section.name] = mapped
	}
	return questions
}

func explanationText(raw json.RawMessage) string {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return ""
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return strings.TrimSpace(text)
	}
	var localized struct {
		EN string `json:"en"`
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &localized); err != nil {
		return ""
	}
	en := strings.TrimSpace(localized.EN)
	id := strings.TrimSpace(localized.ID)
	switch {
	case en != "" && id != "":
		return en + "\n" + id
	case en != "":
		return en
	default:
		return id
	}
}

func selectedIsCorrect(selected string, answers []string, correctAnswer int) bool {
	selected = strings.TrimSpace(selected)
	if selected == "" || correctAnswer < 1 || correctAnswer > len(answers) {
		return false
	}
	return selected == strings.TrimSpace(answers[correctAnswer-1])
}
