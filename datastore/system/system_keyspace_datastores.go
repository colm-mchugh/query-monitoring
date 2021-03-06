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
	"fmt"
	"strings"

	"github.com/couchbase/query/datastore"
	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/expression"
	"github.com/couchbase/query/logging"
	"github.com/couchbase/query/timestamp"
	"github.com/couchbase/query/value"
)

type storeKeyspace struct {
	namespace *namespace
	name      string
	si        datastore.Indexer
}

func (b *storeKeyspace) Release() {
}

func (b *storeKeyspace) NamespaceId() string {
	return b.namespace.Id()
}

func (b *storeKeyspace) Id() string {
	return b.Name()
}

func (b *storeKeyspace) Name() string {
	return b.name
}

func (b *storeKeyspace) Count() (int64, errors.Error) {
	return 1, nil
}

func (b *storeKeyspace) Indexer(name datastore.IndexType) (datastore.Indexer, errors.Error) {
	return b.si, nil
}

func (b *storeKeyspace) Indexers() ([]datastore.Indexer, errors.Error) {
	return []datastore.Indexer{b.si}, nil
}

func (b *storeKeyspace) Fetch(keys []string) ([]datastore.AnnotatedPair, []errors.Error) {
	var errs []errors.Error
	rv := make([]datastore.AnnotatedPair, 0, len(keys))
	for _, k := range keys {
		item, e := b.fetchOne(k)
		if e != nil {
			if errs == nil {
				errs = make([]errors.Error, 0, 1)
			}
			errs = append(errs, e)
			continue
		}

		if item != nil {
			item.SetAttachment("meta", map[string]interface{}{
				"id": k,
			})
		}

		rv = append(rv, datastore.AnnotatedPair{
			Key:   k,
			Value: item,
		})
	}

	return rv, errs
}

func (b *storeKeyspace) fetchOne(key string) (value.AnnotatedValue, errors.Error) {
	if key == b.namespace.store.actualStore.Id() {
		doc := value.NewAnnotatedValue(map[string]interface{}{
			"id":  b.namespace.store.actualStore.Id(),
			"url": b.namespace.store.actualStore.URL(),
		})
		return doc, nil
	}
	return nil, errors.NewSystemDatastoreError(nil, "Key Not Found "+key)
}

func (b *storeKeyspace) Insert(inserts []datastore.Pair) ([]datastore.Pair, errors.Error) {
	// FIXME
	return nil, errors.NewSystemNotImplementedError(nil, "")
}

func (b *storeKeyspace) Update(updates []datastore.Pair) ([]datastore.Pair, errors.Error) {
	// FIXME
	return nil, errors.NewSystemNotImplementedError(nil, "")
}

func (b *storeKeyspace) Upsert(upserts []datastore.Pair) ([]datastore.Pair, errors.Error) {
	// FIXME
	return nil, errors.NewSystemNotImplementedError(nil, "")
}

func (b *storeKeyspace) Delete(deletes []string) ([]string, errors.Error) {
	// FIXME
	return nil, errors.NewSystemNotImplementedError(nil, "")
}

func newStoresKeyspace(p *namespace) (*storeKeyspace, errors.Error) {
	b := new(storeKeyspace)
	b.namespace = p
	b.name = KEYSPACE_NAME_DATASTORES

	b.si = newSystemIndexer(b)
	b.si.CreatePrimaryIndex("", "#primary", nil)

	return b, nil
}

type systemIndexer struct {
	keyspace datastore.Keyspace
	indexes  map[string]datastore.Index
	primary  datastore.PrimaryIndex
}

func newSystemIndexer(keyspace datastore.Keyspace) datastore.Indexer {

	return &systemIndexer{
		keyspace: keyspace,
		indexes:  make(map[string]datastore.Index),
	}
}

func (si *systemIndexer) KeyspaceId() string {
	return si.keyspace.Id()
}

func (si *systemIndexer) Name() datastore.IndexType {
	return datastore.DEFAULT
}

func (si *systemIndexer) IndexIds() ([]string, errors.Error) {
	rv := make([]string, 0, len(si.indexes))
	for name, _ := range si.indexes {
		rv = append(rv, name)
	}
	return rv, nil
}

