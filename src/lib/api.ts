import type { APIResponse, ResolveData } from '$lib/types';

/**
 * In dev (pnpm dev), Go backend runs on :8080.
 * In production, ESA proxies /api/ to Go, so same origin.
 */
const API_BASE = import.meta.env.DEV ? 'http://localhost:8080' : '';

// ── ESA AI Captcha ──

declare global {
	interface Window {
		AliyunCaptchaConfig?: { region: string; prefix: string };
		initAliyunCaptcha: (config: {
			SceneId: string;
			element: string;
			button?: string;
			mode?: 'popup' | 'embed' | 'float';
			success: (captchaVerifyParam: string) => void;
			fail: (result: any) => void;
			getInstance?: (instance: { show: () => void; refresh: () => void }) => void;
			server?: string[];
			slideStyle?: { width: number; height: number };
		}) => void;
	}
}

// 缓存的 captcha 实例，初始化一次，后续重复使用
let captchaInstance: { show: () => void; refresh: () => void } | null = null;
let captchaReady = false;

/**
 * 初始化 ESA 验证码（在页面加载后调用一次即可）
 */
export function initCaptcha(sceneId: string) {
	if (!import.meta.env.DEV && typeof window.initAliyunCaptcha === 'function') {
		window.initAliyunCaptcha({
			SceneId: sceneId,
			mode: 'popup',
			element: '#captcha-element',
			server: ['captcha-esa-open.aliyuncs.com', 'captcha-esa-open-b.aliyuncs.com'],
			slideStyle: { width: 360, height: 40 },
			getInstance: (instance) => {
				captchaInstance = instance;
				captchaReady = true;
			},
			success: (captchaVerifyParam: string) => {
				// success 回调由 show() 触发，这里只存 token
				// 实际的 API 请求在外部 await 后执行
				pendingResolve?.(captchaVerifyParam);
				pendingResolve = null;
				captchaInstance?.refresh();
			},
			fail: () => {
				pendingReject?.(new Error('验证未通过'));
				pendingReject = null;
			}
		});
	}
}

// 用于在 success/fail 回调间传递 token 的临时变量
let pendingResolve: ((token: string) => void) | null = null;
let pendingReject: ((err: Error) => void) | null = null;

/**
 * 触发验证码弹窗，返回 captchaVerifyParam token
 */
function getCaptchaToken(): Promise<string> {
	return new Promise((resolve, reject) => {
		if (!captchaInstance) {
			// 开发环境或 captcha 未就绪时跳过验证
			resolve('');
			return;
		}

		pendingResolve = resolve;
		pendingReject = reject;
		captchaInstance.show();
	});
}

// ── API ──

/**
 * Resolve a Bilibili video by its identifier (BV, av, URL).
 */
export async function resolveVideo(input: string): Promise<ResolveData> {
	const url = `${API_BASE}/api/video/resolve?input=${encodeURIComponent(input)}`;

	// 生产环境：先拿验证码 token，再发请求
	let captchaToken = '';
	if (!import.meta.env.DEV) {
		captchaToken = await getCaptchaToken();
	}

	// 发起业务请求，token 放在 query param 中（ESA 推荐方式）
	const requestUrl = captchaToken
		? `${url}&captcha_verify_param=${encodeURIComponent(captchaToken)}`
		: url;

	const res = await fetch(requestUrl);
	if (!res.ok) {
		const body = await res.json().catch(() => null);
		throw new Error(body?.message ?? `HTTP ${res.status}`);
	}

	// ESA 验证通过后会在响应头返回 X-Captcha-Verify-Code
	const verifyCode = res.headers.get('X-Captcha-Verify-Code');
	if (verifyCode && verifyCode !== 'T001') {
		throw new Error(`验证码验证失败(${verifyCode})`);
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
