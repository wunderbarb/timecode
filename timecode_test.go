// v0.3.0
// Author: Wunderbarb
// Dec 2023

package timecode

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// All the values =have been tested with the following website: https://en.editingtools.io/timecode/ and
// http://www.michaelcinquin.com/tools/timecode_keykode

const (
	cFPS25 = 25.0
	cFPS24 = 24.0
)

var testCounter int

func TestNewFromFrame(t *testing.T) {
	require, assert := Describe(t)

	tests := []struct {
		frame      int
		fps        float64
		expRes     int
		expSuccess bool
	}{
		{0, cFPS25, 0, true},
		{59, cFPS25, 59, true},
		{0, FPS23976fps, 0, true},
		{59, FPS23976fps, 59, true},
		{-1, FPS2997, 0, false},
		{59, -24, 0, false},
	}
	for _, tt := range tests {
		tc, err := NewFromFrame(tt.fps, tt.frame)
		require.Equal(tt.expSuccess, err == nil)
		if err == nil {
			assert.Equal(tt.expRes, tc.Frame())
		}
	}
}

func TestNew(t *testing.T) {
	require, assert := Describe(t)

	tests := []struct {
		sec        float64
		fps        float64
		expRes     int
		expSuccess bool
	}{
		{0, cFPS25, 0, true},
		{0, FPS23976fps, 0, true},
		{1.0, cFPS25, 25, true},
		{1, cFPS24, 24, true},
		{1, FPS23976fps, 23, true},
		{59.0, cFPS25, 1475, true},
		{59.0, FPS23976fps, 1414, true},
		{59.0, cFPS24, 1416, true},

		{-1, FPS2997, 0, false},
		{59, -24, 0, false},
	}
	for i, tt := range tests {
		tc, err := New(tt.fps, tt.sec)
		require.Equal(tt.expSuccess, err == nil)
		if err == nil {
			assert.Equal(tt.expRes, tc.Frame(), "sample %d", i+1)
		}
	}
}
func TestNewFromString(t *testing.T) {
	require, assert := Describe(t)

	tests := []struct {
		tc         string
		fps        float64
		expRes     int
		expSuccess bool
	}{
		{"00:00:00:00", cFPS25, 0, true},
		{"00:00:00:00", FPS23976fps, 0, true},
		{"00:00:01:00", cFPS25, 25, true},
		{"00:00:01:00", cFPS24, 24, true},
		{"00:00:01:00", FPS23976fps, 24, true},
		{"00:00:59:00", cFPS25, 1475, true},
		{"00:00:59:00", FPS23976fps, 1416, true},
		{"00:00:59:00", cFPS24, 1416, true},
		{"bad", FPS2997, 0, false},
		{"00:00:59:00", -24, 0, false},
	}
	for i, tt := range tests {
		tc, err := NewFromString(tt.fps, tt.tc)
		require.Equal(tt.expSuccess, err == nil)
		if err == nil {
			assert.Equal(tt.expRes, tc.Frame(), "sample %d", i+1)
		}
	}
	tests1 := []struct {
		tc         string
		expRes     int
		expSuccess bool
	}{
		{"00:00:00;00", 0, true},
		{"00:00:01;00", 30, true},
		{"00:00:59;29", 1799, true},
		{"00:01:00;00", 1800, false},
		{"00:01:00;02", 1800, true},
		{"00:02:00;02", 3598, true},
		{"00:10:00;00", 17982, true},
		{"00:10:00:00", 17982, false},
	}
	for i, tt := range tests1 {
		tc, err := NewWithDropFrameFromString(tt.tc)
		require.Equal(tt.expSuccess, err == nil, "sample %d", i+1)
		if err == nil {
			assert.Equal(tt.expRes, tc.Frame(), "sample %d", i+1)
		}
	}
}

func TestNewWithDropFrame(t *testing.T) {
	_, assert := Describe(t)

	_, err := NewWithDropFrame(-1.0)
	assert.Error(err)
}
func TestTimecode_String(t *testing.T) {
	require, assert := Describe(t)

	testsFrame := []struct {
		frame  int
		fps    float64
		expRes string
	}{
		{0, FPS23976fps, "00:00:00:00"},
		{1, FPS23976fps, "00:00:00:01"},
		{114, FPS23976fps, "00:00:04:18"},
	}
	for i, tt := range testsFrame {
		tc, err := NewFromFrame(tt.fps, tt.frame)
		require.NoError(err)
		assert.Equal(tt.expRes, tc.String(), "sample %d", i+1)
	}

	tests1 := []struct {
		tc    string
		frame int
	}{
		{"00:00:00;00", 0},
		{"00:00:01;00", 30},
		{"00:00:59;29", 1799},
		{"00:01:00;02", 1800},
		{"00:02:00;02", 3598},
		{"00:10:00;00", 17982},
	}
	for i, tt := range tests1 {
		tc, _ := NewWithDropFrameFromString(tt.tc)
		assert.Equal(tt.tc, tc.String(), "sample %d", i+1)
	}
}

