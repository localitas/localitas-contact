package contact

import (
	"time"
)

type Contact struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	Data      string    `json:"data"`
	Embedding []float32 `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ContactData struct {
	FullName string   `yaml:"full_name" json:"full_name" comment:"Contact's full name"`
	Emails   []string `yaml:"emails,omitempty" json:"emails,omitempty" comment:"Email addresses"`
	Phones   []string `yaml:"phones,omitempty" json:"phones,omitempty" comment:"Phone numbers"`

	Social SocialInfo `yaml:"social,omitempty" json:"social,omitempty" comment:"Social media profiles"`

	Work WorkInfo `yaml:"work,omitempty" json:"work,omitempty" comment:"Work history"`

	Interactions []Interaction `yaml:"interactions,omitempty" json:"interactions,omitempty" comment:"Meeting notes and interactions"`

	Metadata map[string]interface{} `yaml:"metadata,omitempty" json:"metadata,omitempty" comment:"Custom fields (tags, location, notes, etc.)"`
}

type SocialInfo struct {
	LinkedIn  string `yaml:"linkedin,omitempty" json:"linkedin,omitempty" comment:"LinkedIn profile URL or username"`
	X         string `yaml:"x,omitempty" json:"x,omitempty" comment:"X (Twitter) handle"`
	Instagram string `yaml:"instagram,omitempty" json:"instagram,omitempty"`
	WhatsApp  string `yaml:"whatsapp,omitempty" json:"whatsapp,omitempty"`
	GitHub    string `yaml:"github,omitempty" json:"github,omitempty" comment:"GitHub username"`
	Facebook  string `yaml:"facebook,omitempty" json:"facebook,omitempty"`
	TikTok    string `yaml:"tiktok,omitempty" json:"tiktok,omitempty"`
	YouTube   string `yaml:"youtube,omitempty" json:"youtube,omitempty"`
	Website   string `yaml:"website,omitempty" json:"website,omitempty" comment:"Personal or company website"`
}

type WorkInfo struct {
	Current  *Job  `yaml:"current,omitempty" json:"current,omitempty" comment:"Current job"`
	Previous []Job `yaml:"previous,omitempty" json:"previous,omitempty" comment:"Previous jobs"`
}

type Job struct {
	Company string `yaml:"company" json:"company"`
	Title   string `yaml:"title" json:"title"`
	Started string `yaml:"started,omitempty" json:"started,omitempty" comment:"Start date (YYYY-MM)"`
	Ended   string `yaml:"ended,omitempty" json:"ended,omitempty" comment:"End date (YYYY-MM)"`
}

type Interaction struct {
	Type  string `yaml:"type" json:"type" comment:"meeting, call, email, coffee, etc."`
	Date  string `yaml:"date" json:"date" comment:"Date of interaction (YYYY-MM-DD)"`
	Notes string `yaml:"notes,omitempty" json:"notes,omitempty"`
}

type LinkedInData struct {
	Headline string   `yaml:"headline,omitempty" json:"headline,omitempty"`
	Summary  string   `yaml:"summary,omitempty" json:"summary,omitempty"`
	Skills   []string `yaml:"skills,omitempty" json:"skills,omitempty"`
}
