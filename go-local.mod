module github.com/whatap/go-api-example

go 1.14

require (
	github.com/go-sql-driver/mysql v1.6.0
	github.com/whatap/go-api/common/io v0.0.0-20210915084428-9f2466c3bd79 // indirect
	github.com/whatap/go-api/common/lang/pack/udp v0.0.0-20210915084428-9f2466c3bd79 // indirect
	github.com/whatap/go-api/common/net v0.0.0-20210915084428-9f2466c3bd79 // indirect
	github.com/whatap/go-api/common/util/dateutil v0.0.0-20210915084428-9f2466c3bd79 // indirect
	github.com/whatap/go-api/common/util/hash v0.0.0-20210915084428-9f2466c3bd79 // indirect
	github.com/whatap/go-api/common/util/hexa32 v0.0.0-20210915084428-9f2466c3bd79 // indirect
	github.com/whatap/go-api/common/util/keygen v0.0.0-20210915084428-9f2466c3bd79 // indirect
	github.com/whatap/go-api/common/util/paramtext v0.0.0-20210915084428-9f2466c3bd79 // indirect
	github.com/whatap/go-api/common/util/stringutil v0.0.0-20210915084428-9f2466c3bd79 // indirect
	github.com/whatap/go-api/common/util/urlutil v0.0.0-20210915084428-9f2466c3bd79 // indirect
	github.com/whatap/go-api/config v0.0.0-20210915084428-9f2466c3bd79 // indirect
	github.com/whatap/go-api/httpc v0.0.0-20210915084428-9f2466c3bd79
	github.com/whatap/go-api/method v0.0.0-20210915084428-9f2466c3bd79
	github.com/whatap/go-api/sql v0.0.0-20210915084428-9f2466c3bd79
	github.com/whatap/go-api/trace v0.0.0-20210915084428-9f2466c3bd79
	golang.org/x/text v0.3.7 // indirect
)

replace github.com/whatap/go-api/common/io => ../go-api/common/io
replace github.com/whatap/go-api/common/lang/pack/udp => ../go-api/common/lang/pack/udp
replace github.com/whatap/go-api/common/net => ../go-api/common/net
replace github.com/whatap/go-api/common/util/dateutil => ../go-api/common/util/dateutil
replace github.com/whatap/go-api/common/util/hash => ../go-api/common/util/hash
replace github.com/whatap/go-api/common/util/hexa32 => ../go-api/common/util/hexa32
replace github.com/whatap/go-api/common/util/keygen => ../go-api/common/util/keygen
replace github.com/whatap/go-api/common/util/paramtext => ../go-api/common/util/paramtext
replace github.com/whatap/go-api/common/util/stringutil => ../go-api/common/util/stringutil
replace github.com/whatap/go-api/common/util/urlutil => ../go-api/common/util/urlutil
replace github.com/whatap/go-api/config => ../go-api/config
replace github.com/whatap/go-api/httpc => ../go-api/httpc
replace github.com/whatap/go-api/method => ../go-api/method
replace github.com/whatap/go-api/sql => ../go-api/sql
replace github.com/whatap/go-api/trace => ../go-api/trace
