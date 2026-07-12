// Copyright (c) 2026 Ant Group Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"reflect"
	"testing"
)

func TestUnescapedMarshal(t *testing.T) {
	type args struct {
		in interface{}
	}

	type Tmp struct {
		Name string `json:"name"`
	}

	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "html escape >",
			args: args{
				in: Tmp{Name: "xxx > yyy"},
			},
			want:    []byte(`{"name":"xxx > yyy"}`),
			wantErr: false,
		},
		{
			name: "complex content",
			args: args{
				in: Tmp{Name: "xxx > yyy\nzzz\n@#$%^&*()_+{}|:\"<>?"},
			},
			want:    []byte(`{"name":"xxx > yyy\nzzz\n@#$%^&*()_+{}|:\"<>?"}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnescapedMarshal(tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnescapedMarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnescapedMarshal() got = %v, want %v", got, tt.want)
			}
		})
	}
}
