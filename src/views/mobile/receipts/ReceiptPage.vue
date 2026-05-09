<template>
    <f7-page @page:afterin="onPageAfterIn" @page:beforeout="onPageBeforeOut">
        <f7-navbar>
            <f7-nav-left :class="{ 'disabled': loading }" :back-link="tt('Back')"></f7-nav-left>
            <f7-nav-title :title="tt('Receipt')"></f7-nav-title>
            <f7-nav-right :class="{ 'navbar-compact-icons': true, 'disabled': loading }">
                <f7-link icon-f7="checkmark_alt" :class="{ 'disabled': submitting }" @click="save" v-if="!isCreateMode"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-list strong inset dividers class="margin-vertical skeleton-text" v-if="loading">
            <f7-list-input label="Description" placeholder="Receipt Description"></f7-list-input>
            <f7-list-item header="Account" title="Account Name"></f7-list-item>
            <f7-list-item header="Time" title="YYYY/MM/DD HH:mm"></f7-list-item>
            <f7-list-input type="textarea" label="Comment" placeholder="Receipt Comment"></f7-list-input>
        </f7-list>

        <f7-list form strong inset dividers class="margin-vertical" v-else-if="receipt">
            <f7-list-input
                type="text"
                clear-button
                :label="tt('Description')"
                :placeholder="tt('Receipt Description')"
                :disabled="loading || submitting"
                v-model:value="receipt.place"
            ></f7-list-input>

            <f7-list-item
                class="list-item-with-header-and-title"
                :header="tt('Account')"
                :title="accountName"
                link="#"
                no-chevron
                :disabled="loading || submitting"
                @click="showAccountPicker = true"
            >
                <two-column-list-item-selection-sheet primary-key-field="id" primary-value-field="category"
                    primary-title-field="name" primary-footer-field="displayBalance"
                    primary-icon-field="icon" primary-icon-type="account"
                    primary-sub-items-field="accounts"
                    :primary-title-i18n="true"
                    secondary-key-field="id" secondary-value-field="id"
                    secondary-title-field="name" secondary-footer-field="displayBalance"
                    secondary-icon-field="icon" secondary-icon-type="account" secondary-color-field="color"
                    :enable-filter="true" :filter-placeholder="tt('Find account')" :filter-no-items-text="tt('No available account')"
                    :items="accountSelectionItems"
                    v-model:show="showAccountPicker"
                    v-model="receipt.accountId">
                </two-column-list-item-selection-sheet>
            </f7-list-item>

            <f7-list-item
                class="transaction-edit-datetime list-item-with-header-and-title"
                link="#" no-chevron
            >
                <template #header>
                    <div @click="showDateTimeDialog('time')">{{ tt('Time') }}</div>
                </template>
                <template #title>
                    <div class="transaction-edit-datetime-title">
                        <div @click="showDateTimeDialog('date')">{{ receiptDisplayDate }}</div>&nbsp;<div class="transaction-edit-datetime-time" @click="showDateTimeDialog('time')">{{ receiptDisplayTime }}</div>
                    </div>
                </template>
                <date-time-selection-sheet :init-mode="receiptDateTimeSheetMode"
                                           :timezone-utc-offset="receipt.utcOffset"
                                           :model-value="receipt.time"
                                           v-model:show="showReceiptDateTimeSheet"
                                           @update:model-value="updateReceiptTime">
                </date-time-selection-sheet>
            </f7-list-item>

            <f7-list-input
                type="textarea"
                clear-button
                :label="tt('Comment')"
                :placeholder="tt('Receipt Comment')"
                :disabled="loading || submitting"
                v-model:value="receipt.comment"
            ></f7-list-input>

            <f7-list-item>
                <template #header>
                    <span>{{ tt('Total Amount') }}</span>
                </template>
                <template #title>
                    <span :class="{ 'text-income': receipt.totalAmount > 0, 'text-expense': receipt.totalAmount < 0 }">
                        {{ formatAmount(Math.abs(receipt.totalAmount)) }}
                    </span>
                </template>
            </f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-vertical" v-if="receipt">
            <f7-list-item class="transactions-section-header list-item-with-header-and-title">
                <template #header>
                    <div class="transactions-header">
                        <span>{{ tt('Transactions') }}</span>
                        <f7-link :disabled="loading || submitting" @click="addTransaction">
                            {{ tt('Add') }}
                        </f7-link>
                    </div>
                </template>
            </f7-list-item>

            <f7-list-item
                :key="transaction.id"
                v-for="transaction in receipt.transactions || []"
                swipeout
                link="#"
                @click="editTransaction(transaction.id)"
                :disabled="loading"
            >
                <template #title>
                    <div class="transaction-receipt-comment">{{ transaction.comment || '-' }}</div>
                </template>
                <template #after>
                    <div class="transaction-receipt-amount-column">
                        <div :class="{ 'text-expense': transaction.type === TransactionType.Expense, 'text-income': transaction.type === TransactionType.Income, 'text-color-primary': transaction.type === TransactionType.Transfer }">
                            {{ formatAmount(transaction.sourceAmount) }}
                        </div>
                        <div class="transaction-receipt-quantity-row" v-if="transaction.quantity || transaction.unitPrice">
                            {{ formatQuantity(transaction.quantity) }} × {{ formatAmount(transaction.unitPrice) }}
                        </div>
                    </div>
                </template>
                <template #footer>
                    <span>{{ transaction.category?.name || '-' }}</span>
                </template>
                <f7-swipeout-actions right>
                    <f7-swipeout-button color="red" @click="removeTransaction(transaction.id)">
                        {{ tt('Remove') }}
                    </f7-swipeout-button>
                </f7-swipeout-actions>
            </f7-list-item>

            <f7-list-item v-if="!receipt.transactions || receipt.transactions.length < 1">
                <template #title>
                    <span class="text-gray">{{ tt('No data') }}</span>
                </template>
            </f7-list-item>
        </f7-list>
    </f7-page>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { useI18n } from '@/locales/helpers.ts';

