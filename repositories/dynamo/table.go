package dynamo

import (
	"context"
	"fmt"
	"time"

	"github.com/guregu/dynamo"
)

func CreateTable(ctx context.Context, db *dynamo.DB, tableName string) error {
	type schema struct {
		RecordType  string    `dynamo:"pk,hash"`
		ID          string    `dynamo:"sk,range"`
		Name        string    `dynamo:"name" localIndex:"name-index,range"`
		CreatedAt   time.Time `dynamo:"created_at" localIndex:"created_at-index,range"`
		UpdatedAt   time.Time `dynamo:"updated_at" localIndex:"updated_at-index,range"`
		DeleteAfter time.Time `dynamo:"delete_after,unixtime"`
	}

	if err := db.CreateTable(tableName, &schema{}).RunWithContext(ctx); err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}

	if err := db.Table(tableName).WaitWithContext(ctx); err != nil {
		return fmt.Errorf("failed to wait for table %s: %w", tableName, err)
	}

	if err := db.Table(tableName).UpdateTTL("delete_after", true).RunWithContext(ctx); err != nil {
		return fmt.Errorf("failed to enable TTL on table %s: %w", tableName, err)
	}

	return nil
}
