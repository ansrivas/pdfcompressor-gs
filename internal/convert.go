package internal

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/jung-kurt/gofpdf"
)

// ConvertImageToPDF converts PNG or JPEG image to PDF
func ConvertImageToPDF(inputFile, outputFile string) error {
	// Check if input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputFile)
	}

	// Get file extension
	ext := strings.ToLower(filepath.Ext(inputFile))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		return fmt.Errorf("unsupported file format: %s (supported: .png, .jpg, .jpeg)", ext)
	}

	// Open and decode image
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	var img image.Image
	switch ext {
	case ".png":
		img, err = png.Decode(file)
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	}
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Get image dimensions
	bounds := img.Bounds()
	width := float64(bounds.Dx())
	height := float64(bounds.Dy())

	// Calculate PDF dimensions (convert pixels to points, assuming 72 DPI)
	pdfWidth := width * 72 / 300 // Assuming 300 DPI image
	pdfHeight := height * 72 / 300

	// Handle large images by scaling down if necessary
	const maxSize = 500 // Maximum dimension in points
	if pdfWidth > maxSize || pdfHeight > maxSize {
		if pdfWidth > pdfHeight {
			pdfHeight = pdfHeight * maxSize / pdfWidth
			pdfWidth = maxSize
		} else {
			pdfWidth = pdfWidth * maxSize / pdfHeight
			pdfHeight = maxSize
		}
	}

	// Create PDF
	pdf := gofpdf.New("P", "pt", "A4", "")
	pdf.AddPage()

	// Create temporary image file for PDF embedding
	tempImageFile := "temp_image_for_pdf" + ext
	defer os.Remove(tempImageFile)

	// Resize image if needed and save to temporary file
	resizedImg := imaging.Resize(img, int(width), int(height), imaging.Lanczos)
	if err := saveImage(resizedImg, tempImageFile, ext); err != nil {
		return fmt.Errorf("failed to save temporary image: %w", err)
	}

	// Add image to PDF
	imageType := "JPG"
	if ext == ".png" {
		imageType = "PNG"
	}

	// Center the image on the page
	pageWidth, pageHeight := pdf.GetPageSize()
	x := (pageWidth - pdfWidth) / 2
	y := (pageHeight - pdfHeight) / 2

	pdf.ImageOptions(tempImageFile, x, y, pdfWidth, pdfHeight, false,
		gofpdf.ImageOptions{ImageType: imageType, ReadDpi: true}, 0, "")

	// Save PDF
	if err := pdf.OutputFileAndClose(outputFile); err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}

	fmt.Printf("Successfully converted %s to %s\n", inputFile, outputFile)
	return nil
}

// saveImage saves an image to a file with the specified format
func saveImage(img image.Image, filename, format string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	switch format {
	case ".png":
		return png.Encode(file, img)
	case ".jpg", ".jpeg":
		return jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}
