package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

type Owner struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
}

type Stat struct {
	View     int64 `json:"view"`
	Danmaku  int64 `json:"danmaku"`
	Reply    int64 `json:"reply"`
	Like     int64 `json:"like"`
	Coin     int64 `json:"coin"`
	Favorite int64 `json:"favorite"`
	Share    int64 `json:"share"`
}

type Category struct {
	Tid      int64  `json:"tid"`
	Name     string `json:"name"`
	MainID   int64  `json:"mainId"`
	MainName string `json:"mainName"`
}

type ScoreInfo struct {
	Total            int64   `json:"total"`
	PlayScore        float64 `json:"playScore"`
	ConversionScore  float64 `json:"conversionScore"`
	InteractionScore float64 `json:"interactionScore"`
	RawA             float64 `json:"rawA"`
	CorrectionA      float64 `json:"correctionA"`
	RawB             float64 `json:"rawB"`
	CorrectionB      float64 `json:"correctionB"`
	RawC             float64 `json:"rawC"`
	CorrectionC      float64 `json:"correctionC"`
}

type ResolveData struct {
	BVID        string     `json:"bvid"`
	AID         int64      `json:"aid"`
	CID         int64      `json:"cid"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Pic         string     `json:"pic"`
	Owner       Owner      `json:"owner"`
	Stat        Stat       `json:"stat"`
	V1          Category   `json:"v1"`
	V2          *Category  `json:"v2,omitempty"`
	Pubdate     int64      `json:"pubdate"`
	Ctime       int64      `json:"ctime"`
	Score       *ScoreInfo `json:"score,omitempty"`
	Collected   bool       `json:"collected"`
	Ranks       []string   `json:"ranks,omitempty"`
}

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ---------------------------------------------------------------------------
// Bilibili API response structures (only fields we need)
// ---------------------------------------------------------------------------

type bilibiliViewResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    *bilibiliViewData `json:"data"`
}

type bilibiliViewData struct {
	BVID  string `json:"bvid"`
	AID   int64  `json:"aid"`
	CID   int64  `json:"cid"`
	Title string `json:"title"`
	Desc  string `json:"desc"`
	Pic   string `json:"pic"`
	Owner struct {
		Mid  int64  `json:"mid"`
		Name string `json:"name"`
	} `json:"owner"`
	Stat struct {
		View     int64 `json:"view"`
		Danmaku  int64 `json:"danmaku"`
		Reply    int64 `json:"reply"`
		Like     int64 `json:"like"`
		Coin     int64 `json:"coin"`
		Fav      int64 `json:"favorite"`
		Share    int64 `json:"share"`
	} `json:"stat"`
	Tid      int64  `json:"tid"`
	Tname    string `json:"tname"`
	TidV2    int64  `json:"tid_v2"`
	TnameV2  string `json:"tname_v2"`
	Pubdate  int64  `json:"pubdate"`
	Ctime    int64  `json:"ctime"`
}

// v1 category main-id mapping (hard-coded for offline operation)
var v1MainMap = map[int64]struct {
	MainID   int64
	MainName string
}{
	// 动画 (主分区)
	1:   {1, "动画"},
	24:  {1, "动画"},   // MAD·AMV
	25:  {1, "动画"},   // MMD·3D
	47:  {1, "动画"},   // 同人·手书
	257: {1, "动画"},   // 配音
	210: {1, "动画"},   // 手办·模玩
	86:  {1, "动画"},   // 特摄
	253: {1, "动画"},   // 动漫杂谈
	27:  {1, "动画"},   // 综合

	// 番剧 (主分区)
	13:  {13, "番剧"},
	51:  {13, "番剧"},   // 资讯
	152: {13, "番剧"},   // 官方延伸
	32:  {13, "番剧"},   // 完结动画
	33:  {13, "番剧"},   // 连载动画

	// 国创 (主分区)
	167: {167, "国创"},
	153: {167, "国创"},  // 国产动画
	168: {167, "国创"},  // 国产原创相关
	169: {167, "国创"},  // 布袋戏
	170: {167, "国创"},  // 资讯
	195: {167, "国创"},  // 动态漫·广播剧

	// 音乐 (主分区)
	3:   {3, "音乐"},
	28:  {3, "音乐"},   // 原创音乐
	29:  {3, "音乐"},   // 音乐现场
	31:  {3, "音乐"},   // 翻唱
	59:  {3, "音乐"},   // 演奏
	243: {3, "音乐"},   // 乐评盘点
	30:  {3, "音乐"},   // VOCALOID·UTAU
	193: {3, "音乐"},   // MV
	266: {3, "音乐"},   // 音乐粉丝饭拍
	265: {3, "音乐"},   // AI音乐
	267: {3, "音乐"},   // 电台
	244: {3, "音乐"},   // 音乐教学
	130: {3, "音乐"},   // 音乐综合

	// 舞蹈 (主分区)
	129: {129, "舞蹈"},
	20:  {129, "舞蹈"},  // 宅舞
	198: {129, "舞蹈"},  // 街舞
	199: {129, "舞蹈"},  // 明星舞蹈
	200: {129, "舞蹈"},  // 国风舞蹈
	255: {129, "舞蹈"},  // 颜值·网红舞
	154: {129, "舞蹈"},  // 舞蹈综合
	156: {129, "舞蹈"},  // 舞蹈教程

	// 游戏 (主分区)
	4:   {4, "游戏"},
	17:  {4, "游戏"},   // 单机游戏
	171: {4, "游戏"},   // 电子竞技
	172: {4, "游戏"},   // 手机游戏
	65:  {4, "游戏"},   // 网络游戏
	173: {4, "游戏"},   // 桌游棋牌
	121: {4, "游戏"},   // GMV
	136: {4, "游戏"},   // 音游
	19:  {4, "游戏"},   // Mugen

	// 知识 (主分区)
	36:  {36, "知识"},
	201: {36, "知识"},  // 科学科普
	124: {36, "知识"},  // 社科·法律·心理
	228: {36, "知识"},  // 人文历史
	207: {36, "知识"},  // 财经商业
	208: {36, "知识"},  // 校园学习
	209: {36, "知识"},  // 职业职场
	229: {36, "知识"},  // 设计·创意
	122: {36, "知识"},  // 野生技术协会

	// 科技 (主分区)
	188: {188, "科技"},
	95:  {188, "科技"},  // 数码
	230: {188, "科技"},  // 软件应用
	231: {188, "科技"},  // 计算机技术
	232: {188, "科技"},  // 科工机械
	233: {188, "科技"},  // 极客DIY

	// 运动 (主分区)
	234: {234, "运动"},
	235: {234, "运动"},  // 篮球
	249: {234, "运动"},  // 足球
	164: {234, "运动"},  // 健身
	236: {234, "运动"},  // 竞技体育
	237: {234, "运动"},  // 运动文化
	238: {234, "运动"},  // 运动综合

	// 汽车 (主分区)
	223: {223, "汽车"},
	258: {223, "汽车"},  // 汽车知识科普
	227: {223, "汽车"},  // 购车攻略
	247: {223, "汽车"},  // 新能源车
	245: {223, "汽车"},  // 赛车
	246: {223, "汽车"},  // 改装玩车
	240: {223, "汽车"},  // 摩托车
	248: {223, "汽车"},  // 房车
	176: {223, "汽车"},  // 汽车生活

	// 生活 (主分区)
	160: {160, "生活"},
	138: {160, "生活"},  // 搞笑
	254: {160, "生活"},  // 亲子
	250: {160, "生活"},  // 出行
	251: {160, "生活"},  // 三农
	239: {160, "生活"},  // 家居房产
	161: {160, "生活"},  // 手工
	162: {160, "生活"},  // 绘画
	21:  {160, "生活"},  // 日常

	// 美食 (主分区)
	211: {211, "美食"},
	76:  {211, "美食"},  // 美食制作
	212: {211, "美食"},  // 美食侦探
	213: {211, "美食"},  // 美食测评
	214: {211, "美食"},  // 田园美食
	215: {211, "美食"},  // 美食记录

	// 动物圈 (主分区)
	217: {217, "动物圈"},
	218: {217, "动物圈"},  // 喵星人
	219: {217, "动物圈"},  // 汪星人
	222: {217, "动物圈"},  // 小宠异宠
	221: {217, "动物圈"},  // 野生动物
	220: {217, "动物圈"},  // 动物二创
	75:  {217, "动物圈"},  // 动物综合

	// 鬼畜 (主分区)
	119: {119, "鬼畜"},
	22:  {119, "鬼畜"},   // 鬼畜调教
	26:  {119, "鬼畜"},   // 音MAD
	126: {119, "鬼畜"},   // 人力VOCALOID
	216: {119, "鬼畜"},   // 鬼畜剧场
	127: {119, "鬼畜"},   // 教程演示

	// 时尚 (主分区)
	155: {155, "时尚"},
	157: {155, "时尚"},  // 美妆护肤
	252: {155, "时尚"},  // 仿妆cos
	158: {155, "时尚"},  // 穿搭
	159: {155, "时尚"},  // 时尚潮流

	// 资讯 (主分区)
	202: {202, "资讯"},
	203: {202, "资讯"},  // 热点
	204: {202, "资讯"},  // 环球
	205: {202, "资讯"},  // 社会
	206: {202, "资讯"},  // 综合

	// 广告 (主分区)
	165: {165, "广告"},

	// 娱乐 (主分区)
	5:   {5, "娱乐"},
	241: {5, "娱乐"},   // 娱乐杂谈
	262: {5, "娱乐"},   // CP安利
	263: {5, "娱乐"},   // 颜值安利
	242: {5, "娱乐"},   // 娱乐粉丝创作
	264: {5, "娱乐"},   // 娱乐资讯
	137: {5, "娱乐"},   // 明星综合
	71:  {5, "娱乐"},   // 综艺

	// 影视 (主分区)
	181: {181, "影视"},
	182: {181, "影视"},  // 影视杂谈
	183: {181, "影视"},  // 影视剪辑
	260: {181, "影视"},  // 影视整活
	259: {181, "影视"},  // AI影像
	184: {181, "影视"},  // 预告·资讯
	85:  {181, "影视"},  // 小剧场
	256: {181, "影视"},  // 短片
	261: {181, "影视"},  // 影视综合

	// 纪录片 (主分区)
	177: {177, "纪录片"},
	37:  {177, "纪录片"},  // 人文·历史
	178: {177, "纪录片"},  // 科学·探索·自然
	179: {177, "纪录片"},  // 军事
	180: {177, "纪录片"},  // 社会·美食·旅行

	// 电影 (主分区)
	23:  {23, "电影"},
	147: {23, "电影"},  // 华语电影
	145: {23, "电影"},  // 欧美电影
	146: {23, "电影"},  // 日本电影
	83:  {23, "电影"},  // 其他国家

	// 电视剧 (主分区)
	11:  {11, "电视剧"},
	185: {11, "电视剧"}, // 国产剧
	187: {11, "电视剧"}, // 海外剧
}

// v2 category mapping (tid_v2 -> mainId / mainName). Partial.
var v2MainMap = map[int64]struct {
	MainID   int64
	MainName string
}{
	// 动画 (主分区) 1005
	1005: {1005, "动画"},
	2037: {1005, "动画"}, // 同人动画
	2038: {1005, "动画"}, // 模玩周边
	2039: {1005, "动画"}, // cosplay
	2040: {1005, "动画"}, // 二次元线下
	2041: {1005, "动画"}, // 动漫剪辑
	2042: {1005, "动画"}, // 动漫评论
	2043: {1005, "动画"}, // 动漫速读
	2044: {1005, "动画"}, // 动漫配音
	2045: {1005, "动画"}, // 动漫资讯
	2046: {1005, "动画"}, // 网文解读
	2047: {1005, "动画"}, // 虚拟up主
	2048: {1005, "动画"}, // 特摄
	2049: {1005, "动画"}, // 布袋戏
	2050: {1005, "动画"}, // 漫画·动态漫
	2051: {1005, "动画"}, // 广播剧
	2052: {1005, "动画"}, // 动漫reaction
	2053: {1005, "动画"}, // 动漫教学
	2054: {1005, "动画"}, // 二次元其他

	// 游戏 (主分区) 1008
	1008: {1008, "游戏"},
	2064: {1008, "游戏"}, // 单人RPG游戏
	2065: {1008, "游戏"}, // MMORPG游戏
	2066: {1008, "游戏"}, // 单机主机类游戏
	2067: {1008, "游戏"}, // SLG游戏
	2068: {1008, "游戏"}, // 回合制策略游戏
	2069: {1008, "游戏"}, // 即时策略游戏
	2070: {1008, "游戏"}, // MOBA游戏
	2071: {1008, "游戏"}, // 射击游戏
	2072: {1008, "游戏"}, // 体育竞速游戏
	2073: {1008, "游戏"}, // 动作竞技游戏
	2074: {1008, "游戏"}, // 音游舞游
	2075: {1008, "游戏"}, // 模拟经营游戏
	2076: {1008, "游戏"}, // 女性向游戏
	2077: {1008, "游戏"}, // 休闲/小游戏
	2078: {1008, "游戏"}, // 沙盒类
	2079: {1008, "游戏"}, // 其他游戏

	// 鬼畜 (主分区) 1007
	1007: {1007, "鬼畜"},
	2059: {1007, "鬼畜"}, // 鬼畜调教
	2060: {1007, "鬼畜"}, // 鬼畜剧场
	2061: {1007, "鬼畜"}, // 人力VOCALOID
	2062: {1007, "鬼畜"}, // 音MAD
	2063: {1007, "鬼畜"}, // 鬼畜综合

	// 音乐 (主分区) 1003
	1003: {1003, "音乐"},
	2016: {1003, "音乐"}, // 原创音乐
	2017: {1003, "音乐"}, // MV
	2018: {1003, "音乐"}, // 音乐现场
	2019: {1003, "音乐"}, // 乐迷饭拍
	2020: {1003, "音乐"}, // 翻唱
	2021: {1003, "音乐"}, // 演奏
	2022: {1003, "音乐"}, // VOCALOID
	2023: {1003, "音乐"}, // AI音乐
	2024: {1003, "音乐"}, // 电台·歌单
	2025: {1003, "音乐"}, // 音乐教学
	2026: {1003, "音乐"}, // 乐评盘点
	2027: {1003, "音乐"}, // 音乐综合

	// 舞蹈 (主分区) 1004
	1004: {1004, "舞蹈"},
	2028: {1004, "舞蹈"}, // 宅舞
	2029: {1004, "舞蹈"}, // 街舞
	2030: {1004, "舞蹈"}, // 颜值·网红舞
	2031: {1004, "舞蹈"}, // 明星舞蹈
	2032: {1004, "舞蹈"}, // 国风舞蹈
	2033: {1004, "舞蹈"}, // 舞蹈教学
	2034: {1004, "舞蹈"}, // 芭蕾舞
	2035: {1004, "舞蹈"}, // wota艺
	2036: {1004, "舞蹈"}, // 舞蹈综合

	// 影视 (主分区) 1001
	1001: {1001, "影视"},
	2001: {1001, "影视"}, // 影视解读
	2002: {1001, "影视"}, // 影视剪辑
	2003: {1001, "影视"}, // 影视资讯
	2004: {1001, "影视"}, // 影视正片搬运
	2005: {1001, "影视"}, // 短剧短片
	2006: {1001, "影视"}, // AI影视
	2007: {1001, "影视"}, // 影视reaction
	2008: {1001, "影视"}, // 影视综合

	// 娱乐 (主分区) 1002
	1002: {1002, "娱乐"},
	2009: {1002, "娱乐"}, // 娱乐评论
	2010: {1002, "娱乐"}, // 明星剪辑
	2011: {1002, "娱乐"}, // 娱乐饭拍&现场
	2012: {1002, "娱乐"}, // 娱乐资讯
	2013: {1002, "娱乐"}, // 娱乐reaction
	2014: {1002, "娱乐"}, // 娱乐综艺正片
	2015: {1002, "娱乐"}, // 娱乐综合

	// 知识 (主分区) 1010
	1010: {1010, "知识"},
	2084: {1010, "知识"}, // 应试教育
	2085: {1010, "知识"}, // 非应试语言学习
	2086: {1010, "知识"}, // 大学专业知识
	2087: {1010, "知识"}, // 商业财经
	2088: {1010, "知识"}, // 社会观察
	2089: {1010, "知识"}, // 时政解读
	2090: {1010, "知识"}, // 人文历史
	2091: {1010, "知识"}, // 设计艺术
	2092: {1010, "知识"}, // 心理杂谈
	2093: {1010, "知识"}, // 职场发展
	2094: {1010, "知识"}, // 科学科普
	2095: {1010, "知识"}, // 其他知识杂谈

	// 科技数码 (主分区) 1012
	1012: {1012, "科技数码"},
	2099: {1012, "科技数码"}, // 电脑
	2100: {1012, "科技数码"}, // 手机
	2101: {1012, "科技数码"}, // 平板电脑
	2102: {1012, "科技数码"}, // 摄影摄像
	2103: {1012, "科技数码"}, // 工程机械
	2104: {1012, "科技数码"}, // 自制发明/设备
	2105: {1012, "科技数码"}, // 科技数码综合

	// 资讯 (主分区) 1009
	1009: {1009, "资讯"},
	2080: {1009, "资讯"}, // 时政资讯
	2081: {1009, "资讯"}, // 海外资讯
	2082: {1009, "资讯"}, // 社会资讯
	2083: {1009, "资讯"}, // 综合资讯

	// 美食 (主分区) 1020
	1020: {1020, "美食"},
	2149: {1020, "美食"}, // 美食制作
	2150: {1020, "美食"}, // 美食探店
	2151: {1020, "美食"}, // 美食测评
	2152: {1020, "美食"}, // 美食记录
	2153: {1020, "美食"}, // 美食综合

	// 小剧场 (主分区) 1021
	1021: {1021, "小剧场"},
	2154: {1021, "小剧场"}, // 剧情演绎
	2155: {1021, "小剧场"}, // 语言类小剧场
	2156: {1021, "小剧场"}, // UP主小综艺
	2157: {1021, "小剧场"}, // 街头采访

	// 汽车 (主分区) 1013
	1013: {1013, "汽车"},
	2106: {1013, "汽车"}, // 汽车测评
	2107: {1013, "汽车"}, // 汽车文化
	2108: {1013, "汽车"}, // 汽车生活
	2109: {1013, "汽车"}, // 汽车技术
	2110: {1013, "汽车"}, // 汽车综合

	// 时尚美妆 (主分区) 1014
	1014: {1014, "时尚美妆"},
	2111: {1014, "时尚美妆"}, // 美妆
	2112: {1014, "时尚美妆"}, // 护肤
	2113: {1014, "时尚美妆"}, // 仿装cos
	2114: {1014, "时尚美妆"}, // 鞋服穿搭
	2115: {1014, "时尚美妆"}, // 箱包配饰
	2116: {1014, "时尚美妆"}, // 珠宝首饰
	2117: {1014, "时尚美妆"}, // 三坑
	2118: {1014, "时尚美妆"}, // 时尚解读
	2119: {1014, "时尚美妆"}, // 时尚综合

	// 体育运动 (主分区) 1018
	1018: {1018, "体育运动"},
	2133: {1018, "体育运动"}, // 潮流运动
	2134: {1018, "体育运动"}, // 足球
	2135: {1018, "体育运动"}, // 篮球
	2136: {1018, "体育运动"}, // 跑步
	2137: {1018, "体育运动"}, // 武术
	2138: {1018, "体育运动"}, // 格斗
	2139: {1018, "体育运动"}, // 羽毛球
	2140: {1018, "体育运动"}, // 体育资讯
	2141: {1018, "体育运动"}, // 体育赛事
	2142: {1018, "体育运动"}, // 体育综合

	// 动物 (主分区) 1024
	1024: {1024, "动物"},
	2167: {1024, "动物"}, // 猫
	2168: {1024, "动物"}, // 狗
	2169: {1024, "动物"}, // 小宠异宠
	2170: {1024, "动物"}, // 野生动物·动物解说科普
	2171: {1024, "动物"}, // 动物综合·二创

	// vlog (主分区) 1029
	1029: {1029, "vlog"},
	2194: {1029, "vlog"}, // 中外生活vlog
	2195: {1029, "vlog"}, // 学生vlog
	2196: {1029, "vlog"}, // 职业vlog
	2197: {1029, "vlog"}, // 其他vlog

	// 绘画 (主分区) 1006
	1006: {1006, "绘画"},
	2055: {1006, "绘画"}, // 二次元绘画
	2056: {1006, "绘画"}, // 非二次元绘画
	2057: {1006, "绘画"}, // 绘画学习
	2058: {1006, "绘画"}, // 绘画综合

	// 人工智能 (主分区) 1011
	1011: {1011, "人工智能"},
	2096: {1011, "人工智能"}, // AI学习
	2097: {1011, "人工智能"}, // AI资讯
	2098: {1011, "人工智能"}, // AI杂谈

	// 家装房产 (主分区) 1015
	1015: {1015, "家装房产"},
	2120: {1015, "家装房产"}, // 买房租房
	2121: {1015, "家装房产"}, // 家庭装修
	2122: {1015, "家装房产"}, // 家居展示
	2123: {1015, "家装房产"}, // 家用电器

	// 户外潮流 (主分区) 1016
	1016: {1016, "户外潮流"},
	2124: {1016, "户外潮流"}, // 露营
	2125: {1016, "户外潮流"}, // 徒步
	2126: {1016, "户外潮流"}, // 户外探秘
	2127: {1016, "户外潮流"}, // 户外综合

	// 健身 (主分区) 1017
	1017: {1017, "健身"},
	2128: {1017, "健身"}, // 健身科普
	2129: {1017, "健身"}, // 健身跟练教学
	2130: {1017, "健身"}, // 健身记录
	2131: {1017, "健身"}, // 健身身材展示
	2132: {1017, "健身"}, // 健身综合

	// 手工 (主分区) 1019
	1019: {1019, "手工"},
	2143: {1019, "手工"}, // 文具手帐
	2144: {1019, "手工"}, // 轻手作
	2145: {1019, "手工"}, // 传统手工艺
	2146: {1019, "手工"}, // 解压手工
	2147: {1019, "手工"}, // DIY玩具
	2148: {1019, "手工"}, // 其他手工

	// 旅游出行 (主分区) 1022
	1022: {1022, "旅游出行"},
	2158: {1022, "旅游出行"}, // 旅游记录
	2159: {1022, "旅游出行"}, // 旅游攻略
	2160: {1022, "旅游出行"}, // 城市出行
	2161: {1022, "旅游出行"}, // 公共交通

	// 三农 (主分区) 1023
	1023: {1023, "三农"},
	2162: {1023, "三农"}, // 农村种植
	2163: {1023, "三农"}, // 赶海捕鱼
	2164: {1023, "三农"}, // 打野采摘
	2165: {1023, "三农"}, // 农业技术
	2166: {1023, "三农"}, // 农村生活

	// 亲子 (主分区) 1025
	1025: {1025, "亲子"},
	2172: {1025, "亲子"}, // 孕产护理
	2173: {1025, "亲子"}, // 婴幼护理
	2174: {1025, "亲子"}, // 儿童才艺
	2175: {1025, "亲子"}, // 萌娃
	2176: {1025, "亲子"}, // 亲子互动
	2177: {1025, "亲子"}, // 亲子教育
	2178: {1025, "亲子"}, // 亲子综合

	// 健康 (主分区) 1026
	1026: {1026, "健康"},
	2179: {1026, "健康"}, // 健康科普
	2180: {1026, "健康"}, // 养生
	2181: {1026, "健康"}, // 两性知识
	2182: {1026, "健康"}, // 心理健康
	2183: {1026, "健康"}, // 助眠视频·ASMR
	2184: {1026, "健康"}, // 医疗保健综合

	// 情感 (主分区) 1027
	1027: {1027, "情感"},
	2185: {1027, "情感"}, // 家庭关系
	2186: {1027, "情感"}, // 恋爱关系
	2187: {1027, "情感"}, // 人际关系
	2188: {1027, "情感"}, // 自我成长

	// 生活兴趣 (主分区) 1030
	1030: {1030, "生活兴趣"},
	2198: {1030, "生活兴趣"}, // 休闲玩乐
	2199: {1030, "生活兴趣"}, // 线下演出
	2200: {1030, "生活兴趣"}, // 文玩文创
	2201: {1030, "生活兴趣"}, // 潮玩玩具
	2202: {1030, "生活兴趣"}, // 兴趣综合

	// 生活经验 (主分区) 1031
	1031: {1031, "生活经验"},
	2203: {1031, "生活经验"}, // 生活技能
	2204: {1031, "生活经验"}, // 办事流程
	2205: {1031, "生活经验"}, // 婚嫁

	// 神秘学 (主分区) 1028
	1028: {1028, "神秘学"},
	2189: {1028, "神秘学"}, // 塔罗占卜
	2190: {1028, "神秘学"}, // 星座占星
	2191: {1028, "神秘学"}, // 传统玄学
	2192: {1028, "神秘学"}, // 疗愈成长
	2193: {1028, "神秘学"}, // 其他神秘学
}

// ---------------------------------------------------------------------------
// Input parsing
// ---------------------------------------------------------------------------

var (
	reBV       = regexp.MustCompile(`BV\w{10}`)
	reAv       = regexp.MustCompile(`av(\d+)`)
	reBilibili = regexp.MustCompile(`bilibili\.com/video/(BV\w{10})`)
	reB23      = regexp.MustCompile(`https?://b23\.(tv|wtf)/\w+`)
)

