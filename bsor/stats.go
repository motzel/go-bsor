package bsor

import (
	"github.com/motzel/go-bsor/buffer"
	"github.com/motzel/go-bsor/utils"
	"math"
)

const BlockMaxValue = 115

func GetMaxScore(blocks int, maxScorePerBlock int) int {
	var score int

	if blocks >= 14 {
		score += 8 * maxScorePerBlock * (blocks - 13)
	}

	if blocks >= 6 {
		score += 4 * maxScorePerBlock * (int(math.Min(float64(blocks), float64(13))) - 5)
	}

	if blocks >= 2 {
		score += 2 * maxScorePerBlock * (int(math.Min(float64(blocks), float64(5))) - 1)
	}

	score += maxScorePerBlock * int(math.Min(float64(blocks), float64(1)))

	return int(math.Floor(float64(score)))
}

type StatInfo struct {
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
	Score          int32    `json:"score"`
	Mode           string   `json:"mode"`
	Environment    string   `json:"environment"`
	Modifiers      []string `json:"modifiers"`
	JumpDistance   float32  `json:"jumpDistance"`
	LeftHanded     bool     `json:"leftHanded"`
	Height         float32  `json:"height"`
	StartTime      float32  `json:"startTime"`
	FailTime       float32  `json:"failTime"`
	Speed          float32  `json:"speed"`
	Accuracy       float64  `json:"accuracy"`
	WallHit        int      `json:"wallHit"`
	Pauses         int      `json:"pauses"`
}

type HandStat struct {
	AvgAccCut       float64   `json:"avgAccCut"`
	AvgBeforeCut    float64   `json:"avgBeforeCut"`
	AvgAfterCut     float64   `json:"avgAfterCut"`
	AvgScore        float64   `json:"avgScore"`
	AvgScorePercent float64   `json:"avgScorePercent"`
	TimeDependence  float64   `json:"timeDependence"`
	PreSwing        float64   `json:"preSwing"`
	PostSwing       float64   `json:"postSwing"`
	Grid            []float64 `json:"grid"`
	Notes           int       `json:"notes"`
	Misses          int       `json:"misses"`
	BadCuts         int       `json:"badCuts"`
	BombHits        int       `json:"bombHit"`
}

type Stats struct {
	Left  HandStat `json:"left"`
	Right HandStat `json:"right"`
	Total HandStat `json:"total"`
}

type ReplayStats struct {
	Info  StatInfo `json:"info"`
	Stats Stats    `json:"stats"`
}

func (stat *HandStat) calcPercentAcc() {
	stat.AvgScorePercent = stat.AvgScore / BlockMaxValue * 100
}

type StatBuffer struct {
	AccCut         buffer.Buffer[byte, int64]
	BeforeCut      buffer.Buffer[byte, int64]
	AfterCut       buffer.Buffer[byte, int64]
	Score          buffer.Buffer[byte, int64]
	TimeDependence buffer.Buffer[float64, float64]
	PreSwing       buffer.Buffer[float64, float64]
	PostSwing      buffer.Buffer[float64, float64]
	Grid           []buffer.Buffer[byte, int64]
	Notes          int
	Misses         int
	BadCuts        int
	BombHits       int
}

func (buf *StatBuffer) add(goodNoteCut *GoodCutEvent) {
	score := goodNoteCut.AccCut + goodNoteCut.BeforeCut + goodNoteCut.AfterCut

	if goodNoteCut.EventType == Good {
		if goodNoteCut.ScoringType != SliderTail && goodNoteCut.ScoringType != BurstSliderElement {
			buf.BeforeCut.Add(goodNoteCut.BeforeCut)
			buf.PreSwing.Add(float64(goodNoteCut.BeforeCutRating))
		}

		if goodNoteCut.ScoringType != BurstSliderHead && goodNoteCut.ScoringType != BurstSliderElement {
			buf.AccCut.Add(goodNoteCut.AccCut)
			buf.Score.Add(score)
			buf.TimeDependence.Add(float64(goodNoteCut.TimeDependence))
		}

		if goodNoteCut.ScoringType != SliderHead && goodNoteCut.ScoringType != BurstSliderHead && goodNoteCut.ScoringType != BurstSliderElement {
			buf.AfterCut.Add(goodNoteCut.AfterCut)
			buf.PostSwing.Add(float64(goodNoteCut.AfterCutRating))
		}

		if goodNoteCut.ScoringType != BurstSliderHead && goodNoteCut.ScoringType != BurstSliderElement {
			index := (2-goodNoteCut.LineLayer)*4 + goodNoteCut.LineIdx
			if index < 0 || index > 11 {
				index = 0
			}
			buf.Grid[index].Add(score)
		}
	}
}

func (buf *StatBuffer) stat() *HandStat {
	stat := &HandStat{
		AvgAccCut:      buf.AccCut.Avg(),
		AvgBeforeCut:   buf.BeforeCut.Avg(),
		AvgAfterCut:    buf.AfterCut.Avg(),
		AvgScore:       buf.Score.Avg(),
		TimeDependence: buf.TimeDependence.Avg(),
		PreSwing:       buf.PreSwing.Avg() * 100,
		PostSwing:      buf.PostSwing.Avg() * 100,
		Grid:           utils.SliceMap[buffer.Buffer[byte, int64], float64](buf.Grid, func(buf buffer.Buffer[byte, int64]) float64 { return buf.Avg() }),
		Notes:          buf.Notes,
		Misses:         buf.Misses,
		BadCuts:        buf.BadCuts,
		BombHits:       buf.BombHits,
	}

	stat.calcPercentAcc()

	return stat
}

