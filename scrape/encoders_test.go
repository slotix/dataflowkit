package scrape

import (
	"bytes"
	"reflect"
	"testing"
)

func TestJSONEncoder_Encode(t *testing.T) {
	type fields struct {
		paginateResults bool
	}
	type args struct {
		results *Results
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{name: "paginated Results = false",
			fields: fields{
				paginateResults: false,
			},
			args: args{
				results: &Results{
					Output: [][]map[string]interface{}{
						{{"col1": "1", "col2": "2"}},
						{{"col1": "3", "col2": "4"}},
					},
				},
			},
			want:    []byte(`[{"col1":"1","col2":"2"},{"col1":"3","col2":"4"}]` + "\n"),
			wantErr: false,
		},
		
		{name: "paginated Results = true",
			fields: fields{
				paginateResults: true,
			},
			args: args{
				results: &Results{
					Output: [][]map[string]interface{}{
						{{"col1": "1", "col2": "2"}},
						{{"col1": "3", "col2": "4"}},
					},
				},
			},
			want:    []byte(`[[{"col1":"1","col2":"2"}],[{"col1":"3","col2":"4"}]]` + "\n"),
			wantErr: false,
		},
		{name: "complex structure",
			args: args{
				results: &Results{
					Output: [][]map[string]interface{}{
						{{"col1": 1, "col2": 2.345}},
						{{"col1": "3", "col2": "4"}},
						{{"col1": 5, "col2": "6"}},
						{{"col1": "", "col2": 7}},
						{{"col1": []string{"8", "9"}, "col2": 10}},
						{{"col1": []int{11, 12}, "col2": []float64{13.145, 15.16}}},
					},
				},
			},
			want:    []byte(`[{"col1":1,"col2":2.345},{"col1":"3","col2":"4"},{"col1":5,"col2":"6"},{"col1":"","col2":7},{"col1":["8","9"],"col2":10},{"col1":[11,12],"col2":[13.145,15.16]}]` + "\n"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := JSONEncoder{
				paginateResults: tt.fields.paginateResults,
			}
			got, err := e.Encode(tt.args.results)
			buf := new(bytes.Buffer)
			buf.ReadFrom(got)
			t.Log(buf.String())
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONEncoder.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(buf.Bytes(), tt.want) {
				t.Errorf("JSONEncoder.Encode() = \n%v\n, want\n %v", buf.Bytes(), tt.want)
			}
		})
	}
}

func TestCSVEncoder_Encode(t *testing.T) {
	type fields struct {
		partNames []string
		comma     string
	}
	type args struct {
		results *Results
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{name: "1",
			fields: fields{
				partNames: []string{"col1", "col2"},
				comma:     ",",
			},
			args: args{
				results: &Results{
					Output: [][]map[string]interface{}{
						{{"col1": 1, "col2": 2.345}},
						{{"col1": "3", "col2": "4"}},
						{{"col1": 5, "col2": "6"}},
						{{"col1": "", "col2": 7}},
						{{"col1": []string{"8", "9"}, "col2": 10}},
						{{"col1": []int{11, 12}, "col2": []float64{13.145, 15.16}}},
						{{"invalidcol1": "111", "invalidcol2": 000}},
					},
				},
			},
			want:    []byte("col1,col2\n1,2.345\n3,4\n5,6\n,7\n8;9,10\n11;12,13.145;15.16\n,\n"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := CSVEncoder{
				partNames: tt.fields.partNames,
				comma:     tt.fields.comma,
			}
			got, err := e.Encode(tt.args.results)

			buf := new(bytes.Buffer)
			buf.ReadFrom(got)
			t.Log(buf.String())
			t.Log(string(tt.want))

			if (err != nil) != tt.wantErr {
				t.Errorf("CSVEncoder.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(buf.Bytes(), tt.want) {
				t.Errorf("CSVEncoder.Encode() = %v, want %v", buf.Bytes(), tt.want)
			}
		})
	}
}

func TestXMLEncoder_Encode(t *testing.T) {
	type args struct {
		results *Results
	}
	tests := []struct {
		name    string
		e       XMLEncoder
		args    args
		want    []byte
		wantErr bool
	}{
		 {name: "1",
			args: args{
				results: &Results{
					Output: [][]map[string]interface{}{
						{{"col1": 1, "col2": 2.345}},
						{{"col1": "3", "col2": "4"}},
						{{"col1": 5, "col2": "6"}},
						{{"col1": "", "col2": 7}},
						{{"col1": []string{"8", "9"}, "col2": 10}},	
						{{"col1": "11,12", "col2": 13}},		
					},
				},
			},
			want:    []byte(`<?xml version="1.0" encoding="UTF-8"?><root><element><col1>1</col1><col2>2.345</col2></element><element><col1>3</col1><col2>4</col2></element><element><col1>5</col1><col2>6</col2></element><element><col1/><col2>7</col2></element><element><col1>8</col1><col1>9</col1><col2>10</col2></element><element><col1>11,12</col1><col2>13</col2></element></root>`),
			wantErr: false,
		}, 
		{name: "2",
			args: args{
				results: &Results{
					Output: [][]map[string]interface{}{
						{{"col1": []int{11, 12}, "col2": []float64{13.145, 15.16}}},		
					},
				},
			},
			want:    []byte(`<?xml version="1.0" encoding="UTF-8"?><root><element><col1>11</col1><col1>12</col1><col2>13.145</col2><col2>15.16</col2></element></root>`),
			wantErr: false,
		},
	}
	//
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := XMLEncoder{}
			got, err := e.Encode(tt.args.results)

			buf := new(bytes.Buffer)
			buf.ReadFrom(got)
			//t.Log(buf.String())
			//t.Log(string(tt.want))

			if (err != nil) != tt.wantErr {
				t.Errorf("XMLEncoder.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(buf.Bytes(), tt.want) {
				t.Errorf("XMLEncoder.Encode() = %v, want %v", buf.Bytes(), tt.want)
			}
		})
	}
}


