package csvutil

import (
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
