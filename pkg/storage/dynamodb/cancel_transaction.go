package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/chris/delayed-wallet-transactions/pkg/models"
	"github.com/chris/delayed-wallet-transactions/pkg/storage"
)

func (s *Store) CancelTransaction(ctx context.Context, txID string) error {
	tx, err := s.GetTransaction(ctx, txID)
	if err != nil {
		return fmt.Errorf("failed to get transaction for cancellation: %w", err)
	}

	if tx.Status != models.RESERVED {
		return storage.ErrTransactionNotCancellable
	}

	senderWallet, err := s.GetWallet(ctx, tx.FromUserId)
	if err != nil {
		return fmt.Errorf("failed to get sender's wallet for cancellation: %w", err)
	}

	now := time.Now()
	amountAV, err := attributevalue.Marshal(tx.Amount)
	if err != nil {
		return fmt.Errorf("failed to marshal amount for cancellation: %w", err)
	}

	cancelledStatusAV, err := attributevalue.Marshal(models.CANCELLED)
	if err != nil {
		return fmt.Errorf("failed to marshal cancelled status: %w", err)
	}
	nowAV, err := attributevalue.Marshal(now)
	if err != nil {
		return fmt.Errorf("failed to marshal timestamp for cancellation: %w", err)
	}

	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Update: &types.Update{
					TableName: aws.String(s.WalletsTableName),
					Key:       map[string]types.AttributeValue{"user_id": &types.AttributeValueMemberS{Value: tx.FromUserId}},
					UpdateExpression:    aws.String("SET balance = balance + :amount, reserved = reserved - :amount, version = version + :inc"),
					ConditionExpression: aws.String("version = :version"),
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":amount":   amountAV,
						":version":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", senderWallet.Version)},
						":inc":      &types.AttributeValueMemberN{Value: "1"},
					},
				},
			},
			{
				Update: &types.Update{
					TableName: aws.String(s.TransactionsTableName),
					Key:       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: tx.Id}},
					UpdateExpression:    aws.String("SET #status = :cancelled_status, updated_at = :now"),
					ConditionExpression: aws.String("#status = :reserved_status"),
					ExpressionAttributeNames: map[string]string{
						"#status": "status",
					},
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":cancelled_status": cancelledStatusAV,
						":reserved_status":  &types.AttributeValueMemberS{Value: string(models.RESERVED)},
						":now":              nowAV,
					},
				},
			},
		},
	}

	_, err = s.Client.TransactWriteItems(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to execute cancellation transaction: %w", err)
	}

	return nil
}
