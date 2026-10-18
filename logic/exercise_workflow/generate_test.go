package exercise_workflow

import "testing"

func TestSanitizeChatGenerateJSON(t *testing.T) {
	plain := `{"title":"N3"}`
	got, err := sanitizeChatGenerateJSON("```json\n" + plain + "\n```")
	if err != nil {
		t.Fatal(err)
	}
	if got != plain {
		t.Fatalf("fenced json = %q", got)
	}

	got, err = sanitizeChatGenerateJSON("Here is the exercise:\n" + plain + "\nDone.")
	if err != nil {
		t.Fatal(err)
	}
	if got != plain {
		t.Fatalf("prose json = %q", got)
	}

	if _, err = sanitizeChatGenerateJSON("not json"); err == nil {
		t.Fatal("expected invalid json to fail")
	}

	wrapped := "<think>skip {this}</think>\n```json\n{\"title\":\"N3\",}\n```\nnote }"
	got, err = sanitizeChatGenerateJSON(wrapped)
	if err != nil {
		t.Fatal(err)
	}
	if got != plain {
		t.Fatalf("wrapped json = %q", got)
	}
}
