package bsor

import (
	"github.com/motzel/go-bsor/bsor/buffer"
	"github.com/motzel/go-bsor/bsor/utils"
)

const LayersCount = LayerValue(3)
const LinesCount = LineValue(4)

const BlockMaxValue = 115
const GridBufferSize = int(LinesCount) * int(LayersCount)

type BlockPosition byte

const (
	TopLeftBlockPosition BlockPosition = iota
	TopCenterLeftBlockPosition
	TopCenterRightBlockPosition
	TopRightBlockPosition
	MiddleLeftBlockPosition
	MiddleCenterLeftBlockPosition
	MiddleCenterRightBlockPosition
	MiddleRightBlockPosition
	BottomLeftBlockPosition
	BottomCenterLeftBlockPosition
	BottomCenterRightBlockPosition
	BottomRightBlockPosition
)

func (s BlockPosition) String() string {
	switch s {
	case TopLeftBlockPosition:
		return "TopLeft"
	case TopCenterLeftBlockPosition:
		return "TopCenterLeft"
	case TopCenterRightBlockPosition:
		return "TopCenterRight"
	case TopRightBlockPosition:
		return "TopRight"
	case MiddleLeftBlockPosition:
		return "MiddleLeft"
	case MiddleCenterLeftBlockPosition:
		return "MiddleCenterLeft"
	case MiddleCenterRightBlockPosition:
		return "MiddleCenterRight"
	case MiddleRightBlockPosition:
		return "MiddleRight"
	case BottomLeftBlockPosition:
		return "BottomLeft"
	case BottomCenterLeftBlockPosition:
		return "BottomCenterLeft"
	case BottomCenterRightBlockPosition:
		return "BottomCenterRight"
	case BottomRightBlockPosition:
		return "BottomRight"
	default:
		return "Unknown"
	}
}

func NewBlockPosition(layer LayerValue, line LineValue) BlockPosition {
	// layers in BS goes from the bottom to the top, let's reverse it
	index := (LayersCount-1-layer)*LinesCount + line

	if index < 0 || index > (LayersCount*LinesCount-1) {
		index = 0
	}

	return BlockPosition(index)
}

type CutBuffer = buffer.Buffer[CutValue, CutValueSum]
type SwingBuffer = buffer.Buffer[SwingValue, SwingValueSum]

type ReplayStatsInfo struct {
	Info
	EndTime    TimeValue  `json:"endTime"`
	Accuracy   SwingValue `json:"accuracy"`
	FcAccuracy SwingValue `json:"fcAccuracy"`
	CalcScore  Counter    `json:"calcScore"`
	WallHits   Counter    `json:"wallHits"`
	Pauses     Counter    `json:"pauses"`
}

type HandStat struct {
	AccCut         buffer.Stats[CutValue]      `json:"accCut"`
	BeforeCut      buffer.Stats[CutValue]      `json:"beforeCut"`
	AfterCut       buffer.Stats[CutValue]      `json:"afterCut"`
	Score          buffer.Stats[CutValue]      `json:"score"`
	TimeDependence buffer.Stats[SwingValue]    `json:"timeDependence"`
	PreSwing       buffer.Stats[SwingValue]    `json:"preSwing"`
	PostSwing      buffer.Stats[SwingValue]    `json:"postSwing"`
	Grid           buffer.StatsSlice[CutValue] `json:"grid"`
	Notes          Counter                     `json:"notes"`
	Misses         Counter                     `json:"misses"`
	BadCuts        Counter                     `json:"badCuts"`
	BombHits       Counter                     `json:"bombHits"`
	MaxCombo       Counter                     `json:"maxCombo"`
}

type Stats struct {
	Left  HandStat `json:"left"`
	Right HandStat `json:"right"`
	Total HandStat `json:"total"`
}

type ReplayStats struct {
	Info  ReplayStatsInfo `json:"info"`
	Stats Stats           `json:"stats"`
}

type StatBuffer struct {
	AccCut         CutBuffer
	BeforeCut      CutBuffer
	AfterCut       CutBuffer
	Score          CutBuffer
	TimeDependence SwingBuffer
	PreSwing       SwingBuffer
	PostSwing      SwingBuffer
	Grid           []CutBuffer
	Notes          Counter
	Misses         Counter
	BadCuts        Counter
	BombHits       Counter
}

