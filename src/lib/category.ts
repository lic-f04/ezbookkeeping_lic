import { reversed, keys, values } from '@/core/base.ts';
import { type LocalizedPresetCategory, CategoryType } from '@/core/category.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    type TransactionCategoryCreateRequest,
    type TransactionCategoryCreateWithSubCategories,
    TransactionCategory
} from '@/models/transaction_category.ts';

export function transactionTypeToCategoryType(transactionType: TransactionType): CategoryType | null {
    if (transactionType === TransactionType.Income) {
        return CategoryType.Income;
    } else if (transactionType === TransactionType.Expense) {
        return CategoryType.Expense;
    } else if (transactionType === TransactionType.Transfer) {
        return CategoryType.Transfer;
    } else {
        return null;
    }
}

export function categoryTypeToTransactionType(categoryType: CategoryType): TransactionType | null {
    if (categoryType === CategoryType.Income) {
        return TransactionType.Income;
    } else if (categoryType === CategoryType.Expense) {
        return TransactionType.Expense;
    } else if (categoryType === CategoryType.Transfer) {
        return TransactionType.Transfer;
    } else {
        return null;
    }
}

export function localizedPresetCategoryToTransactionCategoryCreateWithSubCategorys(presetCategory: LocalizedPresetCategory): TransactionCategoryCreateWithSubCategories {
    const subCategories: TransactionCategoryCreateRequest[] = [];

    for (const subPresetCategory of presetCategory.subCategories) {
        const subCategory: TransactionCategoryCreateRequest = {
            name: subPresetCategory.name,
            type: subPresetCategory.type,
            parentId: '0',
            icon: subPresetCategory.icon,
            color: subPresetCategory.color,
            comment: '',
            clientSessionId: ''
        };

        subCategories.push(subCategory);
    }

    const categoryWithSubCategories: TransactionCategoryCreateWithSubCategories = {
        name: presetCategory.name,
        type: presetCategory.type,
        icon: presetCategory.icon,
        color: presetCategory.color,
        subCategories: subCategories
    };

    return categoryWithSubCategories;
}

export function localizedPresetCategoriesToTransactionCategoryCreateWithSubCategories(presetCategories: LocalizedPresetCategory[]): TransactionCategoryCreateWithSubCategories[] {
    const categories: TransactionCategoryCreateWithSubCategories[] = [];

    for (const presetCategory of presetCategories) {
        const categoryWithSubCategories = localizedPresetCategoryToTransactionCategoryCreateWithSubCategorys(presetCategory);
        categories.push(categoryWithSubCategories);
    }

    return categories;
}

export function flattenCategoryTree(categories?: TransactionCategory[]): TransactionCategory[] {
    const result: TransactionCategory[] = [];

    if (!categories) {
        return result;
    }

    for (const category of categories) {
        result.push(category);

        if (category.subCategories) {
            result.push(...flattenCategoryTree(category.subCategories));
        }
    }

    return result;
}

export function getAllLeafCategories(categories?: TransactionCategory[]): TransactionCategory[] {
    const result: TransactionCategory[] = [];

    if (!categories) {
        return result;
    }

    for (const category of categories) {
        if (!category.subCategories || category.subCategories.length === 0) {
            result.push(category);
        } else {
            result.push(...getAllLeafCategories(category.subCategories));
        }
    }

    return result;
}

export function getSecondaryTransactionMapByName(allCategories?: TransactionCategory[]): Record<string, TransactionCategory> {
    const ret: Record<string, TransactionCategory> = {};

    if (!allCategories) {
        return ret;
    }

    const allFlat = flattenCategoryTree(allCategories);

    for (const category of allFlat) {
        ret[category.name] = category;
    }

    return ret;
}

