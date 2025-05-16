package localCache

import (
	"sync"
	"time"
)

// 定义缓存项结构
type item struct {
	value  any         // 存储的值
	expiry time.Time   // 过期时间
	timer  *time.Timer // 定时器（用于自动删除）
}

// 缓存对象结构
type Cache struct {
	mu   sync.RWMutex     // 读写锁
	data map[string]*item // 数据存储
}

var (
	instance *Cache    // 单例实例
	once     sync.Once // 单例控制
)

// 获取缓存实例（单例模式）
func NewCache() *Cache {
	once.Do(func() {
		instance = &Cache{
			data: make(map[string]*item),
		}
	})
	return instance
}

// Set 设置缓存（永不过期时expiration传0）
func (c *Cache) Set(key string, value any, expiration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果已存在则停止原有定时器
	if existing, found := c.data[key]; found && existing.timer != nil {
		existing.timer.Stop()
	}

	// 创建新缓存项
	newItem := &item{
		value:  value,
		expiry: time.Now().Add(expiration),
	}

	// 设置自动删除定时器
	if expiration > 0 {
		newItem.timer = time.AfterFunc(expiration, func() {
			c.Delete(key)
		})
	}

	c.data[key] = newItem
}

// Get 获取缓存值
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	item, found := c.data[key]
	c.mu.RUnlock()

	if !found {
		return nil, false
	}

	// 检查是否过期
	if item.expiry.IsZero() || time.Now().Before(item.expiry) {
		return item.value, true
	}

	// 已过期则删除
	c.Delete(key)
	return nil, false
}

// Delete 删除缓存项
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, found := c.data[key]; found {
		// 停止定时器（如果存在）
		if item.timer != nil {
			item.timer.Stop()
		}
		delete(c.data, key)
	}
}
