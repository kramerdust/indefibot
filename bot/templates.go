package bot

type Card struct {
	Word          string
	Transcription string
	Page          int
	Total         int
	Definitions   []string
}

const CardTemplate = `
*{{.Word}}*  
_[{{.Transcription}}]_  
{{range .Definitions}}- {{.}}  
{{end}}({{.Page}}/{{.Total}})`
