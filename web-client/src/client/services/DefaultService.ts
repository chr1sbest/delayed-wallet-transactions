/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { LedgerEntry } from '../models/LedgerEntry';
import type { NewTransaction } from '../models/NewTransaction';
import type { NewWallet } from '../models/NewWallet';
import type { Transaction } from '../models/Transaction';
import type { Wallet } from '../models/Wallet';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class DefaultService {
    /**
     * Schedule a new transaction
     * @param requestBody
     * @returns Transaction Transaction created successfully.
     * @throws ApiError
     */
    public static scheduleTransaction(
        requestBody: NewTransaction,
    ): CancelablePromise<Transaction> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/transactions',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                400: `Invalid request body`,
                422: `Insufficient funds or other processing error`,
            },
        });
    }
    /**
     * Get a transaction by its ID
     * @param transactionId
     * @returns Transaction A single transaction
     * @throws ApiError
     */
    public static getTransactionById(
        transactionId: string,
    ): CancelablePromise<Transaction> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/transactions/{transactionId}',
            path: {
                'transactionId': transactionId,
            },
            errors: {
                404: `Transaction not found`,
            },
        });
    }
    /**
     * Cancel a transaction by its ID
     * @param transactionId
     * @returns void
     * @throws ApiError
     */
    public static cancelTransactionById(
        transactionId: string,
    ): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/transactions/{transactionId}',
            path: {
                'transactionId': transactionId,
            },
            errors: {
                404: `Transaction not found or not in a cancellable state`,
                409: `Transaction is not in a cancellable state`,
            },
        });
    }
    /**
     * Create a new wallet
     * @param requestBody
     * @returns Wallet Wallet created successfully
     * @throws ApiError
     */
    public static createWallet(
        requestBody: NewWallet,
    ): CancelablePromise<Wallet> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/wallets',
            body: requestBody,
            mediaType: 'application/json',
            errors: {
                409: `Wallet for this user already exists`,
            },
        });
    }
    /**
     * List all wallets
     * @returns Wallet A list of wallets
     * @throws ApiError
     */
    public static listWallets(): CancelablePromise<Array<Wallet>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/wallets',
        });
    }
    /**
     * Get a wallet by user ID
     * @param userId
     * @returns Wallet A single wallet
     * @throws ApiError
     */
    public static getWalletByUserId(
        userId: string,
    ): CancelablePromise<Wallet> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/wallets/{userId}',
            path: {
                'userId': userId,
            },
            errors: {
                404: `Wallet not found`,
            },
        });
    }
    /**
     * Delete a wallet by user ID
     * @param userId
     * @returns void
     * @throws ApiError
     */
    public static deleteWallet(
        userId: string,
    ): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/wallets/{userId}',
            path: {
                'userId': userId,
            },
            errors: {
                404: `Wallet not found`,
            },
        });
    }
    /**
     * List all transactions for a user
     * @param userId
     * @returns Transaction A list of transactions for the user
     * @throws ApiError
     */
    public static listTransactionsByUserId(
        userId: string,
    ): CancelablePromise<Array<Transaction>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/users/{userId}/transactions',
            path: {
                'userId': userId,
            },
        });
    }
    /**
     * List recent ledger entries
     * @param limit
     * @returns LedgerEntry A list of recent ledger entries
     * @throws ApiError
     */
    public static listLedgerEntries(
        limit: number = 20,
    ): CancelablePromise<Array<LedgerEntry>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/ledger',
            query: {
                'limit': limit,
            },
        });
    }
}
