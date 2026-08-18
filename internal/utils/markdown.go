package utils

import (
	"bufio"
	"regexp"
	"strings"
)

var (
	reH4         = regexp.MustCompile(`^####\s+(.*)$`)
	reH3         = regexp.MustCompile(`^###\s+(.*)$`)
	reH2         = regexp.MustCompile(`^##\s+(.*)$`)
	reH1         = regexp.MustCompile(`^#\s+(.*)$`)
	reOrdered    = regexp.MustCompile(`^\s*\d+\.\s+(.*)$`)
	reUnordered  = regexp.MustCompile(`^\s*\*\s+(.*)$`)
	reHr         = regexp.MustCompile(`^(\-{3,}|\*{3,}|\_{3,})$`)
	reDoubleStar = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	reDoubleUnd  = regexp.MustCompile(`__([^_]+)__`)
)

// MarkdownToTypst converts common Markdown syntax into equivalent Typst markup
// suitable for embedding into Typst templates.
func MarkdownToTypst(input string) string {
	var out strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(input))

	inCodeBlock := false
	inQuote := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Pass code blocks through untouched
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			if inQuote {
				out.WriteString("]\n")
				inQuote = false
			}
			out.WriteString(line)
			out.WriteString("\n")
			continue
		}

		if inCodeBlock {
			out.WriteString(line)
			out.WriteString("\n")
			continue
		}

		// Handle blockquotes
		if strings.HasPrefix(trimmed, "> ") || trimmed == ">" {
			quoteContent := strings.TrimPrefix(trimmed, "> ")
			if quoteContent == ">" {
				quoteContent = ""
			}
			quoteContent = transformInline(quoteContent)
			if !inQuote {
				out.WriteString("#block(inset: (left: 8pt), stroke: (left: 1.5pt + luma(120)))[\n")
				inQuote = true
			}
			out.WriteString(quoteContent)
			out.WriteString("\n")
			continue
		} else if inQuote {
			out.WriteString("]\n")
			inQuote = false
		}

		// Horizontal rule
		if reHr.MatchString(trimmed) {
			out.WriteString("#line(length: 100%, stroke: 0.5pt)\n")
			continue
		}

		// Headings
		if m := reH4.FindStringSubmatch(line); m != nil {
			out.WriteString("==== " + transformInline(m[1]) + "\n")
			continue
		}
		if m := reH3.FindStringSubmatch(line); m != nil {
			out.WriteString("=== " + transformInline(m[1]) + "\n")
			continue
		}
		if m := reH2.FindStringSubmatch(line); m != nil {
			out.WriteString("== " + transformInline(m[1]) + "\n")
			continue
		}
		if m := reH1.FindStringSubmatch(line); m != nil {
			out.WriteString("= " + transformInline(m[1]) + "\n")
			continue
		}

		// Lists
		if m := reOrdered.FindStringSubmatch(line); m != nil {
			indent := line[:strings.Index(line, strings.TrimSpace(line))]
			out.WriteString(indent + "+ " + transformInline(m[1]) + "\n")
			continue
		}
		if m := reUnordered.FindStringSubmatch(line); m != nil {
			indent := line[:strings.Index(line, strings.TrimSpace(line))]
			out.WriteString(indent + "- " + transformInline(m[1]) + "\n")
			continue
		}

		// General line with inline transformations
		out.WriteString(transformInline(line))
		out.WriteString("\n")
	}

	if inQuote {
		out.WriteString("]\n")
	}

	return out.String()
}

func transformInline(text string) string {
	// Convert **bold** -> *bold*
	text = reDoubleStar.ReplaceAllString(text, "*$1*")
	// Convert __bold__ -> *bold*
	text = reDoubleUnd.ReplaceAllString(text, "*$1*")
	return text
}