// parseInput extracts bvid or aid from various input formats.
// Returns (bvid, aid). One will be empty.
func parseInput(raw string) (bvid string, aid int64) {
	raw = strings.TrimSpace(raw)

	// 1) Try full bilibili URL first
	if m := reBilibili.FindStringSubmatch(raw); len(m) > 1 {
		return m[1], 0
	}

	// 2) Try bare BV
	if m := reBV.FindString(raw); m != "" {
		return m, 0
	}

	// 3) Try av prefix
	if m := reAv.FindStringSubmatch(raw); len(m) > 1 {
		n, _ := strconv.ParseInt(m[1], 10, 64)
		return "", n
	}

	// 4) b23.tv short link – extract URL from text and follow redirect
	if strings.Contains(raw, "b23.tv") || strings.Contains(raw, "b23.wtf") {
		b23url := ""
		if m := reB23.FindString(raw); m != "" {
			b23url = m
		} else {
			// Fallback: prepend https if missing
			if !strings.HasPrefix(raw, "http") {
				b23url = "https://" + raw
			} else {
				b23url = raw
			}
		}
		if bvid, err := resolveB23(b23url); err == nil && bvid != "" {
			return bvid, 0
		}
	}

	// 5) Try bare number as aid
	if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return "", n
	}

	return "", 0
}

