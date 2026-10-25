import { expect, test } from '@playwright/test';
import { display, historyDialog, openApp, openMode, resetHistory } from './app';

test('records, persists, deletes and clears history entries', async ({ page }) => {
	await openMode(page, 'Normal');
	await resetHistory(page);

	await page.getByRole('button', { name: '9', exact: true }).click();
	await page.getByRole('button', { name: 'Multiply' }).click();
	await page.getByRole('button', { name: '9', exact: true }).click();
	await page.getByRole('button', { name: 'Equals' }).click();
	await expect(display(page).locator('.result')).toHaveText('81');

	await page.reload();
	await openApp(page);
	await page.getByRole('button', { name: 'Calculation history' }).click();
	const dialog = historyDialog(page);
	await expect(dialog.getByText('9×9', { exact: true })).toBeVisible();
	await expect(dialog.getByText('= 81', { exact: true })).toBeVisible();
	await expect(dialog.getByText('1 saved entry', { exact: true })).toBeVisible();

	await dialog.getByRole('button', { name: 'Delete 9×9' }).click();
	await expect(dialog.getByText('No calculations yet')).toBeVisible();

	await dialog.getByRole('button', { name: 'Close history' }).click();
	await page.getByRole('button', { name: '1', exact: true }).click();
	await page.getByRole('button', { name: 'Add', exact: true }).click();
	await page.getByRole('button', { name: '1', exact: true }).click();
	await page.getByRole('button', { name: 'Equals' }).click();
	await expect(display(page).locator('.result')).toHaveText('2');

	await page.getByRole('button', { name: 'Calculation history' }).click();
	await dialog.getByRole('button', { name: 'Clear all' }).click();
	await expect(dialog.getByText('No calculations yet')).toBeVisible();
});
