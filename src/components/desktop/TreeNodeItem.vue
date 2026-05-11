<template>
    <div>
        <v-list-item
            :class="{ 'tree-node-selected v-list-item--active text-primary': node['id'] === selectedId }"
            :active="node['id'] === selectedId"
            @click="onSelect"
        >
            <template #prepend>
                <v-btn v-if="hasChildren" density="compact" variant="text" :icon="true" size="small"
                       @click.stop="toggleExpanded">
                    <v-icon :icon="expanded ? mdiChevronDown : mdiChevronRight" size="20" />
                </v-btn>
                <span v-else class="ms-1" style="width: 28px; display: inline-block;"></span>
                <ItemIcon class="me-2" :icon-type="iconType"
                          :icon-id="node['icon']"
                          :color="node['color']">
                </ItemIcon>
            </template>
            <template #title>
                <div class="text-truncate" :style="{ paddingLeft: (depth * 4) + 'px' }">{{ nodeText }}</div>
            </template>
        </v-list-item>

        <template v-if="hasChildren && (expanded || showAllFiltered)">
            <tree-node-item
                v-for="child in filteredChildren"
                :key="child['id'] as string"
                :node="child"
                :selected-id="selectedId"
                :filter-text="filterText"
                :depth="depth + 1"
                @select="onChildSelect">
            </tree-node-item>
        </template>
    </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

import { useI18n } from '@/locales/helpers.ts';

import {
    mdiChevronRight,
    mdiChevronDown
} from '@mdi/js';

interface TreeNodeItemProps {
    node: Record<string, unknown>;
    selectedId: string;
    filterText?: string;
    depth: number;
}

const props = defineProps<TreeNodeItemProps>();

const emit = defineEmits<{
    (e: 'select', id: string): void;
}>();

const { ti } = useI18n();

const expanded = ref<boolean>(false);

const nodeText = computed<string>(() => ti(props.node['name'] as string, false));
const iconType = computed<string>(() => (props.node['iconType'] as string) || 'category');
const children = computed<Record<string, unknown>[]>(() => (props.node['children'] || props.node['subCategories'] || []) as Record<string, unknown>[]);
const hasChildren = computed<boolean>(() => children.value.length > 0);

const showAllFiltered = computed<boolean>(() => {
    return !!props.filterText && hasChildren.value;
});

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

function toggleExpanded(): void {
    expanded.value = !expanded.value;
}

function onSelect(): void {
    emit('select', props.node['id'] as string);
}

function onChildSelect(id: string): void {
    emit('select', id);
}
</script>
