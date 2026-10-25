import { expect, test } from '@playwright/test';
import { display, openMode } from './app';

test('evaluates functions and cycles angle units', async ({ page }) => {
	await openMode(page, 'Scientific');
	const screen = display(page);

	await page.getByRole('button', { name: '√', exact: true }).click();
	await page.getByRole('button', { name: '9', exact: true }).click();
	await page.getByRole('button', { name: ')', exact: true }).click();
	await page.getByRole('button', { name: 'Equals' }).click();

	await expect(screen.locator('.expr')).toContainText('√(9) =');
	await expect(screen.locator('.result')).toHaveText('3');

	const angle = page.getByRole('button', { name: /Angle unit:/ });
	await expect(angle).toHaveAccessibleName('Angle unit: deg');
	await angle.click();
	await expect(angle).toHaveAccessibleName('Angle unit: rad');
	await angle.click();
	await expect(angle).toHaveAccessibleName('Angle unit: grad');
	await angle.click();
	await expect(angle).toHaveAccessibleName('Angle unit: deg');

	await page.getByRole('button', { name: 'Second function' }).click();
	await expect(page.getByRole('button', { name: 'sin⁻¹' })).toBeVisible();
});
