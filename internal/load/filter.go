package load

import (
	"encoding/json"
	"errors"
	"io"
	"os"
)

type FundedByFilter struct {
	Name    string    `json:"name"`
	Amounts []float64 `json:"amounts"`
}

var fundedByFilters map[string]FundedByFilter

func FindFundedByFilter(adress string, amount float64) string {
	if fundedByFilters == nil {
		return ""
	}

	if filter, ok := fundedByFilters[adress]; ok {
		if len(filter.Amounts) == 0 {
			for _, filterAmount := range filter.Amounts {
				if filterAmount == amount {
					return filter.Name
				}
			}
		}
	}

	return ""
}

func LoadFundedByFilters() error {
	if fundedByFilters != nil {
		return errors.New("fundedby filters already loaded")
	}

	// Check if fundedby_filter.json exists
	if _, err := os.Stat("fundedby_filter.json"); os.IsNotExist(err) {
		// File does not exist
		return err
	}

	// Load the file
	file, err := os.Open("fundedby_filter.json")
	if err != nil {
		return err
	}

	// Close the file
	defer file.Close()

	// Read the file
	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &fundedByFilters)
	if err != nil {
		return err
	}

	return nil
}
