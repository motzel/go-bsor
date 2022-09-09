package bsor

import (
	"github.com/motzel/go-bsor/bsor/buffer"
	"math"
	"sort"
)

const fcBufferSize = 5

type NoteRating struct {
	CutDistanceToCenter float32 `json:"cutDistanceToCenter"`
	BeforeCutRating     float32 `json:"beforeCutRating"`
	AfterCutRating      float32 `json:"afterCutRating"`
}

type NoteScore struct {
	BeforeCut byte `json:"beforeCut"`
	AfterCut  byte `json:"afterCut"`
	AccCut    byte `json:"accCut"`
}

func getNoteScore(eventType NoteEventType, scoringType NoteScoringType, cutInfo NoteRating) NoteScore {
	score := NoteScore{}

	if eventType == Good {
		score.BeforeCut = 0
		if scoringType == SliderTail {
			score.BeforeCut = 70
		} else if scoringType != BurstSliderElement {
			score.BeforeCut = byte(math.Round(clamp(float64(cutInfo.BeforeCutRating*70), 0, 70)))
		}

		score.AfterCut = 0
		if scoringType == SliderHead {
			score.AfterCut = 30
		} else if scoringType != BurstSliderElement && scoringType != BurstSliderHead {
			score.AfterCut = byte(math.Round(clamp(float64(cutInfo.AfterCutRating*30), 0, 30)))
		}

		score.AccCut = 0
		if scoringType == BurstSliderElement {
			score.AccCut = 20
		} else {
			score.AccCut = byte(math.Round(15 * (1 - clamp(float64(cutInfo.CutDistanceToCenter/0.3), 0, 1))))
		}
	}

	return score
}

type GameEventI interface {
	GetIdx() int32
	GetTime() float32
	GetColor() ColorType
	GetScore() byte
	GetMaxScore() byte
	DecreasesCombo() bool
	IsNote() bool
	GetAccuracy() float64
	SetAccuracy(acc float64)
	GetFcAccuracy() float64
	SetFcAccuracy(acc float64)
}

type GameEvent struct {
	EventIdx     int32           `json:"-"`
	EventType    NoteEventType   `json:"eventType"`
	ScoringType  NoteScoringType `json:"scoringType"`
	LineIdx      byte            `json:"lineIdx"`
	LineLayer    byte            `json:"lineLayer"`
	ColorType    ColorType       `json:"colorType"`
	CutDirection CutDirection    `json:"cutDirection"`
	EventTime    float32         `json:"eventTime"`
	Accuracy     float64         `json:"accuracy"`
	FcAccuracy   float64         `json:"fcAccuracy"`
	GameEventI   `json:"-"`
}

func (gameEvent *GameEvent) GetIdx() int32 {
	return gameEvent.EventIdx
}

func (gameEvent *GameEvent) GetTime() float32 {
	return gameEvent.EventTime
}

func (gameEvent *GameEvent) GetColor() ColorType {
	return gameEvent.ColorType
}

func (gameEvent *GameEvent) GetScore() byte {
	return 0
}

