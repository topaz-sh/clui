package clui

import (
	"io"
)

type UIOption func(*UI)

func WithInput(input io.Reader) UIOption {
	return func(ui *UI) {
		ui.input = input
	}
}

func WithOutput(output io.Writer) UIOption {
	return func(ui *UI) {
		ui.output = output
	}
}

func WithOutputType(outputType OutputType) UIOption {
	return func(ui *UI) {
		ui.outputType = outputType
	}
}
