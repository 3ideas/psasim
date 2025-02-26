package csvutil

import (
	"fmt"
	"os"

	"github.com/gocarina/gocsv"
)

// ReadItems reads a CSV file and returns a slice of items of type T, the struct tag names must match the column names in the CSV file, the first row is expected to be a header row
// and all tagged fields must be exported (have a capital first letter)
func ReadItems[T any](filename string) ([]T, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	items := make([]T, 0)
	if err := gocsv.UnmarshalFile(file, &items); err != nil {
		return nil, err
	}

	return items, nil
}

func WriteCSV[T any](filename string, items []T) error {
	fh, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer fh.Close()

	err = gocsv.MarshalFile(&items, fh) // Use this to save the CSV back to the file
	if err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}
