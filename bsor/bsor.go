package bsor

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

type ReplayFloat = float32
type ReplayInt = int32

type Version = byte
type LineValue = byte
type LayerValue = byte
type TimeValue = float32
type Modifier = string
type Score = ReplayInt

type Header struct {
	Magic   ReplayInt `json:"-"`
	Version Version   `json:"version"`
}

type Info struct {
	ModVersion     string      `json:"modVersion"`
	GameVersion    string      `json:"gameVersion"`
	TimeSet        time.Time   `json:"timeSet"`
	PlayerId       string      `json:"playerId"`
	PlayerName     string      `json:"playerName"`
	Platform       string      `json:"platform"`
	TrackingSystem string      `json:"trackingSystem"`
	Hmd            string      `json:"hmd"`
	Controller     string      `json:"controller"`
	Hash           string      `json:"hash"`
	SongName       string      `json:"songName"`
	Mapper         string      `json:"mapper"`
	Difficulty     string      `json:"difficulty"`
	Score          Score       `json:"score"`
	Mode           string      `json:"mode"`
	Environment    string      `json:"environment"`
	Modifiers      []Modifier  `json:"modifiers"`
	JumpDistance   ReplayFloat `json:"jumpDistance"`
	LeftHanded     bool        `json:"leftHanded"`
	Height         ReplayFloat `json:"height"`
	StartTime      TimeValue   `json:"startTime"`
	FailTime       TimeValue   `json:"failTime"`
	Speed          ReplayFloat `json:"speed"`
}

type Vector3 struct {
	X ReplayFloat `json:"x"`
	Y ReplayFloat `json:"y"`
	Z ReplayFloat `json:"z"`
}

type Position Vector3

type Rotation struct {
	Vector3
	W ReplayFloat `json:"w"`
}

type PositionAndRotation struct {
	Position Position `json:"position"`
	Rotation Rotation `json:"rotation"`
}

type PartType ReplayInt

const (
	InfoPart PartType = iota
	FramesPart
	NotesPart
	WallsPart
	HeightsPart
	PausesPart
)

func (s PartType) String() string {
	switch s {
	case InfoPart:
		return "Info"
	case FramesPart:
		return "Frames"
	case NotesPart:
		return "Notes"
	case WallsPart:
		return "Walls"
	case HeightsPart:
		return "Heights"
	case PausesPart:
		return "Pauses"
	default:
		return "Unknown"
	}
}

type Frame struct {
	Time      TimeValue           `json:"time"`
	Fps       ReplayInt           `json:"fps"`
	Head      PositionAndRotation `json:"head"`
	LeftHand  PositionAndRotation `json:"leftHand"`
	RightHand PositionAndRotation `json:"rightHand"`
}

type NoteEventType ReplayInt

const (
	Good NoteEventType = iota
	Bad
	Miss
	Bomb
)

func (s NoteEventType) String() string {
	switch s {
	case Good:
		return "Good"
	case Bad:
		return "Bad"
	case Miss:
		return "Miss"
	case Bomb:
		return "Bomb"
	default:
		return "Unknown"
	}
}

type NoteCutInfo struct {
	SpeedOk             bool        `json:"speedOk"`
	DirectionOk         bool        `json:"directionOk"`
	SaberTypeOk         bool        `json:"saberTypeOk"`
	WasCutTooSoon       bool        `json:"wasCutTooSoon"`
	SaberSpeed          ReplayFloat `json:"saberSpeed"`
	SaberDir            Vector3     `json:"saberDir"`
	SaberType           ReplayInt   `json:"saberType"`
	TimeDeviation       ReplayFloat `json:"timeDeviation"`
	CutDirDeviation     ReplayFloat `json:"cutDirDeviation"`
	CutPoint            Vector3     `json:"cutPoint"`
	CutNormal           Vector3     `json:"cutNormal"`
	CutDistanceToCenter ReplayFloat `json:"cutDistanceToCenter"`
	CutAngle            ReplayFloat `json:"cutAngle"`
	BeforeCutRating     ReplayFloat `json:"beforeCutRating"`
	AfterCutRating      ReplayFloat `json:"afterCutRating"`
}

type NoteScoringType byte

const (
	NormalOld NoteScoringType = iota
	Ignore
	NoScore
	Normal
	SliderHead
	SliderTail
	BurstSliderHead
	BurstSliderElement
)

