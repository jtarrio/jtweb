package fromflags

import (
	"flag"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var tf = TimeFlag("test_time_flag", "Usage")

func TestTimeFlagSetValue(t *testing.T) {
	err := flag.Set("test_time_flag", "2012-06-18T15:45:18Z")
	assert.Nil(t, err)
	assert.Equal(t, time.Date(2012, 6, 18, 15, 45, 18, 0, time.UTC), *tf)
}

func TestTimeFlagSetError(t *testing.T) {
	err := flag.Set("test_time_flag", "2012-06-18T15:45:18")
	actual := &time.ParseError{}
	assert.ErrorAs(t, err, &actual)
}
