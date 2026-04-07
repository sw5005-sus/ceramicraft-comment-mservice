package dao

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/sw5005-sus/ceramicraft-comment-mservice/server/log"
	myMongo "github.com/sw5005-sus/ceramicraft-comment-mservice/server/repository/dao/mongo"
	myRedis "github.com/sw5005-sus/ceramicraft-comment-mservice/server/repository/dao/redis"
	"github.com/sw5005-sus/ceramicraft-comment-mservice/server/repository/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	options "go.mongodb.org/mongo-driver/mongo/options"
)

type CommentDao interface {
	Save(ctx context.Context, comment *model.Comment) error
	Get(ctx context.Context, id string) (*model.Comment, error)
	Delete(ctx context.Context, id string) (err error)
	HIncr(ctx context.Context, key string, member string, deta int) (err error)
	SAdd(ctx context.Context, key string, member string) (err error)
	GetListByUserID(ctx context.Context, userID int) (list []*model.Comment, err error)
	GetListByProductID(ctx context.Context, productId int) (list []*model.Comment, err error)
	GetListByQuery(ctx context.Context, productId int, stars int) (list []*model.Comment, err error)
	GetListByStatus(ctx context.Context, status string) (list []*model.Comment, err error)
	HMGet(ctx context.Context, key string, members []string) (likesCntMap map[string]int, err error)
	SMembers(ctx context.Context, key string) (likedReviewIds []string, err error)
	HGet(ctx context.Context, key string, member string) (value string, err error)
	HDel(ctx context.Context, key string, member string) (err error)
	HSet(ctx context.Context, key string, member string, value string) (err error)
	UpdateIsPinnedByID(ctx context.Context, id string, isPinned bool) error
	UpdateReviewByID(ctx context.Context, id string, updates map[string]interface{}) error
}

var (
	commentDaoInstance CommentDao
	commentSyncOnce    sync.Once
)

func GetCommentDao() CommentDao {
	commentSyncOnce.Do(func() {
		commentDaoInstance = &CommentDaoImpl{
			collection:  myMongo.CommentCollection,
			redisClient: myRedis.RedisClient,
		}
	})
	return commentDaoInstance
}

type CommentDaoImpl struct {
	collection  *mongo.Collection
	redisClient *redis.Client
}

// Get implements CommentDao.
func (c *CommentDaoImpl) Get(ctx context.Context, id string) (*model.Comment, error) {
	var returnComment model.Comment
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Logger.Errorf("parse id failed.\terr=%v", err)
		return nil, err
	}
	err = c.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&returnComment)
	if err != nil {
		log.Logger.Errorf("failed to get comment by id %s: %v", id, err)
		return nil, err
	}
	return &returnComment, nil
}

// Save implements CommentDao.
func (c *CommentDaoImpl) Save(ctx context.Context, comment *model.Comment) error {
	ret, err := c.collection.InsertOne(ctx, comment)
	if err != nil {
		log.Logger.Errorf("failed to save comment: %v", err)
		return err
	}
	// 将 InsertedID 转换为字符串
	objectID, ok := ret.InsertedID.(primitive.ObjectID)
	if !ok {
		log.Logger.Fatalf("InsertedID is not of type primitive.ObjectID")
	} else {
		comment.ID = objectID.Hex()
	}
	log.Logger.Infof("comment saved with id: %v", ret.InsertedID)
	return nil
}

func (c *CommentDaoImpl) HIncr(ctx context.Context, key string, member string, deta int) (err error) {
	if c.redisClient == nil {
		log.Logger.Errorf("redis client is nil")
		return nil
	}
	cmd := c.redisClient.HIncrBy(ctx, key, member, int64(deta))
	if cmd.Err() != nil {
		log.Logger.Errorf("HIncr failed\tkey=%s\tmember=%s\tdeta=%d\terr=%v", key, member, deta, cmd.Err())
		return cmd.Err()
	}
	return nil
}

// Delete removes a comment document by its hex id
func (c *CommentDaoImpl) Delete(ctx context.Context, id string) (err error) {
	if c.collection == nil {
		log.Logger.Errorf("mongo collection is nil")
		return nil
	}
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Logger.Errorf("parse id failed. id=%s err=%v", id, err)
		return err
	}
	_, err = c.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		log.Logger.Errorf("DeleteOne failed id=%s err=%v", id, err)
		return err
	}
	return nil
}

