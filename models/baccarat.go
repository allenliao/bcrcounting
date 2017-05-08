package models

type BetType uint8

const (
	Bcr_BETTYPE_BANKER = iota
	Bcr_BETTYPE_PLAYER
	Bcr_BETTYPE_TIE
	Bcr_BETTYPE_BIG
	Bcr_BETTYPE_SMALL
	Bcr_BETTYPE_NONE
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

/*
type CountingResult struct {
	BUCode              string //BU 代碼
	GameIDDisplay       string //局號
	TableNo             uint8  //桌號
	SuggestionBet       string //建議下一局注別
	SuggestionBetAmount int16  //建議下一局下注金額
	Result              string //發牌結果
	GuessResult         bool   //猜測的結果
	HasInit             bool   //初始化算牌數據的 旗標
}
*/

type CountingResult struct {
	BUCode                    string                 //BU 代碼
	GameIDDisplay             string                 //局號
	TableNo                   uint8                  //桌號
	SuggestionBet             string                 //建議下一局注別
	SuggestionBetAmount       int16                  //建議下一局下注金額
	Result                    string                 //發牌結果
	GuessResult               bool                   //猜測的結果
	HasInit                   bool                   //初始化算牌數據的 旗標
	BetSuggestionMap          map[int]*BetSuggestion //計算建議用的參考統計值
	BetSuggestionSliceForSort BetSuggestionSlice     //排序用的

}

/*
//長龍
type CountingResultMethod2 struct {
	CountingResult
}
*/

//排序用的
type BetSuggestionSlice []*BetSuggestion

// Len is part of sort.Interface.
func (d BetSuggestionSlice) Len() int {
	return len(d)
}

// Swap is part of sort.Interface.
func (d BetSuggestionSlice) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

// Less is part of sort.Interface. We use count as the value to sort by
func (d BetSuggestionSlice) Less(i, j int) bool {
	return d[i].HouseEdge > d[j].HouseEdge
}

type BetSuggestion struct {
	BetType      uint8
	HouseEdge    float32 //要大於0才有搞頭(賭場優勢 (莊贏抽水0.05為例) ，若算到後來變正的 變賭場失去優勢)//半年才會碰到一次
	IsSuggestBet bool
}

func (currentCountingResult *CountingResult) ClearGuessResult() {
	currentCountingResult.SuggestionBet = ""
	currentCountingResult.Result = ""
	currentCountingResult.GuessResult = false
}

func (currentCountingResult *CountingResult) InitCountingData() {
	currentCountingResult.BetSuggestionMap[Bcr_BETTYPE_BANKER].HouseEdge = Bcr_BankerHouseEdgeDefault
	currentCountingResult.BetSuggestionMap[Bcr_BETTYPE_BANKER].IsSuggestBet = false
	currentCountingResult.BetSuggestionMap[Bcr_BETTYPE_PLAYER].HouseEdge = Bcr_PlayerHouseEdgeDefault
	currentCountingResult.BetSuggestionMap[Bcr_BETTYPE_PLAYER].IsSuggestBet = false
	currentCountingResult.BetSuggestionMap[Bcr_BETTYPE_TIE].HouseEdge = Bcr_TieHouseEdgeDefault
	currentCountingResult.BetSuggestionMap[Bcr_BETTYPE_TIE].IsSuggestBet = false
}
func TransBetTypeToStr(betType uint8) string {
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
