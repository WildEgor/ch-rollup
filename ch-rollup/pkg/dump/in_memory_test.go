package dump

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInMemoryDump(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "test",
			input: "1",
			want:  "data",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dump := &InMemoryDumper{
				storage: make(map[string]string),
			}
			err := dump.Dump(tt.input, tt.want)
			assert.Nil(t, err)

			data := dump.storage[tt.input]
			assert.Equal(t, tt.want, data)
		})
	}
}

func TestInMemoryProcessNext(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		content string
	}{
		{
			name:    "success",
			input:   "1",
			content: "data",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dump := &InMemoryDumper{
				storage: make(map[string]string),
			}
			err := dump.Dump(tt.input, tt.content)
			assert.Nil(t, err)

			err = dump.ProcessNext(func(id string, content string) error {
				assert.Equal(t, tt.input, id)
				assert.Equal(t, tt.content, content)
				return nil
			})
			assert.Nil(t, err)
		})
	}
}
