//TODO: snowflake should be implemented in the db like discribed here
//https://instagram-engineering.com/sharding-ids-at-instagram-1cf5a71e5a5c
//
package utils

import "github.com/bwmarrin/snowflake"

var generator *snowflake.Node

func InitSnowflakeNode(node int64) error {
	g, err := snowflake.NewNode(node)
	if err == nil {
		generator = g
	}
	return err
}

func GetNextSnowflakeID() int64 {
	return generator.Generate().Int64()
}