export function getTransactionCategoryFullPathName(categoryId: string | null | undefined, allCategories?: TransactionCategory[]): string[] {
    if (!allCategories || !categoryId) {
        return [];
    }

    function search(nodes: TransactionCategory[], targetId: string, path: string[]): boolean {
        for (const node of nodes) {
            path.push(node.name);

            if (node.id === targetId) {
                return true;
            }

            if (node.subCategories && search(node.subCategories, targetId, path)) {
                return true;
            }

            path.pop();
        }

        return false;
    }

    const path: string[] = [];
    search(allCategories, categoryId, path);
    return path;
}

// backward compatibility
export function getTransactionPrimaryCategoryName(categoryId: string | null | undefined, allCategories?: TransactionCategory[]): string {
    const path = getTransactionCategoryFullPathName(categoryId, allCategories);
    return path.length > 0 ? path[0] as string : '';
}

export function getTransactionSecondaryCategoryName(categoryId: string | null | undefined, allCategories?: TransactionCategory[]): string {
    const path = getTransactionCategoryFullPathName(categoryId, allCategories);
    return path.length > 1 ? path[path.length - 1] as string : (path.length > 0 ? path[0] as string : '');
}

function filterCategoryTreeRecursive(categories: TransactionCategory[], lowercaseFilterContent: string, showHidden: boolean): TransactionCategory[] {
    const result: TransactionCategory[] = [];

    for (const category of categories) {
        if (!showHidden && category.hidden) {
            continue;
        }

        const categoryMatchesName = !lowercaseFilterContent || category.name.toLowerCase().includes(lowercaseFilterContent);
        let filteredChildren: TransactionCategory[] | undefined;

        if (category.subCategories && category.subCategories.length > 0) {
            filteredChildren = filterCategoryTreeRecursive(category.subCategories, lowercaseFilterContent, showHidden);
        }

        if (!categoryMatchesName && (!filteredChildren || filteredChildren.length < 1)) {
            continue;
        }

        const filteredCategory = category.clone();
        filteredCategory.subCategories = filteredChildren || [];
        result.push(filteredCategory);
    }

    return result;
}

export function filterTransactionCategories(allTransactionCategories: Record<number, TransactionCategory[]>, allowCategoryTypes?: Record<number, boolean>, allowCategoryName?: string, showHidden?: boolean): Record<string, TransactionCategory[]> {
    const ret: Record<string, TransactionCategory[]> = {};
    const hasAllowCategoryTypes = allowCategoryTypes
        && (allowCategoryTypes[CategoryType.Income]
            || allowCategoryTypes[CategoryType.Expense]
            || allowCategoryTypes[CategoryType.Transfer]);

    const allCategoryTypes = [ CategoryType.Income, CategoryType.Expense, CategoryType.Transfer ];
    const lowercaseFilterContent = allowCategoryName ? allowCategoryName.toLowerCase() : '';

    for (const categoryType of allCategoryTypes) {
        const allCategories = allTransactionCategories[categoryType];

        if (!allCategories || allCategories.length < 1) {
            continue;
        }

        if (hasAllowCategoryTypes && !allowCategoryTypes[categoryType]) {
            continue;
        }

        ret[`${categoryType}`] = filterCategoryTreeRecursive(allCategories, lowercaseFilterContent, showHidden ?? false);
    }

    return ret;
}

export function allVisibleCategoriesByType(allTransactionCategories: Record<number, TransactionCategory[]>, categoryType: number): TransactionCategory[] {
    const allCategories = allTransactionCategories[categoryType];

    if (!allCategories) {
        return [];
    }

    return flattenCategoryTree(allCategories).filter(c => !c.hidden);
}

export function allVisiblePrimaryTransactionCategoriesByType(allTransactionCategories: Record<number, TransactionCategory[]>, categoryType: number): TransactionCategory[] {
    const allCategories = allTransactionCategories[categoryType];
    const visibleCategories: TransactionCategory[] = [];

    if (!allCategories) {
        return visibleCategories;
    }

    for (const category of allCategories) {
        if (category.hidden) {
            continue;
        }

        visibleCategories.push(category);
    }

    return visibleCategories;
}

