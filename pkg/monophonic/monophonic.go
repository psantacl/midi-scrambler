package monophonic

import (
	"os"
	"fmt"
	"bytes"
	// "reflect"
	"gitlab.com/gomidi/midi/v2/smf"
	"com.github/psantacl/midi-scrambler/pkg/logging"
)

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

	ourEvents := []smf.TrackEvent{}
	currentNote := uint8(0)
	inNote  := false
	deltaDrop := uint32(0)
	// i :=  0

	tracksReader.Do(func(ev smf.TrackEvent) {
		// i += 1
		// if (i > 20) {
		// 	return
		// }

		logging.Sugar.Infow("next event",
			"track", ev.TrackNo,
			"ms", ev.AbsMicroSeconds,
			"ticks", ev.AbsTicks,
			"beat-clock ticks", ev.AbsTicks / beatClockRatio,
			"delta", ev.Delta,
			"message", ev.Message)

		logging.Sugar.Sync()

		// fmt.Printf("track %v @%vms smf ticks: %+v beat clock ticks: %+v  %s\n", ev.TrackNo, ev.AbsMicroSeconds/1000, ev.AbsTicks, ev.AbsTicks / beatClockRatio, ev.Message)
		var _ch, _key, _vel uint8
		switch {
		case ev.Message.GetNoteOn(&_ch, &_key, &_vel):
			if (!inNote) {
				// fmt.Printf("Including noteOn %v @ %v \n", _key, ev.AbsTicks)
				inNote = true
				ev.Delta = ev.Delta + deltaDrop
				deltaDrop = 0
				currentNote = _key
				ourEvents = append(ourEvents, ev)
			} else { //already in a note
				logging.Sugar.Infow("dropping NoteOn", "delta", ev.Delta, "key", _key)
				deltaDrop += ev.Delta
			}

		case ev.Message.GetNoteOff(&_ch, &_key, &_vel):
			if (inNote && currentNote == _key) {
				// fmt.Printf("Including noteOff %v @ %v \n", _key, ev.AbsTicks)
				ev.Delta += deltaDrop
				deltaDrop = 0
				currentNote = 0
				inNote = false
				ourEvents = append(ourEvents, ev)
			} else { //noteOf for a dropped note
				logging.Sugar.Infow("dropping NoteOff", "delta", ev.Delta, "key", _key)
				deltaDrop += ev.Delta
			}
		default: //pass all Note events
			ev.Delta += deltaDrop
			deltaDrop = 0
			ourEvents = append(ourEvents, ev)
		}

	})

	var midiData = buildMidiOut(ourEvents, ticks.Resolution())

	err = os.WriteFile("chicken-out.midi", midiData, 0644)

	if err != nil {
		panic("Unable to write output midifile")
	}
}

func buildMidiOut(ourEvents []smf.TrackEvent, ticks uint16) []byte {
	var (
		bf    bytes.Buffer
		clock = smf.MetricTicks(ticks) // resolution: 96 ticks per quarternote 960 is also common
		tr    smf.Track
	)
	for _, ev := range ourEvents {
		tr.Add(ev.Delta, ev.Message)
	}
	tr.Close(0)

	s := smf.New()
	s.TimeFormat = clock
	s.Add(tr)
	s.WriteTo(&bf)
	return bf.Bytes()
}
