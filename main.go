package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	common "github.com/MAK227/go-mql-build/Common"
	catppuccin "github.com/catppuccin/go"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	flag "github.com/spf13/pflag"
)

var readFileCache map[string][]string

func runBuild(mode string, target string, cfg *common.MQLConfig) {
	compileTarget, logfile := common.BuildCompileTarget(target)

	var outputStr string
	status := 0

	switch mode {
	case "compile":
		outputStr, status = common.Compile(target, logfile, compileTarget, cfg)
	case "syntax":
		outputStr, status = common.SyntaxCheck(target, logfile, compileTarget, cfg)
	default:
		fmt.Println("Invalid mode:", mode)
		return
	}

	diagnostics := common.ParseLogFile(outputStr, status, mode)

	common.PrintDiagnostics(diagnostics, readFileCache)

	if !cfg.PreserveLogs {
		os.Remove(logfile)
	} else {
		fmt.Println()
		fmt.Println("Logs are saved in", lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(catppuccin.Latte.Lavender().Hex)).Render(logfile))
	}
}

func main() {
	cfg := &common.MQLConfig{}

	cfg.ParseCLIArgs()

	if cfg.Version {
		fmt.Println("Go-MQL's version:", common.HelpStyle.
			Render(common.VERSION),
		)

		return
	}

	readFileCache = make(map[string][]string)

	common.InitLogger()

	if cfg.Compile != "" {
		runBuild("compile", cfg.Compile, cfg)
		return
	}

	if cfg.Syntax != "" {
		runBuild("syntax", cfg.Syntax, cfg)
		return
	}

	if cfg.Help {
		flag.Usage()
		fmt.Println()
		fmt.Println(common.HelpStyle.Render("Go-MQL's help & usage menu", common.VERSION))
		return
	}

	if !cfg.Version && cfg.Compile == "" && cfg.Syntax == "" {

		var filePicker common.FilePicker
		var err error
		var m tea.Model

		p := tea.NewProgram(filePicker, tea.WithAltScreen())
		if m, err = p.Run(); err != nil {
			log.Fatal(err)
		}

		filePicker = m.(common.FilePicker)

		if len(filePicker.Files) == 0 {
			common.PrintError(errors.New("No .mq4 files found in the current directory"))
			return
		}

		if filePicker.Mode == "compile" {
			runBuild("compile", filePicker.Files[filePicker.CurrIndex].Path, cfg)
			return
		}

		if filePicker.Mode == "syntax" {
			runBuild("syntax", filePicker.Files[filePicker.CurrIndex].Path, cfg)
			return
		}

		// INFO: Shows the help menu (default)
		flag.Usage()
		fmt.Println()
		fmt.Println(common.HelpStyle.Render("Go-MQL's help & usage menu", common.VERSION))
	}
}
