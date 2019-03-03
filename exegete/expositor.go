package exegete

import "io"

type ExpositorProvider interface {
	GetWordExpositor(lang, word string) (Expositor, error)
}

type Expositor interface {
	Word() string
	GetAudio() (io.ReadCloser, error)
	GetSpelling() (string, error)
	GetSenses() []Sense
}

type Sense interface {
	GetDomains() []string
	GetExamples() []string
	GetDefinitions() []string
}