func (buf *StatBuffer) add(goodNoteCut *GoodNoteCutEvent) {
	score := goodNoteCut.AccCut + goodNoteCut.BeforeCut + goodNoteCut.AfterCut

	if goodNoteCut.EventType == Good {
		if goodNoteCut.ScoringType != SliderTail && goodNoteCut.ScoringType != BurstSliderElement {
			buf.BeforeCut.Add(goodNoteCut.BeforeCut)
			buf.PreSwing.Add(goodNoteCut.BeforeCutRating)
		}

		if goodNoteCut.ScoringType != BurstSliderHead && goodNoteCut.ScoringType != BurstSliderElement {
			buf.AccCut.Add(goodNoteCut.AccCut)
			buf.Score.Add(score)
			buf.TimeDependence.Add(goodNoteCut.TimeDependence)
		}

		if goodNoteCut.ScoringType != SliderHead && goodNoteCut.ScoringType != BurstSliderHead && goodNoteCut.ScoringType != BurstSliderElement {
			buf.AfterCut.Add(goodNoteCut.AfterCut)
			buf.PostSwing.Add(goodNoteCut.AfterCutRating)
		}

		if goodNoteCut.ScoringType != BurstSliderHead && goodNoteCut.ScoringType != BurstSliderElement {
			index := NewBlockPosition(goodNoteCut.LineLayer, goodNoteCut.LineIdx)
			buf.Grid[index].Add(score)
		}
	}
}

func (buf *StatBuffer) stat() *HandStat {
	return &HandStat{
		AccCut:         buf.AccCut.Stats(),
		BeforeCut:      buf.BeforeCut.Stats(),
		AfterCut:       buf.AfterCut.Stats(),
		Score:          buf.Score.Stats(),
		TimeDependence: buf.TimeDependence.Stats(),
		PreSwing:       buf.PreSwing.Stats(),
		PostSwing:      buf.PostSwing.Stats(),
		Grid: buffer.StatsSlice[CutValue]{
			Min:    utils.SliceMap[CutBuffer, CutValue](buf.Grid, func(buf CutBuffer) CutValue { return buf.Min() }),
			Avg:    utils.SliceMap[CutBuffer, SwingValue](buf.Grid, func(buf CutBuffer) SwingValue { return buf.Avg() }),
			Median: utils.SliceMap[CutBuffer, CutValue](buf.Grid, func(buf CutBuffer) CutValue { return buf.Median() }),
			Max:    utils.SliceMap[CutBuffer, CutValue](buf.Grid, func(buf CutBuffer) CutValue { return buf.Max() }),
		},
		Notes:    buf.Notes,
		Misses:   buf.Misses,
		BadCuts:  buf.BadCuts,
		BombHits: buf.BombHits,
	}
}

func newStatBuffer(length int) *StatBuffer {
	return &StatBuffer{
		AccCut:         buffer.NewBuffer[CutValue, CutValueSum](length),
		BeforeCut:      buffer.NewBuffer[CutValue, CutValueSum](length),
		AfterCut:       buffer.NewBuffer[CutValue, CutValueSum](length),
		Score:          buffer.NewBuffer[CutValue, CutValueSum](length),
		TimeDependence: buffer.NewBuffer[SwingValue, SwingValueSum](length),
		PreSwing:       buffer.NewBuffer[SwingValue, SwingValueSum](length),
		PostSwing:      buffer.NewBuffer[SwingValue, SwingValueSum](length),
		Grid:           buffer.NewBufferSlice[CutValue, CutValueSum](GridBufferSize, length),
	}
}

func newStatInfo(info *ReplayEventsInfo) *ReplayStatsInfo {
	return &ReplayStatsInfo{
		Info:       info.Info,
		EndTime:    info.EndTime,
		Accuracy:   info.Accuracy,
		FcAccuracy: info.FcAccuracy,
		CalcScore:  info.CalcScore,
		WallHits:   0,
		Pauses:     0,
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

	replayStats.Stats.Left.MaxCombo = replay.Info.MaxLeftCombo
	replayStats.Stats.Right.MaxCombo = replay.Info.MaxRightCombo
	replayStats.Stats.Total.MaxCombo = replay.Info.MaxCombo

	replayStats.Info.Pauses = len(replay.Pauses)
	replayStats.Info.WallHits = len(replay.Walls)

	return replayStats
}
