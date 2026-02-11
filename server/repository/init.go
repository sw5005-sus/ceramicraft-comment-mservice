package repository

import (
	"github.com/sw5005-sus/ceramicraft-comment-mservice/server/repository/dao/mongo"
	"github.com/sw5005-sus/ceramicraft-comment-mservice/server/repository/dao/redis"
)

func Init() {
	mongo.Init()
	redis.Init()
}
