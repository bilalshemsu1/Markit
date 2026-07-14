package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ledongthuc/pdf"
)

type ConversionResult struct {
	Success  bool   `json:"success"`
	Markdown string `json:"markdown"`
	Error    string `json:"error"`
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

// pdf converter

func convertPDFFromPath(filePath string) ConversionResult {
	f, reader, err := pdf.Open(filePath)
	if err != nil {
		return ConversionResult{Success: false, Error: "Cannot open PDF: " + err.Error()}
	}
	defer f.Close()

	var md strings.Builder
	totalPages := reader.NumPage()

	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		page := reader.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}

		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		if totalPages > 1 {
			md.WriteString("<!-- Page " + strconv.Itoa(pageNum) + " -->\n\n")
		}

		lines := strings.Split(text, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				md.WriteString("\n")
			} else {
				md.WriteString(line + "\n")
			}
		}
		md.WriteString("\n")
	}

	result := strings.TrimSpace(md.String())
	if result == "" {
		return ConversionResult{Success: false, Error: "No text content found in PDF. The file might be a scanned/image-based PDF."}
	}

	return ConversionResult{
		Success:  true,
		Markdown: result,
	}
}

func convertPDF(data []byte) ConversionResult {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return ConversionResult{Success: false, Error: "Cannot parse PDF: " + err.Error()}
	}

	var md strings.Builder
	totalPages := reader.NumPage()

	for pageNum := 1; pageNum <= totalPages; pageNum++ {
		page := reader.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}

		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}

		if totalPages > 1 {
			md.WriteString("<!-- Page " + strconv.Itoa(pageNum) + " -->\n\n")
		}

		lines := strings.Split(text, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				md.WriteString("\n")
			} else {
				md.WriteString(line + "\n")
			}
		}
		md.WriteString("\n")
	}

	result := strings.TrimSpace(md.String())
	if result == "" {
		return ConversionResult{Success: false, Error: "No text content found in PDF. The file might be a scanned/image-based PDF."}
	}

	return ConversionResult{
		Success:  true,
		Markdown: result,
	}
}
