//  Copyright (c) 2013 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package system

import (
	"time"

	"github.com/couchbase/query/datastore"
	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/server"
	"github.com/couchbase/query/timestamp"
	"github.com/couchbase/query/value"
)

type activeRequestsKeyspace struct {
	namespace *namespace
	name      string
	indexer   datastore.Indexer
}

func (b *activeRequestsKeyspace) Release() {
}

func (b *activeRequestsKeyspace) NamespaceId() string {
	return b.namespace.Id()
}

func (b *activeRequestsKeyspace) Id() string {
	return b.Name()
}

func (b *activeRequestsKeyspace) Name() string {
	return b.name
}

func (b *activeRequestsKeyspace) Count() (int64, errors.Error) {
	c, err := server.ActiveRequestsCount()
	return int64(c), err
}

func (b *activeRequestsKeyspace) Indexer(name datastore.IndexType) (datastore.Indexer, errors.Error) {
	return b.indexer, nil
}

func (b *activeRequestsKeyspace) Indexers() ([]datastore.Indexer, errors.Error) {
	return []datastore.Indexer{b.indexer}, nil
}

func (b *activeRequestsKeyspace) Fetch(keys []string) ([]datastore.AnnotatedPair, []errors.Error) {
	var errs []errors.Error
	rv := make([]datastore.AnnotatedPair, 0, len(keys))

	server.ActiveRequestsForEach(func(id string, request server.Request) {
		item := value.NewAnnotatedValue(map[string]interface{}{
			"RequestId":     id,
			"RequestTime":   request.RequestTime().String(),
			"ElapsedTime":   time.Since(request.RequestTime()).String(),
			"ExecutionTime": time.Since(request.ServiceTime()).String(),
			"State":         request.State(),
		})
		if request.Statement() != "" {
			item.SetField("Statement", request.Statement())
		}
		if request.Prepared() != nil {
			p := request.Prepared()
			item.SetField("Prepared.Name", p.Name())
			item.SetField("Prepared.Text", p.Text())
		}
		item.SetAttachment("meta", map[string]interface{}{
			"id": id,
		})
		rv = append(rv, datastore.AnnotatedPair{
			Key:   id,
			Value: item,
		})
	})
	return rv, errs
}

func (b *activeRequestsKeyspace) Insert(inserts []datastore.Pair) ([]datastore.Pair, errors.Error) {
	// FIXME
	return nil, errors.NewSystemNotImplementedError(nil, "")
}

func (b *activeRequestsKeyspace) Update(updates []datastore.Pair) ([]datastore.Pair, errors.Error) {
	// FIXME
	return nil, errors.NewSystemNotImplementedError(nil, "")
}

func (b *activeRequestsKeyspace) Upsert(upserts []datastore.Pair) ([]datastore.Pair, errors.Error) {
	// FIXME
	return nil, errors.NewSystemNotImplementedError(nil, "")
}

func (b *activeRequestsKeyspace) Delete(deletes []string) ([]string, errors.Error) {
	// FIXME
	return nil, errors.NewSystemNotImplementedError(nil, "")
}

func newActiveRequestsKeyspace(p *namespace) (*activeRequestsKeyspace, errors.Error) {
	b := new(activeRequestsKeyspace)
	b.namespace = p
	b.name = KEYSPACE_NAME_ACTIVE

	primary := &activeRequestsIndex{name: "#primary", keyspace: b}
	b.indexer = &systemIndexer{keyspace: b, indexes: make(map[string]datastore.Index), primary: primary}

	return b, nil
}

type activeRequestsIndex struct {
	name     string
	keyspace *activeRequestsKeyspace
}

func (pi *activeRequestsIndex) KeyspaceId() string {
	return pi.keyspace.Id()
}

func (pi *activeRequestsIndex) Id() string {
	return pi.Name()
}

func (pi *activeRequestsIndex) Name() string {
	return pi.name
}

func (pi *activeRequestsIndex) Type() datastore.IndexType {
	return datastore.DEFAULT
}

func (pi *activeRequestsIndex) SeekKey() expression.Expressions {
	return nil
}

func (pi *activeRequestsIndex) RangeKey() expression.Expressions {
	return nil
}

func (pi *activeRequestsIndex) Condition() expression.Expression {
	return nil
}

func (pi *activeRequestsIndex) IsPrimary() bool {
	return true
}

func (pi *activeRequestsIndex) State() (state datastore.IndexState, msg string, err errors.Error) {
	return datastore.ONLINE, "", nil
}

func (pi *activeRequestsIndex) Statistics(requestId string, span *datastore.Span) (
	datastore.Statistics, errors.Error) {
	return nil, nil
}

func (pi *activeRequestsIndex) Drop(requestId string) errors.Error {
	return errors.NewSystemIdxNoDropError(nil, "")
}

func (pi *activeRequestsIndex) Scan(requestId string, span *datastore.Span, distinct bool, limit int64,
	cons datastore.ScanConsistency, vector timestamp.Vector, conn *datastore.IndexConnection) {
	defer close(conn.EntryChannel())
	// NOP
}

func (pi *activeRequestsIndex) ScanEntries(requestId string, limit int64, cons datastore.ScanConsistency,
	vector timestamp.Vector, conn *datastore.IndexConnection) {
	defer close(conn.EntryChannel())
	numRequests, err := server.ActiveRequestsCount()
	if err != nil {
		conn.Error(err)
		return
	}
	requestIds := make([]string, numRequests)
	server.ActiveRequestsForEach(func(id string, request server.Request) {
		requestIds = append(requestIds, id)
	})

	for _, name := range requestIds {
		entry := datastore.IndexEntry{PrimaryKey: name}
		conn.EntryChannel() <- &entry
	}
}
