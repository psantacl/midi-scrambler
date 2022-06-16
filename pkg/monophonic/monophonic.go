package monophonic

import (
	"os"
	"fmt"
	"bytes"
	// "reflect"
	"gitlab.com/gomidi/midi/v2/smf"
	"com.github/psantacl/midi-scrambler/pkg/logging"
)

type TrackEventsPrime struct {
	smf.TrackEvent
	survived bool
}

func ProcessFile(midiFile string) {
	data, err := os.ReadFile(midiFile)
	if err != nil {
		logging.Sugar.Errorf("unable to read midi file '%+v'", midiFile)
		panic(err)

	}

	// fmt.Println(string(data))

	bytesReader := bytes.NewReader(data)
	// fmt.Println(reflect.TypeOf(smf.ReadTracksFrom(bytes.NewReader(data))))
	// fmt.Println(reflect.TypeOf(bytes.NewReader(data)))
	tracksReader := smf.ReadTracksFrom(bytesReader)
	ticks := tracksReader.SMF().TimeFormat.(smf.MetricTicks)
	beatClockRatio := int64(ticks.Resolution()) / 24
	trackCount := len(tracksReader.SMF().Tracks)
	if trackCount != 1 {
		panic(fmt.Sprintf("Can only process Midi files with a single track for now!. Found %s tracks", trackCount))
	}
	fmt.Printf("ticks: %+v track_count: %+v\n", ticks.Resolution(), len(tracksReader.SMF().Tracks))

	ourEvents := []TrackEventsPrime{}
	currentNote := uint8(0)
	inNote  := false
	deltaDrop := uint32(0)


	tracksReader.Do(func(ev smf.TrackEvent) {
		survived := false

		logging.Sugar.Infow("next event",
			"track", ev.TrackNo,
			"ms", ev.AbsMicroSeconds,
			"ticks", ev.AbsTicks,
			"beat-clock ticks", ev.AbsTicks / beatClockRatio,
			"delta", ev.Delta,
			"message", ev.Message)

		logging.Sugar.Sync()

		var _ch, _key, _vel uint8
		switch {
		case ev.Message.GetNoteOn(&_ch, &_key, &_vel):
			if (!inNote) {
				inNote = true
				ev.Delta = ev.Delta + deltaDrop
				deltaDrop = 0
				currentNote = _key
				survived = true
			} else { //already in a note
				logging.Sugar.Infow("dropping NoteOn", "delta", ev.Delta, "key", _key)
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
				logging.Sugar.Infow("dropping NoteOff", "delta", ev.Delta, "key", _key)
				deltaDrop += ev.Delta
			}
		default: //pass all Note events
			ev.Delta += deltaDrop
			deltaDrop = 0
			survived = true
		}
		ourEvents = append(ourEvents, TrackEventsPrime{TrackEvent: ev, survived: survived})

	})

	var midiData = buildMidiOut(ourEvents, ticks.Resolution())

	err = os.WriteFile("chicken-out.midi", midiData, 0644)

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
