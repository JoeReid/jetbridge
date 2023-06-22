package dynamo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewBindings(t *testing.T) {
	db := testingDynamoDB(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := CreateTable(ctx, db, "test-table")
	require.NoError(t, err)

	table, err := NewBindings(db, "test-table")
	require.NoError(t, err)
	require.NotNil(t, table)
}
