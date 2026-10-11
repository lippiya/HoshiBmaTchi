package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

var c = context.Background()

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient() *RedisClient{
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", 
		DB:       0,  
	})
	return &RedisClient{client: client}
}

func (r *RedisClient) SetOTP(email string, code string) error{
	return  r.client.Set(c, "otp:"+email, code, 5 * time.Minute).Err()
}

func (r *RedisClient) GetOTP(email string) (string, error){
	return r.client.Get(c, "otp:"+email).Result()
}

func (r *RedisClient) DeleteOTP(email string) error{
	return r.client.Del(c, "otp:"+email).Err()
}

func (r *RedisClient) RateLimitCheck(email string) (bool, int, error){
	key := "rate_limit:otp:" + email
	val, err := r.client.Exists(c, key).Result()
	if err != nil{
		return false, 0, err
	}

	if val > 0{
       ttl := r.client.TTL(c, key).Val()
        return false, int(ttl.Seconds()), nil
    }
    

    r.client.Set(c, key, "1", 60*time.Second)
    return true, 0, nil
}



func (r *RedisClient) SetPendingUser(email string, data string) error {
    return r.client.Set(c, "pending_user:" + email, data, 10*time.Minute).Err()
}

func (r *RedisClient) GetPendingUser(email string) (string, error) {
    return r.client.Get(c, "pending_user:" + email).Result()
}

func (r *RedisClient) DeletePendingUser(email string) error {
    return r.client.Del(c, "pending_user:" + email).Err()
}


func (r *RedisClient) SetUserCache(key string, userData string) error {
    return r.client.Set(c, "user:" + key, userData, 24*time.Hour).Err()
}

func (r *RedisClient) GetUserCache(key string) (string, error) {
    return r.client.Get(c, "user:" + key).Result()
}

func (r *RedisClient) InvalidateUserCache(email string, username string) error {
    r.client.Del(c, "user:" + email)
    r.client.Del(c, "user:" + username)
    return nil
}