export function getFinalCategoryIdsByFilteredCategoryIds(allTransactionCategoriesMap: Record<number, TransactionCategory>, filteredCategoryIds: Record<string, boolean>): string {
    let finalCategoryIds = '';

    if (!allTransactionCategoriesMap) {
        return finalCategoryIds;
    }

    for (const category of values(allTransactionCategoriesMap)) {
        if (filteredCategoryIds && !isCategoryOrSubCategoriesAllChecked(category, filteredCategoryIds)) {
            continue;
        }

        if (finalCategoryIds.length > 0) {
            finalCategoryIds += ',';
        }

        finalCategoryIds += category.id;
    }

    return finalCategoryIds;
}

export function isCategoryIdAvailable(categories: TransactionCategory[], categoryId: string): boolean {
    if (!categories || !categories.length) {
        return false;
    }

    for (const category of categories) {
        if (category.hidden) {
            continue;
        }

        if (category.id === categoryId) {
            return true;
        }

        if (category.subCategories && isCategoryIdAvailable(category.subCategories, categoryId)) {
            return true;
        }
    }

    return false;
}

export function isSubCategoryIdAvailable(categories: TransactionCategory[], categoryId: string): boolean {
    return isCategoryIdAvailable(categories, categoryId);
}

export function getFirstVisibleCategoryId(categories?: TransactionCategory[]): string {
    if (!categories || !categories.length) {
        return '';
    }

    for (const category of categories) {
        if (category.hidden) {
            continue;
        }

        if (category.subCategories && category.subCategories.length > 0) {
            const firstChildId = getFirstVisibleCategoryId(category.subCategories);
            if (firstChildId) {
                return firstChildId;
            }
        }

        return category.id;
    }

    return '';
}

export function getFirstAvailableSubCategoryId(categories: TransactionCategory[], categoryId: string): string {
    if (!categories || !categories.length) {
        return '';
    }

    for (const category of categories) {
        if (category.hidden) {
            continue;
        }

        if (category.id === categoryId) {
            if (category.subCategories && category.subCategories.length > 0) {
                for (const child of category.subCategories) {
                    if (!child.hidden) {
                        return child.id;
                    }
                }
            }
            return '';
        }

        if (category.subCategories) {
            const found = getFirstAvailableSubCategoryId(category.subCategories, categoryId);
            if (found) {
                return found;
            }
        }
    }

    return '';
}

export function isNoAvailableCategory(categories: TransactionCategory[], showHidden: boolean): boolean {
    for (const category of categories) {
        if (showHidden || !category.hidden) {
            return false;
        }
    }

    return true;
}

export function getAvailableCategoryCount(categories: TransactionCategory[], showHidden: boolean): number {
    let count = 0;

    for (const category of categories) {
        if (showHidden || !category.hidden) {
            count++;
        }
    }

    return count;
}

export function getFirstShowingId(categories: TransactionCategory[], showHidden: boolean): string | null {
    for (const category of categories) {
        if (showHidden || !category.hidden) {
            return category.id;
        }
    }

    return null;
}

export function getLastShowingId(categories: TransactionCategory[], showHidden: boolean): string | null {
    for (const category of reversed(categories)) {
        if (showHidden || !category.hidden) {
            return category.id;
        }
    }

    return null;
}

export function selectAllSubCategories(filterCategoryIds: Record<string, boolean>, value: boolean, category?: TransactionCategory): void {
    if (!category || !category.subCategories || !category.subCategories.length) {
        return;
    }

    for (const subCategory of category.subCategories) {
        filterCategoryIds[subCategory.id] = value;
        selectAllSubCategories(filterCategoryIds, value, subCategory);
    }
}

export function selectAll(filterCategoryIds: Record<string, boolean>, allTransactionCategoriesMap: Record<string, TransactionCategory>): void {
    for (const categoryId of keys(filterCategoryIds)) {
        const category = allTransactionCategoriesMap[categoryId];

        if (category) {
            filterCategoryIds[category.id] = false;
        }
    }
}

