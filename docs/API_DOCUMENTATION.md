# Delayed Wallet API
An API for a delayed wallet system, allowing for transactions to be scheduled and processed asynchronously.

## Version: 1.0.0

### /transactions/{transactionId}/notify-settlement

#### POST
##### Summary:

Notify of transaction settlement

##### Description:

An internal endpoint for the settlement service to notify the API that a transaction has been settled.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| transactionId | path |  | Yes | string (uuid) |

##### Responses

| Code | Description |
| ---- | ----------- |
| 204 | Notification accepted. |
| 404 | Transaction not found. |

### /transactions

#### POST
##### Summary:

Schedule a new transaction

##### Responses

| Code | Description |
| ---- | ----------- |
| 201 | Transaction created successfully. |
| 400 | Invalid request body |
| 422 | Insufficient funds or other processing error |

### /transactions/{transactionId}

#### GET
##### Summary:

Get a transaction by its ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| transactionId | path |  | Yes | string |

##### Responses

| Code | Description |
| ---- | ----------- |
| 200 | A single transaction |
| 404 | Transaction not found |

#### DELETE
##### Summary:

Cancel a transaction by its ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| transactionId | path |  | Yes | string |

##### Responses

| Code | Description |
| ---- | ----------- |
| 204 | Transaction cancelled successfully |
| 404 | Transaction not found or not in a cancellable state |
| 409 | Transaction is not in a cancellable state |

### /wallets

#### POST
##### Summary:

Create a new wallet

##### Responses

| Code | Description |
| ---- | ----------- |
| 201 | Wallet created successfully |
| 409 | Wallet for this user already exists |

#### GET
##### Summary:

List all wallets

##### Responses

| Code | Description |
| ---- | ----------- |
| 200 | A list of wallets |

### /wallets/{userId}

#### GET
##### Summary:

Get a wallet by user ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| userId | path |  | Yes | string |

##### Responses

| Code | Description |
| ---- | ----------- |
| 200 | A single wallet |
| 404 | Wallet not found |

#### DELETE
##### Summary:

Delete a wallet by user ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| userId | path |  | Yes | string |

##### Responses

| Code | Description |
| ---- | ----------- |
| 204 | Wallet deleted successfully |
| 404 | Wallet not found |

### /users/{userId}/transactions

#### GET
##### Summary:

List all transactions for a user

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| userId | path |  | Yes | string |

##### Responses

| Code | Description |
| ---- | ----------- |
| 200 | A list of transactions for the user |

### /ledger

#### GET
##### Summary:

List recent ledger entries

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| limit | query |  | No | integer |

##### Responses

| Code | Description |
| ---- | ----------- |
| 200 | A list of recent ledger entries |

### Models


#### Error

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message | string |  | No |

#### NewWallet

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| user_id | string |  | Yes |
| name | string |  | Yes |

#### NewTransaction

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| from_user_id | string |  | Yes |
| to_user_id | string |  | Yes |
| amount | long | The amount of the transaction in the smallest currency unit (e.g., cents). | Yes |
| delay_seconds | integer | An optional delay in seconds before the transaction is processed. Maximum 900 seconds (15 minutes). | No |

#### Transaction

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| from_user_id | string |  | No |
| to_user_id | string |  | No |
| amount | long |  | No |
| status | string |  | No |
| delay_seconds | integer | The delay in seconds before the transaction is processed. | No |
| created_at | dateTime |  | No |
| updated_at | dateTime |  | No |
| ttl | long | A Unix timestamp representing the expiration time of the transaction record. | No |

#### LedgerEntry

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| entry_id | string |  | No |
| transaction_id | string |  | No |
| account_id | string |  | No |
| debit | long |  | No |
| credit | long |  | No |
| timestamp | dateTime |  | No |
| description | string |  | No |

#### Wallet

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| user_id | string |  | No |
| name | string |  | No |
| balance | long | The wallet balance in the smallest currency unit (e.g., cents). | No |
| reserved | long | Funds reserved for pending transactions. | No |
| version | long |  | No |
| created_at | dateTime |  | No |