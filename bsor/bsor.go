package bsor

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"strconv"
)

type Header struct {
	Magic   uint32
	Version byte
}

type Info struct {
	ModVersion     string
	GameVersion    string
	Timestamp      uint32
	PlayerId       string
	PlayerName     string
	Platform       string
	TrackingSystem string
	Hmd            string
	Controller     string
	Hash           string
	SongName       string
	Mapper         string
	Difficulty     string
	Score          uint32
	Mode           string
	Environment    string
	Modifiers      string
	JumpDistance   float32
	LeftHanded     bool
	Height         float32
	StartTime      float32
	FailTime       float32
	Speed          float32
}

type Vector3 struct {
	X float32
	Y float32
	Z float32
}

type Position Vector3

type Rotation struct {
	Position
	W float32
}

type PositionAndRotation struct {
	Position Position
	Rotation Rotation
}

type Frame struct {
	Time      float32
	Fps       uint32
	Header    PositionAndRotation
	LeftHand  PositionAndRotation
	RightHand PositionAndRotation
}

type NoteEventType uint32

const (
	Good NoteEventType = iota
	Bad
	Miss
	Bomb
)

type NoteCutInfo struct {
	SpeedOk             bool
	DirectionOk         bool
	SaberTypeOk         bool
	WasCutTooSoon       bool
	SaberSpeed          float32
	SaberDir            Vector3
	SaberType           uint32
	TimeDeviation       float32
	CutDirDeviation     float32
	CutPoint            Vector3
	CutNormal           Vector3
	CutDistanceToCenter float32
	CutAngle            float32
	BeforeCutRating     float32
	AfterCutRating      float32
}

type Note struct {
	NoteId    uint32
	EventTime float32
	SpawnTime float32
	EventType NoteEventType
	CutInfo   NoteCutInfo
}

type Wall struct {
	WallId    uint32
	Energy    float32
	Time      float32
	SpawnTime float32
}

type Height struct {
	Height float32
	Time   float32
}

type Pause struct {
	Duration uint32
	Time     float32
}

type Bsor struct {
	Header  Header
	Info    Info
	Frames  []Frame
	Notes   []Note
	Walls   []Wall
	Heights []Height
	Pauses  []Pause
}

var byteOrder = binary.LittleEndian

func Read(file os.File, bsor *Bsor) (err error) {
	err = readHeader(file, &bsor.Header)
	if err != nil {
		return
	}

	_, err = readNextBytes(file, 1)
	if err != nil {
		return
	}

	err = readInfo(file, &bsor.Info)
	if err != nil {
		return
	}

	_, err = readNextBytes(file, 1)
	if err != nil {
		return
	}

	err = readFrames(file, &bsor.Frames)
	if err != nil {
		return
	}

	_, err = readNextBytes(file, 1)
	if err != nil {
		return
	}

	err = readNotes(file, &bsor.Notes)
	if err != nil {
		return
	}

	_, err = readNextBytes(file, 1)
	if err != nil {
		return
	}

	err = readWalls(file, &bsor.Walls)
	if err != nil {
		return
	}

	_, err = readNextBytes(file, 1)
	if err != nil {
		return
	}

	err = readHeights(file, &bsor.Heights)
	if err != nil {
		return
	}

	_, err = readNextBytes(file, 1)
	if err != nil {
		return
	}

	err = readPauses(file, &bsor.Pauses)
	if err != nil {
		return
	}

	return
}

func readHeader(file os.File, header *Header) (err error) {
	err = readAny(file, header, binary.Size(*header))
	if err != nil {
		return
	}

	if header.Magic != 0x442d3d69 {
		return errors.New("not a BSOR file")
	}

	if header.Version != 1 {
		return errors.New("unknown BSOR version")
	}

	return nil
}

