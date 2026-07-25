package main

import (
	"bytes"
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/ledongthuc/pdf"
)

type ConversionResult struct {
	Success  bool   `json:"success"`
	Markdown string `json:"markdown"`
	Error    string `json:"error"`
}

type TextSpan struct {
	Font     string
	FontSize float64
	X, Y     float64
	W        float64
	Text     string
	Bold     bool
	Italic   bool
}

type Line struct {
	Y     float64
	Spans []TextSpan
}

type BlockKind int

const (
	BlockParagraph BlockKind = iota
	BlockHeading1
	BlockHeading2
	BlockHeading3
	BlockHeading4
	BlockHeading5
	BlockHeading6
	BlockBulletList
	BlockNumberedList
	BlockTable
	BlockCode
)

type Block struct {
	Kind  BlockKind
	Lines []Line
}

type fontStyleCache struct {
	bold   map[string]bool
	italic map[string]bool
}

func detectFileType(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".pdf":
		return "pdf"
	default:
		return "unknown"
	}
}

func ConvertToMarkdown(filePath string) ConversionResult {
	fileName := filepath.Base(filePath)
	fileType := detectFileType(fileName)
	switch fileType {
	case "pdf":
		return convertPDFFromPath(filePath)
	default:
		return ConversionResult{
			Success: false,
			Error:   fmt.Sprintf("Unsupported file type: %s. Currently only .pdf is supported.", filepath.Ext(fileName)),
		}
	}
}

func ConvertFileContent(fileName string, content []byte) ConversionResult {
	fileType := detectFileType(fileName)
	switch fileType {
	case "pdf":
		return convertPDF(content)
	default:
		return ConversionResult{
			Success: false,
			Error:   fmt.Sprintf("Unsupported file type: %s. Currently only .pdf is supported.", filepath.Ext(fileName)),
		}
	}
}

func convertPDFFromPath(filePath string) ConversionResult {
	f, reader, err := pdf.Open(filePath)
	if err != nil {
		return ConversionResult{Success: false, Error: "Cannot open PDF: " + err.Error()}
	}
	defer f.Close()
	return convertReader(reader)
}

func convertPDF(data []byte) ConversionResult {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return ConversionResult{Success: false, Error: "Cannot parse PDF: " + err.Error()}
	}
	return convertReader(reader)
}

func convertReader(reader *pdf.Reader) ConversionResult {
	var md strings.Builder
	totalPages := reader.NumPage()
	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		page := reader.Page(pageNum)
		if page.V.IsNull() {
			continue
		}
		pageMD := convertPage(page, pageNum)
		if strings.TrimSpace(pageMD) == "" {
			continue
		}
		if totalPages > 1 {
			md.WriteString("<!-- Page " + strconv.Itoa(pageNum) + " -->\n\n")
		}
		md.WriteString(pageMD)
		md.WriteString("\n")
	}
	result := strings.TrimSpace(md.String())
	if result == "" {
		return ConversionResult{Success: false, Error: "No text content found in PDF. The file might be a scanned/image-based PDF."}
	}
	return ConversionResult{Success: true, Markdown: result}
}

func buildFontStyleCache(page pdf.Page) fontStyleCache {
	cache := fontStyleCache{
		bold:   make(map[string]bool),
		italic: make(map[string]bool),
	}
	for _, fontName := range page.Fonts() {
		font := page.Font(fontName)
		bold := false
		italic := false
		fd := font.V.Key("FontDescriptor")
		if fd.Kind() == pdf.Dict {
			flags := fd.Key("Flags").Int64()
			if flags&(1<<18) != 0 {
				bold = true
			}
			if flags&(1<<6) != 0 {
				italic = true
			}
		}
		baseName := font.BaseFont()
		if i := strings.Index(baseName, "+"); i >= 0 {
			baseName = baseName[i+1:]
		}
		lower := strings.ToLower(baseName)
		if !bold {
			bold = strings.Contains(lower, "bold") ||
				strings.Contains(lower, "black") ||
				strings.Contains(lower, "heavy") ||
				strings.Contains(lower, "demibold") ||
				strings.Contains(lower, "semibold")
		}
		if !italic {
			italic = strings.Contains(lower, "italic") ||
				strings.Contains(lower, "oblique") ||
				strings.Contains(lower, "slant")
		}
		cache.bold[fontName] = bold
		cache.italic[fontName] = italic
	}
	return cache
}

