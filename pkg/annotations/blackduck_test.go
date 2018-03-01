/*
Copyright (C) 2018 Black Duck Software, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/

package annotations

import (
	"encoding/json"
	//	"reflect"
	"testing"
)

func createObj(name string, desc string, ts int64, ref string, com bool, label string, score int, index int) BlackDuckAnnotation {
	return BlackDuckAnnotation{
		Name:        name,
		Description: desc,
		Timestamp:   ts,
		Reference:   ref,
		Compliant:   com,
		Summary: []summaryEntry{
			{
				Label:         label,
				Score:         score,
				SeverityIndex: index,
			},
		},
	}
}

func createObjStr(name string, desc string, ts int64, ref string, com bool, label string, score int, index int) string {
	bd := createObj(name, desc, ts, ref, com, label, score, index)
	bytes, _ := json.Marshal(bd)
	return string(bytes)
}

func TestCompare(t *testing.T) {

	testcases := []struct {
		description string
		obj1        BlackDuckAnnotation
		obj2        BlackDuckAnnotation
		shouldPass  bool
	}{
		{
			description: "same objects",
			obj1:        createObj("test", "test", 123456, "test", true, "high", 1, 1),
			obj2:        createObj("test", "test", 123456, "test", true, "high", 1, 1),
			shouldPass:  true,
		},
		{
			description: "empty objects",
			obj1:        BlackDuckAnnotation{},
			obj2:        BlackDuckAnnotation{},
			shouldPass:  true,
		},
		{
			description: "different name",
			obj1:        createObj("test", "test", 123456, "test", true, "high", 1, 1),
			obj2:        createObj("diffName", "test", 123456, "test", true, "high", 1, 1),
			shouldPass:  false,
		},
		{
			description: "different description",
			obj1:        createObj("test", "test", 123456, "test", true, "high", 1, 1),
			obj2:        createObj("test", "diffDescription", 123456, "test", true, "high", 1, 1),
			shouldPass:  false,
		},
		{
			description: "different timestamp",
			obj1:        createObj("test", "test", 123456, "test", true, "high", 1, 1),
			obj2:        createObj("test", "test", 456789, "test", true, "high", 1, 1),
			shouldPass:  true,
		},
		{
			description: "different reference",
			obj1:        createObj("test", "test", 123456, "test", true, "high", 1, 1),
			obj2:        createObj("test", "test", 123456, "diffReference", true, "high", 1, 1),
			shouldPass:  false,
		},
		{
			description: "different compliance",
			obj1:        createObj("test", "test", 123456, "test", true, "high", 1, 1),
			obj2:        createObj("test", "test", 123456, "test", false, "high", 1, 1),
			shouldPass:  false,
		},
		{
			description: "different summary label",
			obj1:        createObj("test", "test", 123456, "test", true, "high", 1, 1),
			obj2:        createObj("test", "test", 123456, "test", true, "otherLabel", 1, 1),
			shouldPass:  false,
		},
		{
			description: "different summary score",
			obj1:        createObj("test", "test", 123456, "test", true, "high", 1, 1),
			obj2:        createObj("test", "test", 123456, "test", true, "high", 2, 1),
			shouldPass:  false,
		},
		{
			description: "different summary severity index",
			obj1:        createObj("test", "test", 123456, "test", true, "high", 1, 1),
			obj2:        createObj("test", "test", 123456, "test", true, "high", 1, 2),
			shouldPass:  false,
		},
	}

	for _, tc := range testcases {
		result := tc.obj1.Compare(&tc.obj2)
		if result != tc.shouldPass {
			t.Errorf("[%s] expected %t got %t: obj1 %v, obj2 %v", tc.description, tc.shouldPass, result, tc.obj1, tc.obj2)
		}
	}
}

func TestNewBlackDuckAnnotationFromJSON(t *testing.T) {
	testcases := []struct {
		description string
		objStr      string
		shouldPass  bool
	}{
		{
			description: "full string",
			objStr:      createObjStr("name", "desc", 123, "url", true, "label", 1, 1),
			shouldPass:  true,
		},
		{
			description: "empty string",
			objStr:      "",
			shouldPass:  false,
		},
		{
			description: "non-json string",
			objStr:      "this is invalid",
			shouldPass:  false,
		},
	}

	for _, tc := range testcases {
		t.Logf("objStr: %s", tc.objStr)
		obj, err := NewBlackDuckAnnotationFromJSON(tc.objStr)
		t.Logf("obj: %v", obj)
		if err != nil && tc.shouldPass {
			t.Fatalf("[%s] error: %v, obj %v", tc.description, err, tc.objStr)
		}
	}
}

func TestCompareBlackDuckAnnotationJSON(t *testing.T) {
	testcases := []struct {
		description string
		objStr1     string
		objStr2     string
		shouldPass  bool
	}{
		{
			description: "same objects",
			objStr1:     createObjStr("test", "test", 123456, "test", true, "high", 1, 1),
			objStr2:     createObjStr("test", "test", 123456, "test", true, "high", 1, 1),
			shouldPass:  true,
		},
		{
			description: "empty strings",
			objStr1:     "",
			objStr2:     "",
			shouldPass:  false,
		},
		{
			description: "different name",
			objStr1:     createObjStr("test", "test", 123456, "test", true, "high", 1, 1),
			objStr2:     createObjStr("diffName", "test", 123456, "test", true, "high", 1, 1),
			shouldPass:  false,
		},
		{
			description: "different description",
			objStr1:     createObjStr("test", "test", 123456, "test", true, "high", 1, 1),
			objStr2:     createObjStr("test", "diffDescription", 123456, "test", true, "high", 1, 1),
			shouldPass:  false,
		},
		{
			description: "different timestamp",
			objStr1:     createObjStr("test", "test", 123456, "test", true, "high", 1, 1),
			objStr2:     createObjStr("test", "test", 456789, "test", true, "high", 1, 1),
			shouldPass:  true,
		},
		{
			description: "different reference",
			objStr1:     createObjStr("test", "test", 123456, "test", true, "high", 1, 1),
			objStr2:     createObjStr("test", "test", 123456, "diffReference", true, "high", 1, 1),
			shouldPass:  false,
		},
		{
			description: "different compliance",
			objStr1:     createObjStr("test", "test", 123456, "test", true, "high", 1, 1),
			objStr2:     createObjStr("test", "test", 123456, "test", false, "high", 1, 1),
			shouldPass:  false,
		},
		{
			description: "different summary label",
			objStr1:     createObjStr("test", "test", 123456, "test", true, "high", 1, 1),
			objStr2:     createObjStr("test", "test", 123456, "test", true, "otherLabel", 1, 1),
			shouldPass:  false,
		},
		{
			description: "different summary score",
			objStr1:     createObjStr("test", "test", 123456, "test", true, "high", 1, 1),
			objStr2:     createObjStr("test", "test", 123456, "test", true, "high", 2, 1),
			shouldPass:  false,
		},
		{
			description: "different summary severity index",
			objStr1:     createObjStr("test", "test", 123456, "test", true, "high", 1, 1),
			objStr2:     createObjStr("test", "test", 123456, "test", true, "high", 1, 2),
			shouldPass:  false,
		},
		{
			description: "invalid json",
			objStr1:     "invalid string",
			objStr2:     createObjStr("test", "test", 123456, "test", true, "high", 1, 2),
			shouldPass:  false,
		},
	}

	for _, tc := range testcases {
		result := CompareBlackDuckAnnotationJSON(tc.objStr1, tc.objStr2)
		if result != tc.shouldPass {
			t.Errorf("[%s] expected %t got %t: objStr1 %v, objStr2 %v", tc.description, tc.shouldPass, result, tc.objStr1, tc.objStr2)
		}
	}

}