func (gameEvent *GameEvent) GetMaxScore() byte {
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

func (gameEvent *GameEvent) GetAccuracy() float64 {
	return gameEvent.Accuracy
}

func (gameEvent *GameEvent) SetAccuracy(acc float64) {
	gameEvent.Accuracy = acc
}

func (gameEvent *GameEvent) GetFcAccuracy() float64 {
	return gameEvent.FcAccuracy
}

func (gameEvent *GameEvent) SetFcAccuracy(acc float64) {
	gameEvent.FcAccuracy = acc
}

type GoodNoteCutEvent struct {
	GameEvent
	TimeDependence float32 `json:"timeDependence"`
	NoteRating
	NoteScore
}

func (note *GoodNoteCutEvent) GetScore() byte {
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

type MissedNoteEvent struct {
	GameEvent
}

func (note *MissedNoteEvent) IsNote() bool {
	return true
}

type BadCutEvent struct {
	GameEvent
	TimeDependence float32 `json:"timeDependence"`
}

func (note *BadCutEvent) IsNote() bool {
	return true
}

type BombHitEvent struct {
	GameEvent
}

type WallHitEvent struct {
	EventIdx   int32   `json:"-"`
	Accuracy   float64 `json:"accuracy"`
	FcAccuracy float64 `json:"fcAccuracy"`
	WallHit
	GameEventI `json:"-"`
}

func (wallHit *WallHitEvent) GetIdx() int32 {
	return wallHit.EventIdx
}

func (wallHit *WallHitEvent) IsNote() bool {
	return false
}

func (wallHit *WallHitEvent) GetTime() float32 {
	return wallHit.Time
}

func (wallHit *WallHitEvent) GetMaxScore() byte {
	return 0
}

func (wallHit *WallHitEvent) GetScore() byte {
	return 0
}

func (wallHit *WallHitEvent) DecreasesCombo() bool {
	return true
}

func (wallHit *WallHitEvent) GetColor() ColorType {
	return NoColor
}

func (wallHit *WallHitEvent) GetAccuracy() float64 {
	return wallHit.Accuracy
}

func (wallHit *WallHitEvent) SetAccuracy(acc float64) {
	wallHit.Accuracy = acc
}

func (wallHit *WallHitEvent) GetFcAccuracy() float64 {
	return wallHit.FcAccuracy
}

func (wallHit *WallHitEvent) SetFcAccuracy(acc float64) {
	wallHit.FcAccuracy = acc
}

type ReplayEventsInfo struct {
	Info
	EndTime       float32 `json:"endTime"`
	Accuracy      float64 `json:"accuracy"`
	FcAccuracy    float64 `json:"fcAccuracy"`
	CalcScore     int32   `json:"calcScore"`
	MaxCombo      int32   `json:"maxCombo"`
	MaxLeftCombo  int32   `json:"maxLeftCombo"`
	MaxRightCombo int32   `json:"maxRightCombo"`
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

	leftFcBuffer := buffer.NewCircularBuffer[byte, int64](fcBufferSize)
	rightFcBuffer := buffer.NewCircularBuffer[byte, int64](fcBufferSize)

	var score, fcScore, maxScore int32
	var maxCombo, maxLeftCombo, maxRightCombo int32
	var currentCombo, currentLeftCombo, currentRightCombo int32
	for i, gameEvent := range gameEvents {
		gameEventScore := int32(gameEvent.GetScore())
		score += gameEventScore * int32(multiplier.Value())
		maxScore += int32(gameEvent.GetMaxScore()) * int32(maxMultiplier.Value())

		isLeft := (events.Info.LeftHanded && gameEvent.GetColor() == Blue) || (!events.Info.LeftHanded && gameEvent.GetColor() == Red)

		if gameEvent.IsNote() {
			if gameEventScore > 0 {
				if isLeft {
					leftFcBuffer.Add(byte(gameEventScore))
				} else {
					rightFcBuffer.Add(byte(gameEventScore))
				}

				fcScore += gameEventScore * int32(maxMultiplier.Value())
			} else if isLeft && leftFcBuffer.Size() > 0 {
				fcScore += int32(math.Round(leftFcBuffer.Avg())) * int32(maxMultiplier.Value())
			} else if !isLeft && rightFcBuffer.Size() > 0 {
				fcScore += int32(math.Round(rightFcBuffer.Avg())) * int32(maxMultiplier.Value())
			} else {
				fcScore += BlockMaxValue * int32(maxMultiplier.Value())
			}
		}

		if maxScore > 0 {
			gameEvents[i].SetAccuracy(float64(score) / float64(maxScore) * 100)
			gameEvents[i].SetFcAccuracy(float64(fcScore) / float64(maxScore) * 100)
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
			events.Info.Accuracy = float64(score) / float64(maxScore) * 100
		} else {
			events.Info.Accuracy = float64(events.Info.Score) / float64(maxScore) * 100
		}
		events.Info.FcAccuracy = float64(fcScore) / float64(maxScore) * 100
	}
}

func NewReplayEvents(replay *Replay) *ReplayEvents {
	events := &ReplayEvents{
		Info:     ReplayEventsInfo{Info: replay.Info, EndTime: 0},
		Hits:     make([]GoodNoteCutEvent, 0, len(replay.Notes)),
		Misses:   make([]MissedNoteEvent, 0),
		BadCuts:  make([]BadCutEvent, 0),
		BombHits: make([]BombHitEvent, 0),
		Walls:    make([]WallHitEvent, 0, len(replay.Walls)),
		Pauses:   replay.Pauses,
	}

	gameEvents := make([]GameEventI, 0, len(replay.Notes)+len(replay.Walls))

	for i := range replay.Notes {
		note := replay.Notes[i]

		timeDependence := float32(math.Abs(float64(note.CutInfo.CutNormal.Z)))

		gameEvent := GameEvent{
			EventIdx:     int32(i),
			EventType:    note.EventType,
			ScoringType:  note.ScoringType,
			LineIdx:      note.LineIdx,
			LineLayer:    note.LineLayer,
			ColorType:    note.ColorType,
			CutDirection: note.CutDirection,
			EventTime:    note.EventTime,
		}

		switch note.EventType {
		case Good:
			noteEvent := GoodNoteCutEvent{
				GameEvent:      gameEvent,
				TimeDependence: timeDependence,
				NoteRating:     NoteRating{BeforeCutRating: note.CutInfo.BeforeCutRating, AfterCutRating: note.CutInfo.AfterCutRating, CutDistanceToCenter: note.CutInfo.CutDistanceToCenter},
				NoteScore: getNoteScore(note.EventType, note.ScoringType, NoteRating{
					BeforeCutRating:     note.CutInfo.BeforeCutRating,
					AfterCutRating:      note.CutInfo.AfterCutRating,
					CutDistanceToCenter: note.CutInfo.CutDistanceToCenter,
				}),
			}

			events.Hits = append(events.Hits, noteEvent)
			gameEvents = append(gameEvents, &events.Hits[len(events.Hits)-1])
		case Bad:
			badCut := BadCutEvent{GameEvent: gameEvent, TimeDependence: timeDependence}
			events.BadCuts = append(events.BadCuts, badCut)
			gameEvents = append(gameEvents, &events.BadCuts[len(events.BadCuts)-1])
		case Miss:
			missedNote := MissedNoteEvent{GameEvent: gameEvent}
			events.Misses = append(events.Misses, missedNote)
			gameEvents = append(gameEvents, &events.Misses[len(events.Misses)-1])
		case Bomb:
			bombHit := BombHitEvent{GameEvent: gameEvent}
			events.BombHits = append(events.BombHits, bombHit)
			gameEvents = append(gameEvents, &events.BombHits[len(events.BombHits)-1])
		}
	}

	numOfNotes := len(replay.Notes)

	for i := range replay.Walls {
		wallHitEvent := WallHitEvent{EventIdx: int32(i) + int32(numOfNotes), WallHit: replay.Walls[i]}
		events.Walls = append(events.Walls, wallHitEvent)
		gameEvents = append(gameEvents, &events.Walls[len(events.Walls)-1])
	}

	if len(replay.Frames) > 0 {
		events.Info.EndTime = replay.Frames[len(replay.Frames)-1].Time
	}

	calculateStats(events, gameEvents)

	return events
}

func NewReplayEventsWithStats(replayEvents *ReplayEvents) *ReplayEventsWithStats {
	replayStats := NewReplayStats(replayEvents)

	return &ReplayEventsWithStats{
		ReplayEvents: *replayEvents,
		Stats:        replayStats.Stats,
	}
}
