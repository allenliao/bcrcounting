//百家樂第一種計算方法
//紀錄每一張牌，並計算出每種Bet type的賭場優勢的影響

package countingmethod

import (
	"bcrcounting/models"
	"fmt"
	"sort"

	"github.com/astaxie/beego"
)

var (
	betSuggestionSliceForSort betSuggestionSlice
)

//排序用的
type betSuggestionSlice []*models.BetSuggestion

// Len is part of sort.Interface.
func (d betSuggestionSlice) Len() int {
	return len(d)
}

// Swap is part of sort.Interface.
func (d betSuggestionSlice) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

// Less is part of sort.Interface. We use count as the value to sort by
func (d betSuggestionSlice) Less(i, j int) bool {
	return d[i].HouseEdge > d[j].HouseEdge
}

func CreateCountingResult(BUCode string, tableNo uint8) models.CountingResult {
	var betSuggestionMap = make(map[int]*models.BetSuggestion)
	betSuggestionMap[models.Bcr_BETTYPE_BANKER] = &models.BetSuggestion{
		BetType:      models.Bcr_BETTYPE_BANKER,
		HouseEdge:    models.Bcr_BankerHouseEdgeDefault,
		IsSuggestBet: false}
	betSuggestionMap[models.Bcr_BETTYPE_PLAYER] = &models.BetSuggestion{
		BetType:      models.Bcr_BETTYPE_PLAYER,
		HouseEdge:    models.Bcr_PlayerHouseEdgeDefault,
		IsSuggestBet: false}
	betSuggestionMap[models.Bcr_BETTYPE_TIE] = &models.BetSuggestion{
		BetType:      models.Bcr_BETTYPE_TIE,
		HouseEdge:    models.Bcr_TieHouseEdgeDefault,
		IsSuggestBet: false}

	currentCountingResult := models.CountingResult{
		BUCode:              BUCode,
		TableNo:             tableNo,
		BetSuggestionMap:    betSuggestionMap,
		SuggestionBet:       "",
		SuggestionBetAmount: 100,
		Result:              "",
		GuessResult:         false}

	betSuggestionSliceForSort = make(betSuggestionSlice, 0, len(betSuggestionMap))

	for _, betSuggestion_adr := range betSuggestionMap {
		betSuggestionSliceForSort = append(betSuggestionSliceForSort, betSuggestion_adr)
	}

	return currentCountingResult
}

//換靴時要呼叫
func resetCountingResult(countingResult models.CountingResult) {
	//countingResult
}

//紀錄每一張牌，並計算出每種Bet type的賭場優勢的影響
//
func Bcr_CountingMethod1(cardList [6]int, currentCountingResult *models.CountingResult) *models.BetSuggestion {

	for _, point := range cardList { //idx, card point
		if point == -1 {
			continue
		}
		currentCountingResult.BetSuggestionMap[models.Bcr_BETTYPE_PLAYER].HouseEdge += models.Bcr_PlayerHouseEdgeEffectList[point]
		currentCountingResult.BetSuggestionMap[models.Bcr_BETTYPE_BANKER].HouseEdge += models.Bcr_BankerHouseEdgeEffectList[point]
		currentCountingResult.BetSuggestionMap[models.Bcr_BETTYPE_TIE].HouseEdge += models.Bcr_TieHouseEdgeEffectList[point]
	}

	currentCountingResult.BetSuggestionMap[models.Bcr_BETTYPE_PLAYER].HouseEdge = -5
	currentCountingResult.BetSuggestionMap[models.Bcr_BETTYPE_BANKER].HouseEdge = 2
	currentCountingResult.BetSuggestionMap[models.Bcr_BETTYPE_TIE].HouseEdge = -1

	//HouseEdge排序算出結果(越大越好)
	sort.Sort(betSuggestionSliceForSort)
	betSuggestion := betSuggestionSliceForSort[0] //最大的

	for idx, betSuggestion := range betSuggestionSliceForSort {
		beego.Info("[" + fmt.Sprint(idx) + "]betSuggestion BetType:" + fmt.Sprint(betSuggestion.BetType) + " HouseEdge:" + fmt.Sprint(betSuggestion.HouseEdge))
	}

	if betSuggestion.HouseEdge > 0 { //擊敗賭場優勢
		betSuggestion.IsSuggestBet = true
		currentCountingResult.SuggestionBet = models.TransBetTypeToStr(betSuggestion.BetType) //建議下一局買甚麼
		return betSuggestion
	}

	return nil

}
