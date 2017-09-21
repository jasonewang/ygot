// Copyright 2017 Google Inc.
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
package ygen

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/openconfig/goyang/pkg/yang"
)

func TestYangTypeToProtoType(t *testing.T) {
	tests := []struct {
		name        string
		in          []resolveTypeArgs
		wantWrapper *mappedType
		wantScalar  *mappedType
		wantSame    bool
		wantErr     bool
	}{{
		name: "integer types",
		in: []resolveTypeArgs{
			{yangType: &yang.YangType{Kind: yang.Yint8}},
			{yangType: &yang.YangType{Kind: yang.Yint16}},
			{yangType: &yang.YangType{Kind: yang.Yint32}},
			{yangType: &yang.YangType{Kind: yang.Yint64}},
		},
		wantWrapper: &mappedType{nativeType: "ywrapper.IntValue"},
		wantScalar:  &mappedType{nativeType: "sint64"},
	}, {
		name: "unsigned integer types",
		in: []resolveTypeArgs{
			{yangType: &yang.YangType{Kind: yang.Yuint8}},
			{yangType: &yang.YangType{Kind: yang.Yuint16}},
			{yangType: &yang.YangType{Kind: yang.Yuint32}},
			{yangType: &yang.YangType{Kind: yang.Yuint64}},
		},
		wantWrapper: &mappedType{nativeType: "ywrapper.UintValue"},
		wantScalar:  &mappedType{nativeType: "uint64"},
	}, {
		name: "bool types",
		in: []resolveTypeArgs{
			{yangType: &yang.YangType{Kind: yang.Ybool}},
			{yangType: &yang.YangType{Kind: yang.Yempty}},
		},
		wantWrapper: &mappedType{nativeType: "ywrapper.BoolValue"},
		wantScalar:  &mappedType{nativeType: "bool"},
	}, {
		name:        "string",
		in:          []resolveTypeArgs{{yangType: &yang.YangType{Kind: yang.Ystring}}},
		wantWrapper: &mappedType{nativeType: "ywrapper.StringValue"},
		wantScalar:  &mappedType{nativeType: "string"},
	}, {
		name:        "decimal64",
		in:          []resolveTypeArgs{{yangType: &yang.YangType{Kind: yang.Ydecimal64}}},
		wantWrapper: &mappedType{nativeType: "ywrapper.Decimal64Value"},
		wantSame:    true,
	}, {
		name: "unmapped types",
		in: []resolveTypeArgs{
			{yangType: &yang.YangType{Kind: yang.Ybinary}},
			{yangType: &yang.YangType{Kind: yang.Ybits}},
		},
		wantErr: true,
	}, {
		name: "union of string, uint32",
		in: []resolveTypeArgs{
			{
				yangType: &yang.YangType{
					Kind: yang.Yunion,
					Type: []*yang.YangType{
						{Kind: yang.Ystring, Name: "string"},
						{Kind: yang.Yuint32, Name: "uint32"},
					},
				},
			},
		},
		wantWrapper: &mappedType{unionTypes: map[string]int{"string": 0, "uint64": 1}},
		wantSame:    true,
	}, {
		name: "union with only strings",
		in: []resolveTypeArgs{{
			yangType: &yang.YangType{
				Kind: yang.Yunion,
				Type: []*yang.YangType{
					{Kind: yang.Ystring, Name: "string"},
					{Kind: yang.Ystring, Name: "string"},
				},
			},
		}},
		wantWrapper: &mappedType{nativeType: "ywrapper.StringValue"},
		wantSame:    true,
	}, {
		name: "derived identityref",
		in: []resolveTypeArgs{{
			yangType: &yang.YangType{
				Kind: yang.Yidentityref,
				Name: "derived-identityref",
			},
			contextEntry: &yang.Entry{
				Type: &yang.YangType{
					Name: "derived-identityref",
					IdentityBase: &yang.Identity{
						Name:   "base-identity",
						Parent: &yang.Module{Name: "base-module"},
					},
				},
				Node: &yang.Leaf{
					Parent: &yang.Module{Name: "base-module"},
				},
			},
		}},
		wantWrapper: &mappedType{
			nativeType:        "basePackage.enumPackage.BaseModuleDerivedIdentityref",
			isEnumeratedValue: true,
		},
		wantSame: true,
	}, {
		name: "enumeration without context",
		in: []resolveTypeArgs{{
			yangType: &yang.YangType{
				Kind: yang.Yenum,
				Name: "enumeration",
			},
		}},
		wantErr: true,
	}, {
		name: "enumeration",
		in: []resolveTypeArgs{{
			yangType: &yang.YangType{
				Kind: yang.Yenum,
				Name: "enumeration",
			},
			contextEntry: &yang.Entry{
				Name: "enumeration-leaf",
				Type: &yang.YangType{
					Name: "enumeration",
					Enum: &yang.EnumType{},
				},
				Parent: &yang.Entry{Name: "base-module"},
			},
		}},
		wantWrapper: &mappedType{
			nativeType:        "EnumerationLeaf",
			isEnumeratedValue: true,
		},
		wantSame: true,
	}, {
		name: "typedef enumeration",
		in: []resolveTypeArgs{{
			yangType: &yang.YangType{Kind: yang.Yenum, Name: "derived-enumeration"},
			contextEntry: &yang.Entry{
				Name: "enumeration-leaf",
				Type: &yang.YangType{
					Name: "derived-enumeration",
					Enum: &yang.EnumType{},
				},
				Node: &yang.Enum{
					Parent: &yang.Module{
						Name: "base-module",
					},
				},
			},
		}},
		wantWrapper: &mappedType{nativeType: "basePackage.enumPackage.BaseModuleDerivedEnumeration", isEnumeratedValue: true},
		wantSame:    true,
	}, {
		name: "identityref",
		in: []resolveTypeArgs{{
			yangType: &yang.YangType{Kind: yang.Yidentityref, Name: "identityref"},
			contextEntry: &yang.Entry{
				Name: "identityref",
				Type: &yang.YangType{
					Name: "identityref",
					IdentityBase: &yang.Identity{
						Name: "base-identity",
						Parent: &yang.Module{
							Name: "test-module",
						},
					},
				},
				Node: &yang.Leaf{
					Parent: &yang.Module{
						Name: "test-module",
					},
				},
			},
		}},
		wantWrapper: &mappedType{nativeType: "basePackage.enumPackage.TestModuleBaseIdentity", isEnumeratedValue: true},
		wantSame:    true,
	}}

	for _, tt := range tests {
		s := newGenState()
		for _, st := range tt.in {
			gotWrapper, err := s.yangTypeToProtoType(st, "basePackage", "enumPackage")
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: yangTypeToProtoType(%v): got unexpected error: %v", tt.name, tt.in, err)
				continue
			}

			if diff := pretty.Compare(gotWrapper, tt.wantWrapper); diff != "" {
				t.Errorf("%s: yangTypeToProtoType(%v): did not get correct type, diff(-got,+want):\n%s", tt.name, tt.in, diff)
			}

			gotScalar, err := s.yangTypeToProtoScalarType(st, "basePackage", "enumPackage")
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: yangTypeToProtoScalarType(%v, basePackage, enumPackage): got unexpected error: %v", tt.name, tt.in, err)
			}

			wantScalar := tt.wantScalar
			if tt.wantSame {
				wantScalar = tt.wantWrapper
			}
			if diff := pretty.Compare(gotScalar, wantScalar); diff != "" {
				t.Errorf("%s: yangTypeToProtoScalarType(%v): did not get correct type, diff(-got,+want):\n%s", tt.name, tt.in, diff)
			}
		}
	}
}

