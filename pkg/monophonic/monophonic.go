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
	tracksReader.Do(func(ev smf.TrackEvent) {
		fmt.Printf("track %v @%vms smf ticks: %+v beat clock ticks: %+v  %s\n", ev.TrackNo, ev.AbsMicroSeconds/1000, ev.AbsTicks, ev.AbsTicks / beatClockRatio, ev.Message)
		ourEvents = append(ourEvents, ev)

	})

	var midiData = buildMidiOut( ourEvents, ticks.Resolution())

	err = os.WriteFile("chicken-out.midi", midiData, 0644)

	// if err != nil {
	// 	panic("Unable to write output midifile")
	// }
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
