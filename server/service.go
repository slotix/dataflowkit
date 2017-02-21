package server

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"

	"github.com/go-kit/kit/log"
	"github.com/slotix/dfk-parser/parser"
	"github.com/spf13/viper"

	"github.com/garyburd/redigo/redis"
)

//	"github.com/slotix/dfk-parser/parser"

var logger log.Logger

// ParseService provides operations on strings.
type ParseService interface {
	GetHTML(string) (string, error)
	MarshalData(payload []byte) (string, error)
	CheckServices()(status map[string]string)
	//	Save(payload []byte) (string, error)
}

type parseService struct{}

func (parseService) GetHTML(url string) (string, error) {
	content, err := parser.GetHTML(url)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (parseService) MarshalData(payload []byte) (string, error) {
	redisConn, err := redis.Dial(viper.GetString("redis.protocol"),
		viper.GetString("redis.address"))

	var p parser.Payloads
	err = p.UnmarshalJSON(payload)
	if err != nil {
		return "", err
	}
	if p.Format == "" {
		p.Format = "json"
	}

	payloadMD5 := generateMD5(payload)

	outRediskey := fmt.Sprintf("%s-%s", p.Format, payloadMD5)
	outRedis, err := redis.Bytes(redisConn.Do("GET", outRediskey))
	if err == nil {
		return string(outRedis), nil
	}

	//if there is no cached value in Redis
	out, err := p.Parse()
	if err != nil {
		return "", err
	}
	var b []byte
	switch p.Format {
	case "xml":
		b, err = out.MarshalXML()
	case "csv":
		b, err = out.MarshalCSV()
	default:
		b, err = out.MarshalJSON()
	}
	if err != nil {
		return "", err
	}

	redisConn.Do("SET", outRediskey, b)
	//set 1 hour before html content key expiration
	redisConn.Do("EXPIRE", outRediskey, viper.GetInt("redis.expire"))

	return string(b), nil

}

func (parseService) CheckServices() (status map[string]string) {
	return CheckServices()//, allAlive
}


// ServiceMiddleware is a chainable behavior modifier for ParseService.
type ServiceMiddleware func(ParseService) ParseService

//func generateMD5(s string) string {
func generateMD5(b []byte) []byte {
	h := md5.New()
	r := bytes.NewReader(b)
	io.Copy(h, r)
	return h.Sum(nil)
}

/*
//MarshalData marshales data from Redis
func (r redisConn) MarshalData(payload []byte) (string, error) {
	var p parser.Payloads

	err := p.UnmarshalJSON(payload)
	if err != nil {
		return "", err
	}
	if p.Format == "" {
		p.Format = "json"
	}
	payloadMD5 := generateMD5(payload)
	outRediskey := fmt.Sprintf("%s-%s", p.Format, payloadMD5)
	outRedis, err := redis.Bytes(r.conn.Do("GET", outRediskey))
	if err == nil {
		return string(outRedis), nil
	}
	return "", err
}

//SetHTML pushes html buffer to Redis
func (r redisConn) push(key string, buf []byte) error {
	reply, err := r.conn.Do("SET", key, buf)
	if err != nil {
		return err
	}
	if reply.(string) == "OK" {
		//set 1 hour 3600 before html content key expiration
		r.conn.Do("EXPIRE", key, 3600)
	}
	return nil
}

//MarshalData marshales data to different formats
func (parseService) MarshalData(payload []byte) (string, error) {

	rc := redisConn{
		protocol: "tcp",
		addr:     "localhost:6379"}
	var err error
	rc.conn, err = redis.Dial(rc.protocol, rc.addr)
	if err != nil {
		return "", fmt.Errorf("%s: %s", parser.ErrRedisDown, err.Error())
	}
	defer rc.conn.Close()

	outRedis, err := rc.MarshalData(payload)
	if err == nil {
		return outRedis, nil
	}
	//if there is no value in Redis
	var p parser.Payloads
	err = p.UnmarshalJSON(payload)
	if err != nil {
		return "", err
	}
	if p.Format == "" {
		p.Format = "json"
	}
	out, err := p.Parse()
	if err != nil {
		return "", err
	}
	var b []byte
	switch p.Format {
	case "xml":
		b, err = out.MarshalXML()
	case "csv":
		b, err = out.MarshalCSV()
	default:
		b, err = out.MarshalJSON()

	}
	if err != nil {
		return "", err
	}
	//push parsed data to Redis
	payloadMD5 := generateMD5(payload)
	outRediskey := fmt.Sprintf("%s-%s", p.Format, payloadMD5)
	err = rc.push(outRediskey, b)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
*/
/*
func (parseService) Save(payload []byte) (string, error) {
	var p parser.Payload
	err := p.UnmarshalJSON([]byte(payload))
	if err != nil {
		return "", err
	}
	if p.Format == ""{
		p.Format = "json"
	}

	out, err := p.Parse([]byte(payload))
	if err != nil {
		logger.Log(err)
	}
	var b []byte

	fName := fmt.Sprintf("/Users/dm/go/src/dataflowkit/testdata/%d.%s",
			time.Now().UnixNano(), p.Format)
	switch p.Format {
	case "xml":
		err = out.SaveXML(fName)
	case "csv":
		b, err = out.MarshalCSV()
	default:
		err = out.SaveJSON(fName)

	}
	if err != nil {
		return "", err
	}
	return fName, nil
}
*/
