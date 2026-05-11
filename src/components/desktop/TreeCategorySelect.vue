<template>
    <v-select
        ref="mainSelect"
        persistent-placeholder
        :density="density"
        :variant="variant"
        :readonly="readonly"
        :disabled="disabled"
        :label="label"
        :menu-props="{ contentClass: 'tree-category-select-menu' }"
        v-model="selectedId"
        v-model:menu="menuState"
        @update:menu="onMenuStateChanged"
    >
        <template #selection>
            <div class="d-flex align-center text-truncate cursor-pointer">
                <span class="text-truncate" v-if="selectedPathText">{{ selectedPathText }}</span>
                <span class="text-truncate" v-if="!selectedId">{{ tt('None') }}</span>
            </div>
        </template>

        <template #no-data>
            <div class="mx-2 mt-2" v-if="enableFilter">
                <v-text-field eager ref="filterInput" density="compact"
                              :prepend-inner-icon="mdiMagnify"
                              :placeholder="filterPlaceholder"
                              v-model="filterContent"
                              @click:control="onInputFocused(filterInput, true)"
                              @update:focused="onInputFocused(filterInput, $event)"></v-text-field>
            </div>
            <div class="mx-4 my-3" v-show="!visibleItems.length">
                {{ filterNoItemsText }}
            </div>
            <div ref="dropdownMenu" class="tree-list-container" v-show="visibleItems.length">
                <v-list density="compact" class="pa-0">
                    <template v-for="item in visibleItems">
                        <tree-node-item
                            :node="item"
                            :selected-id="selectedId"
                            :filter-text="filterContent"
                            :depth="0"
                            @select="onNodeSelected">
                        </tree-node-item>
                    </template>
                </v-list>
            </div>
        </template>
    </v-select>
</template>

<script setup lang="ts">
import { VSelect } from 'vuetify/components/VSelect';
import { VTextField } from 'vuetify/components/VTextField';

import { ref, computed, useTemplateRef, nextTick } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import { useTreeItemSelectionBase } from '@/components/base/TreeItemSelectionBase.ts';
import TreeNodeItem from './TreeNodeItem.vue';

import { scrollToSelectedItem } from '@/lib/ui/common.ts';

import {
    mdiMagnify
} from '@mdi/js';
import type { ComponentDensity, InputVariant } from '@/lib/ui/desktop.ts';

interface TreeCategorySelectProps {
    density?: ComponentDensity;
    variant?: InputVariant;
    disabled?: boolean;
    readonly?: boolean;
    label?: string;
    enableFilter?: boolean;
    filterPlaceholder?: string;
    filterNoItemsText?: string;
    items: Record<string, unknown>[];
    modelValue: string;
}

const props = defineProps<TreeCategorySelectProps>();

const emit = defineEmits<{
    (e: 'update:modelValue', value: string): void;
}>();

const { tt, ti } = useI18n();

const {
    filterContent,
    matchesFilter,
    buildPath
} = useTreeItemSelectionBase(props as any);

const mainSelect = useTemplateRef<VSelect>('mainSelect');
const filterInput = useTemplateRef<VTextField>('filterInput');
const dropdownMenu = useTemplateRef<HTMLElement>('dropdownMenu');

const menuState = ref<boolean>(false);

const selectedId = computed<string>({
    get: () => props.modelValue,
    set: (value) => {
        menuState.value = false;
        emit('update:modelValue', value);
    }
});

const selectedPath = computed<Record<string, unknown>[]>(() => {
    if (!props.modelValue) {
        return [];
    }
    return buildPath(props.items, props.modelValue) || [];
});

const selectedPathText = computed<string>(() => {
    if (!selectedPath.value.length) {
        return '';
    }
    return selectedPath.value.map(n => n['name'] as string).join(' > ');
});

const visibleItems = computed<Record<string, unknown>[]>(() => {
    const result: Record<string, unknown>[] = [];
    const lowerFilter = filterContent.value?.toLowerCase() ?? '';

    for (const item of props.items) {
        if (item['hidden']) {
            continue;
        }

        if (!filterContent.value || matchesFilter(item, lowerFilter, (n) => ti(n['name'] as string, false))) {
            result.push(item);
        }
    }

    return result;
});

function onNodeSelected(id: string): void {
    selectedId.value = id;
}

function updateMenuPosition(): void {
    nextTick(() => {
        const mainSelectRect = mainSelect.value?.$el.getBoundingClientRect();
        const selectMenu = dropdownMenu.value?.parentElement?.parentElement;
        const selectMenuRect = selectMenu?.getBoundingClientRect();

        if (mainSelectRect && selectMenu && selectMenuRect) {
            const newTop = Math.round(mainSelectRect.top + mainSelectRect.height);

            if (newTop + selectMenuRect.height < document.documentElement.scrollHeight) {
                selectMenu.style.top = newTop + 'px';
            }
        }
    });
}

function onMenuStateChanged(state: boolean): void {
    if (state) {
        nextTick(() => {
            if (dropdownMenu.value && dropdownMenu.value.parentElement) {
                scrollToSelectedItem(dropdownMenu.value.parentElement, '.tree-list-container', '.tree-list-container', '.tree-node-selected');
            }
        });
    }
}

function onInputFocused(input: VTextField | null | undefined, focused: boolean): void {
    if (input && focused) {
        nextTick(() => {
            const inputEl = input.$el as HTMLElement;
            const innerInput = inputEl.querySelector('input');
            if (innerInput) {
                innerInput.focus();
            }
            updateMenuPosition();
        });
    }
}
</script>

<style>
.tree-category-select-menu {
    max-height: inherit !important;
}

.tree-category-select-menu > .v-list {
    padding: 0;
}

.tree-category-select-menu .tree-list-container {
    width: 100%;
    max-height: 350px;
    overflow-y: auto;
}
</style>
