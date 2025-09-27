package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// WebSocketConnection represents a record in the WebSocket connections table.
type WebSocketConnection struct {
	ConnectionID string `dynamodbav:"connection_id"`
	PK           string `dynamodbav:"pk"`
}

// AddConnection saves a new WebSocket connection ID to the database.
func (s *Store) AddConnection(ctx context.Context, connectionID string) error {
	conn := WebSocketConnection{ConnectionID: connectionID, PK: "connections"}
	item, err := attributevalue.MarshalMap(conn)
	if err != nil {
		return fmt.Errorf("failed to marshal connection: %w", err)
	}

	_, err = s.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.WebsocketConnectionsTableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}

	return nil
}

// RemoveConnection deletes a WebSocket connection ID from the database.
func (s *Store) RemoveConnection(ctx context.Context, connectionID string) error {
	key, err := attributevalue.MarshalMap(map[string]string{
		"connection_id": connectionID,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal connection key: %w", err)
	}

	_, err = s.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(s.WebsocketConnectionsTableName),
		Key:       key,
	})
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	return nil
}

// GetAllConnections retrieves all active WebSocket connection IDs from the database.
func (s *Store) GetAllConnections(ctx context.Context) ([]string, error) {
	queryOutput, err := s.Client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.WebsocketConnectionsTableName),
		IndexName:              aws.String("pk-index"),
		KeyConditionExpression: aws.String("pk = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "connections"},
		},
		ProjectionExpression: aws.String("connection_id"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query connections table: %w", err)
	}

	var connections []WebSocketConnection
	if err := attributevalue.UnmarshalListOfMaps(queryOutput.Items, &connections); err != nil {
		return nil, fmt.Errorf("failed to unmarshal connections: %w", err)
	}

	connectionIDs := make([]string, len(connections))
	for i, conn := range connections {
		connectionIDs[i] = conn.ConnectionID
	}

	return connectionIDs, nil
}
