import { ref } from 'vue';
import { defineStore } from 'pinia';

import { Receipt, type ReceiptInfoResponse, type ReceiptCreateRequest, type ReceiptModifyRequest } from '@/models/receipt.ts';
import services from '@/lib/services.ts';

export const useReceiptsStore = defineStore('receipts', () => {
    const receipts = ref<Receipt[]>([]);
    const loading = ref<boolean>(false);

    async function loadReceipts(): Promise<void> {
        loading.value = true;

        try {
            const response = await services.getAllReceipts();
            const data = (response.data as any).result as ReceiptInfoResponse[];
            receipts.value = data.map(r => new Receipt(r));
        } finally {
            loading.value = false;
        }
    }

    async function getReceiptById(id: string): Promise<Receipt> {
        const response = await services.getReceipt({ id });
        const data = (response.data as any).result as ReceiptInfoResponse;
        return new Receipt(data);
    }

    async function addReceipt(req: ReceiptCreateRequest): Promise<Receipt> {
        const response = await services.addReceipt(req);
        const data = (response.data as any).result as ReceiptInfoResponse;
        const receipt = new Receipt(data);
        receipts.value.unshift(receipt);
        return receipt;
    }

    async function modifyReceipt(req: ReceiptModifyRequest): Promise<Receipt> {
        const response = await services.modifyReceipt(req);
        const data = (response.data as any).result as ReceiptInfoResponse;
        const updated = new Receipt(data);
        const index = receipts.value.findIndex(r => r.id === updated.id);
        if (index >= 0) {
            receipts.value[index] = updated;
        }
        return updated;
    }

    async function deleteReceipt(id: string): Promise<void> {
        await services.deleteReceipt({ id });
        receipts.value = receipts.value.filter(r => r.id !== id);
    }

    return {
        receipts,
        loading,
        loadReceipts,
        getReceiptById,
        addReceipt,
        modifyReceipt,
        deleteReceipt,
    };
});
