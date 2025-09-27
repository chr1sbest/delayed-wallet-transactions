/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
export type Wallet = {
    user_id?: string;
    name?: string;
    /**
     * The wallet balance in the smallest currency unit (e.g., cents).
     */
    balance?: number;
    /**
     * Funds reserved for pending transactions.
     */
    reserved?: number;
    version?: number;
    created_at?: string;
};

