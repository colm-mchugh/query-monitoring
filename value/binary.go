//  Copieright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package value

import (
	"bytes"
	"fmt"
)

type binaryValue []byte

func (this binaryValue) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("\"<binary (%d b)>\"", len(this))
	return []byte(s), nil
}

func (this binaryValue) Type() Type {
	return BINARY
}

func (this binaryValue) Actual() interface{} {
	return []byte(this)
}

func (this binaryValue) Equals(other Value) bool {
	other = other.unwrap()
	switch other := other.(type) {
	case binaryValue:
		return bytes.Equal(this, other)
	default:
		return false
	}
}

func (this binaryValue) Collate(other Value) int {
	other = other.unwrap()
	switch other := other.(type) {
	case binaryValue:
		return bytes.Compare(this, other)
	default:
		return int(BINARY - other.Type())
	}
}

func (this binaryValue) Truth() bool {
	return len(this) > 0
}

func (this binaryValue) Copy() Value {
	return this
}

func (this binaryValue) CopyForUpdate() Value {
	return this
}

func (this binaryValue) Field(field string) (Value, bool) {
	return missingField(field), false
}

func (this binaryValue) SetField(field string, val interface{}) error {
	return Unsettable(field)
}

func (this binaryValue) UnsetField(field string) error {
	return Unsettable(field)
}

func (this binaryValue) Index(index int) (Value, bool) {
	return missingIndex(index), false
}

func (this binaryValue) SetIndex(index int, val interface{}) error {
	return Unsettable(index)
}

func (this binaryValue) Slice(start, end int) (Value, bool) {
	return NULL_VALUE, false
}

func (this binaryValue) SliceTail(start int) (Value, bool) {
	return NULL_VALUE, false
}

func (this binaryValue) Descendants(buffer []interface{}) []interface{} {
	return buffer
}

func (this binaryValue) Fields() map[string]interface{} {
	return nil
}

func (this binaryValue) Successor() Value {
	return nil
}

func (this binaryValue) unwrap() Value {
	return this
}