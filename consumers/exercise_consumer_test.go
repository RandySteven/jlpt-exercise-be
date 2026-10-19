package consumers

import (
	"encoding/json"
	"testing"

	"github.com/RandySteven/jlpt-exercise-be/entities/models"
)

func TestScoreCalculation(t *testing.T) {
	exercise := &models.Exercise{
		Exercise: json.RawMessage(`{
			"content": {
				"grammar": {
					"questions": [
						{"question": "q1", "answers": ["a", "b"], "correct_answer": 1}
					]
				},
				"language": {
					"questions": [
						{"question": "q2", "answers": ["c", "d"], "correct_answer": 2}
					]
				},
				"listening": {
					"questions": [
						{"question": "q3", "answers": ["e", "f"], "correct_answer": 1}
					]
				}
			}
		}`),
		Answer: json.RawMessage(`{"grammar":{"1":"a"},"language":{"1":"c"},"listening":{}}`),
	}

	score, total, err := scoreCalculation(exercise)
	if err != nil {
		t.Fatalf("score calculation failed: %v", err)
	}
	if score != 1 || total != 3 {
		t.Fatalf("score = %d, total = %d, want 1 of 3", score, total)
	}
}

func TestReviewNoteFromModel(t *testing.T) {
	raw := "```json\n{\"note\":\"Good work.\",\"items\":[]}\n```"
	note, err := reviewNoteFromModel(raw)
	if err != nil {
		t.Fatalf("review note parse failed: %v", err)
	}

	var parsed reviewNote
	if err = json.Unmarshal([]byte(note), &parsed); err != nil {
		t.Fatalf("stored note is not json: %v", err)
	}
	if parsed.Note != "Good work." {
		t.Fatalf("note = %q", parsed.Note)
	}
	if parsed.Items == nil {
		t.Fatal("items should be an empty list")
	}
}

func TestReviewNoteFromModelRejectsEmptyNote(t *testing.T) {
	_, err := reviewNoteFromModel(`{"note":"  ","items":[]}`)
	if err == nil {
		t.Fatal("empty note should be rejected")
	}
}