func (c *CommentDaoImpl) SAdd(ctx context.Context, key string, member string) (err error) {
	if c.redisClient == nil {
		log.Logger.Errorf("redis client is nil")
		return nil
	}
	cmd := c.redisClient.SAdd(ctx, key, member)
	if cmd.Err() != nil {
		log.Logger.Errorf("SAdd failed\tkey=%s\tmember=%s\terr=%v", key, member, cmd.Err())
		return cmd.Err()
	}
	return nil
}

func (c *CommentDaoImpl) GetListByUserID(ctx context.Context, userID int) (list []*model.Comment, err error) {
	if c.collection == nil {
		log.Logger.Errorf("mongo collection is nil")
		return nil, nil
	}
	cursor, err := c.collection.Find(ctx, bson.M{"user_id": userID, "status": "approved"})
	if err != nil {
		log.Logger.Errorf("Find by user_id failed\tuser_id=%d\terr=%v", userID, err)
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Logger.Errorf("failed to close cursor: %v", err)
		}
	}()
	var results []*model.Comment
	for cursor.Next(ctx) {
		var cm model.Comment
		if err := cursor.Decode(&cm); err != nil {
			log.Logger.Errorf("Decode comment failed\terr=%v", err)
			return nil, err
		}
		log.Logger.Infof("cm = %+v\n", cm)
		results = append(results, &cm)
	}
	if err := cursor.Err(); err != nil {
		log.Logger.Errorf("cursor iteration error\terr=%v", err)
		return nil, err
	}
	return results, nil
}

func (c *CommentDaoImpl) GetListByProductID(ctx context.Context, productId int) (list []*model.Comment, err error) {
	if c.collection == nil {
		log.Logger.Errorf("mongo collection is nil")
		return nil, nil
	}
	cursor, err := c.collection.Find(ctx, bson.M{"product_id": productId, "status": "approved"})
	if err != nil {
		log.Logger.Errorf("Find by product_id failed\tproduct_id=%d\terr=%v", productId, err)
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Logger.Errorf("failed to close cursor: %v", err)
		}
	}()
	var results []*model.Comment
	for cursor.Next(ctx) {
		var cm model.Comment
		if err := cursor.Decode(&cm); err != nil {
			log.Logger.Errorf("Decode comment failed\terr=%v", err)
			return nil, err
		}
		results = append(results, &cm)
	}
	if err := cursor.Err(); err != nil {
		log.Logger.Errorf("cursor iteration error\terr=%v", err)
		return nil, err
	}
	return results, nil
}

// GetListByQuery returns comments for a product filtered by stars (if stars>0)
// and ordered by created_at descending.
func (c *CommentDaoImpl) GetListByQuery(ctx context.Context, productId int, stars int) (list []*model.Comment, err error) {
	if c.collection == nil {
		log.Logger.Errorf("mongo collection is nil")
		return nil, nil
	}
	// build filter
	filter := bson.M{}
	if productId > 0 {
		filter["product_id"] = productId
	}
	if stars > 0 {
		filter["stars"] = stars
	}
	filter["status"] = "approved"
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := c.collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Logger.Errorf("Find by product_id and stars failed\tproduct_id=%d\tstars=%d\terr=%v", productId, stars, err)
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Logger.Errorf("failed to close cursor: %v", err)
		}
	}()
	var results []*model.Comment
	for cursor.Next(ctx) {
		var cm model.Comment
		if err := cursor.Decode(&cm); err != nil {
			log.Logger.Errorf("Decode comment failed\terr=%v", err)
			return nil, err
		}
		results = append(results, &cm)
	}
	if err := cursor.Err(); err != nil {
		log.Logger.Errorf("cursor iteration error\terr=%v", err)
		return nil, err
	}
	return results, nil
}

// NOTE: using options.Find() directly above; no helper needed.

func (c *CommentDaoImpl) HMGet(ctx context.Context, key string, members []string) (likesCntMap map[string]int, err error) {
	likesCntMap = make(map[string]int, len(members))
	if c.redisClient == nil {
		log.Logger.Errorf("redis client is nil")
		return likesCntMap, nil
	}
	if len(members) == 0 {
		return likesCntMap, nil
	}
	// redis HMGet accepts variadic keys
	vals, err := c.redisClient.HMGet(ctx, key, members...).Result()
	if err != nil {
		log.Logger.Errorf("HMGet failed\tkey=%s\tmembers=%v\terr=%v", key, members, err)
		return nil, err
	}
	for i, v := range vals {
		member := members[i]
		if v == nil {
			likesCntMap[member] = 0
			continue
		}
		// v can be string or []byte
		var s string
		switch t := v.(type) {
		case string:
			s = t
		case []byte:
			s = string(t)
		default:
			// fallback to sprint
			s = fmt.Sprint(v)
		}
		// parse int
		cnt, perr := strconv.Atoi(s)
		if perr != nil {
			log.Logger.Errorf("parse HMGet value failed\tkey=%s\tmember=%s\tvalue=%v\terr=%v", key, member, v, perr)
			likesCntMap[member] = 0
			continue
		}
		likesCntMap[member] = cnt
	}
	return likesCntMap, nil
}

