package bsor

import (
	"github.com/motzel/go-bsor/bsor/buffer"
	"math"
	"sort"
)

const fcBufferSize = 10

type NoteRating struct {
	CutDistanceToCenter SwingValue `json:"cutDistanceToCenter"`
	BeforeCutRating     SwingValue `json:"beforeCutRating"`
	AfterCutRating      SwingValue `json:"afterCutRating"`
}

type NoteScore struct {
	BeforeCut CutValue `json:"beforeCut"`
	AfterCut  CutValue `json:"afterCut"`
	AccCut    CutValue `json:"accCut"`
}

func getNoteScore(eventType NoteEventType, scoringType NoteScoringType, cutInfo NoteRating) NoteScore {
	score := NoteScore{}

	if eventType == Good {
		score.BeforeCut = 0
		if scoringType == SliderTail {
			score.BeforeCut = 70
		} else if scoringType != BurstSliderElement {
			score.BeforeCut = CutValue(math.Round(clamp(float64(cutInfo.BeforeCutRating*70), 0, 70)))
		}

		score.AfterCut = 0
		if scoringType == SliderHead {
			score.AfterCut = 30
		} else if scoringType != BurstSliderElement && scoringType != BurstSliderHead {
			score.AfterCut = CutValue(math.Round(clamp(float64(cutInfo.AfterCutRating*30), 0, 30)))
		}

		score.AccCut = 0
		if scoringType == BurstSliderElement {
			score.AccCut = 20
		} else {
			score.AccCut = CutValue(math.Round(15 * (1 - clamp(float64(cutInfo.CutDistanceToCenter/0.3), 0, 1))))
		}
	}

	return score
}

type GameEventI interface {
	GetIdx() Counter
	GetTime() TimeValue
	GetColor() ColorType
	GetScore() CutValue
	SetPredictedScore(score CutValue)
	GetMaxScore() CutValue
	DecreasesCombo() bool
	IsNote() bool
	GetAccuracy() SwingValue
	SetAccuracy(acc SwingValue)
	GetFcAccuracy() SwingValue
	SetFcAccuracy(acc SwingValue)
	SetMultiplier(multiplier Counter)
	IsTheSameEvent(event GameEvent) bool
}

type GameEvent struct {
	EventIdx     Counter         `json:"idx"`
	EventType    NoteEventType   `json:"eventType"`
	ScoringType  NoteScoringType `json:"scoringType"`
	LineIdx      LineValue       `json:"lineIdx"`
	LineLayer    LayerValue      `json:"lineLayer"`
	ColorType    ColorType       `json:"colorType"`
	CutDirection CutDirection    `json:"cutDirection"`
	EventTime    TimeValue       `json:"eventTime"`
	Accuracy     SwingValue      `json:"accuracy"`
	FcAccuracy   SwingValue      `json:"fcAccuracy"`
	Multiplier   Counter         `json:"multiplier"`
	GameEventI   `json:"-"`
}

func (gameEvent *GameEvent) IsTheSameEvent(second GameEvent) bool {
	return gameEvent.EventType == second.EventType &&
		gameEvent.ScoringType == second.ScoringType &&
		gameEvent.LineIdx == second.LineIdx &&
		gameEvent.LineLayer == second.LineLayer &&
		gameEvent.ColorType == second.ColorType &&
		gameEvent.CutDirection == second.CutDirection &&
		gameEvent.EventTime == second.EventTime

}

func (gameEvent *GameEvent) GetIdx() Counter {
	return gameEvent.EventIdx
}

func (gameEvent *GameEvent) GetTime() TimeValue {
	return gameEvent.EventTime
}

func (gameEvent *GameEvent) GetColor() ColorType {
	return gameEvent.ColorType
}

func (gameEvent *GameEvent) GetScore() CutValue {
	return 0
}

func (gameEvent *GameEvent) GetMaxScore() CutValue {
	switch gameEvent.ScoringType {
	case BurstSliderHead:
		return 85
	case BurstSliderElement:
		return 20
	default:
		return BlockMaxValue
	}
}

func (gameEvent *GameEvent) DecreasesCombo() bool {
	return gameEvent.EventType != Good
}

func (gameEvent *GameEvent) IsNote() bool {
	return false
}

func (gameEvent *GameEvent) GetAccuracy() SwingValue {
	return gameEvent.Accuracy
}

func (gameEvent *GameEvent) SetAccuracy(acc SwingValue) {
	gameEvent.Accuracy = acc
}

func (gameEvent *GameEvent) GetFcAccuracy() SwingValue {
	return gameEvent.FcAccuracy
}

func (gameEvent *GameEvent) SetFcAccuracy(acc SwingValue) {
	gameEvent.FcAccuracy = acc
}

func (gameEvent *GameEvent) SetPredictedScore(score CutValue) {}

