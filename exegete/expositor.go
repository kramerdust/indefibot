package exegete

import "io"

type Expositor interface {
	GetAudio() (io.ReadCloser, error)
	GetSpelling() (string, error)
	GetSenses() []Sense
}

type Sense interface {
	GetDomains() []string
	GetExamples() []string
	GetDefinitions() []string
}
