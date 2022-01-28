package clui

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ettle/strcase"
	"github.com/fatih/color"
	"github.com/kyokomi/emoji"
	"github.com/olekukonko/tablewriter"
)

// Normal returns a UIMessage that prints a normal message
func (u *UI) Normal() *Message {
	return &Message{
		ui:           u,
		msgType:      normal,
		interactions: []interaction{},
		end:          -1,
	}
}

// Exclamation returns a UIMessage that prints an exclamation message
func (u *UI) Exclamation() *Message {
	return &Message{
		ui:           u,
		msgType:      exclamation,
		interactions: []interaction{},
		end:          -1,
	}
}

// Note returns a UIMessage that prints a note message
func (u *UI) Note() *Message {
	return &Message{
		ui:           u,
		msgType:      note,
		interactions: []interaction{},
		end:          -1,
	}
}

// Success returns a UIMessage that prints a success message
func (u *UI) Success() *Message {
	return &Message{
		ui:           u,
		msgType:      success,
		interactions: []interaction{},
		end:          -1,
	}
}

// Problem returns a Message that prints a message that describes a problem
func (u *UI) Problem() *Message {
	return &Message{
		ui:           u,
		msgType:      problem,
		interactions: []interaction{},
		end:          -1,
	}
}

// Msgf prints a formatted message on the CLI
func (u *Message) Msgf(message string, a ...interface{}) {
	u.Msg(fmt.Sprintf(message, a...))
}

// Do is syntactic sugar for Msg("")
func (u *Message) Do() {
	u.Msg("")
}

// Msg prints a message on the CLI, resolving emoji as it goes
func (u *Message) Msg(message string) {
	switch u.ui.outputType {
	case OutputTypeHuman:
		u.humanMsg(message)
	case OutputTypeJSON:
		u.jsonMsg(message)
	}
}

func (u *Message) humanMsg(message string) {
	message = emoji.Sprint(message)

	// Print a newline before starting output, if not compact.
	if message != "" && !u.compact {
		u.ui.println()
	}

	if !u.noNewline {
		message += "\n"
	}

	switch u.msgType {
	case normal:
	case exclamation:
		message = color.YellowString(message)
	case note:
		message = color.BlueString(message)
	case success:
		message = color.GreenString(message)
	case progress:
	case problem:
		message = color.RedString(message)
	}

	u.ui.printf("%s", message)

	for _, interaction := range u.interactions {
		switch interaction.variant {
		case ask:
			switch interaction.valueType {
			case tBool:
				*(interaction.value.(*bool)) = u.readBool(interaction.name, interaction.boolMap)
			case tInt:
				*(interaction.value.(*int64)) = u.readInt(interaction.name)
			case tString:
				*(interaction.value.(*string)) = u.readString(interaction.name)
			case tPassword:
				*(interaction.value.(*string)) = u.readPassword(interaction.name, interaction.stdin)
			}
		case show:
			switch interaction.valueType {
			case tBool:
				u.ui.printf("%s: %s\n", emoji.Sprint(interaction.name), color.MagentaString("%t", interaction.value))
			case tInt:
				u.ui.printf("%s: %s\n", emoji.Sprint(interaction.name), color.CyanString("%d", interaction.value))
			case tString:
				u.ui.printf("%s: %s\n", emoji.Sprint(interaction.name), color.GreenString("%s", interaction.value))
			case tErr:
				if u.stacks {
					u.ui.printf("%s\n", color.RedString("%+v", interaction.value))
				} else {
					u.ui.printf("%s\n", color.RedString("%v", interaction.value))
				}
			}
		}
	}

	for idx, headers := range u.tableHeaders {
		table := tablewriter.NewWriter(u.ui.output)
		table.SetHeader(headers)
		table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
		table.SetBorder(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator("")
		table.SetHeaderLine(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetColumnSeparator("")

		if idx < len(u.tableData) {
			table.AppendBulk(u.tableData[idx])
		}

		table.Render()
	}

	if u.end > -1 {
		os.Exit(u.end)
	}
}

func (u *Message) jsonMsg(message string) {
	output := map[string]interface{}{}

	output["message"] = message

	// Go through all the questions first
	for _, interaction := range u.interactions {
		if interaction.variant != ask {
			continue
		}

		switch interaction.valueType {
		case tBool:
			*(interaction.value.(*bool)) = u.readBool(interaction.name, interaction.boolMap)
		case tInt:
			*(interaction.value.(*int64)) = u.readInt(interaction.name)
		case tString:
			*(interaction.value.(*string)) = u.readString(interaction.name)
		case tPassword:
			*(interaction.value.(*string)) = u.readPassword(interaction.name, interaction.stdin)
		}
	}

	// Go through all outputs next
	for _, interaction := range u.interactions {
		if interaction.variant != show {
			continue
		}

		output[interaction.name] = interaction.value
	}

	dataOutput := []map[string]interface{}{}

	for idx, headers := range u.tableHeaders {
		for rowIdx, row := range u.tableData[idx] {
			if idx >= len(u.tableData) || rowIdx >= len(u.tableData[idx]) {
				continue
			}

			dataRow := map[string]interface{}{}

			for colIdx, colValue := range row {
				if colIdx >= len(headers) {
					continue
				}

				key := strcase.ToCamel(headers[colIdx])
				dataRow[key] = colValue
			}

			dataOutput = append(dataOutput, dataRow)
		}
	}

	output["data"] = dataOutput

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		u.ui.printf("%s\n", color.RedString("%v", err))
	}

	u.ui.printf("%s\n", string(jsonData))

	if u.end > -1 {
		os.Exit(u.end)
	}
}

// NoNewline disables the printing of a newline after a message output
func (u *Message) NoNewline() *Message {
	u.noNewline = true
	return u
}

// Compact disables the printing of a newline before starting output
func (u *Message) Compact() *Message {
	u.compact = true
	return u
}

// WithEnd ends the entire process after printing the message.
func (u *Message) WithEnd(code int) *Message {
	u.end = code
	return u
}

// WithStack causes error stacks to be printed.
func (u *Message) WithStacks() *Message {
	u.stacks = true
	return u
}

// WithBoolValue adds a bool value to be printed in the message
func (u *Message) WithBoolValue(name string, value bool) *Message {
	u.interactions = append(u.interactions, interaction{
		name:      name,
		variant:   show,
		valueType: tBool,
		value:     value,
	})
	return u
}

// WithStringValue adds a string value to be printed in the message
func (u *Message) WithStringValue(name string, value string) *Message {
	u.interactions = append(u.interactions, interaction{
		name:      name,
		variant:   show,
		valueType: tString,
		value:     value,
	})
	return u
}

// WithErr adds an error value to be printed in the message
func (u *Message) WithErr(err error) *Message {
	u.interactions = append(u.interactions, interaction{
		variant:   show,
		valueType: tErr,
		value:     err,
	})
	return u
}

// WithIntValue adds an int value to be printed in the message
func (u *Message) WithIntValue(name string, value int64) *Message {
	u.interactions = append(u.interactions, interaction{
		name:      name,
		variant:   show,
		valueType: tInt,
		value:     value,
	})
	return u
}
