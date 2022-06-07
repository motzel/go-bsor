package bsor

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Header struct {
	Magic   uint32 `json:"-"`
	Version byte   `json:"version"`
}

type Info struct {
	ModVersion     string   `json:"modVersion"`
	GameVersion    string   `json:"gameVersion"`
	Timestamp      uint32   `json:"timestamp"`
	PlayerId       string   `json:"playerId"`
	PlayerName     string   `json:"playerName"`
	Platform       string   `json:"platform"`
	TrackingSystem string   `json:"trackingSystem"`
	Hmd            string   `json:"hmd"`
	Controller     string   `json:"controller"`
	Hash           string   `json:"hash"`
	SongName       string   `json:"songName"`
	Mapper         string   `json:"mapper"`
	Difficulty     string   `json:"difficulty"`
	Score          uint32   `json:"score"`
	Mode           string   `json:"mode"`
	Environment    string   `json:"environment"`
	Modifiers      []string `json:"modifiers"`
	JumpDistance   float32  `json:"jumpDistance"`
	LeftHanded     bool     `json:"leftHanded"`
	Height         float32  `json:"height"`
	StartTime      float32  `json:"startTime"`
	FailTime       float32  `json:"failTime"`
	Speed          float32  `json:"speed"`
}

type Vector3 struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

type Position Vector3

type Rotation struct {
	Position
	W float32 `json:"w"`
}

type PositionAndRotation struct {
	Position Position `json:"position"`
	Rotation Rotation `json:"rotation"`
}

type PartType uint32

const (
	InfoPart PartType = iota
	FramesPart
	NotesPart
	WallsPart
	HeightsPart
	PausesPart
)

type Frame struct {
	Time      float32             `json:"time"`
	Fps       uint32              `json:"fps"`
	Head      PositionAndRotation `json:"head"`
	LeftHand  PositionAndRotation `json:"leftHand"`
	RightHand PositionAndRotation `json:"rightHand"`
}

type NoteEventType uint32

const (
	Good NoteEventType = iota
	Bad
	Miss
	Bomb
)

type NoteCutInfo struct {
	SpeedOk             bool    `json:"speedOk"`
	DirectionOk         bool    `json:"directionOk"`
	SaberTypeOk         bool    `json:"saberTypeOk"`
	WasCutTooSoon       bool    `json:"wasCutTooSoon"`
	SaberSpeed          float32 `json:"saberSpeed"`
	SaberDir            Vector3 `json:"saberDir"`
	SaberType           uint32  `json:"saberType"`
	TimeDeviation       float32 `json:"timeDeviation"`
	CutDirDeviation     float32 `json:"cutDirDeviation"`
	CutPoint            Vector3 `json:"cutPoint"`
	CutNormal           Vector3 `json:"cutNormal"`
	CutDistanceToCenter float32 `json:"cutDistanceToCenter"`
	CutAngle            float32 `json:"cutAngle"`
	BeforeCutRating     float32 `json:"beforeCutRating"`
	AfterCutRating      float32 `json:"afterCutRating"`
}

type Note struct {
	LineIdx      byte          `json:"lineIdx"`
	LineLayer    byte          `json:"lineLayer"`
	ColorType    byte          `json:"colorType"`
	CutDirection byte          `json:"cutDirection"`
	EventTime    float32       `json:"eventTime"`
	SpawnTime    float32       `json:"spawnTime"`
	EventType    NoteEventType `json:"eventType"`
	CutInfo      NoteCutInfo   `json:"cutInfo"`
}

type Wall struct {
	LineIdx      byte    `json:"lineIdx"`
	ObstacleType byte    `json:"obstacleType"`
	Width        byte    `json:"width"`
	Energy       float32 `json:"energy"`
	Time         float32 `json:"time"`
	SpawnTime    float32 `json:"spawnTime"`
}

type Height struct {
	Height float32 `json:"height"`
	Time   float32 `json:"time"`
}

type Pause struct {
	Duration uint32  `json:"duration"`
	Time     float32 `json:"time"`
}

type Bsor struct {
	Header
	Info    Info     `json:"info"`
	Frames  []Frame  `json:"frames"`
	Notes   []Note   `json:"notes"`
	Walls   []Wall   `json:"walls"`
	Heights []Height `json:"heights"`
	Pauses  []Pause  `json:"pauses"`
}

var byteOrder = binary.LittleEndian

type BsorError struct {
	msg string
}

func (e BsorError) Error() string { return e.msg }

var ErrNotBsorFile = BsorError{"not a BSOR file"}
var ErrUnknownBsorVersion = BsorError{"unknown BSOR version"}
var ErrUnknownPart = BsorError{"unknown file part"}
var ErrDecodeField = BsorError{"invalid value encountered"}

func wrapError(err error) error {
	var e *BsorError
	if errors.As(err, &e) {
		return fmt.Errorf("bsor read error: %w", e)
	}

	return fmt.Errorf("bsor read error: %v", err)
}

