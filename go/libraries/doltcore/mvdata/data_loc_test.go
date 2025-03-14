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

package mvdata

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/liquidata-inc/dolt/go/libraries/doltcore/doltdb"
	"github.com/liquidata-inc/dolt/go/libraries/doltcore/row"
	"github.com/liquidata-inc/dolt/go/libraries/doltcore/schema"
	"github.com/liquidata-inc/dolt/go/libraries/doltcore/schema/encoding"
	"github.com/liquidata-inc/dolt/go/libraries/doltcore/table"
	"github.com/liquidata-inc/dolt/go/libraries/doltcore/table/typed/json"
	"github.com/liquidata-inc/dolt/go/libraries/doltcore/table/typed/noms"
	"github.com/liquidata-inc/dolt/go/libraries/doltcore/table/untyped/csv"
	"github.com/liquidata-inc/dolt/go/libraries/utils/filesys"
	"github.com/liquidata-inc/dolt/go/store/types"
)

var testSchema = `
	{
		"columns": [
		  {
			"name": "a",
			"kind": "string",
			"tag": 0,
			"is_part_of_pk": true,
			"col_constraints":[
			  {
				"constraint_type": "not_null"
			  }
			]
		  },
		  {
			"name": "b",
			"kind": "string",
			"tag": 1,
			"is_part_of_pk": false,
			"col_constraints": [
		  ]}
		]
	}`

var rowMap = []map[string]interface{}{
	{"a": []string{"a", "b", "c"}},
	{"b": []string{"1", "2", "3"}},
}

func createRootAndFS() (*doltdb.DoltDB, *doltdb.RootValue, filesys.Filesys) {

	testHomeDir := "/user/bheni"
	workingDir := "/user/bheni/datasets/states"
	initialDirs := []string{testHomeDir, workingDir}
	fs := filesys.NewInMemFS(initialDirs, nil, workingDir)
	fs.WriteFile("schema.json", []byte(testSchema))
	ddb, _ := doltdb.LoadDoltDB(context.Background(), types.Format_7_18, doltdb.InMemDoltDB)
	ddb.WriteEmptyRepo(context.Background(), "billy bob", "bigbillieb@fake.horse")

	cs, _ := doltdb.NewCommitSpec("HEAD", "master")
	commit, _ := ddb.Resolve(context.Background(), cs)
	root, err := commit.GetRootValue()

	if err != nil {
		panic(err)
	}

	return ddb, root, fs
}

func TestBasics(t *testing.T) {
	tests := []struct {
		dl                   *DataLocation
		expectedFmt          DataFormat
		expectedPath         string
		expectedIsFileType   bool
		expectedReqPK        bool
		expectedMustWrSorted bool
	}{
		{NewDataLocation("table-name", ""), DoltDB, "table-name", false, true, false},
		{NewDataLocation("file.csv", ""), CsvFile, "file.csv", true, false, false},
		{NewDataLocation("file.psv", ""), PsvFile, "file.psv", true, false, false},
		{NewDataLocation("file.json", ""), JsonFile, "file.json", true, false, false},
		//{NewDataLocation("file.nbf", ""), NbfFile, "file.nbf", true, true, true},
	}

	for _, test := range tests {
		if test.expectedFmt != test.dl.Format {
			t.Error(test.dl, "Unexpected format")
		}

		if test.expectedPath != test.dl.Path {
			t.Error("Unexpected path")
		}

		if test.expectedIsFileType != test.dl.IsFileType() {
			t.Error("Unexpected IsFileType result")
		}

		if test.expectedReqPK != test.dl.RequiresPK() {
			t.Error("Unexpected IsFileType result")
		}

		if test.expectedMustWrSorted != test.dl.MustWriteSorted() {
			t.Error("Unexpected IsFileType result")
		}
	}
}

var fakeFields, _ = schema.NewColCollection(
	schema.NewColumn("a", 0, types.StringKind, true, schema.NotNullConstraint{}),
	schema.NewColumn("b", 1, types.StringKind, false),
)

func mustRow(r row.Row, err error) row.Row {
	if err != nil {
		panic(err)
	}

	return r
}

var fakeSchema schema.Schema
var imt *table.InMemTable
var imtRows []row.Row

func init() {
	fakeSchema = schema.SchemaFromCols(fakeFields)

	imtRows = []row.Row{
		mustRow(row.New(types.Format_7_18, fakeSchema, row.TaggedValues{0: types.String("a"), 1: types.String("1")})),
		mustRow(row.New(types.Format_7_18, fakeSchema, row.TaggedValues{0: types.String("b"), 1: types.String("2")})),
		mustRow(row.New(types.Format_7_18, fakeSchema, row.TaggedValues{0: types.String("c"), 1: types.String("3")})),
	}

	imt = table.NewInMemTableWithData(fakeSchema, imtRows)
}

