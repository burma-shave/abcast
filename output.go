package main

import (
	"fmt"
	"os"
	"strings"
)

// ANSI color codes — no external deps.
const (
	reset   = "\033[0m"
	bold    = "\033[1m"
	dim     = "\033[2m"

	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	cyan    = "\033[36m"
	white   = "\033[97m"
	magenta = "\033[35m"
	orange  = "\033[38;5;208m"
)

func colorize(color, s string) string {
	return color + s + reset
}

// Banner prints the startup header.
func Banner(title string) {
	width := 58
	line := strings.Repeat("─", width)
	fmt.Fprintf(os.Stderr, "\n%s%s%s\n", bold+orange, line, reset)
	fmt.Fprintf(os.Stderr, "  %s%s%s\n", bold+white, "ABCAST — Beltalowda Audiobook Publisher", reset)
	if title != "" {
		fmt.Fprintf(os.Stderr, "  %s%s%s\n", cyan, title, reset)
	}
	fmt.Fprintf(os.Stderr, "%s%s%s\n\n", bold+orange, line, reset)
}

// Step prints a major phase header.
func Step(msg string) {
	fmt.Fprintf(os.Stderr, "%s▶  %s%s\n", bold+cyan, msg, reset)
}

// Info prints a regular status line.
func Info(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "   %s%s%s\n", white, msg, reset)
}

// Detail prints a sub-item line.
func Detail(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "   %s%s%s\n", dim+white, msg, reset)
}

// Skip prints a skipped-file notice.
func Skip(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "   %sskip%s  %s%s%s\n", dim+yellow, reset, dim, msg, reset)
}

// Done prints a success line with a checkmark.
func Done(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "   %s✓%s  %s\n", bold+green, reset, msg)
}

// Fatal prints an error in Drummer's voice and exits.
func Fatal(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "\n%s✗  FELOTA — %s%s\n\n", bold+red, msg, reset)
	os.Exit(1)
}

// ── Drummer lines ─────────────────────────────────────────────────────────────

func DrummerGreet(bookTitle string) {
	fmt.Fprintf(os.Stderr, "   %s\"%s — milowda carry it through the void. Sa sa ke?\"%s\n\n",
		dim+magenta, bookTitle, reset)
}

func DrummerChapterFound(n int) {
	if n == 1 {
		fmt.Fprintf(os.Stderr, "\n   %s\"One chapter. Like im sol in the dark. Kowl, we work with what we got.\"%s\n\n",
			dim+magenta, reset)
	} else {
		fmt.Fprintf(os.Stderr, "\n   %s\"%d chapters. Every one of them gets to space. No one gets left behind.\"%s\n\n",
			dim+magenta, n, reset)
	}
}

func DrummerExtracting() {
	fmt.Fprintf(os.Stderr, "\n   %s\"Cutting the feed. ffmpeg does what it's told or it answers to me.\"%s\n\n",
		dim+magenta, reset)
}

func DrummerExtractDone(n int) {
	fmt.Fprintf(os.Stderr, "\n   %s\"%d files. Clean cuts. No waste. Beltalowda don't waste ereluf.\"%s\n\n",
		dim+magenta, n, reset)
}

func DrummerWritingFeed() {
	fmt.Fprintf(os.Stderr, "\n   %s\"Feed goes out on the wire. Any ship that wants it, takes it.\"%s\n\n",
		dim+magenta, reset)
}

func DrummerDone(outDir string) {
	fmt.Fprintf(os.Stderr, "\n   %s\"Work is done. Ship it. Don't make me come back here.\"%s\n",
		dim+magenta, reset)
	fmt.Fprintf(os.Stderr, "   %sUpload %s%s%s to your server and burn the acima.%s\n\n",
		white, bold+cyan, outDir+"/", reset+white, reset)
}

func DrummerUsageError() {
	fmt.Fprintf(os.Stderr, "\n%s\"Oye. You come to me with no file, no URL — what is this, deting?\"%s\n",
		bold+red, reset)
	fmt.Fprintf(os.Stderr, "%s\"Give me something to work with, kopeng.\"%s\n\n",
		red, reset)
}
