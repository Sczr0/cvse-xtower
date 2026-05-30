/** Mirrors the Go backend API response types. */

export interface Owner {
	mid: number;
	name: string;
}

export interface Stat {
	view: number;
	danmaku: number;
	reply: number;
	like: number;
	coin: number;
	favorite: number;
	share: number;
}

export interface Category {
	tid: number;
	name: string;
	mainId: number;
	mainName: string;
}

export interface ScoreInfo {
	total: number;
	playScore: number;
	conversionScore: number;
	interactionScore: number;
	rawA: number;
	correctionA: number;
	rawB: number;
	correctionB: number;
	rawC: number;
	correctionC: number;
}

export interface ResolveData {
	bvid: string;
	aid: number;
	cid: number;
	title: string;
	description: string;
	pic: string;
	owner: Owner;
	stat: Stat;
	v1: Category;
	v2?: Category | null;
	pubdate: number;
	ctime: number;
	score: ScoreInfo | null;
}

export interface APIResponse {
	code: number;
	message: string;
	data?: ResolveData;
}


