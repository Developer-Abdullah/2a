package redis
import ( "context"; "encoding/json"; "errors"; "time"; "github.com/redis/go-redis/v9" )
var ErrCacheMiss = errors.New("cache miss")
type CacheService struct { client *redis.Client }
func NewCacheService(client *redis.Client) *CacheService { return &CacheService{client: client} }
func (c *CacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
bytes, _ := json.Marshal(value)
return c.client.Set(ctx, key, bytes, ttl).Err()
}
func (c *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
val, err := c.client.Get(ctx, key).Bytes()
if err != nil { if errors.Is(err, redis.Nil) { return ErrCacheMiss }; return err }
return json.Unmarshal(val, dest)
}
