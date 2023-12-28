// v0.3.3
// Author: Wunderbarb
// Dec 2023

// Package timecode manages SMPTE timecode.  Its reference is the frame count.  The first frame is always 0.
// It supports drop frames at 29.97 FPS.
// Currently, it does not support frame rate higher than 30.
package timecode

import (
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"time"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

const (
	// FPS2997 is the frame rate 29.97, i.e., 30000/1001.
	FPS2997 = 30000.0 / 1001.0
	// FPS23976fps is the frame rate 23.976, i.e., 24000/1001.
	FPS23976fps = 24000.0 / 1001.0

	cPrecision = 1000
	cModulo24H = 24 * cNumSec * cNumSec
	cNumSec    = 60
)

var (
	// ErrInvalidFPS is returned when the fps or duration is invalid.
	ErrInvalidFPS = errors.New("invalid fps or duration")
	// ErrInconsistentFPS is returned when adding time codes with different FPS or drop frames.
	ErrInconsistentFPS = errors.New("inconsistent fps")
	// ErrInvalidTimeCode is returned when the parsed timecode is not valid.
	ErrInvalidTimeCode = errors.New("invalid timecode")

	_rng = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// Timecode is a structure to handle video timecode as defined by SMPTE.
type Timecode struct {
	fps          float64
	currentFrame int
	dropFrame    bool
}

// New initializes a Timecode structure with the given fps and duration.
func New(fps float64, seconds float64) (*Timecode, error) {

	if fps <= 0.0 || seconds < 0.0 {
		return nil, ErrInvalidFPS
	}
	if seconds == 0.0 {
		return &Timecode{fps: fps, currentFrame: 0}, nil
	}
	s := decimal.NewFromFloat(seconds)
	f := decimal.NewFromFloat(fps)
	return &Timecode{fps: fps,
		currentFrame: int(s.Mul(f).IntPart())}, nil
}

// NewFromFrame initializes a Timecode structure with the given fps and frame.  The first frame
// is 0.
// Frame and frame rate must be positive.
func NewFromFrame(fps float64, frame int) (*Timecode, error) {
	if fps <= 0.0 || frame < 0 {
		return nil, ErrInvalidFPS
	}
	return &Timecode{fps: fps,
		currentFrame: frame}, nil
}

// NewFromString initializes a Timecode structure with the given fps and timecode provided as a string.
func NewFromString(fps float64, timecode string) (*Timecode, error) {
	tc, err := New(fps, 0.0)
	if err != nil {
		return nil, err
	}
	err = tc.Parse(timecode)
	if err != nil {
		return nil, err
	}
	return tc, nil
}

// NewWithDropFrame initializes a Timecode structure with drop frames. Its frame rate is 29.97.
func NewWithDropFrame(seconds float64) (*Timecode, error) {
	tc, err := New(FPS2997, seconds)
	if err != nil {
		return nil, err
	}
	tc.dropFrame = true
	return tc, nil
}

// NewWithDropFrameFromString initializes a Timecode structure with drop frames at 29.97.
func NewWithDropFrameFromString(timecode string) (*Timecode, error) {
	tc, err := NewWithDropFrame(0.0)
	if err != nil {
		return nil, err
	}
	err = tc.Parse(timecode)
	if err != nil {
		return nil, err
	}
	return tc, nil
}

// Add adds the timecode ta to the current timecode.  Their frame rate and drop frame must be the same.
func (t *Timecode) Add(ta Timecode) error {
	if !t.sameFrameRate(ta) {
		return ErrInconsistentFPS
	}
	modulo := cast2Int(cModulo24H * t.fps)
	t.currentFrame += ta.currentFrame
	if t.currentFrame >= modulo {
		t.currentFrame -= modulo
	}
	return nil
}

// AtOffsetFrom returns true if the timecode `t` is at offset `o` from the given timecode `ta`.
//
// The timecodes have to be with the same frame rate and drop frame.
func (t *Timecode) AtOffsetFrom(ta Timecode, o int) bool {
	if !t.sameFrameRate(ta) {
		return false
	}
	return ta.currentFrame+o == t.currentFrame
}

// Before returns true if the timecode `t` is before (or equal to) the given timecode `ta`.
func (t *Timecode) Before(ta Timecode) bool {
	return t.currentFrame <= ta.currentFrame
}

// Clone returns a clone of the timecode.
func Clone(t *Timecode) *Timecode {
	return &Timecode{fps: t.fps, currentFrame: t.currentFrame, dropFrame: t.dropFrame}
}

// Equal returns true if the timecode `t` is equal to the given timecode `ta`.
func (t *Timecode) Equal(ta Timecode) bool {
	return t.currentFrame == ta.currentFrame && t.fps == ta.fps && t.dropFrame == ta.dropFrame
}

// FrameCount returns the number of frames between the timecode `t` and the given timecode `ta`.
// Example: t0, _ := timecode.New(24.0, 0.0)
//
//	t1, _ := timecode.New(24.0, 1.0)
//	fmt.Println(t0.FrameCount(*t1))
//	// Output: 24
func (t *Timecode) FrameCount(ta Timecode) int {
	return ta.currentFrame - t.currentFrame
}

// Offset adds the given number of frames to the timecode.
// The number of frames may be negative.
// If the timecode becomes negative, it is set to 0.
func (t *Timecode) Offset(fra int) {
	t.currentFrame += fra
}

// AsMilliseconds returns the timecode as a properly formatted string. HH:MM:SS.ms
func (t *Timecode) AsMilliseconds() string {
	h1, m1, s1, ms := t.parse()
	return fmt.Sprintf("%02d:%02d:%02d.%03d", h1, m1, s1, cast2Int(cPrecision*ms))
}

// Convert method Converts from one timecode to another without changing the frame rate.
func (t *Timecode) Convert(ts Timecode) {
	t.currentFrame = ts.currentFrame
}

// Frame returns the frame number of the timecode.  The first frame is frame 0.
func (t *Timecode) Frame() int {
	return t.currentFrame
}

// Frames returns the number of frames in the timecode.
func (t *Timecode) Frames() int {
	return t.currentFrame + 1
}

// Milliseconds method returns the number of milliseconds in the timecode at the beginning of the frame.
func (t *Timecode) Milliseconds() int {
	s := decimal.NewFromInt(int64(t.currentFrame)).Mul(decimal.NewFromFloat(cPrecision))
	a := s.DivRound(decimal.NewFromFloat(t.fps), 3)
	return int(a.IntPart())
}

// Parse parses the given timecode string and sets the timecode accordingly.  The timecode must be in the format
// HH:MM:SS:fr or HH:MM:SS;ff. The frame `ff` must comply with the frame rate and drop frame of the timecode.
func (t *Timecode) Parse(ts string) error {
	const (
		cDlm1 = 2
		cDlm2 = 5
		cDlm3 = 8
	)
	if !regexp.MustCompile(`^\d{2}:[0-5]\d:[0-5]\d[:;][0-2]\d$`).MatchString(ts) {
		return ErrInvalidTimeCode
	}
	tsa := []rune(ts)
	h1 := extractHour(tsa[0], tsa[1])
	m1 := extractMin(tsa[cDlm1+1], tsa[cDlm1+2])
	s1 := extractMin(tsa[cDlm2+1], tsa[cDlm2+2])
	f := extractMin(tsa[cDlm3+1], tsa[cDlm3+2])
	if h1 == 0 && m1 == 0 && s1 == 0 && f == 0 {
		t.currentFrame = 0
		return nil
	}
	fr := cast2Round(t.fps)
	if f >= fr {
		return ErrInconsistentFPS
	}
	if !t.dropFrame {
		// We are placing our self at slightly after.  This allows us to avoid rounding issues.
		t.currentFrame = (cNumSec*cNumSec*h1+m1*cNumSec+s1)*fr + f
		return nil
	}
	if tsa[cDlm3] != ';' {
		return ErrInvalidTimeCode
	}

	if s1 == 0 && (f == 0 || f == 1) {
		switch m1 {
		case 0, 10, 20, 30, 40, 50:
		default:
			return ErrInvalidTimeCode
		}
	}
	// See https://www.davidheidelberger.com/2010/06/10/drop-frame-timecode/
	timeBase := int(math.RoundToEven(t.fps))
	cMinFrames := timeBase * cNumSec
	cHourFrames := cNumSec * cMinFrames
	totalMinutes := h1*cNumSec + m1
	t.currentFrame = h1*cHourFrames + m1*cMinFrames + s1*timeBase + f - 2*(totalMinutes-(totalMinutes/10))
	return nil
}

// SetFrame sets the timecode to the given frame.  The first frame is frame 0.
// If the frame is negative, it is set to 0.
func (t *Timecode) SetFrame(fra int) {
	if fra < 0 {
		fra = 0
	}
	t.currentFrame = fra
}

// String returns the timecode as a properly formatted string HH:MM:SS:ff.
func (t *Timecode) String() string {
	fra := cast2Round(t.fps)
	if !t.dropFrame {
		var cMin = cNumSec * fra
		var cHour = cNumSec * cMin
		h1 := t.currentFrame / cHour
		rem := t.currentFrame % cHour
		m1 := rem / cMin
		rem %= cMin
		s1 := rem / fra
		fr := t.currentFrame - (h1*cHour + m1*cMin + s1*fra)
		return fmt.Sprintf("%02d:%02d:%02d:%02d", h1, m1, s1, fr)
	}

	// See https://www.davidheidelberger.com/2010/06/10/drop-frame-timecode/
	dropFrames := 2 // round(framerate * .066666);
	framesPerHour := cast2Round(t.fps * cNumSec * cNumSec)
	framesPerDay := 24 * framesPerHour
	framesPer10Min := cast2Round(t.fps * 10 * cNumSec)
	framesPerMin := cNumSec*cast2Round(t.fps) - dropFrames
	frameNumber := t.currentFrame % framesPerDay
	d := frameNumber / framesPer10Min
	m := frameNumber % framesPer10Min
	frameNumber += 9 * d * dropFrames
	if m > dropFrames {
		frameNumber += dropFrames * ((m - dropFrames) / framesPerMin)
	}
	fr := frameNumber % fra
	s1 := (frameNumber / fra) % cNumSec
	m1 := ((frameNumber / fra) / cNumSec) % cNumSec
	h1 := (((frameNumber / fra) / cNumSec) / cNumSec) % 24
	return fmt.Sprintf("%02d:%02d:%02d;%02d", h1, m1, s1, fr)
}

// Subtract subtracts the timecode ta to the current timecode.
// Their frame rate and drop frame must be the same.
// When `ta` is greater than `t`, the result is modulo 24 hours, i.e., 00:00:01:00 - 00:00:02:00 = 23:59:59:00.
func (t *Timecode) Subtract(ta Timecode) error {
	if !t.sameFrameRate(ta) {
		return ErrInconsistentFPS
	}
	t.currentFrame -= ta.currentFrame
	if t.currentFrame < 0 {
		modulo := cast2Int(cModulo24H * t.fps)
		t.currentFrame += modulo
	}
	return nil
}

// RandomTimecode generates a random timecode with the frame rate `fps` in the range 0 to 12 hours.
func RandomTimecode(fps float64) Timecode {
	// We use `rand` because weak randomness is not an issue.
	t, _ := New(fps,
		float64(_rng.Intn(12*cNumSec*cNumSec*100)/100)) //nolint:gosec
	return *t
}

// parse parses the timecode and returns the hours, minutes, seconds and milliseconds.
func (t *Timecode) parse() (h1 int, m1 int, s1 int, ms float64) {
	const (
		cNumMinute = cNumSec
		cNumHour   = cNumMinute * cNumSec
	)
	if t.currentFrame == 0 {
		return
	}
	dur := t.duration()
	sec := cast2Int(dur)
	h1 = sec / cNumHour
	m := sec % cNumHour
	m1 = m / cNumMinute
	s1 = (sec % cNumHour) % cNumMinute
	_, ms = math.Modf(dur)
	return
}

func (t *Timecode) duration() float64 {
	//return float64(t.currentFrame)/t.fps + cOffsetForRoundingIssues
	f := decimal.NewFromInt(int64(t.currentFrame))
	a, _ := f.DivRound(decimal.NewFromFloat(t.fps), 3).Float64()
	return a
}

func (t *Timecode) sameFrameRate(ta Timecode) bool {
	if t.fps != ta.fps {
		return false
	}
	if t.dropFrame != ta.dropFrame {
		return false
	}
	return true
}

func extractMin(r1 rune, r2 rune) int {
	m1 := isDecimalMinSec(r1)
	m2 := isDigit(r2)
	return 10*m1 + m2
}

func extractHour(r1 rune, r2 rune) int {
	h1 := isDigit(r1)
	h2 := isDigit(r2)
	return 10*h1 + h2
}

func isDigit(r rune) int {
	return int(r - '0')
}

func isDecimalMinSec(r rune) int {
	return int(r - '0')
}

func cast2Int(x float64) int {
	return int(math.Trunc(x))
}

func cast2Round(x float64) int {
	return int(math.Round(x))
}
