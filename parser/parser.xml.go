package parser

//modified lib
//https://blog.diggernaut.com/json-to-xml-or-transform-in-6-seconds/
//-> MongoDB and convert it toXML
import (
	"bufio"
	"os"

	"github.com/Diggernaut/mxj"
)

//original lib
//import "github.com/clbanning/mxj"

//MarshalXML Marshales harvested data as XML
func (cols *Collections) MarshalXML() ([]byte, error) {
	mxj.XMLEscapeChars(true)
	m, err := mxj.NewMapStruct(cols)
	if err != nil {
		return nil, err
	}
	b, err := m.XmlIndent("", "  ", "Collections")

	if err != nil {
		return nil, err
	}
	return b, nil
}

func (cols *Collections) saveXML(fName string) error {
	f, err := os.Create(fName)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := cols.MarshalXML()
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
