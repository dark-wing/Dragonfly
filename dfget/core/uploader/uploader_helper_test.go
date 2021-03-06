/*
 * Copyright The Dragonfly Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package uploader

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/dragonflyoss/Dragonfly/dfget/config"
	"github.com/dragonflyoss/Dragonfly/dfget/core/helper"
	"github.com/dragonflyoss/Dragonfly/dfget/util"
)

var (
	defaultRateLimit    = 1000
	defaultPieceSize    = int64(4 * 1024 * 1024)
	defaultPieceSizeStr = fmt.Sprintf("%d", defaultPieceSize)
)

func pc(origin string) string {
	return pieceContent(defaultPieceSize, origin)
}

func pieceContent(pieceSize int64, origin string) string {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(int64(len(origin))|(pieceSize<<4)))
	buf := bytes.Buffer{}
	buf.Write(b)
	buf.Write([]byte(origin))
	buf.Write([]byte{config.PieceTailChar})
	return buf.String()
}

// newTestPeerServer init the peer server for testing.
func newTestPeerServer(workHome string) {
	buf := &bytes.Buffer{}
	cfg := helper.CreateConfig(buf, workHome)
	p2p = newPeerServer(cfg, 0)
	p2p.totalLimitRate = 1000
	p2p.rateLimiter = util.NewRateLimiter(int32(defaultRateLimit), 2)
}

// initHelper create a temporary file and store it in the syncTaskMap.
func initHelper(fileName, workHome, content string) {
	helper.CreateTestFile(helper.GetServiceFile(fileName, workHome), content)
	p2p.syncTaskMap.Store(fileName, &taskConfig{
		dataDir:   workHome,
		rateLimit: defaultRateLimit,
	})
}

// ----------------------------------------------------------------------------
// handler helper

type HandlerHelper struct {
	method  string
	url     string
	body    io.Reader
	headers map[string]string
}

// ----------------------------------------------------------------------------
// upload header

var defaultUploadHeader = uploadHeader{
	rangeStr: fmt.Sprintf("bytes=0-%d", defaultPieceSize-1),
	num:      "0",
	size:     defaultPieceSizeStr,
}

type uploadHeader struct {
	rangeStr string
	num      string
	size     string
}

func (u uploadHeader) newRange(rangeStr string) uploadHeader {
	newU := u
	if !strings.HasPrefix(rangeStr, "bytes") {
		newU.rangeStr = "bytes=" + rangeStr
	} else {
		newU.rangeStr = rangeStr
	}
	return newU
}

func (u uploadHeader) newNum(num int) uploadHeader {
	newU := u
	newU.num = fmt.Sprintf("%d", num)
	return newU
}

func (u uploadHeader) newSize(size int) uploadHeader {
	newU := u
	newU.size = fmt.Sprintf("%d", size)
	return newU
}

// ----------------------------------------------------------------------------
// upload param

type uploadParamBuilder struct {
	up uploadParam
}

func (upb *uploadParamBuilder) build() *uploadParam {
	return &upb.up
}

func (upb *uploadParamBuilder) padSize(padSize int64) *uploadParamBuilder {
	upb.up.padSize = padSize
	return upb
}

func (upb *uploadParamBuilder) start(start int64) *uploadParamBuilder {
	upb.up.start = start
	return upb
}

func (upb *uploadParamBuilder) length(length int64) *uploadParamBuilder {
	upb.up.length = length
	return upb
}

func (upb *uploadParamBuilder) pieceSize(pieceSize int64) *uploadParamBuilder {
	upb.up.pieceSize = pieceSize
	return upb
}

func (upb *uploadParamBuilder) pieceNum(pieceNum int64) *uploadParamBuilder {
	upb.up.pieceNum = pieceNum
	return upb
}
