package bsor

import (
	"math"
)

type NoteEvent struct {
	EventType    NoteEventType   `json:"eventType"`
	ScoringType  NoteScoringType `json:"scoringType"`
	LineIdx      byte            `json:"lineIdx"`
	LineLayer    byte            `json:"lineLayer"`
	ColorType    ColorType       `json:"colorType"`
	CutDirection CutDirection    `json:"cutDirection"`
	EventTime    float32         `json:"eventTime"`
}

type GoodCutEvent struct {
	NoteEvent
	TimeDependence  float32 `json:"timeDependence"`
	BeforeCutRating float32 `json:"beforeCutRating"`
	AfterCutRating  float32 `json:"afterCutRating"`
	BeforeCut       byte    `json:"beforeCut"`
	AfterCut        byte    `json:"afterCut"`
	AccCut          byte    `json:"accCut"`
}

type MissedNoteEvent struct {
	NoteEvent
}

type BadCutEvent struct {
	NoteEvent
	TimeDependence float32 `json:"timeDependence"`
}

type BombHitEvent struct {
	NoteEvent
}

type ReplayEvents struct {
	Info     Info              `json:"info"`
	Hits     []GoodCutEvent    `json:"notes"`
	Misses   []MissedNoteEvent `json:"misses"`
	BadCuts  []BadCutEvent     `json:"badCuts"`
	BombHits []BombHitEvent    `json:"bombHits"`
	Walls    []WallHit         `json:"walls"`
	Pauses   []Pause           `json:"pauses"`
}

type ReplayEventsWithStats struct {
	ReplayEvents
	Stats Stats `json:"stats"`
}

func NewReplayEvents(replay *Replay) *ReplayEvents {
	events := &ReplayEvents{
		Info:     replay.Info,
		Hits:     make([]GoodCutEvent, 0, len(replay.Notes)),
		Misses:   make([]MissedNoteEvent, 0),
		BadCuts:  make([]BadCutEvent, 0),
		BombHits: make([]BombHitEvent, 0),
		Walls:    replay.Walls,
		Pauses:   replay.Pauses,
	}

	for i := range replay.Notes {
		note := replay.Notes[i]

		timeDependence := float32(math.Abs(float64(note.CutInfo.CutNormal.Z)))

		noteEvent := NoteEvent{
			EventType:    note.EventType,
			ScoringType:  note.ScoringType,
			LineIdx:      note.LineIdx,
			LineLayer:    note.LineLayer,
			ColorType:    note.ColorType,
			CutDirection: note.CutDirection,
			EventTime:    note.EventTime,
		}

		noteSimple := GoodCutEvent{
			NoteEvent:       noteEvent,
			TimeDependence:  timeDependence,
			BeforeCutRating: note.CutInfo.BeforeCutRating,
			AfterCutRating:  note.CutInfo.AfterCutRating,
		}

		if note.EventType == Good {
			noteSimple.BeforeCut = 0
			if note.ScoringType == SliderTail {
				noteSimple.BeforeCut = 70
			} else if note.ScoringType != BurstSliderElement {
				noteSimple.BeforeCut = byte(math.Round(clamp(float64(note.CutInfo.BeforeCutRating*70), 0, 70)))
			}

			noteSimple.AfterCut = 0
			if note.ScoringType == SliderHead {
				noteSimple.AfterCut = 30
			} else if note.ScoringType != BurstSliderElement && note.ScoringType != BurstSliderHead {
				noteSimple.AfterCut = byte(math.Round(clamp(float64(note.CutInfo.AfterCutRating*30), 0, 30)))
			}

			noteSimple.AccCut = 0
			if note.ScoringType == BurstSliderElement {
				noteSimple.AccCut = 20
			} else {
				noteSimple.AccCut = byte(math.Round(15 * (1 - clamp(float64(note.CutInfo.CutDistanceToCenter/0.3), 0, 1))))
			}
		}

		switch note.EventType {
		case Good:
			events.Hits = append(events.Hits, noteSimple)
		case Bad:
			badCut := BadCutEvent{NoteEvent: noteEvent, TimeDependence: timeDependence}
			events.BadCuts = append(events.BadCuts, badCut)
		case Miss:
			missedNote := MissedNoteEvent{NoteEvent: noteEvent}
			events.Misses = append(events.Misses, missedNote)
		case Bomb:
			bombHit := BombHitEvent{NoteEvent: noteEvent}
			events.BombHits = append(events.BombHits, bombHit)
		}
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
