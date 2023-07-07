package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"root/root/config"
	"strconv"

	"github.com/unidoc/unipdf/v3/common"
	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/model"
)

func main() {
	licensePDF()
	getData()
}

func licensePDF() {
	content, err := ioutil.ReadFile(config.APIKEY)
	if err != nil {
		log.Fatal(err)
	}

	err = license.SetMeteredKey(string(content))
	if err != nil {
		fmt.Printf("ERROR: Failed to set metered key: %v\n", err)
		panic(err)
	}
}

func getData() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go input.pdf output_directory count pages")
		return
	}

	// Args from fiels
	inputFile := os.Args[1]
	outputDir := os.Args[2]
	countPages := os.Args[3]

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.Mkdir(outputDir, 0755)
	}

	newCountPages, _ := strconv.Atoi(countPages)

	err := extractPages(inputFile, outputDir, newCountPages)
	if err != nil {
		fmt.Printf("Error extracting pages: %v\n", err)
	}
}

func extractPages(inputFile, outputDir string, interval int) error {
	// Open PDF file
	f, err := os.Open(inputFile)
	if err != nil {
		return err
	}

	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return err
	}

	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return err
	}

	if isEncrypted {
		common.Log.Debug("PDF is encrypted")
		return fmt.Errorf("cannot extract pages from encrypted PDF")
	}

	// Total pages
	totalPages, err := pdfReader.GetNumPages()
	if err != nil {
		return err
	}

	if totalPages < interval {
		return fmt.Errorf("insufficient number of pages")
	}

	pageGroups := totalPages / interval

	for i := 0; i < pageGroups; i++ {
		startPage := i * interval
		endPage := (i+1)*interval - 1

		outputFile := filepath.Join(outputDir, fmt.Sprintf("insurance_%d-%d.pdf", startPage+1, endPage+1))

		outPdfWriter := model.NewPdfWriter()

		for pageNum := startPage; pageNum <= endPage; pageNum++ {
			page, err := pdfReader.GetPage(pageNum + 1) // Pages are 1-indexed.
			if err != nil {
				return err
			}

			outPdfWriter.AddPage(page)
		}

		err = outPdfWriter.WriteToFile(outputFile)
		if err != nil {
			return err
		}
	}

	fmt.Println("Pages extracted successfully!")
	return nil
}
