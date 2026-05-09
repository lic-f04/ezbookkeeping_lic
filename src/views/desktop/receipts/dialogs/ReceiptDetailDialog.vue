<template>
    <v-dialog width="900" persistent v-model="showState">
        <v-card class="pa-sm-1 pa-md-2">
            <template #title>
                <div class="d-flex align-center justify-center">
                    <h4 class="text-h4">{{ tt('Receipt') }}</h4>
                    <v-progress-circular indeterminate size="22" class="ms-2" v-if="loading"></v-progress-circular>
                </div>
            </template>
            <v-card-text v-if="receipt">
                <v-form class="mt-2">
                    <v-row>
                        <v-col cols="12" md="6">
                            <v-text-field
                                type="text"
                                persistent-placeholder
                                :disabled="loading || submitting"
                                :label="tt('Description')"
                                :placeholder="tt('Receipt Description')"
                                v-model="receipt.place"
                            />
                        </v-col>
                        <v-col cols="12" md="6">
                            <date-time-select
                                :disabled="loading || submitting"
                                :label="tt('Time')"
                                :timezone-utc-offset="receipt.utcOffset"
                                v-model="receipt.time"
                                @error="onShowDateTimeError"
                            />
                        </v-col>
                        <v-col cols="12" md="6">
                            <v-select
                                :disabled="loading || submitting"
                                :label="tt('Account')"
                                :items="accountsStore.allPlainAccounts"
                                item-title="name"
                                item-value="id"
                                clearable
                                v-model="receipt.accountId"
                            />
                        </v-col>
                        <v-col cols="12" md="6">
                            <div class="d-flex align-center h-100">
                                <span class="text-subtitle-2">{{ tt('Total Amount') }}:</span>
                                <span class="text-h6 ms-2" :class="{ 'text-income': receipt.totalAmount > 0, 'text-expense': receipt.totalAmount < 0 }">{{ formatAmountToLocalizedNumerals(Math.abs(receipt.totalAmount)) }}</span>
                            </div>
                        </v-col>
                        <v-col cols="12">
                            <v-textarea
                                type="text"
                                persistent-placeholder
                                rows="2"
                                :disabled="loading || submitting"
                                :label="tt('Comment')"
                                :placeholder="tt('Receipt Comment')"
                                v-model="receipt.comment"
                            />
                        </v-col>
                    </v-row>
                </v-form>

                <v-divider class="my-4" />

                <div class="d-flex align-center mb-3">
                    <span class="text-h6">{{ tt('Transactions') }}</span>
                    <v-spacer />
                    <v-btn density="compact" variant="text" size="small"
                           color="primary" :disabled="loading || submitting"
                           @click="addTransaction">{{ tt('Add') }}</v-btn>
                </div>

                <v-table class="table-striped" :hover="!loading">
                    <thead>
                    <tr>
                        <th class="text-no-wrap">{{ tt('Description') }}</th>
                        <th class="text-no-wrap">{{ tt('Category') }}</th>
                        <th class="text-no-wrap">{{ tt('Account') }}</th>
                        <th class="text-no-wrap text-end">{{ tt('Amount') }}</th>
                        <th class="text-no-wrap text-end">{{ tt('Operation') }}</th>
                    </tr>
                    </thead>
                    <tbody v-if="!receipt.transactions || receipt.transactions.length < 1">
                    <tr>
                        <td colspan="5">{{ tt('No data') }}</td>
                    </tr>
                    </tbody>
                    <tbody v-else>
                    <tr :key="transaction.id" v-for="transaction in receipt.transactions"
                        class="cursor-pointer" @click="editTransaction(transaction.id)">
                        <td class="text-no-wrap">{{ transaction.comment || '-' }}</td>
                        <td class="text-no-wrap">{{ transaction.category?.name || '-' }}</td>
                        <td class="text-no-wrap">{{ transaction.sourceAccount?.name || '-' }}</td>
                        <td class="text-end text-no-wrap" :class="{ 'text-expense': transaction.type === TransactionType.Expense, 'text-income': transaction.type === TransactionType.Income, 'text-color-primary': transaction.type === TransactionType.Transfer }">{{ formatAmountToLocalizedNumerals(Math.abs(transaction.sourceAmount)) }}</td>
                        <td class="text-end text-no-wrap">
                            <v-btn density="compact" variant="text" size="small"
                                   :prepend-icon="mdiClose"
                                   @click.stop="removeTransaction(transaction.id)">{{ tt('Remove') }}</v-btn>
                        </td>
                    </tr>
                    </tbody>
                </v-table>
            </v-card-text>
            <v-card-text>
                <div class="w-100 d-flex justify-center flex-wrap mt-sm-1 mt-md-2 gap-4">
                    <v-btn :disabled="loading || submitting" @click="save">
                        {{ tt('Save') }}
                        <v-progress-circular indeterminate size="22" class="ms-2" v-if="submitting"></v-progress-circular>
                    </v-btn>
                    <v-btn color="secondary" variant="tonal"
                           :disabled="loading || submitting" @click="close">{{ tt('Close') }}</v-btn>
                </div>
            </v-card-text>
        </v-card>
    </v-dialog>

    <edit-dialog ref="editDialog" :type="TransactionEditPageType.Transaction" />
    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackbar" />