func (c *CommentDaoImpl) SMembers(ctx context.Context, key string) (likedReviewIds []string, err error) {
	if c.redisClient == nil {
		log.Logger.Errorf("redis client is nil")
		return nil, nil
	}
	vals, err := c.redisClient.SMembers(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		log.Logger.Errorf("SMembers failed\tkey=%s\terr=%v", key, err)
		return nil, err
	}
	return vals, nil
}

func (c *CommentDaoImpl) HGet(ctx context.Context, key string, member string) (value string, err error) {
	if c.redisClient == nil {
		log.Logger.Errorf("redis client is nil")
		return "", nil
	}
	cmd := c.redisClient.HGet(ctx, key, member)
	if cmd.Err() != nil {
		// if key/member not exist, HGet returns redis.Nil; propagate that as empty value with nil error or return error?
		// Follow existing pattern: log and return error
		if cmd.Err() == redis.Nil {
			return "", nil
		}
		log.Logger.Errorf("HGet failed\tkey=%s\tmember=%s\terr=%v", key, member, cmd.Err())
		return "", cmd.Err()
	}
	val, err := cmd.Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		log.Logger.Errorf("HGet result failed\tkey=%s\tmember=%s\terr=%v", key, member, err)
		return "", err
	}
	return val, nil
}

func (c *CommentDaoImpl) HDel(ctx context.Context, key string, member string) (err error) {
	if c.redisClient == nil {
		log.Logger.Errorf("redis client is nil")
		return nil
	}
	cmd := c.redisClient.HDel(ctx, key, member)
	if cmd.Err() != nil {
		log.Logger.Errorf("HDel failed\tkey=%s\tmember=%s\terr=%v", key, member, cmd.Err())
		return cmd.Err()
	}
	return nil
}

func (c *CommentDaoImpl) HSet(ctx context.Context, key string, member string, value string) (err error) {
	if c.redisClient == nil {
		log.Logger.Errorf("redis client is nil")
		return nil
	}
	cmd := c.redisClient.HSet(ctx, key, member, value)
	if cmd.Err() != nil {
		log.Logger.Errorf("HSet failed\tkey=%s\tmember=%s\tvalue=%s\terr=%v", key, member, value, cmd.Err())
		return cmd.Err()
	}
	return nil
}

func (c *CommentDaoImpl) UpdateIsPinnedByID(ctx context.Context, id string, isPinned bool) error {
	if c.collection == nil {
		log.Logger.Errorf("mongo collection is nil")
		return nil
	}
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Logger.Errorf("parse id failed. id=%s err=%v", id, err)
		return err
	}
	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"is_pinned": isPinned}}

	_, err = c.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Logger.Errorf("UpdateOne failed id=%s err=%v", id, err)
		return err
	}
	return nil
}

func (c *CommentDaoImpl) UpdateReviewByID(ctx context.Context, id string, updates map[string]interface{}) error {
	if c.collection == nil {
		log.Logger.Errorf("mongo collection is nil")
		return nil
	}
	if len(updates) == 0 {
		return nil
	}
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Logger.Errorf("parse id failed. id=%s err=%v", id, err)
		return err
	}
	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": updates}

	_, err = c.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Logger.Errorf("UpdateOne failed id=%s err=%v", id, err)
		return err
	}
	return nil
}

func (c *CommentDaoImpl) GetListByStatus(ctx context.Context, status string) (list []*model.Comment, err error) {
	if c.collection == nil {
		log.Logger.Errorf("mongo collection is nil")
		return nil, nil
	}
	
	filter := bson.M{"status": status}
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := c.collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Logger.Errorf("Find by status failed\tstatus=%s\terr=%v", status, err)
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Logger.Errorf("failed to close cursor: %v", err)
		}
	}()
	
	var results []*model.Comment
	for cursor.Next(ctx) {
		var cm model.Comment
		if err := cursor.Decode(&cm); err != nil {
			log.Logger.Errorf("Decode comment failed\terr=%v", err)
			return nil, err
		}
		results = append(results, &cm)
	}
	if err := cursor.Err(); err != nil {
		log.Logger.Errorf("cursor iteration error\terr=%v", err)
		return nil, err
	}
	
	return results, nil
}
