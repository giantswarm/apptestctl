package project

var (
	description = "Command line tool for using app platform in integration tests."
	gitSHA      = "n/a"
	name        = "apptestctl"
	source      = "https://github.com/giantswarm/apptestctl"
	version     = "0.22.1-dev"
)

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}
