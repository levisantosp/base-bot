package i18n

type (
	Type func(lang, content string, args ...Args) string
	Args map[string]any
)

func T(lang, content string, args ...Args) string {
	return content
}