func Read(reader io.Reader) (*Bsor, error) {
	var bsor Bsor
	var err error

	if err = readHeader(reader, &bsor.Header); err != nil {
		return &bsor, wrapError(err)
	}

	for {
		var partType PartType
		if partType, err = readPartType(reader); err != nil {
			if err == io.EOF {
				return &bsor, nil
			}

			return nil, wrapError(err)
		}

		switch partType {
		case InfoPart:
			err = readInfo(reader, &bsor.Info)

		case FramesPart:
			err = readWholeSlice(reader, &bsor.Frames)

		case NotesPart:
			err = readNotes(reader, &bsor.Notes)

		case WallsPart:
			err = readWalls(reader, &bsor.Walls)

		case HeightsPart:
			err = readWholeSlice(reader, &bsor.Heights)

		case PausesPart:
			err = readWholeSlice(reader, &bsor.Pauses)

		default:
			return &bsor, wrapError(ErrUnknownPart)
		}

		if err != nil {
			return nil, wrapError(err)
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
	if err := readAny(reader, header, binary.Size(*header)); err != nil {
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
	info.Timestamp = uint32(timestampInt)

	if info.PlayerId, err = readString(reader); err != nil {
		return err
	}

	if info.PlayerName, err = readString(reader); err != nil {
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

	if info.SongName, err = readString(reader); err != nil {
		return err
	}

	if info.Mapper, err = readString(reader); err != nil {
		return err
	}

	if info.Difficulty, err = readString(reader); err != nil {
		return err
	}

	if err = readAny(reader, &info.Score, binary.Size(info.Score)); err != nil {
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
	info.Modifiers = strings.Split(modifiersCsv, ",")

	if err = readAny(reader, &info.JumpDistance, binary.Size(info.JumpDistance)); err != nil {
		return err
	}

	if err = readAny(reader, &info.LeftHanded, binary.Size(info.LeftHanded)); err != nil {
		return err
	}

	if err = readAny(reader, &info.Height, binary.Size(info.Height)); err != nil {
		return err
	}

	if err = readAny(reader, &info.StartTime, binary.Size(info.StartTime)); err != nil {
		return err
	}

	if err = readAny(reader, &info.FailTime, binary.Size(info.FailTime)); err != nil {
		return err
	}

	if err = readAny(reader, &info.Speed, binary.Size(info.Speed)); err != nil {
		return err
	}

	return nil
}

func readWholeSlice[T any](reader io.Reader, slice *[]T) (err error) {
	var sliceLength uint32
	if sliceLength, err = readUInt32(reader); err != nil {
		return
	}

	*slice = make([]T, sliceLength)

	return readAny(reader, slice, binary.Size(*slice))
}

func readNotes(reader io.Reader, notes *[]Note) (err error) {
	var notesCount uint32
	if notesCount, err = readUInt32(reader); err != nil {
		return
	}

	*notes = make([]Note, notesCount)
	for i := range *notes {
		var noteId uint32
		if noteId, err = readUInt32(reader); err != nil {
			return
		}
		(*notes)[i].LineIdx = byte(noteId / 1000)
		noteId = noteId % 1000
		(*notes)[i].LineLayer = byte(noteId / 100)
		noteId = noteId % 100
		(*notes)[i].ColorType = byte(noteId / 10)
		noteId = noteId % 10
		(*notes)[i].CutDirection = byte(noteId)

		if err = readAny(reader, &(*notes)[i].EventTime, binary.Size((*notes)[i].EventTime)); err != nil {
			return
		}
		if err = readAny(reader, &(*notes)[i].SpawnTime, binary.Size((*notes)[i].SpawnTime)); err != nil {
			return
		}
		if err = readAny(reader, &(*notes)[i].EventType, binary.Size((*notes)[i].EventType)); err != nil {
			return
		}
		if (*notes)[i].EventType == Good || (*notes)[i].EventType == Bad {
			if err = readAny(reader, &(*notes)[i].CutInfo, binary.Size((*notes)[i].CutInfo)); err != nil {
				return
			}
		}
	}

	return
}

func readWalls(reader io.Reader, walls *[]Wall) (err error) {
	var wallsCount uint32
	if wallsCount, err = readUInt32(reader); err != nil {
		return
	}

	*walls = make([]Wall, wallsCount)
	for i := range *walls {
		var wallId uint32
		if wallId, err = readUInt32(reader); err != nil {
			return
		}
		(*walls)[i].LineIdx = byte(wallId / 100)
		wallId = wallId % 100
		(*walls)[i].ObstacleType = byte(wallId / 10)
		wallId = wallId % 10
		(*walls)[i].Width = byte(wallId)

		if err = readAny(reader, &(*walls)[i].Energy, binary.Size((*walls)[i].Energy)); err != nil {
			return
		}
		if err = readAny(reader, &(*walls)[i].Time, binary.Size((*walls)[i].Time)); err != nil {
			return
		}
		if err = readAny(reader, &(*walls)[i].SpawnTime, binary.Size((*walls)[i].SpawnTime)); err != nil {
			return
		}
	}

	return
}

func readAny(reader io.Reader, out any, byteSize int) error {
	return binary.Read(reader, binary.LittleEndian, out)
}

func readUInt32(reader io.Reader) (value uint32, err error) {
	var uintBytes = make([]byte, 4)

	if uintBytes, err = readBytes(reader, 4); err != nil {
		return 0, err
	}

	return byteOrder.Uint32(uintBytes), nil
}

func readString(reader io.Reader) (str string, err error) {
	var size uint32
	if size, err = readUInt32(reader); err != nil {
		return
	}

	stringBytes, err := readBytes(reader, int(size))
	if err != nil {
		return
	}

	return string(stringBytes), nil
}

func readBytes(reader io.Reader, number int) (data []byte, err error) {
	bytes := make([]byte, number)

	if _, err := io.ReadFull(reader, bytes); err != nil {
		return nil, err
	}

	return bytes, nil
}
