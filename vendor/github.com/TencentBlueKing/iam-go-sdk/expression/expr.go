/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS
 * Community Edition) available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package expression

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/TencentBlueKing/iam-go-sdk/expression/eval"
	"github.com/TencentBlueKing/iam-go-sdk/expression/operator"
)

const (
	// KeywordBKIAMPath the field name of BKIAMPath
	KeywordBKIAMPath            = "_bk_iam_path_"
	KeywordBKIAMPathFieldSuffix = "._bk_iam_path_"
)

// ExprCell is the expression cell
type ExprCell struct {
	OP      operator.OP `json:"op"`
	Content []ExprCell  `json:"content"`
	Field   string      `json:"field"`
	Value   interface{} `json:"value"`
}

// Eval will evaluate the expression with ObjectSet, return true or false
func (e *ExprCell) Eval(data ObjectSetInterface) bool {

	switch e.OP {

	case operator.AND:
		for _, c := range e.Content {
			if !c.Eval(data) {
				return false
			}
		}
		return true
	case operator.OR:
		for _, c := range e.Content {
			if c.Eval(data) {
				return true
			}
		}
		return false
	default:
		return evalBinaryOperator(e.OP, e.Field, e.Value, data)
	}

}

// String return the text of expression cell
func (e *ExprCell) String() string {
	switch e.OP {
	case operator.AND, operator.OR:
		separator := fmt.Sprintf(" %s ", e.OP)

		subExprs := make([]string, 0, len(e.Content))
		for _, c := range e.Content {
			subExprs = append(subExprs, c.String())
		}
		return fmt.Sprintf("(%s)", strings.Join(subExprs, separator))
	default:
		return fmt.Sprintf("(%v %s %v)", e.Field, e.OP, e.Value)
	}
}

// Render return the rendered text of expression with ObjectSet
func (e *ExprCell) Render(data ObjectSetInterface) string {
	switch e.OP {
	case operator.AND, operator.OR:
		separator := fmt.Sprintf(" %s ", e.OP)

		subExprs := make([]string, 0, len(e.Content))
		for _, c := range e.Content {
			subExprs = append(subExprs, c.Render(data))
		}
		return fmt.Sprintf("(%s)", strings.Join(subExprs, separator))
	default:
		attrValue := data.GetAttribute(e.Field)
		return fmt.Sprintf("(%v %s %v)", attrValue, e.OP, e.Value)
	}
}

func evalBinaryOperator(op operator.OP, field string, policyValue interface{}, data ObjectSetInterface) bool {
	objectValue := data.GetAttribute(field)

	// support _bk_iam_path_, starts with from `/a,1/b,*/` to `/a,1/b,`
	if op == operator.StartsWith && strings.HasSuffix(field, KeywordBKIAMPathFieldSuffix) {
		v, ok := policyValue.(string)
		if ok {
			if strings.HasSuffix(v, ",*/") {
				policyValue = strings.TrimSuffix(v, "*/")
			}
		}
	}

	switch op {
	case operator.Any:
		return true
	case operator.Eq, operator.Lt, operator.Lte, operator.Gt, operator.Gte:
		return evalPositive(op, objectValue, policyValue)
	case operator.StartsWith, operator.EndsWith, operator.In:
		return evalPositive(op, objectValue, policyValue)
	case operator.NotEq, operator.NotStartsWith, operator.NotEndsWith, operator.NotIn:
		return evalNegative(op, objectValue, policyValue)
	case operator.Contains:
		// NOTE: objectValue is an array, policyValue is single value
		return eval.Contains(objectValue, policyValue)
	case operator.NotContains:
		// NOTE: objectValue is an array, policyValue is single value
		return eval.NotContains(objectValue, policyValue)
	default:
		return false
	}
}
func isValueTypeArray(v interface{}) bool {
	if v == nil {
		return false
	}
	kind := reflect.TypeOf(v).Kind()
	return kind == reflect.Array || kind == reflect.Slice
}

// EvalFunc is the func define of eval
type EvalFunc func(e1, e2 interface{}) bool

// evalPositive
//- 1   hit: return True
//- all miss: return False
func evalPositive(op operator.OP, objectValue, policyValue interface{}) bool {
	var evalFunc EvalFunc

	switch op {
	case operator.Eq:
		evalFunc = eval.Equal
	case operator.Lt:
		evalFunc = eval.Less
	case operator.Lte:
		evalFunc = eval.LessOrEqual
	case operator.Gt:
		evalFunc = eval.Greater
	case operator.Gte:
		evalFunc = eval.GreaterOrEqual
	case operator.StartsWith:
		evalFunc = eval.StartsWith
	case operator.EndsWith:
		evalFunc = eval.EndsWith
	case operator.In:
		evalFunc = eval.In
	}

	// NOTE: here, the policyValue should not be array! It's single value (the In op policyValue is an array)
	//fmt.Println("objectValue isValueTypeArray", objectValue, isValueTypeArray(objectValue))
	if isValueTypeArray(objectValue) {
		//fmt.Println("objectValue is an array", objectValue)
		listValue := reflect.ValueOf(objectValue)
		for i := 0; i < listValue.Len(); i++ {
			if evalFunc(listValue.Index(i).Interface(), policyValue) {
				return true
			}
		}
		return false
	}

	return evalFunc(objectValue, policyValue)
}

// evalNegative:
// - 1   miss: return False
// - all hit: return True
func evalNegative(op operator.OP, objectValue, policyValue interface{}) bool {
	var evalFunc EvalFunc

	switch op {
	case operator.NotEq:
		evalFunc = eval.NotEqual
	case operator.NotStartsWith:
		evalFunc = eval.NotStartsWith
	case operator.NotEndsWith:
		evalFunc = eval.NotEndsWith
	case operator.NotIn:
		evalFunc = eval.NotIn
	}

	// NOTE: here, the policyValue should not be array! It's single value (the In op policyValue is an array)
	if isValueTypeArray(objectValue) {
		listValue := reflect.ValueOf(objectValue)
		for i := 0; i < listValue.Len(); i++ {
			if !evalFunc(listValue.Index(i).Interface(), policyValue) {
				return false
			}
		}
		return true
	}

	return evalFunc(objectValue, policyValue)
}
