package cache

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/rtemka/rbtest/domain"
)

type item = domain.Item
type repo = domain.Repository

type Cache struct {
	mu     sync.RWMutex
	data   []item // здесь можно исользовать мапу, но обойдемся слайсом
	repo   repo
	logger *log.Logger
}

func New(ctx context.Context, db repo, logger *log.Logger, updInterval time.Duration) *Cache {
	c := Cache{
		repo:   db,
		logger: logger,
	}

	go c.cacheLoader(ctx, updInterval) // горутина для обновления кэша.

	return &c
}

// set обновляет кэш целиком, ошибка для удобства
// логируется.
func (c *Cache) update(ctx context.Context) {

	items, err := c.repo.Items(ctx)
	if err != nil {
		c.logger.Println(err)
		return
	}
	c.mu.Lock()
	c.data = items
	c.mu.Unlock()
}

// all возвращает полную копию кэша для дальнешего использования.
func (c *Cache) all(ctx context.Context) []item {
	var out []item
	c.mu.RLock()
	_ = copy(out, c.data)
	c.mu.RUnlock()
	return out
}

func (c *Cache) len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

func (c *Cache) get(id int64) item {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for i := range c.data {
		if c.data[i].ID == id {
			return c.data[i]
		}
	}
	return item{}
}

// cacheLoader обновляет кэш каждый раз через interval.
func (c *Cache) cacheLoader(ctx context.Context, interval time.Duration) {
	upd := func() {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		c.update(ctx)
	}

	upd() // первый раз сразу

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
			upd()
		}
	}
}

// Имитируем контрак БД, чтобы работать поверх неё.

// Close закрываем подключение к БД.
func (c *Cache) Close() error {
	return c.repo.Close()
}

// Items возвращает списком все объекты из БД.
func (c *Cache) Items(ctx context.Context) ([]item, error) {
	if c.len() == 0 {
		c.update(ctx)
	}
	return c.all(ctx), nil
}

// Items возвращает списком все объекты из БД.
func (c *Cache) Item(ctx context.Context, id int64) (item, error) {
	if c.len() == 0 {
		c.update(ctx)
	}
	return c.get(id), nil
}

// DeleteItem удаляет из БД объект по id.
func (c *Cache) DeleteItem(ctx context.Context, id int64) error {
	err := c.repo.DeleteItem(ctx, id)
	if err != nil {
		return err
	}
	go c.update(ctx)
	return nil
}

// UpdateItem обновляет в БД объект.
func (c *Cache) UpdateItem(ctx context.Context, item item) error {
	err := c.repo.UpdateItem(ctx, item)
	if err != nil {
		return err
	}
	go c.update(ctx)
	return nil
}