import type { APIResponse, ResolveData } from '$lib/types';

/**
 * In dev (pnpm dev), Go backend runs on :8080.
 * In production, Nginx proxies /api/ to Go, so same origin.
 */
const API_BASE = import.meta.env.DEV ? 'http://localhost:8080' : '';

/**
 * Resolve a Bilibili video by its identifier (BV, av, URL).
 */
export async function resolveVideo(input: string): Promise<ResolveData> {
	const url = `${API_BASE}/api/video/resolve?input=${encodeURIComponent(input)}`;

	const res = await fetch(url);
	if (!res.ok) {
		const body = await res.json().catch(() => null);
		throw new Error(body?.message ?? `HTTP ${res.status}`);
	}

	const json: APIResponse = await res.json();
	if (json.code !== 0 || !json.data) {
		throw new Error(json.message || 'API returned error');
	}

	return json.data;
}

/**
 * Format a large number with commas.
 */
export function fmtNum(n: number): string {
	return n.toLocaleString('en-US');
}

/**
 * Convert a Unix timestamp (seconds) to a readable string.
 */
export function fmtTime(ts: number): string {
	if (!ts) return '--';
	const d = new Date(ts * 1000);
	return d.toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai' }) + ' CST';
}
