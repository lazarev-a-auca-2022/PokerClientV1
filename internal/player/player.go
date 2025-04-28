// filepath: f:\PokerClientV1\internal\player\player.go
package player

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}
