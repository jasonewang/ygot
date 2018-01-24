// Copyright 2018 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ygot

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/kylelemons/godebug/pretty"
	gnmipb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/util"
)

func TestSchemaPathToGNMIPath(t *testing.T) {
	tests := []struct {
		desc string
		in   []string
		want *gnmipb.Path
	}{{
		desc: "single element",
		in:   []string{"one"},
		want: &gnmipb.Path{
			Elem: []*gnmipb.PathElem{{
				Name: "one",
			}},
		},
	}, {
		desc: "multiple elements",
		in:   []string{"one", "two", "three"},
		want: &gnmipb.Path{
			Elem: []*gnmipb.PathElem{{
				Name: "one",
			}, {
				Name: "two",
			}, {
				Name: "three",
			}},
		},
	}}

	for _, tt := range tests {
		if got := schemaPathTogNMIPath(tt.in); !proto.Equal(got, tt.want) {
			t.Errorf("%s: schemaPathTogNMIPath(%v): did not get expected path, got: %v, want: %v", tt.desc, tt.in, pretty.Sprint(got), pretty.Sprint(tt.want))
		}
	}
}

func TestJoingNMIPaths(t *testing.T) {
	tests := []struct {
		desc     string
		inParent *gnmipb.Path
		inChild  *gnmipb.Path
		want     *gnmipb.Path
	}{{
		desc: "simple parent and child",
		inParent: &gnmipb.Path{
			Elem: []*gnmipb.PathElem{{
				Name: "one",
			}},
		},
		inChild: &gnmipb.Path{
			Elem: []*gnmipb.PathElem{{
				Name: "two",
			}},
		},
		want: &gnmipb.Path{
			Elem: []*gnmipb.PathElem{{
				Name: "one",
			}, {
				Name: "two",
			}},
		},
	}, {
		desc: "simple parent with list in child",
		inParent: &gnmipb.Path{
			Elem: []*gnmipb.PathElem{{
				Name: "one",
			}},
		},
		inChild: &gnmipb.Path{
			Elem: []*gnmipb.PathElem{{
				Name: "two",
			}, {
				Name: "three",
				Key:  map[string]string{"four": "five"},
			}},
		},
		want: &gnmipb.Path{
			Elem: []*gnmipb.PathElem{{
				Name: "one",
			}, {
				Name: "two",
			}, {
				Name: "three",
				Key:  map[string]string{"four": "five"},
			}},
		},
	}, {
		desc: "list in parent, simple child",
		inParent: &gnmipb.Path{
			Elem: []*gnmipb.PathElem{{
				Name: "one",
				Key:  map[string]string{"two": "three"},
			}},
		},
		inChild: &gnmipb.Path{
			Elem: []*gnmipb.PathElem{{
				Name: "four",
			}},
		},
		want: &gnmipb.Path{
			Elem: []*gnmipb.PathElem{{
				Name: "one",
				Key:  map[string]string{"two": "three"},
			}, {
				Name: "four",
			}},
		},
	}}

	for _, tt := range tests {
		if got := joingNMIPaths(tt.inParent, tt.inChild); !proto.Equal(got, tt.want) {
			diff := pretty.Compare(got, tt.want)
			t.Errorf("%s: joingNMIPaths(%v, %v): did not get expected path, diff(-got,+want):\n%s", tt.desc, tt.inParent, tt.inChild, diff)
		}
	}
}

type basicStruct struct {
	StringValue *string                     `path:"string-value"`
	StructValue *basicStructTwo             `path:"struct-value"`
	MapValue    map[string]*basicListMember `path:"map-list"`
}

func (*basicStruct) IsYANGGoStruct() {}

type basicStructTwo struct {
	StringValue *string           `path:"second-string-value"`
	StructValue *basicStructThree `path:"struct-three-value"`
}

type basicListMember struct {
	ListKey *string `path:"list-key"`
}

func (*basicListMember) IsYANGGoStruct() {}
func (b *basicListMember) ΛListKeyMap() (map[string]interface{}, error) {
	return map[string]interface{}{
		"list-key": *b.ListKey,
	}, nil
}

type errorListMember struct {
	StringValue *string `path:"error-list-key"`
}

func (*errorListMember) IsYANGGoStruct() {}
func (b *errorListMember) ΛListKeyMap() (map[string]interface{}, error) {
	return nil, fmt.Errorf("invalid key map")
}

type badListKeyType struct {
	Value *complex128 `path:"error-list-key"`
}

func (*badListKeyType) IsYANGGoStruct() {}
func (b *badListKeyType) ΛListKeyMap() (map[string]interface{}, error) {
	return map[string]interface{}{
		"error-list-key": *b.Value,
	}, nil
}

type basicStructThree struct {
	StringValue *string `path:"third-string-value|config/third-string-value"`
}

