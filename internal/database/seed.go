package database

import (
	"log"

	"purrtopia/internal/database/models"
)

// SeedGroupThreadNames 植入彩虹與釣魚的隨機團名（若資料庫內尚無資料才執行）
func SeedGroupThreadNames() {
	rainbowNames := []string{
		"超夢幻彩虹", "霓虹狂歡派對", "七色光芒", "彩色棉花糖",
		"虹橋特攻", "粉紅泡泡", "天空彩繪師", "彩虹冰淇淋",
		"阿彩彩探險", "七彩神龍", "彩虹棒棒糖", "光譜終結者",
		"彩虹喵星人", "糖果彩虹", "彩虹寶石護衛", "幸運彩虹",
		"彩虹鯉魚旗", "彩虹出沒請注意", "跟著彩虹衝", "彩虹廢材",
		"彩虹今天好嗎安安", "彩虹轟炸機", "虹光閃閃有夠讚", "彩虹啦啦隊",
		"彩霓飛舞大爆發", "彩虹奇蹟冒險島", "彩虹掰掰再見不見", "七色大鳥",
		"彩虹搞笑劇場", "彩虹高手", "彩虹超能力者", "彩虹哇哇哇",
		"彩虹塗鴉牆", "阿嬤的彩虹被", "我愛彩虹", "彩虹閃閃亮",
		"我是彩虹我驕傲", "彩虹闖關記", "彩虹抱抱互助社", "彩虹吃我的灰",
		"彩虹便便奇遇記", "彩虹追追追", "虹之霸主", "彩虹人生笑著過",
		"彩虹偵探事務所", "彩虹大爆笑", "七彩霓虹閃光", "彩繪人生",
		"彩虹今天幾朵", "彩虹出來玩嘍",
	}

	fishingNames := []string{
		"魚兒上鉤記", "等魚等到天荒地老", "釣魚佬俱樂部", "今天是大漁日",
		"魚桶打翻俱樂部", "假裝釣到魚的人", "魚兒魚兒水中游", "我才不是在釣魚",
		"大釣特釣無敵", "魚竿揮揮好快樂", "釣到空氣也算嗎", "鯨魚釣不到",
		"摸魚摸到一身腥", "魚說我不想被釣", "釣魚佬的白日夢", "海鷗來搶餌",
		"今天吃魚全配", "釣到什麼吃什麼", "魚鉤比魚多", "陽光沙灘魚竿",
		"水上漂流探險", "釣竿折斷俱樂部", "魚網破洞修補", "老漁夫說不行",
		"等魚等到手麻了", "釣魚要有耐心", "魚說你釣不到我", "釣魚界的幸運星",
		"今天一定大爆魚", "魚塘霸主", "釣魚心靜如水", "假日釣魚快樂",
		"魚竿保養研究社", "水中的寶藏獵人", "大魚小魚都要釣", "釣魚比賽廢材",
		"魚兒快跑有人追", "浮標咚咚響", "釣魚要起大早", "空竿也要笑著走",
		"釣到人生大道理", "魚說好冷別來了", "釣魚界傳說人物", "水面沒動靜也等",
		"快來啊有魚在", "釣魚吃到吃不下", "漁翁得利", "魚餌消失的謎案",
		"最後一桿釣到了", "今天風很大繼續",
	}

	seedByType("rainbow", rainbowNames)
	seedByType("fishing", fishingNames)
}

func seedByType(actType string, names []string) {
	count, err := GroupThreadNameRepo.CountByType(actType)
	if err != nil {
		log.Printf("[seed] 無法檢查 %s 團名數量: %v", actType, err)
		return
	}
	if count > 0 {
		return // 已有資料，跳過
	}

	entries := make([]models.GroupThreadName, 0, len(names))
	for _, name := range names {
		entries = append(entries, models.GroupThreadName{
			ActivityType: actType,
			Name:         name,
		})
	}

	if err := GroupThreadNameRepo.CreateBatch(entries); err != nil {
		log.Printf("[seed] 插入 %s 團名失敗: %v", actType, err)
		return
	}
	log.Printf("[seed] 已插入 %d 個 %s 團名", len(entries), actType)
}
