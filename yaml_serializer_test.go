package contact

import (
	"strings"
	"testing"
)

func TestSerializeContactWithComments_HappyPath(t *testing.T) {
	data := ContactData{
		FullName: "Alice Johnson",
		Emails:   []string{"alice@example.com", "alice.work@example.com"},
		Phones:   []string{"+15551234567"},
		Social: SocialInfo{
			LinkedIn: "linkedin.com/in/alice",
			GitHub:   "alicej",
		},
		Work: WorkInfo{
			Current: &Job{
				Company: "Acme",
				Title:   "Engineer",
				Started: "2024-01",
			},
		},
	}

	yaml := SerializeContactWithComments(data)

	must := []string{
		"full_name: Alice Johnson",
		"- alice@example.com",
		"- alice.work@example.com",
		"- +15551234567",
		"linkedin: linkedin.com/in/alice",
		"github: alicej",
		"company: Acme",
		"title: Engineer",
		"started: 2024-01",
	}
	for _, substr := range must {
		if !strings.Contains(yaml, substr) {
			t.Errorf("serialized YAML missing %q\n--- full output:\n%s", substr, yaml)
		}
	}
}

func TestNewContactYAML_IsNonEmpty(t *testing.T) {
	out := NewContactYAML()
	if !strings.Contains(out, "full_name: New Contact") {
		t.Errorf("NewContactYAML should include default name line, got:\n%s", out)
	}
}