func extractSpans(page pdf.Page, fontCache *fontStyleCache) []TextSpan {
	content := page.Content()
	spans := make([]TextSpan, 0, len(content.Text))
	for _, t := range content.Text {
		text := strings.TrimSpace(t.S)
		if text == "" {
			continue
		}
		bold := fontCache.bold[t.Font]
		italic := fontCache.italic[t.Font]
		spans = append(spans, TextSpan{
			Font:     t.Font,
			FontSize: t.FontSize,
			X:        t.X,
			Y:        t.Y,
			W:        t.W,
			Text:     text,
			Bold:     bold,
			Italic:   italic,
		})
	}
	return spans
}

func groupLines(spans []TextSpan) []Line {
	if len(spans) == 0 {
		return nil
	}
	sort.Slice(spans, func(i, j int) bool {
		if math.Abs(spans[i].Y-spans[j].Y) > 3 {
			return spans[i].Y > spans[j].Y
		}
		return spans[i].X < spans[j].X
	})
	const yTolerance = 3.0
	lines := []Line{}
	current := Line{Y: spans[0].Y, Spans: []TextSpan{spans[0]}}
	for i := 1; i < len(spans); i++ {
		if math.Abs(spans[i].Y-current.Y) <= yTolerance {
			current.Spans = append(current.Spans, spans[i])
		} else {
			sort.Slice(current.Spans, func(a, b int) bool {
				return current.Spans[a].X < current.Spans[b].X
			})
			current = mergeSpansOnLine(current)
			lines = append(lines, current)
			current = Line{Y: spans[i].Y, Spans: []TextSpan{spans[i]}}
		}
	}
	sort.Slice(current.Spans, func(a, b int) bool {
		return current.Spans[a].X < current.Spans[b].X
	})
	current = mergeSpansOnLine(current)
	lines = append(lines, current)
	return lines
}

func normalizeFontName(font string) string {
	lower := strings.ToLower(font)
	for _, suffix := range []string{"-bold", "-italic", "-bolditalic", "-oblique",
		"bold", "italic", "bolditalic", "oblique",
		"black", "heavy", "demibold", "semibold",
		"mt", "ps", "tt"} {
		lower = strings.TrimSuffix(lower, suffix)
	}
	lower = strings.TrimPrefix(lower, "+")
	return lower
}

// mergeSpansOnLine merges adjacent spans that belong to the same word.
//
// The PDF library does not track word spacing (Tw) in position calculations,
// so gap-based detection fails. Instead, we detect word boundaries from text
// content: punctuation endings, capitalization transitions, and character class
// changes. When no boundary is detected, spans merge (same word). When a
// boundary is found, a space is inserted between them.
func mergeSpansOnLine(line Line) Line {
	if len(line.Spans) <= 1 {
		return line
	}

	type mergeCandidate struct {
		span       TextSpan
		addSpaceBefore bool // insert space before this span
	}

	candidates := []mergeCandidate{{span: line.Spans[0]}}
	for i := 1; i < len(line.Spans); i++ {
		prev := line.Spans[i-1]
		curr := line.Spans[i]

		sameStyle := prev.Bold == curr.Bold && prev.Italic == curr.Italic
		sameFamily := prev.Font == curr.Font ||
			normalizeFontName(prev.Font) == normalizeFontName(curr.Font)

		if !sameFamily || !sameStyle {
			candidates = append(candidates, mergeCandidate{span: curr, addSpaceBefore: true})
			continue
		}

		addSpace := needsWordSpace(prev.Text, curr.Text)
		candidates = append(candidates, mergeCandidate{span: curr, addSpaceBefore: addSpace})
	}

	// Build final spans: merge consecutive same-word spans, insert spaces at boundaries
	merged := []TextSpan{candidates[0].span}
	for i := 1; i < len(candidates); i++ {
		prev := &merged[len(merged)-1]
		c := candidates[i]

		if c.addSpaceBefore {
			// Word boundary detected — start a new span with a leading space
			merged = append(merged, TextSpan{
				Text:     " " + c.span.Text,
				Font:     c.span.Font,
				FontSize: c.span.FontSize,
				X:        c.span.X,
				Y:        c.span.Y,
				W:        c.span.W,
				Bold:     c.span.Bold,
				Italic:   c.span.Italic,
			})
		} else {
			// Same word — merge into previous span
			prev.Text += c.span.Text
			prev.W = (c.span.X + c.span.W) - prev.X
			prev.Bold = prev.Bold || c.span.Bold
			prev.Italic = prev.Italic || c.span.Italic
		}
	}

	line.Spans = merged
	return line
}

