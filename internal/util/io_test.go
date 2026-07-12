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
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMockIoUtil_WriteFile(t *testing.T) {
	type fields struct {
		SuccessMap map[string]bool
	}
	type args struct {
		filename string
		data     []byte
		perm     os.FileMode
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "test",
			fields: fields{
				SuccessMap: map[string]bool{
					"test": true,
				},
			},
			args: args{
				filename: "test",
				data:     []byte("test"),
				perm:     0644,
			},
			wantErr: assert.NoError,
		},
		{
			name: "test for empty mock map",
			fields: fields{
				SuccessMap: nil,
			},
			args: args{
				filename: "test",
				data:     []byte("test"),
				perm:     0644,
			},
			wantErr: assert.Error,
		},
		{
			name: "test",
			fields: fields{
				SuccessMap: map[string]bool{
					"test": true,
				},
			},
			args: args{
				filename: "test",
				data:     []byte("no"),
				perm:     0644,
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MockIoUtil{
				SuccessMap: tt.fields.SuccessMap,
			}
			tt.wantErr(t, m.WriteFile(tt.args.filename, tt.args.data, tt.args.perm), fmt.Sprintf("WriteFile(%v, %v, %v)", tt.args.filename, tt.args.data, tt.args.perm))
		})
	}
}

func TestOs(t *testing.T) {
	m := &MockIoUtil{SuccessMap: map[string]bool{"success": true}}
	f, err := m.Stat("test")
	assert.NoError(t, err)
	assert.Nil(t, f)

	assert.Nil(t, m.MkdirAll("@test", 0755))
	assert.Nil(t, m.WriteFile("@axx", []byte("success"), 0644))
	assert.NotNil(t, m.WriteFile("@axx", []byte("xx"), 0644))
}
