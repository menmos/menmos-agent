package artifact

func platformAliases(platform string) []string {
	switch platform {
	case "linux":
		return []string{"linux"}
	case "windows":
		return []string{"windows"}
	case "darwin":
		return []string{"darwin", "macos"}
	default:
		return []string{}
	}
}

func architectureAliases(arch string) []string {
	switch arch {
	case "arm":
		return []string{"arm"}
	case "amd64":
		return []string{"amd64", "x64", "x86_64"}
	case "arm64":
		return []string{"arm64"}
	default:
		return []string{}
	}
}

func architectureEqual(archA, archB string) bool {
	for _, aAlias := range architectureAliases(archA) {
		for _, bAlias := range architectureAliases(archB) {
			if aAlias == bAlias {
				return true
			}
		}
	}
	return false
}

func platformEqual(platA, platB string) bool {
	for _, aAlias := range platformAliases(platA) {
		for _, bAlias := range platformAliases(platB) {
			if aAlias == bAlias {
				return true
			}
		}
	}
	return false
}
