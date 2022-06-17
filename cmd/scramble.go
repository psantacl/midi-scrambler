package cmd

import (
	"github.com/spf13/cobra"
	// "com.github/psantacl/midi-scrambler/pkg/logging"
	"com.github/psantacl/midi-scrambler/pkg/average"
)


var (
	inMidiFile string
	outMidiFile string
	monophonic bool
)


var averageCmd = &cobra.Command{
	Use:   "average",
	Short: "select notes based on moving average",
	Long:  "select notes based on moving average",
	Run: func(cmd *cobra.Command, args []string) {
		average.ProcessFile(inMidiFile, outMidiFile, monophonic)
	},
}


func init() {
	averageCmd.Flags().StringVarP(&inMidiFile, "in-midi-file", "i", "", "input midi file to scramble (required)")
	averageCmd.Flags().StringVarP(&outMidiFile, "out-midi-file", "o", "", "output midi file name (required)")
	averageCmd.Flags().BoolVarP( &monophonic, "monophonic", "m", false, "reduce file to monophonic")
	averageCmd.MarkFlagRequired("in-midi-file")
	averageCmd.MarkFlagRequired("out-midi-file")

	RootCmd.AddCommand(averageCmd)
}
