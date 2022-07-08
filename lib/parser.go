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
package lib

import (
	"io/ioutil"
	"os"
)

type Parser struct {
	Scanner ScannerI
	App     *AppBuild
	Scope   *Scope
}

type ParserI interface {
	Parse() error
	Load(fileName string) error
}

// default: Ninja parser
func NewParser(app *AppBuild, scope *Scope) ParserI {
	return &NinjaParser{
		&Parser{
			Scanner: NewScanner(SCANNER_NINJA, []byte{}),
			App:     app,
			Scope:   scope,
		},
	}
}

func NewParserWithScanner(scanner_type uint8) ParserI {
	return &NinjaParser{
		&Parser{Scanner: NewScanner(scanner_type, []byte{})},
	}
}

func (p *Parser) Load(fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}

	ct, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	p.Scanner.Reset(ct)
	return nil
}
