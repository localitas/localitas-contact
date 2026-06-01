package contact

import (
	"testing"
	"time"

	"github.com/emersion/go-vcard"
)

func TestContactToVCard(t *testing.T) {
	c := &Contact{
		ID:        "test-123",
		Data:      "full_name: John Doe\nemails:\n  - john@example.com\n  - john@work.com\nphones:\n  - \"+1-555-1234\"\nwork:\n  current:\n    company: Acme Inc\n    title: Engineer\nsocial:\n  website: https://johndoe.com\n  linkedin: https://linkedin.com/in/johndoe",
		UpdatedAt: time.Date(2026, 4, 30, 12, 0, 0, 0, time.UTC),
	}

	card := contactToVCard(c)

	if fn := card.PreferredValue(vcard.FieldFormattedName); fn != "John Doe" {
		t.Errorf("FN = %q, want John Doe", fn)
	}
	if uid := card.PreferredValue(vcard.FieldUID); uid != "test-123@localitas" {
		t.Errorf("UID = %q", uid)
	}

	emails := card[vcard.FieldEmail]
	if len(emails) != 2 {
		t.Errorf("expected 2 emails, got %d", len(emails))
	}

	if org := card.PreferredValue(vcard.FieldOrganization); org != "Acme Inc" {
		t.Errorf("ORG = %q, want Acme Inc", org)
	}
	if title := card.PreferredValue(vcard.FieldTitle); title != "Engineer" {
		t.Errorf("TITLE = %q, want Engineer", title)
	}
}

func TestVCardToContactData(t *testing.T) {
	card := make(vcard.Card)
	card.SetValue(vcard.FieldFormattedName, "Jane Smith")
	card.Add(vcard.FieldEmail, &vcard.Field{Value: "jane@example.com"})
	card.Add(vcard.FieldTelephone, &vcard.Field{Value: "+1-555-5678"})
	card.SetValue(vcard.FieldOrganization, "Tech Corp")
	card.SetValue(vcard.FieldTitle, "CTO")
	card.Add(vcard.FieldURL, &vcard.Field{Value: "https://janesmith.com"})

	data := vcardToContactData(card)

	if data.FullName != "Jane Smith" {
		t.Errorf("FullName = %q", data.FullName)
	}
	if len(data.Emails) != 1 || data.Emails[0] != "jane@example.com" {
		t.Errorf("Emails = %v", data.Emails)
	}
	if len(data.Phones) != 1 || data.Phones[0] != "+1-555-5678" {
		t.Errorf("Phones = %v", data.Phones)
	}
	if data.Work.Current == nil || data.Work.Current.Company != "Tech Corp" {
		t.Errorf("Work.Current.Company = %v", data.Work.Current)
	}
	if data.Work.Current.Title != "CTO" {
		t.Errorf("Work.Current.Title = %q", data.Work.Current.Title)
	}
	if data.Social.Website != "https://janesmith.com" {
		t.Errorf("Social.Website = %q", data.Social.Website)
	}
}

func TestVCardRoundtrip(t *testing.T) {
	c := &Contact{
		ID:        "rt-456",
		Data:      "full_name: Bob Wilson\nemails:\n  - bob@test.com\nphones:\n  - \"555-9999\"\nwork:\n  current:\n    company: StartupCo\n    title: Founder",
		UpdatedAt: time.Date(2026, 4, 30, 12, 0, 0, 0, time.UTC),
	}

	card := contactToVCard(c)
	data := vcardToContactData(card)

	if data.FullName != "Bob Wilson" {
		t.Errorf("FullName = %q after roundtrip", data.FullName)
	}
	if len(data.Emails) != 1 || data.Emails[0] != "bob@test.com" {
		t.Errorf("Emails = %v after roundtrip", data.Emails)
	}
	if data.Work.Current == nil || data.Work.Current.Company != "StartupCo" {
		t.Errorf("Work = %v after roundtrip", data.Work.Current)
	}
}

func TestExtractContactID(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/carddav/default/abc123.vcf", "abc123"},
		{"/carddav/default/abc123", "abc123"},
		{"/carddav/abc123.vcf", "abc123"},
		{"/carddav/default/", ""},
	}
	for _, tt := range tests {
		got := extractContactID(tt.path)
		if got != tt.want {
			t.Errorf("extractContactID(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestContactToAddressObject(t *testing.T) {
	c := &Contact{
		ID:        "obj-789",
		Data:      "full_name: Test User",
		UpdatedAt: time.Date(2026, 4, 30, 12, 0, 0, 0, time.UTC),
	}

	obj := contactToAddressObject(c)
	if obj.Path != "/carddav/default/obj-789.vcf" {
		t.Errorf("Path = %q", obj.Path)
	}
	if obj.ETag == "" {
		t.Error("expected non-empty ETag")
	}
	if obj.Card == nil {
		t.Error("expected non-nil Card")
	}
}
