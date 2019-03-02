package bot

import (
	"github.com/kramerdust/indefibot/exegete"
)

type UserDataProvider interface {
	SetUserExpositor(userID int64, expositor exegete.Expositor)
}

type UserMap map[int64]exegete.Expositor

func NewUserMap() UserDataProvider {
	return make(UserMap)
}

func (um UserMap) SetUserExpositor(userID int64, expositor exegete.Expositor) {
	um[userID] = expositor
}
