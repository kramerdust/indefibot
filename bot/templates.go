package bot

type Card struct {
	Word          string
	Transcription string
	Page          int
	Definition    string
}

const CardTemplate = `
*{{.Word}}*  
_{{.Transcription}}_  
*{{.Page}}*: {{.Definition}}`
