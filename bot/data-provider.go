package bot

import (
	"github.com/kramerdust/indefibot/exegete"
)

type WordDataProvider interface {
	SetWordExpositor(word string, expositor exegete.Expositor)
	GetWordExpositor(word string) (exegete.Expositor, bool)
	SetAudioID(word string, ID string)
	GetAudioID(ID string) (string, bool)
}

type WordMap map[string]*wordEntry

type wordEntry struct {
	expositor exegete.Expositor
	audioID   string
}

func NewWordMap() WordDataProvider {
	return make(WordMap)
}

func (wm WordMap) SetWordExpositor(word string, expositor exegete.Expositor) {
	if _, ok := wm[word]; ok {
		wm[word].expositor = expositor
	} else {
		wm[word] = &wordEntry{expositor: expositor}
	}

}

func (wm WordMap) GetWordExpositor(word string) (exegete.Expositor, bool) {
	we, ok := wm[word]
	if ok {
		return we.expositor, ok
	}
	return nil, ok
}

func (wm WordMap) SetAudioID(word string, ID string) {
	if _, ok := wm[word]; ok {
		wm[word].audioID = ID
	} else {
		wm[word] = &wordEntry{audioID: ID}
	}
}

func (wm WordMap) GetAudioID(word string) (string, bool) {
	we, ok := wm[word]
	if ok {
		return we.audioID, ok
	}
	return "", ok
}