func (s NoteScoringType) String() string {
	switch s {
	case NormalOld:
		return "NormalOld"
	case Ignore:
		return "Ignore"
	case NoScore:
		return "NoScore"
	case Normal:
		return "Normal"
	case SliderHead:
		return "SliderHead"
	case SliderTail:
		return "SliderTail"
	case BurstSliderHead:
		return "BurstSliderHead"
	case BurstSliderElement:
		return "BurstSliderElement"
	default:
		return "Unknown"
	}
}

type ColorType byte

const (
	Red ColorType = iota
	Blue
	NoColor = 255
)

func (s ColorType) String() string {
	switch s {
	case Red:
		return "Red"
	case Blue:
		return "Blue"
	case NoColor:
		return "NoColor"
	default:
		return "Unknown"
	}
}

type CutDirection byte

const (
	TopCenter CutDirection = iota
	BottomCenter
	MiddleLeft
	MiddleRight
	TopLeft
	TopRight
	BottomLeft
	BottomRight
	Dot
)

func (s CutDirection) String() string {
	switch s {
	case TopCenter:
		return "TopCenter"
	case BottomCenter:
		return "BottomCenter"
	case MiddleLeft:
		return "MiddleLeft"
	case MiddleRight:
		return "MiddleRight"
	case TopLeft:
		return "TopLeft"
	case TopRight:
		return "TopRight"
	case BottomLeft:
		return "BottomLeft"
	case BottomRight:
		return "BottomRight"
	case Dot:
		return "Dot"
	default:
		return "Unknown"
	}
}

type Note struct {
	ScoringType  NoteScoringType `json:"scoringType"`
	LineIdx      LineValue       `json:"lineIdx"`
	LineLayer    LayerValue      `json:"lineLayer"`
	ColorType    ColorType       `json:"colorType"`
	CutDirection CutDirection    `json:"cutDirection"`
	EventTime    TimeValue       `json:"eventTime"`
	SpawnTime    TimeValue       `json:"spawnTime"`
	EventType    NoteEventType   `json:"eventType"`
	CutInfo      NoteCutInfo     `json:"cutInfo"`
}

type WallHit struct {
	LineIdx      LineValue   `json:"lineIdx"`
	ObstacleType byte        `json:"obstacleType"`
	Width        byte        `json:"width"`
	Energy       ReplayFloat `json:"energy"`
	Time         TimeValue   `json:"time"`
	SpawnTime    TimeValue   `json:"spawnTime"`
}

type AutomaticHeight struct {
	Height ReplayFloat `json:"height"`
	Time   TimeValue   `json:"time"`
}

type Pause struct {
	Duration int64     `json:"duration"`
	Time     TimeValue `json:"time"`
}

type Replay struct {
	Header
	Info    Info              `json:"info"`
	Frames  []Frame           `json:"frames"`
	Notes   []Note            `json:"notes"`
	Walls   []WallHit         `json:"walls"`
	Heights []AutomaticHeight `json:"heights"`
	Pauses  []Pause           `json:"pauses"`
}

var byteOrder = binary.LittleEndian

type Error struct {
	msg string
}

func (e Error) Error() string { return e.msg }

var ErrNotBsorFile = Error{"not a BSOR file"}
var ErrUnknownBsorVersion = Error{"unknown BSOR version"}
var ErrUnknownPart = Error{"unknown file part"}
var ErrDecodeField = Error{"invalid value encountered"}

func wrapError(err error) error {
	var e *Error
	if errors.As(err, &e) {
		return fmt.Errorf("bsor read error: %w", e)
	}

	return fmt.Errorf("bsor read error: %v", err)
}

func clamp(value float64, min float64, max float64) float64 {
	return math.Min(math.Max(min, value), max)
}

func Read(reader io.Reader) (*Replay, error) {
	var replay Replay
	var err error

	if err = readHeader(reader, &replay.Header); err != nil {
		return nil, wrapError(err)
	}

	for {
		var partType PartType
		if partType, err = readPartType(reader); err != nil {
			if err == io.EOF {
				return &replay, nil
			}

			return nil, wrapError(err)
		}

		switch partType {
		case InfoPart:
			err = readInfo(reader, &replay.Info)

		case FramesPart:
			err = readWholeSlice(reader, &replay.Frames)

		case NotesPart:
			err = readNotes(reader, &replay.Notes)

		case WallsPart:
			err = readWalls(reader, &replay.Walls)

		case HeightsPart:
			err = readWholeSlice(reader, &replay.Heights)

		case PausesPart:
			err = readWholeSlice(reader, &replay.Pauses)

		default:
			return nil, wrapError(ErrUnknownPart)
		}

		if err != nil {
			return nil, wrapError(err)
		}

		if partType == PausesPart {
			return &replay, nil
		}
	}
}