func (si *systemIndexer) IndexNames() ([]string, errors.Error) {
	rv := make([]string, 0, len(si.indexes))
	for name, _ := range si.indexes {
		rv = append(rv, name)
	}
	return rv, nil
}

func (si *systemIndexer) IndexById(id string) (datastore.Index, errors.Error) {
	return si.IndexByName(id)
}

func (si *systemIndexer) IndexByName(name string) (datastore.Index, errors.Error) {
	index, ok := si.indexes[name]
	if !ok {
		return nil, errors.NewSystemIdxNotFoundError(nil, name)
	}
	return index, nil
}

func (si *systemIndexer) PrimaryIndexes() ([]datastore.PrimaryIndex, errors.Error) {
	return []datastore.PrimaryIndex{si.primary}, nil
}

func (si *systemIndexer) Indexes() ([]datastore.Index, errors.Error) {
	return []datastore.Index{si.primary}, nil
}

func (si *systemIndexer) CreatePrimaryIndex(requestId, name string, with value.Value) (
	datastore.PrimaryIndex, errors.Error) {
	if si.primary == nil {
		pi := new(storeIndex)
		si.primary = pi
		pi.keyspace = si.keyspace.(*storeKeyspace)
		pi.name = name
		si.indexes[pi.name] = pi
	}

	return si.primary, nil
}

func (mi *systemIndexer) CreateIndex(requestId, name string, equalKey, rangeKey expression.Expressions,
	where expression.Expression, with value.Value) (datastore.Index, errors.Error) {
	return nil, errors.NewSystemNotSupportedError(nil, "CREATE INDEX is not supported for system datastore.")
}

func (mi *systemIndexer) BuildIndexes(requestId string, names ...string) errors.Error {
	return errors.NewSystemNotSupportedError(nil, "BUILD INDEXES is not supported for system datastore.")
}

func (mi *systemIndexer) Refresh() errors.Error {
	return nil
}

func (mi *systemIndexer) SetLogLevel(level logging.Level) {
	// No-op, uses query engine logger
}

type storeIndex struct {
	name     string
	keyspace *storeKeyspace
}

func (pi *storeIndex) KeyspaceId() string {
	return pi.name
}

func (pi *storeIndex) Id() string {
	return pi.Name()
}

func (pi *storeIndex) Name() string {
	return pi.name
}

func (pi *storeIndex) Type() datastore.IndexType {
	return datastore.DEFAULT
}

func (pi *storeIndex) SeekKey() expression.Expressions {
	return nil
}

func (pi *storeIndex) RangeKey() expression.Expressions {
	return nil
}

func (pi *storeIndex) Condition() expression.Expression {
	return nil
}

func (pi *storeIndex) IsPrimary() bool {
	return true
}

func (pi *storeIndex) State() (state datastore.IndexState, msg string, err errors.Error) {
	return datastore.ONLINE, "", nil
}

func (pi *storeIndex) Statistics(requestId string, span *datastore.Span) (
	datastore.Statistics, errors.Error) {
	return nil, nil
}

func (pi *storeIndex) Drop(requestId string) errors.Error {
	return errors.NewSystemIdxNoDropError(nil, pi.Name())
}

func (pi *storeIndex) Scan(requestId string, span *datastore.Span, distinct bool, limit int64,
	cons datastore.ScanConsistency, vector timestamp.Vector, conn *datastore.IndexConnection) {
	defer close(conn.EntryChannel())

	val := ""

	a := span.Seek[0].Actual()
	switch a := a.(type) {
	case string:
		val = a
	default:
		conn.Error(errors.NewSystemDatastoreError(nil, fmt.Sprintf("Invalid seek value %v of type %T.", a, a)))
		return
	}

	if strings.EqualFold(val, pi.keyspace.namespace.store.actualStore.Id()) {
		entry := datastore.IndexEntry{PrimaryKey: pi.keyspace.namespace.store.actualStore.Id()}
		conn.EntryChannel() <- &entry
	}
}

func (pi *storeIndex) ScanEntries(requestId string, limit int64, cons datastore.ScanConsistency,
	vector timestamp.Vector, conn *datastore.IndexConnection) {
	defer close(conn.EntryChannel())

	entry := datastore.IndexEntry{PrimaryKey: pi.keyspace.namespace.store.actualStore.Id()}
	conn.EntryChannel() <- &entry
}
