package bsor

import (
	"github.com/motzel/go-bsor/bsor/buffer"
	"github.com/motzel/go-bsor/bsor/utils"
)

const BlockMaxValue = 115

type StatInfo struct {
	ReplayEventsInfo
	WallHit int `json:"wallHit"`
	Pauses  int `json:"pauses"`
}

type HandStat struct {
	AccCut         buffer.Stats[uint16]      `json:"accCut"`
	BeforeCut      buffer.Stats[uint16]      `json:"beforeCut"`
	AfterCut       buffer.Stats[uint16]      `json:"afterCut"`
	Score          buffer.Stats[uint16]      `json:"score"`
	TimeDependence buffer.Stats[float64]     `json:"timeDependence"`
	PreSwing       buffer.Stats[float64]     `json:"preSwing"`
	PostSwing      buffer.Stats[float64]     `json:"postSwing"`
	Grid           buffer.StatsSlice[uint16] `json:"grid"`
	Notes          int                       `json:"notes"`
	Misses         int                       `json:"misses"`
	BadCuts        int                       `json:"badCuts"`
	BombHits       int                       `json:"bombHit"`
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

type StatBuffer struct {
	AccCut         buffer.Buffer[uint16, int64]
	BeforeCut      buffer.Buffer[uint16, int64]
	AfterCut       buffer.Buffer[uint16, int64]
	Score          buffer.Buffer[uint16, int64]
	TimeDependence buffer.Buffer[float64, float64]
	PreSwing       buffer.Buffer[float64, float64]
	PostSwing      buffer.Buffer[float64, float64]
	Grid           []buffer.Buffer[uint16, int64]
	Notes          int
	Misses         int
	BadCuts        int
	BombHits       int
}

func (buf *StatBuffer) add(goodNoteCut *GoodNoteCutEvent) {
	score := goodNoteCut.AccCut + goodNoteCut.BeforeCut + goodNoteCut.AfterCut

	if goodNoteCut.EventType == Good {
		if goodNoteCut.ScoringType != SliderTail && goodNoteCut.ScoringType != BurstSliderElement {
			buf.BeforeCut.Add(uint16(goodNoteCut.BeforeCut))
			buf.PreSwing.Add(float64(goodNoteCut.BeforeCutRating))
		}

		if goodNoteCut.ScoringType != BurstSliderHead && goodNoteCut.ScoringType != BurstSliderElement {
			buf.AccCut.Add(uint16(goodNoteCut.AccCut))
			buf.Score.Add(uint16(score))
			buf.TimeDependence.Add(float64(goodNoteCut.TimeDependence))
		}

		if goodNoteCut.ScoringType != SliderHead && goodNoteCut.ScoringType != BurstSliderHead && goodNoteCut.ScoringType != BurstSliderElement {
			buf.AfterCut.Add(uint16(goodNoteCut.AfterCut))
			buf.PostSwing.Add(float64(goodNoteCut.AfterCutRating))
		}

		if goodNoteCut.ScoringType != BurstSliderHead && goodNoteCut.ScoringType != BurstSliderElement {
			index := (2-goodNoteCut.LineLayer)*4 + goodNoteCut.LineIdx
			if index < 0 || index > 11 {
				index = 0
			}
			buf.Grid[index].Add(uint16(score))
		}
	}
}

func (buf *StatBuffer) stat() *HandStat {
	stat := &HandStat{
		AccCut:         buf.AccCut.Stats(),
		BeforeCut:      buf.BeforeCut.Stats(),
		AfterCut:       buf.AfterCut.Stats(),
		Score:          buf.Score.Stats(),
		TimeDependence: buf.TimeDependence.Stats(),
		PreSwing:       buf.PreSwing.Stats(),
		PostSwing:      buf.PostSwing.Stats(),
		Grid: buffer.StatsSlice[uint16]{
			Min:    utils.SliceMap[buffer.Buffer[uint16, int64], uint16](buf.Grid, func(buf buffer.Buffer[uint16, int64]) uint16 { return buf.Min() }),
			Avg:    utils.SliceMap[buffer.Buffer[uint16, int64], float64](buf.Grid, func(buf buffer.Buffer[uint16, int64]) float64 { return buf.Avg() }),
			Median: utils.SliceMap[buffer.Buffer[uint16, int64], uint16](buf.Grid, func(buf buffer.Buffer[uint16, int64]) uint16 { return buf.Median() }),
			Max:    utils.SliceMap[buffer.Buffer[uint16, int64], uint16](buf.Grid, func(buf buffer.Buffer[uint16, int64]) uint16 { return buf.Max() }),
		},
		Notes:    buf.Notes,
		Misses:   buf.Misses,
		BadCuts:  buf.BadCuts,
		BombHits: buf.BombHits,
	}

	return stat
}

func newStatBuffer(length int) *StatBuffer {
	accCutBuffer := buffer.NewBuffer[uint16, int64](length)
	beforeCutBuffer := buffer.NewBuffer[uint16, int64](length)
	afterCutBuffer := buffer.NewBuffer[uint16, int64](length)
	accBuffer := buffer.NewBuffer[uint16, int64](length)

	gridBuffer := [12]buffer.Buffer[uint16, int64]{}

	return &StatBuffer{
		AccCut:    accCutBuffer,
		BeforeCut: beforeCutBuffer,
		AfterCut:  afterCutBuffer,
		Score:     accBuffer,
		Grid: utils.SliceMap[buffer.Buffer[uint16, int64], buffer.Buffer[uint16, int64]](
			gridBuffer[:],
			func(buf buffer.Buffer[uint16, int64]) buffer.Buffer[uint16, int64] {
				return buffer.NewBuffer[uint16, int64](length)
			},
		),
	}
}

func newStatInfo(info *ReplayEventsInfo) *StatInfo {
	return &StatInfo{
		ReplayEventsInfo: *info,
		WallHit:          0,
		Pauses:           0,
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

	replayStats.Info.Pauses = len(replay.Pauses)
	replayStats.Info.WallHit = len(replay.Walls)

	return replayStats
}
