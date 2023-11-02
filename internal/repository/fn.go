package repository

import (
	"fmt"
	"log"
	"strings"

	"github.com/xuri/excelize/v2"
)

func (r *Repo) readFile(fName string) ([][]string, error) {
	f, err := excelize.OpenFile(fName)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Println(err)
		}
	}()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("empty sheetlist in file %s", fName)
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, err
	}

	data := make([][]string, 0, 100)
	for _, row := range rows {
		line := make([]string, 0, 10)
		for _, colCell := range row {
			line = append(line, strings.TrimSpace(colCell))
		}
		data = append(data, line)
	}

	return data, nil
}
