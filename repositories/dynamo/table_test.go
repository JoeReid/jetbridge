package dynamo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateTable(t *testing.T) {
	db := testingDynamoDB(t)

	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	err := CreateTable(ctx, db, "test-table")
	assert.NoError(t, err)
}
