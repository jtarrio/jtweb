module jacobo.tarrio.org/jtweb

go 1.18

require (
	github.com/gorilla/feeds v1.1.1
	github.com/litao91/goldmark-mathjax v0.0.0-20210217064022-a43cf739a50f
	github.com/microcosm-cc/bluemonday v1.0.23
	github.com/yuin/goldmark v1.5.4
	github.com/yuin/goldmark-highlighting v0.0.0-20220208100518-594be1970594
	golang.org/x/text v0.9.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/PuerkitoBio/goquery v1.8.1 // indirect
	github.com/andybalholm/cascadia v1.3.2 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
)

require (
	github.com/alecthomas/chroma v0.10.0 // indirect
	github.com/aymerick/douceur v0.2.0
	github.com/dlclark/regexp2 v1.9.0 // indirect
	github.com/gorilla/css v1.0.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mailerlite/mailerlite-go v1.0.2
	golang.org/x/net v0.9.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace github.com/mailerlite/mailerlite-go => ./email/mailerlitev2/mailerlite-go
