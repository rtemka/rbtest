package mongo

import (
	"context"

	"github.com/rtemka/rbtest/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Когда при выполнении операции не найдено
// ни одного документа
var ErrNoDocuments = mongo.ErrNoDocuments

// псевдоним для объекта хранения БД
type item = domain.Item

// Mongo структура для выполнения CRUD операций с БД
type Mongo struct {
	client *mongo.Client // клиент mongo
	// название текущей db,
	// переключается методом Database()
	database string
	// название текущей collection,
	// переключается методом Collection()
	collection string
}

// New подключается к БД, используя connstr, и возвращает
// объект для работы с БД
func New(connstr, database, collection string) (*Mongo, error) {

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(connstr))
	if err != nil {
		return nil, err
	}

	return &Mongo{
		client:     client,
		database:   database,
		collection: collection,
	}, client.Ping(context.Background(), nil)
}

// Database переключает имя базы данных mongodb
// в структуре *Mongo
func (m *Mongo) Database(database string) *Mongo {
	m.database = database
	return m
}

// Collection переключает имя коллекции в структуре *Mongo.
func (m *Mongo) Collection(collection string) *Mongo {
	m.collection = collection
	return m
}

// Close закрывает соединение с БД
func (m *Mongo) Close() error {
	return m.client.Disconnect(context.Background())
}

// Items возвращает списком все объекты из БД.
func (m *Mongo) Items(ctx context.Context) ([]item, error) {

	col := m.client.Database(m.database).Collection(m.collection)

	cursor, err := col.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	var items []item

	return items, cursor.All(ctx, &items)
}

// AddItem добавляет в БД объект, если он уже
// есть в БД, то no-op.
func (m *Mongo) AddItem(ctx context.Context, item item) error {

	col := m.client.Database(m.database).Collection(m.collection)
	filter := bson.D{bson.E{Key: "id", Value: item.ID}}
	opts := options.Update().SetUpsert(true)
	upd := bson.D{
		bson.E{
			Key: "$setOnInsert", Value: item},
	}

	_, err := col.UpdateOne(ctx, filter, upd, opts)

	return err
}

// Item находит объект по id.
// Возвращает ошибку ErrNoDocuments в случае если документ не найден.
func (m *Mongo) Item(ctx context.Context, id int64) (item, error) {

	col := m.client.Database(m.database).Collection(m.collection)

	var item item

	return item, col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: id}}).Decode(&item)
}

// DeleteItem удаляет из БД объект по id.
func (m *Mongo) DeleteItem(ctx context.Context, id int64) error {
	col := m.client.Database(m.database).Collection(m.collection)
	_, err := col.DeleteOne(ctx, bson.D{bson.E{Key: "id", Value: id}})
	return err
}

// UpdateItem обновляет в БД объект.
func (m *Mongo) UpdateItem(ctx context.Context, item item) error {

	col := m.client.Database(m.database).Collection(m.collection)

	filter := bson.D{bson.E{Key: "id", Value: item.ID}}
	upd := bson.D{
		bson.E{
			Key: "$set", Value: item},
	}
	_, err := col.UpdateOne(ctx, filter, upd)

	return err
}