func TestProtoMsgName(t *testing.T) {
	tests := []struct {
		name                   string
		inEntry                *yang.Entry
		inUniqueProtoMsgNames  map[string]map[string]bool
		inUniqueDirectoryNames map[string]string
		wantCompress           string
		wantUncompress         string
	}{{
		name: "simple message name",
		inEntry: &yang.Entry{
			Name: "msg",
			Parent: &yang.Entry{
				Name: "package",
				Parent: &yang.Entry{
					Name: "module",
				},
			},
		},
		wantCompress:   "Msg",
		wantUncompress: "Msg",
	}, {
		name: "simple message name with compression",
		inEntry: &yang.Entry{
			Name: "msg",
			Parent: &yang.Entry{
				Name: "config",
				Parent: &yang.Entry{
					Name: "container",
					Parent: &yang.Entry{
						Name: "module",
					},
				},
			},
		},
		wantCompress:   "Msg",
		wantUncompress: "Msg",
	}, {
		name: "simple message name with clash when compressing",
		inEntry: &yang.Entry{
			Name: "msg",
			Parent: &yang.Entry{
				Name: "config",
				Kind: yang.DirectoryEntry,
				Dir:  map[string]*yang.Entry{},
				Parent: &yang.Entry{
					Name: "container",
					Parent: &yang.Entry{
						Name: "module",
					},
				},
			},
		},
		inUniqueProtoMsgNames: map[string]map[string]bool{
			"container": {
				"Msg": true,
			},
		},
		wantCompress:   "Msg_",
		wantUncompress: "Msg",
	}, {
		name: "cached name",
		inEntry: &yang.Entry{
			Name: "leaf",
			Parent: &yang.Entry{
				Name: "config",
				Parent: &yang.Entry{
					Name: "container",
					Parent: &yang.Entry{
						Name: "module",
					},
				},
			},
		},
		inUniqueDirectoryNames: map[string]string{"/module/container/config/leaf": "OverriddenName"},
		wantCompress:           "OverriddenName",
		wantUncompress:         "OverriddenName",
	}}

	for _, tt := range tests {
		for compress, want := range map[bool]string{true: tt.wantCompress, false: tt.wantUncompress} {
			s := newGenState()
			// Seed the proto message names with some known input.
			if tt.inUniqueProtoMsgNames != nil {
				s.uniqueProtoMsgNames = tt.inUniqueProtoMsgNames
			}

			if tt.inUniqueDirectoryNames != nil {
				s.uniqueDirectoryNames = tt.inUniqueDirectoryNames
			}

			if got := s.protoMsgName(tt.inEntry, compress); got != want {
				t.Errorf("%s: protoMsgName(%v, %v): did not get expected name, got: %v, want: %v", tt.name, tt.inEntry, compress, got, want)
			}
		}
	}
}

