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
	ContinueBetAmount float64 //累進下注金額
	MaxLoseLimit      float64 //最大累進損失金額
	BetRecordList     map[string]BetRecord
}

func (BetAccount *SimBetAccount) initBetAccount() {

	BetAccount.LoginTime = time.Now()
	BetAccount.BetRecordList = make(map[string]BetRecord)
	//PublishAccountBalance(BetAccount)
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
