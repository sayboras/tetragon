// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package cmdref

import (
	"os"
	"strings"

	"github.com/cilium/tetragon/pkg/logger"
	"github.com/cilium/tetragon/pkg/logger/logfields"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var log = logger.DefaultSlogLogger.With(logfields.LogSubsys, "cmdref")

func linkHandler(s string) string {
	// The generated files have a 'See also' section but the URL's are
	// hardcoded to use Markdown but we only want / have them in HTML
	// later.
	return strings.Replace(s, ".md", ".html", 1)
}

func filePrepend(_ string) string {
	// Prepend a HTML comment that this file is autogenerated. So that
	// users are warned before fixing issues in the Markdown files.  Should
	// never show up on the web.
	return "<!-- This file was autogenerated via tetragon-operator --cmdref, do not edit manually-->" + "\n\n"
}

func GenMarkdown(cmd *cobra.Command, cmdRefDir string) {
	// Remove the line 'Auto generated by spf13/cobra on ...'
	cmd.DisableAutoGenTag = true
	if err := doc.GenMarkdownTreeCustom(cmd, cmdRefDir, filePrepend, linkHandler); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
