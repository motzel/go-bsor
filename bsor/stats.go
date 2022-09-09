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
	MinAccCut            uint16    `json:"minAccCut"`
	AvgAccCut            float64   `json:"avgAccCut"`
	MedianAccCut         uint16    `json:"medAccCut"`
	MaxAccCut            uint16    `json:"maxAccCut"`
	MinBeforeCut         uint16    `json:"minBeforeCut"`
	AvgBeforeCut         float64   `json:"avgBeforeCut"`
	MedianBeforeCut      uint16    `json:"medBeforeCut"`
	MaxBeforeCut         uint16    `json:"maxBeforeCut"`
	MinAfterCut          uint16    `json:"minAfterCut"`
	AvgAfterCut          float64   `json:"avgAfterCut"`
	MedianAfterCut       uint16    `json:"medAfterCut"`
	MaxAfterCut          uint16    `json:"maxAfterCut"`
	MinScore             uint16    `json:"minScore"`
	AvgScore             float64   `json:"avgScore"`
	MedianScore          uint16    `json:"medScore"`
	MaxScore             uint16    `json:"maxScore"`
	AvgScorePercent      float64   `json:"avgScorePercent"`
	MedianScorePercent   float64   `json:"medScorePercent"`
	MinTimeDependence    float64   `json:"minTimeDependence"`
	AvgTimeDependence    float64   `json:"avgTimeDependence"`
	MedianTimeDependence float64   `json:"medTimeDependence"`
	MaxTimeDependence    float64   `json:"maxTimeDependence"`
	MinPreSwing          float64   `json:"minPreSwing"`
	AvgPreSwing          float64   `json:"avgPreSwing"`
	MedianPreSwing       float64   `json:"medPreSwing"`
	MaxPreswing          float64   `json:"maxPreswing"`
	MinPostSwing         float64   `json:"minPostSwing"`
	AvgPostSwing         float64   `json:"avgPostSwing"`
	MedianPostSwing      float64   `json:"medPostSwing"`
	MaxPostSwing         float64   `json:"maxPostSwing"`
	MinGrid              []uint16  `json:"minGrid"`
	AvgGrid              []float64 `json:"avgGrid"`
	MedianGrid           []uint16  `json:"medGrid"`
	MaxGrid              []uint16  `json:"maxGrid"`
	Notes                int       `json:"notes"`
	Misses               int       `json:"misses"`
	BadCuts              int       `json:"badCuts"`
	BombHits             int       `json:"bombHit"`
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
	stat.MedianScorePercent = float64(stat.MedianScore) / BlockMaxValue * 100
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
		MinAccCut:            buf.AccCut.Min(),
		AvgAccCut:            buf.AccCut.Avg(),
		MedianAccCut:         buf.AccCut.Median(),
		MaxAccCut:            buf.AccCut.Max(),
		MinBeforeCut:         buf.BeforeCut.Min(),
		AvgBeforeCut:         buf.BeforeCut.Avg(),
		MedianBeforeCut:      buf.BeforeCut.Median(),
		MaxBeforeCut:         buf.BeforeCut.Max(),
		MinAfterCut:          buf.AfterCut.Min(),
		AvgAfterCut:          buf.AfterCut.Avg(),
		MedianAfterCut:       buf.AfterCut.Median(),
		MaxAfterCut:          buf.AfterCut.Max(),
		MinScore:             buf.Score.Min(),
		AvgScore:             buf.Score.Avg(),
		MedianScore:          buf.Score.Median(),
		MaxScore:             buf.Score.Max(),
		MinTimeDependence:    buf.TimeDependence.Min(),
		AvgTimeDependence:    buf.TimeDependence.Avg(),
		MedianTimeDependence: buf.TimeDependence.Median(),
		MaxTimeDependence:    buf.TimeDependence.Max(),
		MinPreSwing:          buf.PreSwing.Min() * 100,
		AvgPreSwing:          buf.PreSwing.Avg() * 100,
		MedianPreSwing:       buf.PreSwing.Median() * 100,
		MaxPreswing:          buf.PreSwing.Max() * 100,
		MinPostSwing:         buf.PostSwing.Min() * 100,
		AvgPostSwing:         buf.PostSwing.Avg() * 100,
		MedianPostSwing:      buf.PostSwing.Median() * 100,
		MaxPostSwing:         buf.PostSwing.Max() * 100,
		MinGrid:              utils.SliceMap[buffer.Buffer[uint16, int64], uint16](buf.Grid, func(buf buffer.Buffer[uint16, int64]) uint16 { return buf.Min() }),
		AvgGrid:              utils.SliceMap[buffer.Buffer[uint16, int64], float64](buf.Grid, func(buf buffer.Buffer[uint16, int64]) float64 { return buf.Avg() }),
		MedianGrid:           utils.SliceMap[buffer.Buffer[uint16, int64], uint16](buf.Grid, func(buf buffer.Buffer[uint16, int64]) uint16 { return buf.Median() }),
		MaxGrid:              utils.SliceMap[buffer.Buffer[uint16, int64], uint16](buf.Grid, func(buf buffer.Buffer[uint16, int64]) uint16 { return buf.Max() }),
		Notes:                buf.Notes,
		Misses:               buf.Misses,
		BadCuts:              buf.BadCuts,
		BombHits:             buf.BombHits,
	}

	stat.calcPercentAcc()

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