import { useReceiptsStore } from '@/stores/receipt.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useSettingsStore } from '@/stores/setting.ts';

import { getCurrentUnixTime, getBrowserTimezoneOffsetMinutes, parseDateTimeFromUnixTimeWithTimezoneOffset } from '@/lib/datetime.ts';
import { TransactionType } from '@/core/transaction.ts';
import { Receipt, type ReceiptModifyRequest } from '@/models/receipt.ts';
import { useI18nUIComponents, showLoading, hideLoading } from '@/lib/ui/mobile.ts';
import services from '@/lib/services.ts';
import DateTimeSelectionSheet from '@/components/mobile/DateTimeSelectionSheet.vue';

const { tt, formatAmountToLocalizedNumerals, formatDateTimeToLongDate, formatDateTimeToLongTime, getCategorizedAccountsWithDisplayBalance } = useI18n();
const { showAlert } = useI18nUIComponents();

const receiptsStore = useReceiptsStore();
const accountsStore = useAccountsStore();
const settingsStore = useSettingsStore();

const props = defineProps<{
    f7router?: any;
}>();

const loading = ref<boolean>(true);
const submitting = ref<boolean>(false);

const receipt = ref<Receipt | null>(null);
const currentReceiptId = ref<string>('');
const isCreateMode = ref<boolean>(false);
const isReceiptPersisted = ref<boolean>(false);
const showAccountPicker = ref<boolean>(false);
const showReceiptDateTimeSheet = ref<boolean>(false);
const receiptDateTimeSheetMode = ref<string>('time');

const accountName = computed<string>(() => {
    if (!receipt.value?.accountId) return tt('None');
    const account = accountsStore.allAccountsMap[receipt.value.accountId];
    return account?.name || tt('None');
});

const accountSelectionItems = computed(() => {
    return getCategorizedAccountsWithDisplayBalance(
        accountsStore.allVisiblePlainAccounts,
        settingsStore.appSettings.showAccountBalance,
        settingsStore.appSettings.accountCategoryOrders
    );
});