func readPartType(reader io.Reader) (PartType, error) {
	partBytes, err := readBytes(reader, 1)
	if err != nil {
		return 0, err
	}

	return PartType(partBytes[0]), nil
}

func readHeader(reader io.Reader, header *Header) error {
	if err := readAny(reader, header); err != nil {
		return err
	}

	if header.Magic != 0x442d3d69 {
		return ErrNotBsorFile
	}

	if header.Version != 1 {
		return ErrUnknownBsorVersion
	}

	return nil
}

func readInfo(reader io.Reader, info *Info) (err error) {
	if info.ModVersion, err = readString(reader); err != nil {
		return err
	}

	if info.GameVersion, err = readString(reader); err != nil {
		return err
	}

	var str string
	if str, err = readString(reader); err != nil {
		return err
	}
	timestampInt, err := strconv.Atoi(str)
	if err != nil {
		return ErrDecodeField
	}
	info.TimeSet = time.Unix(int64(timestampInt), 0)

	if info.PlayerId, err = readString(reader); err != nil {
		return err
	}

	if info.PlayerName, err = readPotentiallyInvalidString(reader); err != nil {
		return err
	}

	if info.Platform, err = readString(reader); err != nil {
		return err
	}

	if info.TrackingSystem, err = readString(reader); err != nil {
		return err
	}

	if info.Hmd, err = readString(reader); err != nil {
		return err
	}

	if info.Controller, err = readString(reader); err != nil {
		return err
	}

	if info.Hash, err = readString(reader); err != nil {
		return err
	}

	if info.SongName, err = readPotentiallyInvalidString(reader); err != nil {
		return err
	}

	if info.Mapper, err = readPotentiallyInvalidString(reader); err != nil {
		return err
	}

	if info.Difficulty, err = readString(reader); err != nil {
		return err
	}

	if err = readAny(reader, &info.Score); err != nil {
		return err
	}

	if info.Mode, err = readString(reader); err != nil {
		return err
	}

	if info.Environment, err = readString(reader); err != nil {
		return err
	}

	var modifiersCsv string
	if modifiersCsv, err = readString(reader); err != nil {
		return err
	}
	modifiers := strings.Split(modifiersCsv, ",")
	if len(modifiers) > 1 || len(modifiers[0]) > 0 {
		info.Modifiers = modifiers
	} else {
		info.Modifiers = []Modifier{}
	}

	if err = readAny(reader, &info.JumpDistance); err != nil {
		return err
	}

	if err = readAny(reader, &info.LeftHanded); err != nil {
		return err
	}

	if err = readAny(reader, &info.Height); err != nil {
		return err
	}

	if err = readAny(reader, &info.StartTime); err != nil {
		return err
	}

	if err = readAny(reader, &info.FailTime); err != nil {
		return err
	}

	if err = readAny(reader, &info.Speed); err != nil {
		return err
	}

	return nil
}

func readWholeSlice[T any](reader io.Reader, slice *[]T) (err error) {
	var sliceLength ReplayInt
	if sliceLength, err = readBsorInt(reader); err != nil {
		return
	}

	*slice = make([]T, sliceLength)

	return readAny(reader, slice)
}

func readNotes(reader io.Reader, notes *[]Note) (err error) {
	var notesCount ReplayInt
	if notesCount, err = readBsorInt(reader); err != nil {
		return
	}

	*notes = make([]Note, notesCount)
	for i := range *notes {
		var noteId ReplayInt
		if noteId, err = readBsorInt(reader); err != nil {
			return
		}

		(*notes)[i].ScoringType = NoteScoringType(noteId / 10000)
		noteId = noteId % 10000
		(*notes)[i].LineIdx = LineValue(noteId / 1000)
		noteId = noteId % 1000
		(*notes)[i].LineLayer = LayerValue(noteId / 100)
		noteId = noteId % 100
		(*notes)[i].ColorType = ColorType(noteId / 10)
		noteId = noteId % 10
		(*notes)[i].CutDirection = CutDirection(noteId)

		if err = readAny(reader, &(*notes)[i].EventTime); err != nil {
			return
		}
		if err = readAny(reader, &(*notes)[i].SpawnTime); err != nil {
			return
		}
		if err = readAny(reader, &(*notes)[i].EventType); err != nil {
			return
		}
		if (*notes)[i].EventType == Good || (*notes)[i].EventType == Bad {
			if err = readAny(reader, &(*notes)[i].CutInfo); err != nil {
				return
			}
		}
	}

	return
}

