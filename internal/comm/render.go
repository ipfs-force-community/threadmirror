package comm

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"log/slog"
	"sort"
	"strings"

	"github.com/ipfs-force-community/threadmirror/pkg/xscraper"
	"github.com/ipfs-force-community/threadmirror/pkg/xscraper/generated"
)

func RenderThread(threadURLTemplate, threadID string, data any, logger *slog.Logger) (template.HTML, error) {
	funcMap := template.FuncMap{
		"linkify": linkifyTweetText,
		"displayText": func(tweet *xscraper.Tweet) string {
			return tweet.GetDisplayableText()
		},
		"qrcode": func(threadID string) template.URL {
			b, err := GenQrcode(threadURLTemplate, threadID)
			if err != nil {
				logger.Error("failed to generate qrcode", "error", err)
				return ""
			}
			return template.URL(fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(b)))
		},
	}
	tmpl, err := template.New("thread").Funcs(funcMap).Parse(renderTMPL)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

//go:embed render.tmpl
var renderTMPL string

func linkifyTweetText(text string, entities generated.Entities) template.HTML {
	// linkifyTweetText builds the final string based on rune (Unicode code point) indices
	// rather than byte indices to avoid breaking multi-byte characters.
	type entityReplace struct {
		start int // rune index (inclusive)
		end   int // rune index (exclusive)
		repl  string
	}

	var replaces []entityReplace

	// helper converts a generic entity map that contains `indices` field into a replacement
	addEntityReplace := func(ent map[string]any) {
		indicesRaw, ok := ent["indices"].([]any)
		if !ok || len(indicesRaw) != 2 {
			return
		}
		startF, ok1 := indicesRaw[0].(float64)
		endF, ok2 := indicesRaw[1].(float64)
		if !ok1 || !ok2 {
			return
		}
		start, end := int(startF), int(endF)
		if end > start {
			repl := fmt.Sprintf("<a>%s</a>", textRuneSubstring(text, start, end))
			replaces = append(replaces, entityReplace{start: start, end: end, repl: repl})
		}
	}

	// Urls – these have a typed struct with Indices []int
	for _, u := range entities.Urls {
		if len(u.Indices) != 2 {
			continue
		}
		start, end := u.Indices[0], u.Indices[1]
		if end > start {
			repl := fmt.Sprintf("<a>%s</a>", textRuneSubstring(text, start, end))
			replaces = append(replaces, entityReplace{start: start, end: end, repl: repl})
		}
	}

	// Hashtags, Symbols, UserMentions are represented as map[string]any
	for _, h := range entities.Hashtags {
		addEntityReplace(h)
	}
	for _, s := range entities.Symbols {
		addEntityReplace(s)
	}
	for _, m := range entities.UserMentions {
		addEntityReplace(m)
	}

	// sort replacements by start index (ascending)
	sort.Slice(replaces, func(i, j int) bool { return replaces[i].start < replaces[j].start })

	runes := []rune(text)
	var builder strings.Builder
	current := 0
	for _, r := range replaces {
		// ensure indices are within bounds and non-overlapping in rune space
		if r.start < current || r.end > len(runes) {
			continue
		}
		builder.WriteString(string(runes[current:r.start]))
		builder.WriteString(r.repl)
		current = r.end
	}
	builder.WriteString(string(runes[current:]))

	return template.HTML(builder.String())
}

// textRuneSubstring returns the substring of s denoted by rune indices [start, end).
func textRuneSubstring(s string, start, end int) string {
	if start < 0 || end < start {
		return ""
	}
	runes := []rune(s)
	if start >= len(runes) {
		return ""
	}
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[start:end])
}
