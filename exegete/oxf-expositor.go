package exegete

import (
	"fmt"
	"io"
	"net/http"

	oxf "github.com/kramerdust/go-oxford-client/v1/client"
)

type OxfExpositorProvider struct {
	oxfClient *oxf.Client
}
type OxfExpositor struct {
	*oxf.LexicalEntry
	client *http.Client
}

type OxfSense struct {
	*oxf.Sense
}

func NewOxfExpositorProvider(appID, appKey string) ExpositorProvider {
	return &OxfExpositorProvider{
		oxf.NewClient(appID, appKey),
	}
}

func (oep *OxfExpositorProvider) GetWordExpositor(lang, word string) (Expositor, error) {
	entry, err := oep.oxfClient.GetEntry(lang, word)
	if err != nil {
		return nil, fmt.Errorf("Getting oxford entry error: %s ", err)
	}
	return NewOxfExpositor(&entry.Results[0].LexicalEntries[0]), nil
}

func NewOxfExpositor(lex *oxf.LexicalEntry) Expositor {
	return &OxfExpositor{
		lex,
		&http.Client{},
	}
}

func (oe *OxfExpositor) GetAudio() (io.ReadCloser, error) {
	if oe.Pronunciations[0].AudioFile == "" {
		return nil, fmt.Errorf("No audio for this entry")
	}
	url := oe.Pronunciations[0].AudioFile
	resp, err := oe.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Http Get error : %s", err)
	}
	return resp.Body, nil
}

func (oe *OxfExpositor) GetSpelling() (string, error) {
	if oe.Pronunciations[0].PhoneticSpelling == "" {
		return "", fmt.Errorf("No spelling for this entry")
	}
	return oe.Pronunciations[0].PhoneticSpelling, nil
}

func (oe *OxfExpositor) GetSenses() []Sense {
	senses := make([]Sense, len(oe.Entries[0].Senses))
	for i := range oe.LexicalEntry.Entries[0].Senses {
		senses[i] = &OxfSense{&oe.LexicalEntry.Entries[0].Senses[i]}
	}
	return senses
}

func (s *OxfSense) GetDomains() []string {
	return s.Domains
}

func (s *OxfSense) GetExamples() []string {
	examples := make([]string, len(s.Examples))
	for i := range s.Examples {
		examples[i] = s.Examples[i].Text
	}
	return examples
}

func (s *OxfSense) GetDefinitions() []string {
	return s.Definitions
}
