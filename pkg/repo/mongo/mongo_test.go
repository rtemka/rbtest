package mongo

import (
	"context"
	"fmt"
	"os"
	"testing"
)

var tdb *Mongo

const testDBEnv = "TEST_DB_URL"

var testItem1 = item{
	ID:   1,
	Name: "test one",
}

var testItem2 = item{
	ID:   2,
	Name: "test two",
}

var testData = []any{testItem1, testItem2}

func restoreDB(db *Mongo) error {

	err := db.client.Database(db.database).Drop(context.Background())
	if err != nil {
		return err
	}
	col := db.client.Database(db.database).Collection(db.collection)
	_, err = col.InsertMany(context.Background(), testData)

	return err
}

func TestMain(m *testing.M) {

	connstr, ok := os.LookupEnv(testDBEnv)
	if !ok {
		os.Exit(m.Run()) // тест будет пропущен
	}

	var err error
	tdb, err = New(connstr, "testdb", "testcollection")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer tdb.Close()

	if err := restoreDB(tdb); err != nil {
		_ = tdb.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestMongo(t *testing.T) {
	if _, ok := os.LookupEnv(testDBEnv); !ok {
		t.Skipf("you should set %q env variable to run this test, skipped...", testDBEnv)
	}

	t.Run("DeleteItem", func(t *testing.T) {

		want := testItem2

		err := tdb.DeleteItem(context.Background(), want.ID)
		if err != nil {
			t.Fatalf("DeleteItem() = err %v", err)
		}

		got, err := tdb.Item(context.Background(), want.ID)
		if err != nil && err != ErrNoDocuments {
			t.Fatalf("Item() = err %v", err)
		}

		if got != (item{}) {
			t.Errorf("DeleteItem() got = %v, want nothing", got)
		}

	})

	t.Run("AddItem", func(t *testing.T) {

		want := testItem2

		err := tdb.AddItem(context.Background(), want)
		if err != nil {
			t.Fatalf("AddItem() = err %v", err)
		}

		got, err := tdb.Item(context.Background(), want.ID)
		if err != nil {
			t.Fatalf("AddItem() = err %v", err)
		}

		if got != want {
			t.Errorf("AddItem() = %v, want %v", got, want)
		}

	})

	t.Run("UpdateItem()", func(t *testing.T) {
		want := testItem1
		want.Name = "upd name"

		err := tdb.UpdateItem(context.Background(), want)
		if err != nil {
			t.Fatalf("UpdateItem() error = %v", err)
		}

		got, err := tdb.Item(context.Background(), want.ID)
		if err != nil {
			t.Fatalf("Item() error = %v", err)
		}

		if got != want {
			t.Errorf("UpdateItem() got = %v, want = %v", got, want)
		}
	})

}