func TestTimecode_AsMilliseconds(t *testing.T) {
	require, assert := Describe(t)

	tests := []struct {
		time   string
		fps    float64
		expRes string
	}{
		{"00:00:00:00", cFPS25, "00:00:00.000"},
		{"00:00:01:00", cFPS25, "00:00:01.000"},
		{"00:01:00:00", cFPS25, "00:01:00.000"},
		{"01:00:00:00", cFPS25, "01:00:00.000"},
		{"02:02:03:24", cFPS25, "02:02:03.960"},
	}
	for _, tt := range tests {
		tc, err := NewFromString(tt.fps, tt.time)
		require.NoError(err)
		assert.Equal(tt.expRes, tc.AsMilliseconds())
	}
}

func TestTimecode_SetFrame(t *testing.T) {
	_, assert := Describe(t)

	n := Rng.Intn(1000) + 1
	tests := []struct {
		frame     int
		expResult int
	}{
		{n, n},
		{-1, 0},
		{0, 0},
	}
	for _, tt := range tests {
		tc, _ := New(cFPS25, 0)
		tc.SetFrame(tt.frame)
		assert.Equal(tt.expResult, tc.Frame())
		assert.Equal(tt.expResult+1, tc.Frames())
	}
}

func TestTimecode_Add(t *testing.T) {
	require, assert := Describe(t)

	t1, _ := New(cFPS25, 60)
	t2, _ := New(cFPS25, 60)
	t3, _ := New(FPS2997, 120)
	t4, err := NewWithDropFrameFromString("00:05:00;02") // Drop frame, two frames are dropped every minute.
	require.NoError(err)
	t5, _ := NewWithDropFrameFromString("00:05:00;02")
	t6, _ := NewWithDropFrame(0)
	t7, _ := NewWithDropFrame(0)

	t6.SetFrame(107863) // DFTC 00:59:59;00
	t7.SetFrame(349799) // DFTC frame count duration
	_ = t6.Add(*t7)     // Add t6 & t7 to get out time code 04:14:30;19

	require.NoError(t1.Add(*t2))
	assert.Equal("00:02:00:00", t1.String())
	require.EqualError(t1.Add(*t3), ErrInconsistentFPS.Error())
	require.NoError(t4.Add(*t5))
	assert.Equal("00:10:00;02", t4.String())
	require.NoError(t4.Add(*t5))
	assert.Equal("00:15:00;04", t4.String())
	assert.Error(t4.Add(*t3))

	t10, _ := New(cFPS24, 2)
	t20, _ := New(cFPS24, cModulo24H)
	assert.NoError(t10.Add(*t20))
	assert.Equal("00:00:02:00", t10.String())

}

func TestTimecode_Parse(t *testing.T) {
	require, assert := Describe(t)
	tests := []struct {
		str        string
		fps        float64
		expSuccess bool
	}{
		{"00:00:00:00", cFPS25, true},
		{"12:34:56:22", cFPS25, true},
		{"a2:34:56:22", cFPS25, false},
		{"1a:34:56:22", cFPS25, false},
		{"12:a4:56:22", cFPS25, false},
		{"12:3a:56:22", cFPS25, false},
		{"12:34:a6:22", cFPS25, false},
		{"12:34:5a:22", cFPS25, false},
		{"12:34:56:a2", cFPS25, false},
		{"12:34:56:2a", cFPS25, false},
		{"12:64:56:22", cFPS25, false},
		{"12:34:66:22", cFPS25, false},
		{"12:34:56:25", cFPS25, false},
		{"12:34:56", cFPS25, false},
		{"00:00:01:00", FPS23976fps, true},
	}
	for i, tt := range tests {
		tc, err := NewFromString(tt.fps, tt.str)
		require.Equal(tt.expSuccess, err == nil, "sample %d", i+1)
		if err == nil {
			assert.Equal(tt.str, tc.String())
		}

	}
	tests2 := []struct {
		str        string
		expSuccess bool
	}{
		{"00:00:00;00", true},
		{"12:34:00;22", true},
		{"12:34:00;01", false},
	}
	for i, tt := range tests2 {
		tc, _ := NewWithDropFrame(0)
		err := tc.Parse(tt.str)
		require.Equal(tt.expSuccess, err == nil, "sample %d", i+1)
		if err == nil {
			assert.Equal(tt.str, tc.String(), "sample %d", i+1)
		}
	}
}

func TestTimecode_Subtract(t *testing.T) {
	require, assert := Describe(t)

	t1, _ := NewFromFrame(cFPS25, 60)
	t2, _ := NewFromFrame(cFPS25, 31)
	require.NoError(t1.Subtract(*t2))
	assert.Equal(29, t1.Frame())
	require.NoError(t1.Subtract(*t2))
	assert.Equal("23:59:59:23", t1.String())
	t3, _ := NewFromFrame(FPS2997, 60)
	assert.Error(t3.Subtract(*t1))
}