func TestExists(t *testing.T) {
	testLocations := []*DataLocation{
		NewDataLocation("table-name", ""),
		NewDataLocation("file.csv", ""),
		NewDataLocation("file.psv", ""),
		NewDataLocation("file.json", ""),
		//NewDataLocation("file.nbf", ""),
	}

	ddb, root, fs := createRootAndFS()

	for _, loc := range testLocations {
		if exists, err := loc.Exists(context.Background(), root, fs); err != nil {
			t.Error(err)
		} else if exists {
			t.Error("Shouldn't exist before creation")
		}

		if loc.Format == DoltDB {
			schVal, _ := encoding.MarshalAsNomsValue(context.Background(), ddb.ValueReadWriter(), fakeSchema)
			m, err := types.NewMap(context.Background(), ddb.ValueReadWriter())
			assert.NoError(t, err)
			tbl, err := doltdb.NewTable(context.Background(), ddb.ValueReadWriter(), schVal, m)
			assert.NoError(t, err)
			root, err = root.PutTable(context.Background(), ddb, loc.Path, tbl)
			assert.NoError(t, err)
		} else {
			err := fs.WriteFile(loc.Path, []byte("test"))
			assert.NoError(t, err)
		}

		if exists, err := loc.Exists(context.Background(), root, fs); err != nil {
			t.Error(err)
		} else if !exists {
			t.Error("Should already exist after creation")
		}
	}
}

func TestCreateRdWr(t *testing.T) {
	tests := []struct {
		dl          *DataLocation
		expectedRdT reflect.Type
		expectedWrT reflect.Type
	}{
		{NewDataLocation("table-name", ""), reflect.TypeOf((*noms.NomsMapReader)(nil)).Elem(), reflect.TypeOf((*noms.NomsMapCreator)(nil)).Elem()},
		{NewDataLocation("file.csv", ""), reflect.TypeOf((*csv.CSVReader)(nil)).Elem(), reflect.TypeOf((*csv.CSVWriter)(nil)).Elem()},
		{NewDataLocation("file.psv", ""), reflect.TypeOf((*csv.CSVReader)(nil)).Elem(), reflect.TypeOf((*csv.CSVWriter)(nil)).Elem()},
		// TODO (oo): uncomment and fix this for json path test
		{NewDataLocation("file.json", ""), reflect.TypeOf((*json.JSONReader)(nil)).Elem(), reflect.TypeOf((*json.JSONWriter)(nil)).Elem()},
		//{NewDataLocation("file.nbf", ""), reflect.TypeOf((*nbf.NBFReader)(nil)).Elem(), reflect.TypeOf((*nbf.NBFWriter)(nil)).Elem()},
	}

	ddb, root, fs := createRootAndFS()

	mvOpts := &MoveOptions{
		Operation:   OverwriteOp,
		ContOnErr:   false,
		SchFile:     schemaFile,
		MappingFile: mappingFile,
	}

	for _, test := range tests {
		loc := test.dl

		wr, err := loc.CreateOverwritingDataWriter(context.Background(), mvOpts, root, fs, true, fakeSchema, nil)

		if err != nil {
			t.Fatal("Unexpected error creating writer.", err)
		}

		actualWrT := reflect.TypeOf(wr).Elem()
		if actualWrT != test.expectedWrT {
			t.Fatal("Unexpected writer type. Expected:", test.expectedWrT.Name(), "actual:", actualWrT.Name())
		}

		inMemRd := table.NewInMemTableReader(imt)
		_, numBad, pipeErr := table.PipeRows(context.Background(), inMemRd, wr, false)
		wr.Close(context.Background())

		if numBad != 0 || pipeErr != nil {
			t.Fatal("Failed to write data. bad:", numBad, err)
		}

		if nomsWr, ok := wr.(noms.NomsMapWriteCloser); ok {
			vrw := ddb.ValueReadWriter()
			schVal, err := encoding.MarshalAsNomsValue(context.Background(), vrw, nomsWr.GetSchema())

			if err != nil {
				t.Fatal("Unable ta update table")
			}

			tbl, err := doltdb.NewTable(context.Background(), vrw, schVal, *nomsWr.GetMap())
			assert.NoError(t, err)
			root, err = root.PutTable(context.Background(), ddb, test.dl.Path, tbl)
			assert.NoError(t, err)
		}

		// TODO (oo): fix this for json path test
		rd, _, err := loc.CreateReader(context.Background(), root, fs, "schema.json", "")

		if err != nil {
			t.Fatal("Unexpected error creating writer", err)
		}

		actualRdT := reflect.TypeOf(rd).Elem()
		if actualRdT != test.expectedRdT {
			t.Error("Unexpected reader type. Expected:", test.expectedRdT.Name(), "actual:", actualRdT.Name())
		}

		rd.Close(context.Background())
	}
}
