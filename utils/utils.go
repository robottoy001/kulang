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
package utils

import (
	"bytes"
	"encoding/binary"
)

var tag string = "Kulang"

const (
	EdgeSlot uint8 = iota
	InvalidSlot
)

var ids []int = make([]int, InvalidSlot)

func GetId(slot uint8) int {
	var id int
	switch slot {
	case EdgeSlot:
		id = ids[EdgeSlot]
		id++
		ids[EdgeSlot] = id
	default:
		return -1
	}
	return id
}

const (
	seed uint64 = 0xDECAFBADDECAFBAD
	m    uint64 = 0xC6A4a7935BD1E995
)

func Hash(data []byte) uint64 {
	var rotate uint8 = 47
	var length = uint64(len(data))
	var h uint64 = seed ^ (uint64(len(data)) * m)

	buf := bytes.NewBuffer(data)
	for length >= 8 {
		var tmp uint64
		err := binary.Read(buf, binary.LittleEndian, &tmp)
		if err != nil {
			break
		}
		tmp = tmp * m
		tmp = tmp ^ (tmp >> rotate)
		tmp = tmp * m
		h = h ^ tmp
		h = h * m
		length -= 8
	}

	lastBytes := data[len(data)-int(length):]

	switch length & 7 {
	case 7:
		h = h ^ (uint64(lastBytes[6]) << 48)
		fallthrough
	case 6:
		h = h ^ (uint64(lastBytes[5]) << 40)
		fallthrough
	case 5:
		h = h ^ (uint64(lastBytes[4]) << 32)
		fallthrough
	case 4:
		h = h ^ (uint64(lastBytes[3]) << 24)
		fallthrough
	case 3:
		h = h ^ (uint64(lastBytes[2]) << 16)
		fallthrough
	case 2:
		h = h ^ (uint64(lastBytes[1]) << 8)
		fallthrough
	case 1:
		h = h ^ (uint64(lastBytes[0]))
		h = h * m
	}

	h = h ^ (h >> rotate)
	h = h * m
	h = h ^ (h >> rotate)

	return h
}
