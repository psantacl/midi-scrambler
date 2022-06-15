package cmd

import (
	"github.com/spf13/cobra"
	"com.github/psantacl/midi-scrambler/pkg/logging"
	"com.github/psantacl/midi-scrambler/pkg/monophonic"
)


var (
	MidiFile string
)


var monophonicCmd = &cobra.Command{
	Use:   "monophonic",
	Short: "probabilistically reduce midi file to monophonic",
	Long:  "probabilistically reduce midi file to monophonic",
	Run: func(cmd *cobra.Command, args []string) {
		logging.Sugar.Infof("making '%+v' monophonic", MidiFile)
		monophonic.ProcessFile(MidiFile)
	},
}

var averageCmd = &cobra.Command{
	Use:   "average",
	Short: "select notes based on moving average",
	Long:  "select notes based on moving average",
	Run: func(cmd *cobra.Command, args []string) {
		logging.Sugar.Infof("averaging %+v", MidiFile);
	},
}


func init() {
	averageCmd.Flags().StringVarP(&MidiFile, "midi-file", "m", "", "midi file to scramble (required)")
	averageCmd.MarkFlagRequired("midi-file")

	monophonicCmd.Flags().StringVarP(&MidiFile, "midi-file", "m", "", "midi file to scramble (required)")
	monophonicCmd.MarkFlagRequired("midi-file")

	RootCmd.AddCommand(averageCmd)
	RootCmd.AddCommand(monophonicCmd)
}
