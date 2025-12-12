package identity

import (
	"compost-bin/logger"

	"github.com/bwmarrin/snowflake"
)

var alg *snowflake.Node

func GenerateId() int64 {
	return alg.Generate().Int64()
}

func GenerateString() string {
	return alg.Generate().String()
}

func init() {
	var err error
	alg, err = snowflake.NewNode(0)
	if err != nil {
		logger.Fatalf("Failed to init snowflake algorithm: %v", err)
	}
}
