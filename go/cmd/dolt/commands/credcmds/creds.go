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

package credcmds

import (
	"github.com/liquidata-inc/dolt/go/cmd/dolt/cli"
)

var Commands = cli.GenSubCommandHandler([]*cli.Command{
	{Name: "new", Desc: "", Func: New, ReqRepo: false},
	{Name: "rm", Desc: "", Func: Rm, ReqRepo: false},
	{Name: "ls", Desc: "", Func: Ls, ReqRepo: false},
})
