package introspect

import (
	"strings"

	"github.com/tomiwa-a/seedling/pkg/schema"
)

func detectGeneratorHint(col *schema.Column) schema.GeneratorHint {
	name := strings.ToLower(col.Name)

	switch {
	case name == "email" || name == "email_address":
		return schema.HintEmail
	case name == "name" || name == "full_name" || strings.HasSuffix(name, "full_name"):
		return schema.HintName
	case name == "first_name" || name == "firstname" || strings.HasSuffix(name, "first_name"):
		return schema.HintName
	case name == "last_name" || name == "lastname" || strings.HasSuffix(name, "last_name"):
		return schema.HintName
	case name == "city" || strings.HasSuffix(name, "_city"):
		return schema.HintCity
	case name == "country" || strings.HasSuffix(name, "_country"):
		return schema.HintCountry
	case strings.HasSuffix(name, "phone") || strings.HasSuffix(name, "_phone") || name == "phone_number" || name == "telephone":
		return schema.HintPhone
	case name == "ip" || name == "ipv4" || name == "ipv6" || name == "ip_address":
		return schema.HintIP
	case strings.HasSuffix(name, "address") || name == "address" || name == "street":
		return schema.HintAddress
	case strings.HasSuffix(name, "company") || strings.HasSuffix(name, "_company") || name == "organization":
		return schema.HintCompany
	case strings.HasSuffix(name, "title") || name == "job_title":
		return schema.HintJobTitle
	case strings.HasSuffix(name, "url") || strings.HasSuffix(name, "website") || strings.HasSuffix(name, "link"):
		return schema.HintURL
	case name == "uuid" || strings.HasSuffix(name, "_uuid") || name == "guid":
		return schema.HintUUID
	case strings.HasSuffix(name, "currency") || name == "currency_code":
		return schema.HintCurrency
	case name == "created_at" || name == "inserted_at":
		return schema.HintNow
	case name == "updated_at" || name == "modified_at":
		return schema.HintNow
	}

	return schema.HintAuto
}

func hintFromComment(comment string) schema.GeneratorHint {
	if comment == "" {
		return schema.HintAuto
	}

	lower := strings.ToLower(strings.TrimSpace(comment))

	switch {
	case strings.HasPrefix(lower, "email"):
		return schema.HintEmail
	case strings.Contains(lower, "full name") || strings.Contains(lower, "full_name"):
		return schema.HintName
	case strings.HasPrefix(lower, "city"):
		return schema.HintCity
	case strings.HasPrefix(lower, "country"):
		return schema.HintCountry
	case strings.Contains(lower, "phone"):
		return schema.HintPhone
	case strings.Contains(lower, "address"):
		return schema.HintAddress
	case strings.Contains(lower, "company"):
		return schema.HintCompany
	case strings.Contains(lower, "job"):
		return schema.HintJobTitle
	case strings.Contains(lower, "url") || strings.Contains(lower, "website"):
		return schema.HintURL
	case strings.Contains(lower, "currency"):
		return schema.HintCurrency
	}

	return schema.HintAuto
}