func newStatBuffer(length int) *StatBuffer {
	accCutBuffer := buffer.NewBuffer[byte, int64](length)
	beforeCutBuffer := buffer.NewBuffer[byte, int64](length)
	afterCutBuffer := buffer.NewBuffer[byte, int64](length)
	accBuffer := buffer.NewBuffer[byte, int64](length)

	gridBuffer := [12]buffer.Buffer[byte, int64]{}

	return &StatBuffer{
		AccCut:    accCutBuffer,
		BeforeCut: beforeCutBuffer,
		AfterCut:  afterCutBuffer,
		Score:     accBuffer,
		Grid: utils.SliceMap[buffer.Buffer[byte, int64], buffer.Buffer[byte, int64]](
			gridBuffer[:],
			func(buf buffer.Buffer[byte, int64]) buffer.Buffer[byte, int64] {
				return buffer.NewBuffer[byte, int64](length)
			},
		),
	}
}

func newStatInfo(info *Info) *StatInfo {
	return &StatInfo{
		ModVersion:     info.ModVersion,
		GameVersion:    info.GameVersion,
		Timestamp:      info.Timestamp,
		PlayerId:       info.PlayerId,
		PlayerName:     info.PlayerName,
		Platform:       info.Platform,
		TrackingSystem: info.TrackingSystem,
		Hmd:            info.Hmd,
		Controller:     info.Controller,
		Hash:           info.Hash,
		SongName:       info.SongName,
		Mapper:         info.Mapper,
		Difficulty:     info.Difficulty,
		Score:          info.Score,
		Mode:           info.Mode,
		Environment:    info.Environment,
		Modifiers:      info.Modifiers,
		JumpDistance:   info.JumpDistance,
		LeftHanded:     info.LeftHanded,
		Height:         info.Height,
		StartTime:      info.StartTime,
		FailTime:       info.FailTime,
		Speed:          info.Speed,
	}
}

func NewReplayStats(replay *ReplayEvents) *ReplayStats {
	replayStats := &ReplayStats{Info: *newStatInfo(&replay.Info)}

	leftBuf := newStatBuffer(len(replay.Hits))
	rightBuf := newStatBuffer(len(replay.Hits))
	totalBuf := newStatBuffer(len(replay.Hits))

	for i := range replay.Hits {
		isLeft := (replay.Info.LeftHanded && replay.Hits[i].ColorType == Blue) || (!replay.Info.LeftHanded && replay.Hits[i].ColorType == Red)
		isEligibleNoteEvent := replay.Hits[i].ScoringType == Normal || replay.Hits[i].ScoringType == NormalOld || replay.Hits[i].ScoringType == SliderHead || replay.Hits[i].ScoringType == BurstSliderHead

		if isLeft {
			leftBuf.add(&replay.Hits[i])

			if isEligibleNoteEvent {
				leftBuf.Notes++
			}
		} else {
			rightBuf.add(&replay.Hits[i])

			if isEligibleNoteEvent {
				rightBuf.Notes++
			}
		}

		totalBuf.add(&replay.Hits[i])

		if isEligibleNoteEvent {
			totalBuf.Notes++
		}
	}

	for i := range replay.Misses {
		isLeft := (replay.Info.LeftHanded && replay.Misses[i].ColorType == Blue) || (!replay.Info.LeftHanded && replay.Misses[i].ColorType == Red)
		isEligibleNoteEvent := replay.Misses[i].ScoringType == Normal || replay.Misses[i].ScoringType == NormalOld || replay.Misses[i].ScoringType == SliderHead || replay.Misses[i].ScoringType == BurstSliderHead

		if isLeft {
			leftBuf.Misses++

			if isEligibleNoteEvent {
				leftBuf.Notes++
			}
		} else {
			rightBuf.Misses++

			if isEligibleNoteEvent {
				rightBuf.Notes++
			}
		}

		totalBuf.Misses++

		if isEligibleNoteEvent {
			totalBuf.Notes++
		}
	}

	for i := range replay.BadCuts {
		isLeft := (replay.Info.LeftHanded && replay.BadCuts[i].ColorType == Blue) || (!replay.Info.LeftHanded && replay.BadCuts[i].ColorType == Red)

		isEligibleNoteEvent := replay.BadCuts[i].ScoringType == Normal || replay.BadCuts[i].ScoringType == NormalOld || replay.BadCuts[i].ScoringType == SliderHead || replay.BadCuts[i].ScoringType == BurstSliderHead

		if isLeft {
			leftBuf.BadCuts++

			if isEligibleNoteEvent {
				leftBuf.Notes++
			}
		} else {
			rightBuf.BadCuts++

			if isEligibleNoteEvent {
				rightBuf.Notes++
			}
		}

		totalBuf.BadCuts++

		if isEligibleNoteEvent {
			totalBuf.Notes++
		}
	}

	for i := range replay.BombHits {
		isLeft := (replay.Info.LeftHanded && replay.BombHits[i].ColorType == Blue) || (!replay.Info.LeftHanded && replay.BombHits[i].ColorType == Red)

		if isLeft {
			leftBuf.BombHits++
		} else {
			rightBuf.BombHits++
		}

		totalBuf.BombHits++
	}

	replayStats.Stats.Left = *leftBuf.stat()
	replayStats.Stats.Right = *rightBuf.stat()
	replayStats.Stats.Total = *totalBuf.stat()

	maxScore := GetMaxScore(replayStats.Stats.Total.Notes, BlockMaxValue)
	if maxScore > 0 {
		replayStats.Info.Accuracy = float64(replayStats.Info.Score) / float64(maxScore) * 100
	}

	replayStats.Info.Pauses = len(replay.Pauses)
	replayStats.Info.WallHit = len(replay.Walls)

	return replayStats
}
