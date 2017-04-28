// Copyright 2013 Beego Samples authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package models

type BetType int

const (
	EVENT_BANKER = iota
	EVENT_PLAYER
	EVENT_TIE
	EVENT_BIG
	EVENT_SMALL
)

type CountingResult struct {
	BUCode              string           //BU 代碼
	GameIDDisplay       string           //局號
	TableNo             uint8            //桌號
	BetSuggestionData   [2]BetSuggestion //建議值
	SuggestionBet       string
	SuggestionBetAmount int16
	Result              string
	GuessResult         bool
}

type BetSuggestion struct {
	BetType        uint8
	WinProbability float32 //要大於0才有搞頭
	SuggestBet     bool
}
