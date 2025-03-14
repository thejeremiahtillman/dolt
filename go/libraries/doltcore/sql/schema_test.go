// Copyright 2019 Liquidata, Inc.
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

package sql

import (
	"github.com/liquidata-inc/dolt/go/libraries/doltcore/schema"
	"github.com/liquidata-inc/dolt/go/libraries/doltcore/sql/sqltestutil"
	"github.com/liquidata-inc/dolt/go/store/types"

	"testing"

	"github.com/stretchr/testify/assert"
)

const expectedSQL = "CREATE TABLE `table_name` (\n" +
	"  `id` int not null comment 'tag:0',\n" +
	"  `first` varchar not null comment 'tag:1',\n" +
	"  `last` varchar not null comment 'tag:2',\n" +
	"  `is_married` bool comment 'tag:3',\n" +
	"  `age` int comment 'tag:4',\n" +
	"  `rating` float comment 'tag:6',\n" +
	"  `uuid` uuid comment 'tag:7',\n" +
	"  `num_episodes` int unsigned comment 'tag:8',\n" +
	"  primary key (`id`)\n" +
	");"

func TestSchemaAsCreateStmt(t *testing.T) {
	tSchema := sqltestutil.PeopleTestSchema
	str := SchemaAsCreateStmt("table_name", tSchema)

	assert.Equal(t, expectedSQL, str)
}

func TestFmtCol(t *testing.T) {
	tests := []struct {
		Col       schema.Column
		Indent    int
		NameWidth int
		TypeWidth int
		Expected  string
	}{
		{
			schema.NewColumn("first", 0, types.StringKind, true),
			0,
			0,
			0,
			"`first` varchar comment 'tag:0'",
		},
		{
			schema.NewColumn("last", 123, types.IntKind, true),
			2,
			0,
			0,
			"  `last` int comment 'tag:123'",
		},
		{
			schema.NewColumn("title", 2, types.UintKind, true),
			0,
			10,
			0,
			"   `title` int unsigned comment 'tag:2'",
		},
		{
			schema.NewColumn("aoeui", 52, types.UintKind, true),
			0,
			10,
			15,
			"   `aoeui`    int unsigned comment 'tag:52'",
		},
	}

	for _, test := range tests {
		t.Run(test.Expected, func(t *testing.T) {
			actual := FmtCol(test.Indent, test.NameWidth, test.TypeWidth, test.Col)
			assert.Equal(t, test.Expected, actual)
		})
	}
}