func TestTimecode_Offset(t *testing.T) {
	_, assert := Describe(t)

	const cFps = 24
	t1, _ := NewFromString(cFps, "00:00:00:00")
	t1.Offset(1)
	assert.Equal("00:00:00:01", t1.String())
	t1.Offset(1)
	assert.Equal("00:00:00:02", t1.String())
	t1.Offset(1)
	assert.Equal("00:00:00:03", t1.String())
	t1.Offset(-1)
	assert.Equal("00:00:00:02", t1.String())

	t1, _ = NewFromString(cFps, "00:00:01:00")
	t1.Offset(1)
	assert.Equal("00:00:01:01", t1.String())
	t1.Offset(1)
	assert.Equal("00:00:01:02", t1.String())
	t1.Offset(1)
	assert.Equal("00:00:01:03", t1.String())
	t1.Offset(-1)
	assert.Equal("00:00:01:02", t1.String())

	// drift test
	t2, _ := New(cFps, 0)
	const cNumFrames = 2000
	for i := 0; i < cNumFrames; i++ {
		t2.Offset(1)
	}
	t3, _ := New(cFps, 0)
	t3.Offset(cNumFrames)
	assert.Equal(t2.String(), t3.String())

	t10, _ := New(FPS23976fps, 0)
	assert.Equal("00:00:00:00", t10.String())
	for i := 1; i < 25; i++ {
		t10.Offset(1)
	}
	assert.Equal("00:00:01:00", t10.String())

}

func TestTimecode_AtOffsetFrom(t *testing.T) {
	_, assert := Describe(t)

	const cFps = 24
	t1 := RandomTimecode(cFps)
	t2 := t1
	n := Rng.Intn(1000) + 1
	t2.Offset(n)
	assert.True(t2.AtOffsetFrom(t1, n))
	assert.False(t2.AtOffsetFrom(t1, n+1))

	t3, _ := NewWithDropFrame(0)
	assert.False(t3.AtOffsetFrom(t1, 1))
	t4 := RandomTimecode(cFps + 1.0)
	assert.False(t4.AtOffsetFrom(t1, 1))

}

func TestTimecode_Clone(t *testing.T) {
	_, assert := Describe(t)

	t0 := RandomTimecode(24000.0 / 1001)
	t1 := Clone(&t0)
	assert.Equal(t0.String(), t1.String())
}

func TestTimecode_Equal(t *testing.T) {
	_, assert := Describe(t)

	t0 := RandomTimecode(FPS23976fps)
	t1 := Clone(&t0)
	assert.True(t0.Equal(*t1))
	t1.Offset(1)
	assert.False(t0.Equal(*t1))
	t2 := RandomTimecode(FPS23976fps)
	assert.False(t0.Equal(t2))
}

func TestTimecode_Milliseconds(t *testing.T) {
	_, assert := Describe(t)

	tt, _ := NewFromFrame(cFPS25, 25)
	assert.Equal(1000, tt.Milliseconds())

}
func TestTimecode_Before(t *testing.T) {
	_, assert := Describe(t)

	fr := _rng.Intn(12 * cNumSec * cNumSec)
	t1, _ := NewFromFrame(cFPS25, fr)
	t2, _ := NewFromFrame(cFPS25, fr+1)
	assert.False(t2.Before(*t1))
	assert.Equal(-1, t2.FrameCount(*t1))

}

func TestTimecode_Convert(t *testing.T) {
	_, assert := Describe(t)

	r1 := _rng.Intn(10000)
	t1, _ := NewFromFrame(cFPS25, r1)
	t2, _ := NewFromFrame(cFPS24, _rng.Intn(10000))
	t2.Convert(*t1)
	assert.Equal(r1, t2.Frame())
	assert.Equal(cFPS24, t2.fps)
}

func ExampleTimecode_Add() {
	t1, _ := NewWithDropFrame(0)
	t1.SetFrame(44970)
	t7, _ := NewWithDropFrame(3599)
	_ = t1.Add(*t7)
	fmt.Println(t1.String())
	//Output: 01:24:59;14
}

// Describe displays the rank of the test, the name of the function
// and its optional description provided by 'msg'.  It initializes an assert
// and a require function and returns them.
func Describe(t *testing.T, msg ...string) (*require.Assertions,
	*assert.Assertions) {

	dispMsg := ""
	if len(msg) != 0 {
		dispMsg = msg[0]
	}
	name := strings.TrimPrefix(strings.TrimPrefix(t.Name(), "Test"), "_")
	fmt.Printf("Test %d: %s %s\n", testCounter, name, dispMsg)
	testCounter++
	return require.New(t), assert.New(t)
}

// Rng is a randomly seeded random number generator that can be used for tests.
// The random number generator is not cryptographically safe.
var Rng *rand.Rand

// init initializes the random number generator.
func init() {
	Rng = rand.New(rand.NewSource(time.Now().UnixNano())) // #nosec G404  It is not crypto secure. OK for test

}
