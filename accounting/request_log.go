//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

/*

 Packace accounting provides a common API for workload and monitoring data - metrics, statistics, events.
*/
package accounting

import (
	"sync"
	"time"

	"github.com/couchbase/query/plan"
)

type RequestLogEntry struct {
	RequestId    string
	ElapsedTime  time.Duration
	ServiceTime  time.Duration
	Statement    string
	Plan         *plan.Prepared
	ResultCount  int
	ResultSize   int
	ErrorCount   int
	SortCount    uint64
	PreparedName string
	PreparedText string
	Time         time.Time
}

type RequestLog struct {
	sync.RWMutex
	requests map[string]*RequestLogEntry
}

const _CACHE_SIZE = 1 << 10

var requestLog = &RequestLog{
	requests: make(map[string]*RequestLogEntry, _CACHE_SIZE),
}

func (this *RequestLog) add(entry *RequestLogEntry) {
	this.Lock()
	defer this.Unlock()
	this.requests[entry.RequestId] = entry
}

func (this *RequestLog) get(id string) *RequestLogEntry {
	this.RLock()
	defer this.RUnlock()
	return this.requests[id]
}

func (this *RequestLog) size() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.requests)
}

func (this *RequestLog) names() []string {
	i := 0
	this.RLock()
	defer this.RUnlock()
	n := make([]string, len(this.requests))
	for k := range this.requests {
		n[i] = k
		i++
	}
	return n
}

func (this *RequestLog) forEach(f func(string, *RequestLogEntry)) {
	this.RLock()
	defer this.RUnlock()
	for id, entry := range this.requests {
		f(id, entry)
	}
}

func RequestsEntry(id string) *RequestLogEntry {
	return requestLog.get(id)
}

func RequestsCount() int {
	return requestLog.size()
}

func RequestIds() []string {
	return requestLog.names()
}

func RequestsForeach(f func(string, *RequestLogEntry)) {
	requestLog.forEach(f)
}

func LogRequest(acctstore AccountingStore,
	request_time time.Duration, service_time time.Duration,
	result_count int, result_size int,
	error_count int, warn_count int, stmt string,
	sort_count uint64, plan *plan.Prepared, id string) {

	// Don't log requests < 500ms. TODO: make configurable.
	if request_time < time.Millisecond*500 {
		return
	}

	rv := &RequestLogEntry{
		RequestId:   id,
		ElapsedTime: request_time,
		ServiceTime: service_time,
		ResultCount: result_count,
		ResultSize:  result_size,
		ErrorCount:  error_count,
		SortCount:   sort_count,
		Time:        time.Now(),
	}
	if stmt != "" {
		rv.Statement = stmt
	}
	if plan != nil {
		rv.PreparedName = plan.Name()
		rv.PreparedText = plan.Text()
	}
	requestLog.add(rv)
}