func TestProtoPackageName(t *testing.T) {
	tests := []struct {
		name                  string
		inEntry               *yang.Entry
		inDefinedGlobals      map[string]bool
		inUniqueProtoPackages map[string]string
		wantCompress          string
		wantUncompress        string
	}{{
		name: "simple package name",
		inEntry: &yang.Entry{
			Name: "leaf",
			Parent: &yang.Entry{
				Name: "child-container",
				Parent: &yang.Entry{
					Name: "parent-container",
					Kind: yang.DirectoryEntry,
					Dir:  map[string]*yang.Entry{},
					Parent: &yang.Entry{
						Name: "module",
					},
				},
			},
		},
		wantCompress:   "parent_container.child_container",
		wantUncompress: "module.parent_container.child_container",
	}, {
		name: "package name with choice and case",
		inEntry: &yang.Entry{
			Name: "leaf",
			Parent: &yang.Entry{
				Name: "child-container",
				Dir:  map[string]*yang.Entry{},
				Parent: &yang.Entry{
					Name: "case",
					Kind: yang.CaseEntry,
					Dir:  map[string]*yang.Entry{},
					Parent: &yang.Entry{
						Name: "choice",
						Kind: yang.ChoiceEntry,
						Dir:  map[string]*yang.Entry{},
						Parent: &yang.Entry{
							Name: "container",
							Dir:  map[string]*yang.Entry{},
							Parent: &yang.Entry{
								Name: "module",
							},
						},
					},
				},
			},
		},
		wantCompress:   "container.child_container",
		wantUncompress: "module.container.child_container",
	}, {
		name: "clashing names",
		inEntry: &yang.Entry{
			Name: "leaf",
			Parent: &yang.Entry{
				Name: "baz-bat",
				Parent: &yang.Entry{
					Name: "bar",
					Dir:  map[string]*yang.Entry{},
					Parent: &yang.Entry{
						Name: "foo",
						Dir:  map[string]*yang.Entry{},
					},
				},
			},
		},
		inDefinedGlobals: map[string]bool{
			"foo.bar.baz_bat": true, // Clash for uncompressed.
			"bar.baz_bat":     true, // Clash for compressed.
		},
		wantCompress:   "bar.baz_bat_",
		wantUncompress: "foo.bar.baz_bat_",
	}, {
		name: "previously defined parent name",
		inEntry: &yang.Entry{
			Name: "leaf",
			Parent: &yang.Entry{
				Name: "parent",
				Dir:  map[string]*yang.Entry{},
				Kind: yang.DirectoryEntry,
				Parent: &yang.Entry{
					Name: "module",
					Kind: yang.DirectoryEntry,
					Dir:  map[string]*yang.Entry{},
				},
			},
		},
		inUniqueProtoPackages: map[string]string{
			"/module/parent": "explicit.package.name",
		},
		wantCompress:   "explicit.package.name",
		wantUncompress: "explicit.package.name",
	}}

	for _, tt := range tests {
		for compress, want := range map[bool]string{true: tt.wantCompress, false: tt.wantUncompress} {
			s := newGenState()
			if tt.inDefinedGlobals != nil {
				s.definedGlobals = tt.inDefinedGlobals
			}

			if tt.inUniqueProtoPackages != nil {
				s.uniqueProtoPackages = tt.inUniqueProtoPackages
			}

			if got := s.protobufPackage(tt.inEntry, compress); got != want {
				t.Errorf("%s: protobufPackage(%v, %v): did not get expected package name, got: %v, want: %v", tt.name, tt.inEntry, compress, got, want)
			}
		}
	}
}