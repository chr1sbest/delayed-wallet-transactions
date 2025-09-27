/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type NewTransaction = {
    from_user_id: string;
    to_user_id: string;
    /**
     * The amount of the transaction in the smallest currency unit (e.g., cents).
     */
    amount: number;
    /**
     * An optional delay in seconds before the transaction is processed. Maximum 900 seconds (15 minutes).
     */
    delay_seconds?: number;
};