// needsWordSpace determines whether a space should be inserted between two
// consecutive text spans based on content analysis.
func needsWordSpace(prevText, nextText string) bool {
	if len(prevText) == 0 || len(nextText) == 0 {
		return false
	}

	lastRune := rune(prevText[len(prevText)-1])
	firstRune := rune(nextText[0])

	// 1. Space or dash at end — word boundary already marked
	if lastRune == ' ' || lastRune == '-' || lastRune == '–' || lastRune == '—' {
		return false
	}

	// 2. Punctuation at end — word boundary
	if unicode.IsPunct(lastRune) {
		return false
	}

	// 3. New span starts with punctuation — word boundary
	if unicode.IsPunct(firstRune) {
		return false
	}

	// 4. Lowercase → Uppercase transition (e.g., "stack" + "Developer")
	if unicode.IsLower(lastRune) && unicode.IsUpper(firstRune) {
		return true
	}

	// 5. Letter → Digit or Digit → Letter transition
	if (unicode.IsLetter(lastRune) && unicode.IsDigit(firstRune)) ||
		(unicode.IsDigit(lastRune) && unicode.IsLetter(firstRune)) {
		return true
	}

	// 6. Symbol/other → Letter transition (e.g., bullet + text)
	if !unicode.IsLetter(lastRune) && !unicode.IsDigit(lastRune) && unicode.IsLetter(firstRune) {
		return true
	}

	// 7. Letter → Symbol/other transition (e.g., text + bullet)
	if unicode.IsLetter(lastRune) && !unicode.IsLetter(firstRune) && !unicode.IsDigit(firstRune) {
		return true
	}

	// 8. Single uppercase letter followed by single uppercase (likely abbreviation: "B" + "I" → "B I")
	//    But only if the previous text is a single uppercase letter
	if len(prevText) == 1 && unicode.IsUpper(lastRune) && unicode.IsUpper(firstRune) {
		// Check if this looks like an abbreviation (e.g., "B" + "I" in "B I L A L")
		// Abbreviations are usually single uppercase letters
		return true
	}

	// 9. Short uppercase text followed by short uppercase text (abbreviation spacing)
	if len(prevText) <= 2 && len(nextText) <= 2 &&
		unicode.IsUpper(lastRune) && unicode.IsUpper(firstRune) {
		return true
	}

	// Default: no space (assume same word)
	return false
}

func analyzeFontSize(lines []Line) float64 {
	sizeCount := map[float64]int{}
	for _, line := range lines {
		for _, span := range line.Spans {
			if span.FontSize >= 6 {
				rounded := math.Round(span.FontSize*2) / 2
				sizeCount[rounded]++
			}
		}
	}
	if len(sizeCount) == 0 {
		return 12.0
	}
	bestSize := 12.0
	bestCount := 0
	for size, count := range sizeCount {
		if count > bestCount {
			bestCount = count
			bestSize = size
		}
	}
	return bestSize
}

func classifyBlocks(lines []Line, bodySize float64) []Block {
	if len(lines) == 0 {
		return nil
	}
	var blocks []Block
	for _, line := range lines {
		kind := classifyLine(line, bodySize)
		blocks = append(blocks, Block{Kind: kind, Lines: []Line{line}})
	}
	return mergeBlocks(blocks)
}

func classifyLine(line Line, bodySize float64) BlockKind {
	if len(line.Spans) == 0 {
		return BlockParagraph
	}
	maxSize := 0.0
	for _, span := range line.Spans {
		if span.FontSize > maxSize {
			maxSize = span.FontSize
		}
	}
	lineText := mergeLineText(line)
	if len(lineText) < 2 {
		return BlockParagraph
	}
	if isCodeFont(line.Spans) {
		return BlockCode
	}
	if isBulletList(lineText) {
		return BlockBulletList
	}
	if isNumberedList(lineText) {
		return BlockNumberedList
	}
	if bodySize > 0 {
		ratio := maxSize / bodySize
		if ratio >= 1.8 {
			return BlockHeading1
		}
		if ratio >= 1.4 {
			return BlockHeading2
		}
		if ratio >= 1.2 {
			return BlockHeading3
		}
	}
	if allSpansBold(line.Spans) {
		lineLen := len(lineText)
		if lineLen < 50 {
			return BlockHeading3
		}
		if lineLen < 80 {
			return BlockHeading4
		}
	}
	for _, span := range line.Spans {
		if span.Bold && span.FontSize > bodySize*1.08 {
			if len(lineText) < 80 {
				return BlockHeading4
			}
		}
	}
	return BlockParagraph
}

