<template>
    <f7-treeview-item
        :selectable="true"
        :selected="node['id'] === selectedId"
        :label="nodeText"
        :opened="isExpanded"
        @click="onSelect"
    >
        <template #media>
            <ItemIcon icon-type="category" :icon-id="node['icon']" :color="node['color']"></ItemIcon>
        </template>

        <template v-if="hasChildren">
            <tree-view-node
                v-for="child in filteredChildren"
                :key="child['id'] as string"
                :node="child"
                :selected-id="selectedId"
                :filter-text="filterText"
                :depth="depth + 1"
                @select="onChildSelect">
            </tree-view-node>
        </template>
    </f7-treeview-item>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

import { useI18n } from '@/locales/helpers.ts';

interface TreeViewNodeProps {
    node: Record<string, unknown>;
    selectedId: string;
    filterText?: string;
    depth: number;
}

const props = defineProps<TreeViewNodeProps>();

const emit = defineEmits<{
    (e: 'select', id: string): void;
}>();

const { ti } = useI18n();

const isExpanded = ref<boolean>(true);

const nodeText = computed<string>(() => ti(props.node['name'] as string, false));
const children = computed<Record<string, unknown>[]>(() => (props.node['children'] || props.node['subCategories'] || []) as Record<string, unknown>[]);
const hasChildren = computed<boolean>(() => children.value.length > 0);

const filteredChildren = computed<Record<string, unknown>[]>(() => {
    if (!props.filterText) {
        return children.value;
    }

    const lowerFilter = props.filterText.toLowerCase();
    const result: Record<string, unknown>[] = [];

    for (const child of children.value) {
        if (matchesFilterRecursive(child, lowerFilter)) {
            result.push(child);
        }
    }

    return result;
});

function matchesFilterRecursive(item: Record<string, unknown>, lowerFilter: string): boolean {
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

function onSelect(): void {
    emit('select', props.node['id'] as string);
}

function onChildSelect(id: string): void {
    emit('select', id);
}
</script>
