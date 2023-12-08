package fromflags

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

var blf = ByLanguageFlag("test_bylanguage_flag", "Usage")

func TestByLanguageFlagSetValues(t *testing.T) {
	err := flag.Set("test_bylanguage_flag", "en=English")
	assert.Nil(t, err)
	err = flag.Set("test_bylanguage_flag", "gl=Galician")
	assert.Nil(t, err)
	assert.Equal(t, map[string]string{"en": "English", "gl": "Galician"}, *blf)
}

func TestByLanguageFlagSetError(t *testing.T) {
	err := flag.Set("test_bylanguage_flag", "foobar")
	assert.Error(t, err)
}
