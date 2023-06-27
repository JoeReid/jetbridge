package dynamo

import (
	"context"
	"testing"
	"time"

	"github.com/JoeReid/jetbridge/repositories/conformancetest"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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

func TestConformance(t *testing.T) {
	db := testingDynamoDB(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := CreateTable(ctx, db, "test-table")
	require.NoError(t, err)

	peers, err := NewPeers(db, "test-table")
	require.NoError(t, err)

	bindings, err := NewBindings(db, "test-table")
	require.NoError(t, err)

	cs := conformancetest.NewBindingsConformanceSuite(peers, bindings)
	suite.Run(t, cs)
}
