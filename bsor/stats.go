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

func (buf *StatBuffer) add(goodNoteCut *GoodCutEvent) {
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
		MinAccCut:            utils.SliceMin[uint16](buf.AccCut.Values),
		AvgAccCut:            buf.AccCut.Avg(),
		MedianAccCut:         buf.AccCut.Median(),
		MaxAccCut:            utils.SliceMax[uint16](buf.AccCut.Values),
		MinBeforeCut:         utils.SliceMin[uint16](buf.BeforeCut.Values),
		AvgBeforeCut:         buf.BeforeCut.Avg(),
		MedianBeforeCut:      buf.BeforeCut.Median(),
		MaxBeforeCut:         utils.SliceMax[uint16](buf.BeforeCut.Values),
		MinAfterCut:          utils.SliceMin[uint16](buf.AfterCut.Values),
		AvgAfterCut:          buf.AfterCut.Avg(),
		MedianAfterCut:       buf.AfterCut.Median(),
		MaxAfterCut:          utils.SliceMax[uint16](buf.AfterCut.Values),
		MinScore:             utils.SliceMin[uint16](buf.Score.Values),
		AvgScore:             buf.Score.Avg(),
		MedianScore:          buf.Score.Median(),
		MaxScore:             utils.SliceMax[uint16](buf.Score.Values),
		MinTimeDependence:    utils.SliceMin[float64](buf.TimeDependence.Values),
		AvgTimeDependence:    buf.TimeDependence.Avg(),
		MedianTimeDependence: buf.TimeDependence.Median(),
		MaxTimeDependence:    utils.SliceMax[float64](buf.TimeDependence.Values),
		MinPreSwing:          utils.SliceMin[float64](buf.PreSwing.Values),
		AvgPreSwing:          buf.PreSwing.Avg() * 100,
		MedianPreSwing:       buf.PreSwing.Median() * 100,
		MaxPreswing:          utils.SliceMax[float64](buf.PreSwing.Values) * 100,
		MinPostSwing:         utils.SliceMin[float64](buf.PostSwing.Values),
		AvgPostSwing:         buf.PostSwing.Avg() * 100,
		MedianPostSwing:      buf.PostSwing.Median() * 100,
		MaxPostSwing:         utils.SliceMax[float64](buf.PostSwing.Values) * 100,
		MinGrid:              utils.SliceMap[buffer.Buffer[uint16, int64], uint16](buf.Grid, func(buf buffer.Buffer[uint16, int64]) uint16 { return utils.SliceMin[uint16](buf.Values) }),
		AvgGrid:              utils.SliceMap[buffer.Buffer[uint16, int64], float64](buf.Grid, func(buf buffer.Buffer[uint16, int64]) float64 { return buf.Avg() }),
		MedianGrid:           utils.SliceMap[buffer.Buffer[uint16, int64], uint16](buf.Grid, func(buf buffer.Buffer[uint16, int64]) uint16 { return buf.Median() }),
		MaxGrid:              utils.SliceMap[buffer.Buffer[uint16, int64], uint16](buf.Grid, func(buf buffer.Buffer[uint16, int64]) uint16 { return utils.SliceMax[uint16](buf.Values) }),
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