func (gameEvent *GameEvent) SetMultiplier(multiplier Counter) {
	gameEvent.Multiplier = multiplier
}

type GoodNoteCutEvent struct {
	GameEvent
	PredictedScore CutValue   `json:"predictedScore"`
	TimeDependence SwingValue `json:"timeDependence"`
	NoteRating
	NoteScore
}

func (note *GoodNoteCutEvent) GetScore() CutValue {
	if note.EventType == Good {
		score := getNoteScore(
			note.EventType,
			note.ScoringType,
			NoteRating{
				BeforeCutRating:     note.BeforeCutRating,
				AfterCutRating:      note.AfterCutRating,
				CutDistanceToCenter: note.CutDistanceToCenter,
			},
		)

		return score.BeforeCut + score.AccCut + score.AfterCut
	}

	return 0
}

func (note *GoodNoteCutEvent) IsNote() bool {
	return true
}

func (note *GoodNoteCutEvent) SetPredictedScore(score CutValue) {
	note.PredictedScore = score
}

type MissedNoteEvent struct {
	GameEvent
	PredictedScore CutValue `json:"predictedScore"`
}

func (note *MissedNoteEvent) IsNote() bool {
	return true
}

func (note *MissedNoteEvent) SetPredictedScore(score CutValue) {
	note.PredictedScore = score
}

type BadCutEvent struct {
	GameEvent
	PredictedScore CutValue   `json:"predictedScore"`
	TimeDependence SwingValue `json:"timeDependence"`
}

func (note *BadCutEvent) IsNote() bool {
	return true
}

func (note *BadCutEvent) SetPredictedScore(score CutValue) {
	note.PredictedScore = score
}

type BombHitEvent struct {
	GameEvent
}

type WallHitEvent struct {
	EventIdx   Counter    `json:"idx"`
	Accuracy   SwingValue `json:"accuracy"`
	FcAccuracy SwingValue `json:"fcAccuracy"`
	Multiplier Counter    `json:"multiplier"`
	WallHit
	GameEventI `json:"-"`
}

func (wallHit *WallHitEvent) GetIdx() Counter {
	return wallHit.EventIdx
}

func (wallHit *WallHitEvent) IsNote() bool {
	return false
}

func (wallHit *WallHitEvent) GetTime() TimeValue {
	return wallHit.Time
}

func (wallHit *WallHitEvent) GetMaxScore() CutValue {
	return 0
}

func (wallHit *WallHitEvent) GetScore() CutValue {
	return 0
}

func (wallHit *WallHitEvent) DecreasesCombo() bool {
	return true
}

func (wallHit *WallHitEvent) GetColor() ColorType {
	return NoColor
}

func (wallHit *WallHitEvent) GetAccuracy() SwingValue {
	return wallHit.Accuracy
}

func (wallHit *WallHitEvent) SetAccuracy(acc SwingValue) {
	wallHit.Accuracy = acc
}

func (wallHit *WallHitEvent) GetFcAccuracy() SwingValue {
	return wallHit.FcAccuracy
}

func (wallHit *WallHitEvent) SetFcAccuracy(acc SwingValue) {
	wallHit.FcAccuracy = acc
}

func (gameEvent *WallHitEvent) SetMultiplier(multiplier Counter) {
	gameEvent.Multiplier = multiplier
}

type ReplayEventsInfo struct {
	Info
	EndTime       TimeValue  `json:"endTime"`
	CalcScore     Score      `json:"calcScore"`
	Accuracy      SwingValue `json:"accuracy"`
	CalcAccuracy  SwingValue `json:"calcAccuracy"`
	FcAccuracy    SwingValue `json:"fcAccuracy"`
	MaxCombo      Counter    `json:"maxCombo"`
	MaxLeftCombo  Counter    `json:"maxLeftCombo"`
	MaxRightCombo Counter    `json:"maxRightCombo"`
}

type ReplayEvents struct {
	Info     ReplayEventsInfo   `json:"info"`
	Hits     []GoodNoteCutEvent `json:"notes"`
	Misses   []MissedNoteEvent  `json:"misses"`
	BadCuts  []BadCutEvent      `json:"badCuts"`
	BombHits []BombHitEvent     `json:"bombHits"`
	Walls    []WallHitEvent     `json:"walls"`
	Pauses   []Pause            `json:"pauses"`
}

type ReplayEventsWithStats struct {
	ReplayEvents
	Stats Stats `json:"stats"`
}