// resolveB23 follows a b23 short link to extract the destination video URL.
func resolveB23(shortURL string) (string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	// Ensure URL has scheme
	if !strings.HasPrefix(shortURL, "http") {
		shortURL = "https://" + shortURL
	}

	resp, err := client.Get(shortURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	finalURL := resp.Request.URL.String()
	if m := reBilibili.FindStringSubmatch(finalURL); len(m) > 1 {
		return m[1], nil
	}
	return "", fmt.Errorf("not a bilibili video link")
}

// ---------------------------------------------------------------------------
// Bilibili API caller
// ---------------------------------------------------------------------------

var httpClient = &http.Client{Timeout: 10 * time.Second}

func fetchBilibiliView(bvid string, aid int64) (*bilibiliViewData, error) {
	u := "https://api.bilibili.com/x/web-interface/view?"
	if bvid != "" {
		u += "bvid=" + url.QueryEscape(bvid)
	} else {
		u += "aid=" + strconv.FormatInt(aid, 10)
	}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.bilibili.com")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var bResp bilibiliViewResponse
	if err := json.Unmarshal(body, &bResp); err != nil {
		return nil, err
	}

	if bResp.Code != 0 {
		return nil, fmt.Errorf("Bilibili API error: code=%d msg=%s", bResp.Code, bResp.Message)
	}
	if bResp.Data == nil {
		return nil, fmt.Errorf("Bilibili API returned empty data")
	}

	return bResp.Data, nil
}

// ---------------------------------------------------------------------------
// CVSE internal database API caller
// ---------------------------------------------------------------------------

type cvseVideoResponse struct {
	Success bool            `json:"success"`
	Data    *cvseVideoData  `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type cvseVideoData struct {
	BVID         string   `json:"bvid"`
	AVID         string   `json:"avid"`
	Title        string   `json:"title"`
	Desc         string   `json:"desc"`
	Cover        string   `json:"cover"`
	Uploader     string   `json:"uploader"`
	UpFace       string   `json:"up_face"`
	PubTimestamp int64    `json:"pub_timestamp"`
	Pubdate      string   `json:"pubdate"`
	Duration     int64    `json:"duration"`
	Tags         []string `json:"tags"`
	Ranks        []string `json:"ranks"`
	IsExamined   bool     `json:"is_examined"`
	IsRepublish  bool     `json:"is_republish"`
	StaffInfo    string   `json:"staff_info"`
}

var cvseHTTPClient = &http.Client{Timeout: 5 * time.Second}

// CVSE API base URL, injected at build time via -ldflags -X
var CVSEAPIBase = "http://103.40.13.253:20402"

func fetchCVSE(bvid string) (*cvseVideoData, error) {
	u := CVSEAPIBase + "/api/video/" + url.PathEscape(bvid)

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "CVSE-xTower/1.0")

	resp, err := cvseHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cvseResp cvseVideoResponse
	if err := json.Unmarshal(body, &cvseResp); err != nil {
		return nil, err
	}

	if !cvseResp.Success || cvseResp.Data == nil {
		return nil, fmt.Errorf("not collected")
	}

	return cvseResp.Data, nil
}

// ---------------------------------------------------------------------------
// Category helpers
// ---------------------------------------------------------------------------

// Local name fallback for v1 categories when Bilibili API returns empty tname.
var v1NameMap = map[int64]string{
	// 动画
	1:   "动画",
	24:  "MAD·AMV",
	25:  "MMD·3D",
	47:  "同人·手书",
	257: "配音",
	210: "手办·模玩",
	86:  "特摄",
	253: "动漫杂谈",
	27:  "综合",

	// 番剧
	13:  "番剧",
	51:  "资讯",
	152: "官方延伸",
	32:  "完结动画",
	33:  "连载动画",

	// 国创
	167: "国创",
	153: "国产动画",
	168: "国产原创相关",
	169: "布袋戏",
	170: "资讯",
	195: "动态漫·广播剧",

	// 音乐
	3:   "音乐",
	28:  "原创音乐",
	29:  "音乐现场",
	31:  "翻唱",
	59:  "演奏",
	243: "乐评盘点",
	30:  "VOCALOID·UTAU",
	193: "MV",
	266: "音乐粉丝饭拍",
	265: "AI音乐",
	267: "电台",
	244: "音乐教学",
	130: "音乐综合",

	// 舞蹈
	129: "舞蹈",
	20:  "宅舞",
	198: "街舞",
	199: "明星舞蹈",
	200: "国风舞蹈",
	255: "颜值·网红舞",
	154: "舞蹈综合",
	156: "舞蹈教程",

	// 游戏
	4:   "游戏",
	17:  "单机游戏",
	171: "电子竞技",
	172: "手机游戏",
	65:  "网络游戏",
	173: "桌游棋牌",
	121: "GMV",
	136: "音游",
	19:  "Mugen",

	// 知识
	36:  "知识",
	201: "科学科普",
	124: "社科·法律·心理",
	228: "人文历史",
	207: "财经商业",
	208: "校园学习",
	209: "职业职场",
	229: "设计·创意",
	122: "野生技术协会",

	// 科技
	188: "科技",
	95:  "数码",
	230: "软件应用",
	231: "计算机技术",
	232: "科工机械",
	233: "极客DIY",

	// 运动
	234: "运动",
	235: "篮球",
	249: "足球",
	164: "健身",
	236: "竞技体育",
	237: "运动文化",
	238: "运动综合",

	// 汽车
	223: "汽车",
	258: "汽车知识科普",
	227: "购车攻略",
	247: "新能源车",
	245: "赛车",
	246: "改装玩车",
	240: "摩托车",
	248: "房车",
	176: "汽车生活",

	// 生活
	160: "生活",
	138: "搞笑",
	254: "亲子",
	250: "出行",
	251: "三农",
	239: "家居房产",
	161: "手工",
	162: "绘画",
	21:  "日常",

	// 美食
	211: "美食",
	76:  "美食制作",
	212: "美食侦探",
	213: "美食测评",
	214: "田园美食",
	215: "美食记录",

	// 动物圈
	217: "动物圈",
	218: "喵星人",
	219: "汪星人",
	222: "小宠异宠",
	221: "野生动物",
	220: "动物二创",
	75:  "动物综合",

	// 鬼畜
	119: "鬼畜",
	22:  "鬼畜调教",
	26:  "音MAD",
	126: "人力VOCALOID",
	216: "鬼畜剧场",
	127: "教程演示",

	// 时尚
	155: "时尚",
	157: "美妆护肤",
	252: "仿妆cos",
	158: "穿搭",
	159: "时尚潮流",

	// 资讯
	202: "资讯",
	203: "热点",
	204: "环球",
	205: "社会",
	206: "综合",

	// 广告
	165: "广告",

	// 娱乐
	5:   "娱乐",
	241: "娱乐杂谈",
	262: "CP安利",
	263: "颜值安利",
	242: "娱乐粉丝创作",
	264: "娱乐资讯",
	137: "明星综合",
	71:  "综艺",

	// 影视
	181: "影视",
	182: "影视杂谈",
	183: "影视剪辑",
	260: "影视整活",
	259: "AI影像",
	184: "预告·资讯",
	85:  "小剧场",
	256: "短片",
	261: "影视综合",

	// 纪录片
	177: "纪录片",
	37:  "人文·历史",
	178: "科学·探索·自然",
	179: "军事",
	180: "社会·美食·旅行",

	// 电影
	23:  "电影",
	147: "华语电影",
	145: "欧美电影",
	146: "日本电影",
	83:  "其他国家",

	// 电视剧
	11:  "电视剧",
	185: "国产剧",
	187: "海外剧",
}

func resolveV1(tid int64, tname string) Category {
	c := Category{Tid: tid, Name: tname}
	if c.Name == "" {
		if n, ok := v1NameMap[tid]; ok {
			c.Name = n
		}
	}
	if m, ok := v1MainMap[tid]; ok {
		c.MainID = m.MainID
		c.MainName = m.MainName
	} else {
		c.MainID = tid
		if c.Name != "" {
			c.MainName = c.Name
		} else {
			c.MainName = fmt.Sprintf("tid_%d", tid)
		}
	}
	return c
}

func resolveV2(tidV2 int64, tnameV2 string) *Category {
	if tidV2 == 0 {
		return nil
	}
	c := &Category{Tid: tidV2, Name: tnameV2}
	if c.Name == "" {
		if m, ok := v2MainMap[tidV2]; ok {
			c.Name = m.MainName
		}
	}
	if m, ok := v2MainMap[tidV2]; ok {
		c.MainID = m.MainID
		c.MainName = m.MainName
	} else {
		c.MainID = tidV2
		if c.Name != "" {
			c.MainName = c.Name
		} else {
			c.MainName = fmt.Sprintf("tid_v2_%d", tidV2)
		}
	}
	return c
}

// ---------------------------------------------------------------------------
// Score calculation — CVSE Pt. formula
// ---------------------------------------------------------------------------

// computeScore calculates the CVSE score (Pt.) based on the provided formula.
// pubdate is the video's publish timestamp (unix seconds).
func computeScore(stat Stat, pubdate int64) ScoreInfo {
	view := float64(stat.View)
	danmaku := float64(stat.Danmaku)
	reply := float64(stat.Reply)
	like := float64(stat.Like)
	coin := float64(stat.Coin)
	fav := float64(stat.Favorite)
	share := float64(stat.Share)

	// ---- 得点A (播放行为原始分) ----
	rawA := view/2 + like*4 + share*50
	if rawA < 0 {
		rawA = 0
	}

	// ---- 得点B (转化表现原始分) ----
	var rawB float64
	denomB := fav + coin*3
	if denomB > 0 {
		rawB = fav * coin / denomB * 108
	}
	if rawB < 0 {
		rawB = 0
	}

	// ---- 修正A ----
	corrA := ((rawA + rawB*10) / (rawA*10 + rawB)) * 2
	// 稿件发布 ≤ 14天(336小时)，修正A下限为1
	ageHours := (time.Now().Unix() - pubdate) / 3600
	if ageHours <= 336 && corrA < 1.0 {
		corrA = 1.0
	}

	// ---- 修正B ----
	var corrB float64
	denomCorrB := rawA*9 + rawB
	if denomCorrB > 0 {
		corrB = rawB / denomCorrB * 10 / 3
	}

	// ---- 得点C (互动原始分) ----
	rawC := (reply + danmaku*2) * 70
	if rawC < 0 {
		rawC = 0
	}

	// ---- 修正C ----
	var corrC float64
	denomCorrC := rawA*9 + rawB + rawC
	if denomCorrC > 0 {
		corrC = (rawA + rawB*10) / denomCorrC / 3
	}

	// ---- 各维度分 ----
	playScore := rawA * corrA
	conversionScore := rawB * corrB
	interactionScore := rawC * corrC

	// ---- 总分 (四舍五入至整数) ----
	total := playScore + conversionScore + interactionScore

	return ScoreInfo{
		Total:            int64(total + 0.5),
		PlayScore:        playScore,
		ConversionScore:  conversionScore,
		InteractionScore: interactionScore,
		RawA:             rawA,
		CorrectionA:      corrA,
		RawB:             rawB,
		CorrectionB:      corrB,
		RawC:             rawC,
		CorrectionC:      corrC,
	}
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

func handleResolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Code: -1, Message: "method not allowed"})
		return
	}

	input := r.URL.Query().Get("input")
	if input == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Code: -1, Message: "missing query param: input"})
		return
	}

	bvid, aid := parseInput(input)
	if bvid == "" && aid == 0 {
		writeJSON(w, http.StatusBadRequest, APIResponse{Code: -1, Message: "无法解析视频标识，请输入 BV号 / av号 / 视频链接"})
		return
	}

	// ---- 查询 CVSE 内库（检查是否已收录）----
	var collected bool
	var ranks []string
	if bvid != "" {
		if cvseData, err := fetchCVSE(bvid); err == nil {
			collected = true
			ranks = cvseData.Ranks
		}
	}

	data, err := fetchBilibiliView(bvid, aid)
	if err != nil {
		log.Printf("fetch error: %v", err)
		writeJSON(w, http.StatusBadGateway, APIResponse{Code: -2, Message: "B站API请求失败: " + err.Error()})
		return
	}

	v1 := resolveV1(data.Tid, data.Tname)
	var v2 *Category
	if data.TidV2 != 0 {
		v2 = resolveV2(data.TidV2, data.TnameV2)
	}

	respData := ResolveData{
		BVID:        data.BVID,
		AID:         data.AID,
		CID:         data.CID,
		Title:       data.Title,
		Description: data.Desc,
		Pic:         strings.Replace(data.Pic, "http://", "https://", 1),
		Owner: Owner{
			Mid:  data.Owner.Mid,
			Name: data.Owner.Name,
		},
		Stat: Stat{
			View:     data.Stat.View,
			Danmaku:  data.Stat.Danmaku,
			Reply:    data.Stat.Reply,
			Like:     data.Stat.Like,
			Coin:     data.Stat.Coin,
			Favorite: data.Stat.Fav,
			Share:    data.Stat.Share,
		},
		V1:      v1,
		V2:      v2,
		Pubdate: data.Pubdate,
		Ctime:   data.Ctime,
		Collected: collected,
		Ranks:     ranks,
	}

	// ---- 计算周刊分数 ----
	score := computeScore(respData.Stat, respData.Pubdate)
	respData.Score = &score

	writeJSON(w, http.StatusOK, APIResponse{Code: 0, Message: "ok", Data: respData})
}

// ---------------------------------------------------------------------------
// CORS middleware
// ---------------------------------------------------------------------------

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/video/resolve", corsMiddleware(handleResolve))

	// Health check
	mux.HandleFunc("/api/health", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, APIResponse{Code: 0, Message: "ok"})
	}))

	addr := ":8080"
	log.Printf("CVSE xTower Go API server starting on %s", addr)
	log.Printf("Endpoint: GET http://localhost%s/api/video/resolve?input=BV1GJ411x7h7", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
