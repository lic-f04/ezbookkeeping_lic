import type { TransactionInfoResponse } from './transaction.ts';

export interface ReceiptSummaryResponse {
    readonly id: string;
    readonly time: number;
    readonly utcOffset: number;
    readonly place: string;
    readonly comment: string;
    readonly totalAmount: number;
    readonly status: number;
    readonly accountId?: string;
}

export interface ReceiptInfoResponse {
    readonly id: string;
    readonly time: number;
    readonly utcOffset: number;
    readonly accountId?: string;
    readonly place: string;
    readonly comment: string;
    readonly totalAmount: number;
    readonly status: number;
    readonly transactions?: TransactionInfoResponse[];
}

export interface ReceiptCreateRequest {
    readonly time: number;
    readonly utcOffset: number;
    readonly accountId?: string;
    readonly place: string;
    readonly comment: string;
}

export interface ReceiptModifyRequest {
    readonly id: string;
    readonly time: number;
    readonly utcOffset: number;
    readonly accountId?: string;
    readonly place: string;
    readonly comment: string;
    readonly status: number;
}

export interface ReceiptDeleteRequest {
    readonly id: string;
}

export interface ReceiptAddTransactionsRequest {
    readonly receiptId: string;
    readonly transactionIds: string[];
}

export interface ReceiptRemoveTransactionRequest {
    readonly receiptId: string;
    readonly transactionId: string;
}

export interface ReceiptStatusModifyRequest {
    readonly id: string;
    readonly status: number;
}

export class Receipt implements ReceiptInfoResponse {
    public readonly id: string;
    public readonly time: number;
    public readonly utcOffset: number;
    public readonly accountId?: string;
    public readonly place: string;
    public readonly comment: string;
    public readonly totalAmount: number;
    public readonly status: number = 0;
    public transactions?: TransactionInfoResponse[];

    public constructor(data: ReceiptInfoResponse) {
        this.id = data.id;
        this.time = data.time;
        this.utcOffset = data.utcOffset;
        this.accountId = data.accountId;
        this.place = data.place;
        this.comment = data.comment;
        this.totalAmount = data.totalAmount;
        this.status = data.status ?? 0;
        this.transactions = data.transactions;
    }

    public get transactionCount(): number {
        return this.transactions?.length ?? 0;
    }
}
