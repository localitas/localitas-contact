package contact

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

func SerializeContactWithComments(data ContactData) string {
	var sb strings.Builder

	sb.WriteString("# Contact's full name\n")
	sb.WriteString(fmt.Sprintf("full_name: %s\n\n", data.FullName))

	sb.WriteString("# Email addresses\n")
	sb.WriteString("emails:\n")
	if len(data.Emails) > 0 {
		for _, email := range data.Emails {
			if email != "" {
				sb.WriteString(fmt.Sprintf("  - %s\n", email))
			} else {
				sb.WriteString("  - \n")
			}
		}
	} else {
		sb.WriteString("  - \n")
	}
	sb.WriteString("\n")

	sb.WriteString("# Phone numbers\n")
	sb.WriteString("phones:\n")
	if len(data.Phones) > 0 {
		for _, phone := range data.Phones {
			if phone != "" {
				sb.WriteString(fmt.Sprintf("  - %s\n", phone))
			} else {
				sb.WriteString("  - \n")
			}
		}
	} else {
		sb.WriteString("  - \n")
	}
	sb.WriteString("\n")

	sb.WriteString("# Social media profiles\n")
	sb.WriteString("social:\n")
	serializeSocialInfo(&sb, data.Social)
	sb.WriteString("\n")

	sb.WriteString("# Work history\n")
	sb.WriteString("work:\n")
	serializeWorkInfo(&sb, data.Work)
	sb.WriteString("\n")

	sb.WriteString("# Meeting notes and interactions\n")
	sb.WriteString("interactions:\n")
	serializeInteractions(&sb, data.Interactions)
	sb.WriteString("\n")

	sb.WriteString("# Custom fields (tags, location, notes, etc.)\n")
	sb.WriteString("metadata:\n")
	serializeMetadata(&sb, data.Metadata)

	return sb.String()
}

func serializeSocialInfo(sb *strings.Builder, social SocialInfo) {
	t := reflect.TypeOf(social)
	v := reflect.ValueOf(social)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i).String()

		yamlTag := field.Tag.Get("yaml")
		yamlName := strings.Split(yamlTag, ",")[0]

		comment := field.Tag.Get("comment")
		if comment != "" {
			sb.WriteString(fmt.Sprintf("  # %s\n", comment))
		}
		sb.WriteString(fmt.Sprintf("  %s: %s\n", yamlName, value))
	}
}

func serializeWorkInfo(sb *strings.Builder, work WorkInfo) {
	sb.WriteString("  # Current job\n")
	sb.WriteString("  current:\n")
	if work.Current != nil {
		serializeJob(sb, *work.Current, "    ")
	} else {
		sb.WriteString("    company: \n")
		sb.WriteString("    title: \n")
		sb.WriteString("    # Start date (YYYY-MM)\n")
		sb.WriteString(fmt.Sprintf("    started: %s\n", time.Now().Format("2006-01")))
	}

	sb.WriteString("  # Previous jobs\n")
	sb.WriteString("  previous:\n")
	if len(work.Previous) > 0 {
		for _, job := range work.Previous {
			sb.WriteString("    - ")
			serializeJobInline(sb, job)
		}
	} else {
		sb.WriteString("    # - company: Previous Company\n")
		sb.WriteString("    #   title: Previous Title\n")
		sb.WriteString("    #   started: 2020-01\n")
		sb.WriteString("    #   ended: 2023-12\n")
	}
}

func serializeJob(sb *strings.Builder, job Job, indent string) {
	sb.WriteString(fmt.Sprintf("%scompany: %s\n", indent, job.Company))
	sb.WriteString(fmt.Sprintf("%stitle: %s\n", indent, job.Title))
	sb.WriteString(fmt.Sprintf("%s# Start date (YYYY-MM)\n", indent))
	started := job.Started
	if started == "" {
		started = time.Now().Format("2006-01")
	}
	sb.WriteString(fmt.Sprintf("%sstarted: %s\n", indent, started))
	if job.Ended != "" {
		sb.WriteString(fmt.Sprintf("%s# End date (YYYY-MM)\n", indent))
		sb.WriteString(fmt.Sprintf("%sended: %s\n", indent, job.Ended))
	}
}

func serializeJobInline(sb *strings.Builder, job Job) {
	sb.WriteString(fmt.Sprintf("company: %s\n", job.Company))
	sb.WriteString(fmt.Sprintf("      title: %s\n", job.Title))
	if job.Started != "" {
		sb.WriteString(fmt.Sprintf("      started: %s\n", job.Started))
	}
	if job.Ended != "" {
		sb.WriteString(fmt.Sprintf("      ended: %s\n", job.Ended))
	}
}

func serializeInteractions(sb *strings.Builder, interactions []Interaction) {
	if len(interactions) > 0 {
		for _, interaction := range interactions {
			sb.WriteString("  - # meeting, call, email, coffee, etc.\n")
			sb.WriteString(fmt.Sprintf("    type: %s\n", interaction.Type))
			sb.WriteString("    # Date of interaction (YYYY-MM-DD)\n")
			sb.WriteString(fmt.Sprintf("    date: %s\n", interaction.Date))
			if interaction.Notes != "" {
				sb.WriteString(fmt.Sprintf("    notes: %s\n", interaction.Notes))
			} else {
				sb.WriteString("    notes: \n")
			}
		}
	} else {
		sb.WriteString("  - # meeting, call, email, coffee, etc.\n")
		sb.WriteString("    type: meeting\n")
		sb.WriteString("    # Date of interaction (YYYY-MM-DD)\n")
		sb.WriteString(fmt.Sprintf("    date: %s\n", time.Now().UTC().Format("2006-01-02")))
		sb.WriteString("    notes: Initial meeting\n")
	}
}

func serializeMetadata(sb *strings.Builder, metadata map[string]interface{}) {
	if len(metadata) > 0 {
		for key, value := range metadata {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	} else {
		sb.WriteString("  # tags: [prospect, developer]\n")
		sb.WriteString("  # location: San Francisco, CA\n")
		sb.WriteString("  # birthday: 1990-05-15\n")
		sb.WriteString("  # notes: Met at tech conference\n")
	}
}

func NewContactYAML() string {
	data := ContactData{
		FullName: "New Contact",
	}
	return SerializeContactWithComments(data)
}
