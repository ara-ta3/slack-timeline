package slack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceIdFormatToName1(t *testing.T) {
	actual := ReplaceIdFormatToName("<@UXXXXX|dark>")
	assert.Equal(t, "dark", actual)
}

func TestReplaceIdFormatToName2(t *testing.T) {
	actual := ReplaceIdFormatToName("<@UXXXXX|dark> uploaded a file: <https://xxx.slack.com/files/dark/FXXXXX/-.js|Untitled>")
	assert.Equal(t, "dark uploaded a file: <https://xxx.slack.com/files/dark/FXXXXX/-.js|Untitled>", actual)
}

func TestReplaceIdFormatToName3(t *testing.T) {
	actual := ReplaceIdFormatToName("hey i am <@UXXXXX|dark>")
	assert.Equal(t, "hey i am dark", actual)
}
