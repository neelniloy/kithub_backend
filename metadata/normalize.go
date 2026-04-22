package metadata

import (
	"regexp"
	"strings"
	"unicode"
)

var nonWordRE = regexp.MustCompile(`[^a-z0-9]+`)

func Slug(value string) string {
	value = strings.ToLower(stripDiacritics(value))
	value = strings.ReplaceAll(value, "&", " and ")
	value = nonWordRE.ReplaceAllString(value, "_")
	value = strings.Trim(value, "_")
	value = regexp.MustCompile(`_+`).ReplaceAllString(value, "_")
	return value
}

func NormalizeAlias(value string) string {
	value = strings.ToLower(stripDiacritics(value))
	var builder strings.Builder
	for _, r := range value {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			builder.WriteRune(r)
		default:
			builder.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(builder.String()), " ")
}

func splitAliases(value string) []string {
	if value == "" {
		return nil
	}

	parts := regexp.MustCompile(`[,;/|]`).Split(value, -1)
	aliases := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			aliases = append(aliases, part)
		}
	}
	return aliases
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool)
	unique := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	return unique
}

func stripDiacritics(value string) string {
	replacer := strings.NewReplacer(
		"á", "a", "à", "a", "â", "a", "ä", "a", "ã", "a", "å", "a",
		"Á", "a", "À", "a", "Â", "a", "Ä", "a", "Ã", "a", "Å", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"É", "e", "È", "e", "Ê", "e", "Ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"Í", "i", "Ì", "i", "Î", "i", "Ï", "i",
		"ó", "o", "ò", "o", "ô", "o", "ö", "o", "õ", "o",
		"Ó", "o", "Ò", "o", "Ô", "o", "Ö", "o", "Õ", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u",
		"Ú", "u", "Ù", "u", "Û", "u", "Ü", "u",
		"ñ", "n", "Ñ", "n", "ç", "c", "Ç", "c",
		"ş", "s", "Ş", "s", "ğ", "g", "Ğ", "g",
		"ı", "i", "İ", "i",
	)
	return replacer.Replace(value)
}
