import { expect, test } from '@playwright/test';

test('renders the CVSE tool workspace', async ({ page }) => {
	await page.goto('/');

	await expect(page.getByText('CVSE xTower')).toBeVisible();
	await expect(page.getByLabel('视频标识')).toBeVisible();
	await expect(page.getByRole('button', { name: '解析预览' })).toBeVisible();
	await expect(page.getByText('v1 分区')).toBeVisible();
	await expect(page.getByText('v2 分区')).toBeVisible();
	await expect(page.getByText('统计指标')).toBeVisible();
	await expect(page.getByText('周刊分数')).toBeVisible();
});

test('accepts a video id in the resolver input', async ({ page }) => {
	await page.goto('/');

	const input = page.getByLabel('视频标识');
	await input.fill('av2');
	await page.getByRole('button', { name: '解析预览' }).click();

	await expect(input).toHaveValue('av2');
	await expect(page.getByText('preview ready')).toBeVisible();
});
