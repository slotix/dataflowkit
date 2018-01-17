package storage

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func Test_diskv(t *testing.T) {
	d := newDiskvConn("", 1024*1024)
	testValue := []byte("testValue")
	viper.Set("ITEM_EXPIRE_IN", 3600)
	testKey := "testKey"
	err := d.Write(testKey, testValue, 0)
	assert.NoError(t, err, "Expected no error")
	
	value, err := d.Read("NonExistent key")
	assert.Error(t, err, "Expected error")
	
	value, err = d.Read(testKey)
	assert.Equal(t, testValue, value, "Expected equal")

	expired := d.Expired(testKey)
	assert.Equal(t, expired, false, "Expected non expired value")

	err = d.Delete(testKey)
	assert.NoError(t, err, "Expected no error")
	err = d.Delete("NonExistedKey")
	assert.Error(t, err, "Expected error")

	//Add two values to storage
	err = d.Write(testKey, testValue, 0)
	err = d.Write("secondKey", []byte("testValue"), 0)
	//erase all items
	err = d.DeleteAll()
	assert.NoError(t, err, "Expected no error")
}

// func TestDiskvConn_Write1(t *testing.T) {
// 	diskvConn := newDiskvConn("", 1024*1024)
// 	type fields struct {
// 		diskv *diskv.Diskv
// 	}
// 	type args struct {
// 		key     string
// 		value   []byte
// 		expTime int64
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		wantErr bool
// 	}{
// 		{name: "qq",
// 			fields: fields{
// 				diskv: diskvConn.diskv,
// 			},
// 			args: args{
// 				key:     "testKey",
// 				value:   []byte("testValue"),
// 				expTime: 0,
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	// TODO: Add test cases.

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			d := DiskvConn{
// 				diskv: tt.fields.diskv,
// 			}
// 			if err := d.Write(tt.args.key, tt.args.value, tt.args.expTime); (err != nil) != tt.wantErr {
// 				t.Errorf("DiskvConn.Write() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
