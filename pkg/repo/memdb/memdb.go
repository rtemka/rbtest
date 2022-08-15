// Пакет memdb реализует контракт БД, для тестов.
package memdb

import (
	"context"

	"github.com/rtemka/rbtest/domain"
)

type item = domain.Item

type MemDB struct{}

var testItem1 = item{
	ID:   1,
	Name: "test one",
}

var testItem2 = item{
	ID:   2,
	Name: "test two",
}

func New() *MemDB { return &MemDB{} }

// Items возвращает списком все объекты из БД.
func (m *MemDB) Items(context.Context) ([]item, error) {
	return []item{testItem1, testItem2}, nil
}

// Item находит объект по id.
func (m *MemDB) Item(ctx context.Context, id int64) (item, error) {
	return testItem1, nil
}

// DeleteItem удаляет из БД объект по id.
func (m *MemDB) DeleteItem(ctx context.Context, id int64) error {
	return nil
}

// UpdateItem обновляет в БД объект.
func (m *MemDB) UpdateItem(ctx context.Context, item item) error {
	return nil
}

// Close закрывает подключение к БД.
func (m *MemDB) Close() error { return nil }
