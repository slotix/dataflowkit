package storage

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func Test_diskv(t *testing.T) {
	viper.Set("ITEM_EXPIRE_IN", 3600)
	d := newDiskvConn("", 1024*1024)
	testValue := []byte("testValue")
	testKey := "testKey"
	rec := Record{
		Key:     testKey,
		Value:   testValue,
		ExpTime: 100,
	}

	err := d.Write(rec)
	assert.NoError(t, err, "Expected no error")

	value, err := d.Read(Record{
		Type: rec.Type,
		Key:  rec.Key,
	})
	assert.Equal(t, testValue, value, "Expected equal")

	//Write record with Empty key
	recEmptyKey := Record{
		Key:     "",
		Value:   testValue,
		ExpTime: 0,
	}
	err = d.Write(recEmptyKey)
	assert.Error(t, err, "Expected empty key error")

	nonExistentRec := Record{
		Key: "NonExistent key",
	}
	// Read NonExistent key
	value, err = d.Read(nonExistentRec)
	assert.Error(t, err, "Expected error")

	expired := d.Expired(rec)
	assert.Equal(t, expired, false, "Expected non expired value")

	expired = d.Expired(nonExistentRec)
	assert.Equal(t, expired, true, "Expected true for non NonExistentKey")

	err = d.Delete(rec)
	assert.NoError(t, err, "Expected no error")
	err = d.Delete(nonExistentRec)
	assert.Error(t, err, "Expected error")

	//Add two values to storage
	err = d.Write(rec)

	err = d.Write(Record{
		Key:     "OneMoreKey",
		Value:   []byte("OneMoreValue"),
		ExpTime: 100,
	})
	//erase all items
	err = d.DeleteAll()
	assert.NoError(t, err, "Expected no error")
	d.Close()
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