func readWalls(reader io.Reader, walls *[]WallHit) (err error) {
	var wallsCount ReplayInt
	if wallsCount, err = readBsorInt(reader); err != nil {
		return
	}

	*walls = make([]WallHit, wallsCount)
	for i := range *walls {
		var wallId ReplayInt
		if wallId, err = readBsorInt(reader); err != nil {
			return
		}
		(*walls)[i].LineIdx = LineValue(wallId / 100)
		wallId = wallId % 100
		(*walls)[i].ObstacleType = byte(wallId / 10)
		wallId = wallId % 10
		(*walls)[i].Width = byte(wallId)

		if err = readAny(reader, &(*walls)[i].Energy); err != nil {
			return
		}
		if err = readAny(reader, &(*walls)[i].Time); err != nil {
			return
		}
		if err = readAny(reader, &(*walls)[i].SpawnTime); err != nil {
			return
		}
	}

	return
}

func readAny(reader io.Reader, out any) error {
	return binary.Read(reader, binary.LittleEndian, out)
}

func readBsorInt(reader io.Reader) (value ReplayInt, err error) {
	var uintBytes = make([]byte, 4)

	if uintBytes, err = readBytes(reader, 4); err != nil {
		return 0, err
	}

	return ReplayInt(byteOrder.Uint32(uintBytes)), nil
}

func readStringWithLength(reader io.Reader, length int) (str string, err error) {
	stringBytes, err := readBytes(reader, length)
	if err != nil {
		return "", err
	}

	return string(stringBytes), nil
}

func skipResidualsOfIncorrectPreviousStringLength(reader io.Reader, length int) (int, error) {
	bytes := make([]byte, 4)
	byteOrder.PutUint32(bytes[0:], uint32(length))

	var b uint8
	if err := binary.Read(reader, binary.LittleEndian, &b); err != nil {
		return 0, err
	}

	bytes = bytes[1:]
	bytes = append(bytes, b)

	length = int(byteOrder.Uint32(bytes))
	if length > 255 || length < 0 {
		return skipResidualsOfIncorrectPreviousStringLength(reader, length)
	}

	return length, nil
}

func readPotentiallyInvalidStringWithLength(reader io.Reader, length int) (str string, err error) {
	if length > 255 || length < 0 {
		if length, err = skipResidualsOfIncorrectPreviousStringLength(reader, length); err != nil {
			return "", err
		}
	}

	return readStringWithLength(reader, length)
}

func readPotentiallyInvalidString(reader io.Reader) (str string, err error) {
	var length ReplayInt
	if length, err = readBsorInt(reader); err != nil {
		return "", err
	}

	readSeeker, implementsSeeker := interface{}(reader).(io.Seeker)

	if implementsSeeker && length > 0 {
		originalOffset, err := readSeeker.Seek(0, io.SeekCurrent)
		if err != nil {
			return readPotentiallyInvalidStringWithLength(reader, int(length))
		}

		if _, err = readSeeker.Seek(int64(length), io.SeekCurrent); err != nil {
			return readPotentiallyInvalidStringWithLength(reader, int(length))
		}

		var nextPossibleLength ReplayInt

		if nextPossibleLength, err = readBsorInt(reader); err != nil {
			readSeeker.Seek(originalOffset, io.SeekStart)

			return readPotentiallyInvalidStringWithLength(reader, int(length))
		}

		for nextPossibleLength < 0 || nextPossibleLength > 255 {
			if _, err = readSeeker.Seek(-3, io.SeekCurrent); err != nil {
				readSeeker.Seek(originalOffset, io.SeekStart)

				return readPotentiallyInvalidStringWithLength(reader, int(length))
			}

			length++

			if nextPossibleLength, err = readBsorInt(reader); err != nil {
				readSeeker.Seek(originalOffset, io.SeekStart)

				return readPotentiallyInvalidStringWithLength(reader, int(length))
			}
		}

		readSeeker.Seek(int64(-length)-int64(binary.Size(nextPossibleLength)), io.SeekCurrent)

		return readStringWithLength(reader, int(length))
	}

	return readPotentiallyInvalidStringWithLength(reader, int(length))
}

func readString(reader io.Reader) (str string, err error) {
	var length ReplayInt
	if length, err = readBsorInt(reader); err != nil {
		return "", err
	}

	return readStringWithLength(reader, int(length))
}

func readBytes(reader io.Reader, number int) (data []byte, err error) {
	bytes := make([]byte, number)

	if _, err := io.ReadFull(reader, bytes); err != nil {
		return nil, err
	}

	return bytes, nil
}
