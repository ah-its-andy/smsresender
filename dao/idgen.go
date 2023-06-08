package dao

import (
	"log"

	"github.com/bwmarrin/snowflake"
)

var snowGen *snowflake.Node

func InitIDGen(nodeId int64) {
	node, err := snowflake.NewNode(nodeId)
	if err != nil {
		log.Panic("Failed to create snowflake node: %v", err)
	}
	snowGen = node
}

func NextID() uint {
	return uint(snowGen.Generate().Int64())
}
