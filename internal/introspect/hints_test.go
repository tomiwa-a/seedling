package introspect

import (
	"testing"

	"github.com/tomiwa-a/seedling/pkg/schema"
)

func TestDetectGeneratorHint(t *testing.T) {
	tests := []struct {
		name     string
		col      schema.Column
		want     schema.GeneratorHint
	}{
		{"email column", schema.Column{Name: "email"}, schema.HintEmail},
		{"email_address", schema.Column{Name: "email_address"}, schema.HintEmail},
		{"full_name", schema.Column{Name: "full_name"}, schema.HintName},
		{"first_name", schema.Column{Name: "first_name"}, schema.HintName},
		{"last_name", schema.Column{Name: "last_name"}, schema.HintName},
		{"city", schema.Column{Name: "city"}, schema.HintCity},
		{"billing_city", schema.Column{Name: "billing_city"}, schema.HintCity},
		{"country", schema.Column{Name: "country"}, schema.HintCountry},
		{"phone", schema.Column{Name: "phone"}, schema.HintPhone},
		{"phone_number", schema.Column{Name: "phone_number"}, schema.HintPhone},
		{"address", schema.Column{Name: "address"}, schema.HintAddress},
		{"company", schema.Column{Name: "company"}, schema.HintCompany},
		{"organization", schema.Column{Name: "organization"}, schema.HintCompany},
		{"job_title", schema.Column{Name: "job_title"}, schema.HintJobTitle},
		{"website_url", schema.Column{Name: "website_url"}, schema.HintURL},
		{"ip_address", schema.Column{Name: "ip_address"}, schema.HintIP},
		{"uuid", schema.Column{Name: "uuid"}, schema.HintUUID},
		{"currency_code", schema.Column{Name: "currency_code"}, schema.HintCurrency},
		{"created_at", schema.Column{Name: "created_at"}, schema.HintNow},
		{"inserted_at", schema.Column{Name: "inserted_at"}, schema.HintNow},
		{"updated_at", schema.Column{Name: "updated_at"}, schema.HintNow},
		{"fallback to auto", schema.Column{Name: "some_random_column"}, schema.HintAuto},
		{"id column", schema.Column{Name: "id"}, schema.HintAuto},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectGeneratorHint(tt.col)
			if got != tt.want {
				t.Errorf("detectGeneratorHint(%q) = %q, want %q", tt.col.Name, got, tt.want)
			}
		})
	}
}

func TestHintFromComment(t *testing.T) {
	tests := []struct {
		comment string
		want    schema.GeneratorHint
	}{
		{"", schema.HintAuto},
		{"email address of the user", schema.HintEmail},
		{"full name", schema.HintName},
		{"city name", schema.HintCity},
		{"country code", schema.HintCountry},
		{"phone number", schema.HintPhone},
		{"mailing address", schema.HintAddress},
		{"company name", schema.HintCompany},
		{"job title", schema.HintJobTitle},
		{"URL to avatar", schema.HintURL},
		{"website url", schema.HintURL},
		{"currency code", schema.HintCurrency},
		{"some random comment", schema.HintAuto},
	}

	for _, tt := range tests {
		t.Run(tt.comment, func(t *testing.T) {
			got := hintFromComment(tt.comment)
			if got != tt.want {
				t.Errorf("hintFromComment(%q) = %q, want %q", tt.comment, got, tt.want)
			}
		})
	}
}
