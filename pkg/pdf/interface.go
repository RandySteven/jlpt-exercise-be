package pdf_client

type (
	PDFClient interface {
		GeneratePDF(templateFilePath string, data map[string]interface{}) (pdfBytes []byte, err error)
	}
)