func readInfo(file os.File, info *Info) (err error) {
	err = readString(file, &info.ModVersion)
	if err != nil {
		return
	}

	err = readString(file, &info.GameVersion)
	if err != nil {
		return
	}

	var str string
	err = readString(file, &str)
	if err != nil {
		return
	}
	timestampInt, err := strconv.Atoi(str)
	if err != nil {
		return
	}
	info.Timestamp = uint32(timestampInt)

	err = readString(file, &info.PlayerId)
	if err != nil {
		return
	}

	err = readString(file, &info.PlayerName)
	if err != nil {
		return
	}

	err = readString(file, &info.Platform)
	if err != nil {
		return
	}

	err = readString(file, &info.TrackingSystem)
	if err != nil {
		return
	}

	err = readString(file, &info.Hmd)
	if err != nil {
		return
	}

	err = readString(file, &info.Controller)
	if err != nil {
		return
	}

	err = readString(file, &info.Hash)
	if err != nil {
		return
	}

	err = readString(file, &info.SongName)
	if err != nil {
		return
	}

	err = readString(file, &info.Mapper)
	if err != nil {
		return
	}

	err = readString(file, &info.Difficulty)
	if err != nil {
		return
	}

	err = readAny(file, &info.Score, binary.Size(info.Score))
	if err != nil {
		return
	}

	err = readString(file, &info.Mode)
	if err != nil {
		return
	}

	err = readString(file, &info.Environment)
	if err != nil {
		return
	}

	err = readString(file, &info.Modifiers)
	if err != nil {
		return
	}

	err = readAny(file, &info.JumpDistance, binary.Size(info.JumpDistance))
	if err != nil {
		return
	}

	err = readAny(file, &info.LeftHanded, binary.Size(info.LeftHanded))
	if err != nil {
		return
	}

	err = readAny(file, &info.Height, binary.Size(info.Height))
	if err != nil {
		return
	}

	err = readAny(file, &info.StartTime, binary.Size(info.StartTime))
	if err != nil {
		return
	}

	err = readAny(file, &info.FailTime, binary.Size(info.FailTime))
	if err != nil {
		return
	}

	err = readAny(file, &info.Speed, binary.Size(info.Speed))
	if err != nil {
		return
	}

	return nil
}

func readFrames(file os.File, frames *[]Frame) (err error) {
	var framesCount uint32
	err = readUInt32(file, &framesCount)
	if err != nil {
		return
	}

	*frames = make([]Frame, framesCount)
	err = readAny(file, frames, binary.Size(*frames))

	return
}

func readNotes(file os.File, notes *[]Note) (err error) {
	var notesCount uint32
	err = readUInt32(file, &notesCount)
	if err != nil {
		return
	}

	*notes = make([]Note, notesCount)
	for i := range *notes {
		err = readAny(file, &(*notes)[i].NoteId, binary.Size((*notes)[i].NoteId))
		if err != nil {
			return
		}
		err = readAny(file, &(*notes)[i].EventTime, binary.Size((*notes)[i].EventTime))
		if err != nil {
			return
		}
		err = readAny(file, &(*notes)[i].SpawnTime, binary.Size((*notes)[i].SpawnTime))
		if err != nil {
			return
		}
		err = readAny(file, &(*notes)[i].EventType, binary.Size((*notes)[i].EventType))
		if err != nil {
			return
		}
		if (*notes)[i].EventType == Good {
			err = readAny(file, &(*notes)[i].CutInfo, binary.Size((*notes)[i].CutInfo))
			if err != nil {
				return
			}
		}
	}

	return
}

func readWalls(file os.File, walls *[]Wall) (err error) {
	var wallsCount uint32
	err = readUInt32(file, &wallsCount)
	if err != nil {
		return
	}

	*walls = make([]Wall, wallsCount)
	err = readAny(file, walls, binary.Size(*walls))

	return
}

func readHeights(file os.File, heights *[]Height) (err error) {
	var heightsCount uint32
	err = readUInt32(file, &heightsCount)
	if err != nil {
		return
	}

	*heights = make([]Height, heightsCount)
	err = readAny(file, heights, binary.Size(*heights))

	return
}

func readPauses(file os.File, pauses *[]Pause) (err error) {
	var pausesCount uint32
	err = readUInt32(file, &pausesCount)
	if err != nil {
		return
	}

	*pauses = make([]Pause, pausesCount)
	err = readAny(file, pauses, binary.Size(*pauses))

	return
}

func readAny(file os.File, out any, byteSize int) (err error) {
	data, err := readNextBytes(file, byteSize)
	if err != nil {
		return
	}

	buffer := bytes.NewBuffer(data)
	err = binary.Read(buffer, binary.LittleEndian, out)
	if err != nil {
		return
	}

	return nil
}

func readUInt32(file os.File, value *uint32) (err error) {
	uintBytes, err := readNextBytes(file, 4)
	if err != nil {
		return err
	}

	*value = byteOrder.Uint32(uintBytes)

	return nil
}

func readString(file os.File, str *string) (err error) {
	var size uint32
	err = readUInt32(file, &size)
	if err != nil {
		return
	}

	stringBytes, err := readNextBytes(file, int(size))
	if err != nil {
		return
	}

	*str = string(stringBytes)

	return nil
}

func readNextBytes(file os.File, number int) (data []byte, err error) {
	bytes := make([]byte, number)

	n, err := file.Read(bytes)
	if err != nil {
		return nil, err
	}

	if n != number {
		return nil, errors.New("unexpected end of file")
	}

	return bytes, nil
}
