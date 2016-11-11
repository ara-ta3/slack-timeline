package timeline

import (
	"testing"

	"../slack"
	"github.com/stretchr/testify/assert"
)

func TestReplaceUserIDToName(t *testing.T) {

	r := UserRepositoryOnMemory{data: map[string]slack.User{
		"U06ABGQEB": slack.User{ID: "U06ABGQEB", Name: "dark"},
	}}
	f := NewIDReplacerFactory(r)
	replacer, e := f.NewReplacer()
	if assert.NoError(t, e) {
		actual := replacer.Replace("<@U06ABGQEB>")
		assert.Equal(t, "@dark", actual)
	}
}
