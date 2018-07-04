package storage

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gocql/gocql"
)

type cassandra struct {
	cluster *gocql.ClusterConfig
	session *gocql.Session
}

const (
	readQuery                      = "SELECT value from %s WHERE key='%s'"
	writeQuery                     = "INSERT INTO %s (key, value) VALUES ('%s', '%s') USING TTL 86400"
	writeIntermediateQuery         = "INSERT INTO Intermediate (payloadHash, pageID, blockID, fields) VALUES(?, ?, ?, ?) USING TTL 86400"
	writeIntermediateMapQuery      = "INSERT INTO intermediatemaps (payloadHash, map) VALUES(?, ?) USING TTL 86400"
	readIntermediateResultQuery    = "SELECT fields FROM Intermediate WHERE payloadhash=? AND pageID=? AND blockID=?"
	readIntermediateMapQuery       = "SELECT map FROM intermediatemaps WHERE payloadhash=?"
	truncateTableQuery             = "TRUNCATE %s"
	deleteIntermediateRowQuery     = "DELETE FROM intermediate WHERE payloadhash=? AND pageID=? AND blockID=?"
	deleteIntermediateMapsRowQuery = "DELETE FROM intermediatemaps WHERE payloadhash=?"
)

func newCassandra(host string) *cassandra {
	cluster := gocql.NewCluster(host)
	cluster.Keyspace = "dfk"
	cluster.Consistency = gocql.One
	s, err := cluster.CreateSession()
	if err != nil {
		logger.Error(err)
	}
	return &cassandra{cluster: cluster, session: s}
}

// Read loads value according to the specified key from Cassandra storage.
func (c cassandra) Read(key string, recType string) (value []byte, err error) {
	if recType == INTERMEDIATE {
		return c.readIntermediate(key)
	}
	var val string
	query := fmt.Sprintf(readQuery, recType, key)
	err = c.session.Query(query).Scan(&val)
	return []byte(val), err
}

// Write stores key/ value pair along with Expiration time to Cassandra storage.
func (c cassandra) Write(key string, rec *Record, expTime int64) error {
	if rec.RecordType == INTERMEDIATE {
		return c.writeIntermediate(key, rec.Value)
	}
	query := fmt.Sprintf(writeQuery, rec.RecordType, key, string(rec.Value))
	err := c.session.Query(query).Exec()
	return err
}

func (c cassandra) writeIntermediate(key string, value []byte) error {
	//payload-page.no-block.no
	keys := strings.Split(string(key), "-")
	if len(keys) > 1 {
		var results map[string]interface{}
		err := json.Unmarshal(value, &results)
		if err != nil {
			return fmt.Errorf("Failed unmarshal parse results. %s", err.Error())
		}
		return c.session.Query(writeIntermediateQuery, keys[0], keys[1], keys[2], results).Exec()
	}
	return c.session.Query(writeIntermediateMapQuery, key, string(value)).Exec()
}

func (c cassandra) readIntermediate(key string) ([]byte, error) {
	keys := strings.Split(string(key), "-")
	if len(keys) > 1 {
		value := map[string]string{}
		err := c.session.Query(readIntermediateResultQuery, keys[0], keys[1], keys[2]).Scan(&value)
		if err != nil {
			return nil, fmt.Errorf("Failed to read key: %s. %s", keys[0], err.Error())
		}
		val, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("Failed to marshal JSON results for key: %s %s %s. %s", keys[0], keys[1], keys[2], err.Error())
		}
		return val, err
	}
	var val string
	err := c.session.Query(readIntermediateMapQuery, key).Scan(&val)
	return []byte(val), err
}

// Expired returns Expired value of specified key from Cassandra.
func (c cassandra) Expired(key string) bool {
	//There is no need implement TTL with Cassandra 'cause it implements it natively
	return true
}

//Delete deletes specified key from Cassandra storage.
func (c cassandra) Delete(key string) error {
	keys := strings.Split(key, "-")
	var err error
	if len(keys) > 1 {
		err = c.session.Query(deleteIntermediateRowQuery, keys[0], keys[1], keys[2]).Exec()
		if err != nil {
			logger.Warningf("Failed delete intermediate row: payload %s, page %s, block %s. %s", keys[0], keys[1], keys[2], err.Error())
		}
	} else {
		err = c.session.Query(deleteIntermediateMapsRowQuery, keys[0]).Exec()
		if err != nil {
			logger.Warningf("Failed delete intermediate maps row: payload %s. %s", keys[0], err.Error())
		}
	}
	return nil
}

//DeleteAll deletes everything from Cassandra storage.
func (c cassandra) DeleteAll() error {
	query := fmt.Sprintf(truncateTableQuery, "intermediate")
	err := c.session.Query(query).Exec()
	if err != nil {
		logger.Warningf("Failed truncate Intermediate table. %s", err.Error())
	}
	query = fmt.Sprintf(truncateTableQuery, "intermediatemaps")
	err = c.session.Query(query).Exec()
	if err != nil {
		logger.Warningf("Failed truncate IntermediateMaps table. %s", err.Error())
	}
	return nil
}

func (c cassandra) Close() {
	c.session.Close()
}
