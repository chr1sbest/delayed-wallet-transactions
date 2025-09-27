/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type Transaction = {
    id?: string;
    from_user_id?: string;
    to_user_id?: string;
    amount?: number;
    status?: Transaction.status;
    /**
     * The delay in seconds before the transaction is processed.
     */
    delay_seconds?: number;
    created_at?: string;
    updated_at?: string;
    /**
     * A Unix timestamp representing the expiration time of the transaction record.
     */
    ttl?: number;
};
export namespace Transaction {
    export enum status {
        RESERVED = 'RESERVED',
        PENDING_APPROVAL = 'PENDING_APPROVAL',
        APPROVED = 'APPROVED',
        REJECTED = 'REJECTED',
        COMPLETED = 'COMPLETED',
        FAILED = 'FAILED',
    }
}