func TestNodeValuePath(t *testing.T) {
	cmplx := complex(float64(1), float64(2))
	tests := []struct {
		desc          string
		inNI          *util.NodeInfo
		inSchemaPaths [][]string
		wantPathSpec  *pathSpec
		wantErr       string
	}{{
		desc: "root level element",
		inNI: &util.NodeInfo{
			Parent: nil,
		},
		inSchemaPaths: [][]string{{"one", "two"}, {"three", "four"}},
		wantPathSpec: &pathSpec{
			gNMIPaths: []*gnmipb.Path{{
				Elem: []*gnmipb.PathElem{{Name: "one"}, {Name: "two"}},
			}, {
				Elem: []*gnmipb.PathElem{{Name: "three"}, {Name: "four"}},
			}},
		},
	}, {
		desc: "nodeinfo missing parent annotation",
		inNI: &util.NodeInfo{
			Parent: &util.NodeInfo{
				Annotation: []interface{}{},
			},
		},
		wantErr: "could not find path specification annotation",
	}, {
		desc: "nodeinfo for list member",
		inNI: &util.NodeInfo{
			Parent: &util.NodeInfo{
				Annotation: []interface{}{&pathSpec{
					gNMIPaths: []*gnmipb.Path{{
						Elem: []*gnmipb.PathElem{{
							Name: "a-list",
						}},
					}},
				}},
			},
			FieldValue: reflect.ValueOf(&basicListMember{ListKey: String("key-value")}),
		},
		wantPathSpec: &pathSpec{
			gNMIPaths: []*gnmipb.Path{{
				Elem: []*gnmipb.PathElem{{Name: "a-list", Key: map[string]string{"list-key": "key-value"}}},
			}},
		},
	}, {
		desc: "nodeinfo for invalid list member",
		inNI: &util.NodeInfo{
			Parent: &util.NodeInfo{
				Annotation: []interface{}{&pathSpec{
					gNMIPaths: []*gnmipb.Path{{
						Elem: []*gnmipb.PathElem{{
							Name: "a-list",
						}},
					}},
				}},
			},
			FieldValue: reflect.ValueOf(&errorListMember{StringValue: String("foo")}),
		},
		wantErr: "invalid key map",
	}, {
		desc: "nodeinfo for list member with unstringable key",
		inNI: &util.NodeInfo{
			Parent: &util.NodeInfo{
				Annotation: []interface{}{&pathSpec{
					gNMIPaths: []*gnmipb.Path{{
						Elem: []*gnmipb.PathElem{{
							Name: "a-list",
						}},
					}},
				}},
			},
			FieldValue: reflect.ValueOf(&badListKeyType{Value: &cmplx}),
		},
		wantErr: "cannot convert keys to map[string]string: cannot convert type complex128 to a string for use in a key: (1+2i)",
	}, {
		desc: "nodeinfo for list member with no parent",
		inNI: &util.NodeInfo{
			Parent: &util.NodeInfo{
				Annotation: []interface{}{&pathSpec{}},
			},
			FieldValue: reflect.ValueOf(&basicListMember{ListKey: String("key-value")}),
		},
		wantErr: "invalid list member with no parent",
	}}

	for _, tt := range tests {
		got, err := nodeValuePath(tt.inNI, tt.inSchemaPaths)
		if err != nil && err.Error() != tt.wantErr {
			t.Errorf("%s: nodeValuePath(%v, %v): did not get expected error, got: %v, want: %v", tt.desc, tt.inNI, tt.inSchemaPaths, err, tt.wantErr)
		}
		if !reflect.DeepEqual(got, tt.wantPathSpec) {
			diff := pretty.Compare(got, tt.wantPathSpec)
			t.Errorf("%s: nodeValuePath(%v, %v): did not get expected paths, diff(-got,+want): %s", tt.desc, tt.inNI, tt.inSchemaPaths, diff)
		}
	}
}

func TestFindSetLeaves(t *testing.T) {
	tests := []struct {
		desc     string
		inStruct GoStruct
		want     map[*pathSpec]interface{}
		wantErr  string
	}{{
		desc: "test",
		inStruct: &basicStruct{
			StringValue: String("value-one"),
			StructValue: &basicStructTwo{
				StringValue: String("value-two"),
				StructValue: &basicStructThree{
					StringValue: String("value-three"),
				},
			},
		},
	}, {
		desc: "test two",
		inStruct: &basicStruct{
			MapValue: map[string]*basicListMember{
				"one": {ListKey: String("one")},
				"two": {ListKey: String("two")},
			},
		},
	}}

	for _, tt := range tests {
		_, err := findSetLeaves(tt.inStruct)
		if err != nil && (err.Error() != tt.wantErr) {
			t.Errorf("%s: findSetLeaves(%v): did not get expected error: %v", tt.desc, tt.inStruct, err)
			continue
		}
	}
}
