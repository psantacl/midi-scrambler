package average

import (
	"os"
	"fmt"
	"bytes"
	// "reflect"
	"gitlab.com/gomidi/midi/v2/smf"
	"com.github/psantacl/midi-scrambler/pkg/logging"
	"math/rand"
)

type TrackEventsPrime struct {
	smf.TrackEvent
	survived bool
}


func findNeighbors(windowSize uint64, ourEvents []TrackEventsPrime, targetEventIdx int) []uint8{
	logging.Sugar.Infow("findNeighbors", "windowSize", windowSize, "targetEventIdx",  targetEventIdx)
	var totalDelta uint64
	var neighbors []uint8
	subSlice := ourEvents[(targetEventIdx + 1):]
	for idx, ev := range subSlice  {
		totalDelta += uint64(ev.Delta)
		if totalDelta > windowSize {
			break;
		}
		var _ch, _key, _vel uint8
		ev := subSlice[idx]
		if !ev.survived {
			continue
		}
		if ev.Message.GetNoteOn(&_ch, &_key, &_vel) {
			// logging.Sugar.Infof("including neighbor %v", ev)
			// logging.Sugar.Sync()
			neighbors = append(neighbors, _key)
		}

	}

	return neighbors
}


func pickNeighor(neighbors []uint8) uint8 {
	count := len(neighbors)
	if count > 0 {
		return neighbors[rand.Intn(count)]
	}
	return 0
}

func handleAveraging(windowSize uint64, ourEvents []TrackEventsPrime) []TrackEventsPrime {
	logging.Sugar.Infow("handleAveraging", "windowSize", windowSize)
	for idx, ev := range ourEvents {
		var _ch, _key, _vel uint8
		if !ev.survived {
			continue
		}
		if ev.Message.GetNoteOn(&_ch, &_key, &_vel) {
			var neighbors = findNeighbors(windowSize, ourEvents, idx)
			neighbor := pickNeighor(neighbors)
			logging.Sugar.Infow("handleAveraging",
				"idx", idx,
				"ev", fmt.Sprintf("%v", ev),
				"neighbors", fmt.Sprintf("%v", neighbors),
				"neighbor", fmt.Sprintf("%v", neighbor))

		}
	}
	return ourEvents

}

func handleMonophonic(monophonic bool, tracksReader *smf.TracksReader) []TrackEventsPrime {
	ourEvents := []TrackEventsPrime{}
	currentNote := uint8(0)
	inNote  := false
	deltaDrop := uint32(0)


	tracksReader.Do(func(ev smf.TrackEvent) {
		survived := false

		// logging.Sugar.Infow("next event",
		// 	"track", ev.TrackNo,
		// 	"ms", ev.AbsMicroSeconds,
		// 	"ticks", ev.AbsTicks,
		// 	// "beat-clock ticks", ev.AbsTicks / beatClockRatio,
		// 	"delta", ev.Delta,
		// 	"message", ev.Message)

		logging.Sugar.Sync()

		var _ch, _key, _vel uint8
		switch {

		case !monophonic:
			survived = true

		case ev.Message.GetNoteOn(&_ch, &_key, &_vel):
			if (!inNote) {
				inNote = true
				ev.Delta = ev.Delta + deltaDrop
				deltaDrop = 0
				currentNote = _key
				survived = true
			} else { //already in a note
				// logging.Sugar.Infow("dropping NoteOn", "delta", ev.Delta, "key", _key)
				deltaDrop += ev.Delta
			}

		case ev.Message.GetNoteOff(&_ch, &_key, &_vel):
			if (inNote && currentNote == _key) {
				ev.Delta += deltaDrop
				deltaDrop = 0
				currentNote = 0
				inNote = false
				survived = true
			} else { //noteOf for a dropped note
				// logging.Sugar.Infow("dropping NoteOff", "delta", ev.Delta, "key", _key)
				deltaDrop += ev.Delta
			}
		default: //pass all Note events
			ev.Delta += deltaDrop
			deltaDrop = 0
			survived = true
		}
		ourEvents = append(ourEvents, TrackEventsPrime{TrackEvent: ev, survived: survived})

	})
	return ourEvents
}

func ProcessFile(inMidiFile string, outMidiFile string, monophonic bool, windowSize uint64) {
	logging.Sugar.Infow("Average",
		"in-file", inMidiFile,
		"out-file", outMidiFile,
		"monophonic", monophonic,
		"window-size", windowSize)

	data, err := os.ReadFile(inMidiFile)
	if err != nil {
		logging.Sugar.Errorf("unable to read midi file '%+v'", inMidiFile)
		panic(err)

	}

	bytesReader := bytes.NewReader(data)
	tracksReader := smf.ReadTracksFrom(bytesReader)
	ticks := tracksReader.SMF().TimeFormat.(smf.MetricTicks)
	// beatClockRatio := int64(ticks.Resolution()) / 24
	trackCount := len(tracksReader.SMF().Tracks)
	if trackCount != 1 {
		panic(fmt.Sprintf("Can only process Midi files with a single track for now!. Found %s tracks", trackCount))
	}
	logging.Sugar.Infof("ticks: %+v track_count: %+v", ticks.Resolution(), len(tracksReader.SMF().Tracks))


	ourEvents := handleMonophonic(monophonic,  tracksReader)
	ourEvents = handleAveraging(windowSize, ourEvents)
	var midiData = buildMidiOut(ourEvents, ticks.Resolution())

	err = os.WriteFile(outMidiFile, midiData, 0644)

	if err != nil {
		panic("Unable to write output midifile")
	}
}

func buildMidiOut(ourEvents []TrackEventsPrime, ticks uint16) []byte {
	var (
		bf    bytes.Buffer
		clock = smf.MetricTicks(ticks) // resolution: 96 ticks per quarternote 960 is also common
		tr    smf.Track
	)
	for _, ev := range ourEvents {
		if !ev.survived {
			//only include survivors in the file
			continue
		}
		tr.Add(ev.TrackEvent.Delta, ev.TrackEvent.Message)
	}
	tr.Close(0)

	s := smf.New()
	s.TimeFormat = clock
	s.Add(tr)
	s.WriteTo(&bf)
	return bf.Bytes()
}