func calculateStats(events *ReplayEvents, gameEvents []GameEventI) {
	multiplier := NewMultiplierCounter()
	maxMultiplier := NewMultiplierCounter()

	sort.Slice(gameEvents, func(i, j int) bool {
		if gameEvents[i].GetTime() == gameEvents[j].GetTime() {
			return gameEvents[i].GetIdx() <= gameEvents[j].GetIdx()
		}

		return gameEvents[i].GetTime() < gameEvents[j].GetTime()
	})

	leftFcBuffer := buffer.NewCircularBuffer[CutValue, CutValueSum](fcBufferSize)
	rightFcBuffer := buffer.NewCircularBuffer[CutValue, CutValueSum](fcBufferSize)

	var score, fcScore, maxScore Score
	var maxCombo, maxLeftCombo, maxRightCombo Counter
	var currentCombo, currentLeftCombo, currentRightCombo Counter
	for i, gameEvent := range gameEvents {
		gameEventScore := Score(gameEvent.GetScore())
		score += gameEventScore * Score(multiplier.Value())
		maxScore += Score(gameEvent.GetMaxScore()) * Score(maxMultiplier.Value())

		isLeft := gameEvent.GetColor() == Red

		gameEvent.SetMultiplier(Counter(multiplier.Value()))

		if gameEvent.IsNote() {
			predictedScore := CutValue(0)

			if gameEventScore > 0 {
				if isLeft {
					leftFcBuffer.Add(CutValue(gameEventScore))
				} else {
					rightFcBuffer.Add(CutValue(gameEventScore))
				}

				fcScore += gameEventScore * Score(maxMultiplier.Value())

				predictedScore = CutValue(gameEventScore)
			} else if isLeft && leftFcBuffer.Size() > 0 {
				predictedScore = leftFcBuffer.Median()
				fcScore += Score(predictedScore * CutValue(maxMultiplier.Value()))
			} else if !isLeft && rightFcBuffer.Size() > 0 {
				predictedScore = rightFcBuffer.Median()
				fcScore += Score(predictedScore * CutValue(maxMultiplier.Value()))
			} else {
				predictedScore = CutValue(BlockMaxValue)
				fcScore += Score(predictedScore * CutValue(maxMultiplier.Value()))
			}

			gameEvent.SetPredictedScore(predictedScore)
		}

		if maxScore != 0 {
			gameEvents[i].SetAccuracy(SwingValue(score) / SwingValue(maxScore) * 100)
			gameEvents[i].SetFcAccuracy(SwingValue(fcScore) / SwingValue(maxScore) * 100)
		}

		maxMultiplier.Inc()

		if gameEvent.DecreasesCombo() {
			multiplier.Dec()

			if gameEvent.GetColor() != NoColor {
				if isLeft {
					if currentLeftCombo > maxLeftCombo {
						maxLeftCombo = currentLeftCombo
					}

					currentLeftCombo = 0
				} else {
					if currentRightCombo > maxRightCombo {
						maxRightCombo = currentRightCombo
					}

					currentRightCombo = 0
				}
			}

			if currentCombo > maxCombo {
				maxCombo = currentCombo
			}

			currentCombo = 0
		} else {
			multiplier.Inc()

			currentCombo++

			if gameEvent.GetColor() != NoColor {
				if isLeft {
					currentLeftCombo++
				} else {
					currentRightCombo++
				}
			}
		}
	}

	if currentLeftCombo > maxLeftCombo {
		maxLeftCombo = currentLeftCombo
	}

	if currentRightCombo > maxRightCombo {
		maxRightCombo = currentRightCombo
	}

	if currentCombo > maxCombo {
		maxCombo = currentCombo
	}

	events.Info.CalcScore = score
	events.Info.MaxCombo = maxCombo
	events.Info.MaxLeftCombo = maxLeftCombo
	events.Info.MaxRightCombo = maxRightCombo

	if maxScore > 0 {
		if score > 0 {
			events.Info.CalcAccuracy = SwingValue(score) / SwingValue(maxScore) * 100
		} else {
			events.Info.CalcAccuracy = SwingValue(events.Info.Score) / SwingValue(maxScore) * 100
		}
		events.Info.FcAccuracy = SwingValue(fcScore) / SwingValue(maxScore) * 100
		events.Info.Accuracy = SwingValue(events.Info.Score) / SwingValue(maxScore) * 100
	}
}

