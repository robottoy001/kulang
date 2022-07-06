/*
 * Copyright (c) 2022 Huawei Device Co., Ltd.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package main

const (
	SCANNER_NINJA uint8 = iota
)

const (
	INVALID_CHAR = 255
)

type Location struct {
	LineNo    uint32
	Start     uint32
	End       uint32
	LineStart uint32
}

type Token struct {
	Type    uint8
	Literal string
	Loc     Location
}

type Position struct {
	Offset    uint32
	LineNo    uint32
	LineStart uint32
}

type Scanner struct {
	Content   []uint8
	Token     Token
	LastToken Token
	Pos       *Position
}

type ScannerI interface {
	NextToken() Token
	PeekToken(uint8) bool
	BackwardToken()
	GetToken() Token
	Reset([]byte)
	ScanVarValue(bool) (*VarString, error)
	ScanIdent() Token
}

func NewScanner(scanner_type uint8, content []byte) ScannerI {
	switch scanner_type {
	case SCANNER_NINJA:
		return &NinjaScanner{&Scanner{
			Content: content,
			Pos: &Position{
				Offset: 0,
				LineNo: 0,
			},
		},
		}
	default:
		return nil
	}
	return nil
}
