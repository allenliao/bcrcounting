package models

type BetType uint8

const (
	Bcr_BETTYPE_BANKER = iota
	Bcr_BETTYPE_PLAYER
	Bcr_BETTYPE_TIE
	Bcr_BETTYPE_BIG
	Bcr_BETTYPE_SMALL
)

var (
	Bcr_BetTypeCount uint8 = 5
	//賭場優勢 (莊贏抽水0.05為例) ，若算到後來變正的 變賭場失去優勢
	Bcr_PlayerHouseEdgeDefault float32 = -0.0124
	Bcr_BankerHouseEdgeDefault float32 = -0.01053
	Bcr_TieHouseEdgeDefault    float32 = -0.1432
	//牌點 0 ~ 9
	Bcr_PlayerHouseEdgeEffectList [10]float32 = [10]float32{-0.000018, -0.000045, -0.000054, -0.000067, -0.000120, 0.000084, 0.000113, 0.000082, 0.000053, 0.000025}
	Bcr_BankerHouseEdgeEffectList [10]float32 = [10]float32{0.000019, 0.000044, 0.000052, 0.000065, 0.000116, -0.000083, -0.000113, -0.000083, -0.00005, -0.000023}
	Bcr_TieHouseEdgeEffectList    [10]float32 = [10]float32{0.000513, 0.000129, -0.000239, -0.000214, -0.000292, -0.000264, -0.001160, -0.001091, 0.000654, 0.000426}
)

var BetTypeCount uint8 = 5

type CountingResult struct {
	BUCode              string           //BU 代碼
	GameIDDisplay       string           //局號
	TableNo             uint8            //桌號
	BetSuggestionData   [3]BetSuggestion //建議值
	SuggestionBet       string
	SuggestionBetAmount int16
	Result              string
	GuessResult         bool
}

type BetSuggestion struct {
	BetType    uint8
	HouseEdge  float32 //要大於0才有搞頭(賭場優勢 (莊贏抽水0.05為例) ，若算到後來變正的 變賭場失去優勢)
	SuggestBet bool
}

func transBetTypeToStr(betType uint8) string {
	switch betType {
	case Bcr_BETTYPE_BANKER:
		return "莊"
	case Bcr_BETTYPE_PLAYER:
		return "閒"
	case Bcr_BETTYPE_TIE:
		return "和"
	}
	return ""
}