func createReplayEvents(replay *Replay, fixReplayErrors bool) *ReplayEvents {
	hitsCnt := 0
	missesCnt := 0
	badCutsCnt := 0
	bombHitsCnt := 0
	for i := range replay.Notes {
		switch replay.Notes[i].EventType {
		case Good:
			hitsCnt++
		case Bad:
			badCutsCnt++
		case Miss:
			missesCnt++
		case Bomb:
			bombHitsCnt++
		}
	}

	events := &ReplayEvents{
		Info:     ReplayEventsInfo{Info: replay.Info, EndTime: 0},
		Hits:     make([]GoodNoteCutEvent, 0, hitsCnt),
		Misses:   make([]MissedNoteEvent, 0, missesCnt),
		BadCuts:  make([]BadCutEvent, 0, badCutsCnt),
		BombHits: make([]BombHitEvent, 0, bombHitsCnt),
		Walls:    make([]WallHitEvent, 0, len(replay.Walls)),
		Pauses:   replay.Pauses,
	}

	gameEvents := make([]GameEventI, 0, len(replay.Notes)+len(replay.Walls))

	for i := range replay.Notes {
		note := replay.Notes[i]

		timeDependence := SwingValue(math.Abs(float64(note.CutInfo.CutNormal.Z)))

		gameEvent := GameEvent{
			EventIdx:     Counter(i),
			EventType:    note.EventType,
			ScoringType:  note.ScoringType,
			LineIdx:      note.LineIdx,
			LineLayer:    note.LineLayer,
			ColorType:    note.ColorType,
			CutDirection: note.CutDirection,
			EventTime:    note.EventTime,
			Multiplier:   1,
		}

		switch note.EventType {
		case Good:
			noteEvent := GoodNoteCutEvent{
				GameEvent:      gameEvent,
				PredictedScore: 0,
				TimeDependence: timeDependence,
				NoteRating:     NoteRating{BeforeCutRating: SwingValue(note.CutInfo.BeforeCutRating), AfterCutRating: SwingValue(note.CutInfo.AfterCutRating), CutDistanceToCenter: SwingValue(note.CutInfo.CutDistanceToCenter)},
				NoteScore: getNoteScore(note.EventType, note.ScoringType, NoteRating{
					BeforeCutRating:     SwingValue(note.CutInfo.BeforeCutRating),
					AfterCutRating:      SwingValue(note.CutInfo.AfterCutRating),
					CutDistanceToCenter: SwingValue(note.CutInfo.CutDistanceToCenter),
				}),
			}

			if !fixReplayErrors || len(events.Hits) == 0 || (fixReplayErrors && !noteEvent.IsTheSameEvent(events.Hits[len(events.Hits)-1].GameEvent)) {
				events.Hits = append(events.Hits, noteEvent)
				gameEvents = append(gameEvents, &(events.Hits[len(events.Hits)-1]))
			}

		case Bad:
			badCut := BadCutEvent{GameEvent: gameEvent, PredictedScore: 0, TimeDependence: timeDependence}

			if !fixReplayErrors || len(events.BadCuts) == 0 || (fixReplayErrors && !badCut.IsTheSameEvent(events.BadCuts[len(events.BadCuts)-1].GameEvent)) {
				events.BadCuts = append(events.BadCuts, badCut)
				gameEvents = append(gameEvents, &(events.BadCuts[len(events.BadCuts)-1]))
			}

		case Miss:
			missedNote := MissedNoteEvent{GameEvent: gameEvent, PredictedScore: 0}

			if !fixReplayErrors || len(events.Misses) == 0 || (fixReplayErrors && !missedNote.IsTheSameEvent(events.Misses[len(events.Misses)-1].GameEvent)) {
				events.Misses = append(events.Misses, missedNote)
				gameEvents = append(gameEvents, &(events.Misses[len(events.Misses)-1]))
			}
		case Bomb:
			bombHit := BombHitEvent{GameEvent: gameEvent}

			if !fixReplayErrors || len(events.BombHits) == 0 || (fixReplayErrors && !bombHit.IsTheSameEvent(events.BombHits[len(events.BombHits)-1].GameEvent)) {
				events.BombHits = append(events.BombHits, bombHit)
				gameEvents = append(gameEvents, &(events.BombHits[len(events.BombHits)-1]))
			}
		}
	}

	numOfEvents := len(events.Hits) + len(events.Misses) + len(events.BadCuts) + len(events.BombHits)

	for i := range replay.Walls {
		wallHitEvent := WallHitEvent{EventIdx: Counter(i) + Counter(numOfEvents), WallHit: replay.Walls[i]}
		events.Walls = append(events.Walls, wallHitEvent)
		gameEvents = append(gameEvents, &(events.Walls[len(events.Walls)-1]))
	}

	if len(replay.Frames) > 0 {
		events.Info.EndTime = replay.Frames[len(replay.Frames)-1].Time
	}

	calculateStats(events, gameEvents)

	return events
}

func NewReplayEvents(replay *Replay) *ReplayEvents {
	events := createReplayEvents(replay, false)

	if events.Info.Score != events.Info.CalcScore {
		events = createReplayEvents(replay, true)
	}

	return events
}

func NewReplayEventsWithStats(replayEvents *ReplayEvents) *ReplayEventsWithStats {
	replayStats := NewReplayStats(replayEvents)

	return &ReplayEventsWithStats{
		ReplayEvents: *replayEvents,
		Stats:        replayStats.Stats,
	}
}
