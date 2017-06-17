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

import (
	"time"
)

//模擬下注的 帳號
type SimBetAccount struct {
	Balance           float64
	LoginTime         time.Time
	LogoutTime        time.Time
	BetRecordList     map[string]BetRecord
	TotalBetStatistic *BetStatistic
	SubBetStatistic   *BetStatistic
}

type BetRecord struct {
	BetTime           time.Time
	BUCode            string
	TableNo           uint8
	GameIDDisplay     string //局號
	GameResultType    uint8
	GameResultTypeStr string
	BetType           uint8
	BetTypeStr        string
	BetAmmount        float64
	WinAmmount        float64
	Settled           bool
	CurrentBalance    float64
}

type BetStatistic struct {
	BetCount            uint8     //下注次數
	WinBetCount         uint8     //贏的次數
	LoseBetCount        uint8     //輸的次數
	TieBetCount         uint8     //平的次數
	AccumulateBetAmount float64   //累計下注金額
	TotalWinAmount      float64   //累計贏得金額
	StartTime           time.Time //開始統計時間
}
