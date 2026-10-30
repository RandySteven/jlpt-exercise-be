package pdf_client

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	wkhtml "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/google/uuid"
)

type (
	pdfClient struct {
		generator *wkhtml.PDFGenerator
	}
)

func NewPDFClient() (PDFClient, error) {
	generator, err := wkhtml.NewPDFGenerator()
	if err != nil {
		return nil, err
	}
	return &pdfClient{
		generator: generator,
	}, nil
}

func (c *pdfClient) GeneratePDF(templateFilePath string, data map[string]interface{}) (pdfBytes []byte, err error) {
	html, err := template.ParseFiles(templateFilePath)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	if err := html.Execute(&buffer, data); err != nil {
		return nil, err
	}

	c.generator.AddPage(wkhtml.NewPageReader(strings.NewReader(buffer.String())))

	if err := c.generator.Create(); err != nil {
		return nil, err
	}

	outputFilePath := fmt.Sprintf("%s.pdf", uuid.New().String())

	if err := c.generator.WriteFile(outputFilePath); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
