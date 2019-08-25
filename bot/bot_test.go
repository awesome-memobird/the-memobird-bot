package bot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitCmdNPayload(t *testing.T) {
	for input, expect := range map[string]struct {
		Cmd     string
		Payload string
	}{
		"":                    {Cmd: "", Payload: ""},
		" ":                   {Cmd: "", Payload: " "},
		"/bind":               {Cmd: "/bind", Payload: ""},
		"/bind ":              {Cmd: "/bind", Payload: ""},
		"/bind  ":             {Cmd: "/bind", Payload: " "},
		"/bind something":     {Cmd: "/bind", Payload: "something"},
		"/bind  spaces kept ": {Cmd: "/bind", Payload: " spaces kept "},
		"something":           {Cmd: "", Payload: "something"},
		" s p a c e s ":       {Cmd: "", Payload: " s p a c e s "},
		"new\nline":           {Cmd: "", Payload: "new\nline"},
	} {
		cmd, payload := splitCmdNPayload(input)
		assert.Equal(t, expect.Cmd, cmd)
		assert.Equal(t, expect.Payload, payload)
	}
}
