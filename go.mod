module logfleet

go 1.22.2

require (
	golang.org/x/crypto v0.19.0
	gopkg.in/yaml.v3 v3.0.1
)

require golang.org/x/sys v0.17.0 // indirect

replace (
	golang.org/x/crypto => /usr/share/gocode/src/golang.org/x/crypto
	gopkg.in/yaml.v3 => /usr/share/gocode/src/gopkg.in/yaml.v3
)