</template>

<script setup lang="ts">
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import EditDialog from '@/views/desktop/transactions/list/dialogs/EditDialog.vue';
import DateTimeSelect from '@/components/desktop/DateTimeSelect.vue';

import { ref } from 'vue';

import { TransactionEditPageType } from '@/views/base/transactions/TransactionEditPageBase.ts';
import { TransactionType } from '@/core/transaction.ts';

import { useI18n } from '@/locales/helpers.ts';

import { useReceiptsStore } from '@/stores/receipt.ts';
import { useAccountsStore } from '@/stores/account.ts';

import { getCurrentUnixTime, getBrowserTimezoneOffsetMinutes } from '@/lib/datetime.ts';
import { Receipt, type ReceiptModifyRequest } from '@/models/receipt.ts';
import services from '@/lib/services.ts';

import {
    mdiClose
} from '@mdi/js';

type ConfirmDialogType = InstanceType<typeof ConfirmDialog>;
type SnackBarType = InstanceType<typeof SnackBar>;
type EditDialogType = InstanceType<typeof EditDialog>;

const { tt, formatAmountToLocalizedNumerals } = useI18n();

const receiptsStore = useReceiptsStore();
const accountsStore = useAccountsStore();

const confirmDialog = ref<ConfirmDialogType | null>(null);
const snackbar = ref<SnackBarType | null>(null);
const editDialog = ref<EditDialogType | null>(null);

const showState = ref<boolean>(false);
const loading = ref<boolean>(false);
const submitting = ref<boolean>(false);

const receipt = ref<Receipt | null>(null);
const currentReceiptId = ref<string>('');
const isCreateMode = ref<boolean>(false);
const isReceiptPersisted = ref<boolean>(false);

let resolveFunc: (() => void) | null = null;
let rejectFunc: ((reason?: unknown) => void) | null = null;

function open(options?: { id?: string }): Promise<void> {
    showState.value = true;
    loading.value = true;
    submitting.value = false;

    const createMode = !options?.id;
    isCreateMode.value = createMode;

    if (createMode) {
        const now = getCurrentUnixTime();
        const offset = getBrowserTimezoneOffsetMinutes(now);
        receipt.value = new Receipt({
            id: '',
            time: now,
            utcOffset: offset,
            accountId: undefined,
            place: '',
            comment: '',
            totalAmount: 0,
            transactions: []
        });
        currentReceiptId.value = '';
        isReceiptPersisted.value = false;
        loading.value = false;
    } else {
        currentReceiptId.value = options!.id!;
        isReceiptPersisted.value = true;
        receiptsStore.getReceiptById(options!.id!).then(data => {
            receipt.value = data;
            loading.value = false;
        }).catch(error => {
            loading.value = false;
            showState.value = false;

            if (!error.processed && rejectFunc) {
                rejectFunc(error);
            }
        });
    }

    return new Promise<void>((resolve, reject) => {
        resolveFunc = resolve;
        rejectFunc = reject;
    });
}

