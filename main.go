package main

import (
	"fmt"
	"runtime/debug"

	common "github.com/MAK227/go-mql-build/Common"
)

var readFileCache map[string][]string

func main() {
	cfg := &common.MQLConfig{}

	cfg.ParseCLIArgs()

	if cfg.Version {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
			common.VERSION = info.Main.Version
		}
		fmt.Println("Go-MQL's version:", common.HelpStyle.
			Render(common.VERSION),
		)
		return
	}

	readFileCache = make(map[string][]string)

	common.InitLogger()

	if cfg.Compile != "" {

		compileTarget, logfile := common.BuildCompileTarget(cfg.Compile)

		outputStr, status := common.Compile(cfg.Compile, logfile, compileTarget)

		var diagnostics common.Diagnostic

		diagnostics = common.ParseLogFile(outputStr, status, "compile")

		common.PrintDiagnostics(diagnostics, readFileCache)

		return
	}

	if cfg.Syntax != "" {
		syntaxTarget, logfile := common.BuildCompileTarget(cfg.Syntax)

		outputStr, status := common.SyntaxCheck(cfg.Syntax, logfile, syntaxTarget)

		var diagnostics common.Diagnostic

		diagnostics = common.ParseLogFile(outputStr, status, "syntax")

		common.PrintDiagnostics(diagnostics, readFileCache)

		return
	}
	/*
	   ╭──── Script: Scripts/pocketbase.mq4, Line: 191, Char: 20, Type: warning, Code: 43 ───╮
	   │                                                                                     │
	   │                                                                                     │
	   │  191 |     JWT_EXPIRY = json["exp"].ToInt();                                        │
	   │                       │                                                             │
	   │                       ╰─➤ possible loss of data due to type conversion              │
	   ╰─────────────────────────────────────────────────────────────────────────────────────╯
	*/
}
