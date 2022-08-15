package domain

import "context"

type Item struct {
	// так как используем mongo, то тут можно было бы использовать
	// ObjectID mongo, но для простоты используем просто int
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Repository interface {
	Items(context.Context) ([]Item, error)            // Items возвращает списком все объекты из БД.
	Item(ctx context.Context, id int64) (Item, error) // Item находит объект по id.
	DeleteItem(ctx context.Context, id int64) error   // DeleteItem удаляет из БД объект по id.
	UpdateItem(ctx context.Context, item Item) error  // UpdateItem обновляет в БД объект.
	Close() error
}
