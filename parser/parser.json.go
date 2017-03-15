package parser

import (
	"bufio"
	"os"
)

func (cols *Collections) saveJSON(fName string) error {
	f, err := os.Create(fName)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := cols.MarshalJSON()
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	_, err = w.Write(b)
	if err != nil {
		return err
	}
	w.Flush()
	return nil
}
