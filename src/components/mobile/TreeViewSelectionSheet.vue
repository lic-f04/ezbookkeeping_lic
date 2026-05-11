<template>
    <f7-sheet ref="sheet" swipe-to-close swipe-handler=".swipe-handler"
              style="height: auto" :opened="show" @sheet:open="onSheetOpen" @sheet:closed="onSheetClosed">
        <f7-toolbar class="toolbar-with-swipe-handler">
            <div class="swipe-handler"></div>
            <div class="left">
                <f7-link sheet-close icon-f7="xmark"></f7-link>
            </div>
            <f7-searchbar ref="searchbar" custom-searchs
                          :value="filterContent"
                          :placeholder="filterPlaceholder"
                          :disable-button="false"
                          v-if="enableFilter"
                          @input="filterContent = $event.target.value"
                          @focus="onSearchBarFocus">
            </f7-searchbar>
        </f7-toolbar>
        <f7-page-content :class="'margin-top ' + heightClass">
            <f7-list class="no-margin-top no-margin-bottom" v-if="!filteredItems || !filteredItems.length">
                <f7-list-item :title="filterNoItemsText"></f7-list-item>
            </f7-list>
            <f7-treeview class="tree-view-selection-treeview">
                <template v-for="item in filteredItems">
                    <tree-view-node
                        :node="item"
                        :selected-id="currentValue"
                        :filter-text="filterContent"
                        :depth="0"
                        @select="onNodeClicked">
                    </tree-view-node>
                </template>
            </f7-treeview>
        </f7-page-content>
    </f7-sheet>
</template>

<script setup lang="ts">
import { ref, computed, useTemplateRef } from 'vue';
import type { Sheet, Searchbar } from 'framework7/types';

import { useI18n } from '@/locales/helpers.ts';
import TreeViewNode from './TreeViewNode.vue';

import { scrollToSelectedItem } from '@/lib/ui/common.ts';
import { type Framework7Dom, scrollSheetToTop } from '@/lib/ui/mobile.ts';

interface TreeViewSelectionSheetProps {
    show: boolean;
    enableFilter?: boolean;
    filterPlaceholder?: string;
    filterNoItemsText?: string;
    items: Record<string, unknown>[];
    modelValue: string;
}

const props = defineProps<TreeViewSelectionSheetProps>();

const emit = defineEmits<{
    (e: 'update:modelValue', value: string): void;
    (e: 'update:show', value: boolean): void;
}>();

const { ti } = useI18n();

const filterContent = ref<string>('');

const sheet = useTemplateRef<Sheet.Sheet>('sheet');
const searchbar = useTemplateRef<Searchbar.Searchbar>('searchbar');

const currentValue = ref<string>(props.modelValue);

const heightClass = computed<string>(() => {
    const count = visibleItemsCount.value;

    if (count > 6) {
        return 'tree-view-selection-huge-sheet';
    } else if (count > 2) {
        return 'tree-view-selection-large-sheet';
    } else {
        return 'tree-view-selection-default-sheet';
    }
});

const visibleItemsCount = computed<number>(() => {
    let count = 0;

    function countVisible(items: Record<string, unknown>[]): void {
        for (const item of items) {
            if (item['hidden']) {
                continue;
            }
            count++;
        }
    }

    countVisible(props.items);
    return count;
});

function matchesFilterRecursive(item: Record<string, unknown>, lowerFilter: string): boolean {
    if (!lowerFilter) {
        return true;
    }

    const name = ti(item['name'] as string, false).toLowerCase();
    if (name.includes(lowerFilter)) {
        return true;
    }

    const subItems = (item['children'] || item['subCategories'] || []) as Record<string, unknown>[];
    for (const sub of subItems) {
        if (matchesFilterRecursive(sub, lowerFilter)) {
            return true;
        }
    }

    return false;
}

const filteredItems = computed<Record<string, unknown>[]>(() => {
    if (!filterContent.value) {
        return props.items;
    }

    const lowerFilter = filterContent.value.toLowerCase();
    const result: Record<string, unknown>[] = [];

    for (const item of props.items) {
        if (item['hidden']) {
            continue;
        }

        if (matchesFilterRecursive(item, lowerFilter)) {
            result.push(item);
        }
    }

    return result;
});

function onNodeClicked(id: string): void {
    currentValue.value = id;
    emit('update:modelValue', id);
    emit('update:show', false);
}

function onSearchBarFocus(): void {
    scrollSheetToTop(sheet.value?.$el as HTMLElement, window.innerHeight);
}

function onSheetOpen(event: { $el: Framework7Dom }): void {
    currentValue.value = props.modelValue;
    scrollToSelectedItem(event.$el[0], '.sheet-modal-inner', '.page-content', '.treeview-item-selected');
}

function onSheetClosed(): void {
    emit('update:show', false);
    filterContent.value = '';
    searchbar.value?.clear();
}
</script>

<style>
.tree-view-selection-default-sheet {
    height: 310px;
}

@media (min-height: 630px) {
    .tree-view-selection-large-sheet {
        height: 370px;
    }

    .tree-view-selection-huge-sheet {
        height: 500px;
    }
}

@media (max-height: 629px) {
    .tree-view-selection-large-sheet,
    .tree-view-selection-huge-sheet {
        height: 320px;
    }
}

.tree-view-selection-treeview {
    position: relative;
}
</style>