export function selectNone(filterCategoryIds: Record<string, boolean>, allTransactionCategoriesMap: Record<string, TransactionCategory>): void {
    for (const categoryId of keys(filterCategoryIds)) {
        const category = allTransactionCategoriesMap[categoryId];

        if (category) {
            filterCategoryIds[category.id] = true;
        }
    }
}

export function selectInvert(filterCategoryIds: Record<string, boolean>, allTransactionCategoriesMap: Record<string, TransactionCategory>): void {
    for (const categoryId of keys(filterCategoryIds)) {
        const category = allTransactionCategoriesMap[categoryId];

        if (category) {
            filterCategoryIds[category.id] = !filterCategoryIds[category.id];
        }
    }
}

function areAllDescendantsChecked(category: TransactionCategory, filterCategoryIds: Record<string, boolean>): boolean {
    if (!category.subCategories || category.subCategories.length < 1) {
        return !filterCategoryIds[category.id];
    }

    for (const subCategory of category.subCategories) {
        if (filterCategoryIds[subCategory.id]) {
            return false;
        }

        if (!areAllDescendantsChecked(subCategory, filterCategoryIds)) {
            return false;
        }
    }

    return true;
}

export function isCategoryOrSubCategoriesAllChecked(category: TransactionCategory, filterCategoryIds: Record<string, boolean>): boolean {
    return areAllDescendantsChecked(category, filterCategoryIds);
}

export function isCategoryOrAllDescendantsChecked(category: TransactionCategory, filterCategoryIds: Record<string, boolean>): boolean {
    if (!category.subCategories || category.subCategories.length < 1) {
        return !filterCategoryIds[category.id];
    }

    for (const subCategory of category.subCategories) {
        if (filterCategoryIds[subCategory.id]) {
            return false;
        }

        if (!isCategoryOrAllDescendantsChecked(subCategory, filterCategoryIds)) {
            return false;
        }
    }

    return true;
}

function isAllSubTreeChecked(category: TransactionCategory, filterCategoryIds: Record<string, boolean>): boolean {
    if (!category.subCategories || category.subCategories.length < 1) {
        return !filterCategoryIds[category.id];
    }

    for (const subCategory of category.subCategories) {
        if (!isAllSubTreeChecked(subCategory, filterCategoryIds)) {
            return false;
        }
    }

    return true;
}

function countCheckedSubTree(category: TransactionCategory, filterCategoryIds: Record<string, boolean>): { checked: number; total: number } {
    if (!category.subCategories || category.subCategories.length < 1) {
        return {
            checked: filterCategoryIds[category.id] ? 0 : 1,
            total: 1
        };
    }

    let checked = 0;
    let total = 0;

    for (const subCategory of category.subCategories) {
        const result = countCheckedSubTree(subCategory, filterCategoryIds);
        checked += result.checked;
        total += result.total;
    }

    return { checked, total };
}

export function isSubCategoriesAllChecked(category: TransactionCategory, filterCategoryIds: Record<string, boolean>): boolean {
    if (!category.subCategories || category.subCategories.length < 1) {
        return false;
    }

    for (const subCategory of category.subCategories) {
        if (!isAllSubTreeChecked(subCategory, filterCategoryIds)) {
            return false;
        }
    }

    return true;
}

export function isSubCategoriesHasButNotAllChecked(category: TransactionCategory, filterCategoryIds: Record<string, boolean>): boolean {
    if (!category.subCategories || category.subCategories.length < 1) {
        return false;
    }

    let hasChecked = false;
    let hasUnchecked = false;

    for (const subCategory of category.subCategories) {
        const result = countCheckedSubTree(subCategory, filterCategoryIds);

        if (result.checked > 0) {
            hasChecked = true;
        }

        if (result.checked < result.total) {
            hasUnchecked = true;
        }

        if (hasChecked && hasUnchecked) {
            return true;
        }
    }

    return false;
}
