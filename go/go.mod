module github.com/liquidata-inc/dolt/go

require (
	cloud.google.com/go v0.43.0
	github.com/BurntSushi/toml v0.3.1
	github.com/DATA-DOG/go-sqlmock v1.3.3 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/abiosoft/readline v0.0.0-20180607040430-155bce2042db
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d
	github.com/alecthomas/template v0.0.0-20160405071501-a0175ee3bccc // indirect
	github.com/alecthomas/units v0.0.0-20151022065526-2efee857e7cf // indirect
	github.com/attic-labs/kingpin v2.2.7-0.20180312050558-442efcfac769+incompatible
	github.com/aws/aws-sdk-go v1.21.2
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/codahale/blake2 v0.0.0-20150924215134-8d10d0420cbf
	github.com/dustin/go-humanize v1.0.0
	github.com/edsrzf/mmap-go v1.0.0-20181222142022-904c4ced31cd
	github.com/fatih/color v1.7.0
	github.com/flynn-archive/go-shlex v0.0.0-20150515145356-3f9db97f8568
	github.com/gizak/termui/v3 v3.1.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gocraft/dbr v0.0.0-20190708200302-a54124dfc613
	github.com/golang/protobuf v1.3.2
	github.com/golang/snappy v0.0.1
	github.com/google/go-cmp v0.3.0
	github.com/google/uuid v1.1.1
	github.com/jmoiron/sqlx v1.2.0 // indirect
	github.com/jpillora/backoff v0.0.0-20180909062703-3050d21c67d7
	github.com/juju/fslock v0.0.0-20160525022230-4d5c94c67b4b
	github.com/juju/gnuflag v0.0.0-20171113085948-2ce1bb71843d
	github.com/kch42/buzhash v0.0.0-20160816060738-9bdec3dec7c6
	github.com/kr/pretty v0.1.0 // indirect
	github.com/lib/pq v1.1.1 // indirect
	github.com/liquidata-inc/ishell v0.0.0-20190514193646-693241f1f2a0
	github.com/mattn/go-isatty v0.0.8
	github.com/mattn/go-runewidth v0.0.4
	github.com/mattn/go-sqlite3 v1.10.0 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b
	github.com/nsf/termbox-go v0.0.0-20190121233118-02980233997d
	github.com/pkg/errors v0.8.1
	github.com/pkg/profile v1.3.0
	github.com/rivo/uniseg v0.0.0-20190513083848-b9f5b9457d44
	github.com/shirou/gopsutil v2.18.12+incompatible
	github.com/sirupsen/logrus v1.4.2
	github.com/skratchdot/open-golang v0.0.0-20190402232053-79abb63cd66e
	github.com/src-d/go-mysql-server v0.4.1-0.20190624170509-8702d43af506
	github.com/stretchr/testify v1.3.0
	github.com/tealeg/xlsx v1.0.4-0.20190601071628-e2d23f3c43dc
	golang.org/x/crypto v0.0.0-20190605123033-f99c8df09eb5
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859
	golang.org/x/sys v0.0.0-20190624142023-c5567b49c5d0
	google.golang.org/api v0.7.0
	google.golang.org/grpc v1.22.0
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/square/go-jose.v2 v2.3.1
	vitess.io/vitess v3.0.0-rc.3.0.20190602171040-12bfde34629c+incompatible
)

replace github.com/src-d/go-mysql-server => github.com/liquidata-inc/go-mysql-server v0.4.1-0.20190710171053-b2883167103a

//replace github.com/src-d/go-mysql-server => ../../../../liquidata-inc/go-mysql-server

replace vitess.io/vitess => github.com/liquidata-inc/vitess v0.0.0-20190625235908-66745781a796

// For local development, clone vitess into $GOPATH/src like so: git clone git@github.com:liquidata-inc/vitess.git vitess.io/vitess
// Then use this local override:
//replace vitess.io/vitess => ../../../../../vitess.io/vitess