function close(): void {
    resolveFunc?.();
    showState.value = false;
}

function deleteCurrentReceipt(): void {
    const id = currentReceiptId.value;
    if (!id) return;
    receiptsStore.deleteReceipt(id).then(() => {
        isReceiptPersisted.value = false;
        currentReceiptId.value = '';
    }).catch(error => {
        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function save(): void {
    if (!receipt.value) {
        return;
    }

    if (isCreateMode.value && !isReceiptPersisted.value) {
        snackbar.value?.showMessage('Receipt has no transactions to save');
        close();
        return;
    }

    submitting.value = true;

    receiptsStore.modifyReceipt({
        id: currentReceiptId.value,
        time: receipt.value.time,
        utcOffset: receipt.value.utcOffset,
        accountId: receipt.value.accountId,
        place: receipt.value.place,
        comment: receipt.value.comment
    } as ReceiptModifyRequest).then(() => {
        submitting.value = false;
        snackbar.value?.showMessage('You have saved this receipt');
    }).catch(error => {
        submitting.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

async function addTransaction(): Promise<void> {
    if (!receipt.value) {
        return;
    }

    if (isCreateMode.value && !isReceiptPersisted.value) {
        try {
            submitting.value = true;
            const created = await receiptsStore.addReceipt({
                time: receipt.value.time,
                utcOffset: receipt.value.utcOffset,
                accountId: receipt.value.accountId,
                place: receipt.value.place,
                comment: receipt.value.comment
            });
            currentReceiptId.value = created.id;
            receipt.value = created;
            isReceiptPersisted.value = true;
            submitting.value = false;
        } catch (error) {
            submitting.value = false;
            if (!(error as any).processed) {
                snackbar.value?.showError(error as any);
            }
            return;
        }
    }

    const currentId = currentReceiptId.value;

    if (!currentId) {
        return;
    }

    editDialog.value?.open({
        receiptId: currentId
    }).then(() => {
        loading.value = true;
        return receiptsStore.getReceiptById(currentId);
    }).then(data => {
        receipt.value = data;
        loading.value = false;
        snackbar.value?.showMessage('Transaction added to receipt');
    }).catch(error => {
        loading.value = false;

        if (isCreateMode.value && isReceiptPersisted.value && currentReceiptId.value) {
            receiptsStore.getReceiptById(currentReceiptId.value).then(data => {
                if (!data.transactions || data.transactions.length === 0) {
                    deleteCurrentReceipt();
                }
            }).catch(() => {});
        }

        if (error) {
            snackbar.value?.showError(error);
        }
    });
}

function editTransaction(transactionId: string): void {
    editDialog.value?.open({
        id: transactionId
    }).then(() => {
        loading.value = true;
        return receiptsStore.getReceiptById(currentReceiptId.value);
    }).then(data => {
        receipt.value = data;
        loading.value = false;
    }).catch(error => {
        loading.value = false;

        if (error) {
            snackbar.value?.showError(error);
        }
    });
}

function removeTransaction(transactionId: string): void {
    const currentId = currentReceiptId.value;

    if (!currentId) {
        return;
    }

    confirmDialog.value?.open('Are you sure you want to remove this transaction?').then(() => {
        services.deleteTransaction({ id: transactionId }).then(() => {
            return receiptsStore.getReceiptById(currentId);
        }).then(data => {
            receipt.value = data;

            if (!data.transactions || data.transactions.length === 0) {
                receiptsStore.deleteReceipt(currentId).then(() => {
                    snackbar.value?.showMessage('Receipt deleted');
                    close();
                });
            } else {
                snackbar.value?.showMessage('Transaction deleted');
            }
        }).catch(error => {
            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    });
}

function onShowDateTimeError(error: string): void {
    snackbar.value?.showError(error);
}

defineExpose({
    open
});
</script>
