module github.com/comicdeed/scurrytail

go 1.27.2

require (
	fyne.io/systray v1.12.2
	github.com/ardanlabs/conf/v3 v3.13.0
	github.com/godbus/dbus/v5 v5.1.0
	gopkg.in/yaml.v3 v3.0.1
)

require golang.org/x/sys v0.15.0 // indirect

replace fyne.io/systray => github.com/codemedic/fyne-io-systray v1.1.1-0.20261026181359-01d8092ab2cf
