package dao

import (
	"github.com/bwmarrin/snowflake"
)

var snowGen *snowflake.Node

func InitIDGen(nodeId int64) {
	node, err := snowflake.NewNode(nodeId)
	if err != nil {
		panic(err)
	}
	snowGen = node
}

func NextID() uint {
	return uint(snowGen.Generate().Int64())
}
