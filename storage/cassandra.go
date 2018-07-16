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
	writeQuery                     = "INSERT INTO %s (key, value) VALUES ('%s', '%s') USING TTL %d"
	writeIntermediateQuery         = "INSERT INTO Intermediate (payloadHash, pageID, blockID, fields) VALUES(?, ?, ?, ?) USING TTL %d"
	writeIntermediateMapQuery      = "INSERT INTO intermediatemaps (payloadHash, map) VALUES(?, ?) USING TTL %d"
	readIntermediateResultQuery    = "SELECT fields FROM Intermediate WHERE payloadhash=? AND pageID=? AND blockID=?"
	readIntermediateMapQuery       = "SELECT map FROM intermediatemaps WHERE payloadhash=?"
	truncateTableQuery             = "TRUNCATE %s"
	deleteIntermediateRowQuery     = "DELETE FROM intermediate WHERE payloadhash=? AND pageID=? AND blockID=?"
	deleteIntermediateMapsRowQuery = "DELETE FROM intermediatemaps WHERE payloadhash=?"
	deleteCacheRowQuery            = "DELETE FROM cache WHERE key=?"
	deleteCookiesRowQuery          = "DELETE FROM cookies WHERE key=?"
	getTTLQuery                    = "SELECT TTL(%s) from %s"
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
func (c cassandra) Read(rec Record) (value []byte, err error) {
	if rec.Type == INTERMEDIATE {
		return c.readIntermediate(rec.Key)
	}
	var val string
	query := fmt.Sprintf(readQuery, rec.Type, rec.Key)
	err = c.session.Query(query).Scan(&val)
	return []byte(val), err
}

// Write stores key/ value pair along with Expiration time to Cassandra storage.
func (c cassandra) Write(rec Record) error {
	if rec.Type == INTERMEDIATE {
		return c.writeIntermediate(rec.Key, rec.Value, rec.ExpTime)
	}
	query := fmt.Sprintf(writeQuery, rec.Type, rec.Key, string(rec.Value), rec.ExpTime)
	err := c.session.Query(query).Exec()
	return err
}

func (c cassandra) writeIntermediate(key string, value []byte, expTime int64) error {
	//payload-page.no-block.no
	keys := strings.Split(string(key), "-")
	if len(keys) > 1 {
		var results map[string]interface{}
		err := json.Unmarshal(value, &results)
		if err != nil {
			return fmt.Errorf("Failed unmarshal parse results. %s", err.Error())
		}
		for k, v := range results {
			_, ok := v.([]interface{})
			if ok {
				strValue, err := json.Marshal(v)
				if err != nil {
					return fmt.Errorf("Failed to marshal %s array value. %s", key, err.Error())
				}
				results[k] = strValue
			}
		}
		q := fmt.Sprintf(writeIntermediateQuery, expTime)
		return c.session.Query(q, keys[0], keys[1], keys[2], results).Exec()
	}
	q := fmt.Sprintf(writeIntermediateMapQuery, expTime)
	return c.session.Query(q, key, string(value)).Exec()
}

func (c cassandra) readIntermediate(key string) ([]byte, error) {
	keys := strings.Split(string(key), "-")
	if len(keys) > 1 {
		value := map[string]interface{}{}
		get := map[string]string{}
		err := c.session.Query(readIntermediateResultQuery, keys[0], keys[1], keys[2]).Scan(&get)
		if err != nil {
			return nil, fmt.Errorf("Failed to read key: %s. %s", key, err.Error())
		}
		for k, v := range get {
			if string([]rune(v)[0]) == "[" {
				var strArray []string
				err := json.Unmarshal([]byte(v), &strArray)
				if err != nil {
					value[k] = v
				} else {
					value[k] = strArray
				}
			} else {
				value[k] = v
			}
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
func (c cassandra) Expired(rec Record) bool {
	//Expired = Non Existant
	return false
}

//Delete deletes specified key from Cassandra storage.
func (c cassandra) Delete(rec Record) error {
	keys := strings.Split(rec.Key, "-")
	//var err error
	switch rec.Type {
	case INTERMEDIATE:
		if len(keys) > 1 {
			_ = c.session.Query(deleteIntermediateRowQuery, keys[0], keys[1], keys[2]).Exec()
			//the code below will never be executed

			// if err != nil {
			// 	logger.Warningf("Failed delete intermediate row: payload %s, page %s, block %s. %s", keys[0], keys[1], keys[2], err.Error())
			// }
		} else {

			_ = c.session.Query(deleteIntermediateMapsRowQuery, keys[0]).Exec()

			//the code below will never be executed
			// if err != nil {
			// 	logger.Warningf("Failed delete intermediate maps row: payload %s. %s", keys[0], err.Error())
			// }
		}
	case CACHE:
		_ = c.session.Query(deleteCacheRowQuery, keys[0]).Exec()
		//the code below will never be executed
		// if err != nil {
		// 	logger.Warningf("Failed delete cache row: key %s. %s", keys[0], err.Error())
		// }

	case COOKIES:
		_ = c.session.Query(deleteCookiesRowQuery, keys[0]).Exec()
		//the code below will never be executed
		// if err != nil {
		// 	logger.Warningf("Failed delete cookies row: key %s. %s", keys[0], err.Error())
		// }
	}
	return nil
}

//DeleteAll deletes everything from Cassandra storage.
func (c cassandra) DeleteAll() error {
	query := fmt.Sprintf(truncateTableQuery, "intermediate")
	_ = c.session.Query(query).Exec()
	//No need to check the error here as the code below will never be executed
	// if err != nil {
	// 	logger.Warningf("Failed truncate Intermediate table. %s", err.Error())
	// }
	query = fmt.Sprintf(truncateTableQuery, "intermediatemaps")
	_ = c.session.Query(query).Exec()
	// if err != nil {
	// 	logger.Warningf("Failed truncate IntermediateMaps table. %s", err.Error())
	// }
	query = fmt.Sprintf(truncateTableQuery, "cache")
	_ = c.session.Query(query).Exec()
	// if err != nil {
	// 	logger.Warningf("Failed truncate cache table. %s", err.Error())
	// }
	query = fmt.Sprintf(truncateTableQuery, "cookies")
	_ = c.session.Query(query).Exec()
	// if err != nil {
	// 	logger.Warningf("Failed truncate cookies table. %s", err.Error())
	// }
	return nil
}

func (c cassandra) Close() {
	c.session.Close()
}
