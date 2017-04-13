package parser

import (
	"github.com/spf13/viper"
	"github.com/tealeg/xlsx"
)

var file *xlsx.File

//SaveExcel stores harvested data to Excel file
func (cols Collections) SaveExcel(fName string) ([]byte, error) {
	var sheet *xlsx.Sheet
	var err error
	file = xlsx.NewFile()
	for _, c := range cols.Collections {
		sheet, err = file.AddSheet(c.meta.Name)
		if err != nil {
			logger.Printf(err.Error())
		}
		c.marshalExcelSheet(sheet)
	}
	err = file.Save(fName)
	if err != nil {
		logger.Printf(err.Error())
	}
	return nil, nil
}

func (c collection) marshalExcelSheet(sheet *xlsx.Sheet) {
	buf := c.generateTable()
	var row *xlsx.Row
	for _, item := range buf {
		row = sheet.AddRow()
		count := row.WriteSlice(&item, -1)
		if count == -1 {
			logger.Println(viper.GetString("errors.errWriteRow"))
		}
	}
}