const receiptDisplayDate = computed<string>(() => {
    if (!receipt.value) return '';
    const dateTime = parseDateTimeFromUnixTimeWithTimezoneOffset(receipt.value.time, receipt.value.utcOffset);
    return formatDateTimeToLongDate(dateTime);
});

const receiptDisplayTime = computed<string>(() => {
    if (!receipt.value) return '';
    const dateTime = parseDateTimeFromUnixTimeWithTimezoneOffset(receipt.value.time, receipt.value.utcOffset);
    return formatDateTimeToLongTime(dateTime);
});

function formatAmount(amount: number): string {
    return formatAmountToLocalizedNumerals(Math.abs(amount));
}

function formatQuantity(qty: number): string {
    return Number(qty / 1000).toLocaleString(undefined, {
        minimumFractionDigits: 0,
        maximumFractionDigits: 3
    });
}

function showDateTimeDialog(sheetMode: string): void {
    receiptDateTimeSheetMode.value = sheetMode;
    showReceiptDateTimeSheet.value = true;
}

function updateReceiptTime(newTime: number): void {
    if (receipt.value) {
        Object.assign(receipt.value, { time: newTime });
    }
}

function onPageAfterIn(): void {
    load();
}

function onPageBeforeOut(): void {
}

function load(): void {
    const query = props.f7router?.currentRoute?.query || {};
    const id = query.id as string || currentReceiptId.value || '';

    loading.value = true;

    accountsStore.loadAllAccounts({ force: false }).catch(() => {});

    if (!id) {
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
        isCreateMode.value = true;
        isReceiptPersisted.value = false;
        loading.value = false;
        return;
    }

    currentReceiptId.value = id;
    isCreateMode.value = false;
    isReceiptPersisted.value = true;

    receiptsStore.getReceiptById(id).then(data => {
        receipt.value = data;
        loading.value = false;
    }).catch(error => {
        loading.value = false;
        if (!error.processed) {
            showAlert('Unable to retrieve receipt');
        }
        props.f7router?.back();
    });
}

function save(): void {
    if (!receipt.value || !currentReceiptId.value) {
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
        props.f7router?.back();
    }).catch(error => {
        submitting.value = false;
        if (!error.processed) {
            showAlert(error.message || 'Unable to save receipt');
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
                showAlert('Unable to save receipt');
            }
            return;
        }
    }

    const currentId = currentReceiptId.value;

    if (!currentId) {
        return;
    }

    props.f7router?.navigate(`/transaction/add?receiptId=${currentId}`);
}

function editTransaction(transactionId: string): void {
    props.f7router?.navigate(`/transaction/edit?id=${transactionId}`);
}

function removeTransaction(transactionId: string): void {
    const currentId = currentReceiptId.value;

    if (!currentId) {
        return;
    }

    showLoading();
    services.deleteTransaction({ id: transactionId }).then(() => {
        return receiptsStore.getReceiptById(currentId);
    }).then(data => {
        receipt.value = data;
        hideLoading();

        if (!data.transactions || data.transactions.length === 0) {
            receiptsStore.deleteReceipt(currentId).then(() => {
                props.f7router?.back();
            });
        }
    }).catch(error => {
        hideLoading();
        if (!error.processed) {
            showAlert('Unable to remove transaction');
        }
    });
}
</script>

<style>
.transaction-receipt-amount-column {
    text-align: right;
}

.transaction-receipt-amount-column .transaction-receipt-quantity-row {
    font-size: var(--f7-list-item-footer-font-size);
    line-height: var(--f7-list-item-footer-line-height);
    white-space: nowrap;
}

.transaction-receipt-comment {
    overflow: hidden;
    text-overflow: ellipsis;
}

.transactions-section-header .item-content .item-inner .item-title {
    flex: 1;
    overflow: visible;
}

.transactions-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
}
</style>
