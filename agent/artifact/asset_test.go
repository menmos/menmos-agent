package artifact

import "testing"

func Test_asset_stripExtension(t *testing.T) {
	type fields struct {
		FullName    string
		DownloadURL string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"simpleExtension", fields{FullName: "bing.txt"}, "bing"},
		{"spacesInFileName", fields{FullName: "bing bong.txt"}, "bing bong"},
		{"nestedExtensions", fields{FullName: "bing.tar.gz"}, "bing.tar"},
		{"noExtension", fields{FullName: "bingbong"}, "bingbong"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Asset{
				FullName:    tt.fields.FullName,
				DownloadURL: tt.fields.DownloadURL,
			}
			if got := a.stripExtension(); got != tt.want {
				t.Errorf("asset.stripExtension() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_asset_Name(t *testing.T) {
	type fields struct {
		FullName    string
		DownloadURL string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"simpleExtension", fields{FullName: "bing-windows-arm64.exe"}, "bing"},
		{"noExtension", fields{FullName: "bing-linux-arm64"}, "bing"},
		{"extensionOnly", fields{FullName: ".txt"}, "unknown"},
		{"noDashes", fields{FullName: "test.txt"}, "test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Asset{
				FullName:    tt.fields.FullName,
				DownloadURL: tt.fields.DownloadURL,
			}
			if got := a.Name(); got != tt.want {
				t.Errorf("asset.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_asset_Architecture(t *testing.T) {
	type fields struct {
		FullName    string
		DownloadURL string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"parseArm64", fields{FullName: "bing-linux-arm64"}, "arm64"},
		{"parseArm", fields{FullName: "bing-linux-arm"}, "arm"},
		{"parseAmd64", fields{FullName: "bing-linux-amd64"}, "amd64"},
		{"parseAmd64Aliasx86_64", fields{FullName: "bing-linux-x86_64"}, "amd64"},
		{"parseAmd64Aliasx64", fields{FullName: "bing-linux-x64"}, "amd64"},
		{"invertedPlatformAndArch", fields{FullName: "bing-arm64-linux"}, "arm64"},
		{"unknownArchitecture", fields{FullName: "bing-linux-riscv"}, "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Asset{
				FullName:    tt.fields.FullName,
				DownloadURL: tt.fields.DownloadURL,
			}
			if got := a.Architecture(); got != tt.want {
				t.Errorf("asset.Architecture() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_asset_Platform(t *testing.T) {
	type fields struct {
		FullName    string
		DownloadURL string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"parseLinux", fields{FullName: "bing-linux-arm64"}, "linux"},
		{"parseWindows", fields{FullName: "bing-windows-arm"}, "windows"},
		{"parseDarwin", fields{FullName: "bing-darwin-amd64"}, "darwin"},
		{"parseDarwinAliasMacos", fields{FullName: "bing-macos-x86_64"}, "darwin"},
		{"parseAmd64Aliasx64", fields{FullName: "bing-linux-x64"}, "linux"},
		{"invertedPlatformAndArch", fields{FullName: "bing-arm64-windows.exe"}, "windows"},
		{"unknownArchitecture", fields{FullName: "bing-solaris-x64"}, "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Asset{
				FullName:    tt.fields.FullName,
				DownloadURL: tt.fields.DownloadURL,
			}
			if got := a.Platform(); got != tt.want {
				t.Errorf("asset.Platform() = %v, want %v", got, tt.want)
			}
		})
	}
}
