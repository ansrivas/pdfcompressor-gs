package internal

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// CompressPDF compresses a PDF file with the specified quality percentage
func CompressPDF(inputFile, outputFile string, quality int) error {
	// Check if input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputFile)
	}

	// Try Ghostscript first (most effective)
	if isGhostscriptAvailable() {
		fmt.Println("Using Ghostscript for compression...")
		return compressWithGhostscript(inputFile, outputFile, quality)
	}

	// Fallback to pdfcpu (basic optimization)
	fmt.Println("Ghostscript not found, using pdfcpu for basic optimization...")
	return compressWithPdfcpu(inputFile, outputFile, quality)
}

// isGhostscriptAvailable checks if Ghostscript is installed
func isGhostscriptAvailable() bool {
	cmd := "gs"
	if runtime.GOOS == "windows" {
		cmd = "gswin64c" // Try 64-bit version first
	}

	_, err := exec.LookPath(cmd)
	if err != nil && runtime.GOOS == "windows" {
		// Try 32-bit version on Windows
		cmd = "gswin32c"
		_, err = exec.LookPath(cmd)
	}

	return err == nil
}

// compressWithGhostscript uses Ghostscript for effective PDF compression
func compressWithGhostscript(inputFile, outputFile string, quality int) error {
	// Determine Ghostscript command
	cmd := "gs"
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("gswin64c"); err == nil {
			cmd = "gswin64c"
		} else {
			cmd = "gswin32c"
		}
	}

	// Get quality settings based on percentage
	pdfSettings, imageRes := getGhostscriptSettings(quality)

	// Build Ghostscript command
	args := []string{
		"-q",                                  // Quiet mode
		"-dNOPAUSE",                           // Don't pause between pages
		"-dBATCH",                             // Exit after processing
		"-dSAFER",                             // Restrict file operations
		"-sDEVICE=pdfwrite",                   // Output device
		"-dCompatibilityLevel=1.4",            // PDF version
		"-dPDFSETTINGS=" + pdfSettings,        // Compression preset
		"-dEmbedAllFonts=true",                // Embed fonts
		"-dSubsetFonts=true",                  // Subset fonts
		"-dColorImageDownsampleType=/Bicubic", // Color image resampling
		"-dColorImageResolution=" + fmt.Sprintf("%d", imageRes),
		"-dGrayImageDownsampleType=/Bicubic", // Grayscale image resampling
		"-dGrayImageResolution=" + fmt.Sprintf("%d", imageRes),
		"-dMonoImageDownsampleType=/Bicubic", // Monochrome image resampling
		"-dMonoImageResolution=" + fmt.Sprintf("%d", imageRes),
		"-sOutputFile=" + outputFile, // Output file
		inputFile,                    // Input file
	}

	// Execute Ghostscript
	gsCmd := exec.Command(cmd, args...)
	gsCmd.Stderr = os.Stderr

	if err := gsCmd.Run(); err != nil {
		return fmt.Errorf("ghostscript compression failed: %w", err)
	}

	return reportCompressionStats(inputFile, outputFile)
}

// getGhostscriptSettings returns appropriate settings based on quality percentage
func getGhostscriptSettings(quality int) (string, int) {
	switch {
	case quality <= 25:
		return "/screen", 72 // Maximum compression, lowest quality
	case quality <= 50:
		return "/ebook", 150 // High compression, medium-low quality
	case quality <= 75:
		return "/printer", 300 // Medium compression, good quality
	default:
		return "/prepress", 300 // Light compression, highest quality
	}
}

// compressWithPdfcpu provides basic PDF optimization using pdfcpu
func compressWithPdfcpu(inputFile, outputFile string, quality int) error {
	config := model.NewDefaultConfiguration()
	config.ValidationMode = model.ValidationRelaxed

	// Enable compression features based on quality
	if quality < 50 {
		config.WriteObjectStream = true
		config.WriteXRefStream = true
	} else if quality < 80 {
		config.WriteObjectStream = true
	}

	if err := api.OptimizeFile(inputFile, outputFile, config); err != nil {
		return fmt.Errorf("pdfcpu optimization failed: %w", err)
	}

	return reportCompressionStats(inputFile, outputFile)
}

// reportCompressionStats reports compression statistics
func reportCompressionStats(inputFile, outputFile string) error {
	inputInfo, err := os.Stat(inputFile)
	if err != nil {
		return fmt.Errorf("failed to get input file info: %w", err)
	}

	outputInfo, err := os.Stat(outputFile)
	if err != nil {
		return fmt.Errorf("failed to get output file info: %w", err)
	}

	inputSize := inputInfo.Size()
	outputSize := outputInfo.Size()

	if inputSize > 0 {
		compressionRatio := float64(outputSize) / float64(inputSize) * 100
		savings := float64(inputSize-outputSize) / float64(inputSize) * 100

		fmt.Printf("\n📊 Compression Results:\n")
		fmt.Printf("   Original size: %.2f KB (%.2f MB)\n",
			float64(inputSize)/1024, float64(inputSize)/(1024*1024))
		fmt.Printf("   Compressed size: %.2f KB (%.2f MB)\n",
			float64(outputSize)/1024, float64(outputSize)/(1024*1024))
		fmt.Printf("   Final size: %.1f%% of original\n", compressionRatio)
		fmt.Printf("   Space saved: %.1f%%\n", savings)

		if outputSize >= inputSize {
			fmt.Printf("   ⚠️  Note: Output file is not smaller than input\n")
		}
	}

	return nil
}
