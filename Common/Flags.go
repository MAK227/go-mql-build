package Common

import (
	"errors"
	"fmt"
	"os"

	catppuccin "github.com/catppuccin/go"
	"github.com/charmbracelet/lipgloss"
	flag "github.com/spf13/pflag"
)

var VERSION = "unknown (built from source)"

type MQLConfig struct {
	Version        bool
	Compile        string
	Syntax         string
	Help           bool
	MetaEditorPath string
	PreserveLogs   bool
}

var MqlConfig *MQLConfig

var HelpStyle = lipgloss.
	NewStyle().
	Padding(0, 1).
	Background(lipgloss.Color(catppuccin.Latte.Sapphire().Hex)).
	Foreground(lipgloss.Color("#000000"))

func Highlight(s string, highlight string) string {
	return fmt.Sprintf(s, lipgloss.
		NewStyle().
		Foreground(lipgloss.Color(catppuccin.Latte.Red().Hex)).
		Bold(true).
		Render(highlight))
}

func (c *MQLConfig) ParseCLIArgs() {
	// Parse command line flags

	flag.BoolVarP(&c.Version, "version", "v", false, Highlight(
		"Prints the %sersion of go-mql",
		"v",
	))

	flag.StringVarP(&c.Compile, "compile", "c", "", Highlight(
		"Runs the %sompiler on the MQL file",
		"c",
	))

	flag.StringVarP(&c.Syntax, "syntax", "s", "", Highlight(
		"Checks the %syntax of the MQL file",
		"s",
	))

	flag.BoolVarP(&c.Help, "help", "h", false, Highlight(
		"Prints the %selp and usage menu",
		"h",
	))

	flag.BoolVarP(&c.PreserveLogs, "logs", "l", false, Highlight(
		"Makes the %sog file for the compiled EA/script",
		"l",
	))

	defaultMetaEditorPath := os.Getenv("MQL4_METAEDITOR_PATH")

	if defaultMetaEditorPath == "" {
		defaultMetaEditorPath = "../metaeditor.exe"
	}

	flag.StringVarP(&c.MetaEditorPath, "meta-editor", "m", defaultMetaEditorPath, Highlight(
		"Sets the path to the %setaeditor.exe \nOr picks from $MQL4_METAEDITOR_PATH environment variable",
		"m",
	))

	flag.ErrHelp = errors.New("\n" + HelpStyle.Render("Go-MQL's help & usage menu"))
	flag.CommandLine.SortFlags = false

	flag.Parse()

	if !c.Version && c.Compile == "" && c.Syntax == "" {
		flag.Usage()
		fmt.Println()
		fmt.Println(HelpStyle.Render("Go-MQL's help & usage menu"))
	}
}