func mergeLineText(line Line) string {
	var sb strings.Builder
	for i, span := range line.Spans {
		if i > 0 {
			prev := line.Spans[i-1]
			gap := span.X - (prev.X + prev.W)
			if gap > 2 {
				sb.WriteString(" ")
			}
		}
		sb.WriteString(span.Text)
	}
	return sb.String()
}

func isCodeFont(spans []TextSpan) bool {
	if len(spans) == 0 {
		return false
	}
	monoCount := 0
	for _, span := range spans {
		font := strings.ToLower(span.Font)
		if strings.Contains(font, "courier") ||
			strings.Contains(font, "mono") ||
			strings.Contains(font, "consolas") ||
			strings.Contains(font, "menlo") ||
			strings.Contains(font, "fira") {
			monoCount++
		}
	}
	return monoCount == len(spans)
}

func allSpansBold(spans []TextSpan) bool {
	if len(spans) == 0 {
		return false
	}
	for _, span := range spans {
		if !span.Bold {
			return false
		}
	}
	return true
}

func isBulletList(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false
	}
	bulletPrefixes := []string{"\u2022", "\u25cf", "\u25cb", "\u2023", "\u2043", "\u25c6", "\u25c7", "\u25aa", "\u25ab"}
	for _, prefix := range bulletPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
		return true
	}
	return false
}

func isNumberedList(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" || len(trimmed) < 2 {
		return false
	}
	for i, r := range trimmed {
		if unicode.IsDigit(r) {
			continue
		}
		if r == '.' || r == ')' {
			rest := trimmed[i+1:]
			if len(rest) > 0 && rest[0] == ' ' {
				return true
			}
			return false
		}
		if r == '(' && i == 0 {
			continue
		}
		return false
	}
	return false
}

func mergeBlocks(blocks []Block) []Block {
	if len(blocks) == 0 {
		return nil
	}
	var merged []Block
	current := blocks[0]
	for i := 1; i < len(blocks); i++ {
		next := blocks[i]
		if next.Kind == current.Kind && !isHeadingKind(current.Kind) {
			current.Lines = append(current.Lines, next.Lines...)
		} else {
			merged = append(merged, current)
			current = next
		}
	}
	merged = append(merged, current)
	return merged
}

func isHeadingKind(k BlockKind) bool {
	return k >= BlockHeading1 && k <= BlockHeading6
}

func emitMarkdown(blocks []Block) string {
	var md strings.Builder
	for i, block := range blocks {
		if i > 0 {
			prevKind := blocks[i-1].Kind
			if prevKind != block.Kind || isHeadingKind(block.Kind) {
				md.WriteString("\n")
			}
		}
		switch block.Kind {
		case BlockHeading1:
			md.WriteString("# " + emitInlineBlock(block) + "\n")
		case BlockHeading2:
			md.WriteString("## " + emitInlineBlock(block) + "\n")
		case BlockHeading3:
			md.WriteString("### " + emitInlineBlock(block) + "\n")
		case BlockHeading4:
			md.WriteString("#### " + emitInlineBlock(block) + "\n")
		case BlockHeading5:
			md.WriteString("##### " + emitInlineBlock(block) + "\n")
		case BlockHeading6:
			md.WriteString("###### " + emitInlineBlock(block) + "\n")
		case BlockBulletList:
			for _, line := range block.Lines {
				text := mergeLineText(line)
				text = stripBulletPrefix(text)
				text = emitInlineLine(line)
				md.WriteString("- " + text + "\n")
			}
		case BlockNumberedList:
			num := 1
			for _, line := range block.Lines {
				text := mergeLineText(line)
				text = stripNumberPrefix(text)
				text = emitInlineLine(line)
				md.WriteString(strconv.Itoa(num) + ". " + text + "\n")
				num++
			}
		case BlockCode:
			md.WriteString("```\n")
			for _, line := range block.Lines {
				md.WriteString(mergeLineText(line) + "\n")
			}
			md.WriteString("```\n")
		case BlockTable:
			emitTable(&md, block)
		case BlockParagraph:
			text := emitInlineBlock(block)
			md.WriteString(text + "\n")
		}
	}
	return md.String()
}

