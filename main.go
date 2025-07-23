package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ansrivas/pdftool/internal"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pdf-tool",
	Short: "A CLI tool for PDF compression and image-to-PDF conversion",
	Long: `A comprehensive tool for compressing PDF files and converting images (PNG/JPEG) to PDF format.

For best compression results, install Ghostscript:
  - Linux: sudo apt install ghostscript  
  - macOS: brew install ghostscript
  - Windows: Download from ghostscript.com`,
}

var compressCmd = &cobra.Command{
	Use:   "compress [input.pdf] [output.pdf] [quality%]",
	Short: "Compress a PDF file",
	Long: `Compress a PDF file with specified quality percentage.

Quality levels:
  1-25:   Maximum compression, lowest quality (/screen preset)
  26-50:  High compression, medium-low quality (/ebook preset) 
  51-75:  Medium compression, good quality (/printer preset)
  76-100: Light compression, highest quality (/prepress preset)`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		outputFile := args[1]
		qualityStr := args[2]

		quality, err := strconv.Atoi(qualityStr)
		if err != nil {
			return fmt.Errorf("invalid quality percentage: %s (must be 1-100)", qualityStr)
		}

		if quality < 1 || quality > 100 {
			return fmt.Errorf("quality must be between 1 and 100, got: %d", quality)
		}

		// Check if files are the same
		if inputFile == outputFile {
			return fmt.Errorf("input and output files cannot be the same")
		}

		fmt.Printf("🔄 Compressing PDF: %s -> %s (Quality: %d%%)\n", inputFile, outputFile, quality)

		if err := internal.CompressPDF(inputFile, outputFile, quality); err != nil {
			return fmt.Errorf("compression failed: %w", err)
		}

		fmt.Println("✅ PDF compression completed successfully!")
		return nil
	},
}

var convertCmd = &cobra.Command{
	Use:   "convert [input.png/jpg] [output.pdf]",
	Short: "Convert PNG or JPEG to PDF",
	Long:  `Convert PNG or JPEG image files to PDF format with automatic sizing`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := args[0]
		outputFile := args[1]

		fmt.Printf("🔄 Converting image: %s -> %s\n", inputFile, outputFile)

		if err := internal.ConvertImageToPDF(inputFile, outputFile); err != nil {
			return fmt.Errorf("conversion failed: %w", err)
		}

		fmt.Println("✅ Image to PDF conversion completed successfully!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(compressCmd)
	rootCmd.AddCommand(convertCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
}
