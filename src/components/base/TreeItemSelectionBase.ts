import { ref } from 'vue';

export interface TreeNode {
    id: string;
    name: string;
    icon: string;
    iconType?: string;
    color: string;
    hidden?: boolean;
    children?: TreeNode[];
    [key: string]: unknown;
}

export interface TreeItemSelectionBaseProps {
    modelValue: string;
    items: TreeNode[];
    enableFilter?: boolean;
    filterPlaceholder?: string;
    filterNoItemsText?: string;
}

export function useTreeItemSelectionBase(props: TreeItemSelectionBaseProps) {
    const filterContent = ref<string>('');

    function matchesFilter(item: Record<string, unknown>, lowerCaseFilter: string, getNodeTitle: (node: Record<string, unknown>) => string): boolean {
        if (!lowerCaseFilter) {
            return true;
        }

        if (getNodeTitle(item).toLowerCase().includes(lowerCaseFilter)) {
            return true;
        }

        const children = (item['children'] || item['subCategories'] || []) as Record<string, unknown>[];
        for (const child of children) {
            if (matchesFilter(child, lowerCaseFilter, getNodeTitle)) {
                return true;
            }
        }

        return false;
    }

    function findNode(items: Record<string, unknown>[], id: string): Record<string, unknown> | null {
        for (const item of items) {
            if (item['id'] === id) {
                return item;
            }

            const children = (item['children'] || item['subCategories'] || []) as Record<string, unknown>[];
            const found = findNode(children, id);
            if (found) {
                return found;
            }
        }

        return null;
    }

    function buildPath(items: Record<string, unknown>[], id: string): Record<string, unknown>[] {
        const path: Record<string, unknown>[] = [];

        function search(nodes: Record<string, unknown>[], targetId: string, currentPath: Record<string, unknown>[]): boolean {
            for (const node of nodes) {
                currentPath.push(node);

                if (node['id'] === targetId) {
                    path.push(...currentPath);
                    return true;
                }

                const children = (node['children'] || node['subCategories'] || []) as Record<string, unknown>[];
                if (search(children, targetId, currentPath)) {
                    return true;
                }

                currentPath.pop();
            }

            return false;
        }

        search(items, id, []);
        return path;
    }

    return {
        filterContent,
        matchesFilter,
        findNode,
        buildPath
    };
}
