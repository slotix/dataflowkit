package parser

import (
	"fmt"

	"log"

	"github.com/spf13/viper"
	"github.com/tealeg/xlsx"
)

var file *xlsx.File

//SaveExcel stores harvested data to Excel file
func (out Out) SaveExcel(fName string) ([]byte, error) {
	var sheet *xlsx.Sheet
	var err error
	file = xlsx.NewFile()
	for _, o := range out.Element {
		sheet, err = file.AddSheet(o.meta.Name)
		if err != nil {
			fmt.Printf(err.Error())
		}
		o.marshalExcelSheet(sheet)
	}
	err = file.Save(fName)
	if err != nil {
		fmt.Printf(err.Error())
	}
	return nil, nil
}

func (o outItem) marshalExcelSheet(sheet *xlsx.Sheet) {
	buf := o.generateTable()
	var row *xlsx.Row
	for _, item := range buf {
		row = sheet.AddRow()
		count := row.WriteSlice(&item, -1)
		if count == -1 {
			log.Println(viper.GetString("errors.errWriteRow"))
		}
	}
}
