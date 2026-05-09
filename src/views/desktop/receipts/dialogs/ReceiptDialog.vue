<template>
    <v-dialog width="600" persistent v-model="showState">
        <v-card class="pa-sm-1 pa-md-2">
            <template #title>
                <div class="d-flex align-center justify-center">
                    <h4 class="text-h4">{{ tt(title) }}</h4>
                    <v-progress-circular indeterminate size="22" class="ms-2" v-if="loading"></v-progress-circular>
                </div>
            </template>
            <v-card-text>
                <v-form class="mt-2">
                    <v-row>
                        <v-col cols="12">
                            <v-text-field
                                type="text"
                                persistent-placeholder
                                :disabled="loading || submitting"
                                :label="tt('Description')"
                                :placeholder="tt('Receipt Description')"
                                v-model="receipt.place"
                            />
                        </v-col>
                        <v-col cols="12">
                            <date-time-select
                                :disabled="loading || submitting"
                                :label="tt('Time')"
                                :timezone-utc-offset="receipt.utcOffset"
                                v-model="receipt.time"
                                @error="onShowDateTimeError"
                            />
                        </v-col>
                        <v-col cols="12">
                            <v-textarea
                                type="text"
                                persistent-placeholder
                                rows="3"
                                :disabled="loading || submitting"
                                :label="tt('Comment')"
                                :placeholder="tt('Receipt Comment')"
                                v-model="receipt.comment"
                            />
                        </v-col>
                    </v-row>
                </v-form>
            </v-card-text>
            <v-card-text>
                <div class="w-100 d-flex justify-center flex-wrap mt-sm-1 mt-md-2 gap-4">
                    <v-btn :disabled="loading || submitting" @click="save">
                        {{ tt('Save') }}
                        <v-progress-circular indeterminate size="22" class="ms-2" v-if="submitting"></v-progress-circular>
                    </v-btn>
                    <v-btn color="secondary" variant="tonal"
                           :disabled="loading || submitting" @click="cancel">{{ tt('Cancel') }}</v-btn>
                </div>
            </v-card-text>
        </v-card>
    </v-dialog>

    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackbar" />
</template>

<script setup lang="ts">
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import DateTimeSelect from '@/components/desktop/DateTimeSelect.vue';

import { ref, useTemplateRef } from 'vue';

import { useI18n } from '@/locales/helpers.ts';

import { useReceiptsStore } from '@/stores/receipt.ts';

import { Receipt, type ReceiptCreateRequest, type ReceiptModifyRequest } from '@/models/receipt.ts';

import { getCurrentUnixTime, getBrowserTimezoneOffsetMinutes } from '@/lib/datetime.ts';

type ConfirmDialogType = InstanceType<typeof ConfirmDialog>;
type SnackBarType = InstanceType<typeof SnackBar>;

const { tt } = useI18n();

const receiptsStore = useReceiptsStore();

const confirmDialog = useTemplateRef<ConfirmDialogType>('confirmDialog');
const snackbar = useTemplateRef<SnackBarType>('snackbar');

const showState = ref<boolean>(false);
const loading = ref<boolean>(false);
const submitting = ref<boolean>(false);

const receipt = ref<Receipt>(createEmptyReceipt());
const editReceiptId = ref<string | null>(null);

interface ReceiptEditResponse {
    message: string;
}

let resolveFunc: ((value: ReceiptEditResponse) => void) | null = null;
let rejectFunc: ((reason?: unknown) => void) | null = null;

const title = ref<string>('Add Receipt');

function createEmptyReceipt(): Receipt {
    const now = getCurrentUnixTime();
    return new Receipt({
        id: '',
        time: now,
        utcOffset: getBrowserTimezoneOffsetMinutes(now),
        place: '',
        comment: '',
        totalAmount: 0
    });
}

function open(options?: { id?: string }): Promise<ReceiptEditResponse> {
    showState.value = true;
    loading.value = true;
    submitting.value = false;

    receipt.value = createEmptyReceipt();
    editReceiptId.value = null;
    title.value = 'Add Receipt';

    if (options && options.id) {
        title.value = 'Edit Receipt';
        editReceiptId.value = options.id;

        receiptsStore.getReceiptById(options.id).then(data => {
            receipt.value = data;
            loading.value = false;
        }).catch(error => {
            loading.value = false;
            showState.value = false;

            if (!error.processed && rejectFunc) {
                rejectFunc(error);
            }
        });
    } else {
        loading.value = false;
    }

    return new Promise<ReceiptEditResponse>((resolve, reject) => {
        resolveFunc = resolve;
        rejectFunc = reject;
    });
}

function save(): void {
    submitting.value = true;

    const savePromise = editReceiptId.value
        ? receiptsStore.modifyReceipt({
            id: editReceiptId.value,
            time: receipt.value.time,
            utcOffset: receipt.value.utcOffset,
            place: receipt.value.place,
            comment: receipt.value.comment
        } as ReceiptModifyRequest)
        : receiptsStore.addReceipt({
            time: receipt.value.time,
            utcOffset: receipt.value.utcOffset,
            place: receipt.value.place,
            comment: receipt.value.comment
        } as ReceiptCreateRequest);

    savePromise.then(() => {
        submitting.value = false;

        const message = editReceiptId.value ? 'You have saved this receipt' : 'You have added a new receipt';
        resolveFunc?.({ message });
        showState.value = false;
    }).catch(error => {
        submitting.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function cancel(): void {
    rejectFunc?.();
    showState.value = false;
}

function onShowDateTimeError(error: string): void {
    snackbar.value?.showError(error);
}

defineExpose({
    open
});
</script>
