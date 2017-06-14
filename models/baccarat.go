package models

import (
	"fmt"
	"goutils"
	"sort"
	"strings"
)

type BetType uint8

const (
	Bcr_BETTYPE_BANKER = iota //0
	Bcr_BETTYPE_PLAYER        //1
	Bcr_BETTYPE_TIE           //2
	Bcr_BETTYPE_BIG
	Bcr_BETTYPE_SMALL
	Bcr_BETTYPE_NONE
)

var (
	maxBanker, maxTie, maxPlayer float64 = -0.007925005629658699, -0.11981399357318878, -0.009793972596526146
	Bcr_BetTypeCount             uint8   = 5
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

//百家樂第一種計算方法
//紀錄每一張牌，並計算出每種Bet type的賭場優勢的影響
//不符成本效益，放棄此算法

func CreateCurrentCountingResultList(BUCode string, tableNo uint8) map[string]CountingResultInterface {
	//cardCountingMethod := CountingResultMethod1{}//不符成本效益，放棄此算法
	longtrendMethod := CountingResultMethod2{}
	//cardCountingMethod_addr := &cardCountingMethod//不符成本效益，放棄此算法
	longtrendMethod_addr := &longtrendMethod
	//methodObj_addr.InitBaseField(BUCode, tableNo) //沒有改變自身的屬性值 要想辦法為Addr進去不然就是這樣

	//注冊算法
	currentCountingResultList := map[string]CountingResultInterface{
		//"cardCounting": cardCountingMethod_addr,//不符成本效益，放棄此算法
		"longtrend": longtrendMethod_addr}

	for _, methodObj_addr := range currentCountingResultList {
		//methodObj.InitBaseField(BUCode, tableNo)
		methodObj_addr.InitBaseField(BUCode, tableNo) //沒有改變自身的屬性值 要想辦法為Addr進去不然就是這樣
		methodObj_addr.InitCustomField()              //初始化個別算法所需要的參數
		currentCountingResult := methodObj_addr.GetCountingResult()
		goutils.Logger.Info("CreateCurrentCountingResultList currentCountingResult.BUCode:" + currentCountingResult.BUCode)
	}

	return currentCountingResultList

}

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
	PatternName  string
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

type CountingResultInterface interface {
	Counting(cardList [6]int, beadRoadStr string) bool
	InitBaseField(BUCode string, tableNo uint8)
	InitCustomField()
	ClearGuessResult()
	InitChangShoeField()
	GetCountingResult() *CountingResult
}

type CountingResult struct {
	BUCode              string  //BU 代碼
	GameIDDisplay       string  //局號
	TableNo             uint8   //桌號
	SuggestionBet       uint8   //建議下一局注別
	SuggestionBetAmount float64 //建議下一局下注金額
	DefaultBetAmount    float64 //預設下注金額
	TrendName           string  //趨勢名稱
	Result              uint8   //發牌結果
	GuessResult         bool    //猜測的結果
	TieReturn           bool    //開和 若壓莊閒 須返水
	FirstHand           bool    //第一局結果(無法預測不公佈)
	HasInit             bool    //初始化算牌數據的 旗標
	HasBeted            bool    //被下過注了
	MethodName          string
	MethodID            string
	DubleBet            bool //倍投
	DubleBetWhenWin     bool //贏了倍投
	NextBetDubleBet     bool //下一注倍投
	GotCard             bool //拿到CardList了
	GotResult           bool //拿到Result了
	CardList            [6]int
	BeadRoadStr         string
}

func (currentCountingResult *CountingResult) GetCountingResult() *CountingResult {
	return currentCountingResult
}

func (currentCountingResult *CountingResult) InitBaseField(BUCode string, tableNo uint8) {
	currentCountingResult.BUCode = BUCode
	currentCountingResult.TableNo = tableNo
	currentCountingResult.SuggestionBet = Bcr_BETTYPE_NONE
	currentCountingResult.SuggestionBetAmount = 100
	currentCountingResult.DefaultBetAmount = 100
	currentCountingResult.Result = Bcr_BETTYPE_NONE
	currentCountingResult.GuessResult = false
	currentCountingResult.TieReturn = false
	currentCountingResult.FirstHand = false
	currentCountingResult.DubleBet = false
	currentCountingResult.DubleBetWhenWin = false
	currentCountingResult.NextBetDubleBet = false
	currentCountingResult.GotCard = false
	currentCountingResult.GotResult = false

}

func (currentCountingResult *CountingResult) InitCustomField(BUCode string, tableNo uint8) {

}

func (currentCountingResult *CountingResult) InitChangShoeField() {
	currentCountingResult.ClearGuessResult()
	currentCountingResult.StopDubleBet()
}

//Type繼承CountingResult的 model都可以用?
//該桌當局RESULT沒預測結果時會執行
func (currentCountingResult *CountingResult) ClearGuessResult() {
	currentCountingResult.SuggestionBet = Bcr_BETTYPE_NONE
	//currentCountingResult.Result = ""
	//currentCountingResult.GuessResult = false
	//currentCountingResult.TieReturn = false
}

//取結果時 依具Method設定的參數 或 上一次沒有預測 時會執行
func (currentCountingResult *CountingResult) StopDubleBet() {
	currentCountingResult.SuggestionBetAmount = currentCountingResult.DefaultBetAmount
	currentCountingResult.NextBetDubleBet = false
}
func (currentCountingResult *CountingResult) Counting(cardList [6]int, beadRoadStr string) bool {
	return false
}

//用卡牌計算賭場優勢
type CountingResultMethod1 struct {
	CountingResult
	BetSuggestionMap          map[int]*BetSuggestion //計算建議用的參考統計值
	BetSuggestionSliceForSort BetSuggestionSlice     //排序用的
}

//引擎一開始時的初始化
func (currentCountingResult *CountingResultMethod1) InitCustomField() {
	currentCountingResult.MethodID = "M1"
	currentCountingResult.MethodName = "賭場優勢"
	var betSuggestionMap = make(map[int]*BetSuggestion)
	betSuggestionMap[Bcr_BETTYPE_BANKER] = &BetSuggestion{
		BetType:      Bcr_BETTYPE_BANKER,
		HouseEdge:    Bcr_BankerHouseEdgeDefault,
		PatternName:  "莊家失去賭場優勢",
		IsSuggestBet: false}
	betSuggestionMap[Bcr_BETTYPE_PLAYER] = &BetSuggestion{
		BetType:      Bcr_BETTYPE_PLAYER,
		HouseEdge:    Bcr_PlayerHouseEdgeDefault,
		PatternName:  "閒家失去賭場優勢",
		IsSuggestBet: false}
	betSuggestionMap[Bcr_BETTYPE_TIE] = &BetSuggestion{
		BetType:      Bcr_BETTYPE_TIE,
		HouseEdge:    Bcr_TieHouseEdgeDefault,
		PatternName:  "和失去賭場優勢",
		IsSuggestBet: false}

	currentCountingResult.BetSuggestionMap = betSuggestionMap
	currentCountingResult.BetSuggestionSliceForSort = make(BetSuggestionSlice, 0, len(betSuggestionMap))

	for _, betSuggestion_adr := range betSuggestionMap {
		goutils.Logger.Info("CountingResultMethod1.InitBaseField betSuggestion_adr:" + fmt.Sprint(betSuggestion_adr))
		currentCountingResult.BetSuggestionSliceForSort = append(currentCountingResult.BetSuggestionSliceForSort, betSuggestion_adr)
	}
}

func (currentCountingResult *CountingResultMethod1) InitChangShoeField() {
	//goutils.Logger.Info("CountingResultMethod1.InitChangShoeField BUCode:" + currentCountingResult.BUCode + " TableNo:" + fmt.Sprint(currentCountingResult.TableNo))

	if currentCountingResult.BetSuggestionMap == nil {
		goutils.Logger.Info("CountingResultMethod1.BetSuggestionMap==nil")
	}
	//goutils.Logger.Info("CountingResultMethod1.BetSuggestionMap[Bcr_BETTYPE_BANKER]" + fmt.Sprint(currentCountingResult.BetSuggestionMap[Bcr_BETTYPE_BANKER]))
	currentCountingResult.BetSuggestionMap[Bcr_BETTYPE_BANKER].HouseEdge = Bcr_BankerHouseEdgeDefault
	currentCountingResult.BetSuggestionMap[Bcr_BETTYPE_BANKER].IsSuggestBet = false
	currentCountingResult.BetSuggestionMap[Bcr_BETTYPE_PLAYER].HouseEdge = Bcr_PlayerHouseEdgeDefault
	currentCountingResult.BetSuggestionMap[Bcr_BETTYPE_PLAYER].IsSuggestBet = false
	currentCountingResult.BetSuggestionMap[Bcr_BETTYPE_TIE].HouseEdge = Bcr_TieHouseEdgeDefault
	currentCountingResult.BetSuggestionMap[Bcr_BETTYPE_TIE].IsSuggestBet = false
}

//紀錄每一張牌，並計算出每種Bet type的賭場優勢的影響
//Bcr_CountingMethod1
func (currentCountingResult *CountingResultMethod1) Counting(cardList [6]int, beadRoadStr string) bool {
	//goutils.Logger.Info("CountingResultMethod1.Counting" + currentCountingResult.BUCode + " TableNo:" + fmt.Sprint(currentCountingResult.TableNo))

	for _, point := range cardList { //idx, card point
		if point == -1 {
			continue
		}
		if currentCountingResult.BetSuggestionMap == nil {
			goutils.Logger.Info("currentCountingResult.BetSuggestionMap==nil")
		}
		currentCountingResult.BetSuggestionMap[Bcr_BETTYPE_PLAYER].HouseEdge += Bcr_PlayerHouseEdgeEffectList[point]
		currentCountingResult.BetSuggestionMap[Bcr_BETTYPE_BANKER].HouseEdge += Bcr_BankerHouseEdgeEffectList[point]
		currentCountingResult.BetSuggestionMap[Bcr_BETTYPE_TIE].HouseEdge += Bcr_TieHouseEdgeEffectList[point]
	}

	//HouseEdge排序算出結果(越大越好)
	sort.Sort(currentCountingResult.BetSuggestionSliceForSort)
	//betSuggestion := currentCountingResult.BetSuggestionSliceForSort[0] //最大的

	goutils.Logger.Info("TableNo:" + fmt.Sprint(currentCountingResult.TableNo) + "maxBanker:" + fmt.Sprint(maxBanker) + " maxPlayer:" + fmt.Sprint(maxPlayer) + " maxTie:" + fmt.Sprint(maxTie))
	result := false
	for _, betSuggestion := range currentCountingResult.BetSuggestionSliceForSort {
		//goutils.Logger.Info("[" + fmt.Sprint(idx) + "]betSuggestion BetType:" + fmt.Sprint(betSuggestion.BetType) + " HouseEdge:" + fmt.Sprint(betSuggestion.HouseEdge))

		if hitHouseEdge(betSuggestion) { //擊敗賭場優勢 //除非有退庸，不然不可能HouseEdge>0

			betSuggestion.IsSuggestBet = true
			currentCountingResult.SuggestionBet = betSuggestion.BetType //建議下一局買甚麼
			currentCountingResult.TrendName = betSuggestion.PatternName
			//TODO:建議買多少錢
			//TransBetTypeToStr(betSuggestion.BetType)
			result = true
			break
		}
	}

	return result

}

func hitHouseEdge(betSuggestion *BetSuggestion) bool {
	houseEdge := float64(betSuggestion.HouseEdge)
	if betSuggestion.BetType == Bcr_BETTYPE_BANKER {
		if houseEdge > maxBanker {
			maxBanker = houseEdge
			return true
		}
		goutils.Logger.Info("Banker houseEdge:" + fmt.Sprint(houseEdge))
	}
	if betSuggestion.BetType == Bcr_BETTYPE_PLAYER {
		if houseEdge > maxPlayer {
			maxPlayer = houseEdge
			return true
		}
		goutils.Logger.Info("Player houseEdge:" + fmt.Sprint(houseEdge))
	}
	if betSuggestion.BetType == Bcr_BETTYPE_TIE {
		if houseEdge > maxTie {
			maxTie = houseEdge
			return true
		}
		goutils.Logger.Info("Tie houseEdge:" + fmt.Sprint(houseEdge))
	}

	return false
}

//長龍
type CountingResultMethod2 struct {
	CountingResult
	RoadPatternInfoList [4]RoadPatternInfo
}

type RoadPatternInfo struct {
	Pattern string
	//HitCount      uint8 //連續出現幾次
	SuggestionBetType uint8
	PatternName       string
}

//之後可以透過呼叫這個方法餵客製化參數進來
func (currentCountingResult *CountingResultMethod2) InitCustomField() {
	currentCountingResult.MethodID = "M2"
	currentCountingResult.MethodName = "連8斬龍" //連7斬龍 >>1個禮拜24小時不間斷 失敗過一次連18龍 平均一天24H 賺2500~3000
	currentCountingResult.DubleBet = true
	currentCountingResult.DubleBetWhenWin = false //輸了倍投
	currentCountingResult.RoadPatternInfoList[0] = RoadPatternInfo{Pattern: "00000000", SuggestionBetType: 1, PatternName: "長莊"}
	currentCountingResult.RoadPatternInfoList[1] = RoadPatternInfo{Pattern: "11111111", SuggestionBetType: 0, PatternName: "長閒"}
	currentCountingResult.RoadPatternInfoList[2] = RoadPatternInfo{Pattern: "0101010101", SuggestionBetType: 0, PatternName: "莊閒長跳"}
	currentCountingResult.RoadPatternInfoList[3] = RoadPatternInfo{Pattern: "1010101010", SuggestionBetType: 1, PatternName: "閒莊長跳"}

}

//用字串搜尋的 方法
func (currentCountingResult *CountingResultMethod2) Counting(cardList [6]int, beadRoadStr string) bool {
	result := false

	if currentCountingResult.NextBetDubleBet {
		if currentCountingResult.Result == Bcr_BETTYPE_TIE {
			currentCountingResult.SuggestionBetAmount = currentCountingResult.SuggestionBetAmount //開和維持原投注
		} else {
			currentCountingResult.SuggestionBetAmount = currentCountingResult.SuggestionBetAmount * 2
		}

		result = true
		return result
	}

	for _, roadPatternInfo := range currentCountingResult.RoadPatternInfoList {
		Pattern := roadPatternInfo.Pattern
		var endIdx = len(beadRoadStr) - len(Pattern)
		var pIdx = strings.LastIndex(beadRoadStr, Pattern)
		if endIdx == pIdx && endIdx > 0 && pIdx > 0 {
			goutils.Logger.Info("TableNo:" + fmt.Sprint(currentCountingResult.TableNo) + " Pattern:" + Pattern + " endIdx:" + fmt.Sprint(endIdx) + " pIdx:" + fmt.Sprint(pIdx) + " roadPatternInfo.SuggestionBetType:" + TransBetTypeToStr(roadPatternInfo.SuggestionBetType))
			currentCountingResult.SuggestionBet = roadPatternInfo.SuggestionBetType
			currentCountingResult.TrendName = roadPatternInfo.PatternName
			//TransBetTypeToStr(roadPatternInfo.SuggestionBetType) //建議下一局買甚麼
			//SuggestionBetAmount
			result = true
			break
		}

	}

	return result
}