func emitInlineBlock(block Block) string {
	if len(block.Lines) == 0 {
		return ""
	}
	if len(block.Lines) == 1 {
		return emitInlineLine(block.Lines[0])
	}
	var parts []string
	for _, line := range block.Lines {
		text := emitInlineLine(line)
		if text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}

func emitInlineLine(line Line) string {
	if len(line.Spans) == 0 {
		return ""
	}
	type wordToken struct {
		text   string
		bold   bool
		italic bool
	}
	var tokens []wordToken
	currentWord := wordToken{
		text:   line.Spans[0].Text,
		bold:   line.Spans[0].Bold,
		italic: line.Spans[0].Italic,
	}
	for i := 1; i < len(line.Spans); i++ {
		prev := line.Spans[i-1]
		curr := line.Spans[i]
		gap := curr.X - (prev.X + prev.W)
		sameStyle := prev.Bold == curr.Bold && prev.Italic == curr.Italic
		if gap < 3 && sameStyle {
			currentWord.text += curr.Text
			currentWord.bold = currentWord.bold || curr.Bold
			currentWord.italic = currentWord.italic || curr.Italic
		} else {
			if currentWord.text != "" {
				tokens = append(tokens, currentWord)
			}
			currentWord = wordToken{
				text:   curr.Text,
				bold:   curr.Bold,
				italic: curr.Italic,
			}
		}
	}
	if currentWord.text != "" {
		tokens = append(tokens, currentWord)
	}
	var sb strings.Builder
	for i, tok := range tokens {
		if i > 0 {
			sb.WriteString(" ")
		}
		text := tok.text
		if tok.bold {
			text = "**" + text + "**"
		}
		if tok.italic {
			text = "_" + text + "_"
		}
		sb.WriteString(text)
	}
	return sb.String()
}

func stripBulletPrefix(text string) string {
	trimmed := strings.TrimSpace(text)
	bulletPrefixes := []string{"\u2022", "\u25cf", "\u25cb", "\u2023", "\u2043", "\u25c6", "\u25c7", "\u25aa", "\u25ab"}
	for _, prefix := range bulletPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
		}
	}
	if strings.HasPrefix(trimmed, "- ") {
		return strings.TrimPrefix(trimmed, "- ")
	}
	if strings.HasPrefix(trimmed, "* ") {
		return strings.TrimPrefix(trimmed, "* ")
	}
	if strings.HasPrefix(trimmed, "\u2013 ") {
		return strings.TrimPrefix(trimmed, "\u2013 ")
	}
	return trimmed
}

func stripNumberPrefix(text string) string {
	trimmed := strings.TrimSpace(text)
	for i, r := range trimmed {
		if !unicode.IsDigit(r) && r != '(' {
			if (r == '.' || r == ')') && i > 0 {
				rest := trimmed[i+1:]
				if len(rest) > 0 && rest[0] == ' ' {
					return rest[1:]
				}
			}
			break
		}
	}
	if strings.HasPrefix(trimmed, "(") {
		idx := strings.Index(trimmed, ") ")
		if idx > 0 {
			return trimmed[idx+2:]
		}
	}
	return trimmed
}

func emitTable(md *strings.Builder, block Block) {
	if len(block.Lines) == 0 {
		return
	}
	for i, line := range block.Lines {
		spans := line.Spans
		cells := make([]string, len(spans))
		for j, span := range spans {
			cells[j] = strings.TrimSpace(span.Text)
		}
		md.WriteString("| " + strings.Join(cells, " | ") + " |\n")
		if i == 0 {
			seps := make([]string, len(spans))
			for j := range seps {
				seps[j] = "---"
			}
			md.WriteString("| " + strings.Join(seps, " | ") + " |\n")
		}
	}
}

func convertPage(page pdf.Page, pageNum int) string {
	fontCache := buildFontStyleCache(page)
	spans := extractSpans(page, &fontCache)
	if len(spans) == 0 {
		return ""
	}
	lines := groupLines(spans)
	if len(lines) == 0 {
		return ""
	}
	bodySize := analyzeFontSize(lines)
	blocks := classifyBlocks(lines, bodySize)

	_ = pageNum // available for debug logging if needed

	return emitMarkdown(blocks)
}
