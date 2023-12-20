package work

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJobArgumentExtraction(t *testing.T) {
	j := Job{}
	j.setArg("str1", "bar")

	j.setArg("int1", int64(77))
	j.setArg("int2", 77)
	j.setArg("int3", uint64(77))
	j.setArg("int4", float64(77.0))

	j.setArg("bool1", true)

	j.setArg("float1", 3.14)

	// Success cases:
	vString := j.ArgString("str1")
	assert.Equal(t, vString, "bar")
	assert.NoError(t, j.ArgError())

	vInt64 := j.ArgInt64("int1")
	assert.EqualValues(t, vInt64, 77)
	assert.NoError(t, j.ArgError())

	vInt64 = j.ArgInt64("int2")
	assert.EqualValues(t, vInt64, 77)
	assert.NoError(t, j.ArgError())

	vInt64 = j.ArgInt64("int3")
	assert.EqualValues(t, vInt64, 77)
	assert.NoError(t, j.ArgError())

	vInt64 = j.ArgInt64("int4")
	assert.EqualValues(t, vInt64, 77)
	assert.NoError(t, j.ArgError())

	vBool := j.ArgBool("bool1")
	assert.Equal(t, vBool, true)
	assert.NoError(t, j.ArgError())

	vFloat := j.ArgFloat64("float1")
	assert.Equal(t, vFloat, 3.14)
	assert.NoError(t, j.ArgError())

	// Missing key results in error:
	vString = j.ArgString("str_missing")
	assert.Equal(t, vString, "")
	assert.Error(t, j.ArgError())
	j.argError = nil
	assert.NoError(t, j.ArgError())

	vInt64 = j.ArgInt64("int_missing")
	assert.EqualValues(t, vInt64, 0)
	assert.Error(t, j.ArgError())
	j.argError = nil
	assert.NoError(t, j.ArgError())

	vBool = j.ArgBool("bool_missing")
	assert.Equal(t, vBool, false)
	assert.Error(t, j.ArgError())
	j.argError = nil
	assert.NoError(t, j.ArgError())

	vFloat = j.ArgFloat64("float_missing")
	assert.Equal(t, vFloat, 0.0)
	assert.Error(t, j.ArgError())
	j.argError = nil
	assert.NoError(t, j.ArgError())

	// Missing string; Make sure we don't reset it with successes after
	vString = j.ArgString("str_missing")
	assert.Equal(t, vString, "")
	assert.Error(t, j.ArgError())
	_ = j.ArgString("str1")
	_ = j.ArgInt64("int1")
	_ = j.ArgBool("bool1")
	_ = j.ArgFloat64("float1")
	assert.Error(t, j.ArgError())
}
