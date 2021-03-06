//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package execution

import (
	"fmt"
	"time"

	"github.com/couchbase/query/datastore"
	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/plan"
	"github.com/couchbase/query/value"
)

type SendInsert struct {
	base
	plan  *plan.SendInsert
	limit int64
}

func NewSendInsert(plan *plan.SendInsert) *SendInsert {
	rv := &SendInsert{
		base:  newBase(),
		plan:  plan,
		limit: -1,
	}

	rv.output = rv
	return rv
}

func (this *SendInsert) Accept(visitor Visitor) (interface{}, error) {
	return visitor.VisitSendInsert(this)
}

func (this *SendInsert) Copy() Operator {
	return &SendInsert{this.base.copy(), this.plan, this.limit}
}

func (this *SendInsert) RunOnce(context *Context, parent value.Value) {
	this.runConsumer(this, context, parent)
}

func (this *SendInsert) processItem(item value.AnnotatedValue, context *Context) bool {
	rv := this.limit != 0 && this.enbatch(item, this, context)

	if this.limit > 0 {
		this.limit--
	}

	return rv
}

func (this *SendInsert) beforeItems(context *Context, parent value.Value) bool {
	if this.plan.Limit() == nil {
		return true
	}

	limit, err := this.plan.Limit().Evaluate(parent, context)
	if err != nil {
		context.Error(errors.NewError(err, ""))
		return false
	}

	switch l := limit.Actual().(type) {
	case float64:
		this.limit = int64(l)
	default:
		context.Error(errors.NewInvalidValueError(fmt.Sprintf("Invalid LIMIT %v of type %T.", l, l)))
		return false
	}

	return true
}

func (this *SendInsert) afterItems(context *Context) {
	this.flushBatch(context)
}

func (this *SendInsert) flushBatch(context *Context) bool {
	defer this.releaseBatch()

	if len(this.batch) == 0 {
		return true
	}

	dpairs := _INSERT_POOL.Get()
	defer _INSERT_POOL.Put(dpairs)

	keyExpr := this.plan.Key()
	valExpr := this.plan.Value()
	var key, val value.Value
	var err error
	var ok bool
	i := 0

	for _, av := range this.batch {
		dpairs = dpairs[0 : i+1]
		dpair := &dpairs[i]

		if keyExpr != nil {
			// INSERT ... SELECT
			key, err = keyExpr.Evaluate(av, context)
			if err != nil {
				context.Error(errors.NewEvaluationError(err,
					fmt.Sprintf("INSERT key for %v", av.GetValue())))
				continue
			}

			if valExpr != nil {
				val, err = valExpr.Evaluate(av, context)
				if err != nil {
					context.Error(errors.NewEvaluationError(err,
						fmt.Sprintf("INSERT value for %v", av.GetValue())))
					continue
				}
			} else {
				val = av
			}
		} else {
			// INSERT ... VALUES
			key, ok = av.GetAttachment("key").(value.Value)
			if !ok {
				context.Error(errors.NewInsertKeyError(av.GetValue()))
				continue
			}

			val, ok = av.GetAttachment("value").(value.Value)
			if !ok {
				context.Error(errors.NewInsertValueError(av.GetValue()))
				continue
			}
		}

		dpair.Key, ok = key.Actual().(string)
		if !ok {
			context.Error(errors.NewInsertKeyTypeError(key))
			continue
		}

		dpair.Value = val
		i++
	}

	dpairs = dpairs[0:i]

	timer := time.Now()

	// Perform the actual INSERT
	keys, e := this.plan.Keyspace().Insert(dpairs)

	context.AddPhaseTime("insert", time.Since(timer))

	// Update mutation count with number of inserted docs
	context.AddMutationCount(uint64(len(keys)))

	if e != nil {
		context.Error(e)
	}

	// Capture the inserted keys in case there is a RETURNING clause
	for i, k := range keys {
		av := value.NewAnnotatedValue(make(map[string]interface{}))
		av.SetAttachment("meta", map[string]interface{}{"id": k})
		av.SetField(this.plan.Alias(), dpairs[i].Value)
		if !this.sendItem(av) {
			return false
		}
	}

	return true
}

func (this *SendInsert) readonly() bool {
	return false
}

var _INSERT_POOL = datastore.NewPairPool(_BATCH_SIZE)
