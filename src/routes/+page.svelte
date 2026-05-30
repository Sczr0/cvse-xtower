<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { resolveVideo, fmtNum, fmtTime } from '$lib/api';
	import type { ResolveData } from '$lib/types';

	let { children } = $props();

	let videoInput = $state('');
	let loading = $state(false);
	let errorMsg = $state('');
	let result = $state<ResolveData | null>(null);
	let showDisclaimer = $state(false);

	const doResolve = async () => {
		const input = videoInput.trim();
		if (!input) return;

		loading = true;
		errorMsg = '';
		result = null;

		try {
			const data = await resolveVideo(input);
			result = data;
		} catch (e) {
			errorMsg = e instanceof Error ? e.message : '请求失败，请检查输入后重试';
		} finally {
			loading = false;
		}
	};

	const handleKeydown = (event: KeyboardEvent) => {
		if (event.key === 'Enter') doResolve();
	};
</script>

<svelte:head>
	<title>CVSE xTower — B 站视频分区查询与周刊分数</title>
	<meta name="description" content="查询 B 站视频的分区信息与周刊分数，由 CVSE 团队制作。" />
</svelte:head>

<div class="min-h-screen bg-[#fafafa] text-[#0a0a0a]">
	<div class="mx-auto flex min-h-screen w-full max-w-5xl flex-col px-4 py-6 sm:px-8 lg:px-12">
		<!-- ── Header ── -->
		<header class="flex items-center gap-3 pb-6">
			<div
				class="grid size-9 place-items-center rounded-lg border border-[#d4d4d4] bg-white text-sm font-semibold tracking-tight"
				aria-hidden="true"
			>
				CV
			</div>
			<div>
				<h1 class="text-sm font-semibold leading-5">CVSE xTower</h1>
				<p class="text-xs leading-5 text-[#666]">B 站视频分区与周刊分数查询</p>
			</div>
		</header>

		<!-- ── Search ── -->
		<section class="rounded-xl border border-[#e5e5e5] bg-white px-5 py-5 shadow-sm sm:px-6">
			<div class="flex flex-col gap-2 sm:flex-row">
				<input
					id="video-input"
					class="h-11 min-w-0 flex-1 rounded-lg border border-[#d4d4d4] bg-white px-4 text-sm text-[#111] outline-none transition focus:border-[#0a0a0a] focus:ring-1 focus:ring-[#0a0a0a]"
					placeholder="输入 BV 号、av 号或视频链接…"
					bind:value={videoInput}
					onkeydown={handleKeydown}
				/>
				<button
					type="button"
					class="flex h-11 shrink-0 items-center gap-2 rounded-lg bg-[#0a0a0a] px-5 text-sm font-medium text-white transition hover:bg-[#222] disabled:cursor-not-allowed disabled:opacity-40"
					onclick={doResolve}
					disabled={loading || !videoInput.trim()}
				>
					{#if loading}
						<svg class="h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z" />
						</svg>
						<span>查询中</span>
					{:else}
						<span>查询</span>
					{/if}
				</button>
			</div>

			{#if errorMsg}
				<div class="mt-3 rounded-lg border border-[#f5c2c7] bg-[#fff5f5] px-4 py-2.5 text-sm text-[#8a1f2d]">
					{errorMsg}
				</div>
			{/if}
		</section>

		<!-- ── Results ── -->
		{#if result}
			<div class="mt-5 space-y-5">
				<!-- Video info -->
				<section class="rounded-xl border border-[#e5e5e5] bg-white px-5 py-5 shadow-sm sm:px-6">
					<div class="flex items-start gap-4">
						{#if result.pic}
							<div class="hidden shrink-0 sm:block">
								<img
									src={result.pic}
									alt=""
									class="h-24 w-40 rounded-lg border border-[#ededed] object-cover"
									referrerpolicy="no-referrer"
								/>
							</div>
						{/if}
						<div class="min-w-0 flex-1">
							<div class="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-[#888]">
								<span class="font-mono">{result.bvid}</span>
								<span class="hidden sm:inline">·</span>
								<span class="font-mono">av{result.aid}</span>
							</div>
							<h2 class="mt-1.5 text-lg font-semibold leading-6 sm:text-xl">
								{result.title}
							</h2>
							<p class="mt-1 text-sm text-[#555]">
								UP 主：<span class="font-medium text-[#0a0a0a]">{result.owner.name}</span>
							</p>
						</div>
					</div>
				</section>

				<!-- Categories + Score row -->
				<div class="grid gap-5 sm:grid-cols-[1fr_280px]">
					<div class="flex flex-col gap-5">
						<!-- Categories -->
						<section class="rounded-xl border border-[#e5e5e5] bg-white px-5 py-5 shadow-sm sm:px-6">
							<h3 class="text-sm font-semibold">视频分区</h3>
							<div class="mt-4 grid gap-4 sm:grid-cols-2">
								<div>
									<p class="text-xs text-[#888]">v1 分区</p>
									<p class="mt-1 text-xl font-semibold">{result.v1.name}</p>
									<p class="mt-0.5 text-xs text-[#888]">{result.v1.mainName}</p>
								</div>
								<div>
									<p class="text-xs text-[#888]">v2 分区</p>
									<p class="mt-1 text-xl font-semibold">{result.v2?.name ?? '无'}</p>
									<p class="mt-0.5 text-xs text-[#888]">{result.v2?.mainName ?? ''}</p>
								</div>
							</div>
						</section>
						<!-- Timeline -->
						<section class="rounded-xl border border-[#e5e5e5] bg-white px-5 py-3.5 shadow-sm sm:px-6">
							<div class="flex flex-wrap gap-x-6 gap-y-1 text-xs text-[#888]">
								<span>发布时间：<span class="font-medium text-[#0a0a0a]">{fmtTime(result.pubdate)}</span></span>
								<span>投稿时间：<span class="font-medium text-[#0a0a0a]">{fmtTime(result.ctime)}</span></span>
							</div>
						</section>
					</div>

					<!-- Score -->
					<section class="rounded-xl border border-[#e5e5e5] bg-white px-5 py-5 shadow-sm sm:px-6">
						<h3 class="text-sm font-semibold">周刊分数
							{#if result?.score}
								<span class="group relative ml-1.5 inline-flex cursor-help align-middle">
									<svg class="h-3.5 w-3.5 text-[#bbb] hover:text-[#888]" viewBox="0 0 16 16" fill="currentColor">
										<path d="M8 1a7 7 0 110 14A7 7 0 018 1zm0 1.5a5.5 5.5 0 100 11 5.5 5.5 0 000-11zM7.5 7v4.5h1V7h-1zm0-2.5V6h1V4.5h-1z"/>
									</svg>
									<div class="pointer-events-none absolute bottom-full left-1/2 z-10 mb-2 w-52 -translate-x-1/2 rounded-lg border border-[#e5e5e5] bg-white px-3 py-2.5 text-xs text-[#555] shadow-lg opacity-0 transition group-hover:opacity-100">
										<div class="space-y-1.5">
											<div class="flex justify-between gap-2">
												<span>得点A</span>
												<span class="font-mono text-[#0a0a0a]">{result.score.rawA.toFixed(1)}</span>
											</div>
											<div class="flex justify-between gap-2">
												<span>修正A</span>
												<span class="font-mono text-[#0a0a0a]">{result.score.correctionA.toFixed(4)}</span>
											</div>
											<p class="text-[10px] leading-tight text-[#aaa]">收录时间-投稿时间≤14天，修正A下限为1</p>
											<div class="flex justify-between gap-2 border-t border-[#f0f0f0] pt-1.5">
												<span>得点B</span>
												<span class="font-mono text-[#0a0a0a]">{result.score.rawB.toFixed(1)}</span>
											</div>
											<div class="flex justify-between gap-2">
												<span>修正B</span>
												<span class="font-mono text-[#0a0a0a]">{result.score.correctionB.toFixed(4)}</span>
											</div>
											<div class="flex justify-between gap-2 border-t border-[#f0f0f0] pt-1.5">
												<span>得点C</span>
												<span class="font-mono text-[#0a0a0a]">{result.score.rawC.toFixed(1)}</span>
											</div>
											<div class="flex justify-between gap-2">
												<span>修正C</span>
												<span class="font-mono text-[#0a0a0a]">{result.score.correctionC.toFixed(4)}</span>
											</div>
										</div>
										<div class="absolute left-1/2 top-full -translate-x-1/2 border-4 border-transparent border-t-white"></div>
									</div>
								</span>
							{/if}
						</h3>
						<div class="mt-3 flex items-baseline gap-1.5">
							<span class="text-3xl font-bold tracking-tight">
								{result.score?.total ?? '--'}
							</span>
							<span class="text-xs text-[#888]">Pt.</span>
						</div>
						{#if result.score}
							<div class="mt-4 space-y-2.5 border-t border-[#f0f0f0] pt-4 text-xs">
								<div class="flex items-center justify-between">
									<span class="text-[#555]">播放行为</span>
									<span class="font-mono text-[#0a0a0a]">{result.score.playScore.toFixed(1)}</span>
								</div>
								<div class="flex items-center justify-between">
									<span class="text-[#555]">转化表现</span>
									<span class="font-mono text-[#0a0a0a]">{result.score.conversionScore.toFixed(1)}</span>
								</div>
								<div class="flex items-center justify-between">
									<span class="text-[#555]">互动</span>
									<span class="font-mono text-[#0a0a0a]">{result.score.interactionScore.toFixed(1)}</span>
								</div>
							</div>
						{/if}
					</section>
				</div>

				<!-- Stats -->
				<section class="rounded-xl border border-[#e5e5e5] bg-white shadow-sm overflow-hidden">
					<div class="border-b border-[#f0f0f0] px-5 py-3.5 sm:px-6">
						<h3 class="text-sm font-semibold">数据统计</h3>
					</div>
					<div class="grid grid-cols-3 gap-px bg-[#f0f0f0] sm:grid-cols-6 min-w-0">
						<div class="min-w-0 bg-white px-5 py-4 sm:px-6">
							<p class="text-xs text-[#888]">播放</p>
							<p class="mt-1.5 text-base font-semibold tracking-tight sm:text-xl truncate">{fmtNum(result.stat.view)}</p>
						</div>
						<div class="min-w-0 bg-white px-5 py-4 sm:px-6">
							<p class="text-xs text-[#888]">弹幕</p>
							<p class="mt-1.5 text-base font-semibold tracking-tight sm:text-xl truncate">{fmtNum(result.stat.danmaku)}</p>
						</div>
						<div class="min-w-0 bg-white px-5 py-4 sm:px-6">
							<p class="text-xs text-[#888]">评论</p>
							<p class="mt-1.5 text-base font-semibold tracking-tight sm:text-xl truncate">{fmtNum(result.stat.reply)}</p>
						</div>
						<div class="min-w-0 bg-white px-5 py-4 sm:px-6">
							<p class="text-xs text-[#888]">点赞</p>
							<p class="mt-1.5 text-base font-semibold tracking-tight sm:text-xl truncate">{fmtNum(result.stat.like)}</p>
						</div>
						<div class="min-w-0 bg-white px-5 py-4 sm:px-6">
							<p class="text-xs text-[#888]">投币</p>
							<p class="mt-1.5 text-base font-semibold tracking-tight sm:text-xl truncate">{fmtNum(result.stat.coin)}</p>
						</div>
						<div class="min-w-0 bg-white px-5 py-4 sm:px-6">
							<p class="text-xs text-[#888]">收藏</p>
							<p class="mt-1.5 text-base font-semibold tracking-tight sm:text-xl truncate">{fmtNum(result.stat.favorite)}</p>
						</div>
					</div>
				</section>

				</div>
		{:else if !loading}
			<!-- Empty state -->
			<div class="mt-20 text-center">
				<svg class="mx-auto h-10 w-10 text-[#ccc]" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
					<path stroke-linecap="round" stroke-linejoin="round" d="M15.75 10.5l4.72-4.72a.75.75 0 011.28.53v11.38a.75.75 0 01-1.28.53l-4.72-4.72M4.5 18.75h9a2.25 2.25 0 002.25-2.25v-9a2.25 2.25 0 00-2.25-2.25h-9A2.25 2.25 0 002.25 7.5v9a2.25 2.25 0 002.25 2.25z" />
				</svg>
				<p class="mt-4 text-sm text-[#999]">输入视频标识，查询分区与分数</p>
			</div>
		{/if}

		<!-- Footer -->
		<footer class="mt-auto border-t border-[#f0f0f0] pb-2 pt-6 text-center">
			<p class="text-xs text-[#bbb]">数据来源：BiliBili · <a href="https://cvse.cc/" target="_blank" rel="noopener noreferrer" class="underline underline-offset-2 hover:text-[#888]">CVSE 野鸽社</a> · <button type="button" onclick={() => showDisclaimer = true} class="underline underline-offset-2 hover:text-[#888] cursor-pointer">免责声明</button></p>
			<p class="mt-2 text-[11px] text-[#ccc]">使用第三方观测的数据计算得到的分数可能与周刊的分数存在出入<br>计算结果仅供参考，最终数据以周刊公布的数据为准</p>
		</footer>
	</div>
</div>

<!-- Disclaimer modal -->
{#if showDisclaimer}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 px-4"
		onclick={() => showDisclaimer = false}
	>
		<div
			class="max-w-lg rounded-xl border border-[#e5e5e5] bg-white px-6 py-5 shadow-lg sm:px-8 sm:py-6"
			onclick={(e) => e.stopPropagation()}
		>
			<h3 class="text-sm font-semibold">免责声明</h3>
			<div class="mt-3 space-y-3 text-xs leading-6 text-[#555]">
				<p>我们及我们旗下所有排行榜、网站等的数据均为公开，可由公众任意访问，当某人向您展示我们旗下作品（无论是完整版还是截取其中某一部分）、网站、数据表等，均无法证明其为我们的成员。</p>
				<p>即使是我们的成员，其发表的个人言论也不能代我们的任何政策或立场，亦不能理解为是对任何规则的官方解释。</p>
				<p>我们与作品内收录、展示的任何投稿均无任何关联，作品内收录或展示的任何投稿均不代表本刊立场。即使我们的成员甚至我们参与了相关作品，也不能代表我们的任何立场或观点。</p>
				<p class="text-[#bbb]">——<a href="https://cvse.cc/posts/ef4f077d/#:~:text=%E5%90%8D%E5%8D%95%E9%83%A8%E5%88%86%E6%9F%A5%E7%9C%8B%E3%80%82-,%E7%AC%AC%E4%BA%94%E7%AB%A0%20%E5%85%8D%E8%B4%A3,-%E6%88%91%E4%BB%AC%E5%8F%8A%E6%88%91%E4%BB%AC" target="_blank" rel="noopener noreferrer" class="underline underline-offset-2 hover:text-[#888]">CVSE野鸽社版权声明（2025年版）</a></p>
			</div>
			<div class="mt-5 text-right">
				<button
					type="button"
					class="rounded-lg bg-[#0a0a0a] px-4 py-2 text-xs font-medium text-white transition hover:bg-[#222]"
					onclick={() => showDisclaimer = false}
				>我知道了</button>
			</div>
		</div>
	</div>
{/if}
