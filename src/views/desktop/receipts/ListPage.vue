<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <v-card variant="flat" min-height="780">
                    <template #title>
                        <div class="title-and-toolbar d-flex align-center">
                            <span>{{ tt('Receipt List') }}</span>
                            <v-btn class="ms-3" color="default" variant="outlined"
                                   :disabled="loading" @click="add">{{ tt('Add') }}</v-btn>
                            <v-btn density="compact" color="default" variant="text" size="24"
                                   class="ms-2" :icon="true" :loading="loading" @click="reload">
                                <template #loader>
                                    <v-progress-circular indeterminate size="20"/>
                                </template>
                                <v-icon :icon="mdiRefresh" size="24" />
                                <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                            </v-btn>
                        </div>
                    </template>

                    <v-table class="receipts-table table-striped" :hover="!loading">
                        <thead>
                        <tr>
                            <th class="text-no-wrap">{{ tt('Date') }}</th>
                            <th class="text-no-wrap">{{ tt('Description') }}</th>
                            <th class="text-no-wrap text-cell">{{ tt('Comment') }}</th>
                            <th class="text-no-wrap text-end">{{ tt('Total Amount') }}</th>
                            <th class="text-no-wrap text-end">{{ tt('Transaction Count') }}</th>
                            <th class="text-no-wrap text-end">{{ tt('Operation') }}</th>
                        </tr>
                        </thead>

                        <tbody v-if="loading && receipts.length < 1">
                        <tr :key="itemIdx" v-for="itemIdx in [1, 2, 3, 4, 5]">
                            <td colspan="6" class="px-0">
                                <v-skeleton-loader type="text" :loading="true"></v-skeleton-loader>
                            </td>
                        </tr>
                        </tbody>

                        <tbody v-if="!loading && receipts.length < 1">
                        <tr>
                            <td colspan="6">{{ tt('No available receipt') }}</td>
                        </tr>
                        </tbody>

                        <tbody>
                        <tr :key="receipt.id" v-for="receipt in receipts">
                            <td class="text-no-wrap">{{ formatTime(receipt.time) }}</td>
                            <td class="text-no-wrap text-max-width-200">{{ receipt.place || '-' }}</td>
                            <td class="text-cell text-max-width-300">{{ receipt.comment || '-' }}</td>
                            <td class="text-end text-no-wrap">{{ formatAmountToLocalizedNumerals(receipt.totalAmount) }}</td>
                            <td class="text-end text-no-wrap">{{ receipt.transactionCount }}</td>
                            <td class="text-end text-no-wrap">
                                <v-btn density="compact" variant="text" size="small"
                                       :prepend-icon="mdiEyeOutline"
                                       @click="viewReceipt(receipt)">{{ tt('View') }}</v-btn>
                                <v-btn density="compact" variant="text" size="small"
                                       :prepend-icon="mdiPencilOutline"
                                       @click="edit(receipt)">{{ tt('Edit') }}</v-btn>
                                <v-btn density="compact" variant="text" size="small"
                                       :prepend-icon="mdiDeleteOutline"
                                       @click="remove(receipt)">{{ tt('Delete') }}</v-btn>
                            </td>
                        </tr>
                        </tbody>
                    </v-table>
                </v-card>
            </v-card>
        </v-col>
    </v-row>

    <receipt-dialog ref="receiptDialog" />
    <receipt-detail-dialog ref="receiptDetailDialog" />
    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackbar" />
</template>

<script setup lang="ts">
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import ReceiptDialog from './dialogs/ReceiptDialog.vue';
import ReceiptDetailDialog from './dialogs/ReceiptDetailDialog.vue';

import { ref, computed } from 'vue';

import { useI18n } from '@/locales/helpers.ts';

import { useReceiptsStore } from '@/stores/receipt.ts';

import { parseDateTimeFromUnixTime } from '@/lib/datetime.ts';
import type { Receipt } from '@/models/receipt.ts';

import {
    mdiRefresh,
    mdiPencilOutline,
    mdiDeleteOutline,
    mdiEyeOutline
} from '@mdi/js';

type ReceiptDialogType = InstanceType<typeof ReceiptDialog>;
type ReceiptDetailDialogType = InstanceType<typeof ReceiptDetailDialog>;
type ConfirmDialogType = InstanceType<typeof ConfirmDialog>;
type SnackBarType = InstanceType<typeof SnackBar>;

const { tt, formatAmountToLocalizedNumerals, formatDateTimeToShortDate } = useI18n();

const receiptsStore = useReceiptsStore();

const receiptDialog = ref<ReceiptDialogType | null>(null);
const receiptDetailDialog = ref<ReceiptDetailDialogType | null>(null);
const confirmDialog = ref<ConfirmDialogType | null>(null);
const snackbar = ref<SnackBarType | null>(null);

const loading = ref<boolean>(false);

const receipts = computed<Receipt[]>(() => receiptsStore.receipts);

function formatTime(unixTime: number): string {
    return formatDateTimeToShortDate(parseDateTimeFromUnixTime(unixTime));
}

function reload(): void {
    loading.value = true;

    receiptsStore.loadReceipts().then(() => {
        loading.value = false;
        snackbar.value?.showMessage('Receipt list has been updated');
    }).catch(error => {
        loading.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function add(): void {
    receiptDialog.value?.open().then(result => {
        if (result && result.message) {
            snackbar.value?.showMessage(result.message);
        }
    }).catch(error => {
        if (error) {
            snackbar.value?.showError(error);
        }
    });
}

function edit(receipt: Receipt): void {
    receiptDialog.value?.open({ id: receipt.id }).then(result => {
        if (result && result.message) {
            snackbar.value?.showMessage(result.message);
        }
    }).catch(error => {
        if (error) {
            snackbar.value?.showError(error);
        }
    });
}

function viewReceipt(receipt: Receipt): void {
    receiptDetailDialog.value?.open({ id: receipt.id }).catch(error => {
        if (error) {
            snackbar.value?.showError(error);
        }
    });
}

function remove(receipt: Receipt): void {
    confirmDialog.value?.open('Are you sure you want to delete this receipt?').then(() => {
        loading.value = true;

        receiptsStore.deleteReceipt(receipt.id).then(() => {
            loading.value = false;
        }).catch(error => {
            loading.value = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    });
}

reload();
</script>

<style>
.receipts-table .text-max-width-200 {
    max-width: 200px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.receipts-table .text-max-width-300 {
    max-width: 300px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.receipts-table .text-cell {
    max-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
</style>